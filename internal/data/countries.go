package data

import (
	"context"
	"database/sql"
	"time"
)

type Country struct {
	ID        int
	Name      string
	Flag      string
	CreatedAt time.Time
}

type CountryModel struct {
	DB DBTX
}

func (c CountryModel) CreateOrGet(name, flag string) (*Country, error) {
	// First try to get existing country
	query := `SELECT id, name, flag, created_at FROM countries WHERE name = $1`
	country := &Country{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, name).Scan(&country.ID, &country.Name, &country.Flag, &country.CreatedAt)
	if err == nil {
		return country, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new country if not found
	insertQuery := `INSERT INTO countries (name, flag) VALUES ($1, $2) RETURNING id, name, flag, created_at`
	err = c.DB.QueryRowContext(ctx, insertQuery, name, flag).Scan(&country.ID, &country.Name, &country.Flag, &country.CreatedAt)
	return country, err
}
