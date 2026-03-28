package service

import (
	"context"
	"testing"

	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/repository"
)

// mockOrgRepo is a test double for OrgRepository.
type mockOrgRepo struct {
	orgs        map[int64]*model.Org
	slugs       map[string]*model.Org
	memberships map[int64][]model.Membership // keyed by userID
	nextID      int64
}

func newMockOrgRepo() *mockOrgRepo {
	return &mockOrgRepo{
		orgs:        make(map[int64]*model.Org),
		slugs:       make(map[string]*model.Org),
		memberships: make(map[int64][]model.Membership),
		nextID:      1,
	}
}

func (m *mockOrgRepo) Create(ctx context.Context, org *model.Org, ownerUserID int64) error {
	if _, exists := m.slugs[org.Slug]; exists {
		return repository.ErrOrgSlugExists
	}
	org.ID = m.nextID
	m.nextID++
	m.orgs[org.ID] = org
	m.slugs[org.Slug] = org
	m.memberships[ownerUserID] = append(m.memberships[ownerUserID], model.Membership{
		UserID: ownerUserID, OrgID: org.ID, Role: model.RoleOwner,
	})
	return nil
}

func (m *mockOrgRepo) GetByID(ctx context.Context, id int64) (*model.Org, error) {
	org, ok := m.orgs[id]
	if !ok {
		return nil, repository.ErrOrgNotFound
	}
	return org, nil
}

func (m *mockOrgRepo) GetBySlug(ctx context.Context, slug string) (*model.Org, error) {
	org, ok := m.slugs[slug]
	if !ok {
		return nil, repository.ErrOrgNotFound
	}
	return org, nil
}

func (m *mockOrgRepo) ListByUser(ctx context.Context, userID int64) ([]model.Org, error) {
	var orgs []model.Org
	for _, mem := range m.memberships[userID] {
		if org, ok := m.orgs[mem.OrgID]; ok {
			orgs = append(orgs, *org)
		}
	}
	return orgs, nil
}

func (m *mockOrgRepo) GetMembership(ctx context.Context, userID, orgID int64) (*model.Membership, error) {
	for _, mem := range m.memberships[userID] {
		if mem.OrgID == orgID {
			return &mem, nil
		}
	}
	return nil, repository.ErrOrgNotFound
}

func (m *mockOrgRepo) AddMember(ctx context.Context, userID, orgID int64, role model.Role) error {
	m.memberships[userID] = append(m.memberships[userID], model.Membership{
		UserID: userID, OrgID: orgID, Role: role,
	})
	return nil
}

func (m *mockOrgRepo) UpdateMemberRole(ctx context.Context, orgID, userID int64, role model.Role) error {
	for i, mem := range m.memberships[userID] {
		if mem.OrgID == orgID {
			m.memberships[userID][i].Role = role
			return nil
		}
	}
	return repository.ErrOrgNotFound
}

