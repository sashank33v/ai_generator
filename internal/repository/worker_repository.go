package repository

import (
	"context"
	"errors"
	"fmt"

	"generator/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkerRepository struct {
	db *pgxpool.Pool
}

func NewWorkerRepository(db *pgxpool.Pool) *WorkerRepository {
	return &WorkerRepository{
		db: db,
	}
}

func (r *WorkerRepository) ClaimNextJob(
	ctx context.Context,
) (*model.ProcessingJob, error) {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	job := &model.ProcessingJob{}

	query := `
		SELECT id, video_id, status, error_message, created_at, updated_at
		FROM processing_jobs
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`

	err = tx.QueryRow(ctx, query).Scan(
		&job.ID,
		&job.VideoID,
		&job.Status,
		&job.ErrorMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to claim processing job: %w", err)
	}

	updateQuery := `
		UPDATE processing_jobs
		SET status = 'PROCESSING',
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err = tx.Exec(ctx, updateQuery, job.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update job status: %w", err)
	}

	videoQuery := `
		UPDATE videos
		SET status = 'PROCESSING'
		WHERE id = $1
	`

	_, err = tx.Exec(ctx, videoQuery, job.VideoID)
	if err != nil {
		return nil, fmt.Errorf("failed to update video status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	job.Status = "PROCESSING"

	return job, nil
}
func (r *WorkerRepository) UpdateJobStatus(
	ctx context.Context,
	jobID int64,
	videoID int64,
	status string,
	errorMessage *string,
) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	jobQuery := `
        UPDATE processing_jobs
        SET status = $1,
            error_message = $2,
            updated_at = NOW()
        WHERE id = $3
    `

	_, err = tx.Exec(ctx, jobQuery, status, errorMessage, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	videoQuery := `
        UPDATE videos
        SET status = $1
        WHERE id = $2
    `

	_, err = tx.Exec(ctx, videoQuery, status, videoID)
	if err != nil {
		return fmt.Errorf("failed to update video status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit status update: %w", err)
	}

	return nil
}
func (r *WorkerRepository) RecoverInterruptedJobs(
	ctx context.Context,
) error {

	query := `
		WITH recovered AS (
			UPDATE processing_jobs
			SET status = 'PENDING',
			    error_message = NULL,
			    updated_at = NOW()
			WHERE status IN (
				'PROCESSING',
				'TRANSCRIBING',
				'ANALYZING',
				'GENERATING'
			)
			RETURNING video_id
		)
		UPDATE videos
		SET status = 'PENDING'
		WHERE id IN (
			SELECT video_id FROM recovered
		)
	`

	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to recover jobs: %w", err)
	}

	return nil
}

var ErrJobNotRetryable = errors.New("only failed jobs can be retried")

func (r *WorkerRepository) RetryFailedJob(ctx context.Context, jobID int64) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var videoID int64
	var status string

	err = tx.QueryRow(ctx, `
		SELECT video_id, status
		FROM processing_jobs
		WHERE id = $1
		FOR UPDATE
	`, jobID).Scan(&videoID, &status)

	if err != nil {
		return err
	}

	if status != "FAILED" {
		return ErrJobNotRetryable
	}

	_, err = tx.Exec(ctx, `
		UPDATE processing_jobs
		SET status = 'PENDING',
		    error_message = NULL,
		    updated_at = NOW()
		WHERE id = $1
	`, jobID)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE videos
		SET status = 'PENDING'
		WHERE id = $1
	`, videoID)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
func (r *WorkerRepository) ResetJobArtifacts(ctx context.Context, videoID int64) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		DELETE FROM transcriptions
		WHERE video_id = $1
	`, videoID)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM generated_clips
		WHERE video_id = $1
	`, videoID)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
