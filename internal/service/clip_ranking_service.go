package service

import (
	"context"
	"fmt"
	"log"
	"sort"

	"generator/internal/model"
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

	rankedClips := make([]model.RankedClip, 0, len(candidates))

	for i, candidate := range candidates {

		log.Printf(
			"Scoring candidate %d/%d for video %d",
			i+1, len(candidates), videoID,
		)

		score, err := s.scoringService.ScoreCandidate(ctx, candidate)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to score candidate %d: %w",
				i+1, err,
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
