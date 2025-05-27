package jsonl_processing

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/mahesh-singh/review-system/internal/data"
	"github.com/mahesh-singh/review-system/internal/s3"
)

type JSONLProcessingService struct {
	config *ProcessingConfig
	db     *sql.DB
	models data.Models
	logger *slog.Logger
}

func NewJSONLProcessingService(db *sql.DB, config *ProcessingConfig, logger *slog.Logger) *JSONLProcessingService {
	if config == nil {
		config = DefaultProcessingConfig()
	}

	models := data.NewModels(db)

	ValidateConfig(config)

	return &JSONLProcessingService{
		config: config,
		db:     db,
		models: models,
		logger: logger,
	}
}

func (s *JSONLProcessingService) ProcessJSONLFile(
	ctx context.Context,
	reader io.Reader,
	filename string,
	s3Path string,
) (*ProcessingResult, error) {
	// Check if file has already been processed
	isProcessed, err := s.models.ProcessedFiles.IsProcessed(s3Path)
	if err != nil {
		return nil, fmt.Errorf("error checking if file is processed: %w", err)
	}

	if isProcessed {
		log.Printf("File %s has already been processed, skipping", filename)
		return &ProcessingResult{}, nil
	}

	// Create processed file record
	processedFile := &data.ProcessedFile{
		Filename:    filename,
		S3Path:      s3Path,
		ProcessedAt: time.Now(),
		Status:      "Processing",
	}

	err = s.models.ProcessedFiles.Create(processedFile)
	if err != nil {
		return nil, fmt.Errorf("error creating processed file record: %w", err)
	}

	// Start processing
	startTime := time.Now()
	result := &ProcessingResult{
		Errors: make([]ProcessingError, 0),
	}

	// Process file line by line
	scanner := bufio.NewScanner(reader)
	lineNumber := 0
	batch := make([]*HotelReviewData, 0, s.config.BatchSize)

	for scanner.Scan() {
		lineNumber++
		result.TotalRecords++

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Parse JSONL line
		var reviewData HotelReviewData
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &reviewData); err != nil {
			result.ErrorRecords++
			result.Errors = append(result.Errors, ProcessingError{
				LineNumber: lineNumber,
				Error:      fmt.Sprintf("JSON parse error: %v", err),
				RawData:    line,
			})
			continue
		}

		batch = append(batch, &reviewData)

		// Process batch when it reaches the configured size
		if len(batch) >= s.config.BatchSize {
			batchResult := s.processBatch(ctx, batch)
			s.mergeBatchResult(result, batchResult)
			batch = batch[:0] // Reset batch

			// Check error percentage
			if s.shouldStopProcessing(result) {
				log.Printf("Stopping processing due to high error rate: %.2f%%",
					float64(result.ErrorRecords)/float64(result.TotalRecords)*100)
				break
			}
		}
	}

	// Process remaining batch
	if len(batch) > 0 {
		batchResult := s.processBatch(ctx, batch)
		s.mergeBatchResult(result, batchResult)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Calculate duration and update processed file record
	result.Duration = time.Since(startTime)

	// Determine final status
	status := "Success"
	if result.ErrorRecords > 0 {
		if result.SuccessRecords == 0 {
			status = "Failed"
		} else {
			status = "Partial"
		}
	}

	// Update processed file record
	processedFile.RecordsCount = result.SuccessRecords
	processedFile.ErrorsCount = result.ErrorRecords
	processedFile.Status = status

	if err := s.models.ProcessedFiles.Update(processedFile); err != nil {
		log.Printf("Warning: Failed to update processed file record: %v", err)
	}

	log.Printf("Processing completed for %s: %d total, %d success, %d errors in %v",
		filename, result.TotalRecords, result.SuccessRecords, result.ErrorRecords, result.Duration)

	return result, nil
}

// processBatch processes a batch of hotel review data
func (s *JSONLProcessingService) processBatch(ctx context.Context, batch []*HotelReviewData) *ProcessingResult {
	result := &ProcessingResult{
		Errors: make([]ProcessingError, 0),
	}

	for _, reviewData := range batch {
		select {
		case <-ctx.Done():
			return result
		default:
		}

		if err := s.processHotelReviewData(ctx, reviewData); err != nil {
			result.ErrorRecords++
			result.Errors = append(result.Errors, ProcessingError{
				Error: fmt.Sprintf("Processing error for hotel %d: %v", reviewData.HotelID, err),
			})
		} else {
			result.SuccessRecords++
		}
		result.TotalRecords++
	}

	return result
}

