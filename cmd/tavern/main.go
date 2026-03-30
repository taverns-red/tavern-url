package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/taverns-red/tavern-url/internal/auth"
	"github.com/taverns-red/tavern-url/internal/handler"
	tavmiddleware "github.com/taverns-red/tavern-url/internal/middleware"
	"github.com/taverns-red/tavern-url/internal/repository"
	"github.com/taverns-red/tavern-url/internal/service"
	"github.com/taverns-red/tavern-url/templates"
)

func main() {
	// Set up structured logging.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load config from environment.
	port := envOrDefault("PORT", "8080")
	baseURL := envOrDefault("BASE_URL", "http://localhost:"+port)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		slog.Error("SESSION_SECRET environment variable is required")
		os.Exit(1)
	}

	cookieSecure := envOrDefault("COOKIE_SECURE", "true") == "true"

	// Connect to Postgres.
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to database")

	// Wire repositories.
	linkRepo := repository.NewPgLinkRepository(pool)
	userRepo := repository.NewPgUserRepository(pool)
	orgRepo := repository.NewPgOrgRepository(pool)

	// Wire services.
	linkSvc := service.NewLinkService(linkRepo)
	authSvc := auth.NewService(userRepo)
	orgSvc := service.NewOrgService(orgRepo)
	clickRepo := repository.NewPgClickRepository(pool)
	analyticsSvc := service.NewAnalyticsService(clickRepo)
	apiKeyRepo := repository.NewPgAPIKeyRepository(pool)
	apiKeySvc := service.NewAPIKeyService(apiKeyRepo)
	sessionStore := auth.NewSessionStore(sessionSecret, cookieSecure)

	// Wire handlers.
	linkHandler := handler.NewLinkHandler(linkSvc, analyticsSvc, baseURL)
	authHandler := handler.NewAuthHandler(authSvc, sessionStore)
	orgHandler := handler.NewOrgHandler(orgSvc)
	qrSvc := service.NewQRService()
	analyticsHandler := handler.NewAnalyticsHandler(analyticsSvc, qrSvc, linkSvc, baseURL)
	exportHandler := handler.NewExportHandler(analyticsSvc)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeySvc)

	// Google OAuth (optional — only if credentials are configured).
	var googleHandler *handler.GoogleLoginHandler
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if googleClientID != "" && googleClientSecret != "" {
		googleProvider := auth.NewGoogleProvider(auth.GoogleOAuthConfig{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  baseURL + "/auth/google/callback",
		}, userRepo)
		googleHandler = handler.NewGoogleLoginHandler(googleProvider, sessionStore)
		slog.Info("Google OAuth enabled")
	} else {
		slog.Info("Google OAuth disabled", "reason", "GOOGLE_CLIENT_ID/GOOGLE_CLIENT_SECRET not set")
	}

	// Page handler for server-rendered pages.
	pageHandler := handler.NewPageHandler(sessionStore, authSvc, linkSvc, analyticsSvc, apiKeySvc, orgSvc, baseURL)

	// Rate limiter (60 req/min per IP for protected routes).
	rateLimiter := tavmiddleware.NewRateLimiter(60, time.Minute)

	// Set up router.
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(tavmiddleware.SecurityHeaders)

	// Static files.
	r.Handle("/static/*", handler.StaticFileServer("static", nil))

	// Page routes (server-rendered HTML).
	r.Get("/", pageHandler.Home)
	r.Get("/login", pageHandler.Login)
	r.Get("/register", pageHandler.Register)
	r.Get("/dashboard", pageHandler.Dashboard)
	r.Get("/links/{slug}", pageHandler.LinkDetail)
	r.Get("/settings/keys", pageHandler.APIKeys)
	r.Get("/settings/org", pageHandler.Orgs)
	r.Get("/settings/domains", pageHandler.Domains)
	r.Get("/bundles", pageHandler.Bundles)
	r.Get("/notifications", pageHandler.Notifications)
	r.Get("/settings/webhooks", pageHandler.Webhooks)
	r.Get("/integrations", pageHandler.Integrations)
	r.Get("/admin", pageHandler.Admin)
	r.Get("/admin/applications", pageHandler.Applications)
	r.Get("/docs", pageHandler.Docs)
	r.Get("/apply", pageHandler.Apply)

	// API routes.
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (public).
		r.With(rateLimiter.Middleware(tavmiddleware.ByIP)).Post("/auth/register", authHandler.Register)
		r.With(rateLimiter.Middleware(tavmiddleware.ByIP)).Post("/auth/login", authHandler.Login)
		r.Post("/auth/logout", authHandler.Logout)

		// Protected routes (session or API key).
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuthOrAPIKey(sessionStore, authSvc, apiKeySvc))
			r.Use(rateLimiter.Middleware(tavmiddleware.ByIP))
			r.Get("/auth/me", authHandler.Me)
			r.Post("/links", linkHandler.Create)
			r.Post("/links/bulk", linkHandler.BulkCreate)
			r.Get("/links", linkHandler.List)
			r.Put("/links/{id}", linkHandler.Update)
			r.Delete("/links/{id}", linkHandler.Delete)
			r.Get("/links/{id}/analytics", analyticsHandler.GetSummary)
			r.Get("/links/{id}/analytics/export", exportHandler.ExportCSV)
			r.Get("/links/{id}/qr", analyticsHandler.QRCode)
			r.Post("/orgs", orgHandler.Create)
			r.Get("/orgs", orgHandler.List)
			r.Post("/orgs/{slug}/invite", orgHandler.Invite)
			r.Put("/orgs/{slug}/members/{memberID}/role", orgHandler.UpdateRole)
			r.Post("/keys", apiKeyHandler.Create)
			r.Get("/keys", apiKeyHandler.List)
			r.Delete("/keys/{id}", apiKeyHandler.Delete)
		})
	})

	// Google OAuth routes (if enabled).
	if googleHandler != nil {
		r.Get("/auth/google/login", googleHandler.Login)
		r.Get("/auth/google/callback", googleHandler.Callback)
	}

	// Health check.
	r.Get("/health", handler.Health)

	// Redirect — must be last to avoid catching page routes.
	r.Get("/{slug}", linkHandler.Redirect)
	r.Post("/{slug}", linkHandler.Redirect)

	// CSRF protection setup.
	csrfKey := sha256.Sum256([]byte(sessionSecret))
	
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		slog.Error("invalid BASE_URL", "error", err)
		os.Exit(1)
	}

	csrfMw := csrf.Protect(
		csrfKey[:],
		csrf.Secure(cookieSecure),
		csrf.Path("/"),
		csrf.TrustedOrigins([]string{parsedBaseURL.Host}),
	)

	// Middleware to inject CSRF token into context for Templ.
	csrfContextMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			token := csrf.Token(req)
			ctx := context.WithValue(req.Context(), templates.CSRFContextKey, token)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}

	// Middleware to skip CSRF for Bearer token requests.
	skipCSRFForAPI := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if strings.HasPrefix(req.Header.Get("Authorization"), "Bearer ") {
				req = csrf.UnsafeSkipCheck(req)
			}
			next.ServeHTTP(w, req)
		})
	}

	// Wrap the global router.
	protectedHandler := csrfMw(csrfContextMw(r))
	finalHandler := skipCSRFForAPI(protectedHandler)

	// Start server with graceful shutdown.
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      finalHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("tavern-url listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}
	fmt.Println("server stopped cleanly")
}

func envOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
