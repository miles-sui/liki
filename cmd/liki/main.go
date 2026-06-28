package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"liki/internal/agent"
	"liki/internal/dodo"
	"liki/internal/product"
	"liki/internal/xunhu"

	"liki/internal/email"
	apphttp "liki/internal/http"
	"liki/internal/llm"
	"liki/internal/payment"
)

// BuildTime is set at compile time via -ldflags.
var BuildTime = "dev"

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dbPath := flag.String("db", "/var/lib/liki/liki.db", "SQLite database path")
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
	emailFrom := envOr("RESEND_FROM", envOr("EMAIL_FROM", "Liki <report@liki.email>"))
	emailClient := email.New(envOr("RESEND_API_KEY", ""), emailFrom)

	dodoTest := envOrBool("DODO_TEST_MODE", false)
	dodoProducts := map[product.Product]string{
		product.ProductNaming: os.Getenv("DODO_PRODUCT_NAMING"),
	}
	dodoClient := dodo.New(envOr("DODO_API_KEY", ""), envOr("DODO_WEBHOOK_KEY", ""), dodoTest, dodoProducts)

	xunhuClient := xunhu.New(
		envOr("XUNHU_APPID", ""),
		envOr("XUNHU_APPSECRET", ""),
	)

	// Validate tools.json at startup so config errors fail fast.
	if err := agent.ValidateTools(); err != nil {
		slog.Error("validate tools.json", "err", err)
		os.Exit(1)
	}

	// Naming chat tools.
	llmClient := llm.New(envOr("DEEPSEEK_API_KEY", ""))
	namingTools := agent.NewNamingToolRegistry()
	chatAgent := agent.NewChatAgent(llmClient, namingTools, agent.NamingPrompt)

	returnURL := envOr("RETURN_URL", "")
	if returnURL == "" {
		slog.Warn("RETURN_URL not set — post-payment redirect will fail")
	}

	adminEmail := envOr("ADMIN_EMAIL", "")

	// Context for background tasks and server BaseContext.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	paymentSvc := payment.NewService(dodoClient, xunhuClient, emailClient, store, returnURL, adminEmail, ctx)

	// Validate JWT_SECRET at startup so config errors fail fast.
	if err := apphttp.ValidateJWTSecret(); err != nil {
		slog.Error("validate JWT_SECRET", "err", err)
		os.Exit(1)
	}

	devMode := envOrBool("DEV_MODE", false)

	// Setup HTTP
	rateLimiter := apphttp.NewRateLimiter()
	defer rateLimiter.Stop()

	mux := http.NewServeMux()
	h := apphttp.RegisterRoutes(mux, apphttp.ServerDeps{
		Payment:   paymentSvc,
		Store:     store,
		ChatAgent: chatAgent,
		Analytics: apphttp.NewAnalytics(),
		DevMode:   devMode,
	}, BuildTime, rateLimiter)

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

	slog.Info("liki listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
	slog.Info("server stopped")

	// Phased shutdown: drain emails within a deadline so we don't block forever.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := paymentSvc.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown payment service", "err", err)
	}
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
