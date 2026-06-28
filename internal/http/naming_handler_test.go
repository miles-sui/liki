package http

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"liki/internal/agent"
	"liki/internal/llm"
	"liki/internal/payment"
	"liki/internal/product"
)

func setupNamingHandler(t *testing.T, store *payment.Store) http.HandlerFunc {
	t.Helper()
	mockLLM := &agent.MockLLM{
		ToolResps: []*llm.ChatResult{agent.ChatRes(nil, "你好，请告诉我你的出生日期")},
	}
	mockTools := &agent.MockToolRegistry{}
	chat := agent.NewChatAgent(mockLLM, mockTools, "test {locale}\n{phase_instruction}")
	return namingHandler(chat, store)
}

func TestNamingHandler_NoJWT(t *testing.T) {
	store := newAuthTestStore(t)
	handler := setupNamingHandler(t, store)

	r := httptest.NewRequest("POST", "/api/agent/naming", strings.NewReader(`{"message":"你好"}`))
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestNamingHandler_OrderNotPaid(t *testing.T) {
	store := newAuthTestStore(t)
	if err := store.CreateOrder(context.Background(), "order-1", product.ProductNaming, 2990, "CNY", "", "", "", "dodo"); err != nil {
		t.Fatalf("create order: %v", err)
	}
	handler := setupNamingHandler(t, store)

	r := httptest.NewRequest("POST", "/api/agent/naming", strings.NewReader(`{"message":"你好"}`))
	setJWTForTest(t, r,"user@example.com", "order-1")
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestNamingHandler_Expired(t *testing.T) {
	store := newAuthTestStore(t)
	seedActiveOrder(t, store, "order-1", "user@example.com", product.ProductNaming, "2020-01-01 00:00:00")
	handler := setupNamingHandler(t, store)

	r := httptest.NewRequest("POST", "/api/agent/naming", strings.NewReader(`{"message":"你好"}`))
	setJWTForTest(t, r,"user@example.com", "order-1")
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestNamingHandler_EmptyMessage(t *testing.T) {
	store := newAuthTestStore(t)
	seedActiveOrder(t, store, "order-1", "user@example.com", product.ProductNaming, "2027-01-01 00:00:00")
	handler := setupNamingHandler(t, store)

	r := httptest.NewRequest("POST", "/api/agent/naming", strings.NewReader(`{"message":""}`))
	setJWTForTest(t, r,"user@example.com", "order-1")
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestNamingHandler_SSEStream(t *testing.T) {
	store := newAuthTestStore(t)
	seedActiveOrder(t, store, "order-1", "user@example.com", product.ProductNaming, "2027-01-01 00:00:00")

	mockLLM := &agent.MockLLM{
		ToolResps: []*llm.ChatResult{agent.ChatRes(nil, "你好，请告诉我你的出生日期")},
	}
	mockTools := &agent.MockToolRegistry{}
	chat := agent.NewChatAgent(mockLLM, mockTools, "test {locale}\n{phase_instruction}")
	handler := namingHandler(chat, store)

	r := httptest.NewRequest("POST", "/api/agent/naming", strings.NewReader(`{"message":"你好"}`))
	setJWTForTest(t, r,"user@example.com", "order-1")
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}

	var events []agent.ChatEvent
	scanner := bufio.NewScanner(w.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var ev agent.ChatEvent
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				continue
			}
			events = append(events, ev)
		}
	}

	hasThinking := false
	hasTextDelta := false
	for _, ev := range events {
		switch ev.Type {
		case agent.EventThinking:
			hasThinking = true
		case agent.EventTextDelta:
			hasTextDelta = true
		}
	}
	if !hasThinking {
		t.Error("missing thinking SSE event")
	}
	if !hasTextDelta {
		t.Error("missing text-delta SSE event")
	}
}

func TestNamingHandler_ToolCallPhaseEvent(t *testing.T) {
	store := newAuthTestStore(t)
	seedActiveOrder(t, store, "order-1", "user@example.com", product.ProductNaming, "2027-01-01 00:00:00")

	mockLLM := &agent.MockLLM{
		ToolResps: []*llm.ChatResult{
			agent.ChatRes([]llm.ToolCall{agent.ToolCall("compute_chart", `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"gender":"male"}`)}, "好的，让我先推算八字"),
			agent.ChatRes(nil, "基于您的八字，我建议…"),
		},
	}
	mockTools := &agent.MockToolRegistry{}
	chat := agent.NewChatAgent(mockLLM, mockTools, "test {locale}\n{phase_instruction}")
	handler := namingHandler(chat, store)

	r := httptest.NewRequest("POST", "/api/agent/naming", strings.NewReader(`{"message":"帮我起名"}`))
	setJWTForTest(t, r,"user@example.com", "order-1")
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var events []agent.ChatEvent
	scanner := bufio.NewScanner(w.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var ev agent.ChatEvent
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				continue
			}
			events = append(events, ev)
		}
	}

	hasThinking := false
	hasPhase := false
	hasTextDelta := false
	for _, ev := range events {
		switch ev.Type {
		case agent.EventThinking:
			hasThinking = true
		case agent.EventPhase:
			hasPhase = true
		case agent.EventTextDelta:
			hasTextDelta = true
		}
	}
	if !hasThinking {
		t.Error("missing thinking event")
	}
	if !hasPhase {
		t.Error("missing phase event for compute_chart tool call")
	}
	if !hasTextDelta {
		t.Error("missing text-delta event")
	}
}

