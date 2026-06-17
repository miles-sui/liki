package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

)

func TestGenerateFromData_Success(t *testing.T) {
	a := &ChatAgent{
		llm: &MockLLM{
			StreamTokens: []string{"<p>", "报告", "内容", "</p>"},
		},
		prompt: "default prompt {locale}",
		ReportPrompts: map[Product]string{
			ProductChart: "chart report {locale}",
		},
	}

	content, err := a.GenerateFromData(context.Background(), "zh-Hans", ProductChart, json.RawMessage(`{"x":1}`), nil)
	if err != nil {
		t.Fatalf("GenerateFromData: %v", err)
	}
	if content != "<p>报告内容</p>" {
		t.Errorf("content = %q, want <p>报告内容</p>", content)
	}
}

func TestGenerateFromData_FallbackPrompt(t *testing.T) {
	// When no report prompt is configured for the product, fall back to the default prompt.
	a := &ChatAgent{
		llm: &MockLLM{
			StreamTokens: []string{"ok"},
		},
		prompt:        "default {locale}",
		ReportPrompts: nil,
	}

	content, err := a.GenerateFromData(context.Background(), "en", ProductChart, json.RawMessage(`{}`), nil)
	if err != nil {
		t.Fatalf("GenerateFromData: %v", err)
	}
	if content != "ok" {
		t.Errorf("content = %q, want ok", content)
	}
}

func TestGenerateFromData_StreamError(t *testing.T) {
	a := &ChatAgent{
		llm: &MockLLM{
			StreamErr: errors.New("LLM unavailable"),
		},
		prompt: "prompt {locale}",
	}

	_, err := a.GenerateFromData(context.Background(), "zh-Hans", ProductChart, json.RawMessage(`{}`), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "LLM unavailable") {
		t.Errorf("error = %v, want LLM unavailable", err)
	}
}

func TestGenerateFromData_WithCallback(t *testing.T) {
	a := &ChatAgent{
		llm: &MockLLM{
			StreamTokens: []string{"a", "b", "c"},
		},
		prompt: "prompt {locale}",
	}

	var events []ChatEvent
	content, err := a.GenerateFromData(context.Background(), "en", ProductChart, json.RawMessage(`{}`), func(ev ChatEvent) {
		events = append(events, ev)
	})
	if err != nil {
		t.Fatalf("GenerateFromData: %v", err)
	}
	if content != "abc" {
		t.Errorf("content = %q, want abc", content)
	}
	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}
	for i, ev := range events {
		if ev.Type != EventTextDelta {
			t.Errorf("event[%d].Type = %q, want text_delta", i, ev.Type)
		}
	}
	if events[0].Content != "a" || events[1].Content != "b" || events[2].Content != "c" {
		t.Errorf("events = %v, want [a b c]", events)
	}
}

func TestGenerateFromData_EmptyTokens(t *testing.T) {
	a := &ChatAgent{
		llm: &MockLLM{
			StreamTokens: nil, // empty
		},
		prompt: "prompt {locale}",
	}

	content, err := a.GenerateFromData(context.Background(), "zh-Hans", ProductChart, json.RawMessage(`{}`), nil)
	if err != nil {
		t.Fatalf("GenerateFromData: %v", err)
	}
	if content != "" {
		t.Errorf("content = %q, want empty", content)
	}
}

func TestGenerateFromData_LocaleReplacement(t *testing.T) {
	var capturedPrompt string
	a := &ChatAgent{
		llm: &MockLLM{
			StreamTokens: []string{"x"},
			// Override ChatStream to capture the prompt.
		},
		prompt: "base {locale}",
		ReportPrompts: map[Product]string{
			ProductBond: "bond report {locale}",
		},
	}
	// Use a custom mock that captures the system prompt.
	a.llm = &promptCapturingLLM{
		MockLLM:      &MockLLM{StreamTokens: []string{"x"}},
		capturedPrompt: &capturedPrompt,
	}

	_, err := a.GenerateFromData(context.Background(), "zh-Hant", ProductBond, json.RawMessage(`{}`), nil)
	if err != nil {
		t.Fatalf("GenerateFromData: %v", err)
	}
	if !strings.Contains(capturedPrompt, "zh-Hant") {
		t.Errorf("prompt = %q, should contain zh-Hant", capturedPrompt)
	}
	if !strings.Contains(capturedPrompt, "bond report") {
		t.Errorf("prompt = %q, should contain bond report (product-specific)", capturedPrompt)
	}
}

// promptCapturingLLM wraps MockLLM and captures the system prompt.
type promptCapturingLLM struct {
	*MockLLM
	capturedPrompt *string
}

func (m *promptCapturingLLM) ChatStream(ctx context.Context, systemPrompt, userMessage string) (<-chan string, error) {
	*m.capturedPrompt = systemPrompt
	return m.MockLLM.ChatStream(ctx, systemPrompt, userMessage)
}
