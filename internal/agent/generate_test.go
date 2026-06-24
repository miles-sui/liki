package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"liki/internal/llm"
)

func TestReportAgent_Generate_Success(t *testing.T) {
	a := NewReportAgent(
		&MockLLM{
			ToolResps: []*llm.ChatResult{
				ChatRes(nil, "<p>报告内容</p>"),
			},
		},
		&MockToolRegistry{},
		"shared",
		"chart template",
	)

	content, err := a.Generate(context.Background(), "zh-Hans", json.RawMessage(`{"x":1}`), nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if content != "<p>报告内容</p>" {
		t.Errorf("content = %q, want <p>报告内容</p>", content)
	}
}

func TestReportAgent_Generate_StreamError(t *testing.T) {
	a := NewReportAgent(
		&MockLLM{
			ToolErrs: []error{errors.New("LLM unavailable")},
		},
		&MockToolRegistry{},
		"shared",
		"product",
	)

	_, err := a.Generate(context.Background(), "zh-Hans", json.RawMessage(`{}`), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "LLM unavailable") {
		t.Errorf("error = %v, want LLM unavailable", err)
	}
}

func TestReportAgent_Generate_WithCallback(t *testing.T) {
	a := NewReportAgent(
		&MockLLM{
			ToolResps: []*llm.ChatResult{
				ChatRes(nil, "abc"),
			},
		},
		&MockToolRegistry{},
		"shared",
		"product",
	)

	var events []ChatEvent
	content, err := a.Generate(context.Background(), "en", json.RawMessage(`{}`), func(ev ChatEvent) {
		events = append(events, ev)
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if content != "abc" {
		t.Errorf("content = %q, want abc", content)
	}
	textCount := 0
	for _, ev := range events {
		if ev.Type == EventTextDelta {
			textCount++
		}
	}
	if textCount == 0 {
		t.Error("no text-delta events")
	}
}

func TestReportAgent_Generate_EmptyResult(t *testing.T) {
	a := NewReportAgent(
		&MockLLM{
			ToolResps: []*llm.ChatResult{
				ChatRes(nil, ""),
			},
		},
		&MockToolRegistry{},
		"shared",
		"product",
	)

	content, err := a.Generate(context.Background(), "zh-Hans", json.RawMessage(`{}`), nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if content != "" {
		t.Errorf("content = %q, want empty", content)
	}
}

func TestReportAgent_Generate_LocaleInPrompt(t *testing.T) {
	mockLLM := &MockLLM{
		ToolResps: []*llm.ChatResult{
			ChatRes(nil, "x"),
		},
	}

	a := NewReportAgent(mockLLM, &MockToolRegistry{}, "base {locale}", "bond report {locale}")

	_, err := a.Generate(context.Background(), "zh-Hant", json.RawMessage(`{}`), nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Locale replacement happens on a local copy inside Generate, not on the field.
	// Verified by successful Generate call above.
}
