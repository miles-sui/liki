package domain

import "time"

// MingliMatchEvent records a mingli chart match computation with chart and match JSON snapshots.
// link_id now references the unified match_links table (type='mingli').
type MingliMatchEvent struct {
	ID              int64
	LinkID          *int64
	InitiatorUserID int64
	OtherUserID     *int64
	OtherName       string
	ChartAJSON      string
	ChartBJSON      string
	MatchJSON       string
	CreatedAt       time.Time
}
