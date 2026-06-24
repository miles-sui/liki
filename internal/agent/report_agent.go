package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"liki/internal/llm"
)

// ReportAgent generates full reports from computed chart data using a multi-phase workflow.
// Each instance is bound to a specific product prompt (e.g. chart, bond, naming).
type ReportAgent struct {
	llm          LLMClient
	tools        ToolRegistry
	systemPrompt string // report.txt + report-{product}.md concatenated
}

// NewReportAgent creates a ReportAgent for a specific product.
// sharedPrompt is report.txt (multi-phase workflow + check tools).
// productPrompt is the product-specific report template (data contract + chapter structure).
func NewReportAgent(llmClient LLMClient, tools ToolRegistry, sharedPrompt, productPrompt string) *ReportAgent {
	return &ReportAgent{
		llm:          llmClient,
		tools:        tools,
		systemPrompt: sharedPrompt + "\n" + productPrompt,
	}
}

// Generate runs the full report generation pipeline with verification tools.
// onEvent receives text-delta events for real-time streaming; may be nil.
func (a *ReportAgent) Generate(ctx context.Context, locale string, chartJSON json.RawMessage, onEvent func(ChatEvent)) (string, error) {
	systemPrompt := strings.ReplaceAll(a.systemPrompt, "{locale}", locale)
	userMsg := "请根据以下数据生成完整报告:\n" + string(chartJSON)

	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: systemPrompt},
		{Role: llm.RoleUser, Content: userMsg},
	}

	toolDefs := a.tools.Schemas()

	var buf strings.Builder
	for round := 0; round < maxChatRounds; round++ {
		streamCh, err := a.llm.ChatStreamWithTools(ctx, messages, toolDefs)
		if err != nil {
			return "", fmt.Errorf("report agent: %w", err)
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
			messages = append(messages, llm.Message{
				Role:             llm.RoleAssistant,
				Content:          assistantContent,
				ReasoningContent: reasoningBuf.String(),
				ToolCalls:        finalToolCalls,
			})
			buf.WriteString(assistantContent)
		}

		if len(finalToolCalls) == 0 {
			break
		}

		for _, tc := range finalToolCalls {
			result, err := a.tools.Execute(ctx, tc.Function.Name, json.RawMessage(tc.Function.Arguments))
			if err != nil {
				result = json.RawMessage(fmt.Sprintf(`{"error":%q}`, err.Error()))
			}
			messages = append(messages, llm.Message{
				Role:       llm.RoleTool,
				Content:    string(result),
				ToolCallID: tc.ID,
			})
		}
	}

	return buf.String(), nil
}
