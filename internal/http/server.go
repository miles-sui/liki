package http

import (
	"net/http"
	"os"
	"path/filepath"

	"liki/internal/agent"
	"liki/internal/payment"
)

type ServerDeps struct {
	Payment   *payment.Service
	Store     *payment.Store
	ChatAgent *agent.ChatAgent
	Analytics *Analytics
	DevMode   bool
}

func RegisterRoutes(mux *http.ServeMux, deps ServerDeps, buildTime string, rl *RateLimiter) http.Handler {
	// Orders
	mux.HandleFunc("POST /api/orders", handleCreateOrder(deps.Store))
	mux.HandleFunc("POST /api/orders/select", handleOrderSelect(deps.Store))

	// Auth
	mux.HandleFunc("POST /api/auth/login", handleLogin(deps.Store))

	// Payments
	mux.HandleFunc("POST /api/payments/checkout", rl.Wrap(3.0/60, 1, handleCheckout(deps.Payment, deps.Analytics)))
	mux.HandleFunc("POST /api/payments/webhook", handleWebhook(deps.Payment))
	mux.HandleFunc("GET /api/payments/return/{id}", handlePaymentReturn(deps.Store))
	mux.HandleFunc("GET /api/orders/{id}/report", redirectReport())
	mux.HandleFunc("GET /api/orders/{id}/status", handleOrderStatus(deps.Store))
	mux.HandleFunc("POST /api/orders/{id}/retry", handleRetryOrder(deps.Payment))
	mux.HandleFunc("GET /api/reports/{id}", handleReport(deps.Payment, deps.Analytics))
	mux.HandleFunc("GET /api/health", handleHealth())
	// JSON-RPC for external AI agents
	rpcReg := agent.NewRPCRegistry()
	mux.HandleFunc("POST /jsonrpc", rl.Wrap(6000.0/60, 200, handleRPC(rpcReg)))
	mux.HandleFunc("POST /api/analytics/pageview", handlePageView(deps.Analytics))
	mux.HandleFunc("GET /api/stats", handleStats(deps.Analytics))
	mux.HandleFunc("GET /api/version", handleVersion(buildTime))
	mux.HandleFunc("GET /api/location", handleLocation)

	// Agent chat
	mux.HandleFunc("POST /api/agent/naming", rl.Wrap(5.0/60, 3, namingHandler(deps.ChatAgent, deps.Store)))

	handler := Recover(SecurityHeaders(CORSMiddleware(deps.DevMode, BodyLimit(mux))))

	// In dev mode, serve static files from the web/ directory as a fallback.
	// In production, Caddy handles this.
	if deps.DevMode {
		webDir := findWebDir()
		if webDir != "" {
			handler = devFileServer(webDir, handler)
		}
	}

	return handler
}

// findWebDir looks for the web/ directory relative to the working directory.
func findWebDir() string {
	candidates := []string{"web", "../web", "../../web"}
	for _, d := range candidates {
		if fi, err := os.Stat(d); err == nil && fi.IsDir() {
			if abs, err := filepath.Abs(d); err == nil {
				return abs
			}
		}
	}
	return ""
}

// devFileServer wraps an API handler with a static file server fallback.
// If the request path doesn't start with /api/ or /jsonrpc and a file exists
// under webDir, it serves that file. Otherwise it delegates to the API handler.
func devFileServer(webDir string, next http.Handler) http.Handler {
	fs := http.FileServer(http.Dir(webDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API and JSON-RPC requests always go to the API handler.
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			next.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/jsonrpc" {
			next.ServeHTTP(w, r)
			return
		}

		// Try to serve as a static file, with .html fallback.
		p := filepath.Join(webDir, filepath.Clean(r.URL.Path))
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		// Try .html extension (e.g. /chat → /chat.html).
		if fi, err := os.Stat(p + ".html"); err == nil && !fi.IsDir() {
			r2 := r.Clone(r.Context())
			r2.URL.Path = r.URL.Path + ".html"
			fs.ServeHTTP(w, r2)
			return
		}
		// Fall back to API handler (returns 404 for unknown paths).
		next.ServeHTTP(w, r)
	})
}
