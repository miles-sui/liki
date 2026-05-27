package domain

import "time"

// MatchLink is a shareable link for matching. Type is "assessment" or "mingli".
type MatchLink struct {
	ID        int64
	UserID    int64
	Token     string
	Type      string
	CreatedAt time.Time
	DeletedAt *time.Time
}

// BondEvent records a bond computation with a bond_json snapshot.
// link_id is nil for instant compare. other_user_id is nil for anonymous recipients.
type BondEvent struct {
	ID              int64
	LinkID          *int64
	InitiatorUserID int64
	OtherUserID     *int64
	OtherName       string
	AssessmentID    *int64
	BondJSON        string `json:"-"`
	CreatedAt       time.Time
}
