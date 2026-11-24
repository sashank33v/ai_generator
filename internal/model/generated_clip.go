package model

import "time"

type GeneratedClip struct {
	ID        int64     `json:"id"`
	VideoID   int64     `json:"video_id"`
	StartTime int64     `json:"start_time"`
	EndTime   int64     `json:"end_time"`
	AIScore   int       `json:"ai_score"`
	Title     string    `json:"title"`
	Caption   string    `json:"caption"`
	FilePath  string    `json:"file_path"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
