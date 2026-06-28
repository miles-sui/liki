package http

import (
	"net/http"
	"sync"
	"sync/atomic"
)

// Analytics tracks page views and conversion events server-side,
// without cookies or fingerprinting.
type Analytics struct {
	pageViews sync.Map // path -> *int64
	checkouts atomic.Int64
	reports   atomic.Int64
}

// NewAnalytics creates a new Analytics instance.
func NewAnalytics() *Analytics { return &Analytics{} }

// RecordPageView increments the page view counter for a path.
func (a *Analytics) RecordPageView(path string) {
	val, _ := a.pageViews.LoadOrStore(path, new(int64))
	atomic.AddInt64(val.(*int64), 1)
}

// RecordCheckout increments the checkout conversion counter.
func (a *Analytics) RecordCheckout() { a.checkouts.Add(1) }

// RecordReportView increments the report view counter.
func (a *Analytics) RecordReportView() { a.reports.Add(1) }

// Snapshot returns a copy of the current counters.
func (a *Analytics) Snapshot() map[string]any {
	pages := make(map[string]int64)
	a.pageViews.Range(func(k, v any) bool {
		pages[k.(string)] = atomic.LoadInt64(v.(*int64))
		return true
	})
	return map[string]any{
		"page_views": pages,
		"checkouts":  a.checkouts.Load(),
		"reports":    a.reports.Load(),
	}
}

// handlePageView records a page view via POST beacon (no cookies).
func handlePageView(a *Analytics) http.HandlerFunc {
	type pageViewReq struct {
		Path string `json:"path"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		body, ok := decodeJSON[pageViewReq](w, r)
		if !ok {
			return
		}
		path := r.URL.Path
		if body.Path != "" {
			path = body.Path
		}
		a.RecordPageView(path)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleStats(a *Analytics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, a.Snapshot())
	}
}
