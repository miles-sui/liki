package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/25types/25types/internal/app/application/user"
	"github.com/25types/25types/internal/app/domain"
)

// UserHandler holds dependencies for user-related HTTP handlers.
type UserHandler struct {
	Repo       user.UserRepository
	Hasher     user.PasswordHasher
	Claimer    user.Claimer
	Sender     user.EmailSender
	ExportRepo user.ExportRepository
}

// tokenFn is a function that creates a JWT token.
var tokenFn = func(userID int64, tokenVersion int, userName string) (string, error) {
	return CreateToken(JWTClaims{
		UserID:       userID,
		TokenVersion: tokenVersion,
		UserName:     userName,
	})
}

// POST /api/auth/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input user.RegisterUseCaseInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Name, email, and password are required")
		return
	}

	output, err := user.RegisterUseCase(r.Context(), h.Repo, h.Claimer, h.Hasher, h.Sender, tokenFn, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNameAndPasswordRequired):
			respondError(w, http.StatusBadRequest, "invalid_request", "Name, email, and password are required")
		case errors.Is(err, domain.ErrInvalidEmail):
			respondError(w, http.StatusBadRequest, "invalid_request", "Invalid email format")
		case errors.Is(err, domain.ErrPasswordTooShort):
			respondError(w, http.StatusBadRequest, "invalid_request", "Password must be at least 8 characters")
		case errors.Is(err, domain.ErrPasswordContainsName):
			respondError(w, http.StatusBadRequest, "invalid_request", "Password must not contain username")
		case errors.Is(err, domain.ErrUsernameReserved):
			respondError(w, http.StatusConflict, "conflict", "Username is reserved")
		case errors.Is(err, domain.ErrUsernameTaken):
			respondError(w, http.StatusConflict, "conflict", "Username already exists")
		case errors.Is(err, domain.ErrEmailAlreadyVerified):
			respondError(w, http.StatusConflict, "conflict", "Email already registered")
		default:
			respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		}
		return
	}

	respondJSON(w, http.StatusCreated, struct {
		Token string           `json:"token"`
		User  user.RegisteredUser `json:"user"`
	}{Token: output.Token, User: output.User})
}

// POST /api/auth/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Name and password are required")
		return
	}

	output, err := user.LoginUseCase(r.Context(), h.Repo, h.Hasher, tokenFn, req.Name, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNameAndPasswordRequired):
			respondError(w, http.StatusBadRequest, "invalid_request", "Name and password are required")
		case errors.Is(err, domain.ErrInvalidCredentials):
			respondError(w, http.StatusUnauthorized, "unauthorized", "Invalid username or password")
		default:
			respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		}
		return
	}

	respondJSON(w, http.StatusOK, struct {
		Token string           `json:"token"`
		User  user.RegisteredUser `json:"user"`
	}{Token: output.Token, User: output.User})
}

// POST /api/auth/logout
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())
	if err := user.LogoutUseCase(r.Context(), h.Repo, uid); err != nil {
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}
	respondStatus(w, http.StatusOK, "logged_out")
}

// PUT /api/auth/password
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "current_password and new_password are required")
		return
	}

	token, err := user.ChangePasswordUseCase(r.Context(), h.Repo, h.Hasher, tokenFn,
		uid, req.CurrentPassword, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPasswordTooShort):
			respondError(w, http.StatusBadRequest, "invalid_request", "Password must be at least 8 characters")
		case errors.Is(err, domain.ErrCurrentPasswordWrong):
			respondError(w, http.StatusUnauthorized, "incorrect_password", "Current password is incorrect")
		default:
			respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"token":  token,
		"status": "password_changed",
	})
}

// GET /api/auth/verify-email
func (h *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	tok := r.URL.Query().Get("token")
	if tok == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid or expired token")
		return
	}
	if err := h.Repo.VerifyEmailByToken(r.Context(), tok); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid or expired token")
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"email_verified": true})
}

// POST /api/auth/forgot-password
func (h *UserHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}

	locale := r.Header.Get("X-Locale")
	user.ForgotPasswordUseCase(r.Context(), h.Repo, h.Sender, req.Email, locale)
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /api/auth/reset-password
func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid or expired token")
		return
	}
	err := user.ResetPasswordUseCase(r.Context(), h.Repo, h.Hasher, req.Token, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPasswordTooShort):
			respondError(w, http.StatusBadRequest, "invalid_request", "Password must be at least 8 characters")
		case errors.Is(err, domain.ErrTokenExpired):
			respondError(w, http.StatusBadRequest, "invalid_request", "Invalid or expired token")
		default:
			respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		}
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "password_changed"})
}

