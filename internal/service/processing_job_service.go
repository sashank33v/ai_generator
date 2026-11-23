package service

import (
	"context"
	"fmt"

	"generator/internal/model"
	"generator/internal/repository"
)

type ProcessingJobService struct {
	jobRepo *repository.ProcessingJobRepository
}

func NewProcessingJobService(
	jobRepo *repository.ProcessingJobRepository,
) *ProcessingJobService {
	return &ProcessingJobService{
		jobRepo: jobRepo,
	}
}

func (s *ProcessingJobService) CreateJob(
	ctx context.Context,
	videoID int64,
) (*model.ProcessingJob, error) {
	job := &model.ProcessingJob{
		VideoID: videoID,
		Status:  "PENDING",
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create processing job: %w", err)
	}

	return job, nil
}
func (s *ProcessingJobService) GetJobByID(
	ctx context.Context,
	id int64,
) (*model.ProcessingJob, error) {

	job, err := s.jobRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get processing job: %w", err)
	}

	return job, nil
}
