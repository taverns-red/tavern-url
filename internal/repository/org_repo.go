package repository

import (
	"context"
	"errors"

	"github.com/taverns-red/tavern-url/internal/model"
)

// ErrOrgNotFound is returned when an org is not found.
var ErrOrgNotFound = errors.New("org not found")

// ErrOrgSlugExists is returned when an org slug already exists.
var ErrOrgSlugExists = errors.New("org slug already exists")

// OrgRepository defines the interface for org persistence.
type OrgRepository interface {
	// Create persists a new org and adds the creator as owner.
	Create(ctx context.Context, org *model.Org, ownerUserID int64) error

	// GetByID retrieves an org by ID.
	GetByID(ctx context.Context, id int64) (*model.Org, error)

	// GetBySlug retrieves an org by slug.
	GetBySlug(ctx context.Context, slug string) (*model.Org, error)

	// ListByUser returns all orgs a user belongs to.
	ListByUser(ctx context.Context, userID int64) ([]model.Org, error)

	// GetMembership returns the membership for a user in an org.
	GetMembership(ctx context.Context, userID, orgID int64) (*model.Membership, error)

	// AddMember adds a user to an org with the given role.
	AddMember(ctx context.Context, userID, orgID int64, role model.Role) error

	// UpdateMemberRole changes a member's role in an org.
	UpdateMemberRole(ctx context.Context, orgID, userID int64, role model.Role) error
}
