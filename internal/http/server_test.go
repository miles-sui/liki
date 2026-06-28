package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"liki/internal/payment"
)

func TestRegisterRoutes_DoesNotPanic(t *testing.T) {
	mux := http.NewServeMux()
	rl := NewRateLimiter()
	defer rl.Stop()
	deps := ServerDeps{
		Payment:   &payment.Service{},
		Store:     &payment.Store{},
		Analytics: &Analytics{},
	}
	RegisterRoutes(mux, deps, "test", rl)

	r := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("health check after RegisterRoutes: status = %d, want 200", w.Code)
	}
}

func TestRegisterRoutes_JSONRPC(t *testing.T) {
	mux := http.NewServeMux()
	rl := NewRateLimiter()
	defer rl.Stop()
	deps := ServerDeps{
		Payment:   &payment.Service{},
		Store:     &payment.Store{},
		Analytics: &Analytics{},
	}
	RegisterRoutes(mux, deps, "test", rl)

	r := httptest.NewRequest("POST", "/jsonrpc", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code >= 500 {
		t.Errorf("jsonrpc route: status = %d", w.Code)
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
