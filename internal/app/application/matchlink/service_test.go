package matchlink

import (
	"context"
	"errors"
	"testing"

	"github.com/25types/25types/internal/app/application/profile"
	"github.com/25types/25types/internal/app/application/testutil"
	"github.com/25types/25types/internal/app/domain"
	persona "github.com/25types/25types/internal/25types"
)

// =============================================================================
// Stubs (package-specific)
// =============================================================================

type stubMLRepo struct {
	links  map[string]*domain.MatchLink
	byUser map[int64][]*domain.MatchLink
	nextID int64
}

func newStubMLRepo() *stubMLRepo {
	return &stubMLRepo{
		links:  map[string]*domain.MatchLink{},
		byUser: map[int64][]*domain.MatchLink{},
	}
}

func (r *stubMLRepo) Create(ctx context.Context, userID int64, token string, linkType string) (int64, error) {
	r.nextID++
	ml := &domain.MatchLink{ID: r.nextID, UserID: userID, Token: token, Type: linkType}
	r.links[token] = ml
	r.byUser[userID] = append(r.byUser[userID], ml)
	return r.nextID, nil
}

func (r *stubMLRepo) FindByToken(ctx context.Context, token string) (*domain.MatchLink, error) {
	ml, ok := r.links[token]
	if !ok {
		return nil, domain.ErrMatchLinkNotFound
	}
	return ml, nil
}

func (r *stubMLRepo) ListByUser(ctx context.Context, userID int64, linkType string) ([]MatchLinkItem, error) {
	mls := r.byUser[userID]
	items := make([]MatchLinkItem, 0, len(mls))
	for _, ml := range mls {
		if linkType != "" && ml.Type != linkType {
			continue
		}
		items = append(items, MatchLinkItem{ID: ml.ID, Token: ml.Token, Type: ml.Type})
	}
	return items, nil
}

func (r *stubMLRepo) InsertMingliMatchEvent(ctx context.Context, params InsertMingliMatchEventParams) error {
	return nil
}

func (r *stubMLRepo) SoftDelete(ctx context.Context, id, userID int64) (bool, error) {
	for token, ml := range r.links {
		if ml.ID == id && ml.UserID == userID {
			delete(r.links, token)
			return true, nil
		}
	}
	return false, nil
}

// stubBondStore implements profile.BondStore (cannot be in testutil due to import cycle).
type stubBondStore struct {
	inserted []profile.InsertBondParams
}

func (s *stubBondStore) InsertBondEvent(ctx context.Context, params profile.InsertBondParams) error {
	s.inserted = append(s.inserted, params)
	return nil
}
func (s *stubBondStore) ListBondEvents(ctx context.Context, userID int64) ([]domain.BondEvent, error) {
	return nil, nil
}

var _ profile.BondStore = (*stubBondStore)(nil)

// makeTestAnswers generates 30 answers heavily favoring one element.
func makeTestAnswers(primary, secondary string) []persona.Answer {
	answers := make([]persona.Answer, 30)
	for i := range answers {
		qid := "Q" + string(rune('0'+((i/10)+1))) + string(rune('0'+(i%10)))
		if i < 10 {
			qid = "Q0" + string(rune('0'+i+1))
		}
		answers[i] = persona.Answer{QID: qid, Selections: []string{primary, secondary}}
	}
	return answers
}

// =============================================================================
// CreateMatchLink
// =============================================================================

func TestCreateMatchLink(t *testing.T) {
	repo := newStubMLRepo()
	out, err := CreateMatchLink(context.Background(), repo, 1, "assessment")
	if err != nil {
		t.Fatalf("CreateMatchLink: %v", err)
	}
	if out.ID <= 0 {
		t.Errorf("expected ID > 0, got %d", out.ID)
	}
	if out.Token == "" {
		t.Error("expected non-empty token")
	}
	if out.URL == "" {
		t.Error("expected non-empty URL")
	}
}

// =============================================================================
// GetMatchLink
// =============================================================================

func TestGetMatchLink_Valid(t *testing.T) {
	repo := newStubMLRepo()
	out, _ := CreateMatchLink(context.Background(), repo, 42, "assessment")
	info, err := GetMatchLink(context.Background(), repo, out.Token)
	if err != nil {
		t.Fatalf("GetMatchLink: %v", err)
	}
	if !info.Valid {
		t.Error("expected Valid=true")
	}
	if info.CreatorUserID != 42 {
		t.Errorf("expected CreatorUserID=42, got %d", info.CreatorUserID)
	}
}

func TestGetMatchLink_NotFound(t *testing.T) {
	repo := newStubMLRepo()
	_, err := GetMatchLink(context.Background(), repo, "no-such-token")
	if !errors.Is(err, domain.ErrMatchLinkNotFound) {
		t.Errorf("expected ErrMatchLinkNotFound, got %v", err)
	}
}