func (s *JSONLProcessingService) processHotelReviewData(ctx context.Context, reviewData *HotelReviewData) error {
	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, s.config.ContextTimeout)
	defer cancel()

	// Start transaction
	tx, err := s.db.BeginTx(ctxWithTimeout, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Will be ignored if tx.Commit() succeeds

	// Process with retry logic
	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		err = s.processInTransaction(ctxWithTimeout, tx, reviewData)
		if err == nil {
			// Success - commit transaction
			if commitErr := tx.Commit(); commitErr != nil {
				return fmt.Errorf("failed to commit transaction: %w", commitErr)
			}
			return nil
		}

		// If it's the last attempt or a non-retryable error, return
		if attempt == s.config.MaxRetries || !s.isRetryableError(err) {
			return err
		}

		// Wait before retry
		select {
		case <-time.After(s.config.RetryDelay):
		case <-ctxWithTimeout.Done():
			return ctxWithTimeout.Err()
		}

		log.Printf("Retrying hotel %d processing (attempt %d/%d): %v",
			reviewData.HotelID, attempt+1, s.config.MaxRetries, err)
	}

	return err
}

func (s *JSONLProcessingService) processInTransaction(ctx context.Context, tx *sql.Tx, reviewData *HotelReviewData) error {
	// Create models with transaction
	hotelModel := &data.HotelModel{DB: tx}
	reviewModel := &data.ReviewModel{DB: tx}
	hotelProviderRatingModel := &data.HotelProviderRatingModel{DB: tx}
	providerModel := &data.ProviderModel{DB: tx}
	countryModel := &data.CountryModel{DB: tx}
	reviewGroupModel := &data.ReviewGroupModel{DB: tx}

	// 1. Process Hotel
	hotel := &data.Hotel{
		HotelID:  reviewData.HotelID,
		Name:     reviewData.HotelName,
		Platform: reviewData.Platform,
	}

	if err := hotelModel.Create(hotel); err != nil {
		return fmt.Errorf("failed to create/update hotel: %w", err)
	}

	// 2. Process Provider for review
	provider, err := providerModel.CreateOrGet(reviewData.Comment.ReviewProviderText)
	if err != nil {
		return fmt.Errorf("failed to get/create provider: %w", err)
	}

	// 3. Process Country and ReviewGroup for reviewer
	var countryID *int
	var reviewGroupID *int

	if reviewData.Comment.ReviewerInfo.CountryName != "" {
		country, err := countryModel.CreateOrGet(
			reviewData.Comment.ReviewerInfo.CountryName,
			reviewData.Comment.ReviewerInfo.FlagName,
		)
		if err != nil {
			return fmt.Errorf("failed to get/create country: %w", err)
		}
		countryID = &country.ID
	}

	if reviewData.Comment.ReviewerInfo.ReviewGroupName != "" {
		reviewGroup, err := reviewGroupModel.CreateOrGet(reviewData.Comment.ReviewerInfo.ReviewGroupName)
		if err != nil {
			return fmt.Errorf("failed to get/create review group: %w", err)
		}
		reviewGroupID = &reviewGroup.ID
	}

	// 4. Process Review
	review := &data.Review{
		HotelReviewID:           reviewData.Comment.HotelReviewID,
		HotelID:                 reviewData.HotelID,
		ProviderID:              provider.ID,
		Rating:                  reviewData.Comment.Rating,
		CheckInMonthYear:        reviewData.Comment.CheckInDateMonthAndYear,
		EncryptedReviewData:     reviewData.Comment.EncryptedReviewData,
		FormattedRating:         reviewData.Comment.FormattedRating,
		FormattedReviewDate:     reviewData.Comment.FormattedReviewDate,
		RatingText:              reviewData.Comment.RatingText,
		ResponderName:           reviewData.Comment.ResponderName,
		ResponseDateText:        reviewData.Comment.ResponseDateText,
		ResponseTranslateSource: reviewData.Comment.ResponseTranslateSource,
		ReviewComments:          reviewData.Comment.ReviewComments,
		ReviewNegatives:         reviewData.Comment.ReviewNegatives,
		ReviewPositives:         reviewData.Comment.ReviewPositives,
		ReviewProviderLogo:      reviewData.Comment.ReviewProviderLogo,
		ReviewProviderText:      reviewData.Comment.ReviewProviderText,
		ReviewTitle:             reviewData.Comment.ReviewTitle,
		TranslateSource:         reviewData.Comment.TranslateSource,
		TranslateTarget:         reviewData.Comment.TranslateTarget,
		ReviewDate:              reviewData.Comment.ReviewDate,
		OriginalTitle:           reviewData.Comment.OriginalTitle,
		OriginalComment:         reviewData.Comment.OriginalComment,
		FormattedResponseDate:   reviewData.Comment.FormattedResponseDate,
		IsShowReviewResponse:    reviewData.Comment.IsShowReviewResponse,

		// Reviewer Info
		ReviewerCountryName:     reviewData.Comment.ReviewerInfo.CountryName,
		ReviewerDisplayName:     reviewData.Comment.ReviewerInfo.DisplayMemberName,
		ReviewerFlagName:        reviewData.Comment.ReviewerInfo.FlagName,
		ReviewerGroupName:       reviewData.Comment.ReviewerInfo.ReviewGroupName,
		ReviewerRoomTypeName:    reviewData.Comment.ReviewerInfo.RoomTypeName,
		ReviewerCountryID:       countryID,
		ReviewerLengthOfStay:    reviewData.Comment.ReviewerInfo.LengthOfStay,
		ReviewerGroupID:         reviewGroupID,
		ReviewerReviewCount:     reviewData.Comment.ReviewerInfo.ReviewerReviewedCount,
		ReviewerIsExpert:        reviewData.Comment.ReviewerInfo.IsExpertReviewer,
		ReviewerShowGlobalIcon:  reviewData.Comment.ReviewerInfo.IsShowGlobalIcon,
		ReviewerShowReviewCount: reviewData.Comment.ReviewerInfo.IsShowReviewedCount,
	}

	if err := reviewModel.Create(review); err != nil {
		return fmt.Errorf("failed to create/update review: %w", err)
	}

	// 5. Process Hotel Provider Ratings
	for _, providerRating := range reviewData.OverallByProviders {
		// Get or create provider for rating
		ratingProvider, err := providerModel.CreateOrGet(providerRating.Provider)
		if err != nil {
			return fmt.Errorf("failed to get/create rating provider: %w", err)
		}

		rating := &data.HotelProviderRating{
			HotelID:            reviewData.HotelID,
			ProviderID:         ratingProvider.ID,
			ProviderName:       providerRating.Provider,
			OverallScore:       providerRating.OverallScore,
			ReviewCount:        providerRating.ReviewCount,
			Cleanliness:        &providerRating.Grades.Cleanliness,
			Facilities:         &providerRating.Grades.Facilities,
			Location:           &providerRating.Grades.Location,
			RoomComfortQuality: &providerRating.Grades.RoomComfortAndQuality,
			Service:            &providerRating.Grades.Service,
			ValueForMoney:      &providerRating.Grades.ValueForMoney,
		}

		if err := hotelProviderRatingModel.Create(rating); err != nil {
			return fmt.Errorf("failed to create/update hotel provider rating: %w", err)
		}
	}

	return nil
}

