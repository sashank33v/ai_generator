package processor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExtractAudio(t *testing.T) {
	inputPath := filepath.Join("..", "..", "sample.mp4")
	outputPath := filepath.Join(t.TempDir(), "sample.wav")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	audioProcessor := NewAudioProcessor()

	err := audioProcessor.ExtractAudio(ctx, inputPath, outputPath)
	if err != nil {
		t.Fatalf("expected audio extraction to succeed, got: %v", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("expected output audio file to exist: %v", err)
	}

	if info.Size() == 0 {
		t.Fatal("expected audio file to contain data")
	}
}
