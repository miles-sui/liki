package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

)

const defaultBaseURL = "https://api.deepseek.com"
const defaultModel = "deepseek-v4-pro"

// Client is a DeepSeek LLM client with tool calling support.
type Client struct {
	apiKey       string
	baseURL      string
	model        string
	streamClient *http.Client
}

// ToolDef is a tool definition passed to the LLM.
type ToolDef struct {
	Type     string          `json:"type"`
	Function json.RawMessage `json:"function"`
}

// ChatResult holds the LLM response, which may contain content, tool calls, or both.
type ChatResult struct {
	Role             Role       `json:"role"`
	Content          string            `json:"content"`
	ReasoningContent string            `json:"reasoning_content"`
	ToolCalls        []ToolCall `json:"tool_calls"`
}

// StreamEvent is a streaming chunk that may contain text delta, reasoning, tool calls, or finish.
type StreamEvent struct {
	Content          string            // text delta
	ReasoningContent string            // reasoning delta
	ToolCalls        []ToolCall // accumulated tool calls when FinishReason is set
	FinishReason     string            // "stop" | "tool_calls" | "length" | ""
}

// New creates a new LLM client with the default model (deepseek-v4-pro).
func New(apiKey string) *Client {
	return &Client{
		apiKey:       apiKey,
		baseURL:      defaultBaseURL,
		model:        defaultModel,
		streamClient: &http.Client{
		Timeout: 0, // no request timeout; stream is unbounded
		Transport: &http.Transport{
			DialContext: (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
		},
	},
	}
}

// ChatStream sends a chat request with stream=true and returns a channel of content deltas.
// For report generation (Phase 3) only — tool calling uses ChatStreamWithTools.
func (c *Client) ChatStream(ctx context.Context, systemPrompt, userMessage string) (<-chan string, error) {
	messages := []Message{
		{Role: RoleSystem, Content: systemPrompt},
		{Role: RoleUser, Content: userMessage},
	}
	return c.stream(ctx, messages)
}

// streamChunkWithTools is a streaming chunk that may contain tool call deltas.
type streamChunkWithTools struct {
	Choices []struct {
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
		Delta        struct {
			Role             string                `json:"role"`
			Content          string                `json:"content"`
			ReasoningContent string                `json:"reasoning_content"`
			ToolCalls        []streamToolCallDelta `json:"tool_calls"`
		} `json:"delta"`
	} `json:"choices"`
}

type streamToolCallDelta struct {
	Index    int    `json:"index"`
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// doStreamReq marshals the payload, sends the streaming request, and checks for errors.
func (c *Client) doStreamReq(ctx context.Context, payload any) (*http.Response, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("llm: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("llm: new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm: %w", err)
	}
	if resp.StatusCode >= 300 {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024))
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("llm: %d (failed to read body: %v)", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("llm: %d: %s", resp.StatusCode, string(body))
	}
	return resp, nil
}

// sseScanner reads an SSE stream from body line by line. For each valid data line,
// fn is called with the decoded JSON data. If fn returns false, scanning stops.
// Returns any scanner error encountered after the loop.
func sseScanner(ctx context.Context, body io.ReadCloser, fn func([]byte) bool) error {
	defer body.Close()
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return nil
		}
		if !fn([]byte(data)) {
			return nil
		}
	}
	return scanner.Err()
}

// ChatStreamWithTools sends a streaming chat request with tools and returns a channel
// of StreamEvents. Text deltas and reasoning arrive incrementally; tool calls arrive
// with FinishReason set to "tool_calls" or "stop". Used for Phase 1 parameter collection.
func (c *Client) ChatStreamWithTools(ctx context.Context, messages []Message, tools []ToolDef) (<-chan StreamEvent, error) {
	resp, err := c.doStreamReq(ctx, streamRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
		Tools:    tools,
		Thinking: &thinkingParam{Type: "disabled"}, // disabled for deterministic tool-calling latency; enable for report generation if needed
	})
	if err != nil {
		return nil, err
	}

	ch := make(chan StreamEvent, 64)
	go func() {
		defer close(ch)
		var acc []ToolCall

		err := sseScanner(ctx, resp.Body, func(data []byte) bool {
			var chunk streamChunkWithTools
			if err := json.Unmarshal(data, &chunk); err != nil {
				return true
			}
			if len(chunk.Choices) == 0 {
				return true
			}
			choice := chunk.Choices[0]
			delta := choice.Delta

			for _, tc := range delta.ToolCalls {
				for len(acc) <= tc.Index {
					acc = append(acc, ToolCall{})
				}
				if tc.ID != "" {
					acc[tc.Index].ID = tc.ID
				}
				if tc.Type != "" {
					acc[tc.Index].Type = tc.Type
				}
				if tc.Function.Name != "" {
					acc[tc.Index].Function.Name = tc.Function.Name
				}
				acc[tc.Index].Function.Arguments += tc.Function.Arguments
			}

			ev := StreamEvent{
				Content:          delta.Content,
				ReasoningContent: delta.ReasoningContent,
				FinishReason:     choice.FinishReason,
			}

			if choice.FinishReason == "tool_calls" || choice.FinishReason == "stop" {
				var nonEmpty []ToolCall
				for _, tc := range acc {
					if tc.ID != "" && tc.Function.Name != "" {
						nonEmpty = append(nonEmpty, tc)
					}
				}
				if len(nonEmpty) > 0 {
					ev.ToolCalls = nonEmpty
				}
				acc = acc[:0]
			}

			select {
			case ch <- ev:
			case <-ctx.Done():
				return false
			}

			return choice.FinishReason == ""
		})
		if err != nil {
			slog.Debug("sse scanner ended", "err", err)
		}
	}()

	return ch, nil
}

type thinkingParam struct {
	Type string `json:"type"`
}

type streamRequest struct {
	Model    string          `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool            `json:"stream"`
	Tools    []ToolDef       `json:"tools,omitempty"`
	Thinking *thinkingParam  `json:"thinking,omitempty"`
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func (c *Client) stream(ctx context.Context, messages []Message) (<-chan string, error) {
	resp, err := c.doStreamReq(ctx, streamRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		return nil, err
	}

	ch := make(chan string, 64)
	go func() {
		defer close(ch)

		err := sseScanner(ctx, resp.Body, func(data []byte) bool {
			var chunk streamChunk
			if err := json.Unmarshal(data, &chunk); err != nil {
				return true
			}
			if len(chunk.Choices) == 0 {
				return true
			}
			delta := chunk.Choices[0].Delta
			if delta.Content != "" {
				select {
				case ch <- delta.Content:
				case <-ctx.Done():
					return false
				}
			}
			return true
		})
		if err != nil {
			slog.Warn("llm: stream scan", "err", err)
		}
	}()

	return ch, nil
}
