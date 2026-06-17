//go:build integration

package payment

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

)

// openTestDBFile creates a file-based SQLite database for integration tests.
func openTestDBFile(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	db, err := OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	s, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

// ── File persistence ──

func TestStore_ReopenPreservesData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db1, err := OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB (first): %v", err)
	}
	s1, err := NewStore(db1)
	if err != nil {
		t.Fatalf("NewStore (first): %v", err)
	}
	s1.CreateOrder(context.Background(), "reopen-test", agent.ProductChart, 990, "CNY", `{"x":1}`, "", "zh-Hans")
	db1.Close()

	// Reopen — data must survive.
	db2, err := OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB (second): %v", err)
	}
	defer db2.Close()
	s2, err := NewStore(db2)
	if err != nil {
		t.Fatalf("NewStore (second): %v", err)
	}

	o, err := s2.GetOrder(context.Background(), "reopen-test")
	if err != nil {
		t.Fatalf("GetOrder after reopen: %v", err)
	}
	if o.Product != agent.ProductChart {
		t.Errorf("Product = %q, want chart", o.Product)
	}
	if o.ChartJSON != `{"x":1}` {
		t.Errorf("ChartJSON = %q", o.ChartJSON)
	}
}

func TestStore_ReopenWithPaidOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "paid-persist.db")

	db1, _ := OpenDB(path)
	s1, _ := NewStore(db1)
	s1.CreateOrder(context.Background(), "paid-reopen", agent.ProductBond, 1990, "CNY", `{"bond":{}}`, "", "zh-Hans")
	s1.MarkPaidIdempotent(context.Background(), "paid-reopen", "pay-456")
	db1.Close()

	db2, _ := OpenDB(path)
	defer db2.Close()
	s2, _ := NewStore(db2)

	o, err := s2.GetOrder(context.Background(), "paid-reopen")
	if err != nil {
		t.Fatalf("GetOrder after reopen: %v", err)
	}
	if o.Status != OrderPaid {
		t.Errorf("Status = %q, want paid", o.Status)
	}
	if o.PaymentID != "pay-456" {
		t.Errorf("PaymentID = %q, want pay-456", o.PaymentID)
	}
}

// ── Concurrent access ──

func TestStore_MarkPaidIdempotent_Concurrent(t *testing.T) {
	s := openTestDBFile(t)
	ctx := context.Background()

	s.CreateOrder(ctx, "concurrent-pay", agent.ProductChart, 990, "CNY", `{}`, "", "zh-Hans")

	const n = 5
	results := make(chan bool, n)
	errs := make(chan error, n)

	for i := 0; i < n; i++ {
		go func(idx int) {
			pid := "pay-concurrent-" + string(rune('a'+idx))
			newPayment, _, _, _, err := s.MarkPaidIdempotent(ctx, "concurrent-pay", pid)
			results <- newPayment
			errs <- err
		}(i)
	}

	newCount := 0
	for i := 0; i < n; i++ {
		if <-results {
			newCount++
		}
		if err := <-errs; err != nil {
			t.Errorf("concurrent MarkPaid: %v", err)
		}
	}

	if newCount != 1 {
		t.Errorf("expected exactly 1 newPayment=true, got %d", newCount)
	}

	o, _ := s.GetOrder(ctx, "concurrent-pay")
	if o.Status != OrderPaid {
		t.Errorf("Status = %q, want paid", o.Status)
	}
	if o.PaymentID == "" {
		t.Error("PaymentID is empty after payment")
	}
}

// ── UpdatedAt changes ──

func TestStore_UpdatedAt_ChangesOnModification(t *testing.T) {
	s := openTestDBFile(t)
	ctx := context.Background()

	s.CreateOrder(ctx, "ts-order", agent.ProductChart, 990, "CNY", `{}`, "", "zh-Hans")
	o1, _ := s.GetOrder(ctx, "ts-order")
	createdAt := o1.CreatedAt
	updatedAt := o1.UpdatedAt

	time.Sleep(1 * time.Second)

	s.UpdateEmail(ctx, "ts-order", "a@b.com")
	o2, _ := s.GetOrder(ctx, "ts-order")

	if o2.CreatedAt != createdAt {
		t.Error("CreatedAt should not change on update")
	}
	if !o2.UpdatedAt.After(updatedAt) {
		t.Error("UpdatedAt should advance after modification")
	}
	if o2.Email != "a@b.com" {
		t.Errorf("Email = %q", o2.Email)
	}
}

func TestStore_UpdatedAt_ChangesOnPayment(t *testing.T) {
	s := openTestDBFile(t)
	ctx := context.Background()

	s.CreateOrder(ctx, "ts-pay", agent.ProductChart, 990, "CNY", `{}`, "", "zh-Hans")
	o1, _ := s.GetOrder(ctx, "ts-pay")
	updatedAt := o1.UpdatedAt

	time.Sleep(1 * time.Second)

	s.MarkPaidIdempotent(ctx, "ts-pay", "pay-ts")
	o2, _ := s.GetOrder(ctx, "ts-pay")

	if !o2.UpdatedAt.After(updatedAt) {
		t.Error("UpdatedAt should advance after payment")
	}
}

