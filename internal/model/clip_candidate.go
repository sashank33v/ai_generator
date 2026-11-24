package model

type ClipCandidate struct {
	VideoID   int64  `json:"video_id"`
	StartTime int64  `json:"start_time"`
	Endtime   int64  `json:"end_time"`
	Text      string `json:"text"`
}
