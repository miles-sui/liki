package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"liki/internal/i18n"
	"liki/internal/payment"
)

type loginRequest struct {
	Email string `json:"email"`
}

type loginResponse struct {
	Redirect     string              `json:"redirect,omitempty"`
	OrderID      string              `json:"order_id,omitempty"`
	HasBirthInfo bool                `json:"has_birth_info,omitempty"`
	Orders       []loginOrderSummary `json:"orders,omitempty"`
}

type loginOrderSummary struct {
	OrderID       string `json:"order_id"`
	Summary       string `json:"summary"`
	ExpiresAt     string `json:"expires_at"`
	HasBirthInfo  bool   `json:"has_birth_info"`
}

func handleLogin(store *payment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_request", i18n.T(i18n.DetectLang(r), "err.invalid_json"))
			return
		}
		if req.Email == "" {
			respondError(w, http.StatusBadRequest, "invalid_request", i18n.T(i18n.DetectLang(r), "err.email_required"))
			return
		}

		orders, err := store.FindActiveOrdersByEmail(r.Context(), req.Email)
		if err != nil {
			slog.Error("auth: find orders by email", "err", err)
			respondError(w, http.StatusInternalServerError, "internal_error", "查询失败，请稍后重试")
			return
		}

		if len(orders) == 0 {
			respondError(w, http.StatusUnauthorized, "not_found", "未找到有效订单，请先购买")
			return
		}

		if len(orders) == 1 {
			o := orders[0]
			setJWTCookie(w, o.Email, o.OrderID)
			respondJSON(w, http.StatusOK, loginResponse{
				Redirect:     "/chat",
				OrderID:      o.OrderID,
				HasBirthInfo: o.BirthInfo != "",
			})
			return
		}

		// Multiple orders — let user pick.
		summaries := make([]loginOrderSummary, len(orders))
		for i, o := range orders {
			summaries[i] = loginOrderSummary{
				OrderID:       o.OrderID,
				Summary:       "起名",
				ExpiresAt:     o.ChatExpiresAt,
				HasBirthInfo:  o.BirthInfo != "",
			}
		}
		respondJSON(w, http.StatusOK, loginResponse{Orders: summaries})
	}
}

type selectOrderRequest struct {
	OrderID string `json:"order_id"`
	Email   string `json:"email"`
}

func handleOrderSelect(store *payment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req selectOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_request", i18n.T(i18n.DetectLang(r), "err.invalid_json"))
			return
		}
		if req.OrderID == "" || req.Email == "" {
			respondError(w, http.StatusBadRequest, "invalid_request", "order_id and email are required")
			return
		}

		o, err := store.GetOrder(r.Context(), req.OrderID)
		if err != nil {
			respondError(w, http.StatusNotFound, "not_found", "订单不存在")
			return
		}

		if o.Email != req.Email {
			respondError(w, http.StatusForbidden, "forbidden", "无权访问此订单")
			return
		}

		setJWTCookie(w, o.Email, o.OrderID)
		respondJSON(w, http.StatusOK, loginResponse{
			Redirect:     "/chat",
			OrderID:      o.OrderID,
			HasBirthInfo: o.BirthInfo != "",
		})
	}
}

// ValidateJWTSecret checks that JWT_SECRET is set. Call at startup.
func ValidateJWTSecret() error {
	if os.Getenv("JWT_SECRET") == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}
	return nil
}

// jwtSecret returns the JWT signing key from the environment.
func jwtSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

func setJWTCookie(w http.ResponseWriter, email, orderID string) {
	claims := jwt.MapClaims{
		"email":    email,
		"order_id": orderID,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(jwtSecret())
	if err != nil {
		slog.Error("auth: sign jwt", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "liki_token",
		Value:    signed,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})
}

// jwtAuth reads the JWT cookie and returns email and order_id.
func jwtAuth(r *http.Request) (email, orderID string, ok bool) {
	cookie, err := r.Cookie("liki_token")
	if err != nil {
		return "", "", false
	}
	token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (any, error) {
		return jwtSecret(), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return "", "", false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", false
	}
	email, _ = claims["email"].(string)
	orderID, _ = claims["order_id"].(string)
	return email, orderID, email != "" && orderID != ""
}
