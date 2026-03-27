package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/taverns-red/tavern-url/internal/model"
)

// PgLinkRepository implements LinkRepository using PostgreSQL via pgx.
type PgLinkRepository struct {
	pool *pgxpool.Pool
}

// NewPgLinkRepository creates a new PgLinkRepository.
func NewPgLinkRepository(pool *pgxpool.Pool) *PgLinkRepository {
	return &PgLinkRepository{pool: pool}
}

// Create inserts a new link into the database.
// Returns ErrSlugExists if the slug violates the unique constraint.
func (r *PgLinkRepository) Create(ctx context.Context, link *model.Link) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO links (slug, original_url, created_at, updated_at)
		 VALUES ($1, $2, NOW(), NOW())
		 RETURNING id, created_at, updated_at`,
		link.Slug, link.OriginalURL,
	).Scan(&link.ID, &link.CreatedAt, &link.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrSlugExists
		}
		return err
	}
	return nil
}

// GetBySlug retrieves a link by its slug.
func (r *PgLinkRepository) GetBySlug(ctx context.Context, slug string) (*model.Link, error) {
	link := &model.Link{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, slug, original_url, created_at, updated_at
		 FROM links WHERE slug = $1`,
		slug,
	).Scan(&link.ID, &link.Slug, &link.OriginalURL, &link.CreatedAt, &link.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}
	return link, nil
}

// Delete removes a link by its ID.
func (r *PgLinkRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM links WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrLinkNotFound
	}
	return nil
}
