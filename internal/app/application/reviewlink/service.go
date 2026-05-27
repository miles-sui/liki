package reviewlink

import (
	"context"
	"sort"
	"time"

	"github.com/25types/25types/internal/25types"
	"github.com/25types/25types/internal/app/domain"
	"github.com/25types/25types/internal/25types/questionnaire"
)

// CreateReviewLinkUseCase creates a new peer review link.
func CreateReviewLinkUseCase(
	ctx context.Context,
	repo ReviewLinkRepository,
	subjectUserID int64,
) (*CreatedLink, error) {
	token := domain.GenerateToken()
	expires := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)

	id, err := repo.CreateLink(ctx, subjectUserID, token, expires)
	if err != nil {
		return nil, err
	}

	return &CreatedLink{
		ID:        id,
		Token:     token,
		ExpiresAt: expires,
	}, nil
}

// CreatedLink holds the output of link creation.
type CreatedLink struct {
	ID        int64
	Token     string
	ExpiresAt string
}

// GetLinkInfoUseCase returns link validity and recommended QIDs.
func GetLinkInfoUseCase(
	ctx context.Context,
	linkRepo ReviewLinkRepository,
	subRepo ReviewSubmissionRepository,
	token string,
	locale string,
) (*LinkInfo, error) {
	link, err := linkRepo.FindByToken(ctx, token)
	if err != nil || link.IsDeleted() {
		return nil, domain.ErrLinkNotFound
	}

	name := linkRepo.GetSubjectName(ctx, link.SubjectUserID)

	info := &LinkInfo{
		SubjectName: name,
		Valid:       true,
	}

	if link.IsExpired(time.Now()) {
		info.Valid = false
		info.Expired = true
	}

	if info.Valid {
		info.RecommendedQIDs = recommendPeerQIDs(ctx, subRepo, link.ID)
		info.Questions = resolveQuestions(info.RecommendedQIDs, locale)
	}

	return info, nil
}

// LinkInfo holds the output of link info query.
type LinkInfo struct {
	SubjectName     string
	Valid           bool
	Expired         bool
	RecommendedQIDs []string
	Questions       []questionnaire.Question
}

// SubmitPeerReviewInput holds the parameters for SubmitPeerReviewUseCase.
type SubmitPeerReviewInput struct {
	Token        string
	ReviewerName string
	Answers      []persona.Answer
	UserID       *int64
	AnonToken    string
}

// SubmitPeerReviewUseCase processes a peer review submission.
func SubmitPeerReviewUseCase(
	ctx context.Context,
	linkRepo ReviewLinkRepository,
	subRepo ReviewSubmissionRepository,
	input SubmitPeerReviewInput,
) (*PeerSubmitResult, error) {
	if len(input.Answers) == 0 {
		return nil, domain.ErrAnswersRequired
	}

	link, err := linkRepo.FindByToken(ctx, input.Token)
	if err != nil || link.IsDeleted() {
		return nil, domain.ErrLinkNotFound
	}
	if link.IsExpired(time.Now()) {
		return nil, domain.ErrLinkExpired
	}

	if input.ReviewerName == "" && input.UserID != nil {
		input.ReviewerName = linkRepo.GetSubjectName(ctx, *input.UserID)
	}

	profile := domain.ComputeProfileFromAnswers(input.Answers)

	sub := &PeerSubmission{
		UserID:         input.UserID,
		Profile:        profile,
		Answers:        input.Answers,
		ReviewLinkID:   link.ID,
		ReviewerName:   input.ReviewerName,
		AnonymousToken: input.AnonToken,
	}

	if err := subRepo.CreatePeerSubmission(ctx, sub); err != nil {
		return nil, err
	}

	return &PeerSubmitResult{
		SubjectIdentity: profile.Identity,
		AnonToken:       input.AnonToken,
	}, nil
}

// PeerSubmitResult holds the output of a peer review submission.
type PeerSubmitResult struct {
	SubjectIdentity persona.Identity
	AnonToken       string
}

// resolveQuestions returns full Question objects for the given QIDs from the given locale.
func resolveQuestions(qids []string, locale string) []questionnaire.Question {
	q, err := questionnaire.Load(locale)
	if err != nil {
		return nil
	}
	return questionnaire.GetQuestions(q, qids)
}

// recommendPeerQIDs selects 5 recommended questions for peer review.
func recommendPeerQIDs(ctx context.Context, subRepo ReviewSubmissionRepository, linkID int64) []string {
	allQIDs := questionnaire.AllQIDs()
	stats, _ := subRepo.GetPeerQIDStats(ctx, linkID)

	type cand struct {
		id    string
		count int
		maxV  int
	}
	var candidates []cand
	for _, qid := range allQIDs {
		s := stats[qid]
		candidates = append(candidates, cand{id: qid, count: s.Count, maxV: s.MaxVote})
	}

	// Sort: lower count first, then lower maxV (补缺 > 覆盖 > 争议).
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].count < candidates[j].count ||
			(candidates[i].count == candidates[j].count && candidates[i].maxV < candidates[j].maxV)
	})

	n := 5
	if n > len(candidates) {
		n = len(candidates)
	}
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = candidates[i].id
	}
	return out
}
