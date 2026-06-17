package llm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestChatStream_Tokens(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("no flusher")
		}
		chunks := []string{
			`data: {"choices":[{"delta":{"content":"你"},"index":0}]}`,
			`data: {"choices":[{"delta":{"content":"好"},"index":0}]}`,
			`data: {"choices":[{"delta":{"content":"！"},"index":0}]}`,
			`data: [DONE]`,
		}
		for _, c := range chunks {
			fmt.Fprint(w, c+"\n\n")
			flusher.Flush()
		}
	}))
	defer srv.Close()

	c := &Client{
		apiKey:       "test",
		baseURL:      srv.URL,
		model:        "test-model",
		streamClient: &http.Client{Timeout: 0},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch, err := c.ChatStream(ctx, "system prompt", "user message")
	if err != nil {
		t.Fatalf("ChatStream: %v", err)
	}

	var tokens []string
	for token := range ch {
		tokens = append(tokens, token)
	}

	result := strings.Join(tokens, "")
	if result != "你好！" {
		t.Errorf("got %q, want %q", result, "你好！")
	}
}

func TestChatStream_ContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		// Send one chunk then wait for context cancel
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"first\"},\"index\":0}]}\n\n")
		flusher.Flush()
		// Block until client disconnects
		<-r.Context().Done()
	}))
	defer srv.Close()

	c := &Client{
		apiKey:       "test",
		baseURL:      srv.URL,
		model:        "test-model",
		streamClient: &http.Client{Timeout: 0},
	}

	ctx, cancel := context.WithCancel(context.Background())
	ch, err := c.ChatStream(ctx, "system", "user")
	if err != nil {
		t.Fatalf("ChatStream: %v", err)
	}

	// Read first token
	token := <-ch
	if token != "first" {
		t.Errorf("got %q, want %q", token, "first")
	}

	// Cancel and verify channel closes promptly
	cancel()
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("channel should be closed after cancel")
		}
	case <-time.After(5 * time.Second):
		t.Error("channel did not close within 5s of cancel")
	}
}

func TestParseToolDef_Valid(t *testing.T) {
	raw := []byte(`{"name":"compute_chart","description":"compute bazi chart"}`)
	td, err := ParseToolDef(raw)
	if err != nil {
		t.Fatalf("ParseToolDef: %v", err)
	}
	if td.Type != "function" {
		t.Errorf("Type=%q, want function", td.Type)
	}
	if td.Function == nil {
		t.Fatal("Function is nil")
	}
}

func TestParseToolDef_MissingName(t *testing.T) {
	_, err := ParseToolDef([]byte(`{"description":"no name"}`))
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestParseToolDef_InvalidJSON(t *testing.T) {
	_, err := ParseToolDef([]byte(`{bad json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadTool_NotFound(t *testing.T) {
	_, err := LoadTool("nonexistent_tool")
	if err == nil {
		t.Fatal("expected error for nonexistent tool")
	}
}

func TestChatStream_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer srv.Close()

	c := &Client{
		apiKey:       "bad-key",
		baseURL:      srv.URL,
		model:        "test-model",
		streamClient: &http.Client{Timeout: 0},
	}

	ctx := context.Background()
	_, err := c.ChatStream(ctx, "system", "user")
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestChatStreamWithTools_ToolCalls(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		chunks := []string{
			`data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"compute_chart","arguments":""}}]}}]}`,
			`data: {"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"y"}}]}}]}`,
			`data: {"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"ear\":2024}"}}]}}]}`,
			`data: {"choices":[{"index":0,"finish_reason":"tool_calls"}]}`,
			`data: [DONE]`,
		}
		for _, c := range chunks {
			fmt.Fprint(w, c+"\n\n")
			flusher.Flush()
		}
	}))
	defer srv.Close()

	c := &Client{
		apiKey:       "test",
		baseURL:      srv.URL,
		model:        "test-model",
		streamClient: &http.Client{Timeout: 0},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch, err := c.ChatStreamWithTools(ctx, nil, nil)
	if err != nil {
		t.Fatalf("ChatStreamWithTools: %v", err)
	}

	var events []StreamEvent
	for ev := range ch {
		events = append(events, ev)
	}

	if len(events) == 0 {
		t.Fatal("expected at least 1 event")
	}
	last := events[len(events)-1]
	if last.FinishReason != "tool_calls" {
		t.Errorf("FinishReason = %q, want tool_calls", last.FinishReason)
	}
	if len(last.ToolCalls) != 1 {
		t.Fatalf("got %d tool calls, want 1", len(last.ToolCalls))
	}
	if last.ToolCalls[0].Function.Name != "compute_chart" {
		t.Errorf("tool name = %q, want compute_chart", last.ToolCalls[0].Function.Name)
	}
	if last.ToolCalls[0].Function.Arguments != `{"year":2024}` {
		t.Errorf("tool args = %q, want {\"year\":2024}", last.ToolCalls[0].Function.Arguments)
	}
}

func TestChatStreamWithTools_Content(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		chunks := []string{
			`data: {"choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":""}]}`,
			`data: {"choices":[{"index":0,"delta":{"content":" World"},"finish_reason":"stop"}]}`,
			`data: [DONE]`,
		}
		for _, c := range chunks {
			fmt.Fprint(w, c+"\n\n")
			flusher.Flush()
		}
	}))
	defer srv.Close()

	c := &Client{
		apiKey:       "test",
		baseURL:      srv.URL,
		model:        "test-model",
		streamClient: &http.Client{Timeout: 0},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch, err := c.ChatStreamWithTools(ctx, nil, nil)
	if err != nil {
		t.Fatalf("ChatStreamWithTools: %v", err)
	}

	var contents []string
	for ev := range ch {
		if ev.Content != "" {
			contents = append(contents, ev.Content)
		}
	}

	result := strings.Join(contents, "")
	if result != "Hello World" {
		t.Errorf("got %q, want 'Hello World'", result)
	}
}

func TestChatStreamWithTools_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	c := &Client{
		apiKey:       "test",
		baseURL:      srv.URL,
		model:        "test-model",
		streamClient: &http.Client{Timeout: 0},
	}

	_, err := c.ChatStreamWithTools(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("expected error for 500")
	}
}
