package model

import "time"

// CustomDomain represents a custom domain configured for an org.
type CustomDomain struct {
	ID        int64     `json:"id"`
	OrgID     int64     `json:"org_id"`
	Domain    string    `json:"domain"`
	Verified  bool      `json:"verified"`
	DNSToken  string    `json:"dns_token"`
	CreatedAt time.Time `json:"created_at"`
}
