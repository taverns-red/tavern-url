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
