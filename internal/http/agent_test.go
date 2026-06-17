package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"liki/internal/agent"
	
	"liki/internal/llm"
	"liki/internal/payment"
	"liki/internal/session"
)

// mockOrderCreator implements orderStore for test assertions.
type mockOrderCreator struct {
	created []orderRecord
	err     error
}

type orderRecord struct {
	OrderID   string
	Product   agent.Product
	Amount    int
	ChartJSON string
	LlmJSON   string
}

func (m *mockOrderCreator) CreateOrder(ctx context.Context, orderID string, product agent.Product, amount int, currency, chartJSON, llmJSON, locale string) error {
	if m.err != nil {
		return m.err
	}
	m.created = append(m.created, orderRecord{
		OrderID:   orderID,
		Product:   product,
		Amount:    amount,
		ChartJSON: chartJSON,
		LlmJSON:   llmJSON,
	})
	return nil
}

func (m *mockOrderCreator) UpdateEmail(ctx context.Context, orderID, email string) error {
	return nil
}

func (m *mockOrderCreator) GetOrder(ctx context.Context, orderID string) (*payment.Order, error) {
	return nil, nil
}

func (m *mockOrderCreator) UpdateLlmJSON(ctx context.Context, orderID, llmJSON string) error {
	return nil
}

func stubCollectRegistry() *agent.MockToolRegistry {
	return &agent.MockToolRegistry{Defs: []llm.ToolDef{
		{Type: "function", Function: json.RawMessage(`{"name":"get_city_coords"}`)},
	}}
}
func newTestChatAgent(t *testing.T) *agent.ChatAgent {
	t.Helper()
	m := &agent.MockLLM{
		ToolResps: []*llm.ChatResult{
			agent.ChatRes(nil, "请问您的出生年月日时分和性别？"),
		},
	}
	return agent.NewChatAgent(m, stubCollectRegistry(), "test chat content")
}

// newChatAgentWithPurchase returns an agent whose mock calls compute_chart then purchase.
func newChatAgentWithPurchase(t *testing.T) *agent.ChatAgent {
	t.Helper()
	chartData := json.RawMessage(`{"_product":"chart","data":{"chart":{"nianzhu":{"gan":"庚午"}}}}`)
	m := &agent.MockLLM{
		ToolResps: []*llm.ChatResult{
			agent.ChatRes(
				[]llm.ToolCall{agent.ToolCall("compute_chart", `{"year":1990,"month":5,"day":20,"hour":15,"minute":0,"longitude":116.4,"timezone":8,"gender":"male"}`)},
				"好的，已确认。开始排盘。",
			),
			agent.ChatRes(
				[]llm.ToolCall{agent.ToolCall("purchase", `{"product":"chart"}`)},
				"好的，为您创建订单。",
			),
		},
	}
	tools := &agent.MockToolRegistry{
		Defs: []llm.ToolDef{
			{Type: "function", Function: json.RawMessage(`{"name":"compute_chart"}`)},
			{Type: "function", Function: json.RawMessage(`{"name":"purchase"}`)},
		},
		Results: map[string]json.RawMessage{
			"compute_chart": chartData,
		},
	}
	return agent.NewChatAgent(m, tools, "test prompt")
}

// newChatAgentPurchaseOnly returns an agent whose mock immediately calls purchase.
// For tests where compute result is already in session history.
func newChatAgentPurchaseOnly(t *testing.T) *agent.ChatAgent {
	t.Helper()
	m := &agent.MockLLM{
		ToolResps: []*llm.ChatResult{
			agent.ChatRes(
				[]llm.ToolCall{agent.ToolCall("purchase", `{"product":"chart"}`)},
				"好的，为您创建订单。",
			),
		},
	}
	tools := &agent.MockToolRegistry{
		Defs: []llm.ToolDef{
			{Type: "function", Function: json.RawMessage(`{"name":"purchase"}`)},
		},
	}
	return agent.NewChatAgent(m, tools, "test prompt")
}

// newChatAgentError returns an agent whose mock returns an LLM error.
func newChatAgentError(t *testing.T) *agent.ChatAgent {
	t.Helper()
	m := &agent.MockLLM{
		ToolErrs: []error{fmt.Errorf("deepseek API: 502 Bad Gateway")},
	}
	tools := &agent.MockToolRegistry{}
	return agent.NewChatAgent(m, tools, "test prompt")
}

