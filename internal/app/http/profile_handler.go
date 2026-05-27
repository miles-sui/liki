package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/25types/25types/internal/app/application/profile"
	"github.com/25types/25types/internal/app/domain"
)

// ProfileHandler holds dependencies for profile HTTP endpoints.
type ProfileHandler struct {
	PageRepo profile.ProfilePageRepo
	Users    profile.UserLookup
	Profiles domain.ProfileLoader
	Bonds    profile.BondStore
}

// GET /api/profiles/{name}
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		respondError(w, http.StatusNotFound, "not_found", "Profile not found")
		return
	}

	var viewerID *int64
	if uid, ok := UserID(r.Context()); ok {
		viewerID = &uid
	}

	out, err := profile.GetProfile(r.Context(), h.PageRepo, name, viewerID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "Profile not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}

	respondJSON(w, http.StatusOK, out)
}

// GET /api/profiles/{name}/bonds
func (h *ProfileHandler) GetBonds(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	viewerID, ok := UserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	// Only the owner can see their bonds.
	u, err := h.Users.FindByName(r.Context(), name)
	if err != nil || u.ID != viewerID {
		respondError(w, http.StatusNotFound, "not_found", "Profile not found")
		return
	}

	events, err := profile.GetBonds(r.Context(), h.Bonds, viewerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}

	respondList(w, events, len(events))
}

// POST /api/bond
func (h *ProfileHandler) ComputeBond(w http.ResponseWriter, r *http.Request) {
	initiatorID, ok := UserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	var req struct {
		WithUserID int64  `json:"with_user_id"`
		WithName   string `json:"with_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	otherID := req.WithUserID
	if otherID == 0 && req.WithName != "" {
		u, err := h.Users.FindByName(r.Context(), req.WithName)
		if err != nil {
			respondError(w, http.StatusNotFound, "not_found", "User not found")
			return
		}
		otherID = u.ID
	}
	if otherID == 0 {
		respondError(w, http.StatusBadRequest, "invalid_request", "with_user_id or with_name is required")
		return
	}

	if otherID == initiatorID {
		respondError(w, http.StatusBadRequest, "invalid_request", "Cannot compare with yourself")
		return
	}

	result, err := profile.ComputeAndStoreBond(r.Context(), h.Bonds, h.Profiles, initiatorID, otherID)
	if err != nil {
		if errors.Is(err, domain.ErrNoProfile) {
			respondError(w, http.StatusNotFound, "not_found", "Both users must have completed an assessment")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}

	resp := &profile.ComputeBondOutput{
		Self:     result.Bond.Self,
		Other:    result.Bond.Other,
		DeltaA:   result.Bond.DeltaA,
		DeltaB:   result.Bond.DeltaB,
		Concord: result.Bond.Concord,
	}
	if otherUser, lookupErr := h.Users.FindByID(r.Context(), otherID); lookupErr == nil {
		ou := &profile.BondOtherUser{Name: otherUser.Name}
		if result.ProfB != nil {
			ou.IdentityLabel = result.ProfB.Identity.Label
			ou.IdentityID = result.ProfB.Identity.ID
		}
		resp.OtherUser = ou
	}

	respondJSON(w, http.StatusOK, resp)
}
