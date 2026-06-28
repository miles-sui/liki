package payment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"liki/internal/product"

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
    chart_json  TEXT NOT NULL DEFAULT '',
    llm_json    TEXT NOT NULL DEFAULT '',
    birth_info  TEXT NOT NULL DEFAULT '',
    chat_expires_at TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'pending',
    payment_id  TEXT,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_orders_email ON orders(email, status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_payment_id ON orders(payment_id) WHERE payment_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_orders_stale ON orders(status, created_at);

CREATE TABLE IF NOT EXISTS chat_messages (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id   TEXT NOT NULL,
    role       TEXT NOT NULL,
    content    TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_chat_messages_order ON chat_messages(order_id, created_at);
`

// dropPdfPath drops the legacy pdf_path column if it exists (from pre-liki schema).
const dropPdfPath = `ALTER TABLE orders DROP COLUMN pdf_path;`

// addLocale adds locale column for multi-language report generation.
const addLocale = `ALTER TABLE orders ADD COLUMN locale TEXT NOT NULL DEFAULT 'zh-Hans';`

// addProvider adds provider column for tracking which payment gateway was used.
const addProvider = `ALTER TABLE orders ADD COLUMN provider TEXT NOT NULL DEFAULT '';`

// addBirthInfo adds birth_info column for naming order birth data.
const addBirthInfo = `ALTER TABLE orders ADD COLUMN birth_info TEXT NOT NULL DEFAULT '';`

// addChatExpiresAt adds chat_expires_at column for 7-day chat window.
const addChatExpiresAt = `ALTER TABLE orders ADD COLUMN chat_expires_at TEXT NOT NULL DEFAULT '';`

// OrderStatus is the payment status of an order.
type OrderStatus string

const (
	OrderPending OrderStatus = "pending"
	OrderPaid    OrderStatus = "paid"
)

// Order holds the database record for a payment order.
type Order struct {
	OrderID       string
	Product       product.Product
	Amount        int
	Currency      string
	Provider      string
	Email         string
	ChartJSON     string
	LlmJSON       string
	BirthInfo     string
	ChatExpiresAt string
	Status        OrderStatus
	PaymentID     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Store provides SQLite-backed order persistence.
type Store struct{ db *sql.DB }

// ChatExpiryDuration is the chat window from payment success.
const ChatExpiryDuration = 7 * 24 * time.Hour

// DefaultChatExpiry returns the expiry time for a newly paid naming order.
func DefaultChatExpiry() string {
	return time.Now().Add(ChatExpiryDuration).UTC().Format(time.DateTime)
}

// migrateColumn runs an ALTER TABLE and logs duplicate-column errors as info.
func migrateColumn(db *sql.DB, sql, name string) {
	if _, err := db.Exec(sql); err != nil {
		if strings.Contains(err.Error(), "duplicate column") || strings.Contains(err.Error(), "already exists") {
			slog.Info("payment: add "+name+" (expected on upgraded instances)", "err", err)
		} else {
			slog.Error("payment: add "+name+" failed", "err", err)
		}
	}
}

// NewStore creates the orders table and returns a new Store.
func NewStore(db *sql.DB) (*Store, error) {
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("payment: create table: %w", err)
	}
	// Drop legacy pdf_path column from pre-liki schema. Column may not
	// exist on fresh installs — failure is expected and harmless.
	if _, err := db.Exec(dropPdfPath); err != nil {
		if strings.Contains(err.Error(), "no such table") || strings.Contains(err.Error(), "no such column") {
			slog.Info("payment: drop pdf_path (expected on fresh installs)", "err", err)
		} else {
			slog.Error("payment: drop pdf_path failed", "err", err)
		}
	}
	migrateColumn(db, addLocale, "locale")
	migrateColumn(db, addProvider, "provider")
	migrateColumn(db, addBirthInfo, "birth_info")
	migrateColumn(db, addChatExpiresAt, "chat_expires_at")
	return &Store{db: db}, nil
}

// CreateOrder inserts a new pending order into the database.
func (s *Store) CreateOrder(ctx context.Context, orderID string, product product.Product, amount int, currency, email, chartJSON, llmJSON, provider string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO orders (order_id, product, amount, currency, email, chart_json, llm_json, provider) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		orderID, product, amount, currency, email, chartJSON, llmJSON, provider)
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
	if err != nil {
		return fmt.Errorf("payment: update email: %w", err)
	}
	return nil
}

// UpdateProvider updates the payment provider for an order.
func (s *Store) UpdateProvider(ctx context.Context, orderID, provider string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE orders SET provider = ?, updated_at = datetime('now') WHERE order_id = ?`,
		provider, orderID)
	if err != nil {
		return fmt.Errorf("payment: update provider: %w", err)
	}
	return nil
}

