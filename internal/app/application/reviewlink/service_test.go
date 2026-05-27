package reviewlink

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/25types/25types/internal/app/domain"
	persona "github.com/25types/25types/internal/25types"
)

// =============================================================================
// Stubs
// =============================================================================

type stubReviewLinkRepo struct {
	links  map[string]*domain.ReviewLink
	nextID int64
	name   string
	subErr error
}

func newStubReviewLinkRepo() *stubReviewLinkRepo {
	return &stubReviewLinkRepo{
		links: map[string]*domain.ReviewLink{},
	}
}

func (r *stubReviewLinkRepo) CreateLink(ctx context.Context, subjectUserID int64, token string, expiresAt string) (int64, error) {
	exp, _ := time.Parse(time.RFC3339, expiresAt)
	r.nextID++
	r.links[token] = &domain.ReviewLink{
		ID:            r.nextID,
		SubjectUserID: subjectUserID,
		Token:         token,
		ExpiresAt:     exp,
	}
	return r.nextID, nil
}

func (r *stubReviewLinkRepo) FindByToken(ctx context.Context, token string) (*domain.ReviewLink, error) {
	l, ok := r.links[token]
	if !ok {
		return nil, errors.New("not found")
	}
	return l, nil
}

func (r *stubReviewLinkRepo) FindLinkByID(ctx context.Context, id int64) (*domain.ReviewLink, error) {
	return nil, errors.New("not found")
}

func (r *stubReviewLinkRepo) ListBySubject(ctx context.Context, subjectUserID int64) ([]ReviewLinkItem, error) {
	return nil, nil
}

func (r *stubReviewLinkRepo) SoftDelete(ctx context.Context, id int64, subjectUserID int64) (bool, error) {
	for _, l := range r.links {
		if l.ID == id && l.SubjectUserID == subjectUserID {
			now := time.Now()
			l.DeletedAt = &now
			return true, nil
		}
	}
	return false, nil
}

func (r *stubReviewLinkRepo) Renew(ctx context.Context, id int64, subjectUserID int64, newExpires string) (string, bool, error) {
	return "", false, nil
}

func (r *stubReviewLinkRepo) CreatePeerSubmission(ctx context.Context, sub *PeerSubmission) error {
	return r.subErr
}

func (r *stubReviewLinkRepo) GetPeerQIDStats(ctx context.Context, linkID int64) (map[string]domain.QIDStat, error) {
	return nil, nil
}

func (r *stubReviewLinkRepo) GetSubjectName(ctx context.Context, subjectUserID int64) string {
	return r.name
}

func (r *stubReviewLinkRepo) GetReviewSubmissions(ctx context.Context, linkID int64) ([]ReviewSubmissionItem, error) {
	return nil, nil
}

func (r *stubReviewLinkRepo) ListReviewsGivenByUser(ctx context.Context, userID int64) ([]ReviewsGivenItem, error) {
	return nil, nil
}

func (r *stubReviewLinkRepo) ListReviewsGivenByToken(ctx context.Context, anonToken string) ([]ReviewsGivenItem, error) {
	return nil, nil
}

func makeAnswer(qid, a, b string) persona.Answer {
	return persona.Answer{QID: qid, Selections: []string{a, b}}
}

// =============================================================================
// CreateReviewLinkUseCase
// =============================================================================

func TestCreateReviewLink_OK(t *testing.T) {
	repo := newStubReviewLinkRepo()
	result, err := CreateReviewLinkUseCase(context.Background(), repo, 1)
	if err != nil {
		t.Fatalf("CreateReviewLink: %v", err)
	}
	if result.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if result.Token == "" {
		t.Error("expected non-empty token")
	}
}

// =============================================================================
// GetLinkInfoUseCase
// =============================================================================

func TestGetLinkInfo_Valid(t *testing.T) {
	repo := newStubReviewLinkRepo()
	repo.name = "Alice"
	link, _ := CreateReviewLinkUseCase(context.Background(), repo, 1)

	info, err := GetLinkInfoUseCase(context.Background(), repo, repo, link.Token, "en")
	if err != nil {
		t.Fatalf("GetLinkInfo: %v", err)
	}
	if !info.Valid {
		t.Error("expected valid link")
	}
	if info.SubjectName != "Alice" {
		t.Errorf("SubjectName = %q, want Alice", info.SubjectName)
	}
}

