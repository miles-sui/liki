package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/25types/25types/internal/app/application/reports"
	"github.com/25types/25types/internal/app/domain"
)

// ReportsHandler handles report generation, history, detail, deletion, and sharing.
type ReportsHandler struct {
	svc *reports.Service
}

// NewReportsHandler creates a ReportsHandler.
func NewReportsHandler(svc *reports.Service) *ReportsHandler {
	return &ReportsHandler{svc: svc}
}

// Create generates a report with SSE streaming.
// POST /api/reports
func (h *ReportsHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req domain.CreateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	chunks, err := h.svc.Generate(r.Context(), req, userID)
	if err != nil {
		code := "internal"
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrInvalidScene) {
			code, status = "invalid_request", http.StatusBadRequest
		} else if errors.Is(err, domain.ErrEngineDataReq) {
			code, status = "invalid_request", http.StatusBadRequest
		}
		respondError(w, status, code, err.Error())
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		respondError(w, http.StatusInternalServerError, "internal", "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	for c := range chunks {
		if c.Error != nil {
			data, _ := json.Marshal(Envelope{
				Error: &APIError{Code: "internal", Message: c.Error.Error()},
			})
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", data)
			flusher.Flush()
			return
		}
		if c.Done {
			reportID, _ := strconv.ParseInt(c.Text, 10, 64)
			fmt.Fprintf(w, "event: done\ndata: {\"report_id\":%d}\n\n", reportID)
			flusher.Flush()
			return
		}
		data, _ := json.Marshal(Envelope{
			Data: map[string]string{"text": c.Text},
		})
		fmt.Fprintf(w, "event: chunk\ndata: %s\n\n", data)
		flusher.Flush()
	}
}

// List returns the user's report history.
// GET /api/reports?scene=mingli&limit=20&offset=0
func (h *ReportsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	scene := r.URL.Query().Get("scene")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	result, err := h.svc.ListHistory(r.Context(), userID, scene, limit, offset)
	if err != nil {
		log.Printf("reports list err: %v", err)
		respondError(w, http.StatusInternalServerError, "internal", "failed to list reports")
		return
	}
	respondList(w, result.Items, result.Total)
}

// Get returns a single report by ID.
// GET /api/reports/{id}
func (h *ReportsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "invalid report id")
		return
	}

	report, err := h.svc.GetDetail(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrReportNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "report not found")
			return
		}
		log.Printf("reports get err: %v", err)
		respondError(w, http.StatusInternalServerError, "internal", "failed to get report")
		return
	}
	respondJSON(w, http.StatusOK, report)
}

// Delete soft-deletes a report.
// DELETE /api/reports/{id}
func (h *ReportsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "invalid report id")
		return
	}

	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, domain.ErrReportNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "report not found")
			return
		}
		log.Printf("reports delete err: %v", err)
		respondError(w, http.StatusInternalServerError, "internal", "failed to delete report")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// CreateShare creates a public sharing link for a report.
// POST /api/reports/{id}/share
func (h *ReportsHandler) CreateShare(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "invalid report id")
		return
	}

	share, err := h.svc.CreateShare(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrReportNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "report not found")
			return
		}
		log.Printf("reports share err: %v", err)
		respondError(w, http.StatusInternalServerError, "internal", "failed to create share")
		return
	}
	respondJSON(w, http.StatusOK, share)
}

// GetShared returns a publicly shared report by token.
// GET /api/reports/shared/{token}
func (h *ReportsHandler) GetShared(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "token is required")
		return
	}

	report, err := h.svc.GetShared(r.Context(), token)
	if err != nil {
		if errors.Is(err, domain.ErrReportNotFound) {
			respondError(w, http.StatusNotFound, "not_found", "report not found or link expired")
			return
		}
		log.Printf("reports shared err: %v", err)
		respondError(w, http.StatusInternalServerError, "internal", "failed to get shared report")
		return
	}
	// Public endpoint: return content only, no engine_data.
	respondJSON(w, http.StatusOK, map[string]any{
		"id":         report.ID,
		"scene":      report.Scene,
		"sub_scene":  report.SubScene,
		"question":   report.Question,
		"content":    report.Content,
		"locale":     report.Locale,
		"created_at": report.CreatedAt,
	})
}
