package http

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/i18n"
	"liki/internal/payment"
	"liki/internal/product"
)

type createOrderRequest struct {
	Email    string `json:"email"`
	Product  string `json:"product"`
	Currency string `json:"currency"`
}

type checkoutRequest struct {
	OrderID  string `json:"order_id"`
	Email    string `json:"email"`
	Provider string `json:"provider"`
}

func (r createOrderRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email, validation.Required, validation.By(validateEmail)),
		validation.Field(&r.Product, validation.Required, validation.In(string(product.ProductNaming))),
		validation.Field(&r.Currency, validation.Required),
	)
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
		return nil
	}
	if _, err := mail.ParseAddress(s); err != nil {
		return errors.New("invalid email")
	}
	return nil
}

// handleCreateOrder creates a pending naming order (POST /api/orders).
func handleCreateOrder(store *payment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := decodeAndValidate[createOrderRequest](w, r)
		if !ok {
			return
		}
		amount := product.NamingAmountCents(product.Currency(req.Currency))
		if amount == 0 {
			respondError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("unsupported currency: %s", req.Currency))
			return
		}

		orderID := product.NewOrderID()
		if err := store.CreateOrder(r.Context(), orderID, product.Product(req.Product), amount, req.Currency, req.Email, "", "", ""); err != nil {
			slog.Error("payment: create order", "err", err)
			respondError(w, http.StatusInternalServerError, "internal_error", "创建订单失败，请稍后重试")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{"order_id": orderID})
	}
}

func handleCheckout(svc *payment.Service, a *Analytics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := decodeAndValidate[checkoutRequest](w, r)
		if !ok {
			return
		}

		provider := req.Provider
		if provider == "" {
			country := strings.ToUpper(r.Header.Get("CF-IPCountry"))
			if country == "" || country == "CN" {
				provider = payment.ProviderXunhu
			} else {
				provider = payment.ProviderDodo
			}
		}

		result, err := svc.CreateCheckout(r.Context(), provider, req.OrderID, req.Email, ReturnToken(req.OrderID))
		if err != nil {
			if errors.Is(err, payment.ErrOrderNotFound) || errors.Is(err, payment.ErrUnknownProvider) {
				respondError(w, http.StatusNotFound, "not_found", i18n.T(i18n.DetectLang(r), "err.order_not_found"))
				return
			}
			slog.Error("payment: checkout", "err", err)
			respondError(w, http.StatusInternalServerError, "checkout_error", "创建订单失败，请稍后重试")
			return
		}

		a.RecordCheckout()
		respondJSON(w, http.StatusOK, result)
	}
}

// handleOrderStatus returns the current status of an order (GET /api/orders/{id}/status).
// Used by the frontend to load chat state (expiry, birth_info) and validate orderID.
func handleOrderStatus(store *payment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.PathValue("id")
		if orderID == "" {
			respondError(w, http.StatusBadRequest, "invalid_path", "missing order id")
			return
		}
		o, err := store.GetOrder(r.Context(), orderID)
		if err != nil {
			respondError(w, http.StatusNotFound, "not_found", "订单不存在")
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{
			"status":          string(o.Status),
			"product":         string(o.Product),
			"birth_info":      o.BirthInfo,
			"chat_expires_at": o.ChatExpiresAt,
			"email":           o.Email,
		})
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
		status, product, llmJSON, err := svc.GetOrderData(r.Context(), orderID)
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

// handlePaymentReturn handles post-payment redirect (GET /api/payments/return/{id}).
func handlePaymentReturn(store *payment.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.PathValue("id")
		token := r.URL.Query().Get("t")
		if orderID == "" || r.URL.Query().Get("status") != "succeeded" || !verifyReturnToken(orderID, token) {
			slog.Warn("payment: invalid return callback", "orderID", orderID, "has_token", token != "")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Mark order as paid via the callback — this ensures the user can
		// access chat immediately even if the webhook arrives later.
		newPayment, email, _, err := store.MarkPaidViaCallback(r.Context(), orderID)
		if err != nil {
			slog.Error("payment: return mark paid", "orderID", orderID, "err", err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		if email == "" {
			slog.Error("payment: return order has no email", "orderID", orderID)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		if newPayment {
			slog.Info("payment: order marked paid via callback", "orderID", orderID)
		}

		setJWTCookie(w, email, orderID)
		http.Redirect(w, r, "/chat?order_id="+orderID, http.StatusFound)
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

		respondSuccess(w, "success")
	}
}

func respondSuccess(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(body)); err != nil {
		slog.Warn("respondSuccess write", "err", err)
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
