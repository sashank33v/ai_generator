package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type LlamaClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewLlamaClient() *LlamaClient {
	return &LlamaClient{
		baseURL: "http://127.0.0.1:8081",
		httpClient: &http.Client{
			Timeout: 200 * time.Second,
		},
	}
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
	Stream      bool          `json:"stream"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

func (c *LlamaClient) Ask(ctx context.Context, prompt string) (string, error) {

	payload := ChatRequest{
		Model: "Qwen3-1.7B-Q4_K_M.gguf",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: prompt + "\n/no_think",
			},
		},
		Temperature: 0.2,
		MaxTokens:   200,
		Stream:      false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/v1/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("AI request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf(
			"AI server returned %d: %s",
			resp.StatusCode,
			string(errBody),
		)
	}

	var result ChatResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode AI response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("AI returned no choices")
	}

	answer := strings.TrimSpace(result.Choices[0].Message.Content)
	if answer == "" {
		return "", fmt.Errorf("AI returned empty content")
	}

	return answer, nil
}
