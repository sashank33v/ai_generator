package processor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type AudioProcessor struct {
}

func NewAudioProcessor() *AudioProcessor {
	return &AudioProcessor{}
}

func (p *AudioProcessor) ExtractAudio(
	ctx context.Context,
	inputPath string,
	outputPath string,
) error {

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create audio directory: %w", err)
	}

	args := []string{
		"-y",
		"-i", inputPath,
		"-vn",
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"failed to extract audio: %w, output: %s",
			err,
			string(output),
		)
	}

	return nil
}
