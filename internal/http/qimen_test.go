package handler
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleQimenPan_Valid_Shi(t *testing.T) {
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":"shi"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestHandleQimenPan_Valid_Ri(t *testing.T) {
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":"ri"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestHandleQimenPan_Valid_Nian(t *testing.T) {
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":"nian"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestHandleQimenPan_DefaultsToShi(t *testing.T) {
	// kind="" defaults to "shi"
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":""}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (defaults to shi)", w.Code)
	}
}

func TestHandleQimenPan_InvalidKind(t *testing.T) {
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":"xun"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestHandleQimenPan_MissingBirth(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"kind":"shi"}`))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestHandleQimenPan_InvalidJSON(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{bad`))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBlackBox_QiMen_Invariants(t *testing.T) {
	body := `{` + bt15 + `,"kind":"shi"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestEdge_Qimen_InvalidKind(t *testing.T) {
	tests := []string{"SHI", "Shi", "时", "时辰", "day", "YUE", " "}
	for _, kind := range tests {
		t.Run("kind="+kind, func(t *testing.T) {
			body := `{` + bt15 + `,"kind":"` + kind + `"}`
			r := httptest.NewRequest("POST", "/", strings.NewReader(body))
			w := httptest.NewRecorder()
			handleQimenPan(w, r)
			if w.Code == http.StatusOK {
				t.Errorf("BUG: qimen kind=%q accepted", kind)
			}
			if w.Code >= 500 {
				t.Errorf("qimen kind=%q caused 5xx: %d", kind, w.Code)
			}
		})
	}
}

func TestEdge2_Qimen_KindCaseVariations(t *testing.T) {
	tests := []string{"SHI", "Shi", "Ri", "NIAN", "Yue"}
	for _, kind := range tests {
		t.Run(kind, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{
				"birth": map[string]any{"time": "1984-02-15T08:00:00+08:00", "longitude": 116.4},
				"kind":  kind,
			})
			if err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest("POST", "/", strings.NewReader(string(body)))
			w := httptest.NewRecorder()
			handleQimenPan(w, r)
			if w.Code >= 500 {
				t.Errorf("kind=%q caused 5xx: %d", kind, w.Code)
			}
			if w.Code == http.StatusOK {
				t.Errorf("BUG: qimen kind=%q accepted", kind)
			}
		})
	}
}

func TestBug_Qimen_KindYue_Accepted(t *testing.T) {
	// "yue" is valid per validateQimenKind
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":"yue"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("qimen kind=yue: status=%d, want 200", w.Code)
	}
}

func TestBug_Qimen_KindEmpty_DefaultsToShi(t *testing.T) {
	body := `{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":""}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	handleQimenPan(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("qimen kind=empty: status=%d, want 200", w.Code)
	}
}
