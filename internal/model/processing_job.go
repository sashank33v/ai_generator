package model

import "time"

type ProcessingJob struct {
	ID           int64     `json:"id"`
	VideoID      int64     `json:"video_id"`
	Status       string    `json:"status"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
