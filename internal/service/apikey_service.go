package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/taverns-red/tavern-url/internal/model"
)

// APIKeyRepository defines the interface for API key persistence.
type APIKeyRepository interface {
	Create(ctx context.Context, key *model.APIKey) error
	GetByHash(ctx context.Context, hash string) (*model.APIKey, error)
	ListByUser(ctx context.Context, userID int64) ([]model.APIKey, error)
	Delete(ctx context.Context, id int64, userID int64) error
	UpdateLastUsed(ctx context.Context, id int64) error
}

// APIKeyService handles API key business logic.
type APIKeyService struct {
	repo APIKeyRepository
}

// NewAPIKeyService creates a new APIKeyService.
func NewAPIKeyService(repo APIKeyRepository) *APIKeyService {
	return &APIKeyService{repo: repo}
}

// CreateKey generates a new API key and returns the raw key (shown once).
func (s *APIKeyService) CreateKey(ctx context.Context, userID, orgID int64, name string) (rawKey string, key *model.APIKey, err error) {
	if name == "" {
		return "", nil, fmt.Errorf("key name is required")
	}

	// Generate 32 random bytes → 64-char hex string.
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", nil, fmt.Errorf("failed to generate key: %w", err)
	}
	rawKey = "tvn_" + hex.EncodeToString(buf) // tvn_ prefix for easy identification

	// Hash with SHA-256 for storage.
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	key = &model.APIKey{
		UserID:   userID,
		OrgID:    orgID,
		Name:     name,
		KeyHash:  keyHash,
		KeyPrefix: rawKey[:12], // "tvn_" + first 8 hex chars
	}

	if err := s.repo.Create(ctx, key); err != nil {
		return "", nil, err
	}

	return rawKey, key, nil
}

// Authenticate validates an API key and returns the associated key record.
func (s *APIKeyService) Authenticate(ctx context.Context, rawKey string) (*model.APIKey, error) {
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	key, err := s.repo.GetByHash(ctx, keyHash)
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	// Update last used timestamp (fire and forget).
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.repo.UpdateLastUsed(bgCtx, key.ID) //nolint:errcheck
	}()

	return key, nil
}

// ListKeys returns all API keys for a user.
func (s *APIKeyService) ListKeys(ctx context.Context, userID int64) ([]model.APIKey, error) {
	return s.repo.ListByUser(ctx, userID)
}

// DeleteKey deletes an API key.
func (s *APIKeyService) DeleteKey(ctx context.Context, id int64, userID int64) error {
	return s.repo.Delete(ctx, id, userID)
}
