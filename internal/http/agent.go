package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"liki/internal/agent"
	"liki/internal/payment"
)

// namingHandler returns an SSE handler for POST /api/agent/naming.
// Requires JWT auth cookie. Loads chat history from DB,
// persists new messages, and detects report output.
func namingHandler(chat *agent.ChatAgent, store *payment.Store) http.HandlerFunc {
	type namingRequest struct {
		Message string `json:"message"`
		Lang    string `json:"lang,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		_, orderID, ok := jwtAuth(r)
		if !ok {
			respondError(w, http.StatusUnauthorized, "unauthorized", "请先登录或购买")
			return
		}

		o, err := store.GetOrder(r.Context(), orderID)
		if err != nil {
			respondError(w, http.StatusNotFound, "not_found", "订单不存在")
			return
		}

		if o.Status != payment.OrderPaid {
			respondError(w, http.StatusForbidden, "forbidden", "订单未支付")
			return
		}

		expiresAt, err := time.Parse(time.DateTime, o.ChatExpiresAt)
		if err != nil || time.Now().After(expiresAt) {
			respondError(w, http.StatusForbidden, "expired", "聊天已过期")
			return
		}

		var req namingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
			return
		}
		if req.Message == "" {
			respondError(w, http.StatusBadRequest, "invalid_request", "message is required")
			return
		}

		// Load chat history and build messages.
		messages, err := buildChatMessages(r.Context(), store, orderID, req.Message)
		if err != nil {
			slog.Error("naming: load history", "err", err)
			respondError(w, http.StatusInternalServerError, "internal_error", "加载历史记录失败")
			return
		}

		// Save user message — must succeed, or the LLM will lack context.
		if err := store.CreateChatMessage(r.Context(), orderID, payment.RoleUser, req.Message); err != nil {
			slog.Error("naming: save user message", "err", err)
			respondError(w, http.StatusInternalServerError, "internal_error", "保存消息失败")
			return
		}

		// SSE headers.
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		flusher, ok := w.(http.Flusher)
		if !ok {
			respondError(w, http.StatusInternalServerError, "internal_error", "streaming not supported")
			return
		}

		ctx := r.Context()
		locale := langToLocale(req.Lang)

		chatCtx, chatCancel := context.WithCancel(ctx)
		defer chatCancel()

		prevCount := len(messages) - 1 // messages already has the user message appended
		result, err := chat.NamingChat(chatCtx, locale, messages, func(ev agent.ChatEvent) {
			if err := writeSSE(w, flusher, ev); err != nil {
				chatCancel()
			}
		})
		if err != nil {
			if ctx.Err() != nil {
				slog.Info("naming: client disconnected", "err", err)
			} else {
				slog.Error("naming: pipeline error", "err", err)
				if err := writeSSE(w, flusher, agent.ChatEvent{Type: agent.EventError, Content: "服务暂时不可用，请稍后重试"}); err != nil {
					slog.Warn("naming: write error event failed", "err", err)
				}
			}
			return
		}

		// Detect and save report from the last assistant message.
		detectAndSaveReport(r.Context(), store, orderID, result, func(url string) {
			if err := writeSSE(w, flusher, agent.ChatEvent{
				Type:    agent.EventReportReady,
				Content: url,
			}); err != nil {
				slog.Warn("naming: write report-ready event", "err", err)
			}
		})

		// Persist new messages (skip system prompt + history + user message).
		persistChatResult(r.Context(), store, orderID, result, prevCount)

		flushSSE(w, flusher)
	}
}
