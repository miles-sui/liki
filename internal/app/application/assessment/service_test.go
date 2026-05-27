package assessment

import (
	"context"
	"errors"
	"testing"

	"github.com/25types/25types/internal/app/application/testutil"
	"github.com/25types/25types/internal/app/domain"
	persona "github.com/25types/25types/internal/25types"
)

func makeAnswer(qid, a, b string) persona.Answer {
	return persona.Answer{QID: qid, Selections: []string{a, b}}
}

// =============================================================================
// SubmitAssessmentUseCase
// =============================================================================

func TestSubmit_OK(t *testing.T) {
	repo := testutil.NewStubAssessRepo()
	uid := int64(1)
	answers := []persona.Answer{makeAnswer("Q01", "W", "F"), makeAnswer("Q02", "E", "M")}

	result, err := SubmitAssessmentUseCase(context.Background(), repo, answers, &uid, "")
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if result.ID == 0 {
		t.Error("expected non-zero assessment ID")
	}
	if result.Profile.D[0] == 0 && result.Profile.D[1] == 0 {
		t.Error("expected non-zero deviation profile")
	}
}

func TestSubmit_Anonymous(t *testing.T) {
	repo := testutil.NewStubAssessRepo()
	answers := []persona.Answer{makeAnswer("Q01", "W", "F"), makeAnswer("Q02", "E", "M")}

	result, err := SubmitAssessmentUseCase(context.Background(), repo, answers, nil, "anon-token-123")
	if err != nil {
		t.Fatalf("Submit anonymous: %v", err)
	}
	if result.AnonToken != "anon-token-123" {
		t.Errorf("AnonToken = %q, want anon-token-123", result.AnonToken)
	}
}

func TestSubmit_NoAnswers(t *testing.T) {
	_, err := SubmitAssessmentUseCase(context.Background(), nil, nil, nil, "")
	if !errors.Is(err, domain.ErrAnswersRequired) {
		t.Errorf("expected ErrAnswersRequired, got %v", err)
	}
}

func TestSubmit_NoUserAndNoToken(t *testing.T) {
	answers := []persona.Answer{makeAnswer("Q01", "W", "F")}
	result, err := SubmitAssessmentUseCase(context.Background(), testutil.NewStubAssessRepo(), answers, nil, "")
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if result.ID != 0 {
		t.Error("expected zero ID when nothing to persist")
	}
}

// =============================================================================
// PeersUseCase
// =============================================================================

func TestPeers_SelfOnly(t *testing.T) {
	repo := testutil.NewStubAssessRepo()
	repo.Profiles[1] = &domain.PersonalityProfile{}

	result, err := PeersUseCase(context.Background(), repo, 1)
	if err != nil {
		t.Fatalf("Peers: %v", err)
	}
	if result.PeerCount != 0 {
		t.Errorf("PeerCount = %d, want 0", result.PeerCount)
	}
	if result.Peers != nil {
		t.Error("expected nil Peers")
	}
}

func TestPeers_WithPeers(t *testing.T) {
	repo := testutil.NewStubAssessRepo()
	repo.Profiles[1] = &domain.PersonalityProfile{}
	repo.PeerAnswers[1] = []persona.Answer{makeAnswer("Q01", "W", "F"), makeAnswer("Q02", "E", "M")}
	repo.PeerCount[1] = 3

	result, err := PeersUseCase(context.Background(), repo, 1)
	if err != nil {
		t.Fatalf("Peers: %v", err)
	}
	if result.PeerCount != 3 {
		t.Errorf("PeerCount = %d, want 3", result.PeerCount)
	}
	if result.Peers == nil {
		t.Error("expected non-nil Peers")
	}
	if result.Combined == nil {
		t.Error("expected non-nil Combined")
	}
}

func TestPeers_NoProfile(t *testing.T) {
	repo := testutil.NewStubAssessRepo()
	_, err := PeersUseCase(context.Background(), repo, 999)
	if !errors.Is(err, domain.ErrNoProfile) {
		t.Errorf("expected ErrNoProfile, got %v", err)
	}
}

// =============================================================================
// ClaimUseCase
// =============================================================================

func TestClaim_OK(t *testing.T) {
	repo := testutil.NewStubAssessRepo()
	n, err := ClaimUseCase(context.Background(), repo, 1, "anon-token")
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if n != 1 {
		t.Errorf("claimed = %d, want 1", n)
	}
	if repo.ClaimTarget != 1 {
		t.Errorf("ClaimTarget = %d, want 1", repo.ClaimTarget)
	}
}
