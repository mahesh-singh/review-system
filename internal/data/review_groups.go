package data

import (
	"context"
	"database/sql"
	"time"
)

type ReviewGroup struct {
	ID        int
	Name      string
	CreatedAt time.Time
}

type ReviewGroupModel struct {
	DB DBTX
}

func (r ReviewGroupModel) CreateOrGet(name string) (*ReviewGroup, error) {
	// First try to get existing review group
	query := `SELECT id, name, created_at FROM review_groups WHERE name = $1`
	group := &ReviewGroup{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, query, name).Scan(&group.ID, &group.Name, &group.CreatedAt)
	if err == nil {
		return group, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new review group if not found
	insertQuery := `INSERT INTO review_groups (name) VALUES ($1) RETURNING id, name, created_at`
	err = r.DB.QueryRowContext(ctx, insertQuery, name).Scan(&group.ID, &group.Name, &group.CreatedAt)
	return group, err
}
