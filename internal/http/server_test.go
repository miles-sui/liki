package handler
import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"liki/internal/payment"
	"liki/internal/session"
)

func TestRegisterCoreRoutes(t *testing.T) {
	mux := http.NewServeMux()
	rl := NewRateLimiter()
	defer rl.Stop()
	registerCoreRoutes(mux, rl)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		want   int
	}{
		{
			name:   "GET huangli date",
			method: "GET",
			path:   "/api/huangli/date?date=2025-06-01&event=结婚",
			want:   http.StatusOK,
		},
		{
			name:   "GET huangli month",
			method: "GET",
			path:   "/api/huangli/month?month=2025-06&event=结婚",
			want:   http.StatusOK,
		},
		{
			name:   "GET xuankong sanyuan",
			method: "GET",
			path:   "/api/xuankong/sanyuan",
			want:   http.StatusOK,
		},
		{
			name:   "POST bazhai minggua",
			method: "POST",
			path:   "/api/bazhai/minggua",
			body:   `{"gender":"male","birth_year":1990}`,
			want:   http.StatusOK,
		},
		{
			name:   "POST bazi chart",
			method: "POST",
			path:   "/api/bazi/chart",
			body:   `{` + bt + `,"gender":"male"}`,
			want:   http.StatusOK,
		},
		{
			name:   "POST ziwei chart",
			method: "POST",
			path:   "/api/ziwei/chart",
			body:   `{` + bt + `,"gender":"male"}`,
			want:   http.StatusOK,
		},
		{
			name:   "POST qimen pan",
			method: "POST",
			path:   "/api/qimen/pan",
			body:   `{` + bt + `,"kind":"shi"}`,
			want:   http.StatusOK,
		},
		{
			name:   "POST liuyao chart",
			method: "POST",
			path:   "/api/liuyao/chart",
			body:   `{` + bt + `,"yong_shen":"世爻"}`,
			want:   http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r *http.Request
			if tt.body != "" {
				r = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			} else {
				r = httptest.NewRequest(tt.method, tt.path, nil)
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}

func TestRegisterRoutes_DoesNotPanic(t *testing.T) {
	mux := http.NewServeMux()
	rl := NewRateLimiter()
	defer rl.Stop()
	deps := ServerDeps{
		Payment:      &payment.Service{},
		Store:        &payment.Store{},
		SessionStore: &session.Store{},
		Analytics:    &Analytics{},
	}
	RegisterRoutes(mux, deps, "test", rl)

	// Verify a core route still works
	r := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("health check after RegisterRoutes: status = %d, want 200", w.Code)
	}
}

func TestEd3_Version_AllMethods(t *testing.T) {
	h := handleVersion("test-build")
	for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
		t.Run(method, func(t *testing.T) {
			r := httptest.NewRequest(method, "/api/version", nil)
			w := httptest.NewRecorder()
			h(w, r)
			if w.Code >= 500 {
				t.Errorf("%s version: status=%d", method, w.Code)
			}
		})
	}
}

func TestEd3_WrongMethod_OnCoreEndpoints(t *testing.T) {
	// 用 GET 调用 POST-only 的 core handler — 直接调 handler 会怎样?
	tests := []struct {
		name    string
		handler http.HandlerFunc
		body    string
	}{
		{"bazi chart GET", computeChart, `{` + bt15 + `,"gender":"male"}`},
		{"ziwei chart GET", computeZiweiChart, `{` + bt15 + `,"gender":"male"}`},
		{"qimen pan GET", handleQimenPan, `{` + bt15 + `,"kind":"shi"}`},
		{"liuyao chart GET", handleLiuyaoChart, `{` + bt15 + `,"yong_shen":"世爻"}`},
		{"huangli bond date GET", huangliBondDate, `{` + bt15 + `,"event_type":"嫁娶","date":"2025-06-15"}`},
		{"qiming wuge GET", handleWuge, `{"surname":"张","yong_shen":"木"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			tt.handler(w, r)
			// handler 函数本身不检查 method (路由层检查)
			// 所以 GET 也能调 POST handler
			if w.Code >= 500 {
				t.Errorf("GET on POST handler caused 5xx: %d", w.Code)
			}
		})
	}
}
