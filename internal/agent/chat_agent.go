package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"


	"liki/internal/llm"

	"github.com/google/uuid"
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
	EventDone          ChatEventType = "done"
	EventError         ChatEventType = "error"
	EventThinking      ChatEventType = "thinking"
	EventThinkingDelta ChatEventType = "thinking-delta"
)

// ChatEvent is emitted during streaming report generation and sent as SSE.
type ChatEvent struct {
	Type    ChatEventType `json:"type"`
	Content string        `json:"content,omitempty"`
	Data    any           `json:"data,omitempty"`
}

const (
	defaultCurrency  = "USD"
	chatRoundTimeout = 120 * time.Second
	maxChatRounds    = 20
)

// Chat runs the full chat pipeline: collection → compute → teaser → Q&A → purchase.
// onEvent receives text-delta and phase events for real-time client feedback; may be nil.
func (a *ChatAgent) Chat(ctx context.Context, locale string, messages []llm.Message, onEvent func(ChatEvent), orderCreator OrderCreator, amounts map[Product]int) (*ChatResult, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("agent: chat: no messages")
	}
	emitIf(onEvent, ChatEvent{Type: EventThinking})

	msgs := a.ensureSystemPrompt(locale, messages)
	tools := a.tools.Schemas()

	for round := 0; round < maxChatRounds; round++ {
		roundCtx, roundCancel := context.WithTimeout(ctx, chatRoundTimeout)

		streamCh, err := a.llm.ChatStreamWithTools(roundCtx, msgs, tools)
		if err != nil {
			roundCancel()
			return nil, fmt.Errorf("agent: chat round %d: %w", round, err)
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
			if tc.Function.Name == "purchase" {
				purchase, err := a.handlePurchase(ctx, tc, msgs, orderCreator, amounts, locale)
				if err != nil {
					roundCancel()
					return nil, err
				}

				msgs = append(msgs, llm.Message{
					Role:       llm.RoleTool,
					Content:    fmt.Sprintf(`{"status":"ok","order_id":%q}`, purchase.OrderID),
					ToolCallID: tc.ID,
				})

				emitIf(onEvent, ChatEvent{
					Type:    EventPhase,
					Content: "正在创建订单…",
				})

				roundCancel()
				return &ChatResult{
					Messages: msgs,
					Purchase: purchase,
				}, nil
			}

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

	return &ChatResult{Messages: msgs}, nil
}

// handlePurchase extracts product from purchase args, finds the corresponding
// compute result and Q&A, and creates an order.
func (a *ChatAgent) handlePurchase(ctx context.Context, tc llm.ToolCall, msgs []llm.Message, orderCreator OrderCreator, amounts map[Product]int, locale string) (*PurchaseInfo, error) {
	var args struct {
		Product string `json:"product"`
		Email   string `json:"email"`
	}
	if err := json.Unmarshal(json.RawMessage(tc.Function.Arguments), &args); err != nil {
		return nil, fmt.Errorf("agent: purchase: %w", err)
	}
	product := Product(args.Product)

	chartJSON := findComputeResult(msgs, string(product))
	if chartJSON == nil {
		return nil, fmt.Errorf("agent: purchase: no compute result for %s", product)
	}

	qaJSON := extractQAMessages(msgs, string(product))

	// Embed Q&A into chart JSON so the full report can reference user questions.
	if len(qaJSON) > 0 {
		var chartMap map[string]json.RawMessage
		if err := json.Unmarshal(chartJSON, &chartMap); err == nil {
			chartMap["_qa"] = qaJSON
			var err error
			chartJSON, err = json.Marshal(chartMap)
			if err != nil {
				return nil, fmt.Errorf("agent: purchase: marshal chart+qa: %w", err)
			}
		}
	}

	amount, ok := amounts[product]
	if !ok {
		return nil, fmt.Errorf("agent: purchase: no amount configured for %s", product)
	}

	orderID := uuid.New().String()
	if err := orderCreator.CreateOrder(ctx, orderID, product, amount, defaultCurrency, string(chartJSON), "", locale); err != nil {
		return nil, fmt.Errorf("agent: create order: %w", err)
	}

	if args.Email != "" {
		if err := orderCreator.UpdateEmail(ctx, orderID, args.Email); err != nil {
			slog.Warn("agent: update email for order", "orderID", orderID, "err", err)
		}
	}

	return &PurchaseInfo{
		OrderID: orderID,
		Amount:  amount,
		Product: product,
	}, nil
}

func (a *ChatAgent) ensureSystemPrompt(locale string, messages []llm.Message) []llm.Message {
	return append([]llm.Message{{Role: llm.RoleSystem, Content: a.systemPrompt(locale)}}, messages...)
}

// GenerateFromData generates a report from pre-computed engine data without tool calling.
// Uses ReportPrompts map to select the right prompt for the product.
// Used for full report generation after payment (webhook) and as fallback (report API).
func (a *ChatAgent) GenerateFromData(ctx context.Context, locale string, product Product, chartJSON json.RawMessage, onEvent func(ChatEvent)) (string, error) {
	prompt := a.ReportPrompts[product]
	if prompt == "" {
		prompt = a.prompt
	}
	systemPrompt := strings.ReplaceAll(prompt, "{locale}", locale)
	userMsg := "请根据以下数据生成完整报告:\n" + string(chartJSON)

	tokenCh, err := a.llm.ChatStream(ctx, systemPrompt, userMsg)
	if err != nil {
		return "", fmt.Errorf("agent: generate from data: %w", err)
	}

	var buf strings.Builder
	for token := range tokenCh {
		buf.WriteString(token)
		if onEvent != nil {
			onEvent(ChatEvent{Type: EventTextDelta, Content: token})
		}
	}
	return buf.String(), nil
}

func findComputeResult(msgs []llm.Message, product string) json.RawMessage {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role != llm.RoleTool {
			continue
		}
		var wrapper struct {
			Product string          `json:"_product"`
			Data    json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal([]byte(msgs[i].Content), &wrapper); err != nil {
			continue
		}
		if wrapper.Product == product {
			return wrapper.Data
		}
	}
	return nil
}

// extractQAMessages extracts user and assistant messages after the last
// compute_* tool result for the given product.
func extractQAMessages(msgs []llm.Message, product string) json.RawMessage {
	toolResultIdx := -1
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role != llm.RoleTool {
			continue
		}
		var wrapper struct {
			Product string `json:"_product"`
		}
		if err := json.Unmarshal([]byte(msgs[i].Content), &wrapper); err != nil {
			continue
		}
		if wrapper.Product == product {
			toolResultIdx = i
			break
		}
	}
	if toolResultIdx < 0 {
		return nil
	}

	// Collect user and assistant messages after the tool result.
	type qaMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	var qa []qaMsg
	for i := toolResultIdx + 1; i < len(msgs); i++ {
		m := msgs[i]
		if m.Role == llm.RoleUser || m.Role == llm.RoleAssistant {
			qa = append(qa, qaMsg{Role: string(m.Role), Content: m.Content})
		}
	}
	if len(qa) == 0 {
		return nil
	}
	raw, err := json.Marshal(qa)
	if err != nil {
		return nil
	}
	return json.RawMessage(raw)
}
