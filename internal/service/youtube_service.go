package service

import (
	"context"
	"errors"
	"fmt"
	"generator/internal/downloader"
	"generator/internal/model"
	"generator/internal/repository"
	"log"
	"os"
)

var ErrYouTubeDownload = errors.New("Youtube download failed")

type YouTubeService struct {
	videoRepo  *repository.VideoRepository
	downloader *downloader.YouTubeDownloader
}

func NewYoutubeService(videoRepo *repository.VideoRepository, downloader *downloader.YouTubeDownloader) *YouTubeService {
	return &YouTubeService{videoRepo: videoRepo, downloader: downloader}
}
func (s *YouTubeService) ProcessYouTubeURL(ctx context.Context, rawURL string) (*model.Video, *model.ProcessingJob, error) {

	// Step 1: Validate YouTube URL.
	videoURL, err := downloader.NormalizeYouTubeURL(rawURL)
	if err != nil {
		return nil, nil, err
	}

	// Step 2: Download video.
	log.Printf("Downloading YouTube video: %s", videoURL)

	result, err := s.downloader.Download(ctx, videoURL)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrYouTubeDownload, err)
	}

	// Step 3: Prepare video record.
	video := &model.Video{
		OriginalName: result.OriginalName,
		FileName:     result.FileName,
		FilePath:     result.FilePath,
		FileSize:     result.FileSize,
		MimeType:     result.MimeType,
		Status:       "PENDING",
	}

	// Step 4: Prepare processing job.
	job := &model.ProcessingJob{
		Status: "PENDING",
	}

	// Step 5: Save video and job in one transaction.
	if err := s.videoRepo.CreateWithJob(ctx, video, job); err != nil {

		if cleanupErr := os.Remove(result.FilePath); cleanupErr != nil {
			log.Printf("Downloaded file cleanup failed: %v", cleanupErr)
		}

		return nil, nil, fmt.Errorf("failed to create video and job: %w", err)
	}

	log.Printf("YouTube video saved: video_id=%d, job_id=%d", video.ID, job.ID)

	return video, job, nil
}
