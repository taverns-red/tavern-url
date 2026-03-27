package model

import "time"

// APIKey represents an API key for programmatic access.
type APIKey struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	OrgID      int64     `json:"org_id"`
	Name       string    `json:"name"`
	KeyHash    string    `json:"-"` // SHA-256 hash of the key — never exposed
	KeyPrefix  string    `json:"key_prefix"` // First 8 chars for identification
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
