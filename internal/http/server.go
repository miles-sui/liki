package handler

import (
	"net/http"

	"liki/internal/agent"
	"liki/internal/payment"
	"liki/internal/session"
)

type ServerDeps struct {
	Payment      *payment.Service
	Store        *payment.Store
	ChatAgent    *agent.ChatAgent
	SessionStore *session.Store
	Analytics    *Analytics
}

func RegisterRoutes(mux *http.ServeMux, deps ServerDeps, buildTime string, rl *RateLimiter) {
	// Paid API
	mux.HandleFunc("POST /api/payments/checkout", rl.Wrap(3.0/60, 1, handleCheckout(deps.Payment)))
	mux.HandleFunc("POST /api/payments/webhook", handleWebhook(deps.Payment))
	mux.HandleFunc("GET /api/payments/return/{id}", handlePaymentReturn())
	mux.HandleFunc("GET /api/orders/{id}/download", redirectDownload())
	mux.HandleFunc("GET /api/orders/{id}/status", handleOrderStatus(deps.Payment))
	mux.HandleFunc("POST /api/orders/{id}/retry", handleRetryOrder(deps.Payment))
	mux.HandleFunc("GET /api/reports/{id}", handleReport(deps.Payment))
	mux.HandleFunc("GET /api/health", handleHealth())
	mux.HandleFunc("POST /api/analytics/pageview", handlePageView(deps.Analytics))
	mux.HandleFunc("GET /api/stats", handleStats(deps.Analytics))
	mux.HandleFunc("GET /api/version", handleVersion(buildTime))
	mux.HandleFunc("GET /api/location", handleLocation)

	// Agent chat
	mux.HandleFunc("POST /api/agent/chat", rl.Wrap(5.0/60, 3, chatHandler(deps.ChatAgent, deps.Store, deps.SessionStore)))
	mux.HandleFunc("GET /api/agent/session", sessionRestoreHandler(deps.SessionStore))
	mux.HandleFunc("GET /api/agent/greeting", rl.Wrap(30.0/60, 5, greetingHandler(deps.ChatAgent)))

	// Stateless computation
	registerCoreRoutes(mux, rl)
}

func registerCoreRoutes(mux *http.ServeMux, rl *RateLimiter) {
	core := func(h http.HandlerFunc) http.HandlerFunc { return rl.Wrap(60.0/60, 10, h) }

	// Compute endpoints — 60/min per IP
	mux.HandleFunc("GET /api/huangli/query", core(queryHuangli))
	mux.HandleFunc("POST /api/huangli/bond", core(bondHuangli))

	mux.HandleFunc("GET /api/fengshui/sanyuan", core(getSanYuan))
	mux.HandleFunc("POST /api/fengshui/minggua", core(mingGua))
	mux.HandleFunc("POST /api/fengshui/chart", core(fengshuiChart))
	mux.HandleFunc("POST /api/xuankong/chart", core(xuankongChart))

	mux.HandleFunc("POST /api/bazi/chart", core(computeChart))
	mux.HandleFunc("POST /api/tianwen/solartime", core(computeSolarTime))
	mux.HandleFunc("POST /api/bazi/bond", core(bondCharts))
	mux.HandleFunc("POST /api/bazi/liunian", core(liuNian))
	mux.HandleFunc("POST /api/bazi/liuyue", core(liuYue))
	mux.HandleFunc("POST /api/bazi/liuri", core(liuRi))
	mux.HandleFunc("POST /api/bazi/liushi", core(liuShi))
	mux.HandleFunc("POST /api/bazi/xiaoyun", core(xiaoYun))
	mux.HandleFunc("POST /api/bazi/xiaoxian", core(xiaoXian))

	mux.HandleFunc("POST /api/ziwei/chart", core(computeZiweiChart))
	mux.HandleFunc("POST /api/ziwei/daxian", core(computeZiweiDaxian))
	mux.HandleFunc("POST /api/ziwei/liunian", core(computeZiweiLiunian))

	mux.HandleFunc("POST /api/ziwei/liuyue", core(computeZiweiLiuyue))
	mux.HandleFunc("POST /api/ziwei/liuri", core(computeZiweiLiuri))
	mux.HandleFunc("POST /api/ziwei/bond", core(computeZiweiBond))

	mux.HandleFunc("POST /api/qiming/wuge", core(handleWuge))
	mux.HandleFunc("POST /api/qiming/compose", core(handleCompose))
	mux.HandleFunc("POST /api/qiming/detail", core(handleDetail))
	mux.HandleFunc("POST /api/qiming/evaluate", core(handleEvaluate))

	mux.HandleFunc("POST /api/liuyao/chart", core(handleLiuyaoChart))
	mux.HandleFunc("POST /api/qimen/pan", core(handleQimenPan))

}
