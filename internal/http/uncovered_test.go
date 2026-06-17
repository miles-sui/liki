package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"liki/internal/engine/ganzhi"
)

func TestValidStem(t *testing.T) {
	tests := []struct {
		p    ganzhi.Zhu
		want bool
	}{
		{ganzhi.Zhu{Gan: 1, Zhi: 1}, true},
		{ganzhi.Zhu{Gan: 10, Zhi: 12}, true},
		{ganzhi.Zhu{Gan: 0, Zhi: 1}, false},
		{ganzhi.Zhu{Gan: 11, Zhi: 1}, false},
	}
	for _, tc := range tests {
		got := validGan(tc.p)
		if got != tc.want {
			t.Errorf("validGan(%+v)=%v, want %v", tc.p, got, tc.want)
		}
	}
}

func TestIsValidDate(t *testing.T) {
	tests := []struct {
		y, m, d int
		want    bool
	}{
		{2024, 1, 1, true},
		{2024, 2, 29, true},  // leap year
		{2023, 2, 29, false}, // non-leap year
		{2024, 6, 15, true},
		{2024, 12, 31, true},
		{2024, 0, 15, false},
		{2024, 13, 1, false},
		{2024, 6, 0, false},
		{2024, 4, 31, false}, // April has 30 days
	}
	for _, tc := range tests {
		got := isValidDate(tc.y, tc.m, tc.d)
		if got != tc.want {
			t.Errorf("isValidDate(%d,%d,%d)=%v, want %v", tc.y, tc.m, tc.d, got, tc.want)
		}
	}
}

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

	t.Run("empty body uses URL path", func(t *testing.T) {
		a2 := NewAnalytics()
		handler2 := handlePageView(a2)
		req := httptest.NewRequest("POST", "/some-page", nil)
		w := httptest.NewRecorder()
		handler2(w, req)
		snap := a2.Snapshot()
		pv := snap["page_views"].(map[string]int64)
		if pv["/some-page"] != 1 {
			t.Errorf("page_views[/some-page]=%d, want 1", pv["/some-page"])
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