type testOrchDeps struct {
	chat  *agent.ChatAgent
	store *mockOrderCreator
}
func newTestOrchestrator(t *testing.T, chat *agent.ChatAgent) testOrchDeps {
	t.Helper()
	orders := &mockOrderCreator{}
	_ = map[agent.Product]int{agent.ProductChart: 990, agent.ProductBond: 1990, agent.ProductNaming: 2990}
	return testOrchDeps{chat: chat, store: orders}
}

// flushRecorder wraps httptest.ResponseRecorder and implements http.Flusher.
type flushRecorder struct {
	*httptest.ResponseRecorder
}

func (f *flushRecorder) Flush() {
	f.ResponseRecorder.Flush()
}

func newRecorder() *flushRecorder {
	return &flushRecorder{httptest.NewRecorder()}
}

// --- Tests ---

func TestChatHandler_NewSession(t *testing.T) {
	chat := newTestChatAgent(t); chat.Amounts = map[agent.Product]int{agent.ProductChart: 990, agent.ProductBond: 1990, agent.ProductNaming: 2990}
	store := session.NewStore(5 * time.Minute, 0)
	defer store.Stop()

	td := newTestOrchestrator(t,chat)
	handler := chatHandler(td.chat, td.store, store)

	r := httptest.NewRequest("POST", "/api/agent/chat", strings.NewReader(`{"session_id":"","message":"想看八字"}`))
	r.Header.Set("Content-Type", "application/json")
	w := newRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/event-stream") {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}

	sid := w.Header().Get("X-Session-ID")
	if sid == "" {
		t.Error("X-Session-ID header must not be empty")
	}

	sess, ok := store.Get(sid)
	if !ok {
		t.Error("session should exist after chat request")
	}
	// Without confirmation, the pipeline stops after collection — session won't be closed.
	// Confirmed flow is tested in TestChatHandler_OrderCreation and _SSEEvents_Confirmed.
	if sess.IsClosed() {
		t.Error("session should not be closed after incomplete collect")
	}
	if sess.Phase == session.PhaseClosed {
		t.Errorf("session phase = %q, should not be closed", sess.Phase)
	}
}

func TestChatHandler_ExistingSession(t *testing.T) {
	chat := newTestChatAgent(t); chat.Amounts = map[agent.Product]int{agent.ProductChart: 990, agent.ProductBond: 1990, agent.ProductNaming: 2990}
	store := session.NewStore(5 * time.Minute, 0)
	defer store.Stop()

	sess := store.NewSession()
	sess.AppendMessage(llm.Message{Role: llm.RoleUser, Content: "你好"})

	td := newTestOrchestrator(t,chat)
	handler := chatHandler(td.chat, td.store, store)

	r := httptest.NewRequest("POST", "/api/agent/chat", strings.NewReader(`{"session_id":"`+sess.ID+`","message":"想看八字"}`))
	r.Header.Set("Content-Type", "application/json")
	w := newRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	sid := w.Header().Get("X-Session-ID")
	if sid != sess.ID {
		t.Errorf("session ID = %q, want %q", sid, sess.ID)
	}
}

func TestChatHandler_ClosedSession(t *testing.T) {
	chat := newTestChatAgent(t); chat.Amounts = map[agent.Product]int{agent.ProductChart: 990, agent.ProductBond: 1990, agent.ProductNaming: 2990}
	store := session.NewStore(5 * time.Minute, 0)
	defer store.Stop()

	sess := store.NewSession()
	sess.Phase = "closed"

	td := newTestOrchestrator(t,chat)
	handler := chatHandler(td.chat, td.store, store)

	r := httptest.NewRequest("POST", "/api/agent/chat", strings.NewReader(`{"session_id":"`+sess.ID+`","message":"hello"}`))
	r.Header.Set("Content-Type", "application/json")
	w := newRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400 for closed session", w.Code)
	}
}

