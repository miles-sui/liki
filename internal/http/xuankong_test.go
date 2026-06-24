package handler
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestXuankongSanYuan_Default(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/xuankong/sanyuan", nil)
	w := httptest.NewRecorder()
	xuankongSanYuan(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestXuankongSanYuan_WithYear(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/xuankong/sanyuan?year=2025", nil)
	w := httptest.NewRecorder()
	xuankongSanYuan(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestXuankongChart_Valid(t *testing.T) {
	body := `{` + bt + `,"sit_mountain":1,"face_mountain":12}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	xuankongChart(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestXuankongChart_MissingFace(t *testing.T) {
	body := `{` + bt + `,"sit_mountain":1}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	xuankongChart(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestBlackBox_Xuankong_Invariants(t *testing.T) {
	body := `{` + bt15 + `,"sit_mountain":1,"face_mountain":12}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	xuankongChart(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestBlackBox_SanYuan_Invariants(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/xuankong/sanyuan?year=2025", nil)
	w := httptest.NewRecorder()
	xuankongSanYuan(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestEdge_XuankongSanYuan_InvalidYear(t *testing.T) {
	tests := []string{"", "abc", "-100", "0", "10000"}
	for _, year := range tests {
		t.Run("year="+year, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?year="+year, nil)
			w := httptest.NewRecorder()
			xuankongSanYuan(w, r)
			if w.Code >= 500 {
				t.Errorf("year=%q caused 5xx: %d", year, w.Code)
			}
		})
	}
}

func TestEdge_Boundary_XuankongMountainZero(t *testing.T) {
	// sit_mountain=0, face_mountain=0 是有效值
	body, err := json.Marshal(map[string]any{
		"birth":         map[string]any{"time": "1984-02-15T08:00:00+08:00", "longitude": 116.4},
		"sit_mountain":  0,
		"face_mountain": 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/", strings.NewReader(string(body)))
	w := httptest.NewRecorder()
	xuankongChart(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("mountains=0: status=%d", w.Code)
	}
}

func TestEdge_Boundary_MountainMaxValue(t *testing.T) {
	// sit/face mountain max=23
	body, err := json.Marshal(map[string]any{
		"birth":         map[string]any{"time": "1984-02-15T08:00:00+08:00", "longitude": 116.4},
		"sit_mountain":  23,
		"face_mountain": 23,
	})
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/", strings.NewReader(string(body)))
	w := httptest.NewRecorder()
	xuankongChart(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("mountains=23: status=%d", w.Code)
	}
}

func TestEd3_Xuankong_MountainNegativeOne(t *testing.T) {
	body, err := json.Marshal(map[string]any{
		"birth":         map[string]any{"time": "1984-02-15T08:00:00+08:00", "longitude": 116.4},
		"sit_mountain":  -1,
		"face_mountain": 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/", strings.NewReader(string(body)))
	w := httptest.NewRecorder()
	xuankongChart(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("sit_mountain=-1: status=%d, want 422", w.Code)
	}
}

func TestEd3_Xuankong_Mountain24(t *testing.T) {
	body, err := json.Marshal(map[string]any{
		"birth":         map[string]any{"time": "1984-02-15T08:00:00+08:00", "longitude": 116.4},
		"sit_mountain":  0,
		"face_mountain": 24,
	})
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/", strings.NewReader(string(body)))
	w := httptest.NewRecorder()
	xuankongChart(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("face_mountain=24: status=%d, want 422", w.Code)
	}
}

func TestBug_XuankongChart_SitMountainZero_Rejected(t *testing.T) {
	// Mountain index 0 (坎) is valid, but validation.Required on int
	// treats 0 as blank → should this be 200 or 422?
	body := `{` + bt15 + `,"sit_mountain":0,"face_mountain":12}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	xuankongChart(w, r)

	// Current behavior: 422 because validation.Required treats 0 as blank.
	// If this is 422, mountain 0 can never be selected — a design bug.
	if w.Code == http.StatusUnprocessableEntity {
		t.Error("BUG: sit_mountain=0 rejected (validation.Required on int)")
	}
	if w.Code == http.StatusOK {
		t.Log("OK: sit_mountain=0 accepted")
	}
}

func TestBug_XuankongChart_FaceMountainZero_Rejected(t *testing.T) {
	body := `{` + bt15 + `,"sit_mountain":1,"face_mountain":0}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	xuankongChart(w, r)

	if w.Code == http.StatusUnprocessableEntity {
		t.Error("BUG: face_mountain=0 rejected (validation.Required on int)")
	}
}

func TestBug_XuankongChart_SitMountainNegative(t *testing.T) {
	body := `{` + bt15 + `,"sit_mountain":-1,"face_mountain":12}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	xuankongChart(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("xuankong sit=-1: status=%d, want 422 (min=0)", w.Code)
	}
}

func TestBug_XuankongChart_SitMountain24(t *testing.T) {
	body := `{` + bt15 + `,"sit_mountain":24,"face_mountain":12}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	xuankongChart(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("xuankong sit=24: status=%d, want 422 (max=23)", w.Code)
	}
}

func TestBug_SanYuan_YearZero(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CRASH BUG: xuankongSanYuan panics on year=0: %v", r)
		}
	}()
	r := httptest.NewRequest("GET", "/api/xuankong/sanyuan?year=0", nil)
	w := httptest.NewRecorder()
	xuankongSanYuan(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("sanyuan year=0: status=%d, want 200", w.Code)
	}
}

func TestBug_SanYuan_YearNegative(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CRASH BUG: xuankongSanYuan panics on year=-100: %v", r)
		}
	}()
	r := httptest.NewRequest("GET", "/api/xuankong/sanyuan?year=-100", nil)
	w := httptest.NewRecorder()
	xuankongSanYuan(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("sanyuan year=-100: status=%d", w.Code)
	}
}
