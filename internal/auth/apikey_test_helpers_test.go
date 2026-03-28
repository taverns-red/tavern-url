package auth

import (
	"context"

	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/service"
)

// mockAPIKeyRepoForAuth is a test double for service.APIKeyRepository,
// used by RequireAPIKey and RequireAuthOrAPIKey middleware tests.
type mockAPIKeyRepoForAuth struct {
	keys   map[string]*model.APIKey // keyed by hash
	nextID int64
}

func newMockAPIKeyRepoForAuth() *mockAPIKeyRepoForAuth {
	return &mockAPIKeyRepoForAuth{keys: make(map[string]*model.APIKey), nextID: 1}
}

func (m *mockAPIKeyRepoForAuth) Create(_ context.Context, key *model.APIKey) error {
	key.ID = m.nextID
	m.nextID++
	m.keys[key.KeyHash] = key
	return nil
}

func (m *mockAPIKeyRepoForAuth) GetByHash(_ context.Context, hash string) (*model.APIKey, error) {
	key, ok := m.keys[hash]
	if !ok {
		return nil, model.ErrNotFound
	}
	return key, nil
}

func (m *mockAPIKeyRepoForAuth) ListByUser(_ context.Context, userID int64) ([]model.APIKey, error) {
	var result []model.APIKey
	for _, k := range m.keys {
		if k.UserID == userID {
			result = append(result, *k)
		}
	}
	return result, nil
}

func (m *mockAPIKeyRepoForAuth) Delete(_ context.Context, id, userID int64) error {
	for hash, k := range m.keys {
		if k.ID == id && k.UserID == userID {
			delete(m.keys, hash)
			return nil
		}
	}
	return model.ErrNotFound
}

func (m *mockAPIKeyRepoForAuth) UpdateLastUsed(_ context.Context, _ int64) error {
	return nil
}

// newAPIKeyServiceForAuth constructs a real APIKeyService using the mock repo.
func newAPIKeyServiceForAuth(repo *mockAPIKeyRepoForAuth) *service.APIKeyService {
	return service.NewAPIKeyService(repo)
}

// compile-time check.
var _ service.APIKeyRepository = (*mockAPIKeyRepoForAuth)(nil)
