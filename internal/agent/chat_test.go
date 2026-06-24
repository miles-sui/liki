package agent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"liki/internal/llm"
)

func newTestChatAgent(t *testing.T, m *MockLLM, tools *MockToolRegistry) *ChatAgent {
	t.Helper()
	return NewChatAgent(m, tools, "test prompt")
}

// stubPurchaseOrderCreator records order creation calls.
type stubPurchaseOrderCreator struct {
	created     []orderRecord
	emailAddrs  []emailRecord
}

type orderRecord struct {
	orderID   string
	product   Product
	amount    int
	chartJSON string
}

type emailRecord struct {
	orderID string
	email   string
}

func (s *stubPurchaseOrderCreator) CreateOrder(ctx context.Context, orderID string, product Product, amount int, currency, chartJSON, llmJSON, locale, provider string) error {
	s.created = append(s.created, orderRecord{orderID, product, amount, chartJSON})
	return nil
}

func (s *stubPurchaseOrderCreator) UpdateEmail(ctx context.Context, orderID, email string) error {
	s.emailAddrs = append(s.emailAddrs, emailRecord{orderID, email})
	return nil
}

func TestChat_NoToolCalls(t *testing.T) {
	m := &MockLLM{
		ToolResps: []*llm.ChatResult{
			ChatRes(nil, "请问您的出生年月日时分？"),
		},
	}
	tools := &MockToolRegistry{}
	a := newTestChatAgent(t,m, tools)

	result, err := a.Chat(context.Background(), "zh-Hans", []llm.Message{
		{Role: llm.RoleUser, Content: "想看八字"},
	}, nil, &stubPurchaseOrderCreator{}, map[Product]int{})

	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if result.Purchase != nil {
		t.Error("expected no purchase")
	}
	if len(result.Messages) == 0 {
		t.Error("expected messages")
	}
}

func TestChat_PurchaseChart(t *testing.T) {
	chartData := json.RawMessage(`{"_product":"chart","data":{"chart":{"nianzhu":{"gan":"庚午","zhi":"庚午"}}}}`)
	m := &MockLLM{
		ToolResps: []*llm.ChatResult{
			// Round 1: LLM calls compute_chart
			ChatRes(
				[]llm.ToolCall{ToolCall("compute_chart", `{"year":1990,"month":5,"day":20,"hour":15,"minute":0,"longitude":116.4,"timezone":8,"gender":"male"}`)},
				"好的，已确认。开始排盘。",
			),
			// Round 2: LLM calls purchase after seeing user's buy message
			ChatRes(
				[]llm.ToolCall{ToolCall("purchase", `{"product":"chart"}`)},
				"好的，为您创建订单。",
			),
		},
	}
	tools := &MockToolRegistry{
		Results: map[string]json.RawMessage{"compute_chart": chartData},
		Defs: []llm.ToolDef{
			{Type: "function", Function: json.RawMessage(`{"name":"compute_chart"}`)},
			{Type: "function", Function: json.RawMessage(`{"name":"purchase"}`)},
		},
	}
	a := newTestChatAgent(t,m, tools)
	orderCreator := &stubPurchaseOrderCreator{}
	amounts := map[Product]int{ProductChart: 990}

	result, err := a.Chat(context.Background(), "zh-Hans", []llm.Message{
		{Role: llm.RoleUser, Content: "帮我排盘，1990年5月20日15点，北京，男"},
		{Role: llm.RoleUser, Content: "我购买完整报告"},
	}, nil, orderCreator, amounts)

	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if result.Purchase == nil {
		t.Fatal("expected purchase")
	}
	if result.Purchase.Product != ProductChart {
		t.Errorf("product = %q, want chart", result.Purchase.Product)
	}
	if result.Purchase.Amount != 990 {
		t.Errorf("amount = %d, want 990", result.Purchase.Amount)
	}
	if len(orderCreator.created) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orderCreator.created))
	}
}

