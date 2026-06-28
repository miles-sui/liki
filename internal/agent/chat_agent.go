package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"liki/internal/llm"
)

func emitIf(fn func(ChatEvent), ev ChatEvent) {
	if fn != nil {
		fn(ev)
	}
}

// ChatEventType classifies ChatEvent kinds for client-side routing.
type ChatEventType string

// ChatEvent type constants for client-side SSE event routing.
const (
	EventTextDelta     ChatEventType = "text-delta"
	EventPhase         ChatEventType = "phase"
	EventError         ChatEventType = "error"
	EventThinking      ChatEventType = "thinking"
	EventThinkingDelta ChatEventType = "thinking-delta"
	EventReportReady   ChatEventType = "report-ready"
)

// ChatEvent is emitted during streaming report generation and sent as SSE.
type ChatEvent struct {
	Type    ChatEventType `json:"type"`
	Content string        `json:"content,omitempty"`
	Data    any           `json:"data,omitempty"`
}

// IsNamingReport returns true if the assistant content contains a naming report heading.
// Matches lines starting with "# " (h1) followed by "起名报告" (Simplified or Traditional).
func IsNamingReport(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "# ") {
			continue
		}
		if strings.Contains(line, "起名报告") || strings.Contains(line, "起名報吿") {
			return true
		}
	}
	return false
}

const (
	chatRoundTimeout = 120 * time.Second
	maxChatRounds    = 20
)

// NamingChat runs the naming chat pipeline with tool-calling.
// onEvent receives SSE events for real-time client feedback; may be nil.
func (a *ChatAgent) NamingChat(ctx context.Context, locale string, messages []llm.Message, onEvent func(ChatEvent)) ([]llm.Message, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("agent: naming chat: no messages")
	}
	emitIf(onEvent, ChatEvent{Type: EventThinking})

	msgs := a.ensureNamingPrompt(locale, messages)
	tools := a.tools.Schemas()

	for round := 0; round < maxChatRounds; round++ {
		roundCtx, roundCancel := context.WithTimeout(ctx, chatRoundTimeout)

		streamCh, err := a.llm.ChatStreamWithTools(roundCtx, msgs, tools)
		if err != nil {
			roundCancel()
			return nil, fmt.Errorf("agent: naming round %d: %w", round, err)
		}

		var contentBuf strings.Builder
		var reasoningBuf strings.Builder
		var finalToolCalls []llm.ToolCall

		for ev := range streamCh {
			if ev.Content != "" {
				contentBuf.WriteString(ev.Content)
				emitIf(onEvent, ChatEvent{Type: EventTextDelta, Content: ev.Content})
			}
			if ev.ReasoningContent != "" {
				reasoningBuf.WriteString(ev.ReasoningContent)
				emitIf(onEvent, ChatEvent{Type: EventThinkingDelta, Content: ev.ReasoningContent})
			}
			if ev.ToolCalls != nil {
				finalToolCalls = ev.ToolCalls
			}
		}

		assistantContent := contentBuf.String()
		if assistantContent != "" || len(finalToolCalls) > 0 {
			msgs = append(msgs, llm.Message{
				Role:             llm.RoleAssistant,
				Content:          assistantContent,
				ReasoningContent: reasoningBuf.String(),
				ToolCalls:        finalToolCalls,
			})
		}

		if len(finalToolCalls) == 0 {
			roundCancel()
			break
		}

		for _, tc := range finalToolCalls {
			result, err := a.tools.Execute(ctx, tc.Function.Name, json.RawMessage(tc.Function.Arguments))
			if err != nil {
				result = json.RawMessage(fmt.Sprintf(`{"error":%q}`, err.Error()))
			}
			msgs = append(msgs, llm.Message{
				Role:       llm.RoleTool,
				Content:    string(result),
				ToolCallID: tc.ID,
			})

			if strings.HasPrefix(tc.Function.Name, "compute_") {
				emitIf(onEvent, ChatEvent{
					Type:    EventPhase,
					Content: "正在计算命理数据…",
				})
			}
		}
		roundCancel()
	}

	return msgs, nil
}

func (a *ChatAgent) ensureNamingPrompt(locale string, messages []llm.Message) []llm.Message {
	if len(messages) > 0 && messages[0].Role == llm.RoleSystem {
		return messages
	}
	prompt := strings.ReplaceAll(a.prompt, "{locale}", locale)
	return append([]llm.Message{{Role: llm.RoleSystem, Content: prompt}}, messages...)
}
