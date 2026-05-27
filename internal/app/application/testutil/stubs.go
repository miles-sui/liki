// Package testutil provides shared test stubs for application-level service tests.
// It only imports leaf packages (domain, persona) to avoid import cycles with
// application packages that use these stubs in their tests.
package testutil

import (
	"context"

	"github.com/25types/25types/internal/25types"
	"github.com/25types/25types/internal/app/domain"
)

// =============================================================================
// StubAssessRepo — implements assessment.AssessmentRepository
// =============================================================================

type StubAssessRepo struct {
	Profiles    map[int64]*domain.PersonalityProfile
	Answers     map[int64][]persona.Answer
	PeerAnswers map[int64][]persona.Answer
	PeerCount   map[int64]int
	NextID      int64
	ClaimTarget int64
	ClaimErr    error
}

func NewStubAssessRepo() *StubAssessRepo {
	return &StubAssessRepo{
		Profiles:    map[int64]*domain.PersonalityProfile{},
		Answers:     map[int64][]persona.Answer{},
		PeerAnswers: map[int64][]persona.Answer{},
		PeerCount:   map[int64]int{},
	}
}

func (r *StubAssessRepo) CreateSelf(ctx context.Context, userID int64, prof domain.PersonalityProfile, answers []persona.Answer) (int64, error) {
	r.Profiles[userID] = &prof
	r.Answers[userID] = answers
	r.NextID++
	return r.NextID, nil
}

func (r *StubAssessRepo) CreateAnonymous(ctx context.Context, prof domain.PersonalityProfile, answers []persona.Answer, token string) (int64, error) {
	r.NextID++
	return r.NextID, nil
}

func (r *StubAssessRepo) FindLatestProfile(ctx context.Context, userID int64) (*domain.PersonalityProfile, error) {
	p, ok := r.Profiles[userID]
	if !ok {
		return nil, domain.ErrNoProfile
	}
	return p, nil
}

func (r *StubAssessRepo) ListSelf(ctx context.Context, userID int64, offset, limit int) ([]domain.Assessment, int, error) {
	return nil, 0, nil
}

func (r *StubAssessRepo) FindAssessmentByID(ctx context.Context, id int64) (*domain.Assessment, error) {
	return nil, nil
}

func (r *StubAssessRepo) FindAssessmentByIDWithUser(ctx context.Context, id int64) (*domain.Assessment, *string, error) {
	return nil, nil, nil
}

func (r *StubAssessRepo) ClaimAnonymous(ctx context.Context, userID int64, token string) (int64, error) {
	if r.ClaimErr != nil {
		return 0, r.ClaimErr
	}
	r.ClaimTarget = userID
	return 1, nil
}

func (r *StubAssessRepo) FindSelfAnswers(ctx context.Context, userID int64) ([]persona.Answer, error) {
	return r.Answers[userID], nil
}

func (r *StubAssessRepo) ListPeerAnswersForUser(ctx context.Context, userID int64) ([]persona.Answer, int, error) {
	return r.PeerAnswers[userID], r.PeerCount[userID], nil
}

// =============================================================================
// StubProfileLoader — implements domain.ProfileLoader
// =============================================================================

type StubProfileLoader struct {
	Profiles map[int64]*domain.PersonalityProfile
}

func (l *StubProfileLoader) LoadProfile(ctx context.Context, userID int64) (*domain.PersonalityProfile, error) {
	p, ok := l.Profiles[userID]
	if !ok {
		return nil, domain.ErrNoProfile
	}
	return p, nil
}
