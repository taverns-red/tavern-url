package model

import "time"

// ClickEvent represents a single click on a short link.
// Privacy: IP and raw User-Agent are never stored — only derived categories.
type ClickEvent struct {
	ID             int64     `json:"id"`
	LinkID         int64     `json:"link_id"`
	ClickedAt      time.Time `json:"clicked_at"`
	Country        string    `json:"country"`         // ISO 3166-1 alpha-2 or "unknown"
	DeviceCategory string    `json:"device_category"` // "desktop", "mobile", "tablet", "unknown"
	ReferrerDomain string    `json:"referrer_domain"` // e.g. "twitter.com" or "direct"
}

// ClickSummary holds aggregate analytics for a link.
type ClickSummary struct {
	TotalClicks int64            `json:"total_clicks"`
	ByDay       []DayCount       `json:"by_day"`
	ByCountry   []CategoryCount  `json:"by_country"`
	ByDevice    []CategoryCount  `json:"by_device"`
	ByReferrer  []CategoryCount  `json:"by_referrer"`
}

// DayCount is a click count for a specific day.
type DayCount struct {
	Date  string `json:"date"` // YYYY-MM-DD
	Count int64  `json:"count"`
}

// CategoryCount is a click count for a category.
type CategoryCount struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}
