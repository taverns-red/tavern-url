package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	// First 3 should be allowed.
	for i := 0; i < 3; i++ {
		if !rl.Allow("user1") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 4th should be denied.
	if rl.Allow("user1") {
		t.Error("4th request should be denied")
	}

	// Different key should still be allowed.
	if !rl.Allow("user2") {
		t.Error("different key should be allowed")
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := NewRateLimiter(1, 50*time.Millisecond)

	if !rl.Allow("key") {
		t.Error("first request should be allowed")
	}
	if rl.Allow("key") {
		t.Error("second request should be denied")
	}

	// Wait for window to reset.
	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("key") {
		t.Error("request after window reset should be allowed")
	}
}

func TestRateLimiter_Middleware_Allowed(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)

	var called bool
	handler := rl.Middleware(ByIP)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Error("handler should be called when under rate limit")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRateLimiter_Middleware_RateLimited(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)

	handler := rl.Middleware(ByIP)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "5.6.7.8:9999"

	// First request — allowed.
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("first request: expected 200, got %d", w.Code)
	}

	// Second request — rate limited.
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", w2.Code)
	}
}

func TestByIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	got := ByIP(req)
	if got != "192.168.1.1:12345" {
		t.Errorf("expected RemoteAddr, got %q", got)
	}
}

