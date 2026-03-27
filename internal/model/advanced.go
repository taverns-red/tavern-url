package model

import "time"

// RedirectRule defines a conditional redirect for A/B testing, geo, or device targeting.
type RedirectRule struct {
	ID             int64     `json:"id"`
	LinkID         int64     `json:"link_id"`
	ConditionType  string    `json:"condition_type"`  // "geo_country", "device_type", "weighted"
	ConditionValue string    `json:"condition_value"` // e.g., "US", "mobile", "50"
	TargetURL      string    `json:"target_url"`
	Weight         int       `json:"weight"`
	Priority       int       `json:"priority"`
	CreatedAt      time.Time `json:"created_at"`
}

// Webhook represents a webhook subscription for an org.
type Webhook struct {
	ID        int64     `json:"id"`
	OrgID     int64     `json:"org_id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	Secret    string    `json:"-"` // Never expose secret in API responses
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// Application represents an NFP onboarding application.
type Application struct {
	ID           int64      `json:"id"`
	OrgName      string     `json:"org_name"`
	Mission      string     `json:"mission"`
	Website      string     `json:"website,omitempty"`
	ContactEmail string     `json:"contact_email"`
	ContactName  string     `json:"contact_name"`
	Status       string     `json:"status"` // pending, approved, rejected
	ReviewedBy   *int64     `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
