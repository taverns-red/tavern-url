package handler

import (
	"net/http"

	"github.com/taverns-red/tavern-url/internal/auth"
	"github.com/taverns-red/tavern-url/templates"
)

// PageHandler handles server-rendered page requests.
type PageHandler struct {
	sessionStore auth.SessionStore
	authSvc      *auth.Service
}

// NewPageHandler creates a new PageHandler.
func NewPageHandler(sessionStore auth.SessionStore, authSvc *auth.Service) *PageHandler {
	return &PageHandler{sessionStore: sessionStore, authSvc: authSvc}
}

func (h *PageHandler) isAuthenticated(r *http.Request) bool {
	userID, err := h.sessionStore.GetUserID(r)
	return err == nil && userID > 0
}

// Home renders the landing page or redirects to dashboard.
func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	if h.isAuthenticated(r) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	templates.HomePage().Render(r.Context(), w)
}

// Login renders the login page.
func (h *PageHandler) Login(w http.ResponseWriter, r *http.Request) {
	if h.isAuthenticated(r) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	templates.LoginPage("").Render(r.Context(), w)
}

// Register renders the registration page.
func (h *PageHandler) Register(w http.ResponseWriter, r *http.Request) {
	if h.isAuthenticated(r) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	templates.RegisterPage("").Render(r.Context(), w)
}

// Dashboard renders the main dashboard (placeholder for Sprint 5).
func (h *PageHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	// For now, render a simple authenticated page. Sprint 5 will add the full dashboard.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<h1>Dashboard — coming in Sprint 5</h1>"))
}
