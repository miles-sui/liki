package http

import (
	"net/http"

	"liki/internal/payment"
)

func handleReport(svc *payment.Service, a *Analytics) http.HandlerFunc {
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

		a.RecordReportView()
		respondJSON(w, http.StatusOK, report)
	}
}
