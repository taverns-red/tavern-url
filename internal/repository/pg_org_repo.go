package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/taverns-red/tavern-url/internal/model"
)

// PgOrgRepository implements OrgRepository using PostgreSQL.
type PgOrgRepository struct {
	pool *pgxpool.Pool
}

// NewPgOrgRepository creates a new PgOrgRepository.
func NewPgOrgRepository(pool *pgxpool.Pool) *PgOrgRepository {
	return &PgOrgRepository{pool: pool}
}

// Create inserts a new org and adds the creator as owner in a single transaction.
func (r *PgOrgRepository) Create(ctx context.Context, org *model.Org, ownerUserID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	err = tx.QueryRow(ctx,
		`INSERT INTO orgs (name, slug, created_at, updated_at)
		 VALUES ($1, $2, NOW(), NOW())
		 RETURNING id, created_at, updated_at`,
		org.Name, org.Slug,
	).Scan(&org.ID, &org.CreatedAt, &org.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrOrgSlugExists
		}
		return err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO memberships (user_id, org_id, role) VALUES ($1, $2, $3)`,
		ownerUserID, org.ID, string(model.RoleOwner),
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetByID retrieves an org by ID.
func (r *PgOrgRepository) GetByID(ctx context.Context, id int64) (*model.Org, error) {
	org := &model.Org{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, slug, created_at, updated_at FROM orgs WHERE id = $1`, id,
	).Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return org, nil
}

// GetBySlug retrieves an org by slug.
func (r *PgOrgRepository) GetBySlug(ctx context.Context, slug string) (*model.Org, error) {
	org := &model.Org{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, slug, created_at, updated_at FROM orgs WHERE slug = $1`, slug,
	).Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return org, nil
}

// ListByUser returns all orgs a user belongs to.
func (r *PgOrgRepository) ListByUser(ctx context.Context, userID int64) ([]model.Org, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT o.id, o.name, o.slug, o.created_at, o.updated_at
		 FROM orgs o
		 JOIN memberships m ON m.org_id = o.id
		 WHERE m.user_id = $1
		 ORDER BY o.name`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []model.Org
	for rows.Next() {
		var org model.Org
		if err := rows.Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, rows.Err()
}

// GetMembership returns the membership for a user in an org.
func (r *PgOrgRepository) GetMembership(ctx context.Context, userID, orgID int64) (*model.Membership, error) {
	m := &model.Membership{}
	var role string
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, org_id, role FROM memberships WHERE user_id = $1 AND org_id = $2`,
		userID, orgID,
	).Scan(&m.ID, &m.UserID, &m.OrgID, &role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	m.Role = model.Role(role)
	return m, nil
}

// AddMember adds a user to an org with the given role.
func (r *PgOrgRepository) AddMember(ctx context.Context, userID, orgID int64, role model.Role) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO memberships (user_id, org_id, role) VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, org_id) DO UPDATE SET role = $3`,
		userID, orgID, string(role),
	)
	return err
}

// UpdateMemberRole changes a member's role in an org.
func (r *PgOrgRepository) UpdateMemberRole(ctx context.Context, orgID, userID int64, role model.Role) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE memberships SET role = $1 WHERE org_id = $2 AND user_id = $3`,
		string(role), orgID, userID,
	)
	return err
}