// MarkPaidIdempotent marks an order as paid.
// Returns whether this was a new payment, email, and product.
func (s *Store) MarkPaidIdempotent(ctx context.Context, orderID, paymentID string) (bool, string, product.Product, error) {
	var e sql.NullString
	var p string
	err := s.db.QueryRowContext(ctx,
		`UPDATE orders SET status = 'paid', payment_id = ?, chat_expires_at = COALESCE(NULLIF(chat_expires_at,''), datetime('now', '+7 days')), updated_at = datetime('now') WHERE order_id = ? AND status = 'pending' RETURNING email, product`,
		paymentID, orderID).Scan(&e, &p)
	if err == nil {
		return true, e.String, product.Product(p), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, "", "", err
	}

	// Already paid — read existing data.
	err = s.db.QueryRowContext(ctx,
		`SELECT email, product FROM orders WHERE order_id = ? AND status = 'paid'`,
		orderID).Scan(&e, &p)
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", "", ErrOrderNotFound
	}
	if err != nil {
		return false, "", "", err
	}
	return false, e.String, product.Product(p), nil
}

// GetOrder retrieves an order by ID.
func (s *Store) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	var o Order
	var ca, ua string
	var productStr string
	err := s.db.QueryRowContext(ctx,
		`SELECT order_id, product, amount, currency, provider, email, chart_json, llm_json, birth_info, chat_expires_at, status, COALESCE(payment_id,''), created_at, updated_at FROM orders WHERE order_id = ?`,
		orderID).Scan(&o.OrderID, &productStr, &o.Amount, &o.Currency, &o.Provider, &o.Email, &o.ChartJSON, &o.LlmJSON, &o.BirthInfo, &o.ChatExpiresAt, &o.Status, &o.PaymentID, &ca, &ua)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("payment: scan order: %w", err)
	}
	o.Product = product.Product(productStr)
		var err2, err3 error
	o.CreatedAt, err2 = time.Parse(time.DateTime, ca)
	o.UpdatedAt, err3 = time.Parse(time.DateTime, ua)
	if err2 != nil || err3 != nil {
		slog.Warn("payment: parse order timestamps", "orderID", orderID, "created_at", ca, "updated_at", ua, "err", fmt.Errorf("created: %w, updated: %w", err2, err3))
	}
	return &o, nil
}

// SetChatExpiresAtIfEmpty sets chat_expires_at for an order only if it is still empty.
// Used by both webhook and callback — first one wins.
func (s *Store) SetChatExpiresAtIfEmpty(ctx context.Context, orderID, expiresAt string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE orders SET chat_expires_at = ?, updated_at = datetime('now') WHERE order_id = ? AND chat_expires_at = ''`,
		expiresAt, orderID)
	if err != nil {
		return fmt.Errorf("payment: set chat expires: %w", err)
	}
	return nil
}

// OverrideChatExpiresAt unconditionally sets chat_expires_at for an order.
// Used in tests to simulate expired orders after MarkPaidIdempotent now sets it atomically.
func (s *Store) OverrideChatExpiresAt(ctx context.Context, orderID, expiresAt string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE orders SET chat_expires_at = ?, updated_at = datetime('now') WHERE order_id = ?`,
		expiresAt, orderID)
	if err != nil {
		return fmt.Errorf("payment: override chat expires: %w", err)
	}
	return nil
}

