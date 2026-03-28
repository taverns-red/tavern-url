package service

import (
	"context"
	"errors"
	"testing"

	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/repository"
)

// mockLinkRepo is a mock implementation of LinkRepository for testing.
type mockLinkRepo struct {
	links    map[string]*model.Link
	nextID   int64
	createFn func(ctx context.Context, link *model.Link) error
}

func newMockLinkRepo() *mockLinkRepo {
	return &mockLinkRepo{
		links:  make(map[string]*model.Link),
		nextID: 1,
	}
}

func (m *mockLinkRepo) Create(ctx context.Context, link *model.Link) error {
	if m.createFn != nil {
		return m.createFn(ctx, link)
	}
	if _, exists := m.links[link.Slug]; exists {
		return repository.ErrSlugExists
	}
	link.ID = m.nextID
	m.nextID++
	m.links[link.Slug] = link
	return nil
}

func (m *mockLinkRepo) GetBySlug(ctx context.Context, slug string) (*model.Link, error) {
	link, ok := m.links[slug]
	if !ok {
		return nil, repository.ErrLinkNotFound
	}
	return link, nil
}

func (m *mockLinkRepo) Delete(ctx context.Context, id int64) error {
	for slug, link := range m.links {
		if link.ID == id {
			delete(m.links, slug)
			return nil
		}
	}
	return repository.ErrLinkNotFound
}

func (m *mockLinkRepo) GetByID(ctx context.Context, id int64) (*model.Link, error) {
	for _, link := range m.links {
		if link.ID == id {
			return link, nil
		}
	}
	return nil, repository.ErrLinkNotFound
}

func (m *mockLinkRepo) ListAll(ctx context.Context) ([]model.Link, error) {
	var links []model.Link
	for _, link := range m.links {
		links = append(links, *link)
	}
	return links, nil
}

func (m *mockLinkRepo) Update(ctx context.Context, id int64, originalURL string) error {
	for _, link := range m.links {
		if link.ID == id {
			link.OriginalURL = originalURL
			return nil
		}
	}
	return repository.ErrLinkNotFound
}

