package data

import (
	"context"
	"time"
)

type Review struct {
	ID                      int64
	HotelReviewID           int64
	HotelID                 int64
	ProviderID              int
	Rating                  float64
	CheckInMonthYear        string
	EncryptedReviewData     string
	FormattedRating         string
	FormattedReviewDate     string
	RatingText              string
	ResponderName           string
	ResponseDateText        string
	ResponseTranslateSource string
	ReviewComments          string
	ReviewNegatives         string
	ReviewPositives         string
	ReviewProviderLogo      string
	ReviewProviderText      string
	ReviewTitle             string
	TranslateSource         string
	TranslateTarget         string
	ReviewDate              time.Time
	OriginalTitle           string
	OriginalComment         string
	FormattedResponseDate   string
	IsShowReviewResponse    bool

	// Reviewer Info (merged directly)
	ReviewerCountryName     string
	ReviewerDisplayName     string
	ReviewerFlagName        string
	ReviewerGroupName       string
	ReviewerRoomTypeName    string
	ReviewerCountryID       *int
	ReviewerLengthOfStay    int
	ReviewerGroupID         *int
	ReviewerReviewCount     int
	ReviewerIsExpert        bool
	ReviewerShowGlobalIcon  bool
	ReviewerShowReviewCount bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

type ReviewModel struct {
	DB DBTX
}

func (r ReviewModel) Create(review *Review) error {
	query := `INSERT INTO reviews (
		hotel_review_id, hotel_id, provider_id, rating, check_in_month_year,
		encrypted_review_data, formatted_rating, formatted_review_date, rating_text,
		responder_name, response_date_text, response_translate_source, review_comments,
		review_negatives, review_positives, review_provider_logo, review_provider_text,
		review_title, translate_source, translate_target, review_date, original_title,
		original_comment, formatted_response_date, is_show_review_response,
		reviewer_country_name, reviewer_display_name, reviewer_flag_name,
		reviewer_group_name, reviewer_room_type_name, reviewer_country_id,
		reviewer_length_of_stay, reviewer_group_id, reviewer_review_count,
		reviewer_is_expert, reviewer_show_global_icon, reviewer_show_review_count
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
		$18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32,
		$33, $34, $35, $36, $37
	)
	ON CONFLICT (hotel_review_id) DO UPDATE SET
		rating = EXCLUDED.rating,
		review_comments = EXCLUDED.review_comments,
		updated_at = CURRENT_TIMESTAMP
	RETURNING id, created_at, updated_at`

	args := []interface{}{
		review.HotelReviewID,
		review.HotelID,
		review.ProviderID,
		review.Rating,
		review.CheckInMonthYear,
		review.EncryptedReviewData,
		review.FormattedRating,
		review.FormattedReviewDate,
		review.RatingText,
		review.ResponderName,
		review.ResponseDateText,
		review.ResponseTranslateSource,
		review.ReviewComments,
		review.ReviewNegatives,
		review.ReviewPositives,
		review.ReviewProviderLogo,
		review.ReviewProviderText,
		review.ReviewTitle,
		review.TranslateSource,
		review.TranslateTarget,
		review.ReviewDate,
		review.OriginalTitle,
		review.OriginalComment,
		review.FormattedResponseDate,
		review.IsShowReviewResponse,
		review.ReviewerCountryName,
		review.ReviewerDisplayName,
		review.ReviewerFlagName,
		review.ReviewerGroupName,
		review.ReviewerRoomTypeName,
		review.ReviewerCountryID,
		review.ReviewerLengthOfStay,
		review.ReviewerGroupID,
		review.ReviewerReviewCount,
		review.ReviewerIsExpert,
		review.ReviewerShowGlobalIcon,
		review.ReviewerShowReviewCount,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.DB.QueryRowContext(ctx, query, args...).Scan(&review.ID, &review.CreatedAt, &review.UpdatedAt)
}
