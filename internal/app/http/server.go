package http

import (
	"database/sql"
	"net/http"

	"github.com/25types/25types/internal/app/application/assessment"
	"github.com/25types/25types/internal/app/application/commerce"
	"github.com/25types/25types/internal/app/application/flow"
	"github.com/25types/25types/internal/app/application/matchlink"
	"github.com/25types/25types/internal/app/application/profile"
	"github.com/25types/25types/internal/app/application/reports"
	"github.com/25types/25types/internal/app/application/reviewlink"
	"github.com/25types/25types/internal/app/application/user"
	minglihttp "github.com/25types/25types/internal/mingli/http"
)

// ServerConfig holds all dependencies for the HTTP server.
type ServerConfig struct {
	// User / Auth
	UserRepo        user.UserRepository
	UserHasher      user.PasswordHasher
	UserEmailSender user.EmailSender
	TokenValidator  user.TokenValidator
	UserLookup      UserPublicLookup
	ExportRepo      user.ExportRepository
	JWTSecret       string
	AdminUsers      []string

	// Assessment
	AssRepo  assessment.AssessmentRepository
	LinkRepo reviewlink.ReviewLinkRepository
	SubRepo  reviewlink.ReviewSubmissionRepository

	// Flow & Profiles
	Profiles    flow.ProfileLoader
	ProfileRepo profile.ProfilePageRepo
	ProfileUsers profile.UserLookup
	BondStore   profile.BondStore

	// Match links
	MatchLinkRepo matchlink.MatchLinkRepository

	// Reports
	ReportsService *reports.Service

	// Payments
	DodoClient       commerce.PaymentProvider
	DonationRepo     commerce.DonationRepository
	ThankYouSender   commerce.ThankYouSender
	DodoProductID    string
	DodoSubProductID string
	UserEmailFn      UserEmailFn

	// DB
	DB *sql.DB
}

// RegisterRoutes registers all API routes on the given ServeMux.
func RegisterRoutes(mux *http.ServeMux, cfg ServerConfig) {
	if cfg.JWTSecret != "" {
		SetJWTSecret(cfg.JWTSecret)
	}
	SetAdminUsers(cfg.AdminUsers)

	// Middleware factories.
	limit := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
			}
			next(w, r)
		}
	}
	auth := func(next http.HandlerFunc) http.HandlerFunc {
		return RequireAuth(cfg.TokenValidator, limit(next))
	}
	opt := func(next http.HandlerFunc) http.HandlerFunc {
		return OptionalAuth(cfg.TokenValidator, limit(next))
	}

	// Handlers.
	uh := &UserHandler{
		Repo: cfg.UserRepo, Hasher: cfg.UserHasher,
		Claimer: cfg.AssRepo, Sender: cfg.UserEmailSender,
		ExportRepo: cfg.ExportRepo,
	}
	ah := &AssessmentHandler{Repo: cfg.AssRepo, UserLookup: cfg.UserLookup}
	rh := &ReviewHandler{LinkRepo: cfg.LinkRepo, SubRepo: cfg.SubRepo}
	fh := &FlowHandler{Profiles: cfg.Profiles}
	ph := &ProfileHandler{
		PageRepo: cfg.ProfileRepo, Users: cfg.ProfileUsers,
		Profiles: cfg.Profiles, Bonds: cfg.BondStore,
	}
	mh := &MatchLinkHandler{
		Repo: cfg.MatchLinkRepo, Assessments: cfg.AssRepo,
		ProfileLoader: cfg.Profiles, Bonds: cfg.BondStore,
		EmailSender: cfg.UserEmailSender, UserRepo: cfg.UserRepo,
	}
	bnh := &MingliMatchLinkHandler{Repo: cfg.MatchLinkRepo, MingliUsers: cfg.UserRepo}
	pmh := &PaymentHandler{
		DodoClient: cfg.DodoClient, DonationRepo: cfg.DonationRepo,
		ThankYouSender: cfg.ThankYouSender,
		ProductID: cfg.DodoProductID, SubProductID: cfg.DodoSubProductID,
		PlansData: defaultPlans(), UserEmailFn: cfg.UserEmailFn,
	}
	dlh := &DailyHandler{Users: cfg.UserRepo}

	registerHealthRoutes(mux, cfg.DB)
	registerAuthRoutes(mux, limit, auth, uh)
	registerUserRoutes(mux, auth, uh)
	registerAssessmentRoutes(mux, opt, auth, ah)
	registerReviewRoutes(mux, opt, auth, limit, rh)
	registerDebugRoutes(mux, cfg)
	minglihttp.RegisterCoreRoutes(mux)
	registerFlowRoutes(mux, auth, fh)
	registerProfileRoutes(mux, opt, auth, ph)
	registerPaymentRoutes(mux, auth, limit, pmh)
	registerMatchLinkRoutes(mux, auth, opt, mh)
	registerMingliMatchLinkRoutes(mux, auth, opt, bnh)
	registerReportRoutes(mux, auth, cfg.ReportsService)
	registerDailyRoutes(mux, opt, dlh)
	registerCareerRoutes(mux)

	mux.HandleFunc("POST /api/errors/frontend", limit(collectFrontendError(cfg.DB)))
}

func registerHealthRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "ok"
		if db != nil {
			if err := db.PingContext(r.Context()); err != nil {
				dbStatus = "error"
			}
		} else {
			dbStatus = "unavailable"
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok", "db": dbStatus})
	})

	const statsBase = 1000
	mux.HandleFunc("GET /api/stats", func(w http.ResponseWriter, r *http.Request) {
		type stats struct {
			TotalAssessments int `json:"total_assessments"`
		}
		var s stats
		if db != nil {
			db.QueryRowContext(r.Context(),
				`SELECT COUNT(*) FROM assessments WHERE assessment_type = 'self'`).Scan(&s.TotalAssessments)
		}
		s.TotalAssessments += statsBase
		respondJSON(w, http.StatusOK, s)
	})
}

func registerAuthRoutes(mux *http.ServeMux, limit, auth func(http.HandlerFunc) http.HandlerFunc, h *UserHandler) {
	mux.HandleFunc("POST /api/auth/register", limit(h.Register))
	mux.HandleFunc("POST /api/auth/login", limit(h.Login))
	mux.HandleFunc("POST /api/auth/logout", auth(h.Logout))
	mux.HandleFunc("PUT /api/auth/password", auth(h.ChangePassword))
	mux.HandleFunc("GET /api/auth/verify-email", h.VerifyEmail)
	mux.HandleFunc("POST /api/auth/resend-verification", auth(h.ResendVerification))
	mux.HandleFunc("POST /api/auth/forgot-password", limit(h.ForgotPassword))
	mux.HandleFunc("POST /api/auth/reset-password", limit(h.ResetPassword))
}

func registerUserRoutes(mux *http.ServeMux, auth func(http.HandlerFunc) http.HandlerFunc, h *UserHandler) {
	mux.HandleFunc("GET /api/users/me", auth(h.GetMe))
	mux.HandleFunc("PATCH /api/users/me", auth(h.UpdateMe))
	mux.HandleFunc("DELETE /api/users/me", auth(h.DeleteMe))
	mux.HandleFunc("GET /api/users/me/export", auth(h.ExportMe))
	mux.HandleFunc("GET /api/location", (&minglihttp.LocationHandler{}).GetLocation)
}

func registerAssessmentRoutes(mux *http.ServeMux, opt, auth func(http.HandlerFunc) http.HandlerFunc, h *AssessmentHandler) {
	mux.HandleFunc("POST /api/assessments", opt(h.SubmitAssessment))
	mux.HandleFunc("GET /api/assessments/questions", h.GetQuestions)
	mux.HandleFunc("GET /api/assessments", auth(h.ListAssessments))
	mux.HandleFunc("GET /api/assessments/{id}", opt(h.GetAssessment))
	mux.HandleFunc("GET /api/assessments/peers", auth(h.Peers))
	mux.HandleFunc("POST /api/assessments/claim", auth(h.Claim))
}

func registerReviewRoutes(mux *http.ServeMux, opt, auth, limit func(http.HandlerFunc) http.HandlerFunc, h *ReviewHandler) {
	mux.HandleFunc("POST /api/reviews", auth(h.CreateReview))
	mux.HandleFunc("GET /api/reviews", auth(h.ListReviews))
	mux.HandleFunc("GET /api/reviews/{id}", auth(h.GetReview))
	mux.HandleFunc("DELETE /api/reviews/{id}", auth(h.DeleteReview))
	mux.HandleFunc("POST /api/reviews/{id}/renew", auth(h.RenewReview))
	mux.HandleFunc("GET /api/reviews/given", opt(h.ListReviewsGiven))
	mux.HandleFunc("GET /api/r/{token}", h.GetLinkInfo)
	mux.HandleFunc("POST /api/r/{token}", limit(h.SubmitReview))
}

