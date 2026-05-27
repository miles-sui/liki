package assessment

import (
	"context"

	"github.com/25types/25types/internal/25types"
	"github.com/25types/25types/internal/app/domain"
)

// -- Submit --

// SubmitAssessmentUseCase handles self-assessment submission with 3-way fork.
func SubmitAssessmentUseCase(
	ctx context.Context,
	repo AssessmentRepository,
	answers []persona.Answer,
	userID *int64,
	anonToken string,
) (*SubmitResult, error) {
	if len(answers) == 0 {
		return nil, domain.ErrAnswersRequired
	}

	profile := domain.ComputeProfileFromAnswers(answers)

	result := &SubmitResult{
		Profile:  profile,
		Identity: profile.Identity,
	}

	if userID != nil {
		aid, err := repo.CreateSelf(ctx, *userID, profile, answers)
		if err != nil {
			return nil, err
		}
		result.ID = aid
		return result, nil
	}

	if anonToken != "" {
		aid, err := repo.CreateAnonymous(ctx, profile, answers, anonToken)
		if err == nil {
			result.ID = aid
		}
		result.AnonToken = anonToken
	}
	return result, nil
}

// SubmitResult holds the output of assessment submission.
type SubmitResult struct {
	ID        int64
	Profile   domain.PersonalityProfile
	Identity  persona.Identity
	AnonToken string
}

// -- Peers --

// PeersUseCase computes self, peer-aggregated, and combined profiles.
func PeersUseCase(
	ctx context.Context,
	repo AssessmentRepository,
	userID int64,
) (*domain.PeerProfile, error) {
	selfProfile, err := repo.FindLatestProfile(ctx, userID)
	if err != nil {
		return nil, domain.ErrNoProfile
	}

	result := &domain.PeerProfile{
		Self: *selfProfile,
	}

	// Aggregate peer assessments.
	peerAnswers, peerCount, _ := repo.ListPeerAnswersForUser(ctx, userID)
	result.PeerCount = peerCount

	if peerCount > 0 {
		peerProfile := domain.ComputeProfileFromAnswers(peerAnswers)
		result.Peers = &peerProfile

		// Combined = self answers + all peer answers.
		selfAnswers, _ := repo.FindSelfAnswers(ctx, userID)
		combined := append(selfAnswers, peerAnswers...)
		combinedProfile := domain.ComputeProfileFromAnswers(combined)
		result.Combined = &combinedProfile
	}

	return result, nil
}

// -- Claim --

// ClaimUseCase claims anonymous assessments for a logged-in user.
func ClaimUseCase(ctx context.Context, repo AssessmentRepository, userID int64, token string) (int64, error) {
	return repo.ClaimAnonymous(ctx, userID, token)
}
