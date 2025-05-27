package data

import (
	"context"
	"time"
)

type Hotel struct {
	ID        int64
	HotelID   int64
	Name      string
	Platform  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type HotelModel struct {
	DB DBTX
}

func (h HotelModel) Create(hotel *Hotel) error {
	query := `INSERT INTO hotels (hotel_id, name, platform) 
	VALUES ($1, $2, $3)
	ON CONFLICT (hotel_id) DO UPDATE SET
		name = EXCLUDED.name,
		platform = EXCLUDED.platform,
		updated_at = CURRENT_TIMESTAMP
	RETURNING id, created_at, updated_at`

	args := []interface{}{
		hotel.HotelID,
		hotel.Name,
		hotel.Platform,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return h.DB.QueryRowContext(ctx, query, args...).Scan(&hotel.ID, &hotel.CreatedAt, &hotel.UpdatedAt)
}

func (h HotelModel) GetByHotelID(hotelID int64) (*Hotel, error) {
	query := `SELECT id, hotel_id, name, platform, created_at, updated_at FROM hotels WHERE hotel_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	hotel := &Hotel{}
	err := h.DB.QueryRowContext(ctx, query, hotelID).Scan(
		&hotel.ID,
		&hotel.HotelID,
		&hotel.Name,
		&hotel.Platform,
		&hotel.CreatedAt,
		&hotel.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return hotel, nil
}
