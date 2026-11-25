package service

import (
	"context"
	"fmt"
	"log"
	"sort"

	"generator/internal/model"
)

const (
	maxRankingCandidates = 6
	overlapThreshold     = 0.80
)

type ClipRankingService struct {
	candidateService *ClipCandidateService
	scoringService   *ClipScoringService
}

func NewClipRankingService(
	candidateService *ClipCandidateService,
	scoringService *ClipScoringService,
) *ClipRankingService {
	return &ClipRankingService{
		candidateService: candidateService,
		scoringService:   scoringService,
	}
}

// Calculate temporal Intersection over Union (IoU).
func calculateOverlap(a, b model.ClipCandidate) float64 {

	intersectionStart := max(a.StartTime, b.StartTime)
	intersectionEnd := min(a.Endtime, b.Endtime)

	intersection := intersectionEnd - intersectionStart

	if intersection <= 0 {
		return 0
	}

	unionStart := min(a.StartTime, b.StartTime)
	unionEnd := max(a.Endtime, b.Endtime)

	union := unionEnd - unionStart

	return float64(intersection) / float64(union)
}

// Remove near-duplicate candidates and limit AI requests.
func selectDiverseCandidates(
	candidates []model.ClipCandidate,
	limit int,
) []model.ClipCandidate {

	if limit <= 0 || len(candidates) == 0 {
		return nil
	}

	ordered := append([]model.ClipCandidate(nil), candidates...)

	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].StartTime < ordered[j].StartTime
	})

	unique := make([]model.ClipCandidate, 0, len(ordered))

	for _, candidate := range ordered {

		duplicate := false

		for _, selected := range unique {

			if calculateOverlap(candidate, selected) >= overlapThreshold {
				duplicate = true
				break
			}
		}

		if !duplicate {
			unique = append(unique, candidate)
		}
	}

	if len(unique) <= limit {
		return unique
	}

	if limit == 1 {
		return unique[:1]
	}

	// Select candidates across the entire video.
	selected := make([]model.ClipCandidate, 0, limit)

	for i := 0; i < limit; i++ {

		index := i * (len(unique) - 1) / (limit - 1)

		selected = append(selected, unique[index])
	}

	return selected
}

func (s *ClipRankingService) RankCandidates(
	ctx context.Context,
	videoID int64,
) ([]model.RankedClip, error) {

	candidates, err := s.candidateService.GetCandidates(ctx, videoID)

	if err != nil {
		return nil, fmt.Errorf("failed to get candidates: %w", err)
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no valid clip candidates found")
	}

	originalCount := len(candidates)

	candidates = selectDiverseCandidates(
		candidates,
		maxRankingCandidates,
	)

	log.Printf(
		"Candidate optimization: %d -> %d for video %d",
		originalCount,
		len(candidates),
		videoID,
	)

	rankedClips := make([]model.RankedClip, 0, len(candidates))

	for i, candidate := range candidates {

		log.Printf(
			"Scoring candidate %d/%d for video %d",
			i+1,
			len(candidates),
			videoID,
		)

		score, err := s.scoringService.ScoreCandidate(ctx, candidate)

		if err != nil {
			return nil, fmt.Errorf(
				"failed to score candidate %d: %w",
				i+1,
				err,
			)
		}

		rankedClips = append(rankedClips, model.RankedClip{
			Candidate: candidate,
			Score:     *score,
		})
	}

	sort.SliceStable(rankedClips, func(i, j int) bool {
		return rankedClips[i].Score.Total >
			rankedClips[j].Score.Total
	})

	return rankedClips, nil
}
