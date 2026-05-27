package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/25types/25types/internal/app/application/commerce"
)

// DonationRepo persists donation records.
type DonationRepo struct {
	db *sql.DB
}

// NewDonationRepo creates a DonationRepo.
func NewDonationRepo(db *sql.DB) *DonationRepo {
	return &DonationRepo{db: db}
}

// CreateDonation inserts a new donation record. paymentID is the Dodo payment ID
// used as an idempotency key — UNIQUE constraint prevents duplicate processing.
func (r *DonationRepo) CreateDonation(ctx context.Context, userID int64, amount int, paymentID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO donations (user_id, amount, payment_id) VALUES (?, ?, ?)`,
		userID, amount, paymentID)
	if err != nil {
		return fmt.Errorf("CreateDonation: %w", err)
	}
	return nil
}

var _ commerce.DonationRepository = (*DonationRepo)(nil)
