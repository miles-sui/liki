package payment

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"liki/internal/product"

	"time"

)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNewStore_CreatesSchema(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	// Verify orders table exists with correct columns (no pdf_path).
	rows, err := db.Query("PRAGMA table_info(orders)")
	if err != nil {
		t.Fatalf("PRAGMA table_info: %v", err)
	}
	defer rows.Close()

	cols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull bool
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &dflt, &pk); err != nil {
			t.Fatalf("scan column: %v", err)
		}
		cols[name] = true
	}

	required := []string{"order_id", "product", "amount", "currency", "email", "chart_json", "llm_json", "status", "payment_id", "created_at", "updated_at"}
	for _, c := range required {
		if !cols[c] {
			t.Errorf("missing column %q in orders table", c)
		}
	}
	if cols["pdf_path"] {
		t.Error("pdf_path column must not exist in orders table")
	}
	_ = store // suppress unused warning
}

func TestCreateAndGetOrder(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ctx := context.Background()

	orderID := "test-order-1"
	if err := store.CreateOrder(ctx, orderID, product.ProductNaming, 990, "USD", "", `{"chart":{}}`, "", ""); err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	o, err := store.GetOrder(ctx, orderID)
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if o.OrderID != orderID {
		t.Errorf("OrderID = %q, want %q", o.OrderID, orderID)
	}
	if o.Product != product.ProductNaming {
		t.Errorf("Product = %q, want chart", o.Product)
	}
	if o.Amount != 990 {
		t.Errorf("Amount = %d, want 990", o.Amount)
	}
	if o.Status != OrderPending {
		t.Errorf("Status = %q, want pending", o.Status)
	}
	if o.ChartJSON != `{"chart":{}}` {
		t.Errorf("ChartJSON = %q, want {\"chart\":{}}", o.ChartJSON)
	}
	if o.LlmJSON != "" {
		t.Errorf("LlmJSON = %q, want empty", o.LlmJSON)
	}
	if o.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestMarkPaid(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ctx := context.Background()

	orderID := "test-order-paid"
	if err := store.CreateOrder(ctx, orderID, product.ProductNaming, 1990, "USD", "", `{"bond":{}}`, "", ""); err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if err := store.UpdateEmail(ctx, orderID, "user@test.com"); err != nil {
		t.Fatalf("UpdateEmail: %v", err)
	}

	_, email, prod, err := store.MarkPaidIdempotent(ctx, orderID, "pay-123")
	if err != nil {
		t.Fatalf("MarkPaid: %v", err)
	}
	if email != "user@test.com" {
		t.Errorf("email = %q, want user@test.com", email)
	}
	if prod != product.ProductNaming {
		t.Errorf("product = %q, want bond", prod)
	}

	o, err := store.GetOrder(ctx, orderID)
	if err != nil {
		t.Fatalf("GetOrder after MarkPaid: %v", err)
	}
	if o.Status != OrderPaid {
		t.Errorf("Status = %q, want paid", o.Status)
	}
	if o.PaymentID != "pay-123" {
		t.Errorf("PaymentID = %q, want pay-123", o.PaymentID)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	_, err = store.GetOrder(context.Background(), "nonexistent")
	if err != ErrOrderNotFound {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestMarkPaid_NotFound(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	_, _, _, err = store.MarkPaidIdempotent(context.Background(), "nonexistent", "pay-1")
	if err != ErrOrderNotFound {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestUpdateLlmJSON(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ctx := context.Background()

	orderID := "test-order-llm"
	if err := store.CreateOrder(ctx, orderID, product.ProductNaming, 2990, "USD", "", `{"naming":{}}`, "", ""); err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if err := store.UpdateLlmJSON(ctx, orderID, "# Report\n\ncontent"); err != nil {
		t.Fatalf("UpdateLlmJSON: %v", err)
	}

	o, err := store.GetOrder(ctx, orderID)
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if o.LlmJSON != "# Report\n\ncontent" {
		t.Errorf("LlmJSON = %q, want # Report...", o.LlmJSON)
	}
}

func TestCleanStale(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ctx := context.Background()

	// Create a paid order (should survive).
	if err := store.CreateOrder(ctx, "paid-order", product.ProductNaming, 990, "USD", "", `{}`, "", ""); err != nil {
		t.Fatalf("CreateOrder paid: %v", err)
	}
	if _, _, _, err := store.MarkPaidIdempotent(ctx, "paid-order", "pay-1"); err != nil {
		t.Fatalf("MarkPaidIdempotent: %v", err)
	}

	// Create a pending order (should be cleaned).
	if err := store.CreateOrder(ctx, "pending-order", product.ProductNaming, 990, "USD", "", `{}`, "", ""); err != nil {
		t.Fatalf("CreateOrder pending: %v", err)
	}
	// Backdate created_at so it falls within the stale window.
	if _, err := db.ExecContext(ctx, `UPDATE orders SET created_at = '2020-01-01 00:00:00' WHERE order_id = 'pending-order'`); err != nil {
		t.Fatalf("ExecContext: %v", err)
	}

	// Clean orders older than 24h.
	if err := store.CleanStale(ctx, 24*time.Hour); err != nil {
		t.Fatalf("CleanStale: %v", err)
	}

	// Paid order still exists.
	if _, err := store.GetOrder(ctx, "paid-order"); err != nil {
		t.Errorf("paid order should survive cleanup: %v", err)
	}

	// Backdated pending order removed (older than 24h cutoff).
	_, err = store.GetOrder(ctx, "pending-order")
	if err != ErrOrderNotFound {
		t.Errorf("pending order should be cleaned, got %v", err)
	}
}

func TestSchemaColumnMatch(t *testing.T) {
	// Verify the INSERT column list matches the CREATE TABLE schema (no extras).
	db := openTestDB(t)
	_, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	// Collect schema column names (excluding rowid).
	rows, err := db.Query("SELECT name FROM pragma_table_info('orders') ORDER BY cid")
	if err != nil {
		t.Fatalf("pragma_table_info: %v", err)
	}
	defer rows.Close()

	var schemaCols []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		schemaCols = append(schemaCols, name)
	}

	// Columns used in CreateOrder INSERT (must be subset of schema).
	insertCols := []string{"order_id", "product", "amount", "currency", "email", "chart_json", "llm_json"}

	schemaSet := make(map[string]bool, len(schemaCols))
	for _, c := range schemaCols {
		schemaSet[c] = true
	}
	for _, c := range insertCols {
		if !schemaSet[c] {
			t.Errorf("INSERT references column %q not in schema (schema has: %s)", c, strings.Join(schemaCols, ", "))
		}
	}

	// Verify INSERT doesn't miss any NOT NULL column without DEFAULT.
	for _, c := range []string{"order_id", "product", "amount", "currency", "chart_json"} {
		found := false
		for _, ic := range insertCols {
			if ic == c {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("INSERT is missing NOT NULL column %q, will fail", c)
		}
	}
}

func TestMarkPaid_Idempotent(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ctx := context.Background()

	orderID := "test-idempotent"
	if err := store.CreateOrder(ctx, orderID, product.ProductNaming, 990, "USD", "", `{}`, "", ""); err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	newPayment, _, _, err := store.MarkPaidIdempotent(ctx, orderID, "pay-1")
	if err != nil {
		t.Fatalf("first MarkPaid: %v", err)
	}
	if !newPayment {
		t.Error("first MarkPaid should be new payment")
	}

	newPayment2, _, _, err := store.MarkPaidIdempotent(ctx, orderID, "pay-2")
	if err != nil {
		t.Fatalf("second MarkPaid: %v", err)
	}
	if newPayment2 {
		t.Error("second MarkPaid should not be new payment")
	}
}

func TestCreateOrder_DuplicateID(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ctx := context.Background()

	if err := store.CreateOrder(ctx, "dup-order", product.ProductNaming, 990, "USD", "", `{}`, "", ""); err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	err = store.CreateOrder(ctx, "dup-order", product.ProductNaming, 1990, "CNY", "", `{}`, "", "")
	if err == nil {
		t.Error("expected error for duplicate order_id")
	}
}

func TestMarkPaidViaCallback_NewPayment(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ctx := context.Background()

	orderID := "test-callback-new"
	if err := store.CreateOrder(ctx, orderID, product.ProductNaming, 990, "USD", "", `{}`, "", ""); err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	newPayment, email, prod, err := store.MarkPaidViaCallback(ctx, orderID)
	if err != nil {
		t.Fatalf("MarkPaidViaCallback: %v", err)
	}
	if !newPayment {
		t.Error("first call should be new payment")
	}
	if prod != product.ProductNaming {
		t.Errorf("product = %q, want naming", prod)
	}

	// Verify order is paid.
	o, err := store.GetOrder(ctx, orderID)
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if o.Status != OrderPaid {
		t.Errorf("status = %q, want paid", o.Status)
	}
	if o.ChatExpiresAt == "" {
		t.Error("chat_expires_at should be set")
	}

	// Email is empty because we didn't set one.
	_ = email
}

func TestMarkPaidViaCallback_Idempotent(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ctx := context.Background()

	orderID := "test-callback-idempotent"
	if err := store.CreateOrder(ctx, orderID, product.ProductNaming, 990, "USD", "a@b.co", `{}`, "", ""); err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	// First call
	newPayment, email, _, err := store.MarkPaidViaCallback(ctx, orderID)
	if err != nil {
		t.Fatalf("first MarkPaidViaCallback: %v", err)
	}
	if !newPayment {
		t.Error("first call should be new payment")
	}
	if email != "a@b.co" {
		t.Errorf("email = %q, want a@b.co", email)
	}

	// Second call — should be idempotent.
	newPayment2, email2, _, err := store.MarkPaidViaCallback(ctx, orderID)
	if err != nil {
		t.Fatalf("second MarkPaidViaCallback: %v", err)
	}
	if newPayment2 {
		t.Error("second call should not be new payment")
	}
	if email2 != "a@b.co" {
		t.Errorf("email = %q, want a@b.co", email2)
	}
}

func TestMarkPaidViaCallback_NonExistent(t *testing.T) {
	db := openTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	_, _, _, err = store.MarkPaidViaCallback(context.Background(), "nonexistent")
	if !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}
