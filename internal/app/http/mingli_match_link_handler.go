package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/25types/25types/internal/app/application/minglilink"
	"github.com/25types/25types/internal/app/application/matchlink"
	"github.com/25types/25types/internal/app/domain"
)

// MingliMatchLinkHandler holds dependencies for BaZi match link HTTP endpoints.
type MingliMatchLinkHandler struct {
	Repo      matchlink.MatchLinkRepository
	MingliUsers minglilink.BirthInfoLookup
}

// POST /api/mingli-match-links
func (h *MingliMatchLinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())
	out, err := matchlink.CreateMatchLink(r.Context(), h.Repo, uid, "mingli")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}
	respondJSON(w, http.StatusCreated, out)
}

// GET /api/mingli-match-links
func (h *MingliMatchLinkHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())
	items, err := matchlink.ListMatchLinks(r.Context(), h.Repo, uid, "mingli")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}
	respondList(w, items, len(items))
}

// DELETE /api/mingli-match-links/{id}
func (h *MingliMatchLinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

// GET /api/ml/{token}
func (h *MingliMatchLinkHandler) GetLinkInfo(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	info, err := minglilink.GetMingliMatchLink(r.Context(), h.Repo, h.MingliUsers, token)
	if err != nil {
		if errors.Is(err, domain.ErrMatchLinkNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "Match link not found or deleted")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}
	respondJSON(w, http.StatusOK, info)
}

// POST /api/ml/{token}
func (h *MingliMatchLinkHandler) Submit(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	var req struct {
		UseExisting bool              `json:"use_existing"`
		BirthInfo   *domain.BirthInfo `json:"birth_info"`
		OtherName   string            `json:"other_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

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

		if !req.UseExisting && req.BirthInfo != nil {
			bi := req.BirthInfo
			if err := validateBirthInfo(birthParams{
				Year: bi.Year, Month: bi.Month, Day: bi.Day,
				Hour: bi.Hour, Minute: bi.Minute,
				Longitude: bi.Longitude, Timezone: bi.Timezone,
				IsDST: bi.IsDST, Gender: string(bi.Gender),
			}); err != nil {
				respondError(w, http.StatusBadRequest, "invalid_birth_info", err.Error())
				return
			}
		}
	if !req.UseExisting && req.BirthInfo == nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Either use_existing or birth_info is required")
		return
	}

	input := minglilink.SubmitMingliMatchInput{
		Token:       token,
		UseExisting: req.UseExisting,
		BirthInfo:   req.BirthInfo,
		OtherName:   req.OtherName,
		UserID:      userID,
	}

	out, err := minglilink.SubmitMingliMatch(r.Context(), h.Repo, h.MingliUsers, input)
	if err != nil {
		if errors.Is(err, domain.ErrMatchLinkNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "Match link not found or deleted")
			return
		}
		if errors.Is(err, domain.ErrAnswersRequired) {
			respondError(w, http.StatusBadRequest, "invalid_request", "Either use_existing or birth_info is required")
			return
		}
		log.Printf("[mingli-match-link] submit error: %v", err)
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}

	respondJSON(w, http.StatusCreated, out)
}
