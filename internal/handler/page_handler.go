package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/taverns-red/tavern-url/internal/auth"
	"github.com/taverns-red/tavern-url/internal/service"
	"github.com/taverns-red/tavern-url/templates"
)

// PageHandler handles server-rendered page requests.
type PageHandler struct {
	sessionStore auth.SessionStore
	authSvc      *auth.Service
	linkSvc      *service.LinkService
	analyticsSvc *service.AnalyticsService
	baseURL      string
}

// NewPageHandler creates a new PageHandler.
func NewPageHandler(sessionStore auth.SessionStore, authSvc *auth.Service, linkSvc *service.LinkService, analyticsSvc *service.AnalyticsService, baseURL string) *PageHandler {
	return &PageHandler{sessionStore: sessionStore, authSvc: authSvc, linkSvc: linkSvc, analyticsSvc: analyticsSvc, baseURL: baseURL}
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

// Dashboard renders the main dashboard with link list.
func (h *PageHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	links, err := h.linkSvc.ListLinks(r.Context())
	if err != nil {
		http.Error(w, "failed to list links", http.StatusInternalServerError)
		return
	}

	templates.DashboardPage(links, h.baseURL).Render(r.Context(), w)
}

// LinkDetail renders the link detail page with analytics.
func (h *PageHandler) LinkDetail(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	slug := chi.URLParam(r, "slug")
	link, err := h.linkSvc.GetBySlug(r.Context(), slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	summary, err := h.analyticsSvc.GetSummary(r.Context(), link.ID, days)
	if err != nil {
		http.Error(w, "failed to load analytics", http.StatusInternalServerError)
		return
	}

	templates.LinkDetailPage(link, summary, h.baseURL, days).Render(r.Context(), w)
}

