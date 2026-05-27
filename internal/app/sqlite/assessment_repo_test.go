package sqlite

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/25types/25types/internal/app/application/reviewlink"
	"github.com/25types/25types/internal/app/domain"
	persona "github.com/25types/25types/internal/25types"
)

func newTestAssessRepo(t *testing.T) *AssessmentRepo {
	t.Helper()
	return NewAssessmentRepo(openTestDB(t))
}

func newTestReviewLinkRepo(t *testing.T) *ReviewLinkRepo {
	t.Helper()
	return NewReviewLinkRepo(openTestDB(t))
}

func createTestAssessUser(t *testing.T, repo *UserRepo, name string) int64 {
	t.Helper()
	h := PasswordHasher{}
	hash, _ := h.Hash("testpass123")
	id, err := repo.Create(context.Background(), &domain.User{Name: name, PasswordHash: hash})
	if err != nil {
		t.Fatalf("Create user %q: %v", name, err)
	}
	return id
}

// =============================================================================
// CreateSelf + CreateAnonymous
// =============================================================================

func TestAssessmentRepo_CreateSelf(t *testing.T) {
	repo := newTestAssessRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "self-user")

	prof := domain.NewProfile(persona.Deviation{}, persona.Proportion{}, persona.Identity{ID: "IFS", Label: "IFS"})
	answers := []persona.Answer{{QID: "q1", Selections: []string{"A", "B"}}}
	id, err := repo.CreateSelf(context.Background(), uid, prof, answers)
	if err != nil {
		t.Fatalf("CreateSelf: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}

	// Verify it appears in ListSelf.
	items, total, _ := repo.ListSelf(context.Background(), uid, 0, 10)
	if total != 1 || len(items) != 1 {
		t.Errorf("ListSelf: total=%d len=%d, want 1", total, len(items))
	}
}

func TestAssessmentRepo_CreateAnonymous(t *testing.T) {
	repo := newTestAssessRepo(t)

	prof := domain.NewProfile(persona.Deviation{}, persona.Proportion{}, persona.Identity{ID: "FIS", Label: "FIS"})
	answers := []persona.Answer{{QID: "q1", Selections: []string{"W", "F"}}}
	id, err := repo.CreateAnonymous(context.Background(), prof, answers, "anon-token-123")
	if err != nil {
		t.Fatalf("CreateAnonymous: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

// =============================================================================
// FindLatestProfile
// =============================================================================

func TestAssessmentRepo_FindLatestProfile(t *testing.T) {
	repo := newTestAssessRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "profile-user")
	ctx := context.Background()

	prof1 := domain.NewProfile(persona.Deviation{-0.1, 0.2}, persona.Proportion{0.3, 0.7}, persona.Identity{ID: "IFS", Label: "IFS"})
	repo.CreateSelf(ctx, uid, prof1, []persona.Answer{{QID: "q1", Selections: []string{"A", "B"}}})

	prof2 := domain.NewProfile(persona.Deviation{0.5, -0.3}, persona.Proportion{0.6, 0.4}, persona.Identity{ID: "FIS", Label: "FIS"})
	repo.CreateSelf(ctx, uid, prof2, []persona.Answer{{QID: "q2", Selections: []string{"C", "D"}}})

	latest, err := repo.FindLatestProfile(ctx, uid)
	if err != nil {
		t.Fatalf("FindLatestProfile: %v", err)
	}
	if latest == nil {
		t.Fatal("expected profile, got nil")
	}
	if latest.Identity.ID != "FIS" {
		t.Errorf("latest Identity.ID = %q, want FIS", latest.Identity.ID)
	}
}

func TestAssessmentRepo_FindLatestProfile_NotFound(t *testing.T) {
	repo := newTestAssessRepo(t)
	_, err := repo.FindLatestProfile(context.Background(), 99999)
	if !errors.Is(err, domain.ErrNoProfile) {
		t.Errorf("expected ErrNoProfile, got %v", err)
	}
}

// =============================================================================
// ListSelf
// =============================================================================

func TestAssessmentRepo_ListSelf_Empty(t *testing.T) {
	repo := newTestAssessRepo(t)
	items, total, err := repo.ListSelf(context.Background(), 1, 0, 10)
	if err != nil {
		t.Fatalf("ListSelf: %v", err)
	}
	if total != 0 || len(items) != 0 {
		t.Errorf("expected empty, got total=%d items=%d", total, len(items))
	}
}

func TestAssessmentRepo_ListSelf_Pagination(t *testing.T) {
	repo := newTestAssessRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "page-user")
	ctx := context.Background()

	prof := domain.NewProfile(persona.Deviation{}, persona.Proportion{}, persona.Identity{ID: "IFS", Label: "IFS"})
	for i := 0; i < 5; i++ {
		repo.CreateSelf(ctx, uid, prof, []persona.Answer{{QID: "q1", Selections: []string{"A", "B"}}})
	}

	items, total, err := repo.ListSelf(ctx, uid, 0, 2)
	if err != nil {
		t.Fatalf("ListSelf: %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(items) != 2 {
		t.Errorf("limit 2 should return 2 items, got %d", len(items))
	}
}

// =============================================================================
// FindAssessmentByID + FindAssessmentByIDWithUser
// =============================================================================

func TestAssessmentRepo_FindAssessmentByID(t *testing.T) {
	repo := newTestAssessRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "find-by-id")
	ctx := context.Background()

	prof := domain.NewProfile(persona.Deviation{}, persona.Proportion{}, persona.Identity{ID: "IFS", Label: "IFS"})
	id, _ := repo.CreateSelf(ctx, uid, prof, []persona.Answer{{QID: "q1", Selections: []string{"A", "B"}}})

	a, err := repo.FindAssessmentByID(ctx, id)
	if err != nil {
		t.Fatalf("FindAssessmentByID: %v", err)
	}
	if a.ID != id {
		t.Errorf("ID = %d, want %d", a.ID, id)
	}
	if a.AssessmentType != domain.AssessSelf {
		t.Errorf("AssessmentType = %q, want self", a.AssessmentType)
	}
}

func TestAssessmentRepo_FindAssessmentByID_NotFound(t *testing.T) {
	repo := newTestAssessRepo(t)
	_, err := repo.FindAssessmentByID(context.Background(), 99999)
	if err == nil {
		t.Error("expected error for nonexistent assessment")
	}
}

func TestAssessmentRepo_FindAssessmentByIDWithUser(t *testing.T) {
	repo := newTestAssessRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "with-user")
	ctx := context.Background()

	prof := domain.NewProfile(persona.Deviation{}, persona.Proportion{}, persona.Identity{ID: "IFS", Label: "IFS"})
	id, _ := repo.CreateSelf(ctx, uid, prof, []persona.Answer{{QID: "q1", Selections: []string{"A", "B"}}})

	a, userName, err := repo.FindAssessmentByIDWithUser(ctx, id)
	if err != nil {
		t.Fatalf("FindAssessmentByIDWithUser: %v", err)
	}
	if a.ID != id {
		t.Errorf("ID = %d, want %d", a.ID, id)
	}
	if userName == nil || *userName != "with-user" {
		t.Errorf("userName = %v, want with-user", userName)
	}
}

// =============================================================================
// ClaimAnonymous
// =============================================================================

func TestAssessmentRepo_ClaimAnonymous(t *testing.T) {
	repo := newTestAssessRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "claim-user")
	ctx := context.Background()

	// Create anonymous assessment.
	prof := domain.NewProfile(persona.Deviation{}, persona.Proportion{}, persona.Identity{ID: "FIS", Label: "FIS"})
	repo.CreateAnonymous(ctx, prof, []persona.Answer{{QID: "q1", Selections: []string{"W", "F"}}}, "claim-token-xyz")

	claimed, err := repo.ClaimAnonymous(ctx, uid, "claim-token-xyz")
	if err != nil {
		t.Fatalf("ClaimAnonymous: %v", err)
	}
	if claimed != 1 {
		t.Errorf("claimed = %d, want 1", claimed)
	}

	// Second claim should claim 0 rows (already claimed).
	claimed2, _ := repo.ClaimAnonymous(ctx, uid, "claim-token-xyz")
	if claimed2 != 0 {
		t.Errorf("second claim = %d, want 0", claimed2)
	}
}

