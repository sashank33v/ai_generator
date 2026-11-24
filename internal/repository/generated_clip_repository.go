package repository

import (
	"context"
	"fmt"

	"generator/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GeneratedClipRepository struct {
	db *pgxpool.Pool
}

func NewGeneratedClipRepository(db *pgxpool.Pool) *GeneratedClipRepository {
	return &GeneratedClipRepository{db: db}
}

func (r *GeneratedClipRepository) Save(ctx context.Context, clip *model.GeneratedClip) error {

	query := `
		INSERT INTO generated_clips (
			video_id,
			start_time,
			end_time,
			ai_score,
			title,
			caption,
			file_path
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)

		ON CONFLICT (video_id)
		DO UPDATE SET
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			ai_score = EXCLUDED.ai_score,
			title = EXCLUDED.title,
			caption = EXCLUDED.caption,
			file_path = EXCLUDED.file_path,
			updated_at = NOW()

		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		clip.VideoID,
		clip.StartTime,
		clip.EndTime,
		clip.AIScore,
		clip.Title,
		clip.Caption,
		clip.FilePath,
	).Scan(
		&clip.ID,
		&clip.CreatedAt,
		&clip.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save generated clip: %w", err)
	}

	return nil
}

func (r *GeneratedClipRepository) GetByVideoID(
	ctx context.Context,
	videoID int64,
) (*model.GeneratedClip, error) {

	var clip model.GeneratedClip

	query := `
		SELECT
			id,
			video_id,
			start_time,
			end_time,
			ai_score,
			title,
			caption,
			file_path,
			created_at,
			updated_at
		FROM generated_clips
		WHERE video_id = $1
	`

	err := r.db.QueryRow(ctx, query, videoID).Scan(
		&clip.ID,
		&clip.VideoID,
		&clip.StartTime,
		&clip.EndTime,
		&clip.AIScore,
		&clip.Title,
		&clip.Caption,
		&clip.FilePath,
		&clip.CreatedAt,
		&clip.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to retrieve generated clip: %w", err,
		)
	}

	return &clip, nil
}
