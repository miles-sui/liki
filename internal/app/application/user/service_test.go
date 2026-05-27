package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/25types/25types/internal/app/domain"
)

// =============================================================================
// Stubs
// =============================================================================

type stubRepo struct {
	users             map[int64]*domain.User
	names             map[string]*domain.User
	emails            map[string]*domain.User // verified email → user
	pendingEmails     map[string]*domain.User // pending_email → user
	nextID            int64
	resetTokens       map[string]resetTokenEntry
	createErr         error
	findByNameErr     error
	findByEmailErr    error
	findByPendingErr  error
}

type resetTokenEntry struct {
	userID int64
	exp    time.Time
}

func newStubRepo() *stubRepo {
	return &stubRepo{
		users:         map[int64]*domain.User{},
		names:         map[string]*domain.User{},
		emails:        map[string]*domain.User{},
		pendingEmails: map[string]*domain.User{},
		resetTokens:   map[string]resetTokenEntry{},
	}
}

func (r *stubRepo) Create(ctx context.Context, u *domain.User) (int64, error) {
	if r.createErr != nil {
		return 0, r.createErr
	}
	if _, ok := r.names[u.Name]; ok {
		return 0, domain.ErrUsernameTaken
	}
	r.nextID++
	u.ID = r.nextID
	r.users[u.ID] = u
	r.names[u.Name] = u
	if u.Email != "" && u.EmailVerifiedAt != nil {
		r.emails[u.Email] = u
	}
	if u.PendingEmail != nil && *u.PendingEmail != "" {
		r.pendingEmails[*u.PendingEmail] = u
	}
	return u.ID, nil
}