func TestAssessmentRepo_ClaimAnonymous_NoMatch(t *testing.T) {
	repo := newTestAssessRepo(t)
	n, err := repo.ClaimAnonymous(context.Background(), 1, "no-such-token")
	if err != nil {
		t.Fatalf("ClaimAnonymous: %v", err)
	}
	if n != 0 {
		t.Errorf("claimed = %d, want 0", n)
	}
}

// =============================================================================
// FindSelfAnswers
// =============================================================================

func TestAssessmentRepo_FindSelfAnswers(t *testing.T) {
	repo := newTestAssessRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "answers-user")
	ctx := context.Background()

	prof := domain.NewProfile(persona.Deviation{}, persona.Proportion{}, persona.Identity{ID: "IFS", Label: "IFS"})
	answers := []persona.Answer{{QID: "q1", Selections: []string{"A", "B"}}, {QID: "q2", Selections: []string{"C", "D"}}}
	repo.CreateSelf(ctx, uid, prof, answers)

	ans, err := repo.FindSelfAnswers(ctx, uid)
	if err != nil {
		t.Fatalf("FindSelfAnswers: %v", err)
	}
	if len(ans) != 2 {
		t.Errorf("len = %d, want 2", len(ans))
	}
}

func TestAssessmentRepo_FindSelfAnswers_NotFound(t *testing.T) {
	repo := newTestAssessRepo(t)
	_, err := repo.FindSelfAnswers(context.Background(), 99999)
	if err == nil {
		t.Error("expected error for nonexistent user")
	}
}