func TestChatHandler_EmptyMessage(t *testing.T) {
	chat := newTestChatAgent(t); chat.Amounts = map[agent.Product]int{agent.ProductChart: 990, agent.ProductBond: 1990, agent.ProductNaming: 2990}
	store := session.NewStore(5 * time.Minute, 0)
	defer store.Stop()

	td := newTestOrchestrator(t,chat)
	handler := chatHandler(td.chat, td.store, store)

	r := httptest.NewRequest("POST", "/api/agent/chat", strings.NewReader(`{"session_id":"","message":""}`))
	r.Header.Set("Content-Type", "application/json")
	w := newRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestChatHandler_InvalidJSON(t *testing.T) {
	chat := newTestChatAgent(t); chat.Amounts = map[agent.Product]int{agent.ProductChart: 990, agent.ProductBond: 1990, agent.ProductNaming: 2990}
	store := session.NewStore(5 * time.Minute, 0)
	defer store.Stop()

	td := newTestOrchestrator(t,chat)
	handler := chatHandler(td.chat, td.store, store)

	r := httptest.NewRequest("POST", "/api/agent/chat", strings.NewReader(`not json`))
	r.Header.Set("Content-Type", "application/json")
	w := newRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestChatHandler_WrongMethod(t *testing.T) {
	chat := newTestChatAgent(t); chat.Amounts = map[agent.Product]int{agent.ProductChart: 990, agent.ProductBond: 1990, agent.ProductNaming: 2990}
	store := session.NewStore(5 * time.Minute, 0)
	defer store.Stop()

	td := newTestOrchestrator(t,chat)
	handler := chatHandler(td.chat, td.store, store)

	r := httptest.NewRequest("GET", "/api/agent/chat", nil)
	w := newRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400 (handler rejects non-POST methods)", w.Code)
	}
}

func TestChatHandler_SessionNotFound(t *testing.T) {
	chat := newTestChatAgent(t); chat.Amounts = map[agent.Product]int{agent.ProductChart: 990, agent.ProductBond: 1990, agent.ProductNaming: 2990}
	store := session.NewStore(5 * time.Minute, 0)
	defer store.Stop()

	td := newTestOrchestrator(t,chat)
	handler := chatHandler(td.chat, td.store, store)

	r := httptest.NewRequest("POST", "/api/agent/chat", strings.NewReader(`{"session_id":"nonexistent","message":"hello"}`))
	r.Header.Set("Content-Type", "application/json")
	w := newRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != 404 {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestChatHandler_StoreAtCapacity(t *testing.T) {
	// When the session store is full, NewSession returns nil and the handler
	// must return 503 instead of panicking.
	chat := newTestChatAgent(t); chat.Amounts = map[agent.Product]int{agent.ProductChart: 990, agent.ProductBond: 1990, agent.ProductNaming: 2990}
	store := session.NewStore(5*time.Minute, 1)
	defer store.Stop()

	// Fill the store to capacity.
	_ = store.NewSession()

	td := newTestOrchestrator(t, chat)
	handler := chatHandler(td.chat, td.store, store)

	r := httptest.NewRequest("POST", "/api/agent/chat", strings.NewReader(`{"session_id":"","message":"hello"}`))
	r.Header.Set("Content-Type", "application/json")
	w := newRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != 503 {
		t.Errorf("status = %d, want 503 when store is at capacity", w.Code)
	}
	if w.Body.Len() > 0 {
		// Should return JSON error, not SSE.
		ct := w.Header().Get("Content-Type")
		if !strings.Contains(ct, "application/json") {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
	}
}

func TestChatHandler_SSEEvents(t *testing.T) {
	chat := newChatAgentWithPurchase(t); chat.Amounts = map[agent.Product]int{agent.ProductChart: 990, agent.ProductBond: 1990, agent.ProductNaming: 2990}
	store := session.NewStore(5 * time.Minute, 0)
	defer store.Stop()

	td := newTestOrchestrator(t,chat)
	handler := chatHandler(td.chat, td.store, store)

	r := httptest.NewRequest("POST", "/api/agent/chat", strings.NewReader(`{"session_id":"","message":"帮我排盘，1990年5月20日15点，北京，男"}`))
	r.Header.Set("Content-Type", "application/json")
	w := newRecorder()

	handler.ServeHTTP(w, r)

	body := w.Body.String()
	events := parseSSE(t, body)
	if len(events) == 0 {
		t.Fatalf("expected SSE events in response (body len=%d)", len(body))
	}

	var types []string
	hasDone := false
	hasTextDelta := false
	for _, ev := range events {
		types = append(types, fmt.Sprint(ev["type"]))
		switch ev["type"] {
		case "done":
			hasDone = true
		case "text-delta":
			hasTextDelta = true
		case "error":
			t.Errorf("unexpected error event: %v", ev)
		}
	}
	if !hasDone {
		t.Errorf("expected done event; got types: %v", types)
	}
	if !hasTextDelta {
		t.Errorf("expected text-delta events, got types: %v", types)
	}

	for _, ev := range events {
		if ev["type"] != "done" {
			continue
		}
		data, ok := ev["data"].(map[string]any)
		if !ok {
			t.Errorf("done event missing data field: %v", ev)
			continue
		}
		if data["order_id"] == nil || data["order_id"] == "" {
			t.Error("done event data.order_id must be non-empty")
		}
		amt, _ := data["amount"].(float64)
		if amt <= 0 {
			t.Errorf("done event data.amount must be > 0, got %v", data["amount"])
		}
		if data["product"] == nil || data["product"] == "" {
			t.Error("done event data.product must be non-empty")
		}
	}
}

func TestChatHandler_SSEEvents_Confirmed(t *testing.T) {
	chat := newChatAgentPurchaseOnly(t); chat.Amounts = map[agent.Product]int{agent.ProductChart: 990, agent.ProductBond: 1990, agent.ProductNaming: 2990}
	store := session.NewStore(5 * time.Minute, 0)
	defer store.Stop()

	td := newTestOrchestrator(t,chat)
	handler := chatHandler(td.chat, td.store, store)

	sess := store.NewSession()
	sess.AppendMessage(llm.Message{Role: llm.RoleUser, Content: "帮我排盘"})
	sess.AppendMessage(llm.Message{Role: llm.RoleAssistant, Content: "", ToolCalls: []llm.ToolCall{
		{ID: "call_compute", Type: "function", Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{Name: "compute_chart", Arguments: `{"year":1990,"month":5,"day":20,"hour":15,"minute":0,"longitude":116.4,"timezone":8,"gender":"male"}`}},
	}})
	sess.AppendMessage(llm.Message{Role: llm.RoleTool, Content: `{"_product":"chart","data":{"chart":{"nianzhu":{"gan":"庚午"}}}}`, ToolCallID: "call_compute"})
	sess.AppendMessage(llm.Message{Role: llm.RoleAssistant, Content: "您的八字日主为庚金…"})

	r := httptest.NewRequest("POST", "/api/agent/chat", strings.NewReader(`{"session_id":"`+sess.ID+`","message":"我购买完整报告"}`))
	r.Header.Set("Content-Type", "application/json")
	w := newRecorder()

	handler.ServeHTTP(w, r)

	body := w.Body.String()
	events := parseSSE(t, body)
	if len(events) == 0 {
		t.Fatalf("expected SSE events in response (body len=%d)", len(body))
	}

	var types []string
	hasDone := false
	for _, ev := range events {
		types = append(types, fmt.Sprint(ev["type"]))
		switch ev["type"] {
		case "done":
			hasDone = true
			data, ok := ev["data"].(map[string]any)
			if !ok {
				t.Errorf("done event missing data field: %v", ev)
				continue
			}
			if data["order_id"] == nil || data["order_id"] == "" {
				t.Error("done event data.order_id must be non-empty")
			}
			if data["product"] == nil || data["product"] == "" {
				t.Error("done event data.product must be non-empty")
			}
		case "error":
			t.Errorf("unexpected error event: %v", ev)
		}
	}
	if !hasDone {
		t.Errorf("expected done event, got types: %v", types)
	}
}

func TestChatHandler_LLMError(t *testing.T) {
	// When the LLM returns an error (e.g. API downtime), the handler
	// must send an error SSE event rather than crashing.
	chat := newChatAgentError(t)
	store := session.NewStore(5*time.Minute, 0)
	defer store.Stop()

	td := newTestOrchestrator(t, chat)
	handler := chatHandler(td.chat, td.store, store)

	r := httptest.NewRequest("POST", "/api/agent/chat", strings.NewReader(`{"session_id":"","message":"hello"}`))
	r.Header.Set("Content-Type", "application/json")
	w := newRecorder()

	handler.ServeHTTP(w, r)

	// SSE headers are already sent before the error, so HTTP status is 200.
	events := parseSSE(t, w.Body.String())
	hasError := false
	for _, ev := range events {
		if ev["type"] == "error" {
			hasError = true
			if ev["content"] == "" {
				t.Error("error event should have non-empty content")
			}
		}
	}
	if !hasError {
		t.Error("expected an error event in SSE stream")
	}
}

func TestChatHandler_OrderCreation(t *testing.T) {
	chat := newChatAgentWithPurchase(t); chat.Amounts = map[agent.Product]int{agent.ProductChart: 990, agent.ProductBond: 1990, agent.ProductNaming: 2990}
	store := session.NewStore(5 * time.Minute, 0)
	defer store.Stop()

	td := newTestOrchestrator(t,chat)
	handler := chatHandler(td.chat, td.store, store)

	r := httptest.NewRequest("POST", "/api/agent/chat", strings.NewReader(`{"session_id":"","message":"帮我排盘，1990年5月20日15点，北京，男"}`))
	r.Header.Set("Content-Type", "application/json")
	w := newRecorder()

	handler.ServeHTTP(w, r)

	if len(td.store.created) != 1 {
		t.Fatalf("expected 1 order, got %d", len(td.store.created))
	}
	o := td.store.created[0]
	if o.Product != "chart" {
		t.Errorf("order product = %q, want chart", o.Product)
	}
	if o.Amount != 990 {
		t.Errorf("order amount = %d, want 990", o.Amount)
	}
	if o.OrderID == "" {
		t.Error("order ID must not be empty")
	}
}

func TestChatHandler_OrderCreation_Confirmed(t *testing.T) {
	chat := newChatAgentPurchaseOnly(t); chat.Amounts = map[agent.Product]int{agent.ProductChart: 990, agent.ProductBond: 1990, agent.ProductNaming: 2990}
	store := session.NewStore(5 * time.Minute, 0)
	defer store.Stop()

	td := newTestOrchestrator(t,chat)
	handler := chatHandler(td.chat, td.store, store)

	sess := store.NewSession()
	sess.AppendMessage(llm.Message{Role: llm.RoleUser, Content: "帮我排盘"})
	sess.AppendMessage(llm.Message{Role: llm.RoleAssistant, Content: "", ToolCalls: []llm.ToolCall{
		{ID: "call_compute", Type: "function", Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{Name: "compute_chart", Arguments: `{"year":1990,"month":5,"day":20,"hour":15,"minute":0,"longitude":116.4,"timezone":8,"gender":"male"}`}},
	}})
	sess.AppendMessage(llm.Message{Role: llm.RoleTool, Content: `{"_product":"chart","data":{"chart":{"nianzhu":{"gan":"庚午"}}}}`, ToolCallID: "call_compute"})
	sess.AppendMessage(llm.Message{Role: llm.RoleAssistant, Content: "您的八字日主为庚金…"})

	r := httptest.NewRequest("POST", "/api/agent/chat", strings.NewReader(`{"session_id":"`+sess.ID+`","message":"我购买完整报告"}`))
	r.Header.Set("Content-Type", "application/json")
	w := newRecorder()

	handler.ServeHTTP(w, r)

	if len(td.store.created) != 1 {
		t.Fatalf("expected 1 order created, got %d", len(td.store.created))
	}
	o := td.store.created[0]
	if o.Product != "chart" {
		t.Errorf("order product = %q, want chart", o.Product)
	}
	if o.Amount != 990 {
		t.Errorf("order amount = %d, want 990", o.Amount)
	}
	if o.OrderID == "" {
		t.Error("order ID must not be empty")
	}
	if o.ChartJSON == "" {
		t.Error("order ChartJSON must not be empty")
	}
}

func TestSessionRestoreHandler_Success(t *testing.T) {
	store := session.NewStore(5 * time.Minute, 0)
	defer store.Stop()

	sess := store.NewSession()
	sess.AppendMessage(llm.Message{Role: llm.RoleUser, Content: "你好"})
	sess.AppendMessage(llm.Message{Role: llm.RoleAssistant, Content: "你好！"})

	handler := sessionRestoreHandler(store)
	r := httptest.NewRequest("GET", "/api/agent/session?session_id="+sess.ID, nil)
	w := newRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp struct {
		Data struct {
			Messages []llm.Message `json:"messages"`
			Phase    string        `json:"phase"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Data.Messages) != 2 {
		t.Errorf("got %d messages, want 2", len(resp.Data.Messages))
	}
	if resp.Data.Phase == "closed" {
			t.Error("session should not be closed")
		}
	if resp.Data.Phase != "collecting" {
		t.Errorf("session phase = %q, want collecting", resp.Data.Phase)
	}
}

func TestSessionRestoreHandler_MissingParam(t *testing.T) {
	store := session.NewStore(5 * time.Minute, 0)
	defer store.Stop()

	handler := sessionRestoreHandler(store)
	r := httptest.NewRequest("GET", "/api/agent/session", nil)
	w := newRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestSessionRestoreHandler_NotFound(t *testing.T) {
	store := session.NewStore(5 * time.Minute, 0)
	defer store.Stop()

	handler := sessionRestoreHandler(store)
	r := httptest.NewRequest("GET", "/api/agent/session?session_id=nonexistent", nil)
	w := newRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != 404 {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestChatHandler_TouchOnExistingSession(t *testing.T) {
	chat := newTestChatAgent(t); chat.Amounts = map[agent.Product]int{agent.ProductChart: 990, agent.ProductBond: 1990, agent.ProductNaming: 2990}
	store := session.NewStore(5 * time.Minute, 0)
	defer store.Stop()

	sess := store.NewSession()
	sess.AppendMessage(llm.Message{Role: llm.RoleUser, Content: "你好"})

	sess.ExpiresAt = time.Now().Add(1 * time.Second)
	origExpiry := sess.ExpiresAt

	td := newTestOrchestrator(t,chat)
	handler := chatHandler(td.chat, td.store, store)

	r := httptest.NewRequest("POST", "/api/agent/chat", strings.NewReader(`{"session_id":"`+sess.ID+`","message":"想看八字"}`))
	r.Header.Set("Content-Type", "application/json")
	w := newRecorder()

	handler.ServeHTTP(w, r)

	got, ok := store.Get(sess.ID)
	if !ok {
		t.Fatal("session should still exist after Touch")
	}
	if !got.ExpiresAt.After(origExpiry) {
		t.Error("Touch should have extended expiry past original value")
	}
}

func TestGreetingHandler(t *testing.T) {
	chat := &agent.ChatAgent{}
	chat.Greeting = "你好，我是灵机助手！"

	h := greetingHandler(chat)
	r := httptest.NewRequest("GET", "/api/agent/greeting", nil)
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env struct {
		Data struct {
			Greeting string `json:"greeting"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.Greeting != chat.Greeting {
		t.Errorf("greeting = %q, want %q", env.Data.Greeting, chat.Greeting)
	}
}

// --- helpers ---

func parseSSE(t *testing.T, body string) []map[string]any {
	t.Helper()
	var events []map[string]any
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		raw := strings.TrimPrefix(line, "data: ")
		var ev map[string]any
		if err := json.Unmarshal([]byte(raw), &ev); err != nil {
			continue
		}
		events = append(events, ev)
	}
	return events
}

func TestWriteSSE(t *testing.T) {
	w := httptest.NewRecorder()
	ev := agent.ChatEvent{Type: agent.EventTextDelta, Content: "hello"}
	err := writeSSE(w, w, ev)
	if err != nil {
		t.Fatalf("writeSSE: %v", err)
	}
	if w.Body.String() != "data: {\"type\":\"text-delta\",\"content\":\"hello\"}\n\n" {
		t.Errorf("body = %q", w.Body.String())
	}
	if !w.Flushed {
		t.Error("expected Flush to be called")
	}
}

// brokenWriter fails on write, simulating a disconnected client.
type brokenWriter struct {
	httptest.ResponseRecorder
}

func (bw *brokenWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("broken pipe")
}

func TestWriteSSE_ClientDisconnect(t *testing.T) {
	bw := &brokenWriter{}
	err := writeSSE(bw, bw, agent.ChatEvent{Type: agent.EventTextDelta, Content: "x"})
	if err == nil {
		t.Fatal("expected error when client disconnects during write")
	}
}

func TestFlushSSE(t *testing.T) {
	w := httptest.NewRecorder()
	flushSSE(w, w)
	if w.Body.String() != ": ok\n\n" {
		t.Errorf("body = %q, want : ok\\n\\n", w.Body.String())
	}
}
 