func registerDebugRoutes(mux *http.ServeMux, cfg ServerConfig) {
	mux.HandleFunc("GET /debug/db", RequireAdmin(cfg.TokenValidator, func(w http.ResponseWriter, r *http.Request) {
		type tableStat struct {
			Name string `json:"name"`
			Rows int    `json:"rows"`
		}
		type dbStatus struct {
			PageCount  int         `json:"page_count"`
			PageSize   int         `json:"page_size"`
			FileSizeKB int         `json:"file_size_kb"`
			Migrations int         `json:"migrations_applied"`
			Tables     []tableStat `json:"tables"`
		}
		var s dbStatus
		if cfg.DB == nil {
			respondJSON(w, http.StatusOK, dbStatus{Migrations: -1})
			return
		}
		cfg.DB.QueryRowContext(r.Context(), "PRAGMA page_count").Scan(&s.PageCount)
		cfg.DB.QueryRowContext(r.Context(), "PRAGMA page_size").Scan(&s.PageSize)
		s.FileSizeKB = s.PageCount * s.PageSize / 1024
		cfg.DB.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM schema_migrations").Scan(&s.Migrations)

		tables := []string{"users", "assessments", "review_links",
			"user_tokens", "frontend_errors", "match_links", "bond_events", "donations",
			"mingli_match_events"}
		for _, t := range tables {
			var n int
			cfg.DB.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM "+t).Scan(&n)
			s.Tables = append(s.Tables, tableStat{Name: t, Rows: n})
		}
		respondJSON(w, http.StatusOK, s)
	}))
}


func registerFlowRoutes(mux *http.ServeMux, auth func(http.HandlerFunc) http.HandlerFunc, h *FlowHandler) {
	mux.HandleFunc("GET /api/flow", auth(h.GetFlow))
	mux.HandleFunc("GET /api/flow/yearly", auth(h.GetFlowYearly))
	mux.HandleFunc("GET /api/solar-terms", h.GetSolarTerms)
}

func registerProfileRoutes(mux *http.ServeMux, opt, auth func(http.HandlerFunc) http.HandlerFunc, h *ProfileHandler) {
	mux.HandleFunc("GET /api/profiles/{name}", opt(h.GetProfile))
	mux.HandleFunc("GET /api/profiles/{name}/bonds", auth(h.GetBonds))
	mux.HandleFunc("POST /api/bond", auth(h.ComputeBond))
}

func registerPaymentRoutes(mux *http.ServeMux, auth, limit func(http.HandlerFunc) http.HandlerFunc, h *PaymentHandler) {
	mux.HandleFunc("GET /api/payments/plans", h.Plans)
	mux.HandleFunc("POST /api/payments/checkout", auth(h.Checkout))
	mux.HandleFunc("POST /api/payments/confirm", auth(h.Confirm))
	mux.HandleFunc("POST /api/payments/subscribe", auth(h.Subscribe))
	mux.HandleFunc("POST /api/payments/webhook", limit(h.Webhook))
}

func registerMatchLinkRoutes(mux *http.ServeMux, auth, opt func(http.HandlerFunc) http.HandlerFunc, h *MatchLinkHandler) {
	mux.HandleFunc("POST /api/match-links", auth(h.Create))
	mux.HandleFunc("GET /api/match-links", auth(h.List))
	mux.HandleFunc("DELETE /api/match-links/{id}", auth(h.Delete))
	mux.HandleFunc("GET /api/m/{token}", h.GetLinkInfo)
	mux.HandleFunc("POST /api/m/{token}", opt(h.SubmitAssessment))
}

func registerMingliMatchLinkRoutes(mux *http.ServeMux, auth, opt func(http.HandlerFunc) http.HandlerFunc, h *MingliMatchLinkHandler) {
	mux.HandleFunc("POST /api/mingli-match-links", auth(h.Create))
	mux.HandleFunc("GET /api/mingli-match-links", auth(h.List))
	mux.HandleFunc("DELETE /api/mingli-match-links/{id}", auth(h.Delete))
	mux.HandleFunc("GET /api/ml/{token}", h.GetLinkInfo)
	mux.HandleFunc("POST /api/ml/{token}", opt(h.Submit))
}

func registerReportRoutes(mux *http.ServeMux, auth func(http.HandlerFunc) http.HandlerFunc, svc *reports.Service) {
	if svc == nil {
		return
	}
	rph := NewReportsHandler(svc)
	mux.HandleFunc("POST /api/reports", auth(rph.Create))
	mux.HandleFunc("GET /api/reports", auth(rph.List))
	mux.HandleFunc("GET /api/reports/{id}", auth(rph.Get))
	mux.HandleFunc("DELETE /api/reports/{id}", auth(rph.Delete))
	mux.HandleFunc("POST /api/reports/{id}/share", auth(rph.CreateShare))
	mux.HandleFunc("GET /api/reports/shared/{token}", rph.GetShared)
}

func registerDailyRoutes(mux *http.ServeMux, opt func(http.HandlerFunc) http.HandlerFunc, h *DailyHandler) {
	mux.HandleFunc("GET /api/daily/suggestion", opt(h.Suggestion))
	mux.HandleFunc("GET /api/daily/question", opt(h.Question))
}

func registerCareerRoutes(mux *http.ServeMux) {
	ch := &CareerHandler{}
	mux.HandleFunc("GET /api/career/matches", ch.GetMatches)
	mux.HandleFunc("GET /api/career/types", ch.ListTypes)
}