// =============================================================================
// ListPeerAnswersForUser
// =============================================================================

func TestAssessmentRepo_ListPeerAnswersForUser_Empty(t *testing.T) {
	repo := newTestAssessRepo(t)
	answers, count, err := repo.ListPeerAnswersForUser(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListPeerAnswersForUser: %v", err)
	}
	if count != 0 || len(answers) != 0 {
		t.Errorf("expected empty, got count=%d answers=%d", count, len(answers))
	}
}

func TestAssessmentRepo_ListPeerAnswersForUser(t *testing.T) {
	repo := newTestAssessRepo(t)
	rlRepo := newTestReviewLinkRepo(t)
	userRepo := newTestUserRepo(t)
	ctx := context.Background()
	subjectUID := createTestAssessUser(t, userRepo, "peer-subject")

	// Create a review link for subject.
	linkID, _ := rlRepo.CreateLink(ctx, subjectUID, "peer-link-tok", time.Now().UTC().Add(24*time.Hour).Format(time.RFC3339))

	// Submit a peer assessment.
	prof := domain.NewProfile(persona.Deviation{}, persona.Proportion{}, persona.Identity{ID: "IFS", Label: "IFS"})
	err := rlRepo.CreatePeerSubmission(ctx, &reviewlink.PeerSubmission{
		Profile:      prof,
		Answers:      []persona.Answer{{QID: "q1", Selections: []string{"A", "B"}}},
		ReviewLinkID: linkID,
		ReviewerName: "peer1",
	})
	if err != nil {
		t.Fatalf("CreatePeerSubmission: %v", err)
	}

	answers, count, err := repo.ListPeerAnswersForUser(ctx, subjectUID)
	if err != nil {
		t.Fatalf("ListPeerAnswersForUser: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	if len(answers) != 1 {
		t.Errorf("answers len = %d, want 1", len(answers))
	}
}

// =============================================================================
// ReviewLink: CreateLink, FindByToken, FindLinkByID, ListBySubject
// =============================================================================

func TestAssessmentRepo_ReviewLinkCRUD(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "link-subject")
	ctx := context.Background()

	exp := time.Now().UTC().Add(7 * 24 * time.Hour).Format(time.RFC3339)
	id, err := rlRepo.CreateLink(ctx, uid, "test-link-token", exp)
	if err != nil {
		t.Fatalf("CreateLink: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}

	// FindByToken.
	link, err := rlRepo.FindByToken(ctx, "test-link-token")
	if err != nil {
		t.Fatalf("FindByToken: %v", err)
	}
	if link.SubjectUserID != uid {
		t.Errorf("SubjectUserID = %d, want %d", link.SubjectUserID, uid)
	}
	if link.Token != "test-link-token" {
		t.Errorf("Token = %q, want test-link-token", link.Token)
	}

	// FindLinkByID.
	link2, err := rlRepo.FindLinkByID(ctx, id)
	if err != nil {
		t.Fatalf("FindLinkByID: %v", err)
	}
	if link2.ID != id {
		t.Errorf("ID = %d, want %d", link2.ID, id)
	}

	// ListBySubject.
	items, err := rlRepo.ListBySubject(ctx, uid)
	if err != nil {
		t.Fatalf("ListBySubject: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("ListBySubject len = %d, want 1", len(items))
	}
	if items[0].SubmissionCount != 0 {
		t.Errorf("SubmissionCount = %d, want 0", items[0].SubmissionCount)
	}
}

func TestAssessmentRepo_FindByToken_NotFound(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	_, err := rlRepo.FindByToken(context.Background(), "no-such-token")
	if err == nil {
		t.Error("expected error for nonexistent token")
	}
}

func TestAssessmentRepo_FindLinkByID_NotFound(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	_, err := rlRepo.FindLinkByID(context.Background(), 99999)
	if err == nil {
		t.Error("expected error for nonexistent link")
	}
}

func TestAssessmentRepo_ListBySubject_Empty(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	items, err := rlRepo.ListBySubject(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListBySubject: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty, got %d items", len(items))
	}
}

// =============================================================================
// SoftDelete + Renew
// =============================================================================

func TestAssessmentRepo_SoftDelete(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "delete-subj")
	ctx := context.Background()

	id, _ := rlRepo.CreateLink(ctx, uid, "del-token", time.Now().UTC().Add(24*time.Hour).Format(time.RFC3339))

	ok, err := rlRepo.SoftDelete(ctx, id, uid)
	if err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}
	if !ok {
		t.Error("SoftDelete should return true for active link")
	}

	// Second delete should return false (already deleted).
	ok2, _ := rlRepo.SoftDelete(ctx, id, uid)
	if ok2 {
		t.Error("SoftDelete should return false for already-deleted link")
	}

	// FindLinkByID should not find deleted link.
	_, err = rlRepo.FindLinkByID(ctx, id)
	if err == nil {
		t.Error("FindLinkByID should not find soft-deleted link")
	}
}

func TestAssessmentRepo_SoftDelete_WrongUser(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "owner-user")
	otherUID := createTestAssessUser(t, userRepo, "other-user")
	ctx := context.Background()

	id, _ := rlRepo.CreateLink(ctx, uid, "owner-token", time.Now().UTC().Add(24*time.Hour).Format(time.RFC3339))

	// Other user tries to delete — should fail.
	ok, err := rlRepo.SoftDelete(ctx, id, otherUID)
	if err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}
	if ok {
		t.Error("SoftDelete should return false when subjectUserID does not match")
	}
}

