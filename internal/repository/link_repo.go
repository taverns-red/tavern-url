package repository

import (
	"context"
	"errors"

	"github.com/taverns-red/tavern-url/internal/model"
)

// ErrSlugExists is returned when a slug already exists in the database.
var ErrSlugExists = errors.New("slug already exists")

// ErrLinkNotFound is returned when a link is not found.
var ErrLinkNotFound = errors.New("link not found")

// LinkRepository defines the interface for link persistence.
type LinkRepository interface {
	// Create persists a new link. Returns ErrSlugExists if the slug is taken.
	Create(ctx context.Context, link *model.Link) error

	// GetBySlug retrieves a link by its slug. Returns ErrLinkNotFound if not found.
	GetBySlug(ctx context.Context, slug string) (*model.Link, error)

	// GetByID retrieves a link by ID. Returns ErrLinkNotFound if not found.
	GetByID(ctx context.Context, id int64) (*model.Link, error)

	// ListAll returns all links, ordered by newest first.
	ListAll(ctx context.Context) ([]model.Link, error)

	// Delete removes a link by ID. Returns ErrLinkNotFound if not found.
	Delete(ctx context.Context, id int64) error

	// Update updates a link's original URL. Returns ErrLinkNotFound if not found.
	Update(ctx context.Context, id int64, originalURL string) error
}

