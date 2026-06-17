package handler

import (
	"net/http"

	"liki/internal/payment"
)

// Detects paid+llm_json="" and triggers generation in background as fallback.
func handleReport(svc *payment.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.PathValue("id")
		if orderID == "" {
			respondInvalidRequest(w, "missing order id")
			return
		}

		report, err := svc.GetReport(r.Context(), orderID)
		if err != nil {
			respondError(w, http.StatusNotFound, "not_found", "订单不存在")
			return
		}

		// Fallback: paid but llm_json not yet generated (webhook race / failure).
		// StartReportGeneration is idempotent via in-memory mutex map.
		if report.Status == payment.OrderPaid && report.LlmJSON == "" {
			svc.StartReportGeneration(orderID, report.Product, report.ChartJSON)
		}

		respondJSON(w, http.StatusOK, report)
	}
}
