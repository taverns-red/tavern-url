package service

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeriveDevice(t *testing.T) {
	cases := []struct {
		ua       string
		expected string
	}{
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", "desktop"},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X)", "mobile"},
		{"Mozilla/5.0 (Linux; Android 13; Pixel 7)", "mobile"},
		{"Mozilla/5.0 (iPad; CPU OS 16_0 like Mac OS X)", "tablet"},
		{"Googlebot/2.1", "bot"},
		{"", "unknown"},
	}

	for _, tc := range cases {
		got := deriveDevice(tc.ua)
		if got != tc.expected {
			t.Errorf("deriveDevice(%q) = %q, want %q", tc.ua, got, tc.expected)
		}
	}
}

func TestDeriveReferrer(t *testing.T) {
	cases := []struct {
		referer  string
		expected string
	}{
		{"", "direct"},
		{"https://twitter.com/some/path", "twitter.com"},
		{"https://www.google.com/search?q=test", "www.google.com"},
		{"http://example.com:8080/page", "example.com"},
		{"https://t.co/abc123", "t.co"},
	}

	for _, tc := range cases {
		got := deriveReferrer(tc.referer)
		if got != tc.expected {
			t.Errorf("deriveReferrer(%q) = %q, want %q", tc.referer, got, tc.expected)
		}
	}
}

func TestDeriveCountry_CloudflareHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("CF-IPCountry", "gb")
	got := deriveCountry(req)
	if got != "GB" {
		t.Errorf("expected GB, got %q", got)
	}
}

func TestDeriveCountry_NoHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got := deriveCountry(req)
	if got != "unknown" {
		t.Errorf("expected unknown, got %q", got)
	}
}