// GET /api/users/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())
	u, err := user.GetMeUseCase(r.Context(), h.Repo, uid)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	respondJSON(w, http.StatusOK, u)
}

// POST /api/auth/resend-verification
func (h *UserHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())
	locale := r.Header.Get("X-Locale")
	email, err := user.ResendVerificationUseCase(r.Context(), h.Repo, h.Sender, uid, locale)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailAlreadyVerified):
			respondError(w, http.StatusConflict, "already_verified", "Email is already verified")
		case errors.Is(err, domain.ErrNoEmailToVerify):
			respondError(w, http.StatusBadRequest, "no_email", "No email to verify")
		default:
			respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		}
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"email": email, "status": "sent"})
}

// PUT /api/users/me
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())


	var req struct {
		Name      *string            `json:"name,omitempty"`
		Email     *string            `json:"email,omitempty"`
		IsPublic  *bool              `json:"is_public,omitempty"`
		BirthInfo *domain.BirthInfo  `json:"birth_info,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "At least one field is required")
		return
	}

	u, err := user.UpdateMeUseCase(r.Context(), h.Repo, h.Sender, user.UpdateMeInput{
		UserID:    uid,
		Name:      req.Name,
		Email:     req.Email,
		IsPublic:  req.IsPublic,
		BirthInfo: req.BirthInfo,
		Locale:    r.Header.Get("X-Locale"),
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUsernameReserved):
			respondError(w, http.StatusConflict, "conflict", "Username is reserved")
		case errors.Is(err, domain.ErrUsernameTaken):
			respondError(w, http.StatusConflict, "conflict", "Username already taken")
		case errors.Is(err, domain.ErrEmailTaken):
			respondError(w, http.StatusConflict, "conflict", "Email already in use")
		case errors.Is(err, domain.ErrNameEmpty):
			respondError(w, http.StatusBadRequest, "invalid_request", "Name cannot be empty")
		case errors.Is(err, domain.ErrNoFields):
			respondError(w, http.StatusBadRequest, "invalid_request", "At least one field is required")
		default:
			respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		}
		return
	}
	respondJSON(w, http.StatusOK, u)
}

// DELETE /api/users/me
func (h *UserHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())
	reactivateBy, err := user.DeactivateMeUseCase(r.Context(), h.Repo, uid)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{
		"status":        "deactivated",
		"reactivate_by": reactivateBy,
	})
}

type exportAssessmentItem struct {
	ID           int64           `json:"id"`
	Type         string          `json:"type"`
	IdentityID   string          `json:"identity_id"`
	Profile      json.RawMessage `json:"profile"`
	AnswersJSON  json.RawMessage `json:"answers_json"`
	CreatedAt    string          `json:"created_at"`
	ReviewLinkID *int64          `json:"review_link_id,omitempty"`
	ReviewerName string          `json:"reviewer_name"`
}

type exportReviewLinkItem struct {
	ID        int64  `json:"id"`
	Token     string `json:"token"`
	URL       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

// GET /api/users/me/export
func (h *UserHandler) ExportMe(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())

	u, err := user.GetMeUseCase(r.Context(), h.Repo, uid)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred")
		return
	}

	as, _ := h.ExportRepo.GetExportAssessments(r.Context(), uid)
	assessments := make([]exportAssessmentItem, 0, len(as))
	for _, a := range as {
		assessments = append(assessments, exportAssessmentItem{
			ID:           a.ID,
			Type:         a.Type,
			IdentityID:   a.IdentityID,
			Profile:      json.RawMessage(a.ProfileJSON),
			AnswersJSON:  json.RawMessage(a.AnswersJSON),
			CreatedAt:    a.CreatedAt,
			ReviewLinkID: a.ReviewLinkID,
			ReviewerName: a.ReviewerName,
		})
	}

	rl, _ := h.ExportRepo.GetExportReviewLinks(r.Context(), uid)
	links := make([]exportReviewLinkItem, 0, len(rl))
	for _, l := range rl {
		links = append(links, exportReviewLinkItem{
			ID:        l.ID,
			Token:     l.Token,
			URL:       "/r/" + l.Token,
			ExpiresAt: l.ExpiresAt,
			CreatedAt: l.CreatedAt,
		})
	}

	respondJSON(w, http.StatusOK, struct {
		User        *user.RegisteredUser   `json:"user"`
		ExportedAt  string                 `json:"exported_at"`
		Assessments []exportAssessmentItem `json:"assessments"`
		ReviewLinks []exportReviewLinkItem `json:"review_links"`
	}{User: u, ExportedAt: time.Now().UTC().Format(time.RFC3339), Assessments: assessments, ReviewLinks: links})
}
