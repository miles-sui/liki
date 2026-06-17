package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	doc "liki"
	"liki/internal/agent"
	"liki/internal/dodo"
	
	"liki/internal/email"
	"liki/internal/http"
	"liki/internal/llm"
	"liki/internal/payment"
	"liki/internal/session"
)

// BuildTime is set at compile time via -ldflags.
var BuildTime = "dev"

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dbPath := flag.String("db", "/var/lib/lingji/lingji.db", "SQLite database path")
	flag.Parse()

	// Structured logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	// Init database
	db, err := payment.OpenDB(envOr("DB_PATH", *dbPath))
	if err != nil {
		slog.Error("open db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	store, err := payment.NewStore(db)
	if err != nil {
		slog.Error("init store", "err", err)
		os.Exit(1)
	}

	// Init services
	emailFrom := envOr("RESEND_FROM", envOr("EMAIL_FROM", "Liki <report@lingji.email>"))
	emailClient := email.New(envOr("RESEND_API_KEY", ""), emailFrom)

	dodoTest := envOrBool("DODO_TEST_MODE", false)
	dodoClient := dodo.New(envOr("DODO_API_KEY", ""), envOr("DODO_WEBHOOK_KEY", ""), dodoTest)

	// Chat tools call foundation packages (bazi, qiming) directly.
	chatTools := agent.NewChatToolRegistry()
	chatAgent := agent.NewChatAgent(llm.New(envOr("DEEPSEEK_API_KEY", "")), chatTools, "system prompt")
	// Single agent with unified prompt and all tools.
	chatAgent.ReportPrompts = map[agent.Product]string{
		agent.ProductChart:  doc.ChartReportPrompt,
		agent.ProductBond:   doc.BondReportPrompt,
		agent.ProductNaming: doc.NamingReportPrompt,
	}
	chatAgent.Greeting = "你好，我是灵机（Liki），一款 AI 命理助手，为你提供命盘解读、运势分析、合盘配对及起名等服务。有什么事，不妨一起看看。"

	sessionStore := session.NewStore(30*time.Minute, 500)
	defer sessionStore.Stop()

	productIDs := map[agent.Product]string{
		agent.ProductChart:  os.Getenv("DODO_PRODUCT_CHART"),
		agent.ProductBond:   os.Getenv("DODO_PRODUCT_BOND"),
		agent.ProductNaming: os.Getenv("DODO_PRODUCT_NAMING"),
	}

	returnURL := envOr("RETURN_URL", "")
	if returnURL == "" {
		slog.Warn("RETURN_URL not set — post-payment redirect will fail")
	}

	adminEmail := envOr("ADMIN_EMAIL", "")

	// Context for background tasks and server BaseContext.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reportAdapter := &reportGenerator{agent: chatAgent}
	paymentSvc := payment.NewService(dodoClient, emailClient, store, productIDs, returnURL, adminEmail, reportAdapter, ctx)
	chatAgent.Amounts = map[agent.Product]int{agent.ProductChart: 990, agent.ProductBond: 1990, agent.ProductNaming: 2990}

	// Dev mode controls CORS localhost origins.
	if envOrBool("DEV_MODE", false) {
		handler.SetDevMode(true)
	}

	// Setup HTTP
	rateLimiter := handler.NewRateLimiter()
	defer rateLimiter.Stop()

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, handler.ServerDeps{
		Payment:      paymentSvc,
		Store:        store,
		ChatAgent:    chatAgent,
		SessionStore: sessionStore,
		Analytics:    handler.NewAnalytics(),
	}, BuildTime, rateLimiter)

	h := handler.SecurityHeaders(handler.CORSMiddleware(handler.BodyLimit(mux)))

	// Write timeout must accommodate LLM calls (~120s)
	srv := &http.Server{
		Addr:         envOr("LISTEN_ADDR", *addr),
		Handler:      h,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 150 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Background stale order cleanup
	go cleanupStaleOrders(ctx, store)

	// Signal handling
	go handleSignals(srv)

	slog.Info("lingji listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envOrBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			slog.Warn("invalid bool env", "key", key, "value", v, "default", def)
			return def
		}
		return b
	}
	return def
}

func cleanupStaleOrders(ctx context.Context, store *payment.Store) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic in cleanupStaleOrders", "panic", r)
		}
	}()
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := store.CleanStale(ctx, 24*time.Hour); err != nil {
				slog.Error("clean stale orders", "err", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

type reportGenerator struct {
	agent *agent.ChatAgent
}

func (a *reportGenerator) GenerateFromData(ctx context.Context, locale string, product agent.Product, chartJSON json.RawMessage) (string, error) {
	return a.agent.GenerateFromData(ctx, locale, product, chartJSON, nil)
}

func handleSignals(srv *http.Server) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic in handleSignals", "panic", r)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("received signal, shutting down", "signal", sig.String())
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("forced shutdown", "err", err)
	}
}
