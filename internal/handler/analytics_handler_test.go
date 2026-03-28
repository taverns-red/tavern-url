package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/service"
)

// mockClickRepo is a test double for repository.ClickRepository.
type mockClickRepo struct {
	clicks map[int64]int64
}

func newMockClickRepo() *mockClickRepo {
	return &mockClickRepo{clicks: make(map[int64]int64)}
}

func (m *mockClickRepo) Record(_ context.Context, event *model.ClickEvent) error {
	m.clicks[event.LinkID]++
	return nil
}

func (m *mockClickRepo) GetSummary(_ context.Context, linkID int64, _, _ time.Time) (*model.ClickSummary, error) {
	total := m.clicks[linkID]
	return &model.ClickSummary{
		TotalClicks: total,
		ByDay: []model.DayCount{
			{Date: "2026-03-28", Count: total},
		},
		ByCountry: []model.CategoryCount{
			{Category: "US", Count: total},
		},
		ByDevice: []model.CategoryCount{
			{Category: "desktop", Count: total},
		},
		ByReferrer: []model.CategoryCount{
			{Category: "direct", Count: total},
		},
	}, nil
}

func (m *mockClickRepo) GetTotalClicks(_ context.Context, linkID int64) (int64, error) {
	return m.clicks[linkID], nil
}

func setupAnalyticsHandler() (*AnalyticsHandler, *ExportHandler, *chi.Mux) {
	linkRepo := newMockRepo()
	linkSvc := service.NewLinkService(linkRepo)
	clickRepo := newMockClickRepo()
	analyticsSvc := service.NewAnalyticsService(clickRepo)
	qrSvc := service.NewQRService()

	ah := NewAnalyticsHandler(analyticsSvc, qrSvc, linkSvc, "http://localhost:8080")
	eh := NewExportHandler(analyticsSvc)

	// Create a test link.
	_, _ = linkSvc.CreateLink(context.Background(), "https://example.com", strPtr("analytics-test"), nil, nil, "")

	r := chi.NewRouter()
	r.Get("/api/v1/links/{id}/analytics", ah.GetSummary)
	r.Get("/api/v1/links/{id}/qr", ah.QRCode)
	r.Get("/api/v1/links/{id}/analytics/export", eh.ExportCSV)

	return ah, eh, r
}

func strPtr(s string) *string { return &s }

func TestGetSummary_Success(t *testing.T) {
	_, _, r := setupAnalyticsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/1/analytics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var summary model.ClickSummary
	if err := json.NewDecoder(w.Body).Decode(&summary); err != nil {
		t.Fatalf("failed to decode summary: %v", err)
	}
}

func TestGetSummary_InvalidID(t *testing.T) {
	_, _, r := setupAnalyticsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/abc/analytics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetSummary_WithDays(t *testing.T) {
	_, _, r := setupAnalyticsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/1/analytics?days=7", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestQRCode_Success(t *testing.T) {
	_, _, r := setupAnalyticsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/1/qr", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("expected image/png, got %q", ct)
	}
}

func TestQRCode_CustomSize(t *testing.T) {
	_, _, r := setupAnalyticsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/1/qr?size=512", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestQRCode_NotFound(t *testing.T) {
	_, _, r := setupAnalyticsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/999/qr", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestQRCode_InvalidID(t *testing.T) {
	_, _, r := setupAnalyticsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/abc/qr", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestExportCSV_Success(t *testing.T) {
	_, _, r := setupAnalyticsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/1/analytics/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected text/csv, got %q", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); !strings.Contains(cd, "analytics-1.csv") {
		t.Errorf("expected analytics-1.csv in Content-Disposition, got %q", cd)
	}

	body := w.Body.String()
	for _, section := range []string{"Daily Clicks", "Countries", "Devices", "Referrers"} {
		if !strings.Contains(body, section) {
			t.Errorf("expected CSV to contain %q section", section)
		}
	}
}

func TestExportCSV_InvalidID(t *testing.T) {
	_, _, r := setupAnalyticsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/abc/analytics/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// Suppress unused import warning for fmt — used by setupAnalyticsHandler error formatting.
var _ = fmt.Sprintf
