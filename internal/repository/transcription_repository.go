package repository

import (
	"context"
	"fmt"
	"generator/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TranscriptionRepository struct {
	db *pgxpool.Pool
}

func NewTranscriptionRepository(db *pgxpool.Pool) *TranscriptionRepository {
	return &TranscriptionRepository{db: db}
}

func (r *TranscriptionRepository) CreateMany(ctx context.Context, transcriptions []model.Transcription) error {
	if len(transcriptions) == 0 {
		return nil
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Failed to begin transcription transcation: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, transcription := range transcriptions {

		_, err := tx.Exec(ctx, `
			INSERT INTO transcriptions (
				video_id,
				language,
				start_time,
				end_time,
				text
			)
			VALUES ($1, $2, $3, $4, $5)
		`,
			transcription.VideoID,
			transcription.Language,
			transcription.StartTime,
			transcription.EndTime,
			transcription.Text,
		)

		if err != nil {
			return fmt.Errorf("failed to save transcription: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transcription transaction: %w", err)
	}

	return nil
}

func (r *TranscriptionRepository) GetByVideoID(
	ctx context.Context,
	videoID int64,
) ([]model.Transcription, error) {

	rows, err := r.db.Query(ctx, `
		SELECT
			id, video_id, language,
			start_time, end_time, text, created_at
		FROM transcriptions
		WHERE video_id = $1
		ORDER BY start_time ASC, id ASC
	`, videoID)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch transcriptions: %w", err)
	}
	defer rows.Close()

	transcriptions := make([]model.Transcription, 0)

	for rows.Next() {
		var t model.Transcription

		err := rows.Scan(
			&t.ID,
			&t.VideoID,
			&t.Language,
			&t.StartTime,
			&t.EndTime,
			&t.Text,
			&t.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan transcription: %w", err)
		}

		transcriptions = append(transcriptions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read transcription rows: %w", err)
	}

	return transcriptions, nil
}
