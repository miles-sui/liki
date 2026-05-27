package user

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/25types/25types/internal/app/domain"
)

// RegisterUseCaseInput holds registration parameters.
type RegisterUseCaseInput struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	AnonymousToken string `json:"anonymous_token,omitempty"`
}

// RegisterUseCaseOutput holds the result of successful registration.
type RegisterUseCaseOutput struct {
	Token string
	User  RegisteredUser
}

// RegisteredUser is the user view returned after auth operations.
type RegisteredUser struct {
	ID             int64             `json:"id"`
	Name           string            `json:"name"`
	Email          string            `json:"email,omitempty"`
	PendingEmail   *string           `json:"pending_email,omitempty"`
	EmailVerified  bool              `json:"email_verified"`
	IsPublic       bool              `json:"is_public"`
	SupporterSince *time.Time        `json:"supporter_since"`
	BirthInfo      *domain.BirthInfo `json:"birth_info,omitempty"`
}

// toRegisteredUser maps a domain.User to the API view.
func toRegisteredUser(user *domain.User) RegisteredUser {
	return RegisteredUser{
		ID:             user.ID,
		Name:           user.Name,
		Email:          user.Email,
		PendingEmail:   user.PendingEmail,
		EmailVerified:  user.EmailVerifiedAt != nil,
		IsPublic:       user.IsPublic,
		SupporterSince: user.SupporterSince,
		BirthInfo:      user.BirthInfo,
	}
}

// RegisterUseCase creates a new user account and returns a JWT token.
func RegisterUseCase(
	ctx context.Context,
	repo UserRepository,
	claim Claimer,
	hasher PasswordHasher,
	sender EmailSender,
	tokenFn func(userID int64, tokenVersion int, userName string) (string, error),
	input RegisterUseCaseInput,
) (*RegisterUseCaseOutput, error) {
	if err := domain.ValidateRegistration(input.Name, input.Email, input.Password); err != nil {
		return nil, err
	}

	email := input.Email

	// Check if email is already verified by another user.
	if _, err := repo.FindByEmail(ctx, email); err == nil {
		return nil, domain.ErrEmailAlreadyVerified
	} else if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}
	// Also reject if email is pending verification on another account.
	if _, err := repo.FindByPendingEmail(ctx, email); err == nil {
		return nil, domain.ErrEmailAlreadyVerified
	} else if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	hash, err := hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:         input.Name,
		PasswordHash: hash,
		PendingEmail: &email,
	}

	id, err := repo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, domain.ErrUsernameTaken) {
			return nil, err
		}
		return nil, err
	}

	// Send verification email for the new email address.
	verTok := domain.GenerateToken()
	if err := repo.UpdateFields(ctx, id, UpdateUserFields{Email: &email, EmailVerToken: &verTok}); err != nil {
		return nil, err
	}
	if sender != nil {
		if err := sender.SendVerificationEmail(ctx, email, verTok, "en"); err != nil {
			log.Printf("[email] verification email failed for %s: %v", email, err)
		}
	} else {
		log.Printf("[email] no sender — verification link: http://localhost:8080/en/verify-email?token=%s", verTok)
	}

	if input.AnonymousToken != "" {
		claim.ClaimAnonymous(ctx, id, input.AnonymousToken)
	}

	token, err := tokenFn(id, 1, input.Name)
	if err != nil {
		return nil, err
	}

	return &RegisterUseCaseOutput{
		Token: token,
		User: RegisteredUser{
			ID:            id,
			Name:          input.Name,
			Email:         "",
			PendingEmail:  &email,
			EmailVerified: false,
		},
	}, nil
}

