package reviewlink

import (
	"context"

	"github.com/25types/25types/internal/25types"
	"github.com/25types/25types/internal/app/domain"
)

// ReviewLinkRepository is the persistence port for review link CRUD.
type ReviewLinkRepository interface {
	CreateLink(ctx context.Context, subjectUserID int64, token string, expiresAt string) (int64, error)
	FindByToken(ctx context.Context, token string) (*domain.ReviewLink, error)
	FindLinkByID(ctx context.Context, id int64) (*domain.ReviewLink, error)
	ListBySubject(ctx context.Context, subjectUserID int64) ([]ReviewLinkItem, error)
	SoftDelete(ctx context.Context, id int64, subjectUserID int64) (bool, error)
	Renew(ctx context.Context, id int64, subjectUserID int64, newExpires string) (string, bool, error)
	GetSubjectName(ctx context.Context, subjectUserID int64) string
}

// ReviewSubmissionRepository is the persistence port for peer review submissions
// and cross-aggregate reviews-given queries.
type ReviewSubmissionRepository interface {
	CreatePeerSubmission(ctx context.Context, sub *PeerSubmission) error
	GetPeerQIDStats(ctx context.Context, linkID int64) (map[string]domain.QIDStat, error)
	GetReviewSubmissions(ctx context.Context, linkID int64) ([]ReviewSubmissionItem, error)
	ListReviewsGivenByUser(ctx context.Context, userID int64) ([]ReviewsGivenItem, error)
	ListReviewsGivenByToken(ctx context.Context, anonToken string) ([]ReviewsGivenItem, error)
}

// PeerSubmission holds data for a peer review submission.
type PeerSubmission struct {
	UserID         *int64
	Profile        domain.PersonalityProfile
	Answers        []persona.Answer
	ReviewLinkID   int64
	ReviewerName   string
	AnonymousToken string
}

// ReviewLinkItem is a summary of a review link for list display.
type ReviewLinkItem struct {
	ID              int64
	Token           string
	ExpiresAt       string
	CreatedAt       string
	SubmissionCount int
}

// ReviewsGivenItem is a summary of a review the user has given.
type ReviewsGivenItem struct {
	SubjectName   string
	AnsweredCount int
	CreatedAt     string
}

// ReviewSubmissionItem is a single submission in a review link detail view.
type ReviewSubmissionItem struct {
	ReviewerName    string
	AnsweredCount   int
	LastSubmittedAt string
}
