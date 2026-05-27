package minglilink

import (
	"context"

	"github.com/25types/25types/internal/app/domain"
)

// BirthInfoLookup looks up a user's birth info by ID.
type BirthInfoLookup interface {
	FindByID(ctx context.Context, id int64) (*domain.User, error)
}

// MingliMatchLinkInfo is returned by GET /api/m/{token}.
type MingliMatchLinkInfo struct {
	Token       string      `json:"token"`
	CreatorName string      `json:"creator_name"`
	Valid       bool        `json:"valid"`
	ChartA      interface{} `json:"chart_a"`
}

// SubmitMingliMatchInput holds the parameters for submitting a BaZi match via link.
type SubmitMingliMatchInput struct {
	Token       string
	UseExisting bool
	BirthInfo   *domain.BirthInfo
	OtherName   string
	UserID      *int64
}

// SubmitMingliMatchOutput holds the result of a BaZi match link submission.
type SubmitMingliMatchOutput struct {
	ChartA        interface{} `json:"chart_a"`
	ChartB        interface{} `json:"chart_b"`
	Bond          interface{} `json:"bond"`
	CreatorUserID int64       `json:"-"`
}
