package service

import (
	"context"
	"fmt"
	"generator/internal/model"
	"generator/internal/processor"
	"generator/internal/repository"
)

type WorkerService struct {
	workerRepo     *repository.WorkerRepository
	videoRepo      *repository.VideoRepository
	audioProcessor *processor.AudioProcessor
}

func NewWorkerService(workerRepo *repository.WorkerRepository, videoRepo *repository.VideoRepository, audioProcessor *processor.AudioProcessor) *WorkerService {
	return &WorkerService{
		workerRepo:     workerRepo,
		videoRepo:      videoRepo,
		audioProcessor: audioProcessor,
	}
}

func (s *WorkerService) ProcessNextJob(
	ctx context.Context,
) (*model.ProcessingJob, error) {

	job, err := s.workerRepo.ClaimNextJob(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to process next job: %w", err)
	}

	if job == nil {
		return nil, nil
	}

	return job, nil
}
func (s *WorkerService) GetVideoByID(
	ctx context.Context,
	videoID int64,
) (*model.Video, error) {

	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video for processing: %w", err)
	}

	return video, nil
}
func (s *WorkerService) ProcessJob(ctx context.Context, job *model.ProcessingJob) error {
	video, err := s.videoRepo.GetByID(ctx, job.VideoID)
	if err != nil {
		return fmt.Errorf("failed to get video: %w", err)
	}
	outputPath := fmt.Sprintf(
		"storage/audio/%d.wav",
		video.ID,
	)
	err = s.audioProcessor.ExtractAudio(ctx, video.FilePath, outputPath)
	if err != nil {
		return fmt.Errorf("audio extraction failed: %w", err)
	}

	return nil
}
func (s *WorkerService) UpdateJobStatus(
	ctx context.Context,
	job *model.ProcessingJob,
	status string,
	errorMessage *string,
) error {

	err := s.workerRepo.UpdateJobStatus(
		ctx,
		job.ID,
		job.VideoID,
		status,
		errorMessage,
	)

	if err != nil {
		return fmt.Errorf("failed to update processing status: %w", err)
	}

	return nil
}
