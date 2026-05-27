package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/25types/25types/internal/app/application/reviewlink"
	"github.com/25types/25types/internal/app/domain"
	persona "github.com/25types/25types/internal/25types"
	"github.com/25types/25types/internal/25types/questionnaire"
)

// ReviewHandler holds dependencies for review-link HTTP handlers.
type ReviewHandler struct {
	LinkRepo reviewlink.ReviewLinkRepository
	SubRepo  reviewlink.ReviewSubmissionRepository
}

type reviewLinkResponse struct {
	ID        int64  `json:"id"`
	Token     string `json:"token"`
	URL       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
}

type reviewDetailResponse struct {
	ID              int64                     `json:"id"`
	Token           string                    `json:"token"`
	URL             string                    `json:"url"`
	SubjectName     string                    `json:"subject_name"`
	SubmissionCount int                       `json:"submission_count"`
	ExpiresAt       string                    `json:"expires_at"`
	CreatedAt       string                    `json:"created_at"`
	Submissions     []reviewSubmissionItemOut `json:"submissions"`
}

type reviewSubmissionItemOut struct {
	ReviewerName    string `json:"reviewer_name"`
	AnsweredCount   int    `json:"answered_count"`
	LastSubmittedAt string `json:"last_submitted_at"`
}

type reviewRenewResponse struct {
	ID        int64  `json:"id"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

type reviewLinkInfoResponse struct {
	SubjectName     string                   `json:"subject_name"`
	Valid           bool                     `json:"valid"`
	Expired         bool                     `json:"expired"`
	RecommendedQIDs []string                 `json:"recommended_qids"`
	Questions       []questionnaire.Question `json:"questions"`
}

type reviewSubmitResponse struct {
	SubjectIdentity persona.Identity `json:"subject_identity"`
	AnonymousToken  string           `json:"anonymous_token,omitempty"`
}

// POST /api/reviews
func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())

	link, err := reviewlink.CreateReviewLinkUseCase(r.Context(), h.LinkRepo, uid)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}

	respondJSON(w, http.StatusCreated, reviewLinkResponse{
		ID: link.ID, Token: link.Token,
		URL: "/r/" + link.Token, ExpiresAt: link.ExpiresAt,
	})
}

// GET /api/reviews
func (h *ReviewHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())

	items, err := h.LinkRepo.ListBySubject(r.Context(), uid)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}

	name := h.LinkRepo.GetSubjectName(r.Context(), uid)
	type out struct {
		ID              int64  `json:"id"`
		Token           string `json:"token"`
		URL             string `json:"url"`
		SubjectName     string `json:"subject_name,omitempty"`
		SubmissionCount int    `json:"submission_count,omitempty"`
		ExpiresAt       string `json:"expires_at,omitempty"`
		CreatedAt       string `json:"created_at,omitempty"`
	}
	var output []out
	for _, it := range items {
		output = append(output, out{
			ID:              it.ID,
			Token:           it.Token,
			URL:             "/r/" + it.Token,
			SubjectName:     name,
			SubmissionCount: it.SubmissionCount,
			ExpiresAt:       it.ExpiresAt,
			CreatedAt:       it.CreatedAt,
		})
	}
	respondList(w, output, len(output))
}

// GET /api/reviews/{id}
func (h *ReviewHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())
	idStr := r.PathValue("id")
	rid, _ := strconv.ParseInt(idStr, 10, 64)

	link, err := h.LinkRepo.FindLinkByID(r.Context(), rid)
	if err != nil || link.SubjectUserID != uid {
		respondError(w, http.StatusNotFound, "not_found", "Review link not found")
		return
	}

	name := h.LinkRepo.GetSubjectName(r.Context(), uid)

	subs, _ := h.SubRepo.GetReviewSubmissions(r.Context(), rid)

	var subItems []reviewSubmissionItemOut
	for _, s := range subs {
		subItems = append(subItems, reviewSubmissionItemOut{
			ReviewerName:    s.ReviewerName,
			AnsweredCount:   s.AnsweredCount,
			LastSubmittedAt: s.LastSubmittedAt,
		})
	}
	if subItems == nil {
		subItems = []reviewSubmissionItemOut{}
	}

	respondJSON(w, http.StatusOK, reviewDetailResponse{
		ID: rid, Token: link.Token, URL: "/r/" + link.Token,
		SubjectName: name, SubmissionCount: len(subItems),
		ExpiresAt:   link.ExpiresAt.Format(time.RFC3339),
		CreatedAt:   link.CreatedAt.Format(time.RFC3339),
		Submissions: subItems,
	})
}

// DELETE /api/reviews/{id}
func (h *ReviewHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())
	idStr := r.PathValue("id")
	rid, _ := strconv.ParseInt(idStr, 10, 64)

	ok, err := h.LinkRepo.SoftDelete(r.Context(), rid, uid)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}
	if !ok {
		respondError(w, http.StatusNotFound, "not_found", "Review link not found")
		return
	}
	respondStatus(w, http.StatusOK, "deleted")
}

// POST /api/reviews/{id}/renew
func (h *ReviewHandler) RenewReview(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())
	idStr := r.PathValue("id")
	rid, _ := strconv.ParseInt(idStr, 10, 64)

	newExp := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	token, ok, err := h.LinkRepo.Renew(r.Context(), rid, uid, newExp)
	if err != nil || !ok {
		respondError(w, http.StatusNotFound, "not_found", "Review link not found")
		return
	}

	respondJSON(w, http.StatusOK, reviewRenewResponse{
		ID: rid, Token: token, ExpiresAt: newExp,
	})
}

// GET /api/reviews/given
func (h *ReviewHandler) ListReviewsGiven(w http.ResponseWriter, r *http.Request) {
	anonToken := r.URL.Query().Get("anonymous_token")

	var items []reviewlink.ReviewsGivenItem
	if uid, authed := UserID(r.Context()); authed {
		items, _ = h.SubRepo.ListReviewsGivenByUser(r.Context(), uid)
	} else if anonToken != "" {
		items, _ = h.SubRepo.ListReviewsGivenByToken(r.Context(), anonToken)
	}

	type out struct {
		SubjectName   string `json:"subject_name"`
		AnsweredCount int    `json:"answered_count"`
		CreatedAt     string `json:"created_at"`
	}
	var output []out
	for _, it := range items {
		output = append(output, out{
			SubjectName:   it.SubjectName,
			AnsweredCount: it.AnsweredCount,
			CreatedAt:     it.CreatedAt,
		})
	}
	if output == nil {
		output = []out{}
	}
	respondList(w, output, len(output))
}

// GET /api/r/{token}
func (h *ReviewHandler) GetLinkInfo(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	info, err := reviewlink.GetLinkInfoUseCase(r.Context(), h.LinkRepo, h.SubRepo, token, locale)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Review link not found")
		return
	}

	q := info.Questions
	if q == nil {
		q = []questionnaire.Question{}
	}
	respondJSON(w, http.StatusOK, reviewLinkInfoResponse{
		SubjectName:     info.SubjectName,
		Valid:           info.Valid,
		Expired:         info.Expired,
		RecommendedQIDs: info.RecommendedQIDs,
		Questions:       q,
	})
}

// POST /api/r/{token}
func (h *ReviewHandler) SubmitReview(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	link, err := h.LinkRepo.FindByToken(r.Context(), token)
	if err != nil || link.IsDeleted() {
		respondError(w, http.StatusNotFound, "not_found", "Review link not found")
		return
	}
	if link.IsExpired(time.Now()) {
		respondError(w, http.StatusNotFound, "not_found", "Review link has expired")
		return
	}

	var req struct {
		ReviewerName   string          `json:"reviewer_name"`
		Answers        []persona.Answer `json:"answers"`
		AnonymousToken string          `json:"anonymous_token,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Answers) == 0 {
		respondError(w, http.StatusBadRequest, "invalid_request", "answers is required")
		return
	}

	reviewerName := req.ReviewerName
	if reviewerName == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "reviewer_name is required")
		return
	}
	var userID *int64
	uid, authed := UserID(r.Context())
	if authed {
		userID = &uid
	}

	result, err := reviewlink.SubmitPeerReviewUseCase(
		r.Context(), h.LinkRepo, h.SubRepo, reviewlink.SubmitPeerReviewInput{
			Token:        token,
			ReviewerName: reviewerName,
			Answers:      req.Answers,
			UserID:       userID,
			AnonToken:    req.AnonymousToken,
		},
	)
	if err != nil {
		if errors.Is(err, domain.ErrAnswersRequired) {
			respondError(w, http.StatusBadRequest, "invalid_request", "answers is required")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}

	respondJSON(w, http.StatusCreated, reviewSubmitResponse{
		SubjectIdentity: result.SubjectIdentity,
		AnonymousToken:  result.AnonToken,
	})
}
