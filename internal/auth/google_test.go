package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/oauth2"

	"github.com/taverns-red/tavern-url/internal/model"
)

// contains checks if s contains substr (simple helper for test assertions).
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestNewGoogleProvider(t *testing.T) {
	cfg := GoogleOAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/auth/google/callback",
	}

	repo := newMockUserRepo()
	provider := NewGoogleProvider(cfg, repo)

	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
	if provider.config.ClientID != "test-client-id" {
		t.Errorf("expected client ID, got %q", provider.config.ClientID)
	}
}

func TestGoogleProvider_LoginURL(t *testing.T) {
	cfg := GoogleOAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/auth/google/callback",
	}

	repo := newMockUserRepo()
	provider := NewGoogleProvider(cfg, repo)

	url := provider.LoginURL("test-state")
	if url == "" {
		t.Fatal("expected non-empty login URL")
	}
	if !contains(url, "test-client-id") {
		t.Errorf("expected URL to contain client ID, got %q", url)
	}
	if !contains(url, "test-state") {
		t.Errorf("expected URL to contain state, got %q", url)
	}
}

// setupMockOAuthServer creates a mock HTTP server that serves both:
// - POST /token (OAuth token exchange)
// - GET /userinfo (Google userinfo endpoint)
func setupMockOAuthServer(email, name string) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"access_token":  "mock-access-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "mock-refresh-token",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := googleUserInfo{Email: email, Name: name}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	})

	return httptest.NewServer(mux)
}

func TestHandleCallback_NewUser(t *testing.T) {
	mockServer := setupMockOAuthServer("newuser@test.com", "New User")
	defer mockServer.Close()

	repo := newMockUserRepo()
	cfg := GoogleOAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/callback",
	}
	provider := NewGoogleProvider(cfg, repo)

	// Override the oauth2 config + userinfo URL to point at our mock server.
	provider.config.Endpoint = oauth2.Endpoint{
		TokenURL: mockServer.URL + "/token",
	}
	provider.userInfoURL = mockServer.URL + "/userinfo"

	user, err := provider.HandleCallback(t.Context(), "test-code")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "newuser@test.com" {
		t.Errorf("expected email newuser@test.com, got %q", user.Email)
	}
	if user.Name != "New User" {
		t.Errorf("expected name 'New User', got %q", user.Name)
	}
	if user.PasswordHash != "oauth:google" {
		t.Errorf("expected oauth:google sentinel, got %q", user.PasswordHash)
	}
}

func TestHandleCallback_ExistingUser(t *testing.T) {
	mockServer := setupMockOAuthServer("existing@test.com", "Existing User")
	defer mockServer.Close()

	repo := newMockUserRepo()
	// Pre-register the user so HandleCallback returns existing.
	_ = repo.Create(t.Context(), &model.User{
		Email:        "existing@test.com",
		Name:         "Existing User",
		PasswordHash: "some-hash",
	})

	cfg := GoogleOAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/callback",
	}
	provider := NewGoogleProvider(cfg, repo)
	provider.config.Endpoint = oauth2.Endpoint{
		TokenURL: mockServer.URL + "/token",
	}
	provider.userInfoURL = mockServer.URL + "/userinfo"

	user, err := provider.HandleCallback(t.Context(), "test-code")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "existing@test.com" {
		t.Errorf("expected email existing@test.com, got %q", user.Email)
	}
}

func TestHandleCallback_BadUserInfo(t *testing.T) {
	// Server that returns invalid JSON for userinfo.
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"access_token": "tok", "token_type": "Bearer", "expires_in": 3600,
		}
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "not valid json!!!")
	})
	mockServer := httptest.NewServer(mux)
	defer mockServer.Close()

	repo := newMockUserRepo()
	cfg := GoogleOAuthConfig{
		ClientID: "id", ClientSecret: "secret", RedirectURL: "http://localhost/callback",
	}
	provider := NewGoogleProvider(cfg, repo)
	provider.config.Endpoint = oauth2.Endpoint{TokenURL: mockServer.URL + "/token"}
	provider.userInfoURL = mockServer.URL + "/userinfo"

	_, err := provider.HandleCallback(t.Context(), "test-code")
	if err == nil {
		t.Error("expected error for bad userinfo JSON")
	}
	if !contains(err.Error(), "failed to parse user info") {
		t.Errorf("expected 'failed to parse user info' in error, got %q", err.Error())
	}
}

func TestHandleCallback_ExchangeError(t *testing.T) {
	repo := newMockUserRepo()
	cfg := GoogleOAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/callback",
	}
	provider := NewGoogleProvider(cfg, repo)

	// Error-returning token endpoint.
	errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error": "invalid_grant"}`)
	}))
	defer errServer.Close()

	provider.config.Endpoint = oauth2.Endpoint{
		TokenURL: errServer.URL + "/token",
	}

	_, err := provider.HandleCallback(t.Context(), "bad-code")
	if err == nil {
		t.Error("expected error for bad exchange")
	}
}

