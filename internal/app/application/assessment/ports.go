package assessment

import (
	"context"

	"github.com/25types/25types/internal/25types"
	"github.com/25types/25types/internal/app/domain"
)

// AssessmentRepository is the persistence port for self/peer assessments.
type AssessmentRepository interface {
	CreateSelf(ctx context.Context, userID int64, profile domain.PersonalityProfile, answers []persona.Answer) (int64, error)
	CreateAnonymous(ctx context.Context, profile domain.PersonalityProfile, answers []persona.Answer, token string) (int64, error)
	FindLatestProfile(ctx context.Context, userID int64) (*domain.PersonalityProfile, error)
	ListSelf(ctx context.Context, userID int64, offset, limit int) ([]domain.Assessment, int, error)
	FindAssessmentByID(ctx context.Context, id int64) (*domain.Assessment, error)
	FindAssessmentByIDWithUser(ctx context.Context, id int64) (*domain.Assessment, *string, error)
	ClaimAnonymous(ctx context.Context, userID int64, token string) (int64, error)
	FindSelfAnswers(ctx context.Context, userID int64) ([]persona.Answer, error)
	ListPeerAnswersForUser(ctx context.Context, userID int64) ([]persona.Answer, int, error)
}