func TestAssessmentRepo_Renew(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "renew-subj")
	ctx := context.Background()

	id, _ := rlRepo.CreateLink(ctx, uid, "renew-token", time.Now().UTC().Add(24*time.Hour).Format(time.RFC3339))

	newExp := time.Now().UTC().Add(14 * 24 * time.Hour).Format(time.RFC3339)
	tok, ok, err := rlRepo.Renew(ctx, id, uid, newExp)
	if err != nil {
		t.Fatalf("Renew: %v", err)
	}
	if !ok {
		t.Error("Renew should succeed for active link")
	}
	if tok != "renew-token" {
		t.Errorf("token = %q, want renew-token", tok)
	}
}

func TestAssessmentRepo_Renew_Deleted(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "renew-del")
	ctx := context.Background()

	id, _ := rlRepo.CreateLink(ctx, uid, "renew-del-token", time.Now().UTC().Add(24*time.Hour).Format(time.RFC3339))
	rlRepo.SoftDelete(ctx, id, uid)

	_, ok, _ := rlRepo.Renew(ctx, id, uid, time.Now().UTC().Add(14*24*time.Hour).Format(time.RFC3339))
	if ok {
		t.Error("Renew should return false for deleted link")
	}
}

// =============================================================================
// CreatePeerSubmission
// =============================================================================

func TestAssessmentRepo_CreatePeerSubmission(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "peer-subj")
	ctx := context.Background()

	linkID, _ := rlRepo.CreateLink(ctx, uid, "peer-sub-token", time.Now().UTC().Add(24*time.Hour).Format(time.RFC3339))

	prof := domain.NewProfile(persona.Deviation{}, persona.Proportion{}, persona.Identity{ID: "IFS", Label: "IFS"})
	err := rlRepo.CreatePeerSubmission(ctx, &reviewlink.PeerSubmission{
		Profile:      prof,
		Answers:      []persona.Answer{{QID: "q1", Selections: []string{"A", "B"}}},
		ReviewLinkID: linkID,
		ReviewerName: "ReviewerName",
	})
	if err != nil {
		t.Fatalf("CreatePeerSubmission: %v", err)
	}

	// SubmissionCount should now be 1.
	items, _ := rlRepo.ListBySubject(ctx, uid)
	if items[0].SubmissionCount != 1 {
		t.Errorf("SubmissionCount = %d, want 1", items[0].SubmissionCount)
	}
}

