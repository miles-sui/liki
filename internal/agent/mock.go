package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"liki/internal/llm"
)

// MockLLM implements LLMClient with predefined response sequences.
type MockLLM struct {
	ToolResps    []*llm.ChatResult
	ToolErrs     []error
	toolIdx      int
	StreamTokens []string
	StreamErr    error
}

// ChatStreamWithTools converts the next ChatResult in ToolResps into StreamEvents.
func (m *MockLLM) ChatStreamWithTools(ctx context.Context, messages []llm.Message, tools []llm.ToolDef) (<-chan llm.StreamEvent, error) {
	if m.toolIdx < len(m.ToolErrs) && m.ToolErrs[m.toolIdx] != nil {
		err := m.ToolErrs[m.toolIdx]
		m.toolIdx++
		return nil, err
	}
	if m.toolIdx >= len(m.ToolResps) {
		return nil, fmt.Errorf("MockLLM: no ChatStreamWithTools response at index %d", m.toolIdx)
	}
	r := m.ToolResps[m.toolIdx]
	m.toolIdx++

	var events []llm.StreamEvent
	if r.Content != "" {
		events = append(events, llm.StreamEvent{Content: r.Content})
	}
	if len(r.ToolCalls) > 0 {
		events = append(events, llm.StreamEvent{
			ToolCalls:    r.ToolCalls,
			FinishReason: "tool_calls",
		})
	} else {
		events = append(events, llm.StreamEvent{FinishReason: "stop"})
	}

	ch := make(chan llm.StreamEvent, len(events)+1)
	go func() {
		defer close(ch)
		for _, ev := range events {
			select {
			case <-ctx.Done():
				return
			case ch <- ev:
			}
		}
	}()
	return ch, nil
}

// ChatStream returns a channel of streaming tokens for the given prompt.
func (m *MockLLM) ChatStream(ctx context.Context, systemPrompt, userMessage string) (<-chan string, error) {
	if m.StreamErr != nil {
		return nil, m.StreamErr
	}
	ch := make(chan string, len(m.StreamTokens)+1)
	go func() {
		defer close(ch)
		for _, tok := range m.StreamTokens {
			select {
			case <-ctx.Done():
				return
			case ch <- tok:
			}
		}
	}()
	return ch, nil
}

// ToolCall builds a tool call with the given name and JSON arguments.
func ToolCall(name, args string) llm.ToolCall {
	return llm.ToolCall{
		ID:   fmt.Sprintf("call_%s", name),
		Type: "function",
		Function: llm.FunctionCall{
			Name:      name,
			Arguments: args,
		},
	}
}

// ChatRes builds a ChatResult with tool calls and optional content.
func ChatRes(toolCalls []llm.ToolCall, content string) *llm.ChatResult {
	return &llm.ChatResult{
		Role:      llm.RoleAssistant,
		Content:   content,
		ToolCalls: toolCalls,
	}
}

// MockToolRegistry implements ToolRegistry for tests.
type MockToolRegistry struct {
	Results map[string]json.RawMessage
	Errors  map[string]error
	Defs    []llm.ToolDef
}

// Execute runs the named tool with the given JSON arguments.
func (m *MockToolRegistry) Execute(ctx context.Context, name string, args json.RawMessage) (json.RawMessage, error) {
	if err, ok := m.Errors[name]; ok {
		return nil, err
	}
	if r, ok := m.Results[name]; ok {
		return r, nil
	}
	return json.RawMessage(`{"status":"ok"}`), nil
}

// Schemas returns the registered tool definitions for the LLM.
func (m *MockToolRegistry) Schemas() []llm.ToolDef {
	return m.Defs
}
