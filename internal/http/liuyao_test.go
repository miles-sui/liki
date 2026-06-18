package handler
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleLiuyaoChart_Valid(t *testing.T) {
	body := `{` + bt + `,"yong_shen":"世爻"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleLiuyaoChart(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env struct {
		Data struct {
			Name  string `json:"name"`
			Lines []struct {
				Position int    `json:"position"`
				LiuQin   int    `json:"liu_qin"`
			} `json:"lines"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Name == "" {
		t.Error("name is empty")
	}
	if len(env.Data.Lines) != 6 {
		t.Errorf("lines = %d, want 6", len(env.Data.Lines))
	}
}

func TestHandleLiuyaoChart_DefaultsToShiYao(t *testing.T) {
	body := `{` + bt + `,"yong_shen":""}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleLiuyaoChart(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (defaults to 世爻)", w.Code)
	}
	var env struct {
		Data struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Name == "" {
		t.Error("name is empty (should default to 世爻)")
	}
}

func TestHandleLiuyaoChart_InvalidYongShen(t *testing.T) {
	body := `{` + bt + `,"yong_shen":"invalid"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleLiuyaoChart(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestHandleLiuyaoChart_InvalidBody(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{bad`))
	w := httptest.NewRecorder()
	handleLiuyaoChart(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBlackBox_LiuYao_Invariants(t *testing.T) {
	body := `{` + bt15 + `,"yong_shen":"世爻"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleLiuyaoChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}

	var env struct {
		Data struct {
			Name  string `json:"name"`
			Lines []struct {
				Position int `json:"position"`
				LiuQin   int `json:"liu_qin"`
			} `json:"lines"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}

	if env.Data.Name == "" {
		t.Error("hexagram name is empty")
	}
	if len(env.Data.Lines) != 6 {
		t.Errorf("lines = %d, want 6", len(env.Data.Lines))
	}

	// 每爻的 position 应该是 1-6, liu_qin 是六亲编码 0-5
	positions := make(map[int]bool)
	for _, line := range env.Data.Lines {
		if line.Position < 1 || line.Position > 6 {
			t.Errorf("invalid position %d", line.Position)
		}
		positions[line.Position] = true
		if line.LiuQin < 0 || line.LiuQin > 5 {
			t.Errorf("line[%d]: liu_qin=%d out of range [0,5]", line.Position, line.LiuQin)
		}
	}
	if len(positions) != 6 {
		t.Error("duplicate or missing positions")
	}
}

func TestEdge_LiuYao_InvalidFixedValues(t *testing.T) {
	// fixed 数组值应该在 0-3 范围内(老阴/少阴/少阳/老阳)
	// 超出范围可能 crash 或产生异常结果
	tests := []struct {
		name  string
		fixed [6]int
	}{
		{"all 4", [6]int{4, 4, 4, 4, 4, 4}},
		{"all -1", [6]int{-1, -1, -1, -1, -1, -1}},
		{"all 99", [6]int{99, 99, 99, 99, 99, 99}},
		{"mixed invalid", [6]int{0, 1, 2, 3, 4, -1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{
				"birth":     map[string]any{"time": "1984-02-15T08:00:00+08:00", "longitude": 116.4},
				"yong_shen": "世爻",
				"fixed":     tt.fixed,
			})
			if err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest("POST", "/", strings.NewReader(string(body)))
			w := httptest.NewRecorder()
			handleLiuyaoChart(w, r)
			if w.Code >= 500 {
				t.Errorf("invalid fixed caused 5xx: %d", w.Code)
			}
		})
	}
}

func TestBug_Liuyao_YongShenInvalid_Rejected(t *testing.T) {
	body := `{` + bt15 + `,"yong_shen":"子虚"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleLiuyaoChart(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("liuyao yong_shen=子虚: status=%d, want 422", w.Code)
	}
}

func TestBug_Liuyao_YongShenEmpty_DefaultsToShiYao(t *testing.T) {
	body := `{` + bt15 + `,"yong_shen":""}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleLiuyaoChart(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("liuyao yong_shen=empty: status=%d, want 200", w.Code)
	}
}

func TestDomain_LiuYao_Chart(t *testing.T) {
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"yong_shen":"世爻"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleLiuyaoChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, body=%s", w.Code, w.Body.String())
	}
	var env struct {
		Data struct {
			Name  string `json:"name"`
			Lines []struct {
				LiuQin  int    `json:"liu_qin"`
				LiuShou int    `json:"liu_shou"`
				ShiYing string `json:"shi_ying"`
				Zhi     string `json:"zhi"`
			} `json:"lines"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}

	if env.Data.Name == "" {
		t.Error("hexagram name is empty")
	}
	if len(env.Data.Lines) != 6 {
		t.Errorf("lines=%d, want 6", len(env.Data.Lines))
	}
	// 六爻应有世应标识
	hasShi := false
	hasYing := false
	for _, line := range env.Data.Lines {
		if line.ShiYing == "世" {
			hasShi = true
		}
		if line.ShiYing == "应" {
			hasYing = true
		}
		if line.LiuQin < 0 || line.LiuQin > 4 {
			t.Error("liu_qin out of range (0-4)")
		}
		if line.LiuShou < 0 || line.LiuShou > 5 {
			t.Error("liu_shou out of range (0-5)")
		}
		if line.Zhi == "" {
			t.Error("line with empty zhi")
		}
	}
	if !hasShi || !hasYing {
		t.Errorf("世应 missing: shi=%v ying=%v", hasShi, hasYing)
	}
}
