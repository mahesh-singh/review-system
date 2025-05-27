package jsonl_processing

import "time"

type ProcessingConfig struct {
	BatchSize           int           // Number of records to process in each batch
	MaxRetries          int           // Maximum number of retries for failed operations
	RetryDelay          time.Duration // Delay between retries
	MaxErrorsPercentage float64       // Maximum percentage of errors before stopping (0-100)
	ContextTimeout      time.Duration // Timeout for database operations
}

func DefaultProcessingConfig() *ProcessingConfig {
	return &ProcessingConfig{
		BatchSize:           100,
		MaxRetries:          3,
		RetryDelay:          time.Second * 2,
		MaxErrorsPercentage: 10.0,
		ContextTimeout:      time.Minute * 5,
	}
}

// ValidateConfig validates the processing configuration
func ValidateConfig(config *ProcessingConfig) error {
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}
	if config.MaxRetries < 0 {
		config.MaxRetries = 0
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = time.Second * 2
	}
	if config.MaxErrorsPercentage < 0 || config.MaxErrorsPercentage > 100 {
		config.MaxErrorsPercentage = 10.0
	}
	if config.ContextTimeout <= 0 {
		config.ContextTimeout = time.Minute * 5
	}
	return nil
}