// ProcessMultipleFiles processes multiple JSONL files concurrently
func (s *JSONLProcessingService) ProcessMultipleFiles(
	ctx context.Context,
	files []s3.FileInfo,
	fileReader s3.FileReader,
	maxConcurrency int,
) (map[string]*ProcessingResult, error) {
	semaphore := make(chan struct{}, maxConcurrency)
	results := make(map[string]*ProcessingResult)
	resultsChan := make(chan FileResult, len(files))
	errorsChan := make(chan error, len(files))

	// Process files concurrently
	for _, file := range files {
		go func(f s3.FileInfo) {
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			reader, err := fileReader.GetReader(ctx, file.S3Path)
			if err != nil {
				errorsChan <- fmt.Errorf("error processing %s: %w", f.Key, err)
				return
			}
			result, err := s.ProcessJSONLFile(ctx, reader, f.Key, f.S3Path)
			if err != nil {
				errorsChan <- fmt.Errorf("error processing %s: %w", f.Key, err)
				return
			}

			resultsChan <- FileResult{
				Filename: f.Key,
				Result:   result,
			}
		}(file)
	}

	// Collect results
	completed := 0
	var firstError error

	for completed < len(files) {
		select {
		case result := <-resultsChan:
			results[result.Filename] = result.Result
			completed++
		case err := <-errorsChan:
			if firstError == nil {
				firstError = err
			}
			completed++
		case <-ctx.Done():
			return results, ctx.Err()
		}
	}

	return results, firstError
}

// GetProcessedFileStatus returns the status of a processed file
func (s *JSONLProcessingService) GetProcessedFileStatus(s3Path string) (bool, error) {
	return s.models.ProcessedFiles.IsProcessed(s3Path)
}

func (s *JSONLProcessingService) mergeBatchResult(main *ProcessingResult, batch *ProcessingResult) {
	main.SuccessRecords += batch.SuccessRecords
	main.ErrorRecords += batch.ErrorRecords
	main.Errors = append(main.Errors, batch.Errors...)
}

func (s *JSONLProcessingService) shouldStopProcessing(result *ProcessingResult) bool {
	if result.TotalRecords < 100 { // Don't stop early if we haven't processed enough records
		return false
	}

	errorPercentage := float64(result.ErrorRecords) / float64(result.TotalRecords) * 100
	return errorPercentage > s.config.MaxErrorsPercentage
}

func (s *JSONLProcessingService) isRetryableError(err error) bool {
	// Database connection errors, timeout errors, etc. are retryable
	// Constraint violations, JSON parsing errors, etc. are not retryable
	errorStr := strings.ToLower(err.Error())

	// Non-retryable errors
	nonRetryableErrors := []string{
		"duplicate key",
		"constraint violation",
		"invalid input syntax",
		"json parse error",
		"invalid json",
	}

	for _, nonRetryable := range nonRetryableErrors {
		if strings.Contains(errorStr, nonRetryable) {
			return false
		}
	}

	// Retryable errors
	retryableErrors := []string{
		"connection",
		"timeout",
		"network",
		"temporary",
		"deadlock",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(errorStr, retryable) {
			return true
		}
	}

	// Default to retryable for unknown errors
	return true
}
