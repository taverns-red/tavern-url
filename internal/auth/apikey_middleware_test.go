package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taverns-red/tavern-url/internal/model"
)

func TestExtractBearerToken_Valid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer tvn_abc123")
	token := extractBearerToken(req)
	if token != "tvn_abc123" {
		t.Errorf("expected tvn_abc123, got %q", token)
	}
}

func TestExtractBearerToken_MissingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	token := extractBearerToken(req)
	if token != "" {
		t.Errorf("expected empty, got %q", token)
	}
}

func TestExtractBearerToken_WrongScheme(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	token := extractBearerToken(req)
	if token != "" {
		t.Errorf("expected empty for Basic auth, got %q", token)
	}
}

func TestExtractBearerToken_CaseInsensitive(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "bearer tvn_lower")
	token := extractBearerToken(req)
	if token != "tvn_lower" {
		t.Errorf("expected tvn_lower, got %q", token)
	}
}

func TestExtractBearerToken_EmptyToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer ")
	token := extractBearerToken(req)
	if token != "" {
		t.Errorf("expected empty after trim, got %q", token)
	}
}

func TestUserIDFromContext_WithUser(t *testing.T) {
	user := &model.User{ID: 42}
	ctx := ContextWithUser(context.Background(), user)
	id := UserIDFromContext(ctx)
	if id != 42 {
		t.Errorf("expected 42, got %d", id)
	}
}

func TestUserIDFromContext_NoUser(t *testing.T) {
	id := UserIDFromContext(context.Background())
	if id != 0 {
		t.Errorf("expected 0 for empty context, got %d", id)
	}
}

func TestUserFromContext_RoundTrip(t *testing.T) {
	user := &model.User{ID: 7, Email: "test@example.com"}
	ctx := ContextWithUser(context.Background(), user)
	got := UserFromContext(ctx)
	if got == nil {
		t.Fatal("expected user from context")
	}
	if got.ID != 7 || got.Email != "test@example.com" {
		t.Errorf("unexpected user: %+v", got)
	}
}

func TestUserFromContext_Nil(t *testing.T) {
	got := UserFromContext(context.Background())
	if got != nil {
		t.Errorf("expected nil for empty context, got %+v", got)
	}
}

func TestRequireAPIKey_MissingHeader(t *testing.T) {
	repo := newMockUserRepo()
	authSvc := NewService(repo)
	apiKeyRepo := newMockAPIKeyRepoForAuth()
	apiKeySvc := newAPIKeyServiceForAuth(apiKeyRepo)

	middleware := RequireAPIKey(apiKeySvc, authSvc)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without API key")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAPIKey_InvalidKey(t *testing.T) {
	repo := newMockUserRepo()
	authSvc := NewService(repo)
	apiKeyRepo := newMockAPIKeyRepoForAuth()
	apiKeySvc := newAPIKeyServiceForAuth(apiKeyRepo)

	middleware := RequireAPIKey(apiKeySvc, authSvc)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid key")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer tvn_invalid_key")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAPIKey_ValidKey(t *testing.T) {
	repo := newMockUserRepo()
	authSvc := NewService(repo)
	// Register a user first.
	user, err := authSvc.Register(context.Background(), "apikey@test.com", "API User", "password1234")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	apiKeyRepo := newMockAPIKeyRepoForAuth()
	apiKeySvc := newAPIKeyServiceForAuth(apiKeyRepo)
	rawKey, _, err := apiKeySvc.CreateKey(context.Background(), user.ID, 0, "Test Key")
	if err != nil {
		t.Fatalf("create key failed: %v", err)
	}

	var gotUserID int64
	middleware := RequireAPIKey(apiKeySvc, authSvc)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := UserFromContext(r.Context())
		if u != nil {
			gotUserID = u.ID
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+rawKey)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if gotUserID != user.ID {
		t.Errorf("expected user ID %d, got %d", user.ID, gotUserID)
	}
}

func TestRequireAuthOrAPIKey_SessionAuth(t *testing.T) {
	repo := newMockUserRepo()
	authSvc := NewService(repo)
	store := NewSessionStore("test-secret-key-32-bytes-long!!!")
	apiKeyRepo := newMockAPIKeyRepoForAuth()
	apiKeySvc := newAPIKeyServiceForAuth(apiKeyRepo)

	user, err := authSvc.Register(context.Background(), "dual@test.com", "Dual User", "password1234")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Set session.
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	if err := store.SetUserID(w, req, user.ID); err != nil {
		t.Fatalf("SetUserID failed: %v", err)
	}
	cookies := w.Result().Cookies()

	var gotUserID int64
	middleware := RequireAuthOrAPIKey(store, authSvc, apiKeySvc)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := UserFromContext(r.Context())
		if u != nil {
			gotUserID = u.ID
		}
		w.WriteHeader(http.StatusOK)
	}))

	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w2.Code)
	}
	if gotUserID != user.ID {
		t.Errorf("expected user ID %d, got %d", user.ID, gotUserID)
	}
}

func TestRequireAuthOrAPIKey_NoAuth(t *testing.T) {
	repo := newMockUserRepo()
	authSvc := NewService(repo)
	store := NewSessionStore("test-secret-key-32-bytes-long!!!")
	apiKeyRepo := newMockAPIKeyRepoForAuth()
	apiKeySvc := newAPIKeyServiceForAuth(apiKeyRepo)

	middleware := RequireAuthOrAPIKey(store, authSvc, apiKeySvc)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without auth")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuthOrAPIKey_APIKeyAuth(t *testing.T) {
	repo := newMockUserRepo()
	authSvc := NewService(repo)
	store := NewSessionStore("test-secret-key-32-bytes-long!!!")
	apiKeyRepo := newMockAPIKeyRepoForAuth()
	apiKeySvc := newAPIKeyServiceForAuth(apiKeyRepo)

	// Register a user and create an API key.
	user, err := authSvc.Register(context.Background(), "apikey-dual@test.com", "APIKey User", "password1234")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	rawKey, _, err := apiKeySvc.CreateKey(context.Background(), user.ID, 0, "Test Key")
	if err != nil {
		t.Fatalf("create key failed: %v", err)
	}

	var gotUserID int64
	middleware := RequireAuthOrAPIKey(store, authSvc, apiKeySvc)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := UserFromContext(r.Context())
		if u != nil {
			gotUserID = u.ID
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Use API key without session — triggers the fallback path.
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+rawKey)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if gotUserID != user.ID {
		t.Errorf("expected user ID %d, got %d", user.ID, gotUserID)
	}
}

func TestRequireAuthOrAPIKey_InvalidAPIKey(t *testing.T) {
	repo := newMockUserRepo()
	authSvc := NewService(repo)
	store := NewSessionStore("test-secret-key-32-bytes-long!!!")
	apiKeyRepo := newMockAPIKeyRepoForAuth()
	apiKeySvc := newAPIKeyServiceForAuth(apiKeyRepo)

	middleware := RequireAuthOrAPIKey(store, authSvc, apiKeySvc)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid API key")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer tvn_invalid_key")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

