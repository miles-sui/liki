package payment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"liki/internal/agent"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS orders (
    order_id    TEXT PRIMARY KEY,
    product     TEXT NOT NULL,
    amount      INTEGER NOT NULL,
    currency    TEXT NOT NULL,
    provider    TEXT NOT NULL DEFAULT '',
    email       TEXT NOT NULL DEFAULT '',
    chart_json  TEXT NOT NULL,
    llm_json    TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'pending',
    payment_id  TEXT,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_payment_id ON orders(payment_id) WHERE payment_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_orders_stale ON orders(status, created_at);
`

// dropPdfPath drops the legacy pdf_path column if it exists (from pre-lingji schema).
const dropPdfPath = `ALTER TABLE orders DROP COLUMN pdf_path;`

// addLocale adds locale column for multi-language report generation.
const addLocale = `ALTER TABLE orders ADD COLUMN locale TEXT NOT NULL DEFAULT 'zh-Hans';`

// addProvider adds provider column for tracking which payment gateway was used.
const addProvider = `ALTER TABLE orders ADD COLUMN provider TEXT NOT NULL DEFAULT '';`

// OrderStatus is the payment status of an order.
type OrderStatus string

const (
	OrderPending OrderStatus = "pending"
	OrderPaid    OrderStatus = "paid"
)

// Order holds the database record for a payment order.
type Order struct {
	OrderID   string
	Product   agent.Product
	Amount    int
	Currency  string
	Provider  string
	Email     string
	ChartJSON string
	LlmJSON   string
	Status    OrderStatus
	PaymentID string
	Locale    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Store provides SQLite-backed order persistence.
type Store struct{ db *sql.DB }

// NewStore creates the orders table and returns a new Store.
func NewStore(db *sql.DB) (*Store, error) {
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("payment: create table: %w", err)
	}
	// Drop legacy pdf_path column from pre-lingji schema. Column may not
	// exist on fresh installs — failure is expected and harmless.
	if _, err := db.Exec(dropPdfPath); err != nil {
		slog.Info("payment: drop pdf_path (expected on fresh installs)", "err", err)
	}
	// Add locale column for multi-language report generation. Column may
	// already exist on upgraded instances — failure is expected and harmless.
	if _, err := db.Exec(addLocale); err != nil {
		slog.Info("payment: add locale (expected on upgraded instances)", "err", err)
	}
	// Add provider column for tracking payment gateway. Column may
	// already exist on upgraded instances — failure is expected and harmless.
	if _, err := db.Exec(addProvider); err != nil {
		slog.Info("payment: add provider (expected on upgraded instances)", "err", err)
	}
	return &Store{db: db}, nil
}

// CreateOrder inserts a new pending order into the database.
func (s *Store) CreateOrder(ctx context.Context, orderID string, product agent.Product, amount int, currency, chartJSON, llmJSON, locale, provider string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO orders (order_id, product, amount, currency, chart_json, llm_json, locale, provider) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		orderID, product, amount, currency, chartJSON, llmJSON, locale, provider)
	if err != nil {
		return fmt.Errorf("payment: create order: %w", err)
	}
	return nil
}

// UpdateEmail updates the email address for an existing order.
func (s *Store) UpdateEmail(ctx context.Context, orderID, email string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE orders SET email = ?, updated_at = datetime('now') WHERE order_id = ?`,
		email, orderID)
	return err
}

// UpdateProvider updates the payment provider for an order.
func (s *Store) UpdateProvider(ctx context.Context, orderID, provider string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE orders SET provider = ?, updated_at = datetime('now') WHERE order_id = ?`,
		provider, orderID)
	return err
}

// MarkPaidIdempotent marks an order as paid.
// Returns whether this was a new payment, email, product, and chart_json.
func (s *Store) MarkPaidIdempotent(ctx context.Context, orderID, paymentID string) (newPayment bool, email string, product agent.Product, chartJSON string, err error) {
	var e sql.NullString
	var p string
	var cj string
	err = s.db.QueryRowContext(ctx,
		`UPDATE orders SET status = 'paid', payment_id = ?, updated_at = datetime('now') WHERE order_id = ? AND status = 'pending' RETURNING email, product, chart_json`,
		paymentID, orderID).Scan(&e, &p, &cj)
	if err == nil {
		return true, e.String, agent.Product(p), cj, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, "", "", "", err
	}

	// Already paid — read existing data.
	err = s.db.QueryRowContext(ctx,
		`SELECT email, product, chart_json FROM orders WHERE order_id = ? AND status = 'paid'`,
		orderID).Scan(&e, &p, &cj)
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", "", "", ErrOrderNotFound
	}
	if err != nil {
		return false, "", "", "", err
	}
	return false, e.String, agent.Product(p), cj, nil
}

// GetOrder retrieves an order by ID.
func (s *Store) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	var o Order
	var ca, ua string
	var productStr string
	err := s.db.QueryRowContext(ctx,
		`SELECT order_id, product, amount, currency, provider, email, chart_json, llm_json, status, COALESCE(payment_id,''), locale, created_at, updated_at FROM orders WHERE order_id = ?`,
		orderID).Scan(&o.OrderID, &productStr, &o.Amount, &o.Currency, &o.Provider, &o.Email, &o.ChartJSON, &o.LlmJSON, &o.Status, &o.PaymentID, &o.Locale, &ca, &ua)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("payment: scan order: %w", err)
	}
	o.Product = agent.Product(productStr)
		var err2, err3 error
	o.CreatedAt, err2 = time.Parse("2006-01-02 15:04:05", ca)
	o.UpdatedAt, err3 = time.Parse("2006-01-02 15:04:05", ua)
	if err2 != nil || err3 != nil {
		slog.Warn("payment: parse order timestamps", "orderID", orderID, "created_at", ca, "updated_at", ua, "err", fmt.Errorf("created: %w, updated: %w", err2, err3))
	}
	return &o, nil
}

// UpdateLlmJSONIfEmpty atomically sets llm_json only if it is still empty,
// preventing duplicate LLM generations from webhook and report-page race.
func (s *Store) UpdateLlmJSONIfEmpty(ctx context.Context, orderID, llmJSON string) (updated bool, err error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE orders SET llm_json = ?, updated_at = datetime('now') WHERE order_id = ? AND llm_json = ''`,
		llmJSON, orderID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (s *Store) UpdateLlmJSON(ctx context.Context, orderID, llmJSON string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE orders SET llm_json = ?, updated_at = datetime('now') WHERE order_id = ?`,
		llmJSON, orderID)
	return err
}

func (s *Store) CleanStale(ctx context.Context, maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge).UTC().Format("2006-01-02 15:04:05")
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM orders WHERE status = 'pending' AND created_at < ?`, cutoff)
	return err
}

// OpenDB opens a SQLite database at the given path with WAL mode.
func OpenDB(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("payment: create db dir: %w", err)
	}
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)")
	if err != nil {
		return nil, fmt.Errorf("payment: open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	return db, nil
}