func (r *stubRepo) FindByName(ctx context.Context, name string) (*domain.User, error) {
	if r.findByNameErr != nil {
		return nil, r.findByNameErr
	}
	u, ok := r.names[name]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (r *stubRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if r.findByEmailErr != nil {
		return nil, r.findByEmailErr
	}
	u, ok := r.emails[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *stubRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	_, ok := r.emails[email]
	return ok, nil
}

func (r *stubRepo) FindByPendingEmail(ctx context.Context, email string) (*domain.User, error) {
	if r.findByPendingErr != nil {
		return nil, r.findByPendingErr
	}
	u, ok := r.pendingEmails[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *stubRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (r *stubRepo) UpdateTokenVersion(ctx context.Context, id int64) (int, error) {
	u := r.users[id]
	u.TokenVersion++
	return u.TokenVersion, nil
}

func (r *stubRepo) UpdatePasswordHash(ctx context.Context, id int64, hash string) (int, error) {
	u := r.users[id]
	u.PasswordHash = hash
	u.TokenVersion++
	return u.TokenVersion, nil
}

func (r *stubRepo) UpdateFields(ctx context.Context, id int64, fields UpdateUserFields) error {
	u := r.users[id]
	if u == nil {
		return errors.New("user not found")
	}
	if fields.Name != nil {
		u.Name = *fields.Name
	}
	if fields.Email != nil {
		u.PendingEmail = fields.Email
	}
	if fields.EmailVerToken != nil {
		// Store token — VerifyEmailByToken will check this.
	}
	if fields.IsPublic != nil {
		u.IsPublic = *fields.IsPublic
	}
	return nil
}

func (r *stubRepo) SetDeactivated(ctx context.Context, id int64, at time.Time) error {
	u := r.users[id]
	u.DeactivatedAt = &at
	return nil
}

func (r *stubRepo) ReactivateUser(ctx context.Context, id int64) error {
	u := r.users[id]
	u.DeactivatedAt = nil
	return nil
}

func (r *stubRepo) GetExportAssessments(ctx context.Context, userID int64) ([]ExportAssessment, error) {
	return nil, nil
}

func (r *stubRepo) GetExportReviewLinks(ctx context.Context, userID int64) ([]ExportReviewLink, error) {
	return nil, nil
}

func (r *stubRepo) VerifyEmailByToken(ctx context.Context, token string) error { return nil }
func (r *stubRepo) SetPasswordResetToken(ctx context.Context, email, token string, exp time.Time) error {
	r.resetTokens[token] = resetTokenEntry{exp: exp}
	return nil
}
func (r *stubRepo) FindByPasswordResetToken(ctx context.Context, token string) (int64, error) {
	entry, ok := r.resetTokens[token]
	if !ok || time.Now().UTC().After(entry.exp) {
		return 0, domain.ErrTokenExpired
	}
	return entry.userID, nil
}
func (r *stubRepo) ResetPassword(ctx context.Context, id int64, hash string) error {
	r.users[id].PasswordHash = hash
	return nil
}
func (r *stubRepo) DeleteUser(ctx context.Context, id int64) error {
	delete(r.users, id)
	return nil
}

type stubHasher struct {
	hash   string
	verify bool
	rehash string
}

func (h *stubHasher) Hash(password string) (string, error) { return h.hash, nil }
func (h *stubHasher) Verify(password, storedHash string) (bool, string) {
	return h.verify, h.rehash
}

type stubClaimer struct{ claimed []int64 }

func (c *stubClaimer) ClaimAnonymous(ctx context.Context, userID int64, token string) (int64, error) {
	c.claimed = append(c.claimed, userID)
	return 1, nil
}

type stubEmailSender struct{ sent []string }

func (s *stubEmailSender) SendVerificationEmail(ctx context.Context, to, token, locale string) error {
	s.sent = append(s.sent, "verify:"+to)
	return nil
}
func (s *stubEmailSender) SendPasswordResetEmail(ctx context.Context, to, token, locale string) error {
	s.sent = append(s.sent, "reset:"+to)
	return nil
}
func (s *stubEmailSender) SendBondNotification(ctx context.Context, to, otherName, creatorName, locale string) error {
	s.sent = append(s.sent, "bond:"+to)
	return nil
}

func makeTokenFn(token string) func(int64, int, string) (string, error) {
	return func(userID int64, tv int, name string) (string, error) {
		return token, nil
	}
}

// =============================================================================
// RegisterUseCase
// =============================================================================

func TestRegisterUseCase(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{hash: "hashed-pw"}
	claim := &stubClaimer{}
	ctx := context.Background()

	out, err := RegisterUseCase(ctx, repo, claim, hasher, nil, makeTokenFn("jwt-token"),
		RegisterUseCaseInput{Name: "alice", Email: "alice@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if out.Token != "jwt-token" {
		t.Errorf("Token = %q, want jwt-token", out.Token)
	}
	if out.User.Name != "alice" {
		t.Errorf("Name = %q, want alice", out.User.Name)
	}
}

func TestRegisterUseCase_Validation(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{hash: "h"}
	_, err := RegisterUseCase(context.Background(), repo, nil, hasher, nil, nil,
		RegisterUseCaseInput{Name: "", Email: "", Password: ""})
	if !errors.Is(err, domain.ErrNameAndPasswordRequired) {
		t.Errorf("expected ErrNameAndPasswordRequired, got %v", err)
	}
}

func TestRegisterUseCase_Duplicate(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{hash: "h"}
	ctx := context.Background()
	RegisterUseCase(ctx, repo, &stubClaimer{}, hasher, nil, makeTokenFn("x"),
		RegisterUseCaseInput{Name: "bob", Email: "bob@example.com", Password: "password123"})
	_, err := RegisterUseCase(ctx, repo, &stubClaimer{}, hasher, nil, makeTokenFn("x"),
		RegisterUseCaseInput{Name: "bob", Email: "bob2@example.com", Password: "password123"})
	if !errors.Is(err, domain.ErrUsernameTaken) {
		t.Errorf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestRegisterUseCase_WithEmail(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{hash: "h"}
	sender := &stubEmailSender{}
	ctx := context.Background()

	out, err := RegisterUseCase(ctx, repo, &stubClaimer{}, hasher, sender, makeTokenFn("x"),
		RegisterUseCaseInput{Name: "alice", Email: "alice@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if out.User.Email != "" {
		t.Errorf("Email should be empty until verified, got %q", out.User.Email)
	}
	if out.User.EmailVerified {
		t.Error("EmailVerified should be false until verified")
	}
	// Verify email was stored as pending.
	u, _ := repo.FindByID(ctx, out.User.ID)
	if u.PendingEmail == nil || *u.PendingEmail != "alice@example.com" {
		t.Errorf("PendingEmail = %v, want alice@example.com", u.PendingEmail)
	}
	// Verify email was sent.
	if len(sender.sent) != 1 || sender.sent[0] != "verify:alice@example.com" {
		t.Errorf("expected verification email, got %v", sender.sent)
	}
}

func TestRegisterUseCase_EmailAlreadyVerifiedByOther(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{hash: "h"}
	ctx := context.Background()

	// Create user with verified email.
	now := time.Now()
	repo.Create(ctx, &domain.User{Name: "existing", PasswordHash: "h", Email: "dup@example.com", EmailVerifiedAt: &now})

	_, err := RegisterUseCase(ctx, repo, &stubClaimer{}, hasher, nil, makeTokenFn("x"),
		RegisterUseCaseInput{Name: "newbie", Email: "dup@example.com", Password: "password123"})
	if !errors.Is(err, domain.ErrEmailAlreadyVerified) {
		t.Errorf("expected ErrEmailAlreadyVerified, got %v", err)
	}
}

func TestRegisterUseCase_InvalidEmail(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{hash: "h"}
	ctx := context.Background()

	tests := []struct {
		email    string
		wantCode string
	}{
		{"", "name_email_password_required"},
		{"not-an-email", "invalid_email"},
		{"@missing-local", "invalid_email"},
		{"missing-domain@", "invalid_email"},
	}

	for _, tt := range tests {
		_, err := RegisterUseCase(ctx, repo, &stubClaimer{}, hasher, nil, makeTokenFn("x"),
			RegisterUseCaseInput{Name: "testuser", Email: tt.email, Password: "password123"})
		if err == nil {
			t.Errorf("email=%q: expected error", tt.email)
		}
	}
}

func TestRegisterUseCase_WithAnonymousToken(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{hash: "h"}
	claim := &stubClaimer{}
	ctx := context.Background()

	_, err := RegisterUseCase(ctx, repo, claim, hasher, nil, makeTokenFn("x"),
		RegisterUseCaseInput{Name: "carol", Email: "carol@example.com", Password: "password123", AnonymousToken: "anon-token"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if len(claim.claimed) != 1 || claim.claimed[0] != 1 {
		t.Errorf("ClaimAnonymous should have been called with userID=1")
	}
}

// =============================================================================
// LoginUseCase
// =============================================================================

func TestLoginUseCase(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{hash: "h", verify: true}
	ctx := context.Background()
	repo.Create(ctx, &domain.User{Name: "login-user", PasswordHash: "h"})

	out, err := LoginUseCase(ctx, repo, hasher, makeTokenFn("login-token"), "login-user", "pw")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if out.Token != "login-token" {
		t.Errorf("Token = %q, want login-token", out.Token)
	}
}

func TestLoginUseCase_InvalidCredentials(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{verify: false}
	ctx := context.Background()
	repo.Create(ctx, &domain.User{Name: "bad-login", PasswordHash: "h"})

	_, err := LoginUseCase(ctx, repo, hasher, makeTokenFn("x"), "bad-login", "wrongpw")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginUseCase_DeactivatedWithinGrace(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{verify: true}
	ctx := context.Background()

	now := time.Now().UTC()
	repo.Create(ctx, &domain.User{Name: "react-me", PasswordHash: "h"})
	u, _ := repo.FindByName(ctx, "react-me")
	u.DeactivatedAt = &now
	u.TokenVersion = 2

	out, err := LoginUseCase(ctx, repo, hasher, makeTokenFn("react-token"), "react-me", "pw")
	if err != nil {
		t.Fatalf("Login should succeed within grace period: %v", err)
	}
	if out.Token != "react-token" {
		t.Errorf("Token = %q, want react-token", out.Token)
	}
}

func TestLoginUseCase_ByEmail(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{hash: "h", verify: true}
	ctx := context.Background()

	now := time.Now()
	repo.Create(ctx, &domain.User{Name: "email-login", PasswordHash: "h", Email: "login@example.com", EmailVerifiedAt: &now})
	// Also add to emails map for FindByEmail.
	u, _ := repo.FindByName(ctx, "email-login")
	repo.emails["login@example.com"] = u

	out, err := LoginUseCase(ctx, repo, hasher, makeTokenFn("email-token"), "login@example.com", "pw")
	if err != nil {
		t.Fatalf("Login by email: %v", err)
	}
	if out.Token != "email-token" {
		t.Errorf("Token = %q, want email-token", out.Token)
	}
}

func TestLoginUseCase_UnverifiedEmail(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{hash: "h", verify: true}
	ctx := context.Background()

	repo.Create(ctx, &domain.User{Name: "unverified", PasswordHash: "h"})
	u, _ := repo.FindByName(ctx, "unverified")
	u.PendingEmail = strPtr("unverified@example.com")
	// Also register in pendingEmails map so FindByPendingEmail succeeds.
	repo.pendingEmails["unverified@example.com"] = u

	// Login by unverified (pending) email should now succeed.
	out, err := LoginUseCase(ctx, repo, hasher, makeTokenFn("pending-token"), "unverified@example.com", "pw")
	if err != nil {
		t.Fatalf("Login by pending email should succeed: %v", err)
	}
	if out.Token != "pending-token" {
		t.Errorf("Token = %q, want pending-token", out.Token)
	}
}

func TestLoginUseCase_NonexistentEmail(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{hash: "h", verify: true}
	ctx := context.Background()

	_, err := LoginUseCase(ctx, repo, hasher, makeTokenFn("x"), "noone@example.com", "pw")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for nonexistent email, got %v", err)
	}
}

func strPtr(s string) *string { return &s }

func TestLoginUseCase_DeactivatedPastGrace(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{verify: true}
	ctx := context.Background()

	past := time.Now().UTC().Add(-8 * 24 * time.Hour)
	repo.Create(ctx, &domain.User{Name: "too-late", PasswordHash: "h"})
	u, _ := repo.FindByName(ctx, "too-late")
	u.DeactivatedAt = &past

	_, err := LoginUseCase(ctx, repo, hasher, makeTokenFn("x"), "too-late", "pw")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

// =============================================================================
// ChangePasswordUseCase
// =============================================================================

func TestChangePasswordUseCase(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{hash: "new-hash", verify: true}
	ctx := context.Background()
	repo.Create(ctx, &domain.User{Name: "cp-user", PasswordHash: "old-hash"})

	token, err := ChangePasswordUseCase(ctx, repo, hasher, makeTokenFn("new-token"), 1, "oldpw", "newpassword")
	if err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}
	if token != "new-token" {
		t.Errorf("Token = %q, want new-token", token)
	}
}

func TestChangePasswordUseCase_WrongCurrent(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{verify: false}
	ctx := context.Background()
	repo.Create(ctx, &domain.User{Name: "cp-bad", PasswordHash: "h"})

	_, err := ChangePasswordUseCase(ctx, repo, hasher, makeTokenFn("x"), 1, "wrong", "newpassword")
	if !errors.Is(err, domain.ErrCurrentPasswordWrong) {
		t.Errorf("expected ErrCurrentPasswordWrong, got %v", err)
	}
}

func TestChangePasswordUseCase_TooShort(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{verify: true}
	ctx := context.Background()
	repo.Create(ctx, &domain.User{Name: "short-pw", PasswordHash: "h"})

	_, err := ChangePasswordUseCase(ctx, repo, hasher, makeTokenFn("x"), 1, "old", "1234567")
	if !errors.Is(err, domain.ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}

// =============================================================================
// ResetPasswordUseCase
// =============================================================================

func TestResetPasswordUseCase(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{hash: "reset-hash"}
	ctx := context.Background()

	// Setup: create user and a reset token.
	repo.Create(ctx, &domain.User{Name: "reset-user", PasswordHash: "old"})
	repo.SetPasswordResetToken(ctx, "reset@test.com", "valid-token", time.Now().UTC().Add(time.Hour))
	// Manually wire the token to the user.
	repo.resetTokens["valid-token"] = resetTokenEntry{userID: 1, exp: time.Now().UTC().Add(time.Hour)}

	err := ResetPasswordUseCase(ctx, repo, hasher, "valid-token", "new-password")
	if err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}
}

func TestResetPasswordUseCase_ExpiredToken(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{hash: "h"}
	ctx := context.Background()
	repo.Create(ctx, &domain.User{Name: "exp-user", PasswordHash: "old"})
	repo.resetTokens["expired"] = resetTokenEntry{userID: 1, exp: time.Now().UTC().Add(-time.Hour)}

	err := ResetPasswordUseCase(ctx, repo, hasher, "expired", "new-password")
	if !errors.Is(err, domain.ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestResetPasswordUseCase_TooShort(t *testing.T) {
	repo := newStubRepo()
	hasher := &stubHasher{}
	ctx := context.Background()

	err := ResetPasswordUseCase(ctx, repo, hasher, "tok", "1234567")
	if !errors.Is(err, domain.ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}

// =============================================================================
// GetMeUseCase
// =============================================================================

func TestGetMeUseCase(t *testing.T) {
	repo := newStubRepo()
	ctx := context.Background()
	repo.Create(ctx, &domain.User{Name: "me-user", PasswordHash: "h"})

	u, err := GetMeUseCase(ctx, repo, 1)
	if err != nil {
		t.Fatalf("GetMe: %v", err)
	}
	if u.Name != "me-user" {
		t.Errorf("Name = %q, want me-user", u.Name)
	}
}

func TestGetMeUseCase_NotFound(t *testing.T) {
	repo := newStubRepo()
	_, err := GetMeUseCase(context.Background(), repo, 999)
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

// =============================================================================
// UpdateMeUseCase
// =============================================================================

func TestUpdateMeUseCase(t *testing.T) {
	repo := newStubRepo()
	sender := &stubEmailSender{}
	ctx := context.Background()
	repo.Create(ctx, &domain.User{Name: "update-me", PasswordHash: "h"})

	newName := "renamed"
	u, err := UpdateMeUseCase(ctx, repo, sender, UpdateMeInput{UserID: 1, Name: &newName, Locale: "en"})
	if err != nil {
		t.Fatalf("UpdateMe: %v", err)
	}
	if u.Name != "renamed" {
		t.Errorf("Name = %q, want renamed", u.Name)
	}
}

func TestUpdateMeUseCase_EmptyName(t *testing.T) {
	repo := newStubRepo()
	ctx := context.Background()
	repo.Create(ctx, &domain.User{Name: "bad-update", PasswordHash: "h"})

	emptyName := ""
	_, err := UpdateMeUseCase(ctx, repo, nil, UpdateMeInput{UserID: 1, Name: &emptyName, Locale: "en"})
	if !errors.Is(err, domain.ErrNameEmpty) {
		t.Errorf("expected ErrNameEmpty, got %v", err)
	}
}

func TestUpdateMeUseCase_EmailTrigger(t *testing.T) {
	repo := newStubRepo()
	sender := &stubEmailSender{}
	ctx := context.Background()
	repo.Create(ctx, &domain.User{Name: "email-update", PasswordHash: "h"})

	newEmail := "new@example.com"
	_, err := UpdateMeUseCase(ctx, repo, sender, UpdateMeInput{UserID: 1, Email: &newEmail, Locale: "en"})
	if err != nil {
		t.Fatalf("UpdateMe: %v", err)
	}
	if len(sender.sent) != 1 || sender.sent[0] != "verify:new@example.com" {
		t.Errorf("expected verification email sent, got %v", sender.sent)
	}
}

func TestUpdateMeUseCase_NoFieldsChanged(t *testing.T) {
	repo := newStubRepo()
	ctx := context.Background()
	repo.Create(ctx, &domain.User{Name: "no-change", PasswordHash: "h"})

	_, err := UpdateMeUseCase(ctx, repo, nil, UpdateMeInput{UserID: 1, Locale: "en"})
	if !errors.Is(err, domain.ErrNoFields) {
		t.Errorf("expected ErrNoFields, got %v", err)
	}
}

// =============================================================================
// DeactivateMeUseCase
// =============================================================================

// =============================================================================
// ResendVerificationUseCase
// =============================================================================

func TestResendVerificationUseCase_PendingEmail(t *testing.T) {
	repo := newStubRepo()
	sender := &stubEmailSender{}
	ctx := context.Background()

	// Create user with pending email (registered but not yet verified).
	repo.Create(ctx, &domain.User{Name: "resend-pending", PasswordHash: "h"})
	u, _ := repo.FindByName(ctx, "resend-pending")
	pendingEmail := "pending@example.com"
	u.PendingEmail = &pendingEmail

	email, err := ResendVerificationUseCase(ctx, repo, sender, u.ID, "en")
	if err != nil {
		t.Fatalf("ResendVerification: %v", err)
	}
	if email != "pending@example.com" {
		t.Errorf("email = %q, want pending@example.com", email)
	}
	if len(sender.sent) != 1 || sender.sent[0] != "verify:pending@example.com" {
		t.Errorf("expected verification email, got %v", sender.sent)
	}
}

func TestResendVerificationUseCase_UnverifiedEmail(t *testing.T) {
	repo := newStubRepo()
	sender := &stubEmailSender{}
	ctx := context.Background()

	// Create user with unverified email but no pending_email.
	repo.Create(ctx, &domain.User{Name: "resend-unverified", PasswordHash: "h"})
	u, _ := repo.FindByName(ctx, "resend-unverified")
	u.Email = "unverified@example.com"
	// EmailVerifiedAt remains nil — not verified.

	email, err := ResendVerificationUseCase(ctx, repo, sender, u.ID, "en")
	if err != nil {
		t.Fatalf("ResendVerification: %v", err)
	}
	if email != "unverified@example.com" {
		t.Errorf("email = %q, want unverified@example.com", email)
	}
	if len(sender.sent) != 1 || sender.sent[0] != "verify:unverified@example.com" {
		t.Errorf("expected verification email, got %v", sender.sent)
	}
}

func TestResendVerificationUseCase_AlreadyVerified(t *testing.T) {
	repo := newStubRepo()
	sender := &stubEmailSender{}
	ctx := context.Background()

	now := time.Now()
	repo.Create(ctx, &domain.User{Name: "resend-verified", PasswordHash: "h", Email: "done@example.com", EmailVerifiedAt: &now})
	u, _ := repo.FindByName(ctx, "resend-verified")

	_, err := ResendVerificationUseCase(ctx, repo, sender, u.ID, "en")
	if !errors.Is(err, domain.ErrEmailAlreadyVerified) {
		t.Errorf("expected ErrEmailAlreadyVerified, got %v", err)
	}
}

func TestResendVerificationUseCase_NoEmail(t *testing.T) {
	repo := newStubRepo()
	sender := &stubEmailSender{}
	ctx := context.Background()

	// Create user with no email at all.
	repo.Create(ctx, &domain.User{Name: "resend-noemail", PasswordHash: "h"})

	_, err := ResendVerificationUseCase(ctx, repo, sender, 1, "en")
	if !errors.Is(err, domain.ErrNoEmailToVerify) {
		t.Errorf("expected ErrNoEmailToVerify, got %v", err)
	}
}

func TestResendVerificationUseCase_NotFound(t *testing.T) {
	repo := newStubRepo()
	sender := &stubEmailSender{}
	ctx := context.Background()

	_, err := ResendVerificationUseCase(ctx, repo, sender, 999, "en")
	if err == nil {
		t.Error("expected error for nonexistent user")
	}
}

// =============================================================================
// DeactivateMeUseCase
// =============================================================================

func TestDeactivateMeUseCase(t *testing.T) {
	repo := newStubRepo()
	ctx := context.Background()
	repo.Create(ctx, &domain.User{Name: "delete-me", PasswordHash: "h"})

	reactivateBy, err := DeactivateMeUseCase(ctx, repo, 1)
	if err != nil {
		t.Fatalf("DeactivateMe: %v", err)
	}
	if reactivateBy == "" {
		t.Error("expected non-empty reactivation deadline")
	}
	u, _ := repo.FindByID(ctx, 1)
	if !u.IsDeactivated() {
		t.Error("user should be deactivated")
	}
}