// UpdateBirthInfoIfEmpty atomically sets birth_info only if it is still empty.
func (s *Store) UpdateBirthInfoIfEmpty(ctx context.Context, orderID, birthInfo string) (updated bool, err error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE orders SET birth_info = ?, updated_at = datetime('now') WHERE order_id = ? AND birth_info = ''`,
		birthInfo, orderID)
	if err != nil {
		return false, fmt.Errorf("payment: update birth info: %w", err)
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// FindActiveOrdersByEmail finds paid, unexpired orders for an email.
func (s *Store) FindActiveOrdersByEmail(ctx context.Context, email string) ([]Order, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT order_id, product, amount, currency, provider, email, chart_json, llm_json, birth_info, chat_expires_at, status, COALESCE(payment_id,''), created_at, updated_at FROM orders WHERE email = ? AND status = 'paid' AND chat_expires_at > datetime('now') ORDER BY created_at DESC`,
		email)
	if err != nil {
		return nil, fmt.Errorf("payment: find orders by email: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		var ca, ua, productStr string
		if err := rows.Scan(&o.OrderID, &productStr, &o.Amount, &o.Currency, &o.Provider, &o.Email, &o.ChartJSON, &o.LlmJSON, &o.BirthInfo, &o.ChatExpiresAt, &o.Status, &o.PaymentID, &ca, &ua); err != nil {
			return nil, fmt.Errorf("payment: scan order by email: %w", err)
		}
		o.Product = product.Product(productStr)
		var errCA, errUA error
		o.CreatedAt, errCA = time.Parse(time.DateTime, ca)
		o.UpdatedAt, errUA = time.Parse(time.DateTime, ua)
		if errCA != nil || errUA != nil {
			return nil, fmt.Errorf("payment: parse order time: created_at=%w, updated_at=%w", errCA, errUA)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("payment: iter orders by email: %w", err)
	}
	return orders, nil
}

// CreateChatMessage inserts a chat message for an order.
func (s *Store) CreateChatMessage(ctx context.Context, orderID string, role Role, content string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO chat_messages (order_id, role, content) VALUES (?, ?, ?)`,
		orderID, role, content)
	if err != nil {
		return fmt.Errorf("payment: create chat message: %w", err)
	}
	return nil
}

// BatchCreateChatMessages inserts multiple chat messages in a single statement.
func (s *Store) BatchCreateChatMessages(ctx context.Context, orderID string, msgs []ChatMessage) error {
	if len(msgs) == 0 {
		return nil
	}
	stmt := `INSERT INTO chat_messages (order_id, role, content) VALUES `
	args := make([]any, 0, len(msgs)*3)
	for i, m := range msgs {
		if i > 0 {
			stmt += ", "
		}
		stmt += "(?, ?, ?)"
		args = append(args, orderID, m.Role, m.Content)
	}
	_, err := s.db.ExecContext(ctx, stmt, args...)
	if err != nil {
		return fmt.Errorf("payment: batch create chat messages: %w", err)
	}
	return nil
}

// LoadChatHistory loads all chat messages for an order, ordered by creation time.
func (s *Store) LoadChatHistory(ctx context.Context, orderID string) ([]ChatMessage, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT role, content FROM chat_messages WHERE order_id = ? ORDER BY created_at ASC`,
		orderID)
	if err != nil {
		return nil, fmt.Errorf("payment: load chat history: %w", err)
	}
	defer rows.Close()

	var msgs []ChatMessage
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.Role, &m.Content); err != nil {
			return nil, fmt.Errorf("payment: scan chat message: %w", err)
		}
		msgs = append(msgs, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("payment: iter chat messages: %w", err)
	}
	return msgs, nil
}

// Role is the chat message role.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ChatMessage holds a single persisted chat message.
type ChatMessage struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

func (s *Store) UpdateLlmJSON(ctx context.Context, orderID, llmJSON string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE orders SET llm_json = ?, updated_at = datetime('now') WHERE order_id = ?`,
		llmJSON, orderID)
	if err != nil {
		return fmt.Errorf("payment: update llm json: %w", err)
	}
	return nil
}

func (s *Store) CleanStale(ctx context.Context, maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge).UTC().Format(time.DateTime)
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM orders WHERE status = 'pending' AND created_at < ?`, cutoff)
	if err != nil {
		return fmt.Errorf("payment: clean stale orders: %w", err)
	}
	return nil
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
