package data

import (
	"context"
	"database/sql"
	"time"
)

type Provider struct {
	ID        int
	Name      string
	CreatedAt time.Time
}

type ProviderModel struct {
	DB DBTX
}

func (p ProviderModel) CreateOrGet(name string) (*Provider, error) {
	// First try to get existing provider
	query := `SELECT id, name, created_at FROM providers WHERE name = $1`
	provider := &Provider{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRowContext(ctx, query, name).Scan(&provider.ID, &provider.Name, &provider.CreatedAt)
	if err == nil {
		return provider, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new provider if not found
	insertQuery := `INSERT INTO providers (name) VALUES ($1) RETURNING id, name, created_at`
	err = p.DB.QueryRowContext(ctx, insertQuery, name).Scan(&provider.ID, &provider.Name, &provider.CreatedAt)
	return provider, err
}
