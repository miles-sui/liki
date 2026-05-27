package minglihttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/25types/25types/internal/httputil"
	"github.com/25types/25types/internal/mingli/huangli"
)

// HuangliHandler serves huangli (黄历) endpoints — date-only, public, no auth.
type HuangliHandler struct{}

// parseDateParams extracts year/month/day from query parameters. Returns error on invalid input.
func parseDateParams(r *http.Request) (int, int, int, error) {
	ys, ms, ds := r.URL.Query().Get("year"), r.URL.Query().Get("month"), r.URL.Query().Get("day")

	year, err := strconv.Atoi(ys)
	if err != nil {
		return 0, 0, 0, err
	}
	month, err := strconv.Atoi(ms)
	if err != nil {
		return 0, 0, 0, err
	}
	day, err := strconv.Atoi(ds)
	if err != nil {
		return 0, 0, 0, err
	}
	return year, month, day, nil
}

// GET /api/huangli/query?date=2026-05-25  or  ?month=2026-05  or  ?month=2026-05&event=wedding
func (h *HuangliHandler) Query(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	monthStr := r.URL.Query().Get("month")
	eventType := r.URL.Query().Get("event")

	if dateStr != "" {
		entry, err := huangli.QueryDate(dateStr, eventType)
		if err != nil {
			httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "invalid date format, use YYYY-MM-DD")
			return
		}
		httputil.RespondJSON(w, http.StatusOK, entry)
		return
	}

	if monthStr != "" {
		entries, err := huangli.QueryMonth(monthStr, eventType)
		if err != nil {
			httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "invalid month format, use YYYY-MM")
			return
		}
		httputil.RespondJSON(w, http.StatusOK, queryMonthResponse{YearMonth: monthStr, Days: entries})
		return
	}

	httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "date or month parameter is required")
}

type queryMonthResponse struct {
	YearMonth string             `json:"year_month"`
	Days      []huangli.DayEntry `json:"days"`
}

// POST /api/huangli/bond
func (h *HuangliHandler) Bond(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Birth     birthParams `json:"birth_info"`
		Month     string      `json:"month"`
		Date      string      `json:"date"`
		EventType string      `json:"event_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}
	if err := req.Birth.Validate(); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_birth_info", err.Error())
		return
	}

	chart := req.Birth.computeChart()

	if req.Date != "" {
		entry, err := huangli.CrossDate(chart.DayMaster, chart.Day.Branch, req.Date, req.EventType)
		if err != nil {
			httputil.RespondError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		httputil.RespondJSON(w, http.StatusOK, entry)
		return
	}

	if req.Month != "" {
		entries, err := huangli.CrossMonth(chart.DayMaster, chart.Day.Branch, req.Month, req.EventType)
		if err != nil {
			httputil.RespondError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		httputil.RespondJSON(w, http.StatusOK, bondMonthResponse{YearMonth: req.Month, Days: entries})
		return
	}

	httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "month or date is required")
}

type bondMonthResponse struct {
	YearMonth string               `json:"year_month"`
	Days      []huangli.BondDayEntry `json:"days"`
}

type jieQiResponse struct {
	JieQiDepth huangli.JieQiDepth     `json:"jieqi_depth"`
	RenYuan    huangli.RenYuanSiLing  `json:"ren_yuan"`
}

// GET /api/huangli/jieqi?year=2026&month=5&day=25
func (h *HuangliHandler) JieQi(w http.ResponseWriter, r *http.Request) {
	year, month, day, err := parseDateParams(r)
	if err != nil || year < 1900 || month < 1 || month > 12 || day < 1 || day > 31 {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "year/month/day required")
		return
	}

	jqDepth := huangli.ComputeJieQiDepth(year, month, day)
	renYuan := huangli.ComputeRenYuanSiLingForDate(year, month, day)

	httputil.RespondJSON(w, http.StatusOK, jieQiResponse{
		JieQiDepth: jqDepth,
		RenYuan:    renYuan,
	})
}


