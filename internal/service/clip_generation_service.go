package service

import (
	"context"
	"fmt"
	"os"

	"generator/internal/model"
	"generator/internal/processor"
	"generator/internal/repository"

	"github.com/jackc/pgx/v5"
)

type ClipGenerationService struct {
	rankingService    *ClipRankingService
	videoRepo         *repository.VideoRepository
	videoProcessor    *processor.VideoProcessor
	metadataService   *ClipMetadataService
	generatedClipRepo *repository.GeneratedClipRepository
}

func NewClipGenerationService(rankingService *ClipRankingService, videoRepo *repository.VideoRepository, videoProcessor *processor.VideoProcessor, metadataService *ClipMetadataService, generatedClipRepo *repository.GeneratedClipRepository) *ClipGenerationService {
	return &ClipGenerationService{
		rankingService:    rankingService,
		videoRepo:         videoRepo,
		videoProcessor:    videoProcessor,
		metadataService:   metadataService,
		generatedClipRepo: generatedClipRepo,
	}
}

func (s *ClipGenerationService) GenerateBestClip(ctx context.Context, videoID int64, onGenerating func() error) (string, model.RankedClip, *model.ClipMetadata, error) {

	var empty model.RankedClip

	if videoID <= 0 {
		return "", empty, nil, fmt.Errorf("invalid video ID")
	}

	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return "", empty, nil, fmt.Errorf("failed to get video: %w", err)
	}
	if video.Status != "ANALYZING" {
		return "", empty, nil, fmt.Errorf("video is not in analyzing state")
	}
	rankings, err := s.rankingService.RankCandidates(ctx, videoID)
	if err != nil {
		return "", empty, nil, fmt.Errorf("failed to rank clips: %w", err)
	}

	if len(rankings) == 0 {
		return "", empty, nil, fmt.Errorf("no ranked clips found")
	}

	bestClip := rankings[0]

	metadata, err := s.metadataService.GenerateMetadata(ctx, bestClip)
	if err != nil {
		return "", empty, nil, fmt.Errorf("metadata generation failed: %w", err)
	}
	if onGenerating != nil {
		if err := onGenerating(); err != nil {
			return "", empty, nil, fmt.Errorf(
				"failed to update generating status: %w", err,
			)
		}
	}
	landscapePath := fmt.Sprintf("storage/outputs/video_%d_best.mp4", videoID)

	verticalPath := fmt.Sprintf("storage/outputs/video_%d_vertical.mp4", videoID)

	err = s.videoProcessor.GenerateClip(
		ctx,
		video.FilePath,
		landscapePath,
		bestClip.Candidate.StartTime,
		bestClip.Candidate.Endtime,
	)
	if err != nil {
		return "", empty, nil, fmt.Errorf("clip generation failed: %w", err)
	}

	err = s.videoProcessor.ConvertToVertical(
		ctx,
		landscapePath,
		verticalPath,
	)
	if err != nil {
		return "", empty, nil, fmt.Errorf("vertical conversion failed: %w", err)
	}
	generatedClip := &model.GeneratedClip{
		VideoID:   videoID,
		StartTime: bestClip.Candidate.StartTime,
		EndTime:   bestClip.Candidate.Endtime,
		AIScore:   bestClip.Score.Total,
		Title:     metadata.Title,
		Caption:   metadata.Caption,
		FilePath:  verticalPath,
	}

	if err := s.generatedClipRepo.Save(ctx, generatedClip); err != nil {
		return "", empty, nil, fmt.Errorf(
			"failed to save generated clip: %w", err,
		)
	}

	return verticalPath, bestClip, metadata, nil
}

func (s *ClipGenerationService) GetGeneratedClipPath(ctx context.Context, videoID int64) (string, error) {
	clip, err := s.GetGeneratedClip(ctx, videoID)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(clip.FilePath)
	if err != nil {
		return "", fmt.Errorf("generated file unavailable: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("invalid generated file")
	}
	return clip.FilePath, nil
}

func (s *ClipGenerationService) GetGeneratedClip(ctx context.Context, videoID int64) (*model.GeneratedClip, error) {

	if videoID <= 0 {
		return nil, fmt.Errorf("invalid video ID")
	}

	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, err
	}

	if video.Status != "COMPLETED" {
		return nil, fmt.Errorf("clip is not ready: %w", pgx.ErrNoRows)
	}

	return s.generatedClipRepo.GetByVideoID(ctx, videoID)
}
