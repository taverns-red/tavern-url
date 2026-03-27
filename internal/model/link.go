package model

import "time"

// Link represents a shortened URL mapping.
type Link struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	OriginalURL string     `json:"original_url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	MaxClicks   *int64     `json:"max_clicks,omitempty"`
	ClickCount  int64      `json:"click_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// IsExpired returns true if the link has passed its expiration date.
func (l *Link) IsExpired() bool {
	return l.ExpiresAt != nil && time.Now().After(*l.ExpiresAt)
}

// IsExhausted returns true if the link has reached its max click limit.
func (l *Link) IsExhausted() bool {
	return l.MaxClicks != nil && l.ClickCount >= *l.MaxClicks
}
