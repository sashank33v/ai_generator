package model

type RankedClip struct {
	Candidate ClipCandidate `json:"candidate"`
	Score     ClipScore     `json:"score"`
}
