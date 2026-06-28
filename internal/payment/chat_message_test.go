package payment

import (
	"context"
	"testing"
)

func TestCreateChatMessage(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	// Create an order first (FK constraint requires it).
	ctx := context.Background()
	if err := store.CreateOrder(ctx, "order-1", "naming", 2990, "CNY", "", "", "", "dodo"); err != nil {
		t.Fatalf("create order: %v", err)
	}

	if err := store.CreateChatMessage(ctx, "order-1", RoleUser, "你好"); err != nil {
		t.Fatalf("CreateChatMessage: %v", err)
	}
	if err := store.CreateChatMessage(ctx, "order-1", RoleAssistant, "你好，请告诉我出生信息"); err != nil {
		t.Fatalf("CreateChatMessage: %v", err)
	}

	msgs, err := store.LoadChatHistory(ctx, "order-1")
	if err != nil {
		t.Fatalf("LoadChatHistory: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != RoleUser || msgs[0].Content != "你好" {
		t.Errorf("msg[0] = {%q, %q}, want {user, 你好}", msgs[0].Role, msgs[0].Content)
	}
	if msgs[1].Role != RoleAssistant || msgs[1].Content != "你好，请告诉我出生信息" {
		t.Errorf("msg[1] = {%q, %q}, want {assistant, ...}", msgs[1].Role, msgs[1].Content)
	}
}

func TestBatchCreateChatMessages(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	ctx := context.Background()
	if err := store.CreateOrder(ctx, "order-1", "naming", 2990, "CNY", "", "", "", "dodo"); err != nil {
		t.Fatalf("create order: %v", err)
	}

	batch := []ChatMessage{
		{Role: RoleAssistant, Content: "好的"},
		{Role: RoleTool, Content: `{"status":"confirmed"}`},
		{Role: RoleAssistant, Content: "出生信息已确认"},
	}
	if err := store.BatchCreateChatMessages(ctx, "order-1", batch); err != nil {
		t.Fatalf("BatchCreateChatMessages: %v", err)
	}

	msgs, err := store.LoadChatHistory(ctx, "order-1")
	if err != nil {
		t.Fatalf("LoadChatHistory: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
}

func TestBatchCreateChatMessages_Empty(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := store.BatchCreateChatMessages(context.Background(), "order-1", nil); err != nil {
		t.Fatalf("BatchCreateChatMessages with nil: %v", err)
	}
	if err := store.BatchCreateChatMessages(context.Background(), "order-1", []ChatMessage{}); err != nil {
		t.Fatalf("BatchCreateChatMessages with empty: %v", err)
	}
}

func TestLoadChatHistory_Empty(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	msgs, err := store.LoadChatHistory(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("LoadChatHistory: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages, got %d", len(msgs))
	}
}

func TestUpdateBirthInfoIfEmpty(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	ctx := context.Background()
	if err := store.CreateOrder(ctx, "order-1", "naming", 2990, "CNY", "", "", "", "dodo"); err != nil {
		t.Fatalf("create order: %v", err)
	}

	updated, err := store.UpdateBirthInfoIfEmpty(ctx, "order-1", `{"raw":{"year":2026,"month":6,"day":25}}`)
	if err != nil {
		t.Fatalf("UpdateBirthInfoIfEmpty: %v", err)
	}
	if !updated {
		t.Error("expected updated=true on first write")
	}

	// Second write should be a no-op.
	updated, err = store.UpdateBirthInfoIfEmpty(ctx, "order-1", `{"raw":{"year":2027}}`)
	if err != nil {
		t.Fatalf("UpdateBirthInfoIfEmpty second: %v", err)
	}
	if updated {
		t.Error("expected updated=false on second write")
	}

	// Verify value persisted.
	o, err := store.GetOrder(ctx, "order-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if o.BirthInfo != `{"raw":{"year":2026,"month":6,"day":25}}` {
		t.Errorf("BirthInfo = %q, want first write", o.BirthInfo)
	}
}

func TestSetChatExpiresAtIfEmpty(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	ctx := context.Background()
	if err := store.CreateOrder(ctx, "order-1", "naming", 2990, "CNY", "", "", "", "dodo"); err != nil {
		t.Fatalf("create order: %v", err)
	}

	if err := store.SetChatExpiresAtIfEmpty(ctx, "order-1", "2026-07-03 00:00:00"); err != nil {
		t.Fatalf("SetChatExpiresAtIfEmpty: %v", err)
	}

	// Second write is no-op.
	if err := store.SetChatExpiresAtIfEmpty(ctx, "order-1", "2026-07-10 00:00:00"); err != nil {
		t.Fatalf("SetChatExpiresAtIfEmpty second: %v", err)
	}

	o, err := store.GetOrder(ctx, "order-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if o.ChatExpiresAt != "2026-07-03 00:00:00" {
		t.Errorf("ChatExpiresAt = %q, want first write", o.ChatExpiresAt)
	}
}

func TestFindActiveOrdersByEmail(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	ctx := context.Background()
	email := "test@example.com"

	// Create paid naming order.
	if err := store.CreateOrder(ctx, "order-1", "naming", 2990, "CNY", "", "", "", "dodo"); err != nil {
		t.Fatalf("create order: %v", err)
	}
	if err := store.UpdateEmail(ctx, "order-1", email); err != nil {
		t.Fatalf("update email: %v", err)
	}
	if _, _, _, err := store.MarkPaidIdempotent(ctx, "order-1", "pay-1"); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if err := store.SetChatExpiresAtIfEmpty(ctx, "order-1", "2027-01-01 00:00:00"); err != nil {
		t.Fatalf("set chat expiry: %v", err)
	}

	// Create a pending order — should NOT appear.
	if err := store.CreateOrder(ctx, "order-2", "naming", 2990, "CNY", "", "", "", "dodo"); err != nil {
		t.Fatalf("create order: %v", err)
	}
	if err := store.UpdateEmail(ctx, "order-2", email); err != nil {
		t.Fatalf("update email: %v", err)
	}

	// Create an expired order — should NOT appear.
	if err := store.CreateOrder(ctx, "order-3", "naming", 2990, "CNY", "", "", "", "dodo"); err != nil {
		t.Fatalf("create order: %v", err)
	}
	if err := store.UpdateEmail(ctx, "order-3", email); err != nil {
		t.Fatalf("update email: %v", err)
	}
	if _, _, _, err := store.MarkPaidIdempotent(ctx, "order-3", "pay-3"); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if err := store.OverrideChatExpiresAt(ctx, "order-3", "2020-01-01 00:00:00"); err != nil {
		t.Fatalf("set chat expiry: %v", err)
	}

	orders, err := store.FindActiveOrdersByEmail(ctx, email)
	if err != nil {
		t.Fatalf("FindActiveOrdersByEmail: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 active order, got %d", len(orders))
	}
	if orders[0].OrderID != "order-1" {
		t.Errorf("order ID = %q, want order-1", orders[0].OrderID)
	}

	// No orders for different email.
	orders, err = store.FindActiveOrdersByEmail(ctx, "other@example.com")
	if err != nil {
		t.Fatalf("FindActiveOrdersByEmail: %v", err)
	}
	if len(orders) != 0 {
		t.Errorf("expected 0 orders, got %d", len(orders))
	}
}
