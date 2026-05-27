package domain

import "time"

// Assessment is a self or peer assessment. Immutable once created.
type Assessment struct {
	ID             int64
	UserID         *int64
	AssessmentType AssessmentType
	IdentityID     string
	AnswersJSON    string
	ProfileJSON    string
	ReviewLinkID   *int64
	ReviewerName   string
	AnonymousToken string
	CreatedAt      time.Time
}

// ReviewLink is a peer review invitation link.
type ReviewLink struct {
	ID            int64
	SubjectUserID int64
	Token         string
	ExpiresAt     time.Time
	CreatedAt     time.Time
	DeletedAt     *time.Time
}

// IsExpired returns true if the link has passed its expiry.
func (rl *ReviewLink) IsExpired(now time.Time) bool {
	return now.After(rl.ExpiresAt)
}

// IsDeleted returns true if the link has been soft-deleted.
func (rl *ReviewLink) IsDeleted() bool {
	return rl.DeletedAt != nil && !rl.DeletedAt.IsZero()
}

// QIDStat tracks per-question statistics for peer review recommendation.
type QIDStat struct {
	Count   int
	MaxVote int
}
