package http

import (
	"net/http"

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

	return Recover(SecurityHeaders(CORSMiddleware(deps.DevMode, BodyLimit(mux))))
}
