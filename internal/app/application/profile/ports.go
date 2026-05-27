package profile

import (
	"context"

	"github.com/25types/25types/internal/25types"
	"github.com/25types/25types/internal/app/domain"
)

// ProfilePageRepo bundles read access needed by the profile page.  It intentionally
// aggregates user, profile, peer, and review-link data because the profile page
// displays all of these together.
type ProfilePageRepo interface {
	domain.ProfileLoader
	FindByName(ctx context.Context, name string) (*domain.User, error)
	ListPeerAnswersForUser(ctx context.Context, userID int64) ([]persona.Answer, int, error)
	FindActiveReviewLink(ctx context.Context, userID int64) (string, bool)
}

// UserLookup looks up users by name or ID.
type UserLookup interface {
	FindByName(ctx context.Context, name string) (*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
}

// InsertBondParams holds the parameters for persisting a bond event.
type InsertBondParams struct {
	LinkID       *int64
	InitiatorID  int64
	OtherID      int64
	AssessmentID *int64
	Bond         *domain.Bond
}

// BondStore persists bond events with bond_json snapshots.
type BondStore interface {
	InsertBondEvent(ctx context.Context, params InsertBondParams) error
	ListBondEvents(ctx context.Context, userID int64) ([]domain.BondEvent, error)
}
