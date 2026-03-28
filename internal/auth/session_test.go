package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionStore_SetAndGetUserID(t *testing.T) {
	store := NewSessionStore("test-secret-key-32-bytes-long!!!", false)

	// Set the user ID.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	if err := store.SetUserID(w, req, 42); err != nil {
		t.Fatalf("SetUserID failed: %v", err)
	}

	// Extract the cookie from the response.
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie to be set")
	}

	// Make a new request with the session cookie.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}

	userID, err := store.GetUserID(req2)
	if err != nil {
		t.Fatalf("GetUserID failed: %v", err)
	}
	if userID != 42 {
		t.Errorf("expected userID 42, got %d", userID)
	}
}

func TestSessionStore_GetUserID_NoSession(t *testing.T) {
	store := NewSessionStore("test-secret-key-32-bytes-long!!!", false)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	userID, err := store.GetUserID(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != 0 {
		t.Errorf("expected 0 for no session, got %d", userID)
	}
}

func TestSessionStore_Clear(t *testing.T) {
	store := NewSessionStore("test-secret-key-32-bytes-long!!!", false)

	// Set user first.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	if err := store.SetUserID(w, req, 42); err != nil {
		t.Fatalf("SetUserID failed: %v", err)
	}
	cookies := w.Result().Cookies()

	// Clear the session.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	w2 := httptest.NewRecorder()
	if err := store.Clear(w2, req2); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify the cookie is cleared (MaxAge=-1 sets expiry in past).
	clearCookies := w2.Result().Cookies()
	if len(clearCookies) == 0 {
		t.Fatal("expected cleared cookie")
	}
	for _, c := range clearCookies {
		if c.MaxAge > 0 {
			t.Errorf("expected MaxAge <= 0 after clear, got %d", c.MaxAge)
		}
	}
}

func TestNewSessionStore(t *testing.T) {
	store := NewSessionStore("any-secret", false)
	if store.store == nil {
		t.Error("expected non-nil store")
	}
}
