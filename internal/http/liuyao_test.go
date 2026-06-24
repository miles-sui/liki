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
}

func TestHandleLiuyaoChart_DefaultsToShiYao(t *testing.T) {
	body := `{` + bt + `,"yong_shen":""}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleLiuyaoChart(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (defaults to 世爻)", w.Code)
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
