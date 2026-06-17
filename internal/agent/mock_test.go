package agent

import (
	"context"
	"encoding/json"
	"testing"

	"liki/internal/llm"
)

func TestMockLLM_ChatStream(t *testing.T) {
	m := &MockLLM{
		StreamTokens: []string{"你", "好"},
	}
	ch, err := m.ChatStream(context.Background(), "system", "user")
	if err != nil {
		t.Fatalf("ChatStream: %v", err)
	}
	var tokens []string
	for tok := range ch {
		tokens = append(tokens, tok)
	}
	if len(tokens) != 2 || tokens[0] != "你" || tokens[1] != "好" {
		t.Errorf("got %v, want [你 好]", tokens)
	}
}

func TestMockLLM_ChatStream_Error(t *testing.T) {
	m := &MockLLM{
		StreamTokens: []string{"a"},
		StreamErr:    context.DeadlineExceeded,
	}
	_, err := m.ChatStream(context.Background(), "system", "user")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMockLLM_ChatStreamWithTools_ToolCalls(t *testing.T) {
	m := &MockLLM{
		ToolResps: []*llm.ChatResult{
			{ToolCalls: []llm.ToolCall{{ID: "call_1", Type: "function"}}, Role: "assistant"},
		},
	}
	ch, err := m.ChatStreamWithTools(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("ChatStreamWithTools: %v", err)
	}
	var events []llm.StreamEvent
	for ev := range ch {
		events = append(events, ev)
	}
	if len(events) != 1 || events[0].FinishReason != "tool_calls" {
		t.Errorf("got %d events, want 1 with finish_reason=tool_calls", len(events))
	}
}

func TestMockLLM_ChatStreamWithTools_Content(t *testing.T) {
	m := &MockLLM{
		ToolResps: []*llm.ChatResult{
			{Content: "hello", Role: "assistant"},
		},
	}
	ch, err := m.ChatStreamWithTools(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("ChatStreamWithTools: %v", err)
	}
	ev := <-ch
	if ev.Content != "hello" {
		t.Errorf("Content = %q, want hello", ev.Content)
	}
	ev = <-ch
	if ev.FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want stop", ev.FinishReason)
	}
}

func TestMockLLM_ChatStreamWithTools_Error(t *testing.T) {
	m := &MockLLM{
		ToolErrs: []error{context.DeadlineExceeded},
	}
	_, err := m.ChatStreamWithTools(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMockLLM_ChatStreamWithTools_Exhausted(t *testing.T) {
	m := &MockLLM{} // empty, no ToolResps
	_, err := m.ChatStreamWithTools(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("expected error for exhausted responses")
	}
}

func TestMockToolRegistry_Execute_Fallback(t *testing.T) {
	m := &MockToolRegistry{}
	raw, err := m.Execute(context.Background(), "any_tool", nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if string(raw) != `{"status":"ok"}` {
		t.Errorf("got %s, want {\"status\":\"ok\"}", raw)
	}
}

func TestMockToolRegistry_Execute_Error(t *testing.T) {
	m := &MockToolRegistry{
		Errors: map[string]error{"bad_tool": context.DeadlineExceeded},
	}
	_, err := m.Execute(context.Background(), "bad_tool", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMockToolRegistry_Execute_Result(t *testing.T) {
	m := &MockToolRegistry{
		Results: map[string]json.RawMessage{"tool1": json.RawMessage(`{"x":1}`)},
	}
	raw, err := m.Execute(context.Background(), "tool1", nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if string(raw) != `{"x":1}` {
		t.Errorf("got %s, want {\"x\":1}", raw)
	}
}

func TestEmitIf_NilCallback(t *testing.T) {
	// Should not panic
	emitIf(nil, ChatEvent{Type: EventTextDelta, Data: "hello"})
}

func TestEmitIf_WithCallback(t *testing.T) {
	var received []ChatEvent
	cb := func(ev ChatEvent) { received = append(received, ev) }
	emitIf(cb, ChatEvent{Type: EventTextDelta, Data: "hello"})
	if len(received) != 1 {
		t.Errorf("got %d events, want 1", len(received))
	}
}
