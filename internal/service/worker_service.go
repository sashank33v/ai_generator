package service

import (
	"context"
	"errors"
	"fmt"
	"generator/internal/model"
	"generator/internal/processor"
	"generator/internal/repository"
	"log"
	"os"
	"time"
)

type WorkerService struct {
	workerRepo             *repository.WorkerRepository
	videoRepo              *repository.VideoRepository
	audioProcessor         *processor.AudioProcessor
	transcriptionProcessor *processor.TranscriptionProcessor
	transcriptionRepo      *repository.TranscriptionRepository
	generationService      *ClipGenerationService
}

func NewWorkerService(workerRepo *repository.WorkerRepository, videoRepo *repository.VideoRepository, audioProcessor *processor.AudioProcessor, transcriptionProcessor *processor.TranscriptionProcessor, transcriptionRepo *repository.TranscriptionRepository, generationService *ClipGenerationService) *WorkerService {
	return &WorkerService{
		workerRepo:             workerRepo,
		videoRepo:              videoRepo,
		audioProcessor:         audioProcessor,
		transcriptionProcessor: transcriptionProcessor,
		transcriptionRepo:      transcriptionRepo,
		generationService:      generationService,
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

func (s *WorkerService) ProcessJob(
	ctx context.Context,
	job *model.ProcessingJob,
) (retErr error) {

	video, err := s.videoRepo.GetByID(ctx, job.VideoID)
	if err != nil {
		return fmt.Errorf("failed to get video: %w", err)
	}

	// Step 1: Clean up incomplete artifacts if processing fails.
	defer func() {
		if retErr != nil {
			cleanupCtx, cancel := context.WithTimeout(
				context.Background(),
				10*time.Second,
			)
			defer cancel()

			if err := s.cleanupJobArtifacts(cleanupCtx, video.ID); err != nil {
				log.Printf("Cleanup failed for video %d: %v", video.ID, err)
			}
		}
	}()

	// Step 2: Remove artifacts from previous attempts.
	if err := s.cleanupJobArtifacts(ctx, video.ID); err != nil {
		return fmt.Errorf("failed to reset job artifacts: %w", err)
	}

	// Step 3: Extract audio.
	log.Printf("Extracting audio for video %d", video.ID)

	audioPath := fmt.Sprintf("storage/audio/%d.wav", video.ID)

	if err := s.audioProcessor.ExtractAudio(ctx, video.FilePath, audioPath); err != nil {
		return fmt.Errorf("audio extraction failed: %w", err)
	}

	// Step 4: Update status to TRANSCRIBING.
	if err := s.UpdateJobStatus(ctx, job, "TRANSCRIBING", nil); err != nil {
		return fmt.Errorf("failed to update transcription status: %w", err)
	}

	// Step 5: Generate transcription.
	log.Printf("Transcribing video %d", video.ID)

	result, err := s.transcriptionProcessor.Transcribe(ctx, audioPath)
	if err != nil {
		return fmt.Errorf("transcription failed: %w", err)
	}

	// Step 6: Prepare transcription records.
	transcriptions := make([]model.Transcription, 0, len(result.Transcription))

	for _, segment := range result.Transcription {
		transcriptions = append(
			transcriptions,
			model.Transcription{
				VideoID:   video.ID,
				Language:  result.Result.Language,
				StartTime: segment.Offsets.From,
				EndTime:   segment.Offsets.To,
				Text:      segment.Text,
			},
		)
	}

	// Step 7: Store transcription records.
	if err := s.transcriptionRepo.CreateMany(ctx, transcriptions); err != nil {
		return fmt.Errorf("failed to save transcriptions: %w", err)
	}

	log.Printf("Saved %d transcription segments for video %d",
		len(transcriptions), video.ID)

	// Step 8: Update status to ANALYZING.
	if err := s.UpdateJobStatus(ctx, job, "ANALYZING", nil); err != nil {
		return fmt.Errorf("failed to update analysis status: %w", err)
	}

	log.Printf("Starting AI analysis for video %d", video.ID)

	// Step 9: Generate Reel and update GENERATING status.
	outputPath, bestClip, metadata, err :=
		s.generationService.GenerateBestClip(
			ctx,
			video.ID,
			func() error {
				return s.UpdateJobStatus(ctx, job, "GENERATING", nil)
			},
		)

	if err != nil {
		return fmt.Errorf("reel generation failed: %w", err)
	}

	// Step 10: Log successful generation.
	log.Printf("Reel generated: %s", outputPath)
	log.Printf("AI Score: %d/40", bestClip.Score.Total)
	log.Printf("Title: %s", metadata.Title)
	log.Printf("Caption: %s", metadata.Caption)

	return nil
}

func (s *WorkerService) UpdateJobStatus(ctx context.Context, job *model.ProcessingJob, status string, errorMessage *string) error {

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
func (s *WorkerService) RecoverInterruptedJobs(
	ctx context.Context,
) error {
	return s.workerRepo.RecoverInterruptedJobs(ctx)
}

func (s *WorkerService) RetryJob(
	ctx context.Context,
	jobID int64,
) error {

	if jobID <= 0 {
		return fmt.Errorf("invalid job ID")
	}

	return s.workerRepo.RetryFailedJob(ctx, jobID)
}

func (s *WorkerService) cleanupJobArtifacts(
	ctx context.Context,
	videoID int64,
) error {

	var cleanupErr error

	if err := s.workerRepo.ResetJobArtifacts(ctx, videoID); err != nil {
		cleanupErr = errors.Join(cleanupErr, err)
	}

	paths := []string{
		fmt.Sprintf("storage/audio/%d.wav", videoID),
		fmt.Sprintf("storage/audio/%d.wav.transcription.json", videoID),
		fmt.Sprintf("storage/outputs/video_%d_best.mp4", videoID),
		fmt.Sprintf("storage/outputs/video_%d_vertical.mp4", videoID),
		fmt.Sprintf("storage/outputs/video_%d_vertical.mp4.partial.mp4", videoID),
	}

	for _, path := range paths {
		err := os.Remove(path)

		if err != nil && !os.IsNotExist(err) {
			cleanupErr = errors.Join(
				cleanupErr,
				fmt.Errorf("failed to remove %s: %w", path, err),
			)
		}
	}

	return cleanupErr
}