// ResendVerificationUseCase sends a new verification email for the authenticated user.
// It uses pending_email if present, otherwise the current email (if unverified).
func ResendVerificationUseCase(
	ctx context.Context,
	repo UserRepository,
	sender EmailSender,
	userID int64,
	locale string,
) (string, error) {
	user, err := repo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}

	target := ""
	if user.PendingEmail != nil && *user.PendingEmail != "" {
		target = *user.PendingEmail
	} else if user.Email != "" && user.EmailVerifiedAt == nil {
		target = user.Email
	}
	if user.EmailVerifiedAt != nil && (user.PendingEmail == nil || *user.PendingEmail == "") {
		return "", domain.ErrEmailAlreadyVerified
	}
	if target == "" {
		return "", domain.ErrNoEmailToVerify
	}

	token := domain.GenerateToken()
	if err := repo.UpdateFields(ctx, userID, UpdateUserFields{Email: &target, EmailVerToken: &token}); err != nil {
		return "", err
	}
	if sender != nil {
		if err := sender.SendVerificationEmail(ctx, target, token, locale); err != nil {
			log.Printf("[email] resend verification failed for %s: %v", target, err)
		}
	} else {
		log.Printf("[email] no sender — verification link: http://localhost:8080/en/verify-email?token=%s", token)
	}
	return target, nil
}

// LoginUseCase authenticates a user and returns a JWT token.
// The nameOrEmail parameter accepts username, verified email, or pending email.
func LoginUseCase(
	ctx context.Context,
	repo UserRepository,
	hasher PasswordHasher,
	tokenFn func(userID int64, tokenVersion int, userName string) (string, error),
	nameOrEmail, password string,
) (*RegisterUseCaseOutput, error) {
	nameOrEmail = strings.TrimSpace(nameOrEmail)
	if nameOrEmail == "" || password == "" {
		return nil, domain.ErrNameAndPasswordRequired
	}

	var user *domain.User
	var err error
	if strings.Contains(nameOrEmail, "@") {
		user, err = repo.FindByEmail(ctx, nameOrEmail)
		if err != nil {
			user, err = repo.FindByPendingEmail(ctx, nameOrEmail)
		}
	} else {
		user, err = repo.FindByName(ctx, nameOrEmail)
	}
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	valid, rehash := hasher.Verify(password, user.PasswordHash)
	if !valid {
		return nil, domain.ErrInvalidCredentials
	}

	// Auto-rehash on parameter upgrade.
	if rehash != "" {
		repo.UpdatePasswordHash(ctx, user.ID, rehash)
		user.PasswordHash = rehash
	}

	// Handle deactivated account reactivation.
	if user.IsDeactivated() {
		if err := user.ReactivateIfEligible(time.Now()); err != nil {
			return nil, err
		}
		repo.ReactivateUser(ctx, user.ID)
	}

	token, err := tokenFn(user.ID, user.TokenVersion, user.Name)
	if err != nil {
		return nil, err
	}

	return &RegisterUseCaseOutput{
		Token: token,
		User:  toRegisteredUser(user),
	}, nil
}

// LogoutUseCase increments the token_version to invalidate all existing tokens.
func LogoutUseCase(ctx context.Context, repo UserRepository, userID int64) error {
	_, err := repo.UpdateTokenVersion(ctx, userID)
	return err
}

// ChangePasswordUseCase changes the user's password after verifying the current one.
func ChangePasswordUseCase(
	ctx context.Context,
	repo UserRepository,
	hasher PasswordHasher,
	tokenFn func(userID int64, tokenVersion int, userName string) (string, error),
	userID int64,
	currentPassword, newPassword string,
) (string, error) {
	if len(newPassword) < 8 {
		return "", domain.ErrPasswordTooShort
	}

	user, err := repo.FindByID(ctx, userID)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	valid, _ := hasher.Verify(currentPassword, user.PasswordHash)
	if !valid {
		return "", domain.ErrCurrentPasswordWrong
	}

	hash, err := hasher.Hash(newPassword)
	if err != nil {
		return "", err
	}

	newTV, err := repo.UpdatePasswordHash(ctx, userID, hash)
	if err != nil {
		return "", err
	}

	token, err := tokenFn(userID, newTV, user.Name)
	if err != nil {
		return "", err
	}
	return token, nil
}

