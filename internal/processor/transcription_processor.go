package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type WhisperResponse struct {
	Result struct {
		Language string `json:"Language"`
	} `json:"result"`

	Transcription []WhisperSegment `json:"transcription"`
}
type WhisperSegment struct {
	Timestamps struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"timestamps"`

	Offsets struct {
		From int64 `json:"from"`
		To   int64 `json:"to"`
	} `json:"offsets"`

	Text string `json:"text"`
}
type TranscriptionProcessor struct {
	binaryPath string
	modelPath  string
	threads    string
}

func NewTranscriptionProcessor() *TranscriptionProcessor {
	return &TranscriptionProcessor{
		binaryPath: "./whisper.cpp/build/bin/whisper-cli",
		modelPath:  "./whisper.cpp/models/ggml-medium.bin",
		threads:    "4",
	}
}

func (p *TranscriptionProcessor) Transcribe(ctx context.Context, audioPath string) (*WhisperResponse, error) {
	outputPath := audioPath + ".transcription"

	args := []string{
		"-m", p.modelPath,
		"-f", audioPath,
		"-l", "auto",
		"-t", p.threads,
		"-oj",
		"-of", outputPath,
	}

	cmd := exec.CommandContext(ctx, p.binaryPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Whisper transcription failed: %w, output: %s", err, string(output))
	}
	jsonPath := outputPath + ".json"

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read transcription JSON: %w", err)
	}
	var response WhisperResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse transcription JSON: %w", err)
	}
	if len(response.Transcription) == 0 {
		return nil, fmt.Errorf("whisper returned no transcription segments")
	}
	return &response, nil
}

var _ = filepath.Base
