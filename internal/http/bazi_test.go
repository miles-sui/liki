package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// -- computeChart --

func TestComputeChart_Valid(t *testing.T) {
	body := `{"year":1984,"month":2,"day":4,"hour":6,"minute":0,"longitude":120,"timezone":8,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeChart(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data == nil {
		t.Fatal("data is nil")
	}
}

func TestComputeChart_InvalidJSON(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{bad`))
	w := httptest.NewRecorder()
	computeChart(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestComputeChart_MissingYear(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"month":2,"day":4,"hour":6,"gender":"male"}`))
	w := httptest.NewRecorder()
	computeChart(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestComputeChart_InvalidDate(t *testing.T) {
	body := `{"year":2024,"month":2,"day":30,"hour":6,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeChart(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422 for invalid date", w.Code)
	}
}

func TestComputeChart_InvalidLongitude(t *testing.T) {
	body := `{"year":1984,"month":2,"day":4,"hour":6,"longitude":200,"timezone":8,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeChart(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422 for invalid longitude", w.Code)
	}
}

func TestComputeChart_InvalidGender(t *testing.T) {
	body := `{"year":1984,"month":2,"day":4,"hour":6,"gender":"other"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeChart(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422 for invalid gender", w.Code)
	}
}

// -- computeSolarTime --

func TestComputeSolarTime_Valid(t *testing.T) {
	body := `{"year":1984,"month":2,"day":4,"hour":6,"minute":0,"longitude":87,"timezone":8,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeSolarTime(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestComputeSolarTime_InvalidJSON(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{`))
	w := httptest.NewRecorder()
	computeSolarTime(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestComputeSolarTime_InvalidBirth(t *testing.T) {
	body := `{"year":1800,"month":0,"day":0,"hour":0,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeSolarTime(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

// -- bondCharts --

func TestBondCharts_Valid(t *testing.T) {
	body := `{
		"a":{"year":1984,"month":2,"day":4,"hour":6,"gender":"male"},
		"b":{"year":1985,"month":8,"day":15,"hour":12,"gender":"female"}
	}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	bondCharts(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestBondCharts_MissingB(t *testing.T) {
	body := `{"a":{"year":1984,"month":2,"day":4,"hour":6,"gender":"male"}}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	bondCharts(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBondCharts_InvalidA(t *testing.T) {
	body := `{
		"a":{"year":1984,"month":2,"day":4,"hour":6,"gender":"invalid"},
		"b":{"year":1985,"month":8,"day":15,"hour":12,"gender":"female"}
	}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	bondCharts(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

// -- liuNian --

const bzJSON = `"bazi":{"nian":{"gan":"甲","zhi":"子"},"yue":{"gan":"乙","zhi":"丑"},"ri":{"gan":"丙","zhi":"寅"},"shi":{"gan":"丁","zhi":"卯"}}`

func TestLiuNian_Valid(t *testing.T) {
	body := `{` + bzJSON + `,"year":2025}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	liuNian(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestLiuNian_WithDayun(t *testing.T) {
	body := `{` + bzJSON + `,"year":2025,"current_dayun":{"gan":5,"zhi":5}}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	liuNian(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestLiuNian_MissingBaziRi(t *testing.T) {
	body := `{"bazi":{"nian":{"gan":"甲","zhi":"子"}},"year":2025}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	liuNian(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestLiuNian_NegativeYear(t *testing.T) {
	body := `{` + bzJSON + `,"year":-1}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	liuNian(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

// -- liuYue --

func TestLiuYue_Valid(t *testing.T) {
	body := `{` + bzJSON + `,"year":2025,"month":6}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	liuYue(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestLiuYue_MissingMonth(t *testing.T) {
	body := `{"bazi":{"nian":{"gan":"甲","zhi":"子"},"ri":{"gan":"丙","zhi":"寅"},"shi":{"gan":"丁","zhi":"卯"}},"year":2025}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	liuYue(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

// -- liuRi --

func TestLiuRi_Valid(t *testing.T) {
	body := `{"bazi":{"nian":{"gan":"甲","zhi":"子"},"yue":{"gan":"乙","zhi":"丑"},"ri":{"gan":"丙","zhi":"寅"},"shi":{"gan":"丁","zhi":"卯"}},"date":"2025-06-15"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	liuRi(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestLiuRi_InvalidDateFormat(t *testing.T) {
	body := `{"bazi":{"nian":{"gan":"甲","zhi":"子"},"yue":{"gan":"乙","zhi":"丑"},"ri":{"gan":"丙","zhi":"寅"},"shi":{"gan":"丁","zhi":"卯"}},"date":"bad-date"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	liuRi(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestLiuRi_WithPillars(t *testing.T) {
	body := `{"bazi":{"nian":{"gan":"甲","zhi":"子"},"yue":{"gan":"乙","zhi":"丑"},"ri":{"gan":"丙","zhi":"寅"},"shi":{"gan":"丁","zhi":"卯"}},"date":"2025-06-15","dayun_pillar":{"gan":5,"zhi":5},"liunian_pillar":{"gan":6,"zhi":6}}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	liuRi(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

// -- liuShi --

func TestLiuShi_Valid(t *testing.T) {
	body := `{"bazi":{"nian":{"gan":"甲","zhi":"子"},"yue":{"gan":"乙","zhi":"丑"},"ri":{"gan":"丙","zhi":"寅"},"shi":{"gan":"丁","zhi":"卯"}},"date":"2025-06-15","hour":12}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	liuShi(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestLiuShi_MissingDate(t *testing.T) {
	body := `{"bazi":{"nianzhu":{"gan":1,"zhi":1},"rizhu":{"gan":3,"zhi":3},"shizhu":{"gan":4,"zhi":4}},"hour":12}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	liuShi(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestLiuShi_HourOutOfRange(t *testing.T) {
	body := `{"bazi":{"nian":{"gan":"甲","zhi":"子"},"yue":{"gan":"乙","zhi":"丑"},"ri":{"gan":"丙","zhi":"寅"},"shi":{"gan":"丁","zhi":"卯"}},"date":"2025-06-15","hour":25}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	liuShi(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

// -- xiaoYun --

func TestXiaoYun_Valid(t *testing.T) {
	body := `{"gender":"male","day_master":1,"count":10}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	xiaoYun(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestXiaoYun_MissingCount(t *testing.T) {
	body := `{"gender":"male","day_master":1}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	xiaoYun(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestXiaoYun_CountTooLarge(t *testing.T) {
	body := `{"gender":"male","day_master":1,"count":200}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	xiaoYun(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

// -- xiaoXian --

func TestXiaoXian_Valid(t *testing.T) {
	body := `{"gender":"female","count":5}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	xiaoXian(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestXiaoXian_MissingGender(t *testing.T) {
	body := `{"count":5}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	xiaoXian(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestXiaoXian_InvalidGender(t *testing.T) {
	body := `{"gender":"unknown","count":5}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	xiaoXian(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

// -- decodeJSON edge case --

func TestDecodeJSON_EmptyBody(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(""))
	w := httptest.NewRecorder()
	_, ok := decodeJSON[domainBirthParams](w, r)
	if ok {
		t.Error("expected ok=false for empty body")
	}
}

// domainBirthParams is a copy used to avoid import cycle in decodeJSON test.
type domainBirthParams struct {
	Year  int `json:"year"`
	Month int `json:"month"`
}

func (p domainBirthParams) Validate() error { return nil }

// -- respondJSON / respondError via computeChart wrapper --

func TestRespondError_WritesEnvelope(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{bad`))
	w := httptest.NewRecorder()
	computeChart(w, r)

	body, err := io.ReadAll(w.Result().Body)
	if err != nil { t.Fatal(err) }
	if !strings.Contains(string(body), "error") {
		t.Error("error response should contain 'error' key")
	}
}
