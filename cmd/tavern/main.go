package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/taverns-red/tavern-url/internal/auth"
	"github.com/taverns-red/tavern-url/internal/handler"
	"github.com/taverns-red/tavern-url/internal/repository"
	"github.com/taverns-red/tavern-url/internal/service"
)

func main() {
	// Load config from environment.
	port := envOrDefault("PORT", "8080")
	baseURL := envOrDefault("BASE_URL", "http://localhost:"+port)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	sessionSecret := envOrDefault("SESSION_SECRET", "dev-secret-change-me-in-production!!")

	// Connect to Postgres.
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("connected to database")

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
	sessionStore := auth.NewSessionStore(sessionSecret)

	// Wire handlers.
	linkHandler := handler.NewLinkHandler(linkSvc, analyticsSvc, baseURL)
	authHandler := handler.NewAuthHandler(authSvc, sessionStore)
	orgHandler := handler.NewOrgHandler(orgSvc)
	qrSvc := service.NewQRService()
	analyticsHandler := handler.NewAnalyticsHandler(analyticsSvc, qrSvc, linkSvc, baseURL)
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
		log.Println("Google OAuth enabled")
	} else {
		log.Println("Google OAuth disabled (GOOGLE_CLIENT_ID/GOOGLE_CLIENT_SECRET not set)")
	}

	// Page handler for server-rendered pages.
	pageHandler := handler.NewPageHandler(sessionStore, authSvc, linkSvc, analyticsSvc, baseURL)

	// Set up router.
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)

	// Static files.
	r.Handle("/static/*", handler.StaticFileServer("static"))

	// Page routes (server-rendered HTML).
	r.Get("/", pageHandler.Home)
	r.Get("/login", pageHandler.Login)
	r.Get("/register", pageHandler.Register)
	r.Get("/dashboard", pageHandler.Dashboard)
	r.Get("/links/{slug}", pageHandler.LinkDetail)

	// API routes.
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (public).
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/logout", authHandler.Logout)

		// Protected routes (session or API key).
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuthOrAPIKey(sessionStore, authSvc, apiKeySvc))
			r.Get("/auth/me", authHandler.Me)
			r.Post("/links", linkHandler.Create)
			r.Get("/links", linkHandler.List)
			r.Delete("/links/{id}", linkHandler.Delete)
			r.Get("/links/{id}/analytics", analyticsHandler.GetSummary)
			r.Get("/links/{id}/qr", analyticsHandler.QRCode)
			r.Post("/orgs", orgHandler.Create)
			r.Get("/orgs", orgHandler.List)
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

	// Start server with graceful shutdown.
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("tavern-url listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	fmt.Println("server stopped cleanly")
}

func envOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
