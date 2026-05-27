package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/25types/25types/internal/app/application/assessment"
	"github.com/25types/25types/internal/app/domain"
	persona "github.com/25types/25types/internal/25types"
	"github.com/25types/25types/internal/25types/questionnaire"
)

// UserPublicLookup checks whether a user's profile is public.
type UserPublicLookup interface {
	IsPublicByID(ctx context.Context, userID int64) (bool, error)
}

// AssessmentHandler holds dependencies for assessment HTTP handlers.
type AssessmentHandler struct {
	Repo       assessment.AssessmentRepository
	UserLookup UserPublicLookup
}

type submitAssessmentResponse struct {
	ID             int64                     `json:"id,omitempty"`
	Profile        domain.PersonalityProfile `json:"profile"`
	Identity       persona.Identity          `json:"identity"`
	Complete       bool                      `json:"complete"`
	AnonymousToken string                    `json:"anonymous_token,omitempty"`
}

// POST /api/assessments
func (h *AssessmentHandler) SubmitAssessment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Answers        []persona.Answer `json:"answers"`
		AnonymousToken string          `json:"anonymous_token,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Answers) == 0 {
		respondError(w, http.StatusBadRequest, "invalid_request", "answers is required")
		return
	}

	var userID *int64
	uid, authed := UserID(r.Context())
	if authed {
		userID = &uid
	}

	result, err := assessment.SubmitAssessmentUseCase(r.Context(), h.Repo, req.Answers, userID, req.AnonymousToken)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}

	resp := submitAssessmentResponse{
		Profile:        result.Profile,
		Identity:       result.Identity,
		Complete:       len(req.Answers) >= 30,
		ID:             result.ID,
		AnonymousToken: result.AnonToken,
	}
	status := http.StatusOK
	if result.ID != 0 {
		status = http.StatusCreated
	}
	respondJSON(w, status, resp)
}

type assessmentQuestionsResponse struct {
	Rounds []assessmentRoundOut `json:"rounds"`
}

type assessmentRoundOut struct {
	ID        string                   `json:"id"`
	Questions []questionnaire.Question `json:"questions"`
}

// GET /api/assessments/questions
func (h *AssessmentHandler) GetQuestions(w http.ResponseWriter, r *http.Request) {
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	q, err := questionnaire.Load(locale)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal", "Failed to load questions")
		return
	}

	rounds := make([]assessmentRoundOut, len(q.Rounds))
	for i, r := range q.Rounds {
		rounds[i] = assessmentRoundOut{ID: r.ID, Questions: r.Questions}
	}

	respondJSON(w, http.StatusOK, assessmentQuestionsResponse{Rounds: rounds})
}

// GET /api/assessments
func (h *AssessmentHandler) ListAssessments(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	items, total, err := h.Repo.ListSelf(r.Context(), uid, offset, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}

	type out struct {
		ID        int64       `json:"id"`
		Profile   interface{} `json:"profile"`
		Identity  interface{} `json:"identity"`
		CreatedAt string      `json:"created_at"`
	}
	var output []out
	for _, a := range items {
		var prof interface{}
		json.Unmarshal([]byte(a.ProfileJSON), &prof)
		output = append(output, out{
			ID:        a.ID,
			Profile:   prof,
			Identity:  persona.Identity{Label: a.IdentityID, ID: a.IdentityID, Category: persona.DeriveCategory(a.IdentityID)},
			CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	respondList(w, output, total)
}

type assessmentDetailResponse struct {
	ID        int64            `json:"id"`
	Profile   interface{}      `json:"profile"`
	Identity  persona.Identity `json:"identity"`
	CreatedAt string           `json:"created_at"`
	UserName  *string          `json:"user_name,omitempty"`
}

// GET /api/assessments/{id}
func (h *AssessmentHandler) GetAssessment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	aid, _ := strconv.ParseInt(idStr, 10, 64)

	a, userName, err := h.Repo.FindAssessmentByIDWithUser(r.Context(), aid)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Assessment not found")
		return
	}

	// Visibility check: only the owner or public profiles are visible.
	if a.UserID != nil {
		viewerID, authed := UserID(r.Context())
		if !authed || viewerID != *a.UserID {
			if h.UserLookup == nil {
				respondError(w, http.StatusNotFound, "not_found", "Assessment not found")
				return
			}
			isPublic, err := h.UserLookup.IsPublicByID(r.Context(), *a.UserID)
			if err != nil || !isPublic {
				respondError(w, http.StatusNotFound, "not_found", "Assessment not found")
				return
			}
		}
	}

	var prof interface{}
	json.Unmarshal([]byte(a.ProfileJSON), &prof)

	respondJSON(w, http.StatusOK, assessmentDetailResponse{
		ID:        aid,
		Profile:   prof,
		Identity:  persona.Identity{Label: a.IdentityID, ID: a.IdentityID, Category: persona.DeriveCategory(a.IdentityID)},
		CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UserName:  userName,
	})
}

// GET /api/assessments/peers
func (h *AssessmentHandler) Peers(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())

	result, err := assessment.PeersUseCase(r.Context(), h.Repo, uid)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "No profile found")
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// POST /api/assessments/claim
func (h *AssessmentHandler) Claim(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())

	var req struct {
		AnonymousToken string `json:"anonymous_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AnonymousToken == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "anonymous_token is required")
		return
	}

	n, err := assessment.ClaimUseCase(r.Context(), h.Repo, uid, req.AnonymousToken)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}

	respondJSON(w, http.StatusOK, map[string]int64{"claimed": n})
}
