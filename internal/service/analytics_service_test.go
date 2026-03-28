package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/taverns-red/tavern-url/internal/model"
)

// mockClickRepo implements repository.ClickRepository for analytics tests.
type mockClickRepo struct {
	mu     sync.Mutex
	events []*model.ClickEvent
}

func (m *mockClickRepo) Record(_ context.Context, event *model.ClickEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *mockClickRepo) GetSummary(_ context.Context, linkID int64, from, to time.Time) (*model.ClickSummary, error) {
	return &model.ClickSummary{
		TotalClicks: 42,
	}, nil
}

func (m *mockClickRepo) GetTotalClicks(_ context.Context, linkID int64) (int64, error) {
	return 42, nil
}

func TestNewAnalyticsService(t *testing.T) {
	repo := &mockClickRepo{}
	svc := NewAnalyticsService(repo)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestAnalyticsService_RecordClick(t *testing.T) {
	repo := &mockClickRepo{}
	svc := NewAnalyticsService(repo)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone)")
	req.Header.Set("Referer", "https://twitter.com/post")
	req.Header.Set("CF-IPCountry", "us")

	svc.RecordClick(context.Background(), 1, req)

	// RecordClick fires a goroutine — give it a moment.
	time.Sleep(100 * time.Millisecond)

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if len(repo.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(repo.events))
	}
	event := repo.events[0]
	if event.LinkID != 1 {
		t.Errorf("expected linkID 1, got %d", event.LinkID)
	}
	if event.Country != "US" {
		t.Errorf("expected country US, got %q", event.Country)
	}
	if event.DeviceCategory != "mobile" {
		t.Errorf("expected device mobile, got %q", event.DeviceCategory)
	}
}

func TestAnalyticsService_GetSummary(t *testing.T) {
	repo := &mockClickRepo{}
	svc := NewAnalyticsService(repo)

	summary, err := svc.GetSummary(context.Background(), 1, 7)
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}
	if summary.TotalClicks != 42 {
		t.Errorf("expected 42 clicks, got %d", summary.TotalClicks)
	}
}

func TestAnalyticsService_GetTotalClicks(t *testing.T) {
	repo := &mockClickRepo{}
	svc := NewAnalyticsService(repo)

	total, err := svc.GetTotalClicks(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTotalClicks failed: %v", err)
	}
	if total != 42 {
		t.Errorf("expected 42, got %d", total)
	}
}

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

