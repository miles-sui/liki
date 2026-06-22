package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/mail"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/i18n"
	"liki/internal/payment"
)

type checkoutRequest struct {
	OrderID string `json:"order_id"`
	Email   string `json:"email"`
}

func (r checkoutRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.OrderID, validation.Required),
		validation.Field(&r.Email, validation.By(validateEmail)),
	)
}

func validateEmail(value any) error {
	s, ok := value.(string)
	if !ok {
		return nil
	}
	if s == "" {
		return nil // email is optional
	}
	if _, err := mail.ParseAddress(s); err != nil {
		return errors.New("invalid email")
	}
	return nil
}

func handleCheckout(svc *payment.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := decodeAndValidate[checkoutRequest](w, r)
		if !ok {
			return
		}

		result, err := svc.CreateCheckout(r.Context(), req.OrderID, req.Email)
		if err != nil {
			if errors.Is(err, payment.ErrOrderNotFound) {
				respondError(w, http.StatusNotFound, "not_found", i18n.T(i18n.DetectLang(r), "err.order_not_found"))
				return
			}
			slog.Error("payment: checkout", "err", err)
			respondError(w, http.StatusInternalServerError, "checkout_error", "创建订单失败，请稍后重试")
			return
		}

		respondJSON(w, http.StatusOK, result)
	}
}

// handleOrderStatus returns the current status of an order (GET /api/orders/{id}/status).
// Used by the frontend to validate sessionStorage orderID before showing resume banner.
func handleOrderStatus(svc *payment.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.PathValue("id")
		if orderID == "" {
			respondError(w, http.StatusBadRequest, "invalid_path", "missing order id")
			return
		}
		status, product, err := svc.OrderStatus(r.Context(), orderID)
		if err != nil {
			respondError(w, http.StatusNotFound, "not_found", "订单不存在")
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": string(status), "product": string(product)})
	}
}

// handleRetryOrder retries report generation for orders that missed the webhook
// (POST /api/orders/{id}/retry). If the order is paid but has no llm_json,
// triggers background generation and returns the current status.
func handleRetryOrder(svc *payment.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.PathValue("id")
		if orderID == "" {
			respondError(w, http.StatusBadRequest, "invalid_path", "missing order id")
			return
		}
		status, product, llmJSON, err := svc.RetryReportGeneration(r.Context(), orderID)
		if err != nil {
			respondError(w, http.StatusNotFound, "not_found", i18n.T(i18n.DetectLang(r), "err.order_not_found"))
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{
			"status":   string(status),
			"product":  string(product),
			"llm_json": llmJSON,
		})
	}
}

// handlePaymentReturn handles Dodo post-payment redirect (GET /api/payments/return/{id}).
// Email is collected by AI in conversation and stored at order creation time.
func handlePaymentReturn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.PathValue("id")
		if orderID == "" || r.URL.Query().Get("status") != "succeeded" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		http.Redirect(w, r, "/report/"+orderID, http.StatusFound)
	}
}

func handleWebhook(svc *payment.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			slog.Error("payment: webhook read", "err", err)
			respondError(w, http.StatusInternalServerError, "read_error", "读取请求失败")
			return
		}

		if err := svc.HandleWebhook(r.Context(), body, r.Header); err != nil {
			if errors.Is(err, payment.ErrWebhookVerify) {
				slog.Warn("payment: webhook verify", "err", err)
				respondError(w, http.StatusBadRequest, "bad_webhook", i18n.T(i18n.DetectLang(r), "err.webhook_signature_invalid"))
				return
			}
			slog.Error("payment: webhook", "err", err)
			respondError(w, http.StatusInternalServerError, "webhook_error", "处理失败，请稍后重试")
			return
		}

		respondStatus(w, http.StatusOK, "ok")
	}
}

// redirectReport redirects /api/orders/{id}/report to /report/{id} for browser print.
func redirectReport() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.PathValue("id")
		if orderID == "" {
			respondError(w, http.StatusBadRequest, "invalid_path", i18n.T(i18n.DetectLang(r), "err.missing_order_id"))
			return
		}
		http.Redirect(w, r, "/report/"+orderID, http.StatusFound)
	}
}
