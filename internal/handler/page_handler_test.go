package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/taverns-red/tavern-url/internal/auth"
	"github.com/taverns-red/tavern-url/internal/service"
)

func setupPageHandler() (*PageHandler, *chi.Mux) {
	sessionStore := auth.NewSessionStore("test-secret-key-32-bytes-long!!")
	repo := newMockRepo()
	linkSvc := service.NewLinkService(repo)

	h := NewPageHandler(sessionStore, nil, linkSvc, nil, nil, nil, "http://localhost:8080")

	r := chi.NewRouter()
	r.Get("/", h.Home)
	r.Get("/login", h.Login)
	r.Get("/register", h.Register)
	r.Get("/dashboard", h.Dashboard)
	r.Get("/links/{slug}", h.LinkDetail)
	r.Get("/settings/keys", h.APIKeys)
	r.Get("/settings/org", h.Orgs)
	r.Get("/settings/domains", h.Domains)
	r.Get("/settings/webhooks", h.Webhooks)
	r.Get("/bundles", h.Bundles)
	r.Get("/notifications", h.Notifications)
	r.Get("/integrations", h.Integrations)
	r.Get("/admin", h.Admin)
	r.Get("/admin/applications", h.Applications)
	r.Get("/docs", h.Docs)
	r.Get("/apply", h.Apply)

	return h, r
}

// TestPublicPages_Render verifies that public pages (no auth required) return 200.
func TestPublicPages_Render(t *testing.T) {
	_, r := setupPageHandler()

	publicRoutes := []struct {
		name string
		path string
	}{
		{"home", "/"},
		{"login", "/login"},
		{"register", "/register"},
		{"docs", "/docs"},
		{"apply", "/apply"},
	}

	for _, tc := range publicRoutes {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("%s: expected 200, got %d", tc.name, w.Code)
			}
		})
	}
}

// TestProtectedPages_RedirectWhenUnauthenticated verifies that protected pages
// redirect to /login when the user is not authenticated.
func TestProtectedPages_RedirectWhenUnauthenticated(t *testing.T) {
	_, r := setupPageHandler()

	protectedRoutes := []struct {
		name string
		path string
	}{
		{"dashboard", "/dashboard"},
		{"settings_keys", "/settings/keys"},
		{"settings_org", "/settings/org"},
		{"settings_domains", "/settings/domains"},
		{"settings_webhooks", "/settings/webhooks"},
		{"bundles", "/bundles"},
		{"notifications", "/notifications"},
		{"integrations", "/integrations"},
		{"admin", "/admin"},
		{"admin_applications", "/admin/applications"},
	}

	for _, tc := range protectedRoutes {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusFound {
				t.Errorf("%s: expected 302 redirect, got %d", tc.name, w.Code)
			}
			location := w.Header().Get("Location")
			if location != "/login" {
				t.Errorf("%s: expected redirect to /login, got %q", tc.name, location)
			}
		})
	}
}

// TestHome_RedirectsWhenAuthenticated verifies that authenticated users are
// redirected from / to /dashboard.
func TestHome_RedirectsWhenAuthenticated(t *testing.T) {
	h, r := setupPageHandler()

	// Create a request and set session.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Set a session cookie by writing to a temp response first.
	tempW := httptest.NewRecorder()
	if err := h.sessionStore.SetUserID(tempW, req, 1); err != nil {
		t.Fatalf("failed to set session: %v", err)
	}

	// Copy cookies from temp response to the real request.
	for _, cookie := range tempW.Result().Cookies() {
		req.AddCookie(cookie)
	}

	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/dashboard" {
		t.Errorf("expected redirect to /dashboard, got %q", loc)
	}
}
