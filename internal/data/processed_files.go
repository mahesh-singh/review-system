package data

import (
	"context"
	"database/sql"
	"time"
)

type ProcessedFile struct {
	ID           int64
	Filename     string
	S3Path       string
	ProcessedAt  time.Time
	RecordsCount int
	ErrorsCount  int
	Status       string // Success, Failed, Partial
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ProcessedFileModel struct {
	DB DBTX
}

func (p ProcessedFileModel) Create(file *ProcessedFile) error {

	query := `INSERT INTO processed_files (filename, s3path, processed_at, records_count, errors_count, status) 
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, created_at, updated_at`

	args := []interface{}{
		file.Filename,
		file.S3Path,
		file.ProcessedAt,
		file.RecordsCount,
		file.ErrorsCount,
		file.Status}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	return p.DB.QueryRowContext(ctx, query, args...).Scan(&file.ID, &file.CreatedAt, &file.UpdatedAt)
}

func (p ProcessedFileModel) Update(file *ProcessedFile) error {
	query := `UPDATE processed_files 
	SET records_count = $1, errors_count = $2, status = $3, updated_at = CURRENT_TIMESTAMP
	WHERE id = $4`

	args := []interface{}{
		file.RecordsCount,
		file.ErrorsCount,
		file.Status,
		file.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := p.DB.ExecContext(ctx, query, args...)
	return err
}

func (p ProcessedFileModel) IsProcessed(s3Path string) (bool, error) {
	query := `SELECT id FROM processed_files WHERE s3path = $1 AND status IN ('Success', 'Partial')`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int64
	err := p.DB.QueryRowContext(ctx, query, s3Path).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
