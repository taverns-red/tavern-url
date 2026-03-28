package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	h := NewLinkHandler(svc, nil, "http://localhost:8080")

	r := chi.NewRouter()
	r.Post("/api/v1/links", h.Create)
	r.Post("/api/v1/links/bulk", h.BulkCreate)
	r.Get("/api/v1/links", h.List)
	r.Put("/api/v1/links/{id}", h.Update)
	r.Delete("/api/v1/links/{id}", h.Delete)
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

func TestRedirect_PasswordProtected_ShowsGate(t *testing.T) {
	_, r := setupPasswordHandler()

	// Create a password-protected link via JSON API.
	body := `{"url": "https://secret.example.com", "slug": "secret", "password": "hunter2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create failed: %d: %s", w.Code, w.Body.String())
	}

	// GET /secret — should show password gate (200), NOT redirect.
	req = httptest.NewRequest(http.MethodGet, "/secret", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (gate page), got %d", w.Code)
	}
	// Body should contain password form elements.
	if !bytes.Contains(w.Body.Bytes(), []byte("password")) {
		t.Error("expected gate page to mention password")
	}
}

func TestRedirect_PasswordProtected_CorrectPassword(t *testing.T) {
	_, r := setupPasswordHandler()

	body := `{"url": "https://secret2.example.com", "slug": "secret2", "password": "correct-horse"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create failed: %d: %s", w.Code, w.Body.String())
	}

	// POST /secret2 with correct password — should 302 redirect.
	form := "password=correct-horse"
	req = httptest.NewRequest(http.MethodPost, "/secret2", bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302 redirect, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "https://secret2.example.com" {
		t.Errorf("expected redirect to original URL, got %q", loc)
	}
}

func TestRedirect_PasswordProtected_WrongPassword(t *testing.T) {
	_, r := setupPasswordHandler()

	body := `{"url": "https://secret3.example.com", "slug": "secret3", "password": "right-pw"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create failed: %d: %s", w.Code, w.Body.String())
	}

	// POST /secret3 with wrong password — should re-render gate.
	form := "password=wrong-pw"
	req = httptest.NewRequest(http.MethodPost, "/secret3", bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (gate re-render), got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("Incorrect")) {
		t.Error("expected error message about incorrect password")
	}
}

// setupPasswordHandler creates a handler with POST route for password gate testing.
func setupPasswordHandler() (*LinkHandler, *chi.Mux) {
	repo := newMockRepo()
	svc := service.NewLinkService(repo)
	h := NewLinkHandler(svc, nil, "http://localhost:8080")

	r := chi.NewRouter()
	r.Post("/api/v1/links", h.Create)
	r.Get("/{slug}", h.Redirect)
	r.Post("/{slug}", h.Redirect) // Password gate POST
	r.Get("/health", Health)

	return h, r
}

func TestList_EmptyLinks(t *testing.T) {
	_, r := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var links []createLinkResponse
	if err := json.NewDecoder(w.Body).Decode(&links); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("expected 0 links, got %d", len(links))
	}
}

func TestList_WithLinks(t *testing.T) {
	_, r := setupHandler()

	// Create two links.
	for _, url := range []string{"https://a.com", "https://b.com"} {
		body := fmt.Sprintf(`{"url": "%s"}`, url)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var links []createLinkResponse
	json.NewDecoder(w.Body).Decode(&links)
	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}
}

func TestList_SearchFilter(t *testing.T) {
	_, r := setupHandler()

	// Create links.
	for _, pair := range []struct{ url, slug string }{
		{"https://alpha.com", "alpha"},
		{"https://beta.com", "beta"},
	} {
		body := fmt.Sprintf(`{"url": "%s", "slug": "%s"}`, pair.url, pair.slug)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links?q=alpha", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var links []createLinkResponse
	json.NewDecoder(w.Body).Decode(&links)
	if len(links) != 1 {
		t.Errorf("expected 1 filtered link, got %d", len(links))
	}
}

func TestDelete_Success(t *testing.T) {
	_, r := setupHandler()

	// Create a link.
	body := `{"url": "https://delete-me.com", "slug": "del-test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var created createLinkResponse
	json.NewDecoder(w.Body).Decode(&created)

	// Delete it.
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/links/%d", created.ID), nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDelete_NotFound(t *testing.T) {
	_, r := setupHandler()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/links/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpdate_Success(t *testing.T) {
	_, r := setupHandler()

	// Create a link.
	body := `{"url": "https://original.com", "slug": "upd-test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var created createLinkResponse
	json.NewDecoder(w.Body).Decode(&created)

	// Update URL.
	updateBody := `{"url": "https://updated.com"}`
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/links/%d", created.ID), bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdate_NotFound(t *testing.T) {
	_, r := setupHandler()

	body := `{"url": "https://updated.com"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/links/999", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest {
		t.Errorf("expected 404 or 400 for missing link, got %d", w.Code)
	}
}

func TestBulkCreate_Success(t *testing.T) {
	_, r := setupHandler()

	body := `{"urls": ["https://bulk1.com", "https://bulk2.com", "https://bulk3.com"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links/bulk", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var links []createLinkResponse
	json.NewDecoder(w.Body).Decode(&links)
	if len(links) != 3 {
		t.Errorf("expected 3 links, got %d", len(links))
	}
}

func TestBulkCreate_TooMany(t *testing.T) {
	_, r := setupHandler()

	body := `{"urls": []}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links/bulk", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty URLs, got %d", w.Code)
	}
}

func TestCreate_FormEncoded(t *testing.T) {
	_, r := setupHandler()

	form := "url=https%3A%2F%2Fform-test.com&slug=form-link"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Form-encoded creates return HTML partials (200) not JSON (201).
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (HTML partial), got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreate_FormEncoded_MissingURL(t *testing.T) {
	_, r := setupHandler()

	form := "slug=no-url-form"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdate_FormEncoded(t *testing.T) {
	_, r := setupHandler()

	// Create a link first.
	body := `{"url": "https://form-update.com", "slug": "form-upd"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var created createLinkResponse
	json.NewDecoder(w.Body).Decode(&created)

	// Update via form.
	form := "url=https%3A%2F%2Fform-updated.com"
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/links/%d", created.ID), bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdate_InvalidJSON(t *testing.T) {
	_, r := setupHandler()

	req := httptest.NewRequest(http.MethodPut, "/api/v1/links/1", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdate_InvalidID(t *testing.T) {
	_, r := setupHandler()

	body := `{"url": "https://example.com"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/links/abc", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDelete_InvalidID(t *testing.T) {
	_, r := setupHandler()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/links/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid ID, got %d", w.Code)
	}
}

func TestDelete_WithHTMX(t *testing.T) {
	_, r := setupHandler()

	// Create a link.
	body := `{"url": "https://htmx-del.com", "slug": "htmx-del"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var created createLinkResponse
	json.NewDecoder(w.Body).Decode(&created)

	// Delete with HX-Request header.
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/links/%d", created.ID), nil)
	req.Header.Set("HX-Request", "true")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestBulkCreate_InvalidJSON(t *testing.T) {
	_, r := setupHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/links/bulk", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestParseID(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"0", 0},
		{"1", 1},
		{"123", 123},
		{"abc", 0},
		{"12a", 0},
		{"", 0},
	}
	for _, tc := range tests {
		got := parseID(tc.input)
		if got != tc.expected {
			t.Errorf("parseID(%q): expected %d, got %d", tc.input, tc.expected, got)
		}
	}
}

func TestContainsCI(t *testing.T) {
	if !containsCI("Hello World", "hello") {
		t.Error("expected case-insensitive match")
	}
	if containsCI("Hello", "xyz") {
		t.Error("did not expect match")
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestContainsMsg(t *testing.T) {
	err := fmt.Errorf("slug is already taken")
	if !containsMsg(err, "already taken") {
		t.Error("expected containsMsg to find substring")
	}
	if containsMsg(nil, "error") {
		t.Error("expected containsMsg to return false for nil")
	}
}
