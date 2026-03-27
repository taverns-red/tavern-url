package middleware

import (
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
