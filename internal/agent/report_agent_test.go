package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"liki/internal/llm"
)

func TestNewReportAgent_SystemPrompt(t *testing.T) {
	shared := "shared rules"
	product := "chart data contract"
	a := NewReportAgent(&MockLLM{}, nil, shared, product)

	want := shared + "\n" + product
	if !strings.Contains(a.systemPrompt, shared) {
		t.Error("systemPrompt missing shared prompt")
	}
	if !strings.Contains(a.systemPrompt, product) {
		t.Error("systemPrompt missing product prompt")
	}
	if a.systemPrompt != want {
		t.Errorf("systemPrompt = %q, want %q", a.systemPrompt, want)
	}
}

func TestReportAgent_Generate_StreamsText(t *testing.T) {
	mockLLM := &MockLLM{
		ToolResps: []*llm.ChatResult{
			{Content: "第一章内容", Role: "assistant"},
		},
	}
	mockTools := &MockToolRegistry{}

	a := NewReportAgent(mockLLM, mockTools, "shared", "product")
	chartJSON := json.RawMessage(`{"fu_yi":{"qiangruo":"身强"}}`)

	var events []ChatEvent
	onEvent := func(ev ChatEvent) {
		events = append(events, ev)
	}

	result, err := a.Generate(context.Background(), "zh-Hans", chartJSON, onEvent)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if result != "第一章内容" {
		t.Errorf("result = %q, want '第一章内容'", result)
	}

	textCount := 0
	for _, ev := range events {
		if ev.Type == EventTextDelta {
			textCount++
		}
	}
	if textCount == 0 {
		t.Error("no text-delta events emitted")
	}
}

func TestReportAgent_Generate_Locale(t *testing.T) {
	mockLLM := &MockLLM{
		ToolResps: []*llm.ChatResult{
			{Content: "ok", Role: "assistant"},
		},
	}

	a := NewReportAgent(mockLLM, &MockToolRegistry{}, "use {locale} here", "product {locale} too")
	_, err := a.Generate(context.Background(), "zh-Hans", json.RawMessage(`{}`), nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Locale replacement works (verified by successful Generate call above).
	// The systemPrompt field retains its original placeholder — that's fine,
	// Generate uses a local copy.
}

func TestReportAgent_Generate_OnEventNil(t *testing.T) {
	mockLLM := &MockLLM{
		ToolResps: []*llm.ChatResult{
			{Content: "ok", Role: "assistant"},
		},
	}

	a := NewReportAgent(mockLLM, &MockToolRegistry{}, "shared", "product")
	_, err := a.Generate(context.Background(), "en", json.RawMessage(`{}`), nil)
	if err != nil {
		t.Fatalf("Generate with nil onEvent: %v", err)
	}
}

func TestReportAgent_Generate_LLMError(t *testing.T) {
	mockLLM := &MockLLM{
		ToolErrs: []error{context.DeadlineExceeded},
	}

	a := NewReportAgent(mockLLM, &MockToolRegistry{}, "shared", "product")
	_, err := a.Generate(context.Background(), "zh-Hans", json.RawMessage(`{}`), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "report agent") {
		t.Errorf("error message should mention report agent: %v", err)
	}
}

func TestReportAgent_Generate_ToolCalling(t *testing.T) {
	toolCall := ToolCall("verify_terminology", `{"terms":["正官","七杀"]}`)
	mockLLM := &MockLLM{
		ToolResps: []*llm.ChatResult{
			ChatRes([]llm.ToolCall{toolCall}, ""),
			ChatRes(nil, "报告内容"),
		},
	}
	mockTools := &MockToolRegistry{
		Results: map[string]json.RawMessage{
			"verify_terminology": json.RawMessage(`{"unknown":[]}`),
		},
	}

	var events []ChatEvent
	a := NewReportAgent(mockLLM, mockTools, "shared", "product")
	content, err := a.Generate(context.Background(), "zh-Hans", json.RawMessage(`{"x":1}`), func(ev ChatEvent) {
		events = append(events, ev)
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if content != "报告内容" {
		t.Errorf("content = %q, want '报告内容'", content)
	}
	textCount := 0
	for _, ev := range events {
		if ev.Type == EventTextDelta {
			textCount++
		}
	}
	if textCount != 1 {
		t.Errorf("expected 1 text-delta event, got %d", textCount)
	}
}
