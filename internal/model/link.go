package model

import "time"

// Link represents a shortened URL mapping.
type Link struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
