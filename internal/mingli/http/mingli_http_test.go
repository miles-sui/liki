package minglihttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/25types/25types/internal/ganzhi"
)

func TestValidateBirthInfo(t *testing.T) {
	tests := []struct {
		name    string
		bp      birthParams
		wantErr bool
	}{
		{"valid", birthParams{Year: 2000, Month: 6, Day: 15, Hour: 12, Minute: 30, Longitude: 120, Timezone: 8, Gender: "male"}, false},
		{"valid female", birthParams{Year: 1995, Month: 1, Day: 1, Hour: 0, Minute: 0, Longitude: -75, Timezone: -5, Gender: "female"}, false},
		{"year too low", birthParams{Year: 1800, Month: 1, Day: 1, Hour: 0, Minute: 0}, true},
		{"year too high", birthParams{Year: 2300, Month: 1, Day: 1, Hour: 0, Minute: 0}, true},
		{"month too low", birthParams{Year: 2000, Month: 0, Day: 1, Hour: 0, Minute: 0}, true},
		{"month too high", birthParams{Year: 2000, Month: 13, Day: 1, Hour: 0, Minute: 0}, true},
		{"day too low", birthParams{Year: 2000, Month: 1, Day: 0, Hour: 0, Minute: 0}, true},
		{"day too high", birthParams{Year: 2000, Month: 1, Day: 32, Hour: 0, Minute: 0}, true},
		{"hour too low", birthParams{Year: 2000, Month: 1, Day: 1, Hour: -1, Minute: 0}, true},
		{"hour too high", birthParams{Year: 2000, Month: 1, Day: 1, Hour: 24, Minute: 0}, true},
		{"minute too low", birthParams{Year: 2000, Month: 1, Day: 1, Hour: 0, Minute: -1}, true},
		{"minute too high", birthParams{Year: 2000, Month: 1, Day: 1, Hour: 0, Minute: 60}, true},
		{"bad longitude", birthParams{Year: 2000, Month: 1, Day: 1, Hour: 0, Minute: 0, Longitude: 200}, true},
		{"bad longitude negative", birthParams{Year: 2000, Month: 1, Day: 1, Hour: 0, Minute: 0, Longitude: -200}, true},
		{"invalid gender", birthParams{Year: 2000, Month: 1, Day: 1, Hour: 0, Minute: 0, Gender: "other"}, true},
		{"empty gender fails", birthParams{Year: 2000, Month: 1, Day: 1, Hour: 0, Minute: 0, Gender: ""}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.bp.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("birthParams.Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var envelope struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	json.NewDecoder(rec.Body).Decode(&envelope)
	if envelope.Data.Status != "ok" {
		t.Errorf("status = %q, want %q", envelope.Data.Status, "ok")
	}
}

func TestComputeDailySuggestion_NoDayMaster(t *testing.T) {
	ds := ComputeDailySuggestion(0)
	if ds.Date == "" {
		t.Error("Date is empty")
	}
	if ds.DayPillar == "" {
		t.Error("DayPillar is empty")
	}
	if ds.Personalized {
		t.Error("Personalized should be false when dayMaster is zero")
	}
}

func TestComputeDailySuggestion_WithDayMaster(t *testing.T) {
	// Test with a valid day master (Jia Wood, index 1)
	ds := ComputeDailySuggestion(ganzhi.Stem(1))
	if !ds.Personalized {
		t.Error("Personalized should be true with a valid day master")
	}
	if ds.Suggestion == "" {
		t.Error("Suggestion is empty")
	}
	if ds.Question == "" {
		t.Error("Question is empty")
	}
}

func TestDayMasterFromBirthInfo(t *testing.T) {
	dm := DayMasterFromBirthInfo(2000, 6, 15, 12, 30, 120, 8)
	if dm < 1 || dm > 10 {
		t.Errorf("DayMaster = %d, want 1-10", dm)
	}
}

func TestBaZiChartHandler_InvalidJSON(t *testing.T) {
	h := &MingliHandler{}
	req := httptest.NewRequest(http.MethodPost, "/api/bazi/chart", strings.NewReader("not json"))
	rec := httptest.NewRecorder()
	h.ComputeChart(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestBaZiChartHandler_InvalidBirth(t *testing.T) {
	h := &MingliHandler{}
	body := `{"year": 1800, "month": 1, "day": 1, "hour": 0, "minute": 0}`
	req := httptest.NewRequest(http.MethodPost, "/api/bazi/chart", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ComputeChart(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestBaZiChartHandler_Valid(t *testing.T) {
	h := &MingliHandler{}
	body := `{"year": 2000, "month": 6, "day": 15, "hour": 12, "minute": 30, "longitude": 120, "timezone": 8, "gender": "male"}`
	req := httptest.NewRequest(http.MethodPost, "/api/bazi/chart", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ComputeChart(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("response missing data envelope")
	}

	// Check essential fields
	required := []string{"year_pillar", "month_pillar", "day_pillar", "hour_pillar", "day_master", "element_count"}
	for _, field := range required {
		if _, ok := data[field]; !ok {
			t.Errorf("missing field %q", field)
		}
	}

	// Day master should be a stem name string like "甲"
	dm, ok := data["day_master"].(string)
	if !ok || dm == "" {
		t.Errorf("day_master = %v, want non-empty string", data["day_master"])
	}
}

func TestBaZiBondHandler_MissingCharts(t *testing.T) {
	h := &MingliHandler{}
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/bazi/bond", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.BondCharts(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestBaZiBondHandler_Valid(t *testing.T) {
	h := &MingliHandler{}
	birth := `{"year": 2000, "month": 6, "day": 15, "hour": 12, "minute": 30, "longitude": 120, "timezone": 8, "gender": "male"}`
	body := `{"a": ` + birth + `, "b": ` + birth + `}`
	req := httptest.NewRequest(http.MethodPost, "/api/bazi/bond", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.BondCharts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("missing data envelope")
	}
	for _, field := range []string{"chart_a", "chart_b", "bond"} {
		if _, ok := data[field]; !ok {
			t.Errorf("missing field %q", field)
		}
	}
}

func TestSolarTermsHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/solar-terms", nil)
	rec := httptest.NewRecorder()
	SolarTerms(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("missing data envelope")
	}
	if _, ok := data["year"]; !ok {
		t.Error("missing year")
	}
	if _, ok := data["current"]; !ok {
		t.Error("missing current month")
	}
	months, ok := data["months"].([]interface{})
	if !ok || len(months) != 12 {
		t.Errorf("months count = %d, want 12", len(months))
	}
}

func TestReferenceStemsHandler(t *testing.T) {
	h := &ReferenceHandler{}
	req := httptest.NewRequest(http.MethodGet, "/api/reference/stems", nil)
	rec := httptest.NewRecorder()
	h.Stems(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	items := data["items"].([]interface{})
	if len(items) != 10 {
		t.Errorf("stems count = %d, want 10", len(items))
	}
}

func TestReferenceBranchesHandler(t *testing.T) {
	h := &ReferenceHandler{}
	req := httptest.NewRequest(http.MethodGet, "/api/reference/branches", nil)
	rec := httptest.NewRecorder()
	h.Branches(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	items := data["items"].([]interface{})
	if len(items) != 12 {
		t.Errorf("branches count = %d, want 12", len(items))
	}
}

func TestReferenceNayinHandler(t *testing.T) {
	h := &ReferenceHandler{}
	req := httptest.NewRequest(http.MethodGet, "/api/reference/nayin", nil)
	rec := httptest.NewRecorder()
	h.Nayin(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	items := data["items"].([]interface{})
	if len(items) != 60 {
		t.Errorf("nayin count = %d, want 60", len(items))
	}
}

func TestLiuYue_InvalidMonth(t *testing.T) {
	h := &MingliHandler{}
	body := `{"bazi": {"year":{"stem":8,"branch":10},"month":{"stem":3,"branch":9},"day":{"stem":2,"branch":12},"hour":{"stem":4,"branch":2}}, "year": 2024, "month": 13}`
	req := httptest.NewRequest(http.MethodPost, "/api/bazi/liuyue", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.LiuYue(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLiuYue_Valid(t *testing.T) {
	h := &MingliHandler{}
	body := `{"bazi": {"year":{"stem":8,"branch":10},"month":{"stem":3,"branch":9},"day":{"stem":2,"branch":12},"hour":{"stem":4,"branch":2}}, "year": 2024, "month": 5}`
	req := httptest.NewRequest(http.MethodPost, "/api/bazi/liuyue", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.LiuYue(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestLiuShi_InvalidHour(t *testing.T) {
	h := &MingliHandler{}
	body := `{"bazi": {"year":{"stem":8,"branch":10},"month":{"stem":3,"branch":9},"day":{"stem":2,"branch":12},"hour":{"stem":4,"branch":2}}, "date": "2024-05-20", "hour": 25}`
	req := httptest.NewRequest(http.MethodPost, "/api/bazi/liushi", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.LiuShi(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLiuNian_InvalidYear(t *testing.T) {
	h := &MingliHandler{}
	body := `{"bazi": {"year":{"stem":8,"branch":10},"month":{"stem":3,"branch":9},"day":{"stem":2,"branch":12},"hour":{"stem":4,"branch":2}}, "year": 2500}`
	req := httptest.NewRequest(http.MethodPost, "/api/bazi/liunian", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.LiuNian(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestXiaoYun_Defaults(t *testing.T) {
	h := &MingliHandler{}
	body := `{"birth": {"year": 2000, "month": 6, "day": 15, "hour": 12, "minute": 30, "longitude": 120, "timezone": 8, "gender": "male"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/bazi/xiao-yun", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.XiaoYun(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestXiaoXian_Defaults(t *testing.T) {
	h := &MingliHandler{}
		body := `{"gender": "male", "count": 5}`
	req := httptest.NewRequest(http.MethodPost, "/api/bazi/xiao-xian", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.XiaoXian(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
