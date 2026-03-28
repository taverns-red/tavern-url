package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/taverns-red/tavern-url/internal/model"
)

type mockAPIKeyRepo struct {
	keys   map[string]*model.APIKey // keyed by hash
	byUser map[int64][]*model.APIKey
	nextID int64
}

func newMockAPIKeyRepo() *mockAPIKeyRepo {
	return &mockAPIKeyRepo{
		keys:   make(map[string]*model.APIKey),
		byUser: make(map[int64][]*model.APIKey),
		nextID: 1,
	}
}

func (m *mockAPIKeyRepo) Create(ctx context.Context, key *model.APIKey) error {
	key.ID = m.nextID
	m.nextID++
	m.keys[key.KeyHash] = key
	m.byUser[key.UserID] = append(m.byUser[key.UserID], key)
	return nil
}

func (m *mockAPIKeyRepo) GetByHash(ctx context.Context, hash string) (*model.APIKey, error) {
	key, ok := m.keys[hash]
	if !ok {
		return nil, model.ErrNotFound
	}
	return key, nil
}

func (m *mockAPIKeyRepo) ListByUser(ctx context.Context, userID int64) ([]model.APIKey, error) {
	var keys []model.APIKey
	for _, k := range m.byUser[userID] {
		keys = append(keys, *k)
	}
	return keys, nil
}

func (m *mockAPIKeyRepo) Delete(ctx context.Context, id int64, userID int64) error {
	for hash, k := range m.keys {
		if k.ID == id && k.UserID == userID {
			delete(m.keys, hash)
			return nil
		}
	}
	return model.ErrNotFound
}

func (m *mockAPIKeyRepo) UpdateLastUsed(ctx context.Context, id int64) error {
	return nil
}

func TestCreateKey(t *testing.T) {
	svc := NewAPIKeyService(newMockAPIKeyRepo())

	rawKey, key, err := svc.CreateKey(context.Background(), 1, 1, "My Key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rawKey) < 20 {
		t.Errorf("expected long key, got %d chars", len(rawKey))
	}
	if rawKey[:4] != "tvn_" {
		t.Errorf("expected tvn_ prefix, got %q", rawKey[:4])
	}
	if key.Name != "My Key" {
		t.Errorf("expected name 'My Key', got %q", key.Name)
	}
	if key.KeyPrefix != rawKey[:12] {
		t.Errorf("expected prefix %q, got %q", rawKey[:12], key.KeyPrefix)
	}
}

func TestAuthenticate(t *testing.T) {
	svc := NewAPIKeyService(newMockAPIKeyRepo())

	rawKey, _, err := svc.CreateKey(context.Background(), 1, 1, "Test Key")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Verify the raw key can authenticate.
	key, err := svc.Authenticate(context.Background(), rawKey)
	if err != nil {
		t.Fatalf("authenticate failed: %v", err)
	}
	if key.Name != "Test Key" {
		t.Errorf("expected name 'Test Key', got %q", key.Name)
	}
}

func TestAuthenticate_InvalidKey(t *testing.T) {
	svc := NewAPIKeyService(newMockAPIKeyRepo())

	_, err := svc.Authenticate(context.Background(), "tvn_invalid_key")
	if err == nil {
		t.Error("expected error for invalid key")
	}
}

func TestKeyHash_NotReversible(t *testing.T) {
	svc := NewAPIKeyService(newMockAPIKeyRepo())

	rawKey, _, err := svc.CreateKey(context.Background(), 1, 1, "Hash Test")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Verify the stored hash is a SHA-256 hash of the raw key.
	hash := sha256.Sum256([]byte(rawKey))
	expectedHash := hex.EncodeToString(hash[:])

	key, _ := svc.Authenticate(context.Background(), rawKey)
	if key.KeyHash != expectedHash {
		t.Errorf("stored hash doesn't match SHA-256 of raw key")
	}
}

func TestCreateKey_EmptyName(t *testing.T) {
	svc := NewAPIKeyService(newMockAPIKeyRepo())

	_, _, err := svc.CreateKey(context.Background(), 1, 1, "")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestListKeys(t *testing.T) {
	svc := NewAPIKeyService(newMockAPIKeyRepo())

	// Create 3 keys for user 1.
	for _, name := range []string{"Key A", "Key B", "Key C"} {
		_, _, err := svc.CreateKey(context.Background(), 1, 1, name)
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
	}

	keys, err := svc.ListKeys(context.Background(), 1)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
}

func TestListKeys_EmptyForOtherUser(t *testing.T) {
	svc := NewAPIKeyService(newMockAPIKeyRepo())

	_, _, _ = svc.CreateKey(context.Background(), 1, 1, "User1 Key")

	keys, err := svc.ListKeys(context.Background(), 999)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected 0 keys for other user, got %d", len(keys))
	}
}

func TestDeleteKey(t *testing.T) {
	svc := NewAPIKeyService(newMockAPIKeyRepo())

	rawKey, key, _ := svc.CreateKey(context.Background(), 1, 1, "To Delete")

	err := svc.DeleteKey(context.Background(), key.ID, 1)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Verify key no longer authenticates.
	_, err = svc.Authenticate(context.Background(), rawKey)
	if err == nil {
		t.Error("expected error authenticating deleted key")
	}
}

func TestDeleteKey_NotFound(t *testing.T) {
	svc := NewAPIKeyService(newMockAPIKeyRepo())

	err := svc.DeleteKey(context.Background(), 999, 1)
	if err == nil {
		t.Error("expected error for non-existent key")
	}
}
