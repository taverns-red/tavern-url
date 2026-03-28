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
	"github.com/taverns-red/tavern-url/internal/auth"
	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/service"
)

// mockAPIKeyRepoH implements service.APIKeyRepository for handler tests.
type mockAPIKeyRepoH struct {
	keys   map[int64]*model.APIKey
	byHash map[string]*model.APIKey
	nextID int64
}

func newMockAPIKeyRepoH() *mockAPIKeyRepoH {
	return &mockAPIKeyRepoH{
		keys:   make(map[int64]*model.APIKey),
		byHash: make(map[string]*model.APIKey),
		nextID: 1,
	}
}

func (m *mockAPIKeyRepoH) Create(_ context.Context, key *model.APIKey) error {
	key.ID = m.nextID
	m.nextID++
	m.keys[key.ID] = key
	m.byHash[key.KeyHash] = key
	return nil
}

func (m *mockAPIKeyRepoH) GetByHash(_ context.Context, hash string) (*model.APIKey, error) {
	key, ok := m.byHash[hash]
	if !ok {
		return nil, model.ErrNotFound
	}
	return key, nil
}

func (m *mockAPIKeyRepoH) ListByUser(_ context.Context, userID int64) ([]model.APIKey, error) {
	var result []model.APIKey
	for _, k := range m.keys {
		if k.UserID == userID {
			result = append(result, *k)
		}
	}
	return result, nil
}

func (m *mockAPIKeyRepoH) Delete(_ context.Context, id, userID int64) error {
	for _, k := range m.keys {
		if k.ID == id && k.UserID == userID {
			delete(m.keys, id)
			delete(m.byHash, k.KeyHash)
			return nil
		}
	}
	return model.ErrNotFound
}

func (m *mockAPIKeyRepoH) UpdateLastUsed(_ context.Context, _ int64) error {
	return nil
}

func setupAPIKeyHandler() (*APIKeyHandler, *chi.Mux) {
	repo := newMockAPIKeyRepoH()
	apiKeySvc := service.NewAPIKeyService(repo)
	h := NewAPIKeyHandler(apiKeySvc)

	r := chi.NewRouter()
	// Wrap routes with auth context.
	r.Post("/api/v1/keys", func(w http.ResponseWriter, r *http.Request) {
		user := &model.User{ID: 1, Email: "test@test.com"}
		ctx := auth.ContextWithUser(r.Context(), user)
		h.Create(w, r.WithContext(ctx))
	})
	r.Get("/api/v1/keys", func(w http.ResponseWriter, r *http.Request) {
		user := &model.User{ID: 1, Email: "test@test.com"}
		ctx := auth.ContextWithUser(r.Context(), user)
		h.List(w, r.WithContext(ctx))
	})
	r.Delete("/api/v1/keys/{id}", func(w http.ResponseWriter, r *http.Request) {
		user := &model.User{ID: 1, Email: "test@test.com"}
		ctx := auth.ContextWithUser(r.Context(), user)
		h.Delete(w, r.WithContext(ctx))
	})

	return h, r
}

func TestAPIKeyHandler_Create_JSON(t *testing.T) {
	_, r := setupAPIKeyHandler()

	body := `{"name":"My Key"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp createKeyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.RawKey == "" {
		t.Error("expected raw_key in response")
	}
}

func TestAPIKeyHandler_Create_MissingName(t *testing.T) {
	_, r := setupAPIKeyHandler()

	body := `{"name":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAPIKeyHandler_Create_InvalidJSON(t *testing.T) {
	_, r := setupAPIKeyHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAPIKeyHandler_Create_Unauthenticated(t *testing.T) {
	repo := newMockAPIKeyRepoH()
	apiKeySvc := service.NewAPIKeyService(repo)
	h := NewAPIKeyHandler(apiKeySvc)

	r := chi.NewRouter()
	r.Post("/api/v1/keys", h.Create)

	body := `{"name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAPIKeyHandler_Create_FormEncoded(t *testing.T) {
	_, r := setupAPIKeyHandler()

	form := "name=Form+Key"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAPIKeyHandler_Create_FormEncoded_MissingName(t *testing.T) {
	_, r := setupAPIKeyHandler()

	form := "name="
	req := httptest.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAPIKeyHandler_List(t *testing.T) {
	_, r := setupAPIKeyHandler()

	// Create a key first.
	body := `{"name":"List Key"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// List keys.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/keys", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var keys []keyResponse
	if err := json.NewDecoder(w.Body).Decode(&keys); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(keys) != 1 {
		t.Errorf("expected 1 key, got %d", len(keys))
	}
}

func TestAPIKeyHandler_List_Unauthenticated(t *testing.T) {
	repo := newMockAPIKeyRepoH()
	apiKeySvc := service.NewAPIKeyService(repo)
	h := NewAPIKeyHandler(apiKeySvc)

	r := chi.NewRouter()
	r.Get("/api/v1/keys", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/keys", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAPIKeyHandler_Delete(t *testing.T) {
	_, r := setupAPIKeyHandler()

	// Create a key.
	body := `{"name":"Delete Key"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/keys", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var created createKeyResponse
	json.NewDecoder(w.Body).Decode(&created)

	// Delete.
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/keys/%d", created.ID), nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAPIKeyHandler_Delete_NotFound(t *testing.T) {
	_, r := setupAPIKeyHandler()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/keys/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestAPIKeyHandler_Delete_InvalidID(t *testing.T) {
	_, r := setupAPIKeyHandler()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/keys/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAPIKeyHandler_Delete_Unauthenticated(t *testing.T) {
	repo := newMockAPIKeyRepoH()
	apiKeySvc := service.NewAPIKeyService(repo)
	h := NewAPIKeyHandler(apiKeySvc)

	r := chi.NewRouter()
	r.Delete("/api/v1/keys/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/keys/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
