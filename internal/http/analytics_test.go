package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAnalytics(t *testing.T) {
	a := NewAnalytics()

	// Record page views
	a.RecordPageView("/")
	a.RecordPageView("/")
	a.RecordPageView("/chart")
	a.RecordCheckout()
	a.RecordCheckout()
	a.RecordReportView()

	snap := a.Snapshot()
	pv := snap["page_views"].(map[string]int64)
	if pv["/"] != 2 {
		t.Errorf("page_views[/]=%d, want 2", pv["/"])
	}
	if pv["/chart"] != 1 {
		t.Errorf("page_views[/chart]=%d, want 1", pv["/chart"])
	}
	if c := snap["checkouts"].(int64); c != 2 {
		t.Errorf("checkouts=%d, want 2", c)
	}
	if r := snap["reports"].(int64); r != 1 {
		t.Errorf("reports=%d, want 1", r)
	}
}

func TestHandlePageView(t *testing.T) {
	a := NewAnalytics()
	handler := handlePageView(a)

	t.Run("valid path", func(t *testing.T) {
		body, err := json.Marshal(map[string]string{"path": "/test"})
		if err != nil { t.Fatal(err) }
		req := httptest.NewRequest("POST", "/api/analytics/pageview", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusNoContent {
			t.Errorf("status=%d, want 204", w.Code)
		}
		snap := a.Snapshot()
		pv := snap["page_views"].(map[string]int64)
		if pv["/test"] != 1 {
			t.Errorf("page_views[/test]=%d, want 1", pv["/test"])
		}
	})

	t.Run("empty body returns 400", func(t *testing.T) {
		a2 := NewAnalytics()
		handler2 := handlePageView(a2)
		req := httptest.NewRequest("POST", "/some-page", nil)
		w := httptest.NewRecorder()
		handler2(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("status=%d, want 400 for empty body", w.Code)
		}
		// No page view should be recorded on decode failure.
		snap := a2.Snapshot()
		pv := snap["page_views"].(map[string]int64)
		if pv["/some-page"] != 0 {
			t.Errorf("page_views[/some-page]=%d, want 0 (decode failure)", pv["/some-page"])
		}
	})
}

func TestHandleStats(t *testing.T) {
	a := NewAnalytics()
	a.RecordPageView("/")
	a.RecordCheckout()
	handler := handleStats(a)
	req := httptest.NewRequest("GET", "/api/analytics/stats", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status=%d, want 200", w.Code)
	}
	var body struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data == nil {
		t.Fatal("data is nil")
	}
}

func TestEd3_Analytics_PageView_Valid(t *testing.T) {
	a := NewAnalytics()
	h := handlePageView(a)

	body := `{"path":"/bazi"}`
	r := httptest.NewRequest("POST", "/api/analytics/pageview", strings.NewReader(body))
	w := httptest.NewRecorder()
	h(w, r)
	if w.Code != http.StatusNoContent {
		t.Errorf("status=%d, want 204", w.Code)
	}

	// stats 应有数据
	r2 := httptest.NewRequest("GET", "/api/stats", nil)
	w2 := httptest.NewRecorder()
	handleStats(a)(w2, r2)
	var env struct {
		Data struct {
			PageViews map[string]int64 `json:"page_views"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w2.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Data.PageViews["/bazi"] != 1 {
		t.Errorf("page_views[/bazi]=%d, want 1", env.Data.PageViews["/bazi"])
	}
}

func TestEd3_Analytics_PageView_EmptyBody(t *testing.T) {
	a := NewAnalytics()
	h := handlePageView(a)

	r := httptest.NewRequest("POST", "/api/analytics/pageview", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	h(w, r)
	// 空 body 用 r.URL.Path 作为 path
	if w.Code != http.StatusNoContent {
		t.Errorf("status=%d, want 204", w.Code)
	}
}

func TestEd3_Analytics_PageView_InvalidJSON(t *testing.T) {
	a := NewAnalytics()
	h := handlePageView(a)

	r := httptest.NewRequest("POST", "/api/analytics/pageview", strings.NewReader("bad"))
	w := httptest.NewRecorder()
	h(w, r)
	if w.Code >= 500 {
		t.Errorf("invalid JSON caused 5xx: %d", w.Code)
	}
}

func TestEd3_Analytics_PageView_VeryLongPath(t *testing.T) {
	a := NewAnalytics()
	h := handlePageView(a)

	longPath := "/" + strings.Repeat("x", 10000)
	body := `{"path":"` + longPath + `"}`
	r := httptest.NewRequest("POST", "/api/analytics/pageview", strings.NewReader(body))
	w := httptest.NewRecorder()
	h(w, r)
	if w.Code >= 500 {
		t.Errorf("very long path caused 5xx: %d", w.Code)
	}
}

func TestEd3_Analytics_Stats_Concurrent(t *testing.T) {
	a := NewAnalytics()
	// 应能安全并发读写
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			a.RecordPageView("/test")
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	snap := a.Snapshot()
	pv := snap["page_views"].(map[string]int64)
	if pv["/test"] != 10 {
		t.Errorf("concurrent page_views: got %d, want 10", pv["/test"])
	}
}

func TestEd3_Stats_AllMethods(t *testing.T) {
	a := NewAnalytics()
	h := handleStats(a)
	for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
		t.Run(method, func(t *testing.T) {
			r := httptest.NewRequest(method, "/api/stats", nil)
			w := httptest.NewRecorder()
			h(w, r)
			if w.Code >= 500 {
				t.Errorf("%s stats: status=%d", method, w.Code)
			}
		})
	}
}
