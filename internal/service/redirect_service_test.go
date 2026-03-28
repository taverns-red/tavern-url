package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taverns-red/tavern-url/internal/model"
)

func TestEvaluate_NoRules(t *testing.T) {
	svc := NewRedirectService()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	result := svc.Evaluate(nil, req)
	if result != "" {
		t.Errorf("expected empty string for no rules, got %q", result)
	}
}

func TestEvaluate_GeoCountry_Match(t *testing.T) {
	svc := NewRedirectService()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("CF-IPCountry", "US")

	rules := []model.RedirectRule{
		{ConditionType: "geo_country", ConditionValue: "US", TargetURL: "https://us.example.com"},
	}

	result := svc.Evaluate(rules, req)
	if result != "https://us.example.com" {
		t.Errorf("expected US target, got %q", result)
	}
}

func TestEvaluate_GeoCountry_NoMatch(t *testing.T) {
	svc := NewRedirectService()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("CF-IPCountry", "GB")

	rules := []model.RedirectRule{
		{ConditionType: "geo_country", ConditionValue: "US", TargetURL: "https://us.example.com"},
	}

	result := svc.Evaluate(rules, req)
	if result != "" {
		t.Errorf("expected empty for non-matching country, got %q", result)
	}
}

func TestEvaluate_GeoCountry_CaseInsensitive(t *testing.T) {
	svc := NewRedirectService()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("CF-IPCountry", "us")

	rules := []model.RedirectRule{
		{ConditionType: "geo_country", ConditionValue: "US", TargetURL: "https://us.example.com"},
	}

	result := svc.Evaluate(rules, req)
	if result != "https://us.example.com" {
		t.Errorf("expected case-insensitive geo match, got %q", result)
	}
}

func TestEvaluate_DeviceType_Mobile(t *testing.T) {
	svc := NewRedirectService()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)")

	rules := []model.RedirectRule{
		{ConditionType: "device_type", ConditionValue: "mobile", TargetURL: "https://m.example.com"},
	}

	result := svc.Evaluate(rules, req)
	if result != "https://m.example.com" {
		t.Errorf("expected mobile target, got %q", result)
	}
}

func TestEvaluate_DeviceType_Desktop(t *testing.T) {
	svc := NewRedirectService()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")

	rules := []model.RedirectRule{
		{ConditionType: "device_type", ConditionValue: "desktop", TargetURL: "https://desktop.example.com"},
	}

	result := svc.Evaluate(rules, req)
	if result != "https://desktop.example.com" {
		t.Errorf("expected desktop target, got %q", result)
	}
}

func TestEvaluate_Weighted_FullWeight(t *testing.T) {
	svc := NewRedirectService()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Weight 100 should always match.
	rules := []model.RedirectRule{
		{ConditionType: "weighted", Weight: 100, TargetURL: "https://variant-a.com"},
	}

	result := svc.Evaluate(rules, req)
	if result != "https://variant-a.com" {
		t.Errorf("expected 100%% weight to always match, got %q", result)
	}
}

func TestEvaluate_Weighted_ZeroWeight(t *testing.T) {
	svc := NewRedirectService()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Weight 0 should never match.
	rules := []model.RedirectRule{
		{ConditionType: "weighted", Weight: 0, TargetURL: "https://variant-a.com"},
	}

	result := svc.Evaluate(rules, req)
	if result != "" {
		t.Errorf("expected 0%% weight to never match, got %q", result)
	}
}

func TestEvaluate_MultipleRules_FirstMatch(t *testing.T) {
	svc := NewRedirectService()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("CF-IPCountry", "US")
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone)")

	rules := []model.RedirectRule{
		{ConditionType: "geo_country", ConditionValue: "US", TargetURL: "https://geo-first.com"},
		{ConditionType: "device_type", ConditionValue: "mobile", TargetURL: "https://mobile-second.com"},
	}

	result := svc.Evaluate(rules, req)
	if result != "https://geo-first.com" {
		t.Errorf("expected first matching rule to win, got %q", result)
	}
}

func TestDetectDevice_Tablet(t *testing.T) {
	result := detectDevice("mozilla/5.0 (ipad; cpu os 14_0 like mac os x)")
	if result != "tablet" {
		t.Errorf("expected tablet, got %q", result)
	}
}

func TestDetectDevice_Android(t *testing.T) {
	result := detectDevice("mozilla/5.0 (linux; android 11)")
	if result != "mobile" {
		t.Errorf("expected mobile for android, got %q", result)
	}
}
