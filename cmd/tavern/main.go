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

	// Verify connection.
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("connected to database")

	// Wire layers.
	linkRepo := repository.NewPgLinkRepository(pool)
	userRepo := repository.NewPgUserRepository(pool)
	linkSvc := service.NewLinkService(linkRepo)
	authSvc := auth.NewService(userRepo)
	sessionStore := auth.NewSessionStore(sessionSecret)

	linkHandler := handler.NewLinkHandler(linkSvc, baseURL)
	authHandler := handler.NewAuthHandler(authSvc, sessionStore)

	// Set up router.
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)

	// API routes.
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (public).
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/logout", authHandler.Logout)

		// Protected auth routes.
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth(sessionStore, authSvc))
			r.Get("/auth/me", authHandler.Me)
			r.Post("/links", linkHandler.Create)
		})
	})

	// Health check.
	r.Get("/health", handler.Health)

	// Redirect — must be last to avoid catching API routes.
	r.Get("/{slug}", linkHandler.Redirect)

	// Start server with graceful shutdown.
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Run server in a goroutine.
	go func() {
		log.Printf("tavern-url listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	// Graceful shutdown with 30s timeout.
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