func TestNamingHandler_ReportReady(t *testing.T) {
	store := newAuthTestStore(t)
	seedActiveOrder(t, store, "order-1", "user@example.com", product.ProductNaming, "2027-01-01 00:00:00")

	mockLLM := &agent.MockLLM{
		ToolResps: []*llm.ChatResult{
			agent.ChatRes(nil, "好的，以下是您的起名报告：\n\n# 起名报告\n\n## 一、命盘分析\n\n日主为甲木…"),
		},
	}
	mockTools := &agent.MockToolRegistry{}
	chat := agent.NewChatAgent(mockLLM, mockTools, "test {locale}\n{phase_instruction}")
	handler := namingHandler(chat, store)

	r := httptest.NewRequest("POST", "/api/agent/naming", strings.NewReader(`{"message":"给我出报告"}`))
	setJWTForTest(t, r,"user@example.com", "order-1")
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var hasReportReady bool
	scanner := bufio.NewScanner(w.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var ev agent.ChatEvent
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				continue
			}
			if ev.Type == agent.EventReportReady {
				hasReportReady = true
			}
		}
	}
	if !hasReportReady {
		t.Error("missing report-ready event for output containing # 起名报告")
	}
}

func TestNamingHandler_LLMError(t *testing.T) {
	store := newAuthTestStore(t)
	seedActiveOrder(t, store, "order-1", "user@example.com", product.ProductNaming, "2027-01-01 00:00:00")

	mockLLM := &agent.MockLLM{
		ToolErrs: []error{fmt.Errorf("LLM service unavailable")},
	}
	mockTools := &agent.MockToolRegistry{}
	chat := agent.NewChatAgent(mockLLM, mockTools, "test {locale}\n{phase_instruction}")
	handler := namingHandler(chat, store)

	r := httptest.NewRequest("POST", "/api/agent/naming", strings.NewReader(`{"message":"帮我起名"}`))
	setJWTForTest(t, r,"user@example.com", "order-1")
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var hasError bool
	scanner := bufio.NewScanner(w.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var ev agent.ChatEvent
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				continue
			}
			if ev.Type == agent.EventError {
				hasError = true
			}
		}
	}
	if !hasError {
		t.Error("missing error event for LLM failure")
	}
}

func TestNamingHandler_PersistsMessages(t *testing.T) {
	store := newAuthTestStore(t)
	seedActiveOrder(t, store, "order-1", "user@example.com", product.ProductNaming, "2027-01-01 00:00:00")

	mockLLM := &agent.MockLLM{
		ToolResps: []*llm.ChatResult{agent.ChatRes(nil, "起名建议如下…")},
	}
	mockTools := &agent.MockToolRegistry{}
	chat := agent.NewChatAgent(mockLLM, mockTools, "test {locale}\n{phase_instruction}")
	handler := namingHandler(chat, store)

	r := httptest.NewRequest("POST", "/api/agent/naming", strings.NewReader(`{"message":"帮我起名"}`))
	setJWTForTest(t, r,"user@example.com", "order-1")
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	msgs, err := store.LoadChatHistory(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("LoadChatHistory: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("message count = %d, want 2 (user + assistant)", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[0].Content != "帮我起名" {
		t.Errorf("msg[0] = {%q, %q}, want {user, 帮我起名}", msgs[0].Role, msgs[0].Content)
	}
	if msgs[1].Role != "assistant" || msgs[1].Content != "起名建议如下…" {
		t.Errorf("msg[1] = {%q, %q}, want {assistant, 起名建议如下…}", msgs[1].Role, msgs[1].Content)
	}
}

func TestNamingHandler_LoadsHistory(t *testing.T) {
	store := newAuthTestStore(t)
	seedActiveOrder(t, store, "order-1", "user@example.com", product.ProductNaming, "2027-01-01 00:00:00")

	if err := store.CreateChatMessage(context.Background(), "order-1", payment.RoleUser, "我的出生日期是2026-06-25"); err != nil {
		t.Fatalf("seed history: %v", err)
	}
	if err := store.CreateChatMessage(context.Background(), "order-1", payment.RoleAssistant, "好的，请确认"); err != nil {
		t.Fatalf("seed history: %v", err)
	}

	mockLLM := &agent.MockLLM{
		ToolResps: []*llm.ChatResult{agent.ChatRes(nil, "确认无误，开始起名")},
	}
	mockTools := &agent.MockToolRegistry{}
	chat := agent.NewChatAgent(mockLLM, mockTools, "test {locale}\n{phase_instruction}")
	handler := namingHandler(chat, store)

	r := httptest.NewRequest("POST", "/api/agent/naming", strings.NewReader(`{"message":"确认"}`))
	setJWTForTest(t, r,"user@example.com", "order-1")
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	// 2 historical + user + assistant = 4.
	msgs, err := store.LoadChatHistory(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("LoadChatHistory: %v", err)
	}
	if len(msgs) != 4 {
		t.Fatalf("message count = %d, want 4", len(msgs))
	}
}

func setJWTForTest(t *testing.T, r *http.Request, email, orderID string) {
	t.Helper()
	w := httptest.NewRecorder()
	setJWTCookie(w, email, orderID)
	for _, c := range w.Result().Cookies() {
		if c.Name == "liki_token" {
			r.AddCookie(c)
			return
		}
	}
}