// =============================================================================
// ListMatchLinks
// =============================================================================

func TestListMatchLinks(t *testing.T) {
	repo := newStubMLRepo()
	for i := 0; i < 3; i++ {
		CreateMatchLink(context.Background(), repo, 7, "assessment")
	}
	items, err := ListMatchLinks(context.Background(), repo, 7, "assessment")
	if err != nil {
		t.Fatalf("ListMatchLinks: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
}

func TestListMatchLinks_Empty(t *testing.T) {
	repo := newStubMLRepo()
	items, err := ListMatchLinks(context.Background(), repo, 99, "assessment")
	if err != nil {
		t.Fatalf("ListMatchLinks: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

// =============================================================================
// DeleteMatchLink
// =============================================================================

func TestDeleteMatchLink(t *testing.T) {
	repo := newStubMLRepo()
	out, _ := CreateMatchLink(context.Background(), repo, 1, "assessment")
	err := DeleteMatchLink(context.Background(), repo, out.ID, 1)
	if err != nil {
		t.Fatalf("DeleteMatchLink: %v", err)
	}
	_, err = GetMatchLink(context.Background(), repo, out.Token)
	if !errors.Is(err, domain.ErrMatchLinkNotFound) {
		t.Errorf("expected ErrMatchLinkNotFound after delete, got %v", err)
	}
}

func TestDeleteMatchLink_NotOwner(t *testing.T) {
	repo := newStubMLRepo()
	out, _ := CreateMatchLink(context.Background(), repo, 1, "assessment")
	err := DeleteMatchLink(context.Background(), repo, out.ID, 2)
	if !errors.Is(err, domain.ErrMatchLinkNotFound) {
		t.Errorf("expected ErrMatchLinkNotFound for non-owner, got %v", err)
	}
}

// =============================================================================
// SubmitMatchAssessment
// =============================================================================

func TestSubmitMatchAssessment_Answers(t *testing.T) {
	repo := newStubMLRepo()
	out, _ := CreateMatchLink(context.Background(), repo, 10, "assessment")
	loader := &testutil.StubProfileLoader{Profiles: map[int64]*domain.PersonalityProfile{
		10: {D: persona.Deviation{0.2, 0.1, 0, -0.1, -0.2}},
	}}
	bonds := &stubBondStore{}

	result, err := SubmitMatchAssessment(context.Background(), repo,
		testutil.NewStubAssessRepo(), loader, bonds,
		SubmitMatchAssessmentInput{
			Token:   out.Token,
			Answers: makeTestAnswers("W", "F"),
		})
	if err != nil {
		t.Fatalf("SubmitMatchAssessment: %v", err)
	}
	if result.Bond == nil {
		t.Error("expected Bond when both profiles exist")
	}
	if len(bonds.inserted) != 1 {
		t.Errorf("expected 1 bond stored, got %d", len(bonds.inserted))
	}
}

func TestSubmitMatchAssessment_Answers_NoCreatorProfile(t *testing.T) {
	repo := newStubMLRepo()
	out, _ := CreateMatchLink(context.Background(), repo, 10, "assessment")
	loader := &testutil.StubProfileLoader{Profiles: map[int64]*domain.PersonalityProfile{}}
	bonds := &stubBondStore{}

	result, err := SubmitMatchAssessment(context.Background(), repo,
		testutil.NewStubAssessRepo(), loader, bonds,
		SubmitMatchAssessmentInput{
			Token:   out.Token,
			Answers: makeTestAnswers("W", "F"),
		})
	if err != nil {
		t.Fatalf("SubmitMatchAssessment: %v", err)
	}
	if result.Bond != nil {
		t.Error("expected no Bond when creator has no profile")
	}
}

func TestSubmitMatchAssessment_InvalidToken(t *testing.T) {
	repo := newStubMLRepo()
	_, err := SubmitMatchAssessment(context.Background(), repo,
		testutil.NewStubAssessRepo(), &testutil.StubProfileLoader{}, &stubBondStore{},
		SubmitMatchAssessmentInput{Token: "bad-token", Answers: makeTestAnswers("F", "E")})
	if !errors.Is(err, domain.ErrMatchLinkNotFound) {
		t.Errorf("expected ErrMatchLinkNotFound, got %v", err)
	}
}

func TestSubmitMatchAssessment_NoAnswersAndNoUseExisting(t *testing.T) {
	repo := newStubMLRepo()
	out, _ := CreateMatchLink(context.Background(), repo, 10, "assessment")
	_, err := SubmitMatchAssessment(context.Background(), repo,
		testutil.NewStubAssessRepo(), &testutil.StubProfileLoader{}, &stubBondStore{},
		SubmitMatchAssessmentInput{Token: out.Token})
	if !errors.Is(err, domain.ErrAnswersRequired) {
		t.Errorf("expected ErrAnswersRequired, got %v", err)
	}
}
