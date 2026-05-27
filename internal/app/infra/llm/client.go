// Package llm provides an Anthropic Messages API streaming client.
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/25types/25types/internal/app/application/reports"
)

var _ reports.Streamer = (*Client)(nil)

// Client implements reports.Streamer via the Anthropic Messages API.
type Client struct {
	apiKey    string
	baseURL   string
	model     string
	maxTokens int
	timeout   time.Duration
	http      *http.Client
}

// New creates a new Anthropic client from environment variables and config.
func New(model string, maxTokens int, timeout time.Duration) *Client {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	return &Client{
		apiKey:    apiKey,
		baseURL:   baseURL,
		model:     model,
		maxTokens: maxTokens,
		timeout:   timeout,
		http:      &http.Client{Timeout: timeout},
	}
}

// Stream sends a streaming request to the Anthropic Messages API.
func (c *Client) Stream(ctx context.Context, systemPrompt string, messages []reports.Message) (<-chan reports.Chunk, error) {
	if c.apiKey == "" {
		return nil, errors.New("ANTHROPIC_API_KEY is not set")
	}

	apiMessages := make([]map[string]string, len(messages))
	for i, m := range messages {
		apiMessages[i] = map[string]string{"role": m.Role, "content": m.Content}
	}

	body := map[string]any{
		"model":      c.model,
		"max_tokens": c.maxTokens,
		"messages":   apiMessages,
		"stream":     true,
	}
	if systemPrompt != "" {
		body["system"] = systemPrompt
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("llm: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/messages", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("llm: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm: request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		resp.Body.Close()
		return nil, fmt.Errorf("llm: API returned %d: %s", resp.StatusCode, string(body))
	}

	ch := make(chan reports.Chunk, 64)
	go c.readStream(ctx, resp.Body, ch)
	return ch, nil
}

func (c *Client) readStream(ctx context.Context, body io.ReadCloser, ch chan<- reports.Chunk) {
	defer body.Close()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("llm: panic in readStream: %v\n%s", r, debug.Stack())
			select {
			case ch <- reports.Chunk{Error: fmt.Errorf("llm: internal error"), Done: true}:
			default:
			}
		}
		close(ch)
	}()

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var accumulatedText strings.Builder
	startTime := time.Now()

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			ch <- reports.Chunk{Error: ctx.Err(), Done: true}
			return
		default:
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var event struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		switch event.Type {
		case "content_block_delta":
			if event.Delta.Type == "text_delta" && event.Delta.Text != "" {
				accumulatedText.WriteString(event.Delta.Text)
				ch <- reports.Chunk{Text: event.Delta.Text}
			}
		case "message_stop":
			log.Printf("llm stream done model=%s latency_ms=%d chars=%d",
				c.model, time.Since(startTime).Milliseconds(), accumulatedText.Len())
			ch <- reports.Chunk{Done: true}
			return
		case "error":
			ch <- reports.Chunk{Error: fmt.Errorf("llm: stream error: %s", data), Done: true}
			return
		}
	}

	if err := scanner.Err(); err != nil && !errors.Is(err, context.Canceled) {
		ch <- reports.Chunk{Error: fmt.Errorf("llm: stream read: %w", err), Done: true}
	}
}