// ── UpdateLlmJSON overwrites unconditionally ──

func TestStore_UpdateLlmJSON_Overwrites(t *testing.T) {
	s := openTestDBFile(t)
	ctx := context.Background()

	s.CreateOrder(ctx, "llm-overwrite", agent.ProductChart, 990, "CNY", `{}`, "", "zh-Hans")
	s.UpdateLlmJSONIfEmpty(ctx, "llm-overwrite", "v1")

	err := s.UpdateLlmJSON(ctx, "llm-overwrite", "v2")
	if err != nil {
		t.Fatalf("UpdateLlmJSON: %v", err)
	}

	o, _ := s.GetOrder(ctx, "llm-overwrite")
	if o.LlmJSON != "v2" {
		t.Errorf("LlmJSON = %q, want v2 (should be overwritten)", o.LlmJSON)
	}
}

// ── OpenDB creates directory ──

func TestOpenDB_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "new", "subdir")
	path := filepath.Join(dir, "test.db")

	db, err := OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	db.Close()

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("directory was not created by OpenDB")
	}
}

// ── Locale field ──

func TestStore_Locale_Persists(t *testing.T) {
	s := openTestDBFile(t)
	ctx := context.Background()

	tests := []struct {
		orderID string
		locale  string
	}{
		{"loc-zh", "zh-Hans"},
		{"loc-hk", "zh-Hant"},
		{"loc-en", "en"},
	}

	for _, tt := range tests {
		if err := s.CreateOrder(ctx, tt.orderID, agent.ProductChart, 990, "CNY", `{}`, "", tt.locale); err != nil {
			t.Errorf("CreateOrder(%s): %v", tt.orderID, err)
		}
	}

	for _, tt := range tests {
		o, err := s.GetOrder(ctx, tt.orderID)
		if err != nil {
			t.Errorf("GetOrder(%s): %v", tt.orderID, err)
			continue
		}
		if o.Locale != tt.locale {
			t.Errorf("%s: Locale = %q, want %q", tt.orderID, o.Locale, tt.locale)
		}
	}
}

// ── ChartJSON with llm_json initially set ──

func TestStore_CreateOrder_WithInitialLlmJSON(t *testing.T) {
	s := openTestDBFile(t)
	ctx := context.Background()

	s.CreateOrder(ctx, "prefilled-llm", agent.ProductChart, 990, "CNY", `{"chart":{}}`, "# Pre-generated\n\nLLM content", "zh-Hans")

	o, _ := s.GetOrder(ctx, "prefilled-llm")
	if o.LlmJSON != "# Pre-generated\n\nLLM content" {
		t.Errorf("LlmJSON = %q, want pre-generated content", o.LlmJSON)
	}

	// UpdateLlmJSONIfEmpty should NOT overwrite.
	updated, _ := s.UpdateLlmJSONIfEmpty(ctx, "prefilled-llm", "new")
	if updated {
		t.Error("UpdateLlmJSONIfEmpty should not overwrite pre-filled llm_json")
	}
}

// ── CleanStale with age boundary ──

func TestStore_CleanStale_AgeBoundary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "boundary.db")

	db, err := OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close()
	s, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ctx := context.Background()

	s.CreateOrder(ctx, "very-old", agent.ProductChart, 990, "CNY", `{}`, "", "zh-Hans")
	s.CreateOrder(ctx, "very-new", agent.ProductChart, 990, "CNY", `{}`, "", "zh-Hans")

	// Backdate the old order.
	db.ExecContext(ctx, `UPDATE orders SET created_at = '2020-01-01 00:00:00' WHERE order_id = 'very-old'`)

	// Clean only very old orders (created before 2021).
	err = s.CleanStale(ctx, 365*24*time.Hour) // ~1 year from now — very-old is from 2020
	if err != nil {
		t.Fatalf("CleanStale: %v", err)
	}

	_, err = s.GetOrder(ctx, "very-old")
	if err != ErrOrderNotFound {
		t.Errorf("very-old should be cleaned, got %v", err)
	}

	o, err := s.GetOrder(ctx, "very-new")
	if err != nil {
		t.Fatalf("very-new should survive: %v", err)
	}
	if o.Status != OrderPending {
		t.Errorf("very-new status = %q, want pending", o.Status)
	}
}

// ── GetOrder timestamps are valid ──

func TestStore_GetOrder_Timestamps(t *testing.T) {
	s := openTestDBFile(t)
	ctx := context.Background()

	before := time.Now()
	s.CreateOrder(ctx, "ts-valid", agent.ProductChart, 990, "CNY", `{}`, "", "zh-Hans")
	after := time.Now()

	o, _ := s.GetOrder(ctx, "ts-valid")

	if o.CreatedAt.Before(before.Add(-1*time.Second)) || o.CreatedAt.After(after.Add(1*time.Second)) {
		t.Errorf("CreatedAt = %v, expected between %v and %v", o.CreatedAt, before, after)
	}
	if o.UpdatedAt != o.CreatedAt {
		t.Error("UpdatedAt should equal CreatedAt for new orders")
	}
}
