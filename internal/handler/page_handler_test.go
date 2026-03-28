package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/go-chi/chi/v5"
	"github.com/taverns-red/tavern-url/internal/auth"
	"github.com/taverns-red/tavern-url/internal/service"
)

func setupPageHandler() (*PageHandler, *chi.Mux) {
	sessionStore := auth.NewSessionStore("test-secret-key-32-bytes-long!!", false)
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

// setAuthCookie is a test helper that sets a session cookie on the request.
func setAuthCookie(t *testing.T, h *PageHandler, req *http.Request, userID int64) {
	t.Helper()
	tempW := httptest.NewRecorder()
	if err := h.sessionStore.SetUserID(tempW, req, userID); err != nil {
		t.Fatalf("failed to set session: %v", err)
	}
	for _, cookie := range tempW.Result().Cookies() {
		req.AddCookie(cookie)
	}
}

func TestDashboard_Authenticated(t *testing.T) {
	h, r := setupPageHandler()

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	setAuthCookie(t, h, req, 1)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestDashboard_WithSearchQuery(t *testing.T) {
	h, r := setupPageHandler()

	req := httptest.NewRequest(http.MethodGet, "/dashboard?q=test", nil)
	setAuthCookie(t, h, req, 1)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestDashboard_HTMX(t *testing.T) {
	h, r := setupPageHandler()

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	setAuthCookie(t, h, req, 1)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLogin_RedirectsWhenAuthenticated(t *testing.T) {
	h, r := setupPageHandler()

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	setAuthCookie(t, h, req, 1)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/dashboard" {
		t.Errorf("expected redirect to /dashboard, got %q", loc)
	}
}

func TestRegister_RedirectsWhenAuthenticated(t *testing.T) {
	h, r := setupPageHandler()

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	setAuthCookie(t, h, req, 1)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/dashboard" {
		t.Errorf("expected redirect to /dashboard, got %q", loc)
	}
}

func TestLinkDetail_NotFound(t *testing.T) {
	h, r := setupPageHandler()

	req := httptest.NewRequest(http.MethodGet, "/links/nonexistent-slug", nil)
	setAuthCookie(t, h, req, 1)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestLinkDetail_Unauthenticated(t *testing.T) {
	_, r := setupPageHandler()

	req := httptest.NewRequest(http.MethodGet, "/links/some-slug", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302 redirect, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/login" {
		t.Errorf("expected redirect to /login, got %q", loc)
	}
}

func TestBundles_Authenticated(t *testing.T) {
	h, r := setupPageHandler()

	req := httptest.NewRequest(http.MethodGet, "/bundles", nil)
	setAuthCookie(t, h, req, 1)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// TestProtectedPages_AuthenticatedRender verifies that authenticated users
// see the actual pages instead of being redirected.
func TestProtectedPages_AuthenticatedRender(t *testing.T) {
	h, r := setupPageHandler()

	authenticatedRoutes := []struct {
		name string
		path string
	}{
		{"notifications", "/notifications"},
		{"integrations", "/integrations"},
		{"admin", "/admin"},
		{"admin_applications", "/admin/applications"},
		{"settings_domains", "/settings/domains"},
		{"settings_webhooks", "/settings/webhooks"},
	}

	for _, tc := range authenticatedRoutes {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			setAuthCookie(t, h, req, 1)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("%s: expected 200, got %d", tc.name, w.Code)
			}
		})
	}
}

func TestStaticFileServer_Disk(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/test.js", []byte("console.log('hi')"), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	handler := StaticFileServer(dir, nil)

	r := chi.NewRouter()
	r.Handle("/static/*", handler)

	req := httptest.NewRequest(http.MethodGet, "/static/test.js", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestStaticFileServer_Embedded(t *testing.T) {
	fsys := fstest.MapFS{
		"test.css": &fstest.MapFile{
			Data: []byte("body { color: red; }"),
		},
	}

	handler := StaticFileServer("", fsys)

	r := chi.NewRouter()
	r.Handle("/static/*", handler)

	req := httptest.NewRequest(http.MethodGet, "/static/test.css", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuthHandler_Register_WeakPassword_Form(t *testing.T) {
	_, r := setupAuthHandlerWithService()

	form := url.Values{}
	form.Set("email", "weakform@test.com")
	form.Set("name", "Weak")
	form.Set("password", "123")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for weak password form, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "weak") && !strings.Contains(w.Body.String(), "8 characters") {
		t.Errorf("expected weak password message, got %q", w.Body.String())
	}
}

func TestAuthHandler_Register_DuplicateEmail_Form(t *testing.T) {
	_, r := setupAuthHandlerWithService()

	// Register first via JSON.
	body := `{"email":"dup-form@test.com","name":"First","password":"StrongPass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Duplicate via form.
	form := url.Values{}
	form.Set("email", "dup-form@test.com")
	form.Set("name", "Second")
	form.Set("password", "StrongPass123!")
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 for dup email form, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthHandler_Login_WrongPassword_Form(t *testing.T) {
	_, r := setupAuthHandlerWithService()

	// Register first.
	body := `{"email":"form-pw@test.com","name":"FormPW","password":"StrongPass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Login via form with wrong password.
	form := url.Values{}
	form.Set("email", "form-pw@test.com")
	form.Set("password", "WrongPassword")
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong password form, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGoogleLoginHandler_Login(t *testing.T) {
	cfg := auth.GoogleOAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		RedirectURL:  "http://localhost/callback",
	}
	repo := newMockUserRepoForAuthHandler()
	provider := auth.NewGoogleProvider(cfg, repo)
	store := auth.NewSessionStore("test-secret-key-32-bytes-long!!!", false)
	h := NewGoogleLoginHandler(provider, store)

	r := chi.NewRouter()
	r.Get("/auth/google/login", h.Login)

	req := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected 307, got %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if !strings.Contains(loc, "test-client-id") {
		t.Errorf("expected redirect to Google with client ID, got %q", loc)
	}
}

func setupFullPageHandler() (*PageHandler, *chi.Mux) {
	sessionStore := auth.NewSessionStore("test-secret-key-32-bytes-long!!", false)
	repo := newMockRepo()
	linkSvc := service.NewLinkService(repo)
	orgRepo := newMockOrgRepo()
	orgSvc := service.NewOrgService(orgRepo)
	apiKeyRepo := newMockAPIKeyRepoH()
	apiKeySvc := service.NewAPIKeyService(apiKeyRepo)

	h := NewPageHandler(sessionStore, nil, linkSvc, nil, apiKeySvc, orgSvc, "http://localhost:8080")

	r := chi.NewRouter()
	r.Get("/settings/keys", h.APIKeys)
	r.Get("/settings/org", h.Orgs)

	return h, r
}

func TestAPIKeys_Authenticated(t *testing.T) {
	h, r := setupFullPageHandler()

	req := httptest.NewRequest(http.MethodGet, "/settings/keys", nil)
	setAuthCookie(t, h, req, 1)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestOrgs_Authenticated(t *testing.T) {
	h, r := setupFullPageHandler()

	req := httptest.NewRequest(http.MethodGet, "/settings/org", nil)
	setAuthCookie(t, h, req, 1)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuthHandler_Logout_ClearError(t *testing.T) {
	store := auth.NewSessionStore("test-secret-key-32-bytes-long!!!", false)
	h := &AuthHandler{authSvc: nil, sessionStore: store}

	r := chi.NewRouter()
	r.Post("/api/v1/auth/logout", h.Logout)

	// Just call logout without a session — Clear should succeed anyway
	// because gorilla/sessions creates a new empty session.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
