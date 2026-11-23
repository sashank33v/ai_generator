package model

import "time"

type Video struct {
	ID           int64     `json:"id"`
	OriginalName string    `json:"original_name"`
	FileName     string    `json:"file_name"`
	FilePath     string    `json:"-"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}
