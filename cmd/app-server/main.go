package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/25types/25types/internal/app/application/commerce"
	"github.com/25types/25types/internal/app/application/reports"
	"github.com/25types/25types/internal/app/application/user"
	"github.com/25types/25types/internal/app/db"
	httpx "github.com/25types/25types/internal/app/http"
	"github.com/25types/25types/internal/app/infra/dodo"
	"github.com/25types/25types/internal/app/infra/llm"
	"github.com/25types/25types/internal/app/infra/resend"
	"github.com/25types/25types/internal/app/infra/tencent"
	"github.com/25types/25types/internal/app/sqlite"
	"github.com/25types/25types/internal/tianwen"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dbPath := flag.String("db", "data/25types.db", "path to SQLite database")
	flag.Parse()

	if v := os.Getenv("LISTEN_ADDR"); v != "" {
		*addr = v
	}
	if v := os.Getenv("DATABASE_PATH"); v != "" {
		*dbPath = v
	}
	database, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()
	log.Printf("database opened at %s", *dbPath)

	nc := len(tianwen.LoadedCities())
	log.Printf("%d cities loaded", nc)
	zhCount := 0
	for _, c := range tianwen.LoadedCities() {
		if c.NameZh != "" {
			zhCount++
		}
	}
	if zhCount > 0 {
		log.Printf("%d Chinese city name mappings loaded", zhCount)
	}

	userRepo := sqlite.NewUserRepo(database)
	assRepo := sqlite.NewAssessmentRepo(database)
	reviewLinkRepo := sqlite.NewReviewLinkRepo(database)

	from := os.Getenv("EMAIL_FROM")
	var userSender user.EmailSender
	if from == "" {
		log.Println("[email] EMAIL_FROM not set — transactional emails disabled")
	} else {
		switch os.Getenv("EMAIL_PROVIDER") {
		case "tencent":
			id := os.Getenv("TENCENT_SECRET_ID")
			key := os.Getenv("TENCENT_SECRET_KEY")
			region := os.Getenv("TENCENT_REGION")
			tencentFrom := os.Getenv("TENCENT_FROM")
			if tencentFrom == "" {
				tencentFrom = from
			}
			if id != "" && key != "" {
				c := tencent.New(id, key, tencentFrom, region)
				userSender = c
				log.Printf("[email] Tencent SES client initialized (region=%s)", region)
			} else {
				log.Println("[email] TENCENT_SECRET_ID/KEY not set — disabling email")
			}
		default:
			if key := os.Getenv("RESEND_API_KEY"); key != "" {
				c := resend.New(key, from)
				userSender = c
				log.Println("[email] Resend client initialized")
			} else {
				log.Println("[email] no provider configured — transactional emails disabled")
			}
		}
	}

	profileRepo := sqlite.NewProfileRepo(userRepo, assRepo)
	matchLinkRepo := sqlite.NewMatchLinkRepo(database)

	var dodoClient *dodo.Client
	if apiKey := os.Getenv("DODO_API_KEY"); apiKey != "" {
		webhookKey := os.Getenv("DODO_WEBHOOK_KEY")
		testMode := os.Getenv("DODO_TEST_MODE") == "1"
		dodoClient = dodo.New(apiKey, webhookKey, testMode)
		log.Printf("[dodo] client initialized (test=%v)", testMode)
	} else {
		log.Println("[dodo] DODO_API_KEY not set — donation checkout disabled")
	}

	donationRepo := sqlite.NewDonationRepo(database)

	// LLM + Reports
	var reportsService *reports.Service
	llmConfig := llm.LoadModelConfig("")
	timeout := llm.ParseTimeout(llmConfig.Timeout, 15*time.Second)

	templates := llm.MustLoadTemplates()
	log.Printf("llm: templates loaded")

	llmClient := llm.New(llmConfig.Model, llmConfig.MaxTokens, timeout)
	reportsRepo := sqlite.NewReportsRepo(database)
	reportsService = reports.NewService(reportsRepo, llmClient, templates)
	log.Printf("reports: service initialized (model=%s max_tokens=%d timeout=%s)", llmConfig.Model, llmConfig.MaxTokens, llmConfig.Timeout)

	var thankYouSender commerce.ThankYouSender
	if userSender != nil {
		if ts, ok := userSender.(commerce.ThankYouSender); ok {
			thankYouSender = ts
		}
	}

	productID := os.Getenv("DODO_PRODUCT_DONATION")
	subProductID := os.Getenv("DODO_PRODUCT_SUBSCRIBE")

	mux := http.NewServeMux()
	jwtSecret := os.Getenv("JWT_SECRET")

	var adminUsers []string
	if s := os.Getenv("ADMIN_USERS"); s != "" {
		for _, u := range strings.Split(s, ",") {
			u = strings.TrimSpace(u)
			if u != "" {
				adminUsers = append(adminUsers, u)
			}
		}
	}

	httpx.RegisterRoutes(mux, httpx.ServerConfig{
		UserRepo:   userRepo,
		UserHasher: sqlite.PasswordHasher{},

		AssRepo:  assRepo,
		LinkRepo: reviewLinkRepo,
		SubRepo:  reviewLinkRepo,

		Profiles:      assRepo,
		ProfileRepo:   profileRepo,
		ProfileUsers:  profileRepo,
		BondStore:     profileRepo,
		MatchLinkRepo: matchLinkRepo,

		UserEmailSender: userSender,
		TokenValidator:  userRepo,
		UserLookup:      userRepo,
		ExportRepo:      userRepo,
		DB:              database,
		ReportsService: reportsService,

		DodoClient:     dodoClient,
		DonationRepo:   donationRepo,
		ThankYouSender:   thankYouSender,
		DodoProductID:    productID,
		DodoSubProductID: subProductID,
		UserEmailFn: func(ctx context.Context, userID int64) (string, bool) {
			u, err := userRepo.FindByID(ctx, userID)
			if err != nil || u == nil {
				return "", false
			}
			if u.Email != "" {
				return u.Email, true
			}
			if u.PendingEmail != nil && *u.PendingEmail != "" {
				return *u.PendingEmail, true
			}
			return "", false
		},

		JWTSecret:  jwtSecret,
		AdminUsers: adminUsers,
	})

	srv := &http.Server{
		Addr:         *addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-quit
		log.Printf("received %s, shutting down gracefully...", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("server forced to shutdown: %v", err)
		}
	}()

	log.Printf("25types API server listening on %s", *addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
	log.Println("server stopped")
}
