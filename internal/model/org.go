package model

import "time"

// Role represents a user's role within an organization.
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

// Org represents an organization.
type Org struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Membership represents a user's membership in an organization.
type Membership struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"user_id"`
	OrgID  int64 `json:"org_id"`
	Role   Role  `json:"role"`
}
