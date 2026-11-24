package model

import "time"

type Transcription struct {
	ID        int64     `json:"id"`
	VideoID   int64     `json:"video_id"`
	Language  string    `json:"language"`
	StartTime int64     `json:"start_time"`
	EndTime   int64     `json:"end_time"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