func TestChat_EmptyAssistantNotAppended(t *testing.T) {
	// Regression: LLM returns empty content + no tool calls → empty assistant
	// message must NOT be saved to message history, or the next request will
	// fail with DeepSeek 400: "Invalid assistant message: content or tool_calls must be set"
	m := &MockLLM{
		ToolResps: []*llm.ChatResult{
			ChatRes(nil, ""), // reasoning-only response, no content, no tool calls
		},
	}
	a := newTestChatAgent(t, m, &MockToolRegistry{})

	result, err := a.Chat(context.Background(), "zh-Hans", []llm.Message{
		{Role: llm.RoleUser, Content: "你好"},
	}, nil, &stubPurchaseOrderCreator{}, map[Product]int{})

	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if result.Purchase != nil {
		t.Error("expected no purchase")
	}
	// Verify no empty assistant message in returned messages.
	for _, msg := range result.Messages {
		if msg.Role == llm.RoleAssistant && msg.Content == "" && len(msg.ToolCalls) == 0 {
			t.Error("empty assistant message should not be appended")
		}
	}
}

func TestChat_LLMError(t *testing.T) {
	m := &MockLLM{
		ToolErrs: []error{context.DeadlineExceeded},
	}
	a := newTestChatAgent(t,m, &MockToolRegistry{})

	_, err := a.Chat(context.Background(), "zh-Hans", []llm.Message{
		{Role: llm.RoleUser, Content: "hi"},
	}, nil, &stubPurchaseOrderCreator{}, map[Product]int{})

	if err == nil {
		t.Fatal("expected error from ChatStreamWithTools")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, want context.DeadlineExceeded", err)
	}
}