func TestCreateOrg_Success(t *testing.T) {
	svc := NewOrgService(newMockOrgRepo())
	org, err := svc.CreateOrg(context.Background(), "Habitat for Humanity", "habitat", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.Name != "Habitat for Humanity" {
		t.Errorf("expected name, got %q", org.Name)
	}
	if org.Slug != "habitat" {
		t.Errorf("expected slug habitat, got %q", org.Slug)
	}
}

func TestCreateOrg_DuplicateSlug(t *testing.T) {
	svc := NewOrgService(newMockOrgRepo())
	_, err := svc.CreateOrg(context.Background(), "Org1", "my-org", 1)
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	_, err = svc.CreateOrg(context.Background(), "Org2", "my-org", 2)
	if err == nil {
		t.Error("expected error for duplicate slug")
	}
}

func TestCreateOrg_InvalidSlug(t *testing.T) {
	svc := NewOrgService(newMockOrgRepo())
	cases := []string{"ab", "has space", "special!", "under_score"}
	for _, slug := range cases {
		_, err := svc.CreateOrg(context.Background(), "Test", slug, 1)
		if err == nil {
			t.Errorf("expected error for slug %q", slug)
		}
	}
}

func TestCreateOrg_EmptyName(t *testing.T) {
	svc := NewOrgService(newMockOrgRepo())
	_, err := svc.CreateOrg(context.Background(), "", "valid-slug", 1)
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestListUserOrgs(t *testing.T) {
	repo := newMockOrgRepo()
	svc := NewOrgService(repo)

	_, err := svc.CreateOrg(context.Background(), "Org1", "org-1", 1)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	_, err = svc.CreateOrg(context.Background(), "Org2", "org-2", 1)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	orgs, err := svc.ListUserOrgs(context.Background(), 1)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(orgs) != 2 {
		t.Errorf("expected 2 orgs, got %d", len(orgs))
	}
}

func TestGetOrg_AccessDenied(t *testing.T) {
	repo := newMockOrgRepo()
	svc := NewOrgService(repo)

	_, err := svc.CreateOrg(context.Background(), "Private Org", "private", 1)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// User 2 should not have access.
	_, err = svc.GetOrg(context.Background(), 1, 2)
	if err == nil {
		t.Error("expected access denied error")
	}
	if err != nil && err.Error() != "access denied or org not found" {
		t.Errorf("expected access denied error, got: %v", err)
	}
}

func TestInviteMember_Success(t *testing.T) {
	repo := newMockOrgRepo()
	svc := NewOrgService(repo)

	_, err := svc.CreateOrg(context.Background(), "Invite Org", "invite-org", 1)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	err = svc.InviteMember(context.Background(), "invite-org", 1, "alice@example.com", "member")
	if err != nil {
		t.Errorf("expected successful invite, got: %v", err)
	}
}

func TestInviteMember_NonOwner(t *testing.T) {
	repo := newMockOrgRepo()
	svc := NewOrgService(repo)

	_, err := svc.CreateOrg(context.Background(), "Owner Only", "owner-only", 1)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// User 2 is not a member — should fail.
	err = svc.InviteMember(context.Background(), "owner-only", 2, "bob@example.com", "member")
	if err == nil {
		t.Error("expected permission denied error")
	}
}

func TestInviteMember_InvalidRole(t *testing.T) {
	repo := newMockOrgRepo()
	svc := NewOrgService(repo)

	err := svc.InviteMember(context.Background(), "any-org", 1, "x@test.com", "superadmin")
	if err == nil {
		t.Error("expected error for invalid role")
	}
}

func TestInviteMember_OrgNotFound(t *testing.T) {
	repo := newMockOrgRepo()
	svc := NewOrgService(repo)

	err := svc.InviteMember(context.Background(), "nonexistent", 1, "x@test.com", "member")
	if err == nil {
		t.Error("expected error for nonexistent org")
	}
}

func TestUpdateMemberRole_Success(t *testing.T) {
	repo := newMockOrgRepo()
	svc := NewOrgService(repo)

	_, err := svc.CreateOrg(context.Background(), "Role Org", "role-org", 1)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Add user 2 as member.
	_ = svc.InviteMember(context.Background(), "role-org", 1, "member@test.com", "member")

	// Owner (user 1) tries to change their own role — should be blocked.
	err = svc.UpdateMemberRole(context.Background(), "role-org", 1, 1, "admin")
	if err == nil {
		t.Error("expected error when changing owner's role")
	} else if err.Error() != "cannot change the owner's role" {
		t.Errorf("expected 'cannot change the owner's role', got: %v", err)
	}
}

func TestUpdateMemberRole_InvalidRole(t *testing.T) {
	repo := newMockOrgRepo()
	svc := NewOrgService(repo)

	err := svc.UpdateMemberRole(context.Background(), "any-org", 1, 2, "dictator")
	if err == nil {
		t.Error("expected error for invalid role")
	}
}

func TestUpdateMemberRole_PermissionDenied(t *testing.T) {
	repo := newMockOrgRepo()
	svc := NewOrgService(repo)

	_, err := svc.CreateOrg(context.Background(), "Perm Org", "perm-org", 1)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// User 2 has no membership — should be denied.
	err = svc.UpdateMemberRole(context.Background(), "perm-org", 2, 1, "admin")
	if err == nil {
		t.Error("expected permission denied error")
	}
}
