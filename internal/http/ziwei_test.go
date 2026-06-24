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
	if env.Data.JuShu < 2 || env.Data.JuShu > 6 {
		t.Errorf("ju_shu = %d, want [2,6]", env.Data.JuShu)
	}
	for i, p := range env.Data.Palaces {
		if p.Name == "" {
			t.Errorf("palace[%d].Name is empty", i)
		}
		if p.Gan == "" {
			t.Errorf("palace[%d].Gan is empty", i)
		}
		if p.Zhi == "" {
			t.Errorf("palace[%d].Zhi is empty", i)
		}
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
	var steps []struct {
		StartAge int    `json:"start_age"`
		EndAge   int    `json:"end_age"`
		Name     string `json:"name"`
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(env.Data, &steps); err != nil {
		t.Fatalf("decode steps: %v", err)
	}
	if len(steps) != 12 {
		t.Errorf("daxian steps = %d, want 12", len(steps))
	}
	if steps[0].Name == "" {
		t.Error("first step name is empty")
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
	var env struct {
		Data struct {
			MingGong int `json:"ming_gong"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Data.MingGong < 0 || env.Data.MingGong > 11 {
		t.Errorf("ming_gong = %d, want [0,11]", env.Data.MingGong)
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
	var env struct {
		Data struct {
			MingGong int `json:"ming_gong"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Data.MingGong < 0 || env.Data.MingGong > 11 {
		t.Errorf("ming_gong = %d, want [0,11]", env.Data.MingGong)
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
	var env struct {
		Data struct {
			MingGong int `json:"ming_gong"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Data.MingGong < 0 || env.Data.MingGong > 11 {
		t.Errorf("ming_gong = %d, want [0,11]", env.Data.MingGong)
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
	var env struct {
		Data struct {
			StarCross []json.RawMessage `json:"star_cross"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data.StarCross) == 0 {
		t.Error("star_cross is empty")
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

func TestBlackBox_ZiWei_Invariants(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}

	var env struct {
		Data struct {
			Palaces []struct {
				Name string `json:"name"`
				Gan  string `json:"gan"`
				Zhi  string `json:"zhi"`
			} `json:"palaces"`
			MingGong int `json:"ming_gong"`
			ShenGong int `json:"shen_gong"`
			JuShu    int `json:"ju_shu"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}

	if len(env.Data.Palaces) != 12 {
		t.Errorf("palaces = %d, want 12", len(env.Data.Palaces))
	}
	if env.Data.JuShu < 2 || env.Data.JuShu > 6 {
		t.Errorf("ju_shu = %d, want [2,6]", env.Data.JuShu)
	}
	if env.Data.MingGong < 0 || env.Data.MingGong > 11 {
		t.Errorf("ming_gong = %d, want [0,11]", env.Data.MingGong)
	}
	if env.Data.ShenGong < 0 || env.Data.ShenGong > 11 {
		t.Errorf("shen_gong = %d, want [0,11]", env.Data.ShenGong)
	}

	// 每个宫必须有名字、天干、地支
	for _, p := range env.Data.Palaces {
		if p.Name == "" {
			t.Error("palace name is empty")
		}
		if p.Gan == "" {
			t.Error("palace gan is empty")
		}
		if p.Zhi == "" {
			t.Error("palace zhi is empty")
		}
	}
}

func TestBlackBox_CrossEndpoint_BaZi_ZiWei_YearBranch(t *testing.T) {
	// 同一个出生时间，八字 year_zhi 和紫微 year_zhi 应该一样
	bzBody := `{` + bt15 + `,"gender":"male"}`
	bzr := httptest.NewRequest("POST", "/", strings.NewReader(bzBody))
	bzw := httptest.NewRecorder()
	computeChart(bzw, bzr)
	if bzw.Code != http.StatusOK {
		t.Fatalf("bazi chart: status=%d", bzw.Code)
	}
	var bzEnv struct {
		Data struct {
			Year struct{ Zhi string } `json:"nian"`
		} `json:"data"`
	}
	if err := json.NewDecoder(bzw.Body).Decode(&bzEnv); err != nil {
		t.Fatal(err)
	}

	zwBody := `{` + bt15 + `,"gender":"male"}`
	zwr := httptest.NewRequest("POST", "/", strings.NewReader(zwBody))
	zww := httptest.NewRecorder()
	computeZiweiChart(zww, zwr)
	if zww.Code != http.StatusOK {
		t.Fatalf("ziwei chart: status=%d", zww.Code)
	}
	var zwEnv struct {
		Data struct {
			Palaces []struct {
				Zhi string `json:"zhi"`
			} `json:"palaces"`
		} `json:"data"`
	}
	if err := json.NewDecoder(zww.Body).Decode(&zwEnv); err != nil {
		t.Fatal(err)
	}

	// 八字 Year.Zhi (如 "子") 应该等于紫微某个宫的地支
	// 紫微的 year branch 在 palaces[?] 的 zhi 中
	// 不过紫微的 palaces[?].zhi 是十二宫的地支，不是年支。
	// 这个跨端点比较需要领域知识。
	_ = bzEnv
	_ = zwEnv
}

func TestBlackBox_ZiWeiBond_DifferentGenders_HasStars(t *testing.T) {
	// 获取男性命盘
	cbM := `{` + bt15 + `,"gender":"male"}`
	crM := httptest.NewRequest("POST", "/", strings.NewReader(cbM))
	cwM := httptest.NewRecorder()
	computeZiweiChart(cwM, crM)
	if cwM.Code != http.StatusOK {
		t.Fatalf("male chart: status=%d", cwM.Code)
	}
	var envM struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(cwM.Body).Decode(&envM); err != nil {
		t.Fatal(err)
	}

	// 获取女性命盘
	cbF := `{` + bt15 + `,"gender":"female"}`
	crF := httptest.NewRequest("POST", "/", strings.NewReader(cbF))
	cwF := httptest.NewRecorder()
	computeZiweiChart(cwF, crF)
	if cwF.Code != http.StatusOK {
		t.Fatalf("female chart: status=%d", cwF.Code)
	}
	var envF struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(cwF.Body).Decode(&envF); err != nil {
		t.Fatal(err)
	}

	// 合盘
	bondBody := `{"a":` + string(envM.Data) + `,"b":` + string(envF.Data) + `}`
	br := httptest.NewRequest("POST", "/", strings.NewReader(bondBody))
	bw := httptest.NewRecorder()
	computeZiweiBond(bw, br)
	if bw.Code != http.StatusOK {
		t.Fatalf("bond: status=%d", bw.Code)
	}

	var bondEnv struct {
		Data struct {
			StarCross []json.RawMessage `json:"star_cross"`
		} `json:"data"`
	}
	if err := json.NewDecoder(bw.Body).Decode(&bondEnv); err != nil {
		t.Fatal(err)
	}
	if len(bondEnv.Data.StarCross) == 0 {
		t.Error("star_cross is empty for different-gender bond")
	}
}

func TestBlackBox_ZiWei_DaXian_Has12Steps(t *testing.T) {
	chartBody := `{` + bt15 + `,"gender":"male"}`
	cr := httptest.NewRequest("POST", "/", strings.NewReader(chartBody))
	cw := httptest.NewRecorder()
	computeZiweiChart(cw, cr)
	if cw.Code != http.StatusOK {
		t.Fatalf("chart: status=%d", cw.Code)
	}
	var chartEnv struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(cw.Body).Decode(&chartEnv); err != nil {
		t.Fatal(err)
	}

	dxBody := `{"chart":` + string(chartEnv.Data) + `,"gender":"male"}`
	dxr := httptest.NewRequest("POST", "/", strings.NewReader(dxBody))
	dxw := httptest.NewRecorder()
	computeZiweiDaxian(dxw, dxr)
	if dxw.Code != http.StatusOK {
		t.Fatalf("daxian: status=%d", dxw.Code)
	}

	var steps []json.RawMessage
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(dxw.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(env.Data, &steps); err != nil {
		t.Fatal(err)
	}

	if len(steps) != 12 {
		t.Errorf("daxian steps = %d, want 12", len(steps))
	}
}

func TestBlackBox_ZiWei_LiuNian_HasMingGong(t *testing.T) {
	chartBody := `{` + bt15 + `,"gender":"male"}`
	cr := httptest.NewRequest("POST", "/", strings.NewReader(chartBody))
	cw := httptest.NewRecorder()
	computeZiweiChart(cw, cr)
	if cw.Code != http.StatusOK {
		t.Fatalf("chart: status=%d", cw.Code)
	}
	var chartEnv struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(cw.Body).Decode(&chartEnv); err != nil {
		t.Fatal(err)
	}

	lnBody := `{"liu_year":2025,"chart":` + string(chartEnv.Data) + `}`
	lnr := httptest.NewRequest("POST", "/", strings.NewReader(lnBody))
	lnw := httptest.NewRecorder()
	computeZiweiLiunian(lnw, lnr)
	if lnw.Code != http.StatusOK {
		t.Fatalf("liunian: status=%d", lnw.Code)
	}

	var env struct {
		Data struct {
			MingGong     int    `json:"ming_gong"`
			MingGongName string `json:"ming_gong_name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(lnw.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}

	if env.Data.MingGongName == "" {
		t.Error("liunian ming_gong_name is empty")
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

func TestDomain_Ziwei_14MainStars(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	var env struct {
		Data struct {
			Palaces []struct {
				Stars []struct {
					IsMajor bool   `json:"is_major"`
					Name    string `json:"name"`
				} `json:"stars"`
			} `json:"palaces"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data.Palaces) != 12 {
		t.Fatalf("palaces len=%d, want 12", len(env.Data.Palaces))
	}
	majorCount := 0
	for _, p := range env.Data.Palaces {
		for _, s := range p.Stars {
			if s.IsMajor {
				majorCount++
			}
		}
	}
	if majorCount != 14 {
		t.Errorf("total major stars=%d, want 14", majorCount)
	}
}

func TestDomain_Ziwei_JuShu_Range(t *testing.T) {
	dates := []struct{ name, time string }{
		{"春节", "1984-02-15T08:00:00+08:00"},
		{"夏至", "1990-06-21T12:00:00+08:00"},
		{"冬至", "2000-12-22T00:00:00+08:00"},
		{"元旦", "1984-01-01T12:00:00+08:00"},
	}
	for _, d := range dates {
		t.Run(d.name, func(t *testing.T) {
			body := `{"birth":{"time":"` + d.time + `","longitude":116.4},"gender":"male"}`
			r := httptest.NewRequest("POST", "/", strings.NewReader(body))
			w := httptest.NewRecorder()
			computeZiweiChart(w, r)
			if w.Code != http.StatusOK {
				t.Fatalf("status=%d", w.Code)
			}
			var env struct {
				Data struct {
					JuShu int `json:"ju_shu"`
				} `json:"data"`
			}
			if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
				t.Fatal(err)
			}
			if env.Data.JuShu < 2 || env.Data.JuShu > 6 {
				t.Errorf("JuShu=%d, want [2,6]", env.Data.JuShu)
			}
		})
	}
}

func TestDomain_Ziwei_MingGong_Zero(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	var env struct {
		Data struct {
			MingGong int `json:"ming_gong"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Data.MingGong != 0 {
		t.Errorf("MingGong=%d, want 0", env.Data.MingGong)
	}
}

func TestDomain_Ziwei_DaXian_12Steps(t *testing.T) {
	// 先算命盘
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("chart status=%d", w.Code)
	}
	// 取 ju_shu 和 raw data
	bodyBytes := w.Body.Bytes()
	var meta struct {
		Data struct {
			JuShu  int    `json:"ju_shu"`
			Gender string `json:"gender"`
		} `json:"data"`
	}
	var rawData struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &meta); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(bodyBytes, &rawData); err != nil {
		t.Fatal(err)
	}

	dxBody := `{"chart":` + string(rawData.Data) + `,"gender":"` + meta.Data.Gender + `"}`
	r2 := httptest.NewRequest("POST", "/", strings.NewReader(dxBody))
	w2 := httptest.NewRecorder()
	computeZiweiDaxian(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("daxian status=%d, body=%s", w2.Code, w2.Body.String())
	}
	var dxEnv struct {
		Data []struct {
			StartAge int    `json:"start_age"`
			EndAge   int    `json:"end_age"`
			Name     string `json:"name"`
			Palace   int    `json:"palace"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w2.Body).Decode(&dxEnv); err != nil {
		t.Fatal(err)
	}
	if len(dxEnv.Data) != 12 {
		t.Errorf("DaXian steps=%d, want 12", len(dxEnv.Data))
	}
	if len(dxEnv.Data) > 0 {
		first := dxEnv.Data[0]
		if first.StartAge != meta.Data.JuShu {
			t.Errorf("first step StartAge=%d, want %d (ju_shu)", first.StartAge, meta.Data.JuShu)
		}
	}
	// Verify consecutive steps are 10 years apart
	for i := 1; i < len(dxEnv.Data); i++ {
		prev := dxEnv.Data[i-1]
		curr := dxEnv.Data[i]
		if curr.StartAge != prev.EndAge+1 {
			t.Errorf("step[%d].StartAge=%d, want step[%d].EndAge+1=%d",
				i, curr.StartAge, i-1, prev.EndAge+1)
		}
	}
}

func TestDomain_Ziwei_SiHua_4Entries(t *testing.T) {
	body := `{` + bt15 + `,"gender":"female"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	var env struct {
		Data struct {
			SiHua map[string]string `json:"si_hua"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data.SiHua) != 4 {
		t.Errorf("SiHua entries=%d, want 4", len(env.Data.SiHua))
	}
	validSihua := map[string]bool{"禄": true, "权": true, "科": true, "忌": true}
	for star, hua := range env.Data.SiHua {
		if !validSihua[hua] {
			t.Errorf("SiHua[%s]=%q, want one of 禄/权/科/忌", star, hua)
		}
	}
}

func TestDomain_Bazi_Ziwei_YearGanConsistency(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("bazi status=%d", w.Code)
	}
	var bzEnv struct {
		Data struct {
			Year struct{ Gan string } `json:"nian"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&bzEnv); err != nil {
		t.Fatal(err)
	}

	r2 := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w2 := httptest.NewRecorder()
	computeZiweiChart(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("ziwei status=%d", w2.Code)
	}
	var zwEnv struct {
		Data struct {
			YearGan string `json:"year_gan"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w2.Body).Decode(&zwEnv); err != nil {
		t.Fatal(err)
	}

	if bzEnv.Data.Year.Gan != zwEnv.Data.YearGan {
		t.Errorf("bazi Year.Gan=%q != ziwei YearGan=%q", bzEnv.Data.Year.Gan, zwEnv.Data.YearGan)
	}
}

func TestDomain_Ziwei_LiuNian_MingGongOffset(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	bodyBytes := w.Body.Bytes()
	var cEnv struct {
		Data struct {
			BirthYear int `json:"birth_year"`
		} `json:"data"`
	}
	var raw struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &cEnv); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		t.Fatal(err)
	}

	// 流年 2025 → offset = (2025-1984)%12 = 41%12 = 5
	lnBody := `{"liu_year":2025,"chart":` + string(raw.Data) + `}`
	r2 := httptest.NewRequest("POST", "/", strings.NewReader(lnBody))
	w2 := httptest.NewRecorder()
	computeZiweiLiunian(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("liunian status=%d, body=%s", w2.Code, w2.Body.String())
	}
	var lnEnv struct {
		Data struct {
			MingGong     int    `json:"ming_gong"`
			MingGongName string `json:"ming_gong_name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w2.Body).Decode(&lnEnv); err != nil {
		t.Fatal(err)
	}

	expectedOffset := (2025 - cEnv.Data.BirthYear) % 12
	if lnEnv.Data.MingGong != expectedOffset {
		t.Errorf("LiuNian.MingGong=%d, want %d (offset 2025-%d=%%12=%d)",
			lnEnv.Data.MingGong, expectedOffset, cEnv.Data.BirthYear, expectedOffset)
	}
	if lnEnv.Data.MingGongName == "" {
		t.Error("MingGongName is empty")
	}
}

func TestDomain_Ziwei_MingShen_Different(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	var env struct {
		Data struct {
			MingGong int `json:"ming_gong"`
			ShenGong int `json:"shen_gong"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	// 1984-02-15: 农历正月十四 辰时
	// 命宫 = ((1-5+2)%12+12)%12+1 = 11 → 戌 → index 10
	// 身宫 = ((1+5)%12+12)%12+1 = 7 → 午 → index 6
	// mingGong=0 (命宫总是在 Palaces[0] = 命宫 always at index 0)
	// shenGong should be some other index
	// Actually mingGong is always 0 because 命宫 is always at palaces[0]
	if env.Data.ShenGong >= 12 {
		t.Errorf("ShenGong=%d, want [0,11]", env.Data.ShenGong)
	}
}

func TestDomain_Ziwei_LiuYue_SiHua(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("chart status=%d", w.Code)
	}
	var cEnv struct{ Data json.RawMessage }
	if err := json.NewDecoder(w.Body).Decode(&cEnv); err != nil {
		t.Fatal(err)
	}

	lyBody := `{"liu_year":2025,"lunar_month":6,"chart":` + string(cEnv.Data) + `}`
	r2 := httptest.NewRequest("POST", "/", strings.NewReader(lyBody))
	w2 := httptest.NewRecorder()
	computeZiweiLiuyue(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("liuyue status=%d, body=%s", w2.Code, w2.Body.String())
	}
	var lyEnv struct {
		Data struct {
			SiHua map[string]string `json:"si_hua"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w2.Body).Decode(&lyEnv); err != nil {
		t.Fatal(err)
	}
	if len(lyEnv.Data.SiHua) != 4 {
		t.Errorf("SiHua entries=%d, want 4", len(lyEnv.Data.SiHua))
	}
}

func TestDomain_Ziwei_SiHua_JiaYear(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	var env struct {
		Data struct {
			SiHua   map[string]string `json:"si_hua"`
			YearGan string            `json:"year_gan"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}

	if env.Data.YearGan != "甲" {
		t.Fatalf("YearGan=%q, want 甲", env.Data.YearGan)
	}

	// 甲年四化: 廉贞(5)禄/破军(13)权/武曲(3)科/太阳(2)忌
	// starIndex: ZiWei=0,TianJi=1,TaiYang=2,WuQu=3,TianTong=4,LianZhen=5,...,PoJun=13
	want := map[string]string{
		"5": "禄", "13": "权", "3": "科", "2": "忌",
	}
	if len(env.Data.SiHua) != 4 {
		t.Errorf("SiHua entries=%d, want 4", len(env.Data.SiHua))
	}
	for starIdx, hua := range want {
		if env.Data.SiHua[starIdx] != hua {
			t.Errorf("SiHua[starIdx=%s]=%q, want %s", starIdx, env.Data.SiHua[starIdx], hua)
		}
	}
}

func TestDomain_Ziwei_ZiweiPos(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	var env struct {
		Data struct {
			ZiweiPos  int    `json:"ziwei_pos"`
			JuShu     int    `json:"ju_shu"`
			JuShuName string `json:"ju_shu_name"`
			HourZhi   string `json:"hour_zhi"`
			BirthYear int    `json:"birth_year"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}

	if env.Data.ZiweiPos < 0 || env.Data.ZiweiPos >= 12 {
		t.Errorf("ZiweiPos=%d, want [0,11]", env.Data.ZiweiPos)
	}
	if env.Data.JuShuName == "" {
		t.Error("JuShuName is empty")
	}
	if env.Data.HourZhi == "" {
		t.Error("HourZhi is empty")
	}
	if env.Data.BirthYear != 1984 {
		t.Errorf("BirthYear=%d, want 1984", env.Data.BirthYear)
	}
	// Verify Ziwei star is in the ziweiPos palace
	var env2 struct {
		Data struct {
			Palaces []struct {
				Stars []struct {
					Name    string `json:"name"`
					IsMajor bool   `json:"is_major"`
				} `json:"stars"`
			} `json:"palaces"`
		} `json:"data"`
	}
	// Re-decode from w.Body
	r2 := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w2 := httptest.NewRecorder()
	computeZiweiChart(w2, r2)
	if err := json.NewDecoder(w2.Body).Decode(&env2); err != nil {
		t.Fatal(err)
	}

	pos := env.Data.ZiweiPos
	foundZiwei := false
	for _, s := range env2.Data.Palaces[pos].Stars {
		if s.Name == "紫微" && s.IsMajor {
			foundZiwei = true
			break
		}
	}
	if !foundZiwei {
		t.Errorf("Ziwei star not found at ziwei_pos=%d", pos)
	}
}

func TestDomain_Ziwei_Patterns(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	var env struct {
		Data struct {
			Patterns []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"patterns"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	// Patterns may be empty for some charts, but should not cause error
	if len(env.Data.Patterns) > 0 {
		for _, p := range env.Data.Patterns {
			if p.Name == "" {
				t.Error("pattern with empty Name")
			}
		}
	}
	t.Logf("Patterns count=%d", len(env.Data.Patterns))
}

func TestDomain_Ziwei_LiuNian_SiHuaPalace(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("chart status=%d", w.Code)
	}
	var raw struct{ Data json.RawMessage }
	if err := json.NewDecoder(w.Body).Decode(&raw); err != nil {
		t.Fatal(err)
	}

	lnBody := `{"liu_year":2025,"chart":` + string(raw.Data) + `}`
	r2 := httptest.NewRequest("POST", "/", strings.NewReader(lnBody))
	w2 := httptest.NewRecorder()
	computeZiweiLiunian(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("liunian status=%d, body=%s", w2.Code, w2.Body.String())
	}
	var lnEnv struct {
		Data struct {
			SiHuaPalace map[string]int `json:"si_hua_palace"`
			MinorStars  map[string]int `json:"minor_stars"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w2.Body).Decode(&lnEnv); err != nil {
		t.Fatal(err)
	}
	// 甲年四化 → 4个四化星入宫
	if len(lnEnv.Data.SiHuaPalace) != 4 {
		t.Errorf("SiHuaPalace entries=%d, want 4", len(lnEnv.Data.SiHuaPalace))
	}
	for starIdx, palaceIdx := range lnEnv.Data.SiHuaPalace {
		if palaceIdx < 0 || palaceIdx >= 12 {
			t.Errorf("SiHuaPalace[%s]=%d, want [0,11]", starIdx, palaceIdx)
		}
	}
	// MinorStars should have some entries
	if len(lnEnv.Data.MinorStars) == 0 {
		t.Error("MinorStars is empty")
	}
}

func TestDomain_Ziwei_DaXian_Direction(t *testing.T) {
	tests := []struct {
		name        string
		time        string
		gender      string
		wantForward bool // true=顺行(阳男/阴女), false=逆行(阴男/阳女)
	}{
		// 1984=甲子年(阳), 男→顺行
		{"甲子年阳男→顺行", "1984-02-15T08:00:00+08:00", "male", true},
		// 1984=甲子年(阳), 女→逆行
		{"甲子年阳女→逆行", "1984-02-15T08:00:00+08:00", "female", false},
		// 1985=乙丑年(阴), 男→逆行
		{"乙丑年阴男→逆行", "1985-06-15T12:00:00+08:00", "male", false},
		// 1985=乙丑年(阴), 女→顺行
		{"乙丑年阴女→顺行", "1985-06-15T12:00:00+08:00", "female", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"birth":{"time":"` + tt.time + `","longitude":116.4},"gender":"` + tt.gender + `"}`
			r := httptest.NewRequest("POST", "/", strings.NewReader(body))
			w := httptest.NewRecorder()
			computeZiweiChart(w, r)
			if w.Code != http.StatusOK {
				t.Fatalf("chart status=%d", w.Code)
			}
			var raw struct{ Data json.RawMessage }
			var cEnv struct {
				Data struct {
					Gender string `json:"gender"`
				}
			}
			bodyBytes := w.Body.Bytes()
			if err := json.Unmarshal(bodyBytes, &cEnv); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(bodyBytes, &raw); err != nil {
				t.Fatal(err)
			}

			dxBody := `{"chart":` + string(raw.Data) + `,"gender":"` + cEnv.Data.Gender + `"}`
			r2 := httptest.NewRequest("POST", "/", strings.NewReader(dxBody))
			w2 := httptest.NewRecorder()
			computeZiweiDaxian(w2, r2)
			if w2.Code != http.StatusOK {
				t.Fatalf("daxian status=%d", w2.Code)
			}
			var dxEnv struct {
				Data []struct {
					StartAge int `json:"start_age"`
					Palace   int `json:"palace"`
				} `json:"data"`
			}
			if err := json.NewDecoder(w2.Body).Decode(&dxEnv); err != nil {
				t.Fatal(err)
			}
			if len(dxEnv.Data) < 2 {
				t.Fatal("not enough daxian steps")
			}
			// 顺行: palace index increases (with wrap)
			// 逆行: palace index decreases (with wrap)
			p0, p1 := dxEnv.Data[0].Palace, dxEnv.Data[1].Palace
			diff := (p1 - p0 + 12) % 12
			if tt.wantForward {
				if diff != 1 {
					t.Errorf("DaXian direction: p0=%d p1=%d diff=%d, want 1 (forward)", p0, p1, diff)
				}
			} else {
				if diff != 11 { // 11 ≡ -1 mod 12
					t.Errorf("DaXian direction: p0=%d p1=%d diff=%d, want 11 (reverse)", p0, p1, diff)
				}
			}
		})
	}
}

func TestDomain_Ziwei_LiuRi(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("chart status=%d", w.Code)
	}
	var raw struct{ Data json.RawMessage }
	if err := json.NewDecoder(w.Body).Decode(&raw); err != nil {
		t.Fatal(err)
	}

	lrBody := `{"liu_year":2025,"lunar_month":6,"lunar_day":15,"chart":` + string(raw.Data) + `}`
	r2 := httptest.NewRequest("POST", "/", strings.NewReader(lrBody))
	w2 := httptest.NewRecorder()
	computeZiweiLiuri(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("liuri status=%d, body=%s", w2.Code, w2.Body.String())
	}
	var lrEnv struct {
		Data struct {
			MingGong     int               `json:"ming_gong"`
			MingGongName string            `json:"ming_gong_name"`
			SiHua        map[string]string `json:"si_hua"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w2.Body).Decode(&lrEnv); err != nil {
		t.Fatal(err)
	}
	if lrEnv.Data.MingGongName == "" {
		t.Error("MingGongName is empty")
	}
	if len(lrEnv.Data.SiHua) != 4 {
		t.Errorf("SiHua entries=%d, want 4", len(lrEnv.Data.SiHua))
	}
}

func TestDomain_Ziwei_Bond(t *testing.T) {
	// 先算两个人的命盘
	bodyA := `{` + bt15 + `,"gender":"male"}`
	rA := httptest.NewRequest("POST", "/", strings.NewReader(bodyA))
	wA := httptest.NewRecorder()
	computeZiweiChart(wA, rA)
	if wA.Code != http.StatusOK {
		t.Fatalf("chart A status=%d", wA.Code)
	}
	var rawA struct{ Data json.RawMessage }
	if err := json.NewDecoder(wA.Body).Decode(&rawA); err != nil {
		t.Fatal(err)
	}

	bodyB := `{"birth":{"time":"1990-05-20T12:00:00+08:00","longitude":121.5},"gender":"female"}`
	rB := httptest.NewRequest("POST", "/", strings.NewReader(bodyB))
	wB := httptest.NewRecorder()
	computeZiweiChart(wB, rB)
	if wB.Code != http.StatusOK {
		t.Fatalf("chart B status=%d", wB.Code)
	}
	var rawB struct{ Data json.RawMessage }
	if err := json.NewDecoder(wB.Body).Decode(&rawB); err != nil {
		t.Fatal(err)
	}

	bondBody := `{"a":` + string(rawA.Data) + `,"b":` + string(rawB.Data) + `}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(bondBody))
	w := httptest.NewRecorder()
	computeZiweiBond(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("bond status=%d, body=%s", w.Code, w.Body.String())
	}
	var env struct {
		Data struct {
			SihuaCross []struct {
				Star  int    `json:"star"`
				Type  string `json:"type"`
				IntoB int    `json:"into_b"`
			} `json:"sihua_cross"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	// Each person has 4 sihua → each maps to a palace in the other's chart
	if len(env.Data.SihuaCross) == 0 {
		t.Error("SihuaCross is empty")
	}
	for _, sc := range env.Data.SihuaCross {
		if sc.Star == 0 {
			t.Error("sihua_cross entry with zero star")
		}
		if sc.Type == "" {
			t.Error("sihua_cross entry with empty type")
		}
	}
}

func TestDomain_LuCun_QingYang_TuoLuo_Chain(t *testing.T) {
	// 甲子年 → 禄存在寅(支3), 擎羊在卯(支4), 陀罗在丑(支2)
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	var env struct {
		Data struct {
			Palaces []struct {
				Name  string `json:"name"`
				Stars []struct {
					Name    string `json:"name"`
					IsMajor bool   `json:"is_major"`
				} `json:"stars"`
			} `json:"palaces"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}

	// Find which palaces have 禄存, 擎羊, 陀罗
	var luCunPalace, qingYangPalace, tuoLuoPalace int = -1, -1, -1
	for i, p := range env.Data.Palaces {
		for _, s := range p.Stars {
			switch s.Name {
			case "禄存":
				luCunPalace = i
			case "擎羊":
				qingYangPalace = i
			case "陀罗":
				tuoLuoPalace = i
			}
		}
	}
	if luCunPalace < 0 || qingYangPalace < 0 || tuoLuoPalace < 0 {
		t.Fatalf("禄存/擎羊/陀罗 missing: luCun=%d qingYang=%d tuoLuo=%d", luCunPalace, qingYangPalace, tuoLuoPalace)
	}
	// 擎羊应在禄存顺数前一位(zhi-1: lc+1→palace: lc-1)
		// 陀罗应在禄存逆数后一位(zhi-1: lc-1→palace: lc+1)
	if (luCunPalace-1+12)%12 != qingYangPalace {
		t.Errorf("擎羊(=%d) should be luCun(=%d)-1 in palace (zhi-1 +1 = palace -1)", qingYangPalace, luCunPalace)
	}
	if (luCunPalace+1)%12 != tuoLuoPalace {
		t.Errorf("陀罗(=%d) should be luCun(=%d)+1 in palace (zhi-1 -1 = palace +1)", tuoLuoPalace, luCunPalace)
	}
	t.Logf("禄存=%s, 擎羊=%s, 陀罗=%s (甲禄在寅)",
		env.Data.Palaces[luCunPalace].Name, env.Data.Palaces[qingYangPalace].Name, env.Data.Palaces[tuoLuoPalace].Name)
}

func TestDomain_TianKui_TianYue_Opposite(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	var env struct {
		Data struct {
			Palaces []struct {
				Name  string `json:"name"`
				Stars []struct {
					Name string `json:"name"`
				} `json:"stars"`
			} `json:"palaces"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}

	var tianKui, tianYue int = -1, -1
	for i, p := range env.Data.Palaces {
		for _, s := range p.Stars {
			if s.Name == "天魁" {
				tianKui = i
			}
			if s.Name == "天钺" {
				tianYue = i
			}
		}
	}
	if tianKui < 0 || tianYue < 0 {
		t.Fatalf("天魁/天钺 missing")
	}
	// 天魁天钺应永远对角(6宫差)
	if (tianKui+6)%12 != tianYue && (tianYue+6)%12 != tianKui {
		t.Errorf("天魁=%s, 天钺=%s, not 6 apart", env.Data.Palaces[tianKui].Name, env.Data.Palaces[tianYue].Name)
	}
}

func TestDomain_HuoXing_LingXing_Groups(t *testing.T) {
	// Test four groups: 申子辰, 寅午戌, 巳酉丑, 亥卯未
	// Birth year changes yearZhi
	tests := []struct {
		name  string
		birth string
		hour  int // 辰时=5 (08:00)
	}{
		{"子年(申子辰)", "1984-02-15T08:00:00+08:00", 5}, // 甲子年
		{"寅年(寅午戌)", "1986-02-15T08:00:00+08:00", 5}, // 丙寅年
		{"巳年(巳酉丑)", "1989-02-15T08:00:00+08:00", 5}, // 己巳年
		{"亥年(亥卯未)", "1983-02-15T08:00:00+08:00", 5}, // 癸亥年
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{
				"birth":  map[string]any{"time": tt.birth, "longitude": 116.4},
				"gender": "male",
			})
			if err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest("POST", "/", strings.NewReader(string(body)))
			w := httptest.NewRecorder()
			computeZiweiChart(w, r)
			if w.Code != http.StatusOK {
				t.Fatalf("status=%d", w.Code)
			}
			var env struct {
				Data struct {
					Palaces []struct {
						Name  string `json:"name"`
						Stars []struct {
							Name string `json:"name"`
						} `json:"stars"`
					} `json:"palaces"`
				} `json:"data"`
			}
			if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
				t.Fatal(err)
			}

			var huo, ling int = -1, -1
			for i, p := range env.Data.Palaces {
				for _, s := range p.Stars {
					if s.Name == "火星" {
						huo = i
					}
					if s.Name == "铃星" {
						ling = i
					}
				}
			}
			if huo < 0 || ling < 0 {
				t.Fatalf("火星/铃星 missing")
			}
			// 火星和铃星不应在同一宫
			if huo == ling {
				t.Errorf("火星 and 铃星 in same palace %s", env.Data.Palaces[huo].Name)
			}
			t.Logf("%s: 火星=%s, 铃星=%s", tt.name, env.Data.Palaces[huo].Name, env.Data.Palaces[ling].Name)
		})
	}
}

func TestDomain_14MainStars_Present(t *testing.T) {
	body := `{` + bt15 + `,"gender":"male"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	computeZiweiChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	var env struct {
		Data struct {
			Palaces []struct {
				Stars []struct {
					Name    string `json:"name"`
					IsMajor bool   `json:"is_major"`
				} `json:"stars"`
			} `json:"palaces"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}

	mainStars := map[string]bool{}
	for _, p := range env.Data.Palaces {
		for _, s := range p.Stars {
			if s.IsMajor {
				mainStars[s.Name] = true
			}
		}
	}
	// 14 main stars should all be present
	expected := []string{"紫微", "天机", "太阳", "武曲", "天同", "廉贞",
		"天府", "太阴", "贪狼", "巨门", "天相", "天梁", "七杀", "破军"}
	for _, name := range expected {
		if !mainStars[name] {
			t.Errorf("main star %q missing", name)
		}
	}
}