func TestFindComputeResult(t *testing.T) {
	msgs := []llm.Message{
		{Role: llm.RoleUser, Content: "帮我排盘"},
		{Role: llm.RoleAssistant, Content: "好的", ToolCalls: []llm.ToolCall{ToolCall("compute_chart", `{}`)}},
		{Role: llm.RoleTool, Content: `{"_product":"chart","data":{"chart":{"nianzhu":{"gan":"庚午"}}}}`, ToolCallID: "call_compute_chart"},
		{Role: llm.RoleAssistant, Content: "您的八字…"},
	}

	result := findComputeResult(msgs, "chart")
	if result == nil {
		t.Fatal("expected chart data")
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(result, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := data["chart"]; !ok {
		t.Error("expected chart key")
	}
}

func TestFindComputeResult_NotFound(t *testing.T) {
	msgs := []llm.Message{
		{Role: llm.RoleUser, Content: "hi"},
	}
	if result := findComputeResult(msgs, "chart"); result != nil {
		t.Error("expected nil for missing product")
	}
}

func TestExtractQAMessages(t *testing.T) {
	msgs := []llm.Message{
		{Role: llm.RoleAssistant, Content: "已确认", ToolCalls: []llm.ToolCall{ToolCall("compute_chart", `{}`)}},
		{Role: llm.RoleTool, Content: `{"_product":"chart","data":{}}`, ToolCallID: "call_compute_chart"},
		{Role: llm.RoleAssistant, Content: "您的日主为庚金"},
		{Role: llm.RoleUser, Content: "用神为什么是火？"},
		{Role: llm.RoleAssistant, Content: "因为日主庚金生于巳月…"},
		{Role: llm.RoleUser, Content: "我购买完整报告"},
	}

	result := extractQAMessages(msgs, "chart")
	if result == nil {
		t.Fatal("expected Q&A messages")
	}

	var qa []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(result, &qa); err != nil {
		t.Fatalf("unmarshal Q&A: %v", err)
	}
	if len(qa) != 4 {
		t.Errorf("expected 4 Q&A messages, got %d: %v", len(qa), qa)
	}
}

func TestExtractQAMessages_NoProduct(t *testing.T) {
	msgs := []llm.Message{
		{Role: llm.RoleUser, Content: "hi"},
	}
	if result := extractQAMessages(msgs, "bond"); result != nil {
		t.Error("expected nil when product not found")
	}
}

func TestChat_PurchaseWithEmail(t *testing.T) {
	chartData := json.RawMessage(`{"_product":"chart","data":{"chart":{"nianzhu":{"gan":"庚午"}}}}`)
	m := &MockLLM{
		ToolResps: []*llm.ChatResult{
			ChatRes(
				[]llm.ToolCall{ToolCall("compute_chart", `{"year":1990,"month":5,"day":20,"hour":15,"minute":0,"longitude":116.4,"timezone":8,"gender":"male"}`)},
				"排盘完成。",
			),
			ChatRes(
				[]llm.ToolCall{ToolCall("purchase", `{"product":"chart","email":"user@example.com"}`)},
				"订单已创建，报告将发送至您的邮箱。",
			),
		},
	}
	tools := &MockToolRegistry{
		Results: map[string]json.RawMessage{"compute_chart": chartData},
		Defs: []llm.ToolDef{
			{Type: "function", Function: json.RawMessage(`{"name":"compute_chart"}`)},
			{Type: "function", Function: json.RawMessage(`{"name":"purchase"}`)},
		},
	}
	a := newTestChatAgent(t, m, tools)
	orderCreator := &stubPurchaseOrderCreator{}
	amounts := map[Product]int{ProductChart: 990}

	result, err := a.Chat(context.Background(), "zh-Hans", []llm.Message{
		{Role: llm.RoleUser, Content: "排盘 1990年5月20日15点 北京 男"},
		{Role: llm.RoleUser, Content: "购买，邮箱 user@example.com"},
	}, nil, orderCreator, amounts)

	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if result.Purchase == nil {
		t.Fatal("expected purchase")
	}
	if len(orderCreator.emailAddrs) != 1 {
		t.Fatalf("expected 1 email update, got %d", len(orderCreator.emailAddrs))
	}
	if orderCreator.emailAddrs[0].email != "user@example.com" {
		t.Errorf("email = %q, want user@example.com", orderCreator.emailAddrs[0].email)
	}
}

func TestChat_PurchaseEmbedsQA(t *testing.T) {
	// Q&A messages after the compute result must be embedded as _qa
	// so the full report can reference user questions.
	chartData := json.RawMessage(`{"_product":"chart","data":{"chart":{"nianzhu":{"gan":"庚午"}}}}`)
	m := &MockLLM{
		ToolResps: []*llm.ChatResult{
			// The LLM receives the full conversation history including prior
			// compute result + Q&A, and issues purchase in a single round.
			ChatRes(
				[]llm.ToolCall{ToolCall("purchase", `{"product":"chart"}`)},
				"为您创建订单。",
			),
		},
	}
	tools := &MockToolRegistry{
		Defs: []llm.ToolDef{
			{Type: "function", Function: json.RawMessage(`{"name":"purchase"}`)},
		},
	}
	a := newTestChatAgent(t, m, tools)
	orderCreator := &stubPurchaseOrderCreator{}
	amounts := map[Product]int{ProductChart: 990}

	// Simulate a multi-turn session: initial request → compute → teaser → Q&A → purchase.
	_, err := a.Chat(context.Background(), "zh-Hans", []llm.Message{
		{Role: llm.RoleUser, Content: "排盘 1990年5月20日15点 北京 男"},
		{Role: llm.RoleAssistant, Content: "好的", ToolCalls: []llm.ToolCall{ToolCall("compute_chart", `{}`)}},
		{Role: llm.RoleTool, Content: string(chartData), ToolCallID: "call_compute"},
		{Role: llm.RoleAssistant, Content: "您的日主为庚金，生于巳月…"},
		{Role: llm.RoleUser, Content: "我的用神是什么？"},
		{Role: llm.RoleAssistant, Content: "庚金生于巳月火旺，以水为用神调候…"},
		{Role: llm.RoleUser, Content: "购买完整报告"},
	}, nil, orderCreator, amounts)

	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if len(orderCreator.created) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orderCreator.created))
	}

	var stored map[string]json.RawMessage
	if err := json.Unmarshal([]byte(orderCreator.created[0].chartJSON), &stored); err != nil {
		t.Fatalf("stored chartJSON is not a valid object: %v", err)
	}
	qaRaw, ok := stored["_qa"]
	if !ok {
		t.Fatal("stored chartJSON missing _qa key — Q&A was not embedded")
	}
	var qa []map[string]string
	if err := json.Unmarshal(qaRaw, &qa); err != nil {
		t.Fatalf("_qa is not valid JSON: %v", err)
	}
	if len(qa) == 0 {
		t.Error("_qa is empty, expected Q&A messages")
	}
	foundQuestion := false
	foundAnswer := false
	for _, m := range qa {
		if m["role"] == "user" && m["content"] == "我的用神是什么？" {
			foundQuestion = true
		}
		if m["role"] == "assistant" && m["content"] == "庚金生于巳月火旺，以水为用神调候…" {
			foundAnswer = true
		}
	}
	if !foundQuestion {
		t.Error("_qa missing user question")
	}
	if !foundAnswer {
		t.Error("_qa missing assistant answer")
	}
}

