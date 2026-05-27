package sqlite

import (
	"context"
	"testing"
	"time"
)

func TestDonationRepo_CreateDonation(t *testing.T) {
	repo := NewDonationRepo(openTestDB(t))
	userRepo := newTestUserRepo(t)
	ctx := context.Background()

	userID := createTestUser(t, userRepo, "donor1")

	// Create a donation
	err := repo.CreateDonation(ctx, userID, 990, "pay_test_1")
	if err != nil {
		t.Fatalf("CreateDonation: %v", err)
	}

	// Verify it's in the database
	var count int
	err = repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM donations WHERE user_id = ? AND amount = 990 AND payment_id = 'pay_test_1'", userID).Scan(&count)
	if err != nil {
		t.Fatalf("query donations: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 donation, got %d", count)
	}
}

func TestDonationRepo_MultipleDonations(t *testing.T) {
	repo := NewDonationRepo(openTestDB(t))
	userRepo := newTestUserRepo(t)
	ctx := context.Background()

	userID := createTestUser(t, userRepo, "donor2")

	// Create multiple donations (same user, different amounts)
	pids := []string{"pay_a", "pay_b", "pay_c"}
	for i, amount := range []int{990, 1990, 2990} {
		if err := repo.CreateDonation(ctx, userID, amount, pids[i]); err != nil {
			t.Fatalf("CreateDonation(%d): %v", amount, err)
		}
	}

	var count int
	err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM donations WHERE user_id = ?", userID).Scan(&count)
	if err != nil {
		t.Fatalf("query donations: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 donations, got %d", count)
	}
}

func TestDonationRepo_CreatedAtRFC3339(t *testing.T) {
	repo := NewDonationRepo(openTestDB(t))
	userRepo := newTestUserRepo(t)
	ctx := context.Background()

	userID := createTestUser(t, userRepo, "rfc3339-donor")
	if err := repo.CreateDonation(ctx, userID, 1990, "pay_rfc3339"); err != nil {
		t.Fatalf("CreateDonation: %v", err)
	}

	// Verify created_at is stored in RFC3339 format (parseable by time.Parse).
	var createdStr string
	err := repo.db.QueryRowContext(ctx, "SELECT created_at FROM donations WHERE user_id = ?", userID).Scan(&createdStr)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	tm, err := time.Parse(time.RFC3339, createdStr)
	if err != nil {
		t.Fatalf("created_at %q is not RFC3339: %v", createdStr, err)
	}
	if time.Since(tm) > 10*time.Second {
		t.Errorf("donation time too old: %v", tm)
	}
}

func TestDonationRepo_ImplementsInterface(t *testing.T) {
	// Compile-time check — if this compiles, the interface is satisfied.
	var _ = NewDonationRepo(openTestDB(t))
}
