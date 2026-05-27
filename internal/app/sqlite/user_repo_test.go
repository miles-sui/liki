package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/25types/25types/internal/app/application/user"
	"github.com/25types/25types/internal/app/db"
	"github.com/25types/25types/internal/app/domain"
)

// openTestDB creates an in-memory SQLite database with all migrations applied.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database, err := db.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func newTestUserRepo(t *testing.T) *UserRepo {
	t.Helper()
	return NewUserRepo(openTestDB(t))
}

func createTestUser(t *testing.T, repo *UserRepo, name string) int64 {
	t.Helper()
	h := PasswordHasher{}
	hash, _ := h.Hash("testpass123")
	id, err := repo.Create(context.Background(), &domain.User{Name: name, PasswordHash: hash})
	if err != nil {
		t.Fatalf("Create(%q): %v", name, err)
	}
	return id
}

// =============================================================================
// Create + FindByName + FindByID
// =============================================================================

func TestUserRepo_CreateAndFind(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()

	// Create — spec: name 1-64 chars, password argon2id.
	h := PasswordHasher{}
	hash, _ := h.Hash("secret123")
	id, err := repo.Create(ctx, &domain.User{Name: "alice", PasswordHash: hash})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}

	// FindByName.
	u, err := repo.FindByName(ctx, "alice")
	if err != nil {
		t.Fatalf("FindByName: %v", err)
	}
	if u.ID != id {
		t.Errorf("ID = %d, want %d", u.ID, id)
	}
	if u.Name != "alice" {
		t.Errorf("Name = %q, want alice", u.Name)
	}
	if u.TokenVersion != 1 {
		t.Errorf("TokenVersion = %d, want 1 (initial)", u.TokenVersion)
	}

	// FindByID.
	u2, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if u2.Name != "alice" {
		t.Errorf("Name = %q, want alice", u2.Name)
	}
}

