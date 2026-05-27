package user

import (
	"context"
	"time"

	"github.com/25types/25types/internal/app/domain"
)

// UserRepository is the write-side port for User persistence.
type UserRepository interface {
	FindByName(ctx context.Context, name string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByPendingEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) (int64, error)
	UpdateTokenVersion(ctx context.Context, id int64) (int, error)
	UpdatePasswordHash(ctx context.Context, id int64, hash string) (int, error)
	UpdateFields(ctx context.Context, id int64, fields UpdateUserFields) error
	SetDeactivated(ctx context.Context, id int64, at time.Time) error
	ReactivateUser(ctx context.Context, id int64) error
	VerifyEmailByToken(ctx context.Context, token string) error
	SetPasswordResetToken(ctx context.Context, email, token string, exp time.Time) error
	FindByPasswordResetToken(ctx context.Context, token string) (int64, error)
	ResetPassword(ctx context.Context, id int64, hash string) error
	DeleteUser(ctx context.Context, id int64) error
}

// ExportRepository provides cross-aggregate read access for GDPR data export.
// It intentionally queries across User, Assessment, and ReviewLink aggregates.
type ExportRepository interface {
	GetExportAssessments(ctx context.Context, userID int64) ([]ExportAssessment, error)
	GetExportReviewLinks(ctx context.Context, userID int64) ([]ExportReviewLink, error)
}

// UpdateUserFields holds the mutable user profile fields.
type UpdateUserFields struct {
	Name          *string
	Email         *string
	IsPublic      *bool
	BirthInfo     *string // JSON string, set to "" to clear
	EmailVerToken *string // set alongside Email to trigger verification flow
}

// TokenValidator checks that a user's token_version matches the stored value.
type TokenValidator interface {
	GetTokenVersion(ctx context.Context, userID int64) (int, error)
}

// PasswordHasher abstracts password hashing and verification.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, storedHash string) (valid bool, newHash string)
}

// Claimer is a narrow interface for claiming anonymous assessments during registration.
type Claimer interface {
	ClaimAnonymous(ctx context.Context, userID int64, token string) (int64, error)
}

// EmailSender abstracts transactional email delivery.
type EmailSender interface {
	SendVerificationEmail(ctx context.Context, to, token, locale string) error
	SendPasswordResetEmail(ctx context.Context, to, token, locale string) error
	SendBondNotification(ctx context.Context, to, otherName, creatorName, locale string) error
}

// ExportAssessment is an assessment entry in the export.
type ExportAssessment struct {
	ID           int64
	Type         string
	IdentityID   string
	ProfileJSON  string
	AnswersJSON  string
	CreatedAt    string
	ReviewLinkID *int64
	ReviewerName string
}

// ExportReviewLink is a review link entry in the export.
type ExportReviewLink struct {
	ID        int64
	Token     string
	ExpiresAt string
	CreatedAt string
}