func TestGetLinkInfo_NotFound(t *testing.T) {
	repo := newStubReviewLinkRepo()
	_, err := GetLinkInfoUseCase(context.Background(), repo, repo, "nonexistent", "en")
	if !errors.Is(err, domain.ErrLinkNotFound) {
		t.Errorf("expected ErrLinkNotFound, got %v", err)
	}
}

func TestGetLinkInfo_Deleted(t *testing.T) {
	repo := newStubReviewLinkRepo()
	link, _ := CreateReviewLinkUseCase(context.Background(), repo, 1)
	repo.SoftDelete(context.Background(), link.ID, 1)

	_, err := GetLinkInfoUseCase(context.Background(), repo, repo, link.Token, "en")
	if !errors.Is(err, domain.ErrLinkNotFound) {
		t.Errorf("expected ErrLinkNotFound for deleted link, got %v", err)
	}
}

func TestGetLinkInfo_Expired(t *testing.T) {
	repo := newStubReviewLinkRepo()
	repo.name = "Bob"
	token := "expired-token"
	repo.links[token] = &domain.ReviewLink{
		ID:            1,
		SubjectUserID: 1,
		Token:         token,
		ExpiresAt:     time.Now().Add(-1 * time.Hour),
	}

	info, err := GetLinkInfoUseCase(context.Background(), repo, repo, token, "en")
	if err != nil {
		t.Fatalf("GetLinkInfo expired: %v", err)
	}
	if info.Valid {
		t.Error("expected invalid (expired) link")
	}
	if !info.Expired {
		t.Error("expected expired=true")
	}
}

// =============================================================================
// SubmitPeerReviewUseCase
// =============================================================================

func TestSubmitPeerReview_OK(t *testing.T) {
	linkRepo := newStubReviewLinkRepo()
	link, _ := CreateReviewLinkUseCase(context.Background(), linkRepo, 1)

	answers := []persona.Answer{makeAnswer("Q01", "W", "F"), makeAnswer("Q02", "E", "M")}
	result, err := SubmitPeerReviewUseCase(context.Background(), linkRepo, linkRepo, SubmitPeerReviewInput{Token: link.Token, ReviewerName: "Reviewer", Answers: answers})
	if err != nil {
		t.Fatalf("SubmitPeerReview: %v", err)
	}
	if result.SubjectIdentity.ID == "" {
		t.Error("expected non-empty identity ID")
	}
}

func TestSubmitPeerReview_NoAnswers(t *testing.T) {
	_, err := SubmitPeerReviewUseCase(context.Background(), nil, nil, SubmitPeerReviewInput{Token: "tok", ReviewerName: "R"})
	if !errors.Is(err, domain.ErrAnswersRequired) {
		t.Errorf("expected ErrAnswersRequired, got %v", err)
	}
}

func TestSubmitPeerReview_ExpiredLink(t *testing.T) {
	repo := newStubReviewLinkRepo()
	repo.links["expired"] = &domain.ReviewLink{
		ID:            1,
		SubjectUserID: 1,
		Token:         "expired",
		ExpiresAt:     time.Now().Add(-1 * time.Hour),
	}
	answers := []persona.Answer{makeAnswer("Q01", "W", "F")}

	_, err := SubmitPeerReviewUseCase(context.Background(), repo, repo, SubmitPeerReviewInput{Token: "expired", ReviewerName: "R", Answers: answers})
	if !errors.Is(err, domain.ErrLinkExpired) {
		t.Errorf("expected ErrLinkExpired, got %v", err)
	}
}

func TestSubmitPeerReview_DeletedLink(t *testing.T) {
	repo := newStubReviewLinkRepo()
	link, _ := CreateReviewLinkUseCase(context.Background(), repo, 1)
	repo.SoftDelete(context.Background(), link.ID, 1)

	answers := []persona.Answer{makeAnswer("Q01", "W", "F")}
	_, err := SubmitPeerReviewUseCase(context.Background(), repo, repo, SubmitPeerReviewInput{Token: link.Token, ReviewerName: "R", Answers: answers})
	if !errors.Is(err, domain.ErrLinkNotFound) {
		t.Errorf("expected ErrLinkNotFound, got %v", err)
	}
}
