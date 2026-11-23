package repository

import (
	"context"
	"fmt"

	"generator/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type VideoRepository struct {
	db *pgxpool.Pool
}

func NewVideoRepository(db *pgxpool.Pool) *VideoRepository {
	return &VideoRepository{
		db: db,
	}
}

func (r *VideoRepository) Create(ctx context.Context, video *model.Video) error {
	query := `
		INSERT INTO videos (
			original_name,
			file_name,
			file_path,
			file_size,
			mime_type,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		video.OriginalName,
		video.FileName,
		video.FilePath,
		video.FileSize,
		video.MimeType,
		video.Status,
	).Scan(&video.ID, &video.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create video: %w", err)
	}

	return nil
}
func (r *VideoRepository) CreateWithJob(
	ctx context.Context,
	video *model.Video,
	job *model.ProcessingJob,
) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
        INSERT INTO videos (
            original_name,
            file_name,
            file_path,
            file_size,
            mime_type,
            status
        )
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at
    `

	err = tx.QueryRow(
		ctx,
		query,
		video.OriginalName,
		video.FileName,
		video.FilePath,
		video.FileSize,
		video.MimeType,
		video.Status,
	).Scan(&video.ID, &video.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create video: %w", err)
	}

	job.VideoID = video.ID

	jobQuery := `
        INSERT INTO processing_jobs (video_id, status)
        VALUES ($1, $2)
        RETURNING id, created_at, updated_at
    `

	err = tx.QueryRow(
		ctx,
		jobQuery,
		job.VideoID,
		job.Status,
	).Scan(&job.ID, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create processing job: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
func (r *VideoRepository) GetByID(
	ctx context.Context,
	id int64,
) (*model.Video, error) {

	query := `
        SELECT id, original_name, file_name, file_path,
               file_size, mime_type, status, created_at
        FROM videos
        WHERE id = $1
    `

	video := &model.Video{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&video.ID,
		&video.OriginalName,
		&video.FileName,
		&video.FilePath,
		&video.FileSize,
		&video.MimeType,
		&video.Status,
		&video.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	return video, nil
}
