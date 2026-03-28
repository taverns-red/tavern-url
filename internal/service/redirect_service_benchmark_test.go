package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taverns-red/tavern-url/internal/model"
)

// BenchmarkEvaluate_GeoRule benchmarks the redirect rule evaluation for geo-country rules.
func BenchmarkEvaluate_GeoRule(b *testing.B) {
	svc := NewRedirectService()
	rules := []model.RedirectRule{
		{ConditionType: "geo_country", ConditionValue: "US", TargetURL: "https://us.example.com"},
		{ConditionType: "geo_country", ConditionValue: "GB", TargetURL: "https://uk.example.com"},
		{ConditionType: "device_type", ConditionValue: "mobile", TargetURL: "https://m.example.com"},
	}
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("CF-IPCountry", "US")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.Evaluate(rules, req)
	}
}

// BenchmarkEvaluate_DeviceRule benchmarks device-type rule evaluation.
func BenchmarkEvaluate_DeviceRule(b *testing.B) {
	svc := NewRedirectService()
	rules := []model.RedirectRule{
		{ConditionType: "device_type", ConditionValue: "mobile", TargetURL: "https://m.example.com"},
	}
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.Evaluate(rules, req)
	}
}

// BenchmarkEvaluate_NoMatch benchmarks the worst case — no rule matches.
func BenchmarkEvaluate_NoMatch(b *testing.B) {
	svc := NewRedirectService()
	rules := []model.RedirectRule{
		{ConditionType: "geo_country", ConditionValue: "JP", TargetURL: "https://jp.example.com"},
		{ConditionType: "geo_country", ConditionValue: "DE", TargetURL: "https://de.example.com"},
		{ConditionType: "device_type", ConditionValue: "tablet", TargetURL: "https://tablet.example.com"},
	}
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("CF-IPCountry", "US")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh)")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.Evaluate(rules, req)
	}
}
