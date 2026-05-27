package http

import (
	"context"

	"github.com/25types/25types/internal/app/domain"
)

// BirthInfoLookup looks up a user's birth info by ID.
type BirthInfoLookup interface {
	FindByID(ctx context.Context, id int64) (*domain.User, error)
}