// =============================================================================
// GetPeerQIDStats
// =============================================================================

func TestAssessmentRepo_GetPeerQIDStats(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "qid-subj")
	ctx := context.Background()

	linkID, _ := rlRepo.CreateLink(ctx, uid, "qid-token", time.Now().UTC().Add(24*time.Hour).Format(time.RFC3339))

	prof := domain.NewProfile(persona.Deviation{}, persona.Proportion{}, persona.Identity{ID: "IFS", Label: "IFS"})
	rlRepo.CreatePeerSubmission(ctx, &reviewlink.PeerSubmission{
		Profile:      prof,
		Answers:      []persona.Answer{{QID: "q1", Selections: []string{"A", "B"}}, {QID: "q2", Selections: []string{"C"}}},
		ReviewLinkID: linkID,
		ReviewerName: "r1",
	})
	rlRepo.CreatePeerSubmission(ctx, &reviewlink.PeerSubmission{
		Profile:      prof,
		Answers:      []persona.Answer{{QID: "q1", Selections: []string{"W"}}},
		ReviewLinkID: linkID,
		ReviewerName: "r2",
	})

	stats, err := rlRepo.GetPeerQIDStats(ctx, linkID)
	if err != nil {
		t.Fatalf("GetPeerQIDStats: %v", err)
	}
	if stats["q1"].Count != 2 {
		t.Errorf("q1 count = %d, want 2", stats["q1"].Count)
	}
	if stats["q2"].Count != 1 {
		t.Errorf("q2 count = %d, want 1", stats["q2"].Count)
	}
}

func TestAssessmentRepo_GetPeerQIDStats_Empty(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	stats, err := rlRepo.GetPeerQIDStats(context.Background(), 99999)
	if err != nil {
		t.Fatalf("GetPeerQIDStats: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected empty stats, got %d", len(stats))
	}
}

// =============================================================================
// GetSubjectName
// =============================================================================

func TestAssessmentRepo_GetSubjectName(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "subject-name-user")

	name := rlRepo.GetSubjectName(context.Background(), uid)
	if name != "subject-name-user" {
		t.Errorf("name = %q, want subject-name-user", name)
	}
}

func TestAssessmentRepo_GetSubjectName_NotFound(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	name := rlRepo.GetSubjectName(context.Background(), 99999)
	if name != "" {
		t.Errorf("expected empty string, got %q", name)
	}
}

// =============================================================================
// GetReviewSubmissions
// =============================================================================

func TestAssessmentRepo_GetReviewSubmissions_Empty(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	subs, err := rlRepo.GetReviewSubmissions(context.Background(), 99999)
	if err != nil {
		t.Fatalf("GetReviewSubmissions: %v", err)
	}
	if len(subs) != 0 {
		t.Errorf("expected empty, got %d", len(subs))
	}
}

func TestAssessmentRepo_GetReviewSubmissions(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestAssessUser(t, userRepo, "rev-subj")
	ctx := context.Background()

	linkID, _ := rlRepo.CreateLink(ctx, uid, "rev-sub-token", time.Now().UTC().Add(24*time.Hour).Format(time.RFC3339))

	prof := domain.NewProfile(persona.Deviation{}, persona.Proportion{}, persona.Identity{ID: "IFS", Label: "IFS"})
	rlRepo.CreatePeerSubmission(ctx, &reviewlink.PeerSubmission{
		Profile:      prof,
		Answers:      []persona.Answer{{QID: "q1", Selections: []string{"A"}}, {QID: "q2", Selections: []string{"B"}}},
		ReviewLinkID: linkID,
		ReviewerName: "peer-x",
	})

	subs, err := rlRepo.GetReviewSubmissions(ctx, linkID)
	if err != nil {
		t.Fatalf("GetReviewSubmissions: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 submission, got %d", len(subs))
	}
	if subs[0].ReviewerName != "peer-x" {
		t.Errorf("ReviewerName = %q, want peer-x", subs[0].ReviewerName)
	}
	if subs[0].AnsweredCount != 2 {
		t.Errorf("AnsweredCount = %d, want 2", subs[0].AnsweredCount)
	}
}

