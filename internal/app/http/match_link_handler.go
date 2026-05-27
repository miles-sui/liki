package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/25types/25types/internal/app/application/assessment"
	"github.com/25types/25types/internal/app/application/matchlink"
	"github.com/25types/25types/internal/app/application/profile"
	"github.com/25types/25types/internal/app/application/user"
	"github.com/25types/25types/internal/app/domain"
	persona "github.com/25types/25types/internal/25types"
)

// MatchLinkHandler holds dependencies for assessment match link HTTP endpoints.
type MatchLinkHandler struct {
	Repo          matchlink.MatchLinkRepository
	Assessments   assessment.AssessmentRepository
	ProfileLoader domain.ProfileLoader
	Bonds         profile.BondStore
	EmailSender   user.EmailSender
	UserRepo      user.UserRepository
}

// POST /api/match-links
func (h *MatchLinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())
	out, err := matchlink.CreateMatchLink(r.Context(), h.Repo, uid, "assessment")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}
	respondJSON(w, http.StatusCreated, out)
}

// GET /api/match-links
func (h *MatchLinkHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())
	items, err := matchlink.ListMatchLinks(r.Context(), h.Repo, uid, "assessment")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}
	respondList(w, items, len(items))
}

// DELETE /api/match-links/{id}
func (h *MatchLinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid ID")
		return
	}

	if err := matchlink.DeleteMatchLink(r.Context(), h.Repo, id, uid); err != nil {
		if errors.Is(err, domain.ErrMatchLinkNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "Match link not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}
	respondStatus(w, http.StatusOK, "deleted")
}

// GET /api/m/{token}
func (h *MatchLinkHandler) GetLinkInfo(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	info, err := matchlink.GetMatchLink(r.Context(), h.Repo, token)
	if err != nil {
		if errors.Is(err, domain.ErrMatchLinkNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "Match link not found or deleted")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}
	// Look up creator name for display (best-effort).
	if h.UserRepo != nil {
		if u, lookupErr := h.UserRepo.FindByID(r.Context(), info.CreatorUserID); lookupErr == nil {
			info.CreatorName = u.Name
		}
	}
	respondJSON(w, http.StatusOK, info)
}

// POST /api/m/{token}
func (h *MatchLinkHandler) SubmitAssessment(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	var req struct {
		Answers        []persona.Answer `json:"answers"`
		AnonymousToken string          `json:"anonymous_token"`
		UseExisting    bool            `json:"use_existing"`
		OtherName      string          `json:"other_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Use existing profile requires authentication.
	var userID *int64
	if req.UseExisting {
		uid, ok := UserID(r.Context())
		if !ok {
			respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
			return
		}
		userID = &uid
	} else if uid, ok := UserID(r.Context()); ok {
		userID = &uid
	}

	input := matchlink.SubmitMatchAssessmentInput{
		Token:       token,
		Answers:     req.Answers,
		AnonToken:   req.AnonymousToken,
		UseExisting: req.UseExisting,
		OtherName:   req.OtherName,
		UserID:      userID,
	}

	out, err := matchlink.SubmitMatchAssessment(r.Context(), h.Repo, h.Assessments, h.ProfileLoader, h.Bonds, input)
	if err != nil {
		if errors.Is(err, domain.ErrMatchLinkNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "Match link not found or deleted")
			return
		}
		if errors.Is(err, domain.ErrNoProfile) {
			respondError(w, http.StatusNotFound, "not_found", "No existing profile found — complete an assessment first")
			return
		}
		if errors.Is(err, domain.ErrAnswersRequired) {
			respondError(w, http.StatusBadRequest, "invalid_request", "answers is required")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}

	// Best-effort notification to link creator.
	if h.EmailSender != nil && h.UserRepo != nil {
		otherName := req.OtherName
		if otherName == "" && userID != nil {
			if u, lookupErr := h.UserRepo.FindByID(r.Context(), *userID); lookupErr == nil {
				otherName = u.Name
			}
		}
		if otherName == "" {
			otherName = "Anonymous"
		}
		notifyBondCreated(r.Context(), h.EmailSender, h.UserRepo, out.CreatorUserID, otherName, r.Header.Get("X-Locale"))
	}

	resp := matchlink.SubmitMatchAPIResponse{
		Profile:      out.Profile,
		AssessmentID: out.AssessmentID,
		Bond:         out.Bond,
	}
	respondJSON(w, http.StatusCreated, resp)
}

// notifyBondCreated sends a bond notification email to the link creator (best-effort).
func notifyBondCreated(ctx context.Context, emailSender user.EmailSender, userRepo user.UserRepository, creatorID int64, otherName, locale string) {
	if emailSender == nil || userRepo == nil {
		return
	}
	creator, err := userRepo.FindByID(ctx, creatorID)
	if err != nil {
		return
	}
	email := creator.Email
	if email == "" && creator.PendingEmail != nil {
		email = *creator.PendingEmail
	}
	if email == "" {
		return
	}
	if locale == "" {
		locale = "en"
	}
	if err := emailSender.SendBondNotification(ctx, email, otherName, creator.Name, locale); err != nil {
		log.Printf("[email] bond notification failed for user %d: %v", creatorID, err)
	}
}
