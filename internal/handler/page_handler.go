package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/taverns-red/tavern-url/internal/auth"
	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/service"
	"github.com/taverns-red/tavern-url/templates"
)

// PageHandler handles server-rendered page requests.
type PageHandler struct {
	sessionStore auth.SessionStore
	authSvc      *auth.Service
	linkSvc      *service.LinkService
	analyticsSvc *service.AnalyticsService
	apiKeySvc    *service.APIKeyService
	orgSvc       *service.OrgService
	baseURL      string
}

// NewPageHandler creates a new PageHandler.
func NewPageHandler(sessionStore auth.SessionStore, authSvc *auth.Service, linkSvc *service.LinkService, analyticsSvc *service.AnalyticsService, apiKeySvc *service.APIKeyService, orgSvc *service.OrgService, baseURL string) *PageHandler {
	return &PageHandler{sessionStore: sessionStore, authSvc: authSvc, linkSvc: linkSvc, analyticsSvc: analyticsSvc, apiKeySvc: apiKeySvc, orgSvc: orgSvc, baseURL: baseURL}
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

	// Filter by search query if provided.
	query := r.URL.Query().Get("q")
	if query != "" {
		var filtered []model.Link
		q := strings.ToLower(query)
		for _, l := range links {
			if strings.Contains(strings.ToLower(l.Slug), q) || strings.Contains(strings.ToLower(l.OriginalURL), q) {
				filtered = append(filtered, l)
			}
		}
		links = filtered
	}

	// HTMX request — return just the link list partial.
	if r.Header.Get("HX-Request") == "true" {
		templates.LinkList(links, h.baseURL).Render(r.Context(), w)
		return
	}

	templates.DashboardPage(links, h.baseURL, query).Render(r.Context(), w)
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

	templates.LinkDetailPage(link, summary, h.baseURL, days, nil).Render(r.Context(), w)
}

// APIKeys renders the API key management page.
func (h *PageHandler) APIKeys(w http.ResponseWriter, r *http.Request) {
	userID, err := h.sessionStore.GetUserID(r)
	if err != nil || userID == 0 {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	keys, err := h.apiKeySvc.ListKeys(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to list keys", http.StatusInternalServerError)
		return
	}

	templates.APIKeysPage(keys).Render(r.Context(), w)
}

// Orgs renders the organization management page.
func (h *PageHandler) Orgs(w http.ResponseWriter, r *http.Request) {
	userID, err := h.sessionStore.GetUserID(r)
	if err != nil || userID == 0 {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	orgs, err := h.orgSvc.ListUserOrgs(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to list organizations", http.StatusInternalServerError)
		return
	}

	templates.OrgPage(orgs).Render(r.Context(), w)
}

// Domains renders the custom domain management page.
func (h *PageHandler) Domains(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	// Pass empty domains list — domain CRUD repo TBD.
	templates.DomainsPage(nil).Render(r.Context(), w)
}

// Bundles renders the link bundles (bio page) management page.
func (h *PageHandler) Bundles(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	links, _ := h.linkSvc.ListLinks(r.Context())
	templates.BundlesPage(links, h.baseURL).Render(r.Context(), w)
}

// Notifications renders the notification center page.
func (h *PageHandler) Notifications(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	templates.NotificationsPage().Render(r.Context(), w)
}

// Webhooks renders the webhook management page.
func (h *PageHandler) Webhooks(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	templates.WebhooksPage(nil).Render(r.Context(), w)
}

// Integrations renders the integrations & tools page.
func (h *PageHandler) Integrations(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	templates.IntegrationsPage().Render(r.Context(), w)
}
