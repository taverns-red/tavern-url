package auth

import (
	"testing"
)

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

// contains checks if s contains substr (simple helper for test assertions).
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
