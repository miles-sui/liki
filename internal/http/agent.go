package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"liki/internal/agent"
	
	"liki/internal/llm"

	"liki/internal/session"
)

// chatHandler returns an SSE handler for POST /api/agent/chat.
func chatHandler(chat *agent.ChatAgent, orders agent.OrderCreator, store *session.Store) http.HandlerFunc {
	type chatRequest struct {
		SessionID string `json:"session_id"`
		Message   string `json:"message"`
		Country   string `json:"country,omitempty"`
		City      string `json:"city,omitempty"`
		Lang      string `json:"lang,omitempty"` // frontend language: zh/hk/en
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "bad_request", "invalid request body")
			return
		}
		if req.Message == "" {
			respondError(w, http.StatusBadRequest, "bad_request", "message is required")
			return
		}

		// Load or create session.
		var sess *session.Session
		if req.SessionID != "" {
			var ok bool
			sess, ok = store.Get(req.SessionID)
			if !ok {
				respondError(w, http.StatusNotFound, "not_found", "session not found or expired")
				return
			}
			if sess.IsClosed() {
				respondError(w, http.StatusBadRequest, "session_closed", "session is closed")
				return
			}
			store.Touch(sess.ID)
		} else {
			sess = store.NewSession()
			if sess == nil {
				respondError(w, http.StatusServiceUnavailable, "server_error", "service is busy, please try again")
				return
			}
			if req.Country != "" {
				loc := req.City
				if loc == "" {
					loc = req.Country
				}
				loc = sanitizeLocation(loc)
				if loc != "" {
					sess.AppendMessage(llm.Message{Role: llm.RoleSystem, Content: "[" + loc + "] 用户IP位于此地区。收集出生信息时可据此建议默认城市和时区，但必须经用户明确确认。"})
				}
			}
		}

		sess.AppendMessage(llm.Message{Role: llm.RoleUser, Content: req.Message})

		// Set SSE headers.
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Session-ID", sess.ID)
		w.Header().Set("X-Accel-Buffering", "no")

		flusher, ok := w.(http.Flusher)
		if !ok {
			respondError(w, http.StatusInternalServerError, "server_error", "streaming not supported")
			return
		}

		ctx := r.Context()
		locale := langToLocale(req.Lang)

		chatCtx, chatCancel := context.WithCancel(ctx)
		defer chatCancel()
		result, err := chat.Chat(chatCtx, locale, sess.SnapshotMessages(), func(ev agent.ChatEvent) {
			if err := writeSSE(w, flusher, ev); err != nil {
				chatCancel()
			}
		}, orders, chat.Amounts)
		if err != nil {
			if ctx.Err() != nil {
				slog.Info("chat: client disconnected", "err", err)
			} else {
				slog.Error("chat: pipeline error", "err", err)
				if err := writeSSE(w, flusher, agent.ChatEvent{Type: agent.EventError, Content: "服务暂时不可用，请稍后重试"}); err != nil {
					slog.Warn("chat: write error event failed", "err", err)
				}
			}
			return
		}

		sess.SetMessages(result.Messages)

		if result.Purchase == nil {
			flushSSE(w, flusher)
			return
		}

		if err := writeSSE(w, flusher, agent.ChatEvent{
			Type: agent.EventDone,
			Data: map[string]any{
				"order_id": result.Purchase.OrderID,
				"amount":   result.Purchase.Amount,
				"product":  result.Purchase.Product,
			},
		}); err != nil {
			slog.Error("write SSE done event", "err", err)
		}

		sess.SetPhase(session.PhaseClosed)
	}
}

// greetingHandler serves the cached LLM-generated greeting (GET /api/agent/greeting).
func greetingHandler(chat *agent.ChatAgent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]any{"greeting": chat.Greeting})
	}
}

// sessionRestoreHandler returns session history (GET /api/agent/session).
func sessionRestoreHandler(store *session.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid := r.URL.Query().Get("session_id")
		if sid == "" {
			respondError(w, http.StatusBadRequest, "bad_request", "session_id is required")
			return
		}
		sess, ok := store.Get(sid)
		if !ok {
			respondError(w, http.StatusNotFound, "not_found", "session not found or expired")
			return
		}
		store.Touch(sid)
		_, phase, msgs := sess.Snapshot()
		respondJSON(w, http.StatusOK, map[string]any{
			"messages": msgs,
			"phase":    phase,
		})
	}
}

// sanitizeLocation strips characters that could be used for prompt injection.
func sanitizeLocation(s string) string {
	for _, r := range s {
		if r > 127 {
			continue // allow CJK
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == ' ' || r == '-' {
			continue
		}
		return ""
	}
	if len(s) > 64 {
		return ""
	}
	return s
}

// langToLocale maps frontend language code to BCP 47 locale.
func langToLocale(lang string) string {
	switch lang {
	case "zh":
		return "zh-Hans"
	case "hk":
		return "zh-Hant"
	case "en":
		return "en"
	default:
		return "zh-Hans"
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event agent.ChatEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "data: %s\n\n", data)
	if err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

// flushSSE sends an SSE comment to ensure framing is complete before handler returns.
// Prevents data loss when the handler returns immediately after streaming.
func flushSSE(w http.ResponseWriter, flusher http.Flusher) {
	fmt.Fprint(w, ": ok\n\n")
	flusher.Flush()
}
