package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"generator/internal/model"
	"generator/internal/repository"

	"github.com/google/uuid"
)

type VideoService struct {
	videoRepo  *repository.VideoRepository
	jobService *ProcessingJobService
	storageDir string
}

func NewVideoService(videoRepo *repository.VideoRepository, jobService *ProcessingJobService, storageDir string) *VideoService {
	return &VideoService{
		videoRepo:  videoRepo,
		jobService: jobService,
		storageDir: storageDir,
	}
}

func (s *VideoService) UploadVideo(
	ctx context.Context,
	file multipart.File,
	header *multipart.FileHeader,
) (*model.Video, error) {
	const maxFileSize = 500 << 20 // 500 MB

	if header.Size <= 0 || header.Size > maxFileSize {
		return nil, fmt.Errorf("video size must be between 1 byte and 500 MB")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".mp4" && ext != ".mov" && ext != ".mkv" {
		return nil, fmt.Errorf("unsupported video format: %s", ext)
	}
	buffer := make([]byte, 512)

	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read video file: %w", err)
	}

	mimeType := http.DetectContentType(buffer)

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to reset file position: %w", err)
	}
	if err := os.MkdirAll(s.storageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	fileName := uuid.NewString() + ext
	filePath := filepath.Join(s.storageDir, fileName)

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create video file: %w", err)
	}

	_, copyErr := io.Copy(dst, io.LimitReader(file, maxFileSize+1))
	closeErr := dst.Close()

	if copyErr != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save video: %w", copyErr)
	}

	if closeErr != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to close video file: %w", closeErr)
	}

	video := &model.Video{
		OriginalName: header.Filename,
		FileName:     fileName,
		FilePath:     filePath,
		FileSize:     header.Size,
		MimeType:     mimeType,
		Status:       "PENDING",
	}

	job := &model.ProcessingJob{
		Status: "PENDING",
	}

	err = s.videoRepo.CreateWithJob(ctx, video, job)
	if err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save video and create job: %w", err)
	}

	return video, nil
}