// ForgotPasswordUseCase generates a password reset token, stores it, and sends it.
func ForgotPasswordUseCase(ctx context.Context, repo UserRepository, sender EmailSender, email, locale string) {
	if email == "" {
		return
	}

	tok := domain.GenerateToken()
	exp := time.Now().Add(15 * time.Minute).UTC()

	repo.SetPasswordResetToken(ctx, email, tok, exp)
	if sender != nil {
		if err := sender.SendPasswordResetEmail(ctx, email, tok, locale); err != nil {
			log.Printf("[email] password reset email failed for %s: %v", email, err)
		}
	}
}

// ResetPasswordUseCase changes the password using a valid reset token.
func ResetPasswordUseCase(
	ctx context.Context,
	repo UserRepository,
	hasher PasswordHasher,
	token, password string,
) error {
	if len(password) < 8 {
		return domain.ErrPasswordTooShort
	}
	uid, err := repo.FindByPasswordResetToken(ctx, token)
	if err != nil {
		return domain.ErrTokenExpired
	}
	hash, err := hasher.Hash(password)
	if err != nil {
		return err
	}
	return repo.ResetPassword(ctx, uid, hash)
}

// GetMeUseCase returns the authenticated user's profile.
func GetMeUseCase(ctx context.Context, repo UserRepository, userID int64) (*RegisteredUser, error) {
	user, err := repo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	ru := toRegisteredUser(user)
	return &ru, nil
}

// UpdateMeInput holds the parameters for UpdateMeUseCase.
type UpdateMeInput struct {
	UserID    int64
	Name      *string
	Email     *string
	IsPublic  *bool
	BirthInfo *domain.BirthInfo
	Locale    string
}

// UpdateMeUseCase updates the authenticated user's profile.
func UpdateMeUseCase(
	ctx context.Context,
	repo UserRepository,
	sender EmailSender,
	input UpdateMeInput,
) (*RegisteredUser, error) {
	var fields UpdateUserFields
	var hasField bool

	name := input.Name
	if name != nil {
		*name = strings.TrimSpace(*name)
		if *name == "" {
			return nil, domain.ErrNameEmpty
		}
		if domain.IsReservedName(*name) {
			return nil, domain.ErrUsernameReserved
		}
		fields.Name = name
		hasField = true
	}
	var emailChanged string
	var emailVerToken string
	if input.Email != nil && *input.Email != "" {
		emailChanged = *input.Email
		if existing, err := repo.FindByEmail(ctx, emailChanged); err == nil && existing.ID != input.UserID {
			return nil, domain.ErrEmailTaken
		}
		if existing, err := repo.FindByPendingEmail(ctx, emailChanged); err == nil && existing.ID != input.UserID {
			return nil, domain.ErrEmailTaken
		}
		fields.Email = input.Email
		emailVerToken = domain.GenerateToken()
		fields.EmailVerToken = &emailVerToken
		hasField = true
	}
	if input.IsPublic != nil {
		fields.IsPublic = input.IsPublic
		hasField = true
	}

	if input.BirthInfo != nil {
		if input.BirthInfo.Year == 0 {
			empty := ""
			fields.BirthInfo = &empty
		} else {
			raw, _ := json.Marshal(input.BirthInfo)
			s := string(raw)
			fields.BirthInfo = &s
		}
		hasField = true
	}

	if !hasField {
		return nil, domain.ErrNoFields
	}

	if err := repo.UpdateFields(ctx, input.UserID, fields); err != nil {
		return nil, err
	}
	if emailChanged != "" && sender != nil {
		if err := sender.SendVerificationEmail(ctx, emailChanged, emailVerToken, input.Locale); err != nil {
			log.Printf("[email] verification email failed for %s: %v", emailChanged, err)
		}
	}

	// Re-read to return current state.
	user, err := repo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	ru := toRegisteredUser(user)
	return &ru, nil
}

// DeactivateMeUseCase soft-deactivates the user's account, allowing reactivation within 7 days.
func DeactivateMeUseCase(ctx context.Context, repo UserRepository, userID int64) (string, error) {
	now := time.Now().UTC()
	reactivateBy := now.Add(7 * 24 * time.Hour)
	if err := repo.SetDeactivated(ctx, userID, now); err != nil {
		return "", err
	}
	return reactivateBy.Format(time.RFC3339), nil
}
