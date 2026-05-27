package matchlink

import (
	"context"
	"github.com/25types/25types/internal/25types"
	"github.com/25types/25types/internal/app/application/assessment"
	"github.com/25types/25types/internal/app/application/profile"
	"github.com/25types/25types/internal/app/domain"
)

// CreateMatchLink creates a new match link for the given user.
func CreateMatchLink(ctx context.Context, repo MatchLinkRepository, userID int64, linkType string) (*CreateMatchLinkOutput, error) {
	token := domain.GenerateToken()

	id, err := repo.Create(ctx, userID, token, linkType)
	if err != nil {
		return nil, err
	}
	urlPrefix := "/m/"
	if linkType == "mingli" {
		urlPrefix = "/ml/"
	}
	return &CreateMatchLinkOutput{ID: id, Token: token, Type: linkType, URL: urlPrefix + token}, nil
}

// GetMatchLink returns match link info by token (public).
func GetMatchLink(ctx context.Context, repo MatchLinkRepository, token string) (*MatchLinkInfo, error) {
	ml, err := repo.FindByToken(ctx, token)
	if err != nil {
		return nil, domain.ErrMatchLinkNotFound
	}
	return &MatchLinkInfo{Token: token, Valid: true, CreatorUserID: ml.UserID}, nil
}

// ListMatchLinks returns all non-deleted match links for a user.
func ListMatchLinks(ctx context.Context, repo MatchLinkRepository, userID int64, linkType string) ([]MatchLinkItem, error) {
	items, err := repo.ListByUser(ctx, userID, linkType)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []MatchLinkItem{}
	}
	return items, nil
}

// DeleteMatchLink soft-deletes a match link.
func DeleteMatchLink(ctx context.Context, repo MatchLinkRepository, id, userID int64) error {
	ok, err := repo.SoftDelete(ctx, id, userID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrMatchLinkNotFound
	}
	return nil
}

// SubmitMatchAssessmentInput holds the parameters for submitting an assessment via match link.
type SubmitMatchAssessmentInput struct {
	Token       string
	Answers     []persona.Answer
	AnonToken   string
	UseExisting bool
	OtherName   string
	UserID      *int64
}

// SubmitMatchAssessmentOutput holds the result of a match link assessment submission.
type SubmitMatchAssessmentOutput struct {
	AssessmentID  int64
	Profile       domain.PersonalityProfile
	Bond          *domain.Bond
	CreatorUserID int64
}

// SubmitMatchAPIResponse is the JSON-serializable response for POST /api/m/{token}.
type SubmitMatchAPIResponse struct {
	Profile      domain.PersonalityProfile `json:"profile"`
	AssessmentID int64                     `json:"assessment_id,omitempty"`
	Bond         *domain.Bond              `json:"bond,omitempty"`
}

// SubmitMatchAssessment processes an assessment submission via a match link token.
// It handles both anonymous (answers) and authenticated (use_existing) paths,
// computes a bond when both profiles are available, and stores the bond event.
func SubmitMatchAssessment(
	ctx context.Context,
	linkRepo MatchLinkRepository,
	assessments assessment.AssessmentRepository,
	profileLoader domain.ProfileLoader,
	bonds profile.BondStore,
	input SubmitMatchAssessmentInput,
) (*SubmitMatchAssessmentOutput, error) {
	ml, err := linkRepo.FindByToken(ctx, input.Token)
	if err != nil {
		return nil, domain.ErrMatchLinkNotFound
	}

	out := &SubmitMatchAssessmentOutput{
		CreatorUserID: ml.UserID,
	}

	if input.UseExisting {
		result, err := profile.ComputeAndStoreBondWithOpts(ctx, bonds, profileLoader, profile.InsertBondParams{
			InitiatorID: ml.UserID,
			OtherID:     *input.UserID,
			LinkID:      &ml.ID,
		})
		if err != nil {
			return nil, err
		}

		if result.ProfB != nil {
			out.Profile = *result.ProfB
		}
		out.Bond = result.Bond
		return out, nil
	}

	if len(input.Answers) == 0 {
		return nil, domain.ErrAnswersRequired
	}

	prof := domain.ComputeProfileFromAnswers(input.Answers)

	assessmentID, err := assessments.CreateAnonymous(ctx, prof, input.Answers, input.AnonToken)
	if err != nil {
		return nil, err
	}

	out.AssessmentID = assessmentID
	out.Profile = prof

	creatorProf, err := profileLoader.LoadProfile(ctx, ml.UserID)
	if err != nil {
		return out, nil
	}

	bond := domain.NewBond(prof, *creatorProf)
	if bonds != nil {
		var otherID int64
		if input.UserID != nil {
			otherID = *input.UserID
		}
		aid := assessmentID
		if err := bonds.InsertBondEvent(ctx, profile.InsertBondParams{
			LinkID: &ml.ID, InitiatorID: ml.UserID, OtherID: otherID, AssessmentID: &aid, Bond: &bond,
		}); err != nil {
			return nil, err
		}
	}

	out.Bond = &bond
	return out, nil
}
