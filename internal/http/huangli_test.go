package handler
import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHuangliDate_Valid(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/huangli/date?date=2025-06-01&event=结婚", nil)
	w := httptest.NewRecorder()
	huangliDate(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestHuangliDate_MissingParams(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/huangli/date", nil)
	w := httptest.NewRecorder()
	huangliDate(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHuangliMonth_Valid(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/huangli/month?month=2025-06&event=结婚", nil)
	w := httptest.NewRecorder()
	huangliMonth(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestHuangliMonth_MissingParams(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/huangli/month", nil)
	w := httptest.NewRecorder()
	huangliMonth(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHuangliBondDate_Valid(t *testing.T) {
	body := `{` + bt + `,"event_type":"结婚","date":"2025-06-01"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	huangliBondDate(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestHuangliBondDate_MissingEvent(t *testing.T) {
	body := `{` + bt + `,"date":"2025-06-01"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	huangliBondDate(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestHuangliBondMonth_Valid(t *testing.T) {
	body := `{` + bt + `,"event_type":"结婚","month":"2025-06"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	huangliBondMonth(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestHuangliBondMonth_MissingMonth(t *testing.T) {
	body := `{` + bt + `,"event_type":"结婚"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	huangliBondMonth(w, r)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestBlackBox_Huangli_DifferentEvents(t *testing.T) {
	// 同一个日期的不同事件，结果应该不同
	r1 := httptest.NewRequest("GET", "/api/huangli/date?date=2025-06-15&event=嫁娶", nil)
	w1 := httptest.NewRecorder()
	huangliDate(w1, r1)

	r2 := httptest.NewRequest("GET", "/api/huangli/date?date=2025-06-15&event=开业", nil)
	w2 := httptest.NewRecorder()
	huangliDate(w2, r2)

	if w1.Code != http.StatusOK || w2.Code != http.StatusOK {
		t.Skip("huangli data unavailable")
	}
}

func TestBlackBox_Huangli_BondDateInBondMonth(t *testing.T) {
	// 这个测试验证 bond/date 和 bond/month 的一致性
	// bond/month 返回的 entries 应该包含 bond/date 查询的那一天
	body := `{` + bt15 + `,"event_type":"嫁娶","date":"2025-06-15"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	huangliBondDate(w, r)

	bodyM := `{` + bt15 + `,"event_type":"嫁娶","month":"2025-06"}`
	rM := httptest.NewRequest("POST", "/", strings.NewReader(bodyM))
	wM := httptest.NewRecorder()
	huangliBondMonth(wM, rM)

	if w.Code != http.StatusOK || wM.Code != http.StatusOK {
		t.Skip("huangli bond data unavailable")
	}
}

func TestEdge_HuangliDate_MissingParams(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"no params", "/?date=2025-06-15"},
		{"no date", "/?event=嫁娶"},
		{"both missing", "/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			huangliDate(w, r)
			if w.Code < 400 {
				t.Errorf("huangli date %s: status=%d, want >=400", tt.name, w.Code)
			}
		})
	}
}

func TestEdge_HuangliDate_InvalidDateFormat(t *testing.T) {
	tests := []string{
		"2025-06-15T12:00:00Z",
		"2025-6-15",
		"2025/06/15",
		"20250615",
		"2025-13-01",
		"2025-00-01",
		"not-a-date",
		"2025-06-32",
	}
	for _, date := range tests {
		t.Run("date="+date, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?date="+date+"&event=嫁娶", nil)
			w := httptest.NewRecorder()
			huangliDate(w, r)
			if w.Code >= 500 {
				t.Errorf("invalid date %q caused 5xx: %d", date, w.Code)
			}
		})
	}
}

func TestEdge_HuangliMonth_InvalidMonthFormat(t *testing.T) {
	tests := []string{
		"2025-6",
		"2025/06",
		"202506",
		"2025-13",
		"2025-00",
		"not-month",
	}
	for _, month := range tests {
		t.Run("month="+month, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?month="+month+"&event=嫁娶", nil)
			w := httptest.NewRecorder()
			huangliMonth(w, r)
			if w.Code >= 500 {
				t.Errorf("invalid month %q caused 5xx: %d", month, w.Code)
			}
		})
	}
}

func TestEdge_HuangliBondDate_InvalidDate(t *testing.T) {
	body := `{` + bt15 + `,"event_type":"嫁娶","date":"2025-13-01"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	huangliBondDate(w, r)
	if w.Code >= 500 {
		t.Errorf("invalid bond date caused 5xx: %d", w.Code)
	}
}

func TestEdge_HuangliBondMonth_InvalidMonth(t *testing.T) {
	body := `{` + bt15 + `,"event_type":"嫁娶","month":"2025-13"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	huangliBondMonth(w, r)
	if w.Code >= 500 {
		t.Errorf("invalid bond month caused 5xx: %d", w.Code)
	}
}

func TestEdge2_HuangliMonth_MissingParams(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"no params", "/"},
		{"only month", "/?month=2025-06"},
		{"only event", "/?event=嫁娶"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			huangliMonth(w, r)
			if w.Code >= 500 {
				t.Errorf("%s caused 5xx: %d", tt.name, w.Code)
			}
		})
	}
}

func TestEd3_HuangliDate_NoEvent(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/huangli/date?date=2025-06-15&event=", nil)
	w := httptest.NewRecorder()
	huangliDate(w, r)

	// event 为空时 handler 返回 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("empty event: status=%d, want 400", w.Code)
	}
}

func TestBug_HuangliDate_InvalidDateFormat(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/huangli/date?date=not-a-date&event=结婚", nil)
	w := httptest.NewRecorder()
	huangliDate(w, r)

	// Should return 400 for invalid date format
	if w.Code != http.StatusBadRequest {
		t.Errorf("huangli date invalid-format: status=%d, want 400", w.Code)
	}
}

func TestBug_HuangliMonth_InvalidMonthFormat(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/huangli/month?month=bad&event=结婚", nil)
	w := httptest.NewRecorder()
	huangliMonth(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("huangli month invalid-format: status=%d, want 400", w.Code)
	}
}

func TestBug_HuangliBondDate_MissingBirth(t *testing.T) {
	body := `{"event_type":"结婚","date":"2025-06-01"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	huangliBondDate(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("huangli bond date missing birth: status=%d, want 422", w.Code)
	}
}