func TestUserRepo_CreateDuplicate(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()

	h := PasswordHasher{}
	hash, _ := h.Hash("pw12345678")
	repo.Create(ctx, &domain.User{Name: "bob", PasswordHash: hash})

	_, err := repo.Create(ctx, &domain.User{Name: "bob", PasswordHash: hash})
	if !errors.Is(err, domain.ErrUsernameTaken) {
		t.Errorf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestUserRepo_FindByName_NotFound(t *testing.T) {
	repo := newTestUserRepo(t)
	_, err := repo.FindByName(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent user")
	}
}

func TestUserRepo_FindByID_NotFound(t *testing.T) {
	repo := newTestUserRepo(t)
	_, err := repo.FindByID(context.Background(), 99999)
	if err == nil {
		t.Error("expected error for nonexistent user ID")
	}
}

// =============================================================================
// FindByEmail — spec: only returns users with verified email
// =============================================================================

func TestUserRepo_FindByEmail(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()
	id := createTestUser(t, repo, "email-find")

	// Verify email.
	now := time.Now().UTC()
	repo.db.ExecContext(ctx,
		`UPDATE users SET email = 'find@example.com', email_verified_at = ? WHERE id = ?`,
		now.Format(time.RFC3339), id)

	u, err := repo.FindByEmail(ctx, "find@example.com")
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if u.ID != id {
		t.Errorf("ID = %d, want %d", u.ID, id)
	}
}

func TestUserRepo_FindByEmail_NotVerified(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()
	id := createTestUser(t, repo, "unverified-email")

	// Set email but NOT verified.
	repo.db.ExecContext(ctx,
		`UPDATE users SET email = 'unverified@example.com' WHERE id = ?`, id)

	_, err := repo.FindByEmail(ctx, "unverified@example.com")
	if err == nil {
		t.Error("FindByEmail should not find user with unverified email")
	}
}

func TestUserRepo_FindByEmail_NotFound(t *testing.T) {
	repo := newTestUserRepo(t)
	_, err := repo.FindByEmail(context.Background(), "nonexistent@example.com")
	if err == nil {
		t.Error("expected error for nonexistent email")
	}
}

// =============================================================================
// UpdateTokenVersion — spec: token_version + 1 on password change / logout
// =============================================================================

func TestUserRepo_UpdateTokenVersion(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()
	id := createTestUser(t, repo, "token-test")

	newVer, err := repo.UpdateTokenVersion(ctx, id)
	if err != nil {
		t.Fatalf("UpdateTokenVersion: %v", err)
	}
	if newVer != 2 {
		t.Errorf("version = %d, want 2", newVer)
	}

	// Second increment.
	newVer2, _ := repo.UpdateTokenVersion(ctx, id)
	if newVer2 != 3 {
		t.Errorf("version = %d, want 3", newVer2)
	}
}

// =============================================================================
// UpdatePasswordHash — spec: new hash + token_version + 1
// =============================================================================

func TestUserRepo_UpdatePasswordHash(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()
	id := createTestUser(t, repo, "pw-test")

	h := PasswordHasher{}
	newHash, _ := h.Hash("newpassword456")
	newVer, err := repo.UpdatePasswordHash(ctx, id, newHash)
	if err != nil {
		t.Fatalf("UpdatePasswordHash: %v", err)
	}
	if newVer != 2 {
		t.Errorf("version = %d, want 2 (incremented)", newVer)
	}

	// Verify the hash changed.
	u, _ := repo.FindByID(ctx, id)
	if u.PasswordHash != newHash {
		t.Error("password hash was not updated")
	}
}

// =============================================================================
// UpdateFields — spec: partial update, name/email/is_public
// =============================================================================

func TestUserRepo_UpdateFields(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()
	id := createTestUser(t, repo, "fields-test")

	// Update name.
	newName := "renamed"
	err := repo.UpdateFields(ctx, id, user.UpdateUserFields{Name: &newName})
	if err != nil {
		t.Fatalf("UpdateFields: %v", err)
	}
	u, _ := repo.FindByID(ctx, id)
	if u.Name != "renamed" {
		t.Errorf("Name = %q, want renamed", u.Name)
	}
}

func TestUserRepo_UpdateFields_DuplicateName(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()
	createTestUser(t, repo, "bob")
	id := createTestUser(t, repo, "alice")

	newName := "bob"
	err := repo.UpdateFields(ctx, id, user.UpdateUserFields{Name: &newName})
	if !errors.Is(err, domain.ErrUsernameTaken) {
		t.Errorf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestUserRepo_UpdateFields_Email(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()
	id := createTestUser(t, repo, "email-test")

	newEmail := "new@example.com"
	err := repo.UpdateFields(ctx, id, user.UpdateUserFields{Email: &newEmail})
	if err != nil {
		t.Fatalf("UpdateFields email: %v", err)
	}
	u, _ := repo.FindByID(ctx, id)
	if u.PendingEmail == nil || *u.PendingEmail != "new@example.com" {
		t.Errorf("PendingEmail = %v, want new@example.com", u.PendingEmail)
	}
	// Per spec: email field keeps old value until verification.
	if u.Email != "" {
		t.Errorf("Email should stay empty until verified, got %q", u.Email)
	}
}

func TestUserRepo_UpdateFields_IsPublic(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()
	id := createTestUser(t, repo, "public-test")

	public := true
	err := repo.UpdateFields(ctx, id, user.UpdateUserFields{IsPublic: &public})
	if err != nil {
		t.Fatalf("UpdateFields is_public: %v", err)
	}
	u, _ := repo.FindByID(ctx, id)
	if !u.IsPublic {
		t.Errorf("IsPublic = %v, want true", u.IsPublic)
	}
}

// =============================================================================
// SetDeactivated — spec: token_version + 1, deactivated_at = now
// =============================================================================

func TestUserRepo_SetDeactivated(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()
	id := createTestUser(t, repo, "deactivate-me")

	now := time.Now().UTC()
	err := repo.SetDeactivated(ctx, id, now)
	if err != nil {
		t.Fatalf("SetDeactivated: %v", err)
	}

	u, _ := repo.FindByID(ctx, id)
	if u.DeactivatedAt == nil {
		t.Error("expected DeactivatedAt to be set")
	}
	// FindByName should still work (row exists, just deactivated).
	u2, err := repo.FindByName(ctx, "deactivate-me")
	if err != nil {
		t.Errorf("FindByName should still find deactivated user: %v", err)
	}
	if u2.DeactivatedAt == nil {
		t.Error("expected DeactivatedAt after FindByName")
	}
}

// =============================================================================
// Password reset flow — spec: SetPasswordResetToken → FindByPasswordResetToken → ResetPassword
// =============================================================================

func TestUserRepo_PasswordResetFlow(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()
	id := createTestUser(t, repo, "reset-me")

	// First verify email to allow password reset (SetPasswordResetToken requires email_verified_at).
	repo.db.ExecContext(ctx,
		`UPDATE users SET email = 'reset@test.com', email_verified_at = ? WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), id)

	token := "reset-token-hex-12345"
	exp := time.Now().UTC().Add(15 * time.Minute)
	err := repo.SetPasswordResetToken(ctx, "reset@test.com", token, exp)
	if err != nil {
		t.Fatalf("SetPasswordResetToken: %v", err)
	}

	// Find by valid token.
	foundID, err := repo.FindByPasswordResetToken(ctx, token)
	if err != nil {
		t.Fatalf("FindByPasswordResetToken: %v", err)
	}
	if foundID != id {
		t.Errorf("FindByPasswordResetToken ID = %d, want %d", foundID, id)
	}

	// Reset password.
	h := PasswordHasher{}
	newHash, _ := h.Hash("brand-new-password")
	err = repo.ResetPassword(ctx, id, newHash)
	if err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}

	// Token should be consumed — find again should fail.
	_, err = repo.FindByPasswordResetToken(ctx, token)
	if !errors.Is(err, domain.ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired after reset, got %v", err)
	}

	// Token version should be incremented.
	u, _ := repo.FindByID(ctx, id)
	if u.TokenVersion != 2 {
		t.Errorf("TokenVersion = %d, want 2 after reset", u.TokenVersion)
	}
	if u.PasswordHash != newHash {
		t.Error("password hash should be updated")
	}
}

func TestUserRepo_FindByPasswordResetToken_Expired(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()
	id := createTestUser(t, repo, "expired-reset")

	repo.db.ExecContext(ctx,
		`UPDATE users SET email = 'exp@test.com', email_verified_at = ? WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), id)

	// Set token with past expiration.
	exp := time.Now().UTC().Add(-1 * time.Hour)
	repo.SetPasswordResetToken(ctx, "exp@test.com", "expired-token", exp)

	_, err := repo.FindByPasswordResetToken(ctx, "expired-token")
	if !errors.Is(err, domain.ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired for expired token, got %v", err)
	}
}

// =============================================================================
// TokenValidator — spec: GetTokenVersion
// =============================================================================

func TestUserRepo_GetTokenVersion(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()
	id := createTestUser(t, repo, "version-test")

	tv, err := repo.GetTokenVersion(ctx, id)
	if err != nil {
		t.Fatalf("GetTokenVersion: %v", err)
	}
	if tv != 1 {
		t.Errorf("initial version = %d, want 1", tv)
	}

	repo.UpdateTokenVersion(ctx, id)
	tv, _ = repo.GetTokenVersion(ctx, id)
	if tv != 2 {
		t.Errorf("version after increment = %d, want 2", tv)
	}
}

func TestUserRepo_GetTokenVersion_Deactivated(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()
	id := createTestUser(t, repo, "deact-ver")

	repo.SetDeactivated(ctx, id, time.Now().UTC())

	_, err := repo.GetTokenVersion(ctx, id)
	if err == nil {
		t.Error("GetTokenVersion should fail for deactivated user")
	}
}

// =============================================================================// VerifyEmailByToken
// =============================================================================

func TestUserRepo_VerifyEmailByToken(t *testing.T) {
	repo := newTestUserRepo(t)
	ctx := context.Background()
	id := createTestUser(t, repo, "verify-me")

	token := "verify-hex-token-abc"
	repo.db.ExecContext(ctx,
		`UPDATE users SET pending_email = 'new@test.com' WHERE id = ?`, id)
	repo.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO user_tokens (user_id, token_type, token, expires_at)
		 VALUES (?, 'email_verify', ?, ?)`,
		id, token, time.Now().Add(24*time.Hour).UTC().Format(time.RFC3339))

	err := repo.VerifyEmailByToken(ctx, token)
	if err != nil {
		t.Fatalf("VerifyEmailByToken: %v", err)
	}

	u, _ := repo.FindByID(ctx, id)
	if u.Email != "new@test.com" {
		t.Errorf("Email = %q, want new@test.com (pending→email)", u.Email)
	}
	if u.PendingEmail != nil && *u.PendingEmail != "" {
		t.Error("PendingEmail should be cleared after verification")
	}
}

func TestUserRepo_VerifyEmailByToken_Invalid(t *testing.T) {
	repo := newTestUserRepo(t)
	err := repo.VerifyEmailByToken(context.Background(), "bogus-token")
	if !errors.Is(err, domain.ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired for invalid token, got %v", err)
	}
}

// =============================================================================
// Export
// =============================================================================

func TestUserRepo_GetExportAssessments_Empty(t *testing.T) {
	repo := newTestUserRepo(t)
	id := createTestUser(t, repo, "export-empty")

	as, err := repo.GetExportAssessments(context.Background(), id)
	if err != nil {
		t.Fatalf("GetExportAssessments: %v", err)
	}
	if len(as) != 0 {
		t.Errorf("expected 0 assessments, got %d", len(as))
	}
}

// =============================================================================
// parseNullTime / nullStrToStr — unit tests
// =============================================================================

func TestParseNullTime(t *testing.T) {
	rfc := "2024-01-15T10:30:00Z"
	expected := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name  string
		input sql.NullString
		want  *time.Time
	}{
		{"not valid", sql.NullString{Valid: false, String: ""}, nil},
		{"empty string", sql.NullString{Valid: true, String: ""}, nil},
		{"RFC3339", sql.NullString{Valid: true, String: rfc}, &expected},
		{"non-RFC3339", sql.NullString{Valid: true, String: "2024-01-15 10:30:00"}, nil},
		{"garbage", sql.NullString{Valid: true, String: "not-a-date"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseNullTime(tt.input)
			if tt.want == nil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}
			if got == nil {
				t.Errorf("expected %v, got nil", *tt.want)
				return
			}
			if !got.Equal(*tt.want) {
				t.Errorf("expected %v, got %v", *tt.want, *got)
			}
		})
	}
}

func TestNullStrToStr(t *testing.T) {
	tests := []struct {
		name  string
		input sql.NullString
		want  *string
	}{
		{"not valid", sql.NullString{Valid: false, String: ""}, nil},
		{"empty", sql.NullString{Valid: true, String: ""}, nil},
		{"value", sql.NullString{Valid: true, String: "hello"}, strPtr("hello")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nullStrToStr(tt.input)
			if tt.want == nil {
				if got != nil {
					t.Errorf("expected nil, got %v", *got)
				}
				return
			}
			if got == nil || *got != *tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func strPtr(s string) *string { return &s }

// =============================================================================
// SupporterSince — E2E via FindByID
// =============================================================================

func TestUserRepo_SupporterSince(t *testing.T) {
	repo := newTestUserRepo(t)
	donationRepo := NewDonationRepo(openTestDB(t))
	ctx := context.Background()

	userID := createTestUser(t, repo, "supporter-e2e")

	// Before donation: supporter_since should be nil.
	u, err := repo.FindByID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if u.SupporterSince != nil {
		t.Error("expected nil SupporterSince before any donation")
	}

	// Insert a donation.
	if err := donationRepo.CreateDonation(ctx, userID, 1990, "pay_supporter_test"); err != nil {
		t.Fatalf("CreateDonation: %v", err)
	}

	// After donation: supporter_since should be set.
	u, err = repo.FindByID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if u.SupporterSince == nil {
		t.Fatal("expected non-nil SupporterSince after donation")
	}
	if u.SupporterSince.IsZero() {
		t.Error("SupporterSince is zero time")
	}
}

func TestUserRepo_GetExportReviewLinks_Empty(t *testing.T) {
	repo := newTestUserRepo(t)
	id := createTestUser(t, repo, "export-rl")

	rl, err := repo.GetExportReviewLinks(context.Background(), id)
	if err != nil {
		t.Fatalf("GetExportReviewLinks: %v", err)
	}
	if len(rl) != 0 {
		t.Errorf("expected 0 review links, got %d", len(rl))
	}
}
