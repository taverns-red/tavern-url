package model

import (
	"testing"
	"time"
)

func TestIsExpired_NotExpired(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	link := &Link{ExpiresAt: &future}
	if link.IsExpired() {
		t.Error("expected link with future expiration to not be expired")
	}
}

func TestIsExpired_Expired(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	link := &Link{ExpiresAt: &past}
	if !link.IsExpired() {
		t.Error("expected link with past expiration to be expired")
	}
}

func TestIsExpired_NoExpiration(t *testing.T) {
	link := &Link{}
	if link.IsExpired() {
		t.Error("expected link with no expiration to not be expired")
	}
}

func TestIsExhausted_NotExhausted(t *testing.T) {
	max := int64(100)
	link := &Link{MaxClicks: &max, ClickCount: 50}
	if link.IsExhausted() {
		t.Error("expected link with clicks below max to not be exhausted")
	}
}

func TestIsExhausted_Exhausted(t *testing.T) {
	max := int64(100)
	link := &Link{MaxClicks: &max, ClickCount: 100}
	if !link.IsExhausted() {
		t.Error("expected link at max clicks to be exhausted")
	}
}

func TestIsExhausted_OverLimit(t *testing.T) {
	max := int64(10)
	link := &Link{MaxClicks: &max, ClickCount: 15}
	if !link.IsExhausted() {
		t.Error("expected link over max clicks to be exhausted")
	}
}

func TestIsExhausted_NoLimit(t *testing.T) {
	link := &Link{ClickCount: 9999}
	if link.IsExhausted() {
		t.Error("expected link with no max to not be exhausted")
	}
}

func TestIsPasswordProtected_WithPassword(t *testing.T) {
	hash := "$2a$12$somehash"
	link := &Link{PasswordHash: &hash}
	if !link.IsPasswordProtected() {
		t.Error("expected link with password hash to be protected")
	}
}

func TestIsPasswordProtected_EmptyHash(t *testing.T) {
	empty := ""
	link := &Link{PasswordHash: &empty}
	if link.IsPasswordProtected() {
		t.Error("expected link with empty hash to not be protected")
	}
}

func TestIsPasswordProtected_NilHash(t *testing.T) {
	link := &Link{}
	if link.IsPasswordProtected() {
		t.Error("expected link with nil hash to not be protected")
	}
}
