package processor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type VideoProcessor struct{}

func NewVideoProcessor() *VideoProcessor {
	return &VideoProcessor{}
}

func (p *VideoProcessor) GenerateClip(
	ctx context.Context,
	inputPath string,
	outputPath string,
	startTime int64,
	endTime int64,
) error {

	if startTime < 0 || endTime <= startTime {
		return fmt.Errorf("invalid clip timestamps")
	}

	// Convert milliseconds to seconds.
	start := float64(startTime) / 1000
	duration := float64(endTime-startTime) / 1000

	// Create output directory.
	if err := os.MkdirAll(
		filepath.Dir(outputPath), 0755,
	); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	args := []string{
		"-y",
		"-i", inputPath,
		"-ss", fmt.Sprintf("%.3f", start),
		"-t", fmt.Sprintf("%.3f", duration),
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-crf", "23",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", "128k",
		"-movflags", "+faststart",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"failed to generate clip: %w, ffmpeg: %s",
			err,
			string(output),
		)
	}

	return nil
}

func (p *VideoProcessor) ConvertToVertical(ctx context.Context, inputPath string, outputPath string) error {

	if inputPath == outputPath {
		return fmt.Errorf("input and output paths cannot be same")
	}

	if err := os.MkdirAll(
		filepath.Dir(outputPath), 0755,
	); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filter := "[0:v]split=2[bg][fg];" +
		"[bg]scale=720:1280:force_original_aspect_ratio=increase," +
		"crop=720:1280,boxblur=20:10[blur];" +
		"[fg]scale=720:1280:force_original_aspect_ratio=decrease[sharp];" +
		"[blur][sharp]overlay=(W-w)/2:(H-h)/2,format=yuv420p[v]"

	args := []string{
		"-y",
		"-i", inputPath,
		"-filter_complex", filter,
		"-map", "[v]",
		"-map", "0:a?",
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		"-movflags", "+faststart",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"vertical conversion failed: %w, ffmpeg: %s",
			err,
			string(output),
		)
	}

	return nil
}

func (p *VideoProcessor) GenerateVerticalClip(
	ctx context.Context,
	inputPath string,
	outputPath string,
	startTime int64,
	endTime int64,
) error {

	// Validate timestamps.
	if startTime < 0 || endTime <= startTime {
		return fmt.Errorf("invalid clip timestamps")
	}

	if filepath.Clean(inputPath) == filepath.Clean(outputPath) {
		return fmt.Errorf("input and output paths must differ")
	}

	// Convert milliseconds into seconds.
	startSeconds := float64(startTime) / 1000
	duration := float64(endTime-startTime) / 1000

	// Create output directory.
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Temporary output prevents publishing incomplete videos.
	tempPath := outputPath + ".partial.mp4"
	defer os.Remove(tempPath)

	// Create blurred background and sharp foreground.
	filter := "[0:v]split=2[bg][fg];" +
		"[bg]scale=720:1280:force_original_aspect_ratio=increase," +
		"crop=720:1280,boxblur=20:10,setsar=1[blur];" +
		"[fg]scale=720:1280:force_original_aspect_ratio=decrease," +
		"setsar=1[sharp];" +
		"[blur][sharp]overlay=(W-w)/2:(H-h)/2," +
		"setsar=1,format=yuv420p[v]"

	args := []string{
		"-y",

		"-ss", fmt.Sprintf("%.3f", startSeconds),
		"-i", inputPath,

		"-t", fmt.Sprintf("%.3f", duration),

		"-filter_complex", filter,

		"-map", "[v]",
		"-map", "0:a:0?",

		"-c:v", "libx264",
		"-preset", "veryfast",
		"-crf", "23",

		"-c:a", "aac",
		"-b:a", "128k",

		"-movflags", "+faststart",

		tempPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"vertical clip generation failed: %w: %s",
			err,
			string(output),
		)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	// Publish the completed video.
	if err := os.Rename(tempPath, outputPath); err != nil {
		return fmt.Errorf("failed to publish generated clip: %w", err)
	}

	return nil
}
