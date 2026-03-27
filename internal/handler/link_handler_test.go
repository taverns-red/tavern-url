package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/repository"
	"github.com/taverns-red/tavern-url/internal/service"
)

// mockLinkRepo is a test double for repository.LinkRepository.
type mockLinkRepo struct {
	links  map[string]*model.Link
	nextID int64
}

func newMockRepo() *mockLinkRepo {
	return &mockLinkRepo{
		links:  make(map[string]*model.Link),
		nextID: 1,
	}
}

func (m *mockLinkRepo) Create(ctx context.Context, link *model.Link) error {
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

func setupHandler() (*LinkHandler, *chi.Mux) {
	repo := newMockRepo()
	svc := service.NewLinkService(repo)
	h := NewLinkHandler(svc, "http://localhost:8080")

	r := chi.NewRouter()
	r.Post("/api/v1/links", h.Create)
	r.Get("/{slug}", h.Redirect)
	r.Get("/health", Health)

	return h, r
}

func TestCreate_Success(t *testing.T) {
	_, r := setupHandler()

	body := `{"url": "https://www.habitat.org"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp createLinkResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Slug == "" {
		t.Error("expected slug in response")
	}
	if resp.OriginalURL != "https://www.habitat.org" {
		t.Errorf("expected original URL, got %q", resp.OriginalURL)
	}
	if resp.ShortURL == "" {
		t.Error("expected short_url in response")
	}
}

func TestCreate_CustomSlug(t *testing.T) {
	_, r := setupHandler()

	body := `{"url": "https://www.redcross.org", "slug": "red-cross"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp createLinkResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Slug != "red-cross" {
		t.Errorf("expected slug %q, got %q", "red-cross", resp.Slug)
	}
}

func TestCreate_DuplicateSlug(t *testing.T) {
	_, r := setupHandler()

	body := `{"url": "https://example.com", "slug": "dup-test"}`

	// First create.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("first create failed: %d", w.Code)
	}

	// Second create — should conflict.
	req = httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreate_MissingURL(t *testing.T) {
	_, r := setupHandler()

	body := `{"slug": "no-url"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreate_InvalidURL(t *testing.T) {
	_, r := setupHandler()

	body := `{"url": "not-a-url"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	_, r := setupHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreate_ShortSlug(t *testing.T) {
	_, r := setupHandler()

	body := `{"url": "https://example.com", "slug": "ab"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for short slug, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRedirect_Success(t *testing.T) {
	_, r := setupHandler()

	// Create a link first.
	body := `{"url": "https://www.habitat.org", "slug": "habitat"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Follow the redirect.
	req = httptest.NewRequest(http.MethodGet, "/habitat", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", w.Code)
	}
	location := w.Header().Get("Location")
	if location != "https://www.habitat.org" {
		t.Errorf("expected Location %q, got %q", "https://www.habitat.org", location)
	}
}

func TestRedirect_NotFound(t *testing.T) {
	_, r := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHealth(t *testing.T) {
	_, r := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %q", resp["status"])
	}
}
