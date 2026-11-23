package repository

import (
	"context"
	"fmt"

	"generator/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProcessingJobRepository struct {
	db *pgxpool.Pool
}

func NewProcessingJobRepository(db *pgxpool.Pool) *ProcessingJobRepository {
	return &ProcessingJobRepository{
		db: db,
	}
}

func (r *ProcessingJobRepository) Create(ctx context.Context, job *model.ProcessingJob) error {
	query := `
		INSERT INTO processing_jobs (video_id, status)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		job.VideoID,
		job.Status,
	).Scan(
		&job.ID,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create processing job: %w", err)
	}

	return nil
}
func (r *ProcessingJobRepository) GetByID(ctx context.Context, id int64) (*model.ProcessingJob, error) {
	var job model.ProcessingJob
	query := `SELECT id, video_id, status, error_message, created_at, updated_at FROM processing_jobs WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&job.ID,
		&job.VideoID,
		&job.Status,
		&job.ErrorMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve processing job: %w", err)
	}

	return &job, nil
}
