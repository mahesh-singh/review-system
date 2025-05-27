package data

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type Models struct {
	ProcessedFiles      ProcessedFileModel
	Hotel               HotelModel
	Review              ReviewModel
	HotelProviderRating HotelProviderRatingModel
	Provider            ProviderModel
	Country             CountryModel
	ReviewGroup         ReviewGroupModel
}

func NewModels(dbtx DBTX) Models {
	return Models{
		ProcessedFiles:      ProcessedFileModel{DB: dbtx},
		Hotel:               HotelModel{DB: dbtx},
		Review:              ReviewModel{DB: dbtx},
		HotelProviderRating: HotelProviderRatingModel{DB: dbtx},
		Provider:            ProviderModel{DB: dbtx},
		Country:             CountryModel{DB: dbtx},
		ReviewGroup:         ReviewGroupModel{DB: dbtx},
	}
}
