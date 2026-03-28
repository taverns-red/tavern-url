package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/taverns-red/tavern-url/internal/auth"
	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/repository"
	"github.com/taverns-red/tavern-url/internal/service"
)

// mockOrgRepo is a test double for repository.OrgRepository.
type mockOrgRepo struct {
	orgs        map[int64]*model.Org
	slugs       map[string]*model.Org
	memberships map[int64][]model.Membership
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

func setupOrgHandler() (*OrgHandler, *chi.Mux) {
	repo := newMockOrgRepo()
	orgSvc := service.NewOrgService(repo)
	h := NewOrgHandler(orgSvc)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := auth.ContextWithUser(r.Context(), &model.User{ID: 1, Email: "test@example.com"})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Post("/api/v1/orgs", h.Create)
	r.Get("/api/v1/orgs", h.List)
	r.Post("/api/v1/orgs/{slug}/invite", h.Invite)
	r.Put("/api/v1/orgs/{slug}/members/{memberID}/role", h.UpdateRole)

	return h, r
}

func TestOrgCreate_JSON(t *testing.T) {
	_, r := setupOrgHandler()

	body := `{"name": "Test Org", "slug": "test-org"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orgs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp orgResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Name != "Test Org" {
		t.Errorf("expected name %q, got %q", "Test Org", resp.Name)
	}
}

func TestOrgCreate_Form(t *testing.T) {
	_, r := setupOrgHandler()

	form := url.Values{"name": {"Form Org"}, "slug": {"form-org"}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orgs", bytes.NewBufferString(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (HTML partial), got %d: %s", w.Code, w.Body.String())
	}
}

func TestOrgCreate_MissingFields(t *testing.T) {
	_, r := setupOrgHandler()

	body := `{"name": ""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orgs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestOrgList_Success(t *testing.T) {
	_, r := setupOrgHandler()

	// Create an org first.
	body := `{"name": "List Org", "slug": "list-org"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orgs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// List orgs.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/orgs", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var orgs []orgResponse
	if err := json.NewDecoder(w.Body).Decode(&orgs); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(orgs) != 1 {
		t.Errorf("expected 1 org, got %d", len(orgs))
	}
}

func TestOrgInvite_JSON(t *testing.T) {
	_, r := setupOrgHandler()

	// Create org first.
	body := `{"name": "Invite Org", "slug": "invite-org"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orgs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Invite member.
	inviteBody := `{"email": "alice@example.com", "role": "member"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/orgs/invite-org/invite", bytes.NewBufferString(inviteBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestOrgInvite_Form(t *testing.T) {
	_, r := setupOrgHandler()

	// Create org first.
	body := `{"name": "Form Invite Org", "slug": "form-invite"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orgs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Invite via form.
	form := url.Values{"email": {"bob@example.com"}, "role": {"admin"}}
	req = httptest.NewRequest(http.MethodPost, "/api/v1/orgs/form-invite/invite", bytes.NewBufferString(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (HTML partial), got %d: %s", w.Code, w.Body.String())
	}
}

func TestOrgInvite_MissingEmail(t *testing.T) {
	_, r := setupOrgHandler()

	body := `{"role": "admin"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orgs/any-org/invite", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestOrgCreate_Unauthenticated(t *testing.T) {
	repo := newMockOrgRepo()
	orgSvc := service.NewOrgService(repo)
	h := NewOrgHandler(orgSvc)

	r := chi.NewRouter()
	r.Post("/api/v1/orgs", h.Create)

	body := `{"name": "Unauth Org", "slug": "unauth-org"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orgs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// suppress unused import lint
var _ = context.Background
