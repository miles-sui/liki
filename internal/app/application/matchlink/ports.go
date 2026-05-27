package matchlink

import (
	"context"

	"github.com/25types/25types/internal/app/domain"
)

// MatchLinkRepository is the persistence port for match links.
type MatchLinkRepository interface {
	Create(ctx context.Context, userID int64, token string, linkType string) (int64, error)
	FindByToken(ctx context.Context, token string) (*domain.MatchLink, error)
	ListByUser(ctx context.Context, userID int64, linkType string) ([]MatchLinkItem, error)
	SoftDelete(ctx context.Context, id int64, userID int64) (bool, error)
	InsertMingliMatchEvent(ctx context.Context, params InsertMingliMatchEventParams) error
}

// MatchLinkItem is a summary of a match link for list display.
type MatchLinkItem struct {
	ID         int64  `json:"id"`
	Token      string `json:"token"`
	Type       string `json:"type"`
	BondCount  int    `json:"bond_count"`
	MatchCount int    `json:"match_count"`
	CreatedAt  string `json:"created_at"`
}

// MatchLinkInfo is returned by GET /api/m/{token}.
type MatchLinkInfo struct {
	Token         string `json:"token"`
	CreatorName   string `json:"creator_name"`
	Valid         bool   `json:"valid"`
	CreatorUserID int64  `json:"-"`
}

// CreateMatchLinkOutput is returned after creating a match link.
type CreateMatchLinkOutput struct {
	ID    int64  `json:"id"`
	Token string `json:"token"`
	Type  string `json:"type"`
	URL   string `json:"url"`
}

// InsertMingliMatchEventParams holds the data for storing a mingli chart match event.
type InsertMingliMatchEventParams struct {
	LinkID          *int64
	InitiatorUserID int64
	OtherUserID     *int64
	OtherName       string
	ChartAJSON      string
	ChartBJSON      string
	MatchJSON       string
}
