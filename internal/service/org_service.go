package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/repository"
)

// OrgService handles org business logic.
type OrgService struct {
	orgRepo repository.OrgRepository
}

// NewOrgService creates a new OrgService.
func NewOrgService(orgRepo repository.OrgRepository) *OrgService {
	return &OrgService{orgRepo: orgRepo}
}

var orgSlugPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// CreateOrg creates a new org and makes the given user the owner.
func (s *OrgService) CreateOrg(ctx context.Context, name string, slug string, ownerUserID int64) (*model.Org, error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(strings.ToLower(slug))

	if name == "" {
		return nil, fmt.Errorf("org name is required")
	}
	if len(slug) < 3 || len(slug) > 64 {
		return nil, fmt.Errorf("org slug must be 3-64 characters")
	}
	if !orgSlugPattern.MatchString(slug) {
		return nil, fmt.Errorf("org slug must be lowercase alphanumeric and hyphens only")
	}

	org := &model.Org{Name: name, Slug: slug}
	if err := s.orgRepo.Create(ctx, org, ownerUserID); err != nil {
		if err == repository.ErrOrgSlugExists {
			return nil, fmt.Errorf("org slug %q is already taken", slug)
		}
		return nil, err
	}
	return org, nil
}

// ListUserOrgs returns all orgs a user belongs to.
func (s *OrgService) ListUserOrgs(ctx context.Context, userID int64) ([]model.Org, error) {
	return s.orgRepo.ListByUser(ctx, userID)
}

// GetOrg retrieves an org by ID, verifying the user has access.
func (s *OrgService) GetOrg(ctx context.Context, orgID int64, userID int64) (*model.Org, error) {
	_, err := s.orgRepo.GetMembership(ctx, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("access denied or org not found")
	}
	return s.orgRepo.GetByID(ctx, orgID)
}
