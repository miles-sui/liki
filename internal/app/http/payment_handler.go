package http

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/25types/25types/internal/app/application/commerce"
)

// SubscriptionPlan is a publicly visible pricing plan.
type SubscriptionPlan struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	NameEn   string   `json:"name_en"`
	Amount   float64  `json:"amount"`
	Interval string   `json:"interval"`
	Features []string `json:"features"`
}

// PaymentHandler handles donation checkout, subscription, and webhook endpoints.
type PaymentHandler struct {
	DodoClient     commerce.PaymentProvider
	DonationRepo   commerce.DonationRepository
	ThankYouSender commerce.ThankYouSender
	ProductID      string
	SubProductID   string
	PlansData      []SubscriptionPlan
	UserEmailFn    UserEmailFn
}

// UserEmailFn looks up a user email by ID. Returns (email, found).
type UserEmailFn func(ctx context.Context, userID int64) (string, bool)

// Plans handles GET /api/payments/plans (public).
func (h *PaymentHandler) Plans(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, struct {
		SingleReportPrice float64              `json:"single_report_price"`
		Currency          string               `json:"currency"`
		Plans             []SubscriptionPlan   `json:"plans"`
	}{SingleReportPrice: 9.9, Currency: "CNY", Plans: h.PlansData})
}

// Subscribe handles POST /api/payments/subscribe (auth required).
func (h *PaymentHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	if h.DodoClient == nil || h.SubProductID == "" {
		respondError(w, http.StatusServiceUnavailable, "service_unavailable", "Payment service not configured")
		return
	}

	var req struct {
		PlanID string `json:"plan_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	var plan *SubscriptionPlan
	for _, p := range h.PlansData {
		if p.ID == req.PlanID {
			plan = &p
			break
		}
	}
	if plan == nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "plan_id is missing or invalid")
		return
	}

	userEmail := ""
	if h.UserEmailFn != nil {
		if email, ok := h.UserEmailFn(r.Context(), uid); ok {
			userEmail = email
		}
	}

	locale := r.Header.Get("X-Locale")
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	returnURL := scheme + "://" + r.Host + "/" + locale + "/app"

	// Subscription checkout via Dodo — the Dodo product is configured as recurring.
	amount := int(plan.Amount * 100) // convert to cents for Dodo
	result, err := h.DodoClient.CreateCheckout(r.Context(), h.SubProductID, amount, uid, userEmail, returnURL)
	if err != nil {
		log.Printf("[payments] subscribe checkout failed for user %d plan %s: %v", uid, req.PlanID, err)
		respondError(w, http.StatusInternalServerError, "internal", "Failed to create checkout session")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"url": result.CheckoutURL})
}

// Checkout handles POST /api/payments/checkout (auth required).
func (h *PaymentHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	if h.DodoClient == nil {
		respondError(w, http.StatusServiceUnavailable, "service_unavailable", "Payment service not configured")
		return
	}

	var req struct {
		Amount int `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if !commerce.ValidateAmount(req.Amount) {
		respondError(w, http.StatusBadRequest, "invalid_request", "Amount is required and must be 990, 1990, or 2990")
		return
	}

	userEmail := ""
	if h.UserEmailFn != nil {
		if email, ok := h.UserEmailFn(r.Context(), uid); ok {
			userEmail = email
		}
	}

	locale := r.Header.Get("X-Locale")
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	returnURL := scheme + "://" + r.Host + "/" + locale + "/donate"

	output, err := commerce.CreateDonationCheckout(r.Context(), h.DodoClient, h.ProductID, commerce.CreateCheckoutInput{
		UserID:    uid,
		UserEmail: userEmail,
		Amount:    req.Amount,
		Locale:    locale,
		ReturnURL: returnURL,
	})
	if err != nil {
		log.Printf("[payments] checkout creation failed for user %d: %v", uid, err)
		respondError(w, http.StatusInternalServerError, "internal", "Failed to create checkout session")
		return
	}

	respondJSON(w, http.StatusOK, output)
}

// Webhook handles POST /api/payments/webhook (public — no auth).
func (h *PaymentHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Failed to read request body")
		return
	}
	defer r.Body.Close()

	event, err := h.DodoClient.VerifyWebhook(rawBody, r.Header)
	if err != nil {
		log.Printf("[payments] webhook verification failed: %v", err)
		respondError(w, http.StatusUnauthorized, "unauthorized", "Invalid webhook signature")
		return
	}

	locale := r.Header.Get("X-Locale")
	if err := commerce.HandleDonationWebhook(r.Context(), h.DonationRepo, h.ThankYouSender, event, locale); err != nil {
		log.Printf("[payments] webhook processing failed: %v", err)
		respondError(w, http.StatusInternalServerError, "internal", "Failed to process webhook")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Confirm handles POST /api/payments/confirm (auth required).
// Verifies the payment with Dodo's API and creates a donation record if confirmed.
// Payment ID (UNIQUE) ensures idempotency across confirm and webhook paths.
func (h *PaymentHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	if h.DodoClient == nil {
		respondError(w, http.StatusServiceUnavailable, "service_unavailable", "Payment service not configured")
		return
	}

	var req struct {
		PaymentID string `json:"payment_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PaymentID == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "payment_id is required")
		return
	}

	result, err := commerce.ConfirmDonation(r.Context(), h.DodoClient, h.DonationRepo, commerce.ConfirmDonationInput{
		UserID:    uid,
		PaymentID: req.PaymentID,
	})
	if err != nil {
		log.Printf("[payments] confirm failed for user %d payment %s: %v", uid, req.PaymentID, err)
		respondError(w, http.StatusInternalServerError, "internal", "Failed to confirm payment")
		return
	}

	respondJSON(w, http.StatusOK, result)
}
