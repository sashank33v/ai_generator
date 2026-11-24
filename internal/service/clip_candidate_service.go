package service

import (
	"context"
	"fmt"
	"strings"

	"generator/internal/model"
	"generator/internal/repository"
)

const (
	minClipDuration int64 = 35_000
	maxClipDuration int64 = 60_000
)

type ClipCandidateService struct {
	transcriptionRepo *repository.TranscriptionRepository
}

func NewClipCandidateService(transcriptionRepo *repository.TranscriptionRepository) *ClipCandidateService {
	return &ClipCandidateService{
		transcriptionRepo: transcriptionRepo,
	}
}

func GenerateClipCandidates(segments []model.Transcription) []model.ClipCandidate {

	candidates := make([]model.ClipCandidate, 0)

	for i := range segments {

		if i > 0 && !isSentenceEnd(segments[i-1].Text) {
			continue
		}

		start := segments[i].StartTime
		var words []string

		for j := i; j < len(segments); j++ {

			duration := segments[j].EndTime - start

			if duration > maxClipDuration {
				break
			}

			words = append(words, strings.TrimSpace(segments[j].Text))

			if duration >= minClipDuration &&
				isSentenceEnd(segments[j].Text) {

				candidates = append(candidates, model.ClipCandidate{
					VideoID:   segments[i].VideoID,
					StartTime: start,
					Endtime:   segments[j].EndTime,
					Text:      strings.Join(words, " "),
				})

				break
			}
		}
	}

	return candidates
}

func isSentenceEnd(text string) bool {
	text = strings.TrimSpace(text)

	return strings.HasSuffix(text, ".") ||
		strings.HasSuffix(text, "?") ||
		strings.HasSuffix(text, "!")
}
func (s *ClipCandidateService) GetCandidates(ctx context.Context, videoID int64) ([]model.ClipCandidate, error) {
	segments, err := s.transcriptionRepo.GetByVideoID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transcriptions: %w", err)
	}
	if len(segments) == 0 {
		return nil, fmt.Errorf("no transcriptions found for video %d", videoID)
	}
	candidates := GenerateClipCandidates(segments)
	return candidates, nil
}