func TestChat_PurchaseNoComputeResult(t *testing.T) {
	// purchase without prior compute_chart is a usage error — Chat returns the error.
	m := &MockLLM{
		ToolResps: []*llm.ChatResult{
			ChatRes(
				[]llm.ToolCall{ToolCall("purchase", `{"product":"chart"}`)},
				"",
			),
		},
	}
	tools := &MockToolRegistry{
		Defs: []llm.ToolDef{
			{Type: "function", Function: json.RawMessage(`{"name":"purchase"}`)},
		},
	}
	a := newTestChatAgent(t, m, tools)
	orderCreator := &stubPurchaseOrderCreator{}

	_, err := a.Chat(context.Background(), "zh-Hans", []llm.Message{
		{Role: llm.RoleUser, Content: "我要购买"},
	}, nil, orderCreator, map[Product]int{ProductChart: 990})

	if err == nil {
		t.Fatal("expected error when purchase has no prior compute result")
	}
	if len(orderCreator.created) != 0 {
		t.Errorf("expected 0 orders on error, got %d", len(orderCreator.created))
	}
}

func TestChat_PurchaseUnknownProduct(t *testing.T) {
	chartData := json.RawMessage(`{"_product":"chart","data":{"chart":{}}}`)
	m := &MockLLM{
		ToolResps: []*llm.ChatResult{
			ChatRes(
				[]llm.ToolCall{ToolCall("compute_chart", `{"year":1990,"month":5,"day":20,"hour":15,"minute":0,"longitude":116.4,"timezone":8,"gender":"male"}`)},
				"排盘完成。",
			),
			ChatRes(
				[]llm.ToolCall{ToolCall("purchase", `{"product":"unknown_product"}`)},
				"",
			),
		},
	}
	tools := &MockToolRegistry{
		Results: map[string]json.RawMessage{"compute_chart": chartData},
		Defs: []llm.ToolDef{
			{Type: "function", Function: json.RawMessage(`{"name":"compute_chart"}`)},
			{Type: "function", Function: json.RawMessage(`{"name":"purchase"}`)},
		},
	}
	a := newTestChatAgent(t, m, tools)
	orderCreator := &stubPurchaseOrderCreator{}

	_, err := a.Chat(context.Background(), "zh-Hans", []llm.Message{
		{Role: llm.RoleUser, Content: "排盘 1990年5月20日15点 北京 男"},
		{Role: llm.RoleUser, Content: "购买"},
	}, nil, orderCreator, map[Product]int{})

	if err == nil {
		t.Fatal("expected error when product has no amount configured")
	}
	if len(orderCreator.created) != 0 {
		t.Errorf("expected 0 orders, got %d", len(orderCreator.created))
	}
}
