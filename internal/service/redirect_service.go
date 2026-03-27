package service

import (
	"math/rand"
	"net/http"
	"strings"

	"github.com/taverns-red/tavern-url/internal/model"
)

// RedirectService evaluates redirect rules to determine the target URL.
type RedirectService struct{}

// NewRedirectService creates a new RedirectService.
func NewRedirectService() *RedirectService {
	return &RedirectService{}
}

// Evaluate checks rules against the request and returns the target URL.
// Returns empty string if no rules match (fall back to original URL).
func (s *RedirectService) Evaluate(rules []model.RedirectRule, r *http.Request) string {
	if len(rules) == 0 {
		return ""
	}

	for _, rule := range rules {
		switch rule.ConditionType {
		case "geo_country":
			country := r.Header.Get("CF-IPCountry")
			if strings.EqualFold(country, rule.ConditionValue) {
				return rule.TargetURL
			}
		case "device_type":
			ua := strings.ToLower(r.UserAgent())
			device := detectDevice(ua)
			if strings.EqualFold(device, rule.ConditionValue) {
				return rule.TargetURL
			}
		case "weighted":
			// Weighted random (A/B testing).
			if rand.Intn(100) < rule.Weight {
				return rule.TargetURL
			}
		}
	}
	return ""
}

func detectDevice(ua string) string {
	switch {
	case strings.Contains(ua, "mobile") || strings.Contains(ua, "iphone") || strings.Contains(ua, "android"):
		return "mobile"
	case strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad"):
		return "tablet"
	default:
		return "desktop"
	}
}