// =============================================================================
// ListReviewsGivenByUser + ListReviewsGivenByToken
// =============================================================================

func TestAssessmentRepo_ListReviewsGivenByUser_Empty(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	items, err := rlRepo.ListReviewsGivenByUser(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListReviewsGivenByUser: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty, got %d items", len(items))
	}
}

func TestAssessmentRepo_ListReviewsGivenByUser(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	userRepo := newTestUserRepo(t)
	ctx := context.Background()
	subjectUID := createTestAssessUser(t, userRepo, "review-subject")
	reviewerUID := createTestAssessUser(t, userRepo, "reviewer-user")

	linkID, _ := rlRepo.CreateLink(ctx, subjectUID, "rev-given-token", time.Now().UTC().Add(24*time.Hour).Format(time.RFC3339))

	prof := domain.NewProfile(persona.Deviation{}, persona.Proportion{}, persona.Identity{ID: "IFS", Label: "IFS"})
	rlRepo.CreatePeerSubmission(ctx, &reviewlink.PeerSubmission{
		UserID:       &reviewerUID,
		Profile:      prof,
		Answers:      []persona.Answer{{QID: "q1", Selections: []string{"A"}}},
		ReviewLinkID: linkID,
		ReviewerName: "the-reviewer",
	})

	items, err := rlRepo.ListReviewsGivenByUser(ctx, reviewerUID)
	if err != nil {
		t.Fatalf("ListReviewsGivenByUser: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].SubjectName != "review-subject" {
		t.Errorf("SubjectName = %q, want review-subject", items[0].SubjectName)
	}
	if items[0].AnsweredCount != 1 {
		t.Errorf("AnsweredCount = %d, want 1", items[0].AnsweredCount)
	}
}

func TestAssessmentRepo_ListReviewsGivenByToken(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	userRepo := newTestUserRepo(t)
	ctx := context.Background()
	subjectUID := createTestAssessUser(t, userRepo, "token-subject")

	linkID, _ := rlRepo.CreateLink(ctx, subjectUID, "token-rev-link", time.Now().UTC().Add(24*time.Hour).Format(time.RFC3339))

	prof := domain.NewProfile(persona.Deviation{}, persona.Proportion{}, persona.Identity{ID: "IFS", Label: "IFS"})
	rlRepo.CreatePeerSubmission(ctx, &reviewlink.PeerSubmission{
		AnonymousToken: "anon-reviewer-token",
		Profile:        prof,
		Answers:        []persona.Answer{{QID: "q1", Selections: []string{"W"}}},
		ReviewLinkID:   linkID,
		ReviewerName:   "anon-reviewer",
	})

	items, err := rlRepo.ListReviewsGivenByToken(ctx, "anon-reviewer-token")
	if err != nil {
		t.Fatalf("ListReviewsGivenByToken: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].SubjectName != "token-subject" {
		t.Errorf("SubjectName = %q, want token-subject", items[0].SubjectName)
	}
}

func TestAssessmentRepo_ListReviewsGivenByToken_Empty(t *testing.T) {
	rlRepo := newTestReviewLinkRepo(t)
	items, err := rlRepo.ListReviewsGivenByToken(context.Background(), "no-such-token")
	if err != nil {
		t.Fatalf("ListReviewsGivenByToken: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty, got %d items", len(items))
	}
}

// =============================================================================
// LoadProfile (domain.ProfileLoader)
// =============================================================================

func TestAssessmentRepo_LoadProfile(t *testing.T) {
	repo := newTestAssessRepo(t)
	userRepo := newTestUserRepo(t)
	uid := createTestUser(t, userRepo, "profile-loader")
	ctx := context.Background()

	prof := domain.NewProfile(persona.Deviation{}, persona.Proportion{}, persona.Identity{ID: "IFS", Label: "IFS"})
	repo.CreateSelf(ctx, uid, prof, []persona.Answer{{QID: "q1", Selections: []string{"A", "B"}}})

	loaded, err := repo.LoadProfile(ctx, uid)
	if err != nil {
		t.Fatalf("LoadProfile: %v", err)
	}
	if loaded.Identity.ID != "IFS" {
		t.Errorf("Identity.ID = %q, want IFS", loaded.Identity.ID)
	}
}

func TestAssessmentRepo_LoadProfile_NotFound(t *testing.T) {
	repo := newTestAssessRepo(t)
	_, err := repo.LoadProfile(context.Background(), 99999)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}
