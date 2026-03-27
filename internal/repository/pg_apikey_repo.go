package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/taverns-red/tavern-url/internal/model"
)

// PgAPIKeyRepository implements service.APIKeyRepository using PostgreSQL.
type PgAPIKeyRepository struct {
	pool *pgxpool.Pool
}

// NewPgAPIKeyRepository creates a new PgAPIKeyRepository.
func NewPgAPIKeyRepository(pool *pgxpool.Pool) *PgAPIKeyRepository {
	return &PgAPIKeyRepository{pool: pool}
}

// Create inserts a new API key.
func (r *PgAPIKeyRepository) Create(ctx context.Context, key *model.APIKey) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO api_keys (user_id, org_id, name, key_hash, key_prefix, created_at)
		 VALUES ($1, $2, $3, $4, $5, NOW())
		 RETURNING id, created_at`,
		key.UserID, key.OrgID, key.Name, key.KeyHash, key.KeyPrefix,
	).Scan(&key.ID, &key.CreatedAt)
}

// GetByHash retrieves an API key by its SHA-256 hash.
func (r *PgAPIKeyRepository) GetByHash(ctx context.Context, hash string) (*model.APIKey, error) {
	key := &model.APIKey{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, org_id, name, key_hash, key_prefix, last_used_at, created_at
		 FROM api_keys WHERE key_hash = $1`,
		hash,
	).Scan(&key.ID, &key.UserID, &key.OrgID, &key.Name, &key.KeyHash, &key.KeyPrefix, &key.LastUsedAt, &key.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return key, nil
}

// ListByUser returns all API keys for a user.
func (r *PgAPIKeyRepository) ListByUser(ctx context.Context, userID int64) ([]model.APIKey, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, org_id, name, key_prefix, last_used_at, created_at
		 FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []model.APIKey
	for rows.Next() {
		var k model.APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.OrgID, &k.Name, &k.KeyPrefix, &k.LastUsedAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// Delete removes an API key by ID, scoped to a user.
func (r *PgAPIKeyRepository) Delete(ctx context.Context, id int64, userID int64) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM api_keys WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

// UpdateLastUsed updates the last_used_at timestamp.
func (r *PgAPIKeyRepository) UpdateLastUsed(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`, id,
	)
	return err
}
