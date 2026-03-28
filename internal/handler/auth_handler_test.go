package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/taverns-red/tavern-url/internal/auth"
	"github.com/taverns-red/tavern-url/internal/model"
)


func TestAuthHandler_Register_JSON_MissingFields(t *testing.T) {
	store := auth.NewSessionStore("test-secret-key-32-bytes-long!!!")
	// We can't easily construct an auth.Service from handler package
	// since it requires repository.UserRepository. Instead, test the
	// validation paths that don't require service interaction.
	// The handler checks for empty fields before calling the service.

	// Create a handler with nil service — it should reject before calling it.
	// Actually, we can't do that safely. Let's test form validation.
	h := &AuthHandler{authSvc: nil, sessionStore: store}

	r := chi.NewRouter()
	r.Post("/api/v1/auth/register", h.Register)

	// Test missing fields — JSON path.
	body := `{"email":"","name":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthHandler_Register_FormEncoded_MissingFields(t *testing.T) {
	store := auth.NewSessionStore("test-secret-key-32-bytes-long!!!")
	h := &AuthHandler{authSvc: nil, sessionStore: store}

	r := chi.NewRouter()
	r.Post("/api/v1/auth/register", h.Register)

	form := url.Values{}
	form.Set("email", "")
	form.Set("name", "")
	form.Set("password", "")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "required") {
		t.Errorf("expected 'required' in error, got %q", w.Body.String())
	}
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	store := auth.NewSessionStore("test-secret-key-32-bytes-long!!!")
	h := &AuthHandler{authSvc: nil, sessionStore: store}

	r := chi.NewRouter()
	r.Post("/api/v1/auth/register", h.Register)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAuthHandler_Login_JSON_MissingFields(t *testing.T) {
	store := auth.NewSessionStore("test-secret-key-32-bytes-long!!!")
	h := &AuthHandler{authSvc: nil, sessionStore: store}

	r := chi.NewRouter()
	r.Post("/api/v1/auth/login", h.Login)

	body := `{"email":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthHandler_Login_FormEncoded_MissingFields(t *testing.T) {
	store := auth.NewSessionStore("test-secret-key-32-bytes-long!!!")
	h := &AuthHandler{authSvc: nil, sessionStore: store}

	r := chi.NewRouter()
	r.Post("/api/v1/auth/login", h.Login)

	form := url.Values{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	store := auth.NewSessionStore("test-secret-key-32-bytes-long!!!")
	h := &AuthHandler{authSvc: nil, sessionStore: store}

	r := chi.NewRouter()
	r.Post("/api/v1/auth/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAuthHandler_Logout_JSON(t *testing.T) {
	store := auth.NewSessionStore("test-secret-key-32-bytes-long!!!")
	h := &AuthHandler{authSvc: nil, sessionStore: store}

	r := chi.NewRouter()
	r.Post("/api/v1/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["message"] != "logged out" {
		t.Errorf("expected 'logged out', got %q", resp["message"])
	}
}

func TestAuthHandler_Logout_HTMX(t *testing.T) {
	store := auth.NewSessionStore("test-secret-key-32-bytes-long!!!")
	h := &AuthHandler{authSvc: nil, sessionStore: store}

	r := chi.NewRouter()
	r.Post("/api/v1/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if redirect := w.Header().Get("HX-Redirect"); redirect != "/" {
		t.Errorf("expected HX-Redirect '/', got %q", redirect)
	}
}

func TestAuthHandler_Me_Unauthenticated(t *testing.T) {
	h := &AuthHandler{}

	r := chi.NewRouter()
	r.Get("/api/v1/auth/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthHandler_Me_Authenticated(t *testing.T) {
	h := &AuthHandler{}

	r := chi.NewRouter()
	r.Get("/api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		user := &model.User{ID: 42, Email: "test@test.com", Name: "Test User"}
		ctx := auth.ContextWithUser(r.Context(), user)
		h.Me(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp userResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.ID != 42 || resp.Email != "test@test.com" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestWriteFormError(t *testing.T) {
	w := httptest.NewRecorder()
	writeFormError(w, http.StatusBadRequest, "Test error")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("expected text/html, got %q", ct)
	}
	if !strings.Contains(w.Body.String(), "Test error") {
		t.Errorf("expected body to contain 'Test error', got %q", w.Body.String())
	}
}