func TestCreateLink_AutoSlug(t *testing.T) {
	repo := newMockLinkRepo()
	svc := NewLinkService(repo)

	link, err := svc.CreateLink(context.Background(), "https://www.habitat.org", nil, nil, nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.Slug == "" {
		t.Error("expected auto-generated slug, got empty string")
	}
	if len(link.Slug) != 6 {
		t.Errorf("expected 6-char slug, got %d: %q", len(link.Slug), link.Slug)
	}
	if link.OriginalURL != "https://www.habitat.org" {
		t.Errorf("expected original URL to be preserved, got %q", link.OriginalURL)
	}
	if link.ID == 0 {
		t.Error("expected link to have an ID after creation")
	}
}

func TestCreateLink_CustomSlug(t *testing.T) {
	repo := newMockLinkRepo()
	svc := NewLinkService(repo)

	slug := "spring-gala"
	link, err := svc.CreateLink(context.Background(), "https://www.redcross.org", &slug, nil, nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.Slug != "spring-gala" {
		t.Errorf("expected slug %q, got %q", "spring-gala", link.Slug)
	}
}

func TestCreateLink_DuplicateCustomSlug(t *testing.T) {
	repo := newMockLinkRepo()
	svc := NewLinkService(repo)

	slug := "my-link"
	_, err := svc.CreateLink(context.Background(), "https://example.com", &slug, nil, nil, "")
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err = svc.CreateLink(context.Background(), "https://other.com", &slug, nil, nil, "")
	if err == nil {
		t.Error("expected error for duplicate slug, got nil")
	}
}

func TestCreateLink_InvalidURL(t *testing.T) {
	repo := newMockLinkRepo()
	svc := NewLinkService(repo)

	cases := []string{
		"",
		"not-a-url",
		"ftp://example.com",
		"://missing-scheme",
	}

	for _, u := range cases {
		_, err := svc.CreateLink(context.Background(), u, nil, nil, nil, "")
		if err == nil {
			t.Errorf("expected error for invalid URL %q, got nil", u)
		}
	}
}

func TestCreateLink_InvalidCustomSlug(t *testing.T) {
	repo := newMockLinkRepo()
	svc := NewLinkService(repo)

	slug := "ab" // too short
	_, err := svc.CreateLink(context.Background(), "https://example.com", &slug, nil, nil, "")
	if err == nil {
		t.Error("expected error for too-short custom slug, got nil")
	}
}

func TestCreateLink_AutoSlugRetry(t *testing.T) {
	repo := newMockLinkRepo()
	callCount := 0
	repo.createFn = func(ctx context.Context, link *model.Link) error {
		callCount++
		if callCount <= 3 {
			return repository.ErrSlugExists
		}
		link.ID = 1
		return nil
	}

	svc := NewLinkService(repo)
	link, err := svc.CreateLink(context.Background(), "https://example.com", nil, nil, nil, "")
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if link == nil {
		t.Fatal("expected link, got nil")
	}
	if callCount != 4 {
		t.Errorf("expected 4 attempts, got %d", callCount)
	}
}

func TestCreateLink_AutoSlugExhausted(t *testing.T) {
	repo := newMockLinkRepo()
	repo.createFn = func(ctx context.Context, link *model.Link) error {
		return repository.ErrSlugExists
	}

	svc := NewLinkService(repo)
	_, err := svc.CreateLink(context.Background(), "https://example.com", nil, nil, nil, "")
	if err == nil {
		t.Error("expected error after exhausted retries, got nil")
	}
}

func TestGetBySlug_Found(t *testing.T) {
	repo := newMockLinkRepo()
	svc := NewLinkService(repo)

	slug := "test-link"
	_, err := svc.CreateLink(context.Background(), "https://example.com", &slug, nil, nil, "")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	link, err := svc.GetBySlug(context.Background(), "test-link")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if link.OriginalURL != "https://example.com" {
		t.Errorf("expected URL %q, got %q", "https://example.com", link.OriginalURL)
	}
}

func TestGetBySlug_NotFound(t *testing.T) {
	repo := newMockLinkRepo()
	svc := NewLinkService(repo)

	_, err := svc.GetBySlug(context.Background(), "nonexistent")
	if !errors.Is(err, repository.ErrLinkNotFound) {
		t.Errorf("expected ErrLinkNotFound, got: %v", err)
	}
}

func TestDeleteLink_Success(t *testing.T) {
	repo := newMockLinkRepo()
	svc := NewLinkService(repo)

	slug := "to-delete"
	link, err := svc.CreateLink(context.Background(), "https://example.com", &slug, nil, nil, "")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	err = svc.DeleteLink(context.Background(), link.ID)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err = svc.GetBySlug(context.Background(), "to-delete")
	if !errors.Is(err, repository.ErrLinkNotFound) {
		t.Error("expected link to be deleted")
	}
}

func TestDeleteLink_NotFound(t *testing.T) {
	repo := newMockLinkRepo()
	svc := NewLinkService(repo)

	err := svc.DeleteLink(context.Background(), 999)
	if !errors.Is(err, repository.ErrLinkNotFound) {
		t.Errorf("expected ErrLinkNotFound, got: %v", err)
	}
}

func TestGetByID_Found(t *testing.T) {
	repo := newMockLinkRepo()
	svc := NewLinkService(repo)

	slug := "by-id"
	link, err := svc.CreateLink(context.Background(), "https://example.com", &slug, nil, nil, "")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	got, err := svc.GetByID(context.Background(), link.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Slug != "by-id" {
		t.Errorf("expected slug by-id, got %q", got.Slug)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := newMockLinkRepo()
	svc := NewLinkService(repo)

	_, err := svc.GetByID(context.Background(), 999)
	if !errors.Is(err, repository.ErrLinkNotFound) {
		t.Errorf("expected ErrLinkNotFound, got: %v", err)
	}
}

func TestListLinks(t *testing.T) {
	repo := newMockLinkRepo()
	svc := NewLinkService(repo)

	for _, slug := range []string{"list-a", "list-b"} {
		s := slug
		if _, err := svc.CreateLink(context.Background(), "https://example.com/"+s, &s, nil, nil, ""); err != nil {
			t.Fatalf("create failed: %v", err)
		}
	}

	links, err := svc.ListLinks(context.Background())
	if err != nil {
		t.Fatalf("ListLinks failed: %v", err)
	}
	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}
}

func TestUpdateLink_Success(t *testing.T) {
	repo := newMockLinkRepo()
	svc := NewLinkService(repo)

	slug := "update-me"
	link, err := svc.CreateLink(context.Background(), "https://old.com", &slug, nil, nil, "")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	newURL := "https://new.com"
	updated, err := svc.UpdateLink(context.Background(), link.ID, &newURL, nil)
	if err != nil {
		t.Fatalf("UpdateLink failed: %v", err)
	}
	if updated.OriginalURL != "https://new.com" {
		t.Errorf("expected new URL, got %q", updated.OriginalURL)
	}
}

func TestUpdateLink_NotFound(t *testing.T) {
	repo := newMockLinkRepo()
	svc := NewLinkService(repo)

	newURL := "https://nope.com"
	_, err := svc.UpdateLink(context.Background(), 999, &newURL, nil)
	if !errors.Is(err, repository.ErrLinkNotFound) {
		t.Errorf("expected ErrLinkNotFound, got: %v", err)
	}
}

func TestUpdateLink_InvalidURL(t *testing.T) {
	repo := newMockLinkRepo()
	svc := NewLinkService(repo)

	slug := "update-invalid"
	link, err := svc.CreateLink(context.Background(), "https://valid.com", &slug, nil, nil, "")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	badURL := "not-a-url"
	_, err = svc.UpdateLink(context.Background(), link.ID, &badURL, nil)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

