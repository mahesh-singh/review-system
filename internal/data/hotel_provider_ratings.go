package data

import (
	"context"
	"time"
)

type HotelProviderRating struct {
	ID                 int64
	HotelID            int64
	ProviderID         int
	ProviderName       string
	OverallScore       float64
	ReviewCount        int
	Cleanliness        *float64
	Facilities         *float64
	Location           *float64
	RoomComfortQuality *float64
	Service            *float64
	ValueForMoney      *float64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type HotelProviderRatingModel struct {
	DB DBTX
}

func (h HotelProviderRatingModel) Create(rating *HotelProviderRating) error {
	query := `INSERT INTO hotel_provider_ratings (
		hotel_id, provider_id, provider_name, overall_score, review_count,
		cleanliness, facilities, location, room_comfort_quality, service, value_for_money
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	ON CONFLICT (hotel_id, provider_id) DO UPDATE SET
		overall_score = EXCLUDED.overall_score,
		review_count = EXCLUDED.review_count,
		cleanliness = EXCLUDED.cleanliness,
		facilities = EXCLUDED.facilities,
		location = EXCLUDED.location,
		room_comfort_quality = EXCLUDED.room_comfort_quality,
		service = EXCLUDED.service,
		value_for_money = EXCLUDED.value_for_money,
		updated_at = CURRENT_TIMESTAMP
	RETURNING id, created_at, updated_at`

	args := []interface{}{
		rating.HotelID,
		rating.ProviderID,
		rating.ProviderName,
		rating.OverallScore,
		rating.ReviewCount,
		rating.Cleanliness,
		rating.Facilities,
		rating.Location,
		rating.RoomComfortQuality,
		rating.Service,
		rating.ValueForMoney,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return h.DB.QueryRowContext(ctx, query, args...).Scan(&rating.ID, &rating.CreatedAt, &rating.UpdatedAt)
}
