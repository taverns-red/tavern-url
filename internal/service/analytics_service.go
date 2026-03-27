package service

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/repository"
)

// AnalyticsService handles click tracking and analytics retrieval.
type AnalyticsService struct {
	clickRepo repository.ClickRepository
}

// NewAnalyticsService creates a new AnalyticsService.
func NewAnalyticsService(clickRepo repository.ClickRepository) *AnalyticsService {
	return &AnalyticsService{clickRepo: clickRepo}
}

// RecordClick records a click event from an HTTP request.
// Privacy: IP is used only for country derivation and never stored.
func (s *AnalyticsService) RecordClick(ctx context.Context, linkID int64, r *http.Request) {
	event := &model.ClickEvent{
		LinkID:         linkID,
		Country:        deriveCountry(r),
		DeviceCategory: deriveDevice(r.UserAgent()),
		ReferrerDomain: deriveReferrer(r.Referer()),
	}

	// Fire and forget — don't block the redirect on analytics.
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.clickRepo.Record(bgCtx, event) //nolint:errcheck
	}()
}

// GetSummary returns analytics for a link within a date range.
func (s *AnalyticsService) GetSummary(ctx context.Context, linkID int64, days int) (*model.ClickSummary, error) {
	to := time.Now()
	from := to.AddDate(0, 0, -days)
	return s.clickRepo.GetSummary(ctx, linkID, from, to)
}

// GetTotalClicks returns the total click count for a link.
func (s *AnalyticsService) GetTotalClicks(ctx context.Context, linkID int64) (int64, error) {
	return s.clickRepo.GetTotalClicks(ctx, linkID)
}

// deriveCountry extracts a country code from the request.
// In production, this would use MaxMind GeoLite2 or similar.
// For now, uses Cloudflare's CF-IPCountry header or "unknown".
func deriveCountry(r *http.Request) string {
	// Cloudflare adds this header automatically.
	if country := r.Header.Get("CF-IPCountry"); country != "" {
		return strings.ToUpper(country)
	}
	// Fallback: could use a GeoIP database here.
	return "unknown"
}

// deriveDevice categorizes the device from the User-Agent string.
func deriveDevice(ua string) string {
	ua = strings.ToLower(ua)

	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "tablet"
	}
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		return "mobile"
	}
	if strings.Contains(ua, "bot") || strings.Contains(ua, "crawler") || strings.Contains(ua, "spider") {
		return "bot"
	}
	if ua == "" {
		return "unknown"
	}
	return "desktop"
}

// deriveReferrer extracts the domain from the Referer header.
func deriveReferrer(referer string) string {
	if referer == "" {
		return "direct"
	}

	// Strip protocol.
	r := referer
	if idx := strings.Index(r, "://"); idx >= 0 {
		r = r[idx+3:]
	}

	// Extract host.
	host := r
	if idx := strings.IndexAny(r, "/?#"); idx >= 0 {
		host = r[:idx]
	}

	// Strip port.
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	if host == "" {
		return "direct"
	}
	return strings.ToLower(host)
}
