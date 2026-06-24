package handler
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var btZiwei = `"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4}`

func TestComputeZiweiChart_Valid(t *testing.T) {
	body := `{` + btZiwei + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env struct {
		Data struct {
			MingGong int `json:"ming_gong"`
			ShenGong int `json:"shen_gong"`
			JuShu    int `json:"ju_shu"`
			ZiweiPos int `json:"ziwei_pos"`
			Palaces  []struct {
				Name string `json:"name"`
				Gan  string `json:"gan"`
				Zhi  string `json:"zhi"`
			} `json:"palaces"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(env.Data.Palaces) != 12 {
		t.Errorf("palaces len = %d, want 12", len(env.Data.Palaces))
	}
}

func TestComputeZiweiChart_InvalidJSON(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{bad`))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestComputeZiweiChart_MissingBirth(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"gender":"male"}`))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestComputeZiweiChart_InvalidGender(t *testing.T) {
	body := `{` + btZiwei + `,"gender":"other"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestComputeZiweiDaxian_Valid(t *testing.T) {
	// First get a chart from computeZiweiChart
	chartBody := `{` + btZiwei + `,"gender":"male"}`
	cr := httptest.NewRequest("POST", "/", strings.NewReader(chartBody))
	cw := httptest.NewRecorder()
	computeZiweiChart(cw, cr)
	if cw.Code != http.StatusOK {
		t.Fatalf("setup chart: status = %d", cw.Code)
	}
	var chartEnv struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(cw.Body).Decode(&chartEnv); err != nil {
		t.Fatal(err)
	}

	daxianBody := `{"chart":` + string(chartEnv.Data) + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(daxianBody))
	w := httptest.NewRecorder()
	computeZiweiDaxian(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestComputeZiweiLiunian_Valid(t *testing.T) {
	chartBody := `{` + btZiwei + `,"gender":"male"}`
	cr := httptest.NewRequest("POST", "/", strings.NewReader(chartBody))
	cw := httptest.NewRecorder()
	computeZiweiChart(cw, cr)
	if cw.Code != http.StatusOK {
		t.Fatalf("setup chart: status = %d", cw.Code)
	}
	var chartEnv struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(cw.Body).Decode(&chartEnv); err != nil {
		t.Fatal(err)
	}

	lnBody := `{"liu_year":2025,"chart":` + string(chartEnv.Data) + `}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(lnBody))
	w := httptest.NewRecorder()
	computeZiweiLiunian(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestComputeZiweiLiuyue_Valid(t *testing.T) {
	chartBody := `{` + btZiwei + `,"gender":"male"}`
	cr := httptest.NewRequest("POST", "/", strings.NewReader(chartBody))
	cw := httptest.NewRecorder()
	computeZiweiChart(cw, cr)
	if cw.Code != http.StatusOK {
		t.Fatalf("setup chart: status = %d", cw.Code)
	}
	var chartEnv struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(cw.Body).Decode(&chartEnv); err != nil {
		t.Fatal(err)
	}

	lyBody := `{"liu_year":2025,"lunar_month":1,"chart":` + string(chartEnv.Data) + `}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(lyBody))
	w := httptest.NewRecorder()
	computeZiweiLiuyue(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestComputeZiweiLiuri_Valid(t *testing.T) {
	chartBody := `{` + btZiwei + `,"gender":"male"}`
	cr := httptest.NewRequest("POST", "/", strings.NewReader(chartBody))
	cw := httptest.NewRecorder()
	computeZiweiChart(cw, cr)
	if cw.Code != http.StatusOK {
		t.Fatalf("setup chart: status = %d", cw.Code)
	}
	var chartEnv struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(cw.Body).Decode(&chartEnv); err != nil {
		t.Fatal(err)
	}

	lrBody := `{"liu_year":2025,"lunar_month":1,"lunar_day":1,"chart":` + string(chartEnv.Data) + `}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(lrBody))
	w := httptest.NewRecorder()
	computeZiweiLiuri(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestComputeZiweiLiuri_InvalidLunarMonth(t *testing.T) {
	body := `{"liu_year":2025,"lunar_month":13,"lunar_day":1,"chart":{}}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiLiuri(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestComputeZiweiBond_Valid(t *testing.T) {
	chartBody := `{` + btZiwei + `,"gender":"male"}`
	cr := httptest.NewRequest("POST", "/", strings.NewReader(chartBody))
	cw := httptest.NewRecorder()
	computeZiweiChart(cw, cr)
	if cw.Code != http.StatusOK {
		t.Fatalf("setup chart A: status = %d", cw.Code)
	}
	var chartEnvA struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(cw.Body).Decode(&chartEnvA); err != nil {
		t.Fatal(err)
	}

	chartBodyB := `{` + btZiwei + `,"gender":"female"}`
	crB := httptest.NewRequest("POST", "/", strings.NewReader(chartBodyB))
	cwB := httptest.NewRecorder()
	computeZiweiChart(cwB, crB)
	if cwB.Code != http.StatusOK {
		t.Fatalf("setup chart B: status = %d", cwB.Code)
	}
	var chartEnvB struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(cwB.Body).Decode(&chartEnvB); err != nil {
		t.Fatal(err)
	}

	bondBody := `{"a":` + string(chartEnvA.Data) + `,"b":` + string(chartEnvB.Data) + `}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(bondBody))
	w := httptest.NewRecorder()
	computeZiweiBond(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestComputeZiweiBond_MissingB(t *testing.T) {
	chartBody := `{` + btZiwei + `,"gender":"male"}`
	cr := httptest.NewRequest("POST", "/", strings.NewReader(chartBody))
	cw := httptest.NewRecorder()
	computeZiweiChart(cw, cr)
	var chartEnv struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(cw.Body).Decode(&chartEnv); err != nil {
		t.Fatal(err)
	}

	bondBody := `{"a":` + string(chartEnv.Data) + `}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(bondBody))
	w := httptest.NewRecorder()
	computeZiweiBond(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422 for missing b", w.Code)
	}
}

func TestEdge_ZiWei_DaXian_EmptyChart(t *testing.T) {
	body := `{"chart":{},"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiDaxian(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("empty chart daxian: status=%d, want 422", w.Code)
	}
}

func TestEdge_ZiWei_LiuNian_EmptyChart(t *testing.T) {
	body := `{"liu_year":2025,"chart":{}}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiLiunian(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("empty chart liunian: status=%d, want 422", w.Code)
	}
}

func TestEdge2_ZiWeiBond_CorruptedChartJSON(t *testing.T) {
	// chart 字段给一个不是有效 Chart 的东西
	body := `{"a":{"palaces":[],"ming_gong":-99},"b":{"palaces":[],"ming_gong":-99}}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiBond(w, r)
	if w.Code >= 500 {
		t.Errorf("corrupted charts caused 5xx: %d", w.Code)
	}
}

func TestEdge2_ZiWei_LiuNian_WithBirthInChart(t *testing.T) {
	// chart 中嵌入的 birth 数据可能影响结果
	chartBody := `{` + bt15 + `,"gender":"male"}`
	cr := httptest.NewRequest("POST", "/", strings.NewReader(chartBody))
	cw := httptest.NewRecorder()
	computeZiweiChart(cw, cr)
	if cw.Code != http.StatusOK {
		t.Fatalf("setup chart: status=%d", cw.Code)
	}
	var chartEnv struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(cw.Body).Decode(&chartEnv); err != nil {
		t.Fatal(err)
	}

	lnBody := `{"liu_year":2025,"chart":` + string(chartEnv.Data) + `}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(lnBody))
	w := httptest.NewRecorder()
	computeZiweiLiunian(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("ziwei liunian: status=%d", w.Code)
	}
}

func TestEdge2_ZiWei_LiuYue_MonthEdge(t *testing.T) {
	chartBody := `{` + bt15 + `,"gender":"male"}`
	cr := httptest.NewRequest("POST", "/", strings.NewReader(chartBody))
	cw := httptest.NewRecorder()
	computeZiweiChart(cw, cr)
	if cw.Code != http.StatusOK {
		t.Fatalf("setup chart: status=%d", cw.Code)
	}
	var chartEnv struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(cw.Body).Decode(&chartEnv); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		month int
	}{
		{"month 0", 0},
		{"month 13", 13},
		{"month -1", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{
				"liu_year":    2025,
				"lunar_month": tt.month,
				"chart":       chartEnv.Data,
			})
			if err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest("POST", "/", strings.NewReader(string(body)))
			w := httptest.NewRecorder()
			computeZiweiLiuyue(w, r)
			if w.Code >= 500 {
				t.Errorf("month=%d caused 5xx: %d", tt.month, w.Code)
			}
		})
	}
}

func TestBug_ZiweiBond_MissingB_SilentlyPasses(t *testing.T) {
	// First get a valid chart for A
	chartBody := `{` + bt15 + `,"gender":"male"}`
	cr := httptest.NewRequest("POST", "/", strings.NewReader(chartBody))
	cw := httptest.NewRecorder()
	computeZiweiChart(cw, cr)
	if cw.Code != http.StatusOK {
		t.Fatalf("setup chart: status=%d", cw.Code)
	}
	var chartEnv struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(cw.Body).Decode(&chartEnv); err != nil {
		t.Fatal(err)
	}

	// Send bond request with only 'a', no 'b'
	bondBody := `{"a":` + string(chartEnv.Data) + `}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(bondBody))
	w := httptest.NewRecorder()
	computeZiweiBond(w, r)

	// BUG: decodeJSON doesn't validate, so missing 'b' gives zero-value Chart.
	// The handler returns 200 with a result computed against an empty chart.
	if w.Code == http.StatusOK {
		t.Error("BUG: ziwei bond with missing 'b' returns 200 (no validation)")
	}
	if w.Code == http.StatusUnprocessableEntity {
		t.Log("OK: missing 'b' returns 422")
	}
}

func TestBug_ZiweiBond_EmptyCharts_ReturnsResult(t *testing.T) {
	// Both charts are empty objects — now rejected by ju_shu validation.
	body := `{"a":{},"b":{}}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiBond(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("empty charts: status=%d, want 422", w.Code)
	}
}

func TestBug_ZiweiDaxian_EmptyChart(t *testing.T) {
	body := `{"chart":{},"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiDaxian(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("empty chart daxian: status=%d, want 422", w.Code)
	}
}

func TestBug_ZiweiLiunian_EmptyChart(t *testing.T) {
	body := `{"liu_year":2025,"chart":{}}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiLiunian(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("empty chart liunian: status=%d, want 422", w.Code)
	}
}
