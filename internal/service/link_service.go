package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/repository"
)

const maxSlugRetries = 5

// LinkService handles link business logic.
type LinkService struct {
	repo repository.LinkRepository
}

// NewLinkService creates a new LinkService.
func NewLinkService(repo repository.LinkRepository) *LinkService {
	return &LinkService{repo: repo}
}

// CreateLink creates a new short link.
// If customSlug is non-nil and non-empty, it is used as the slug.
// Otherwise, a random slug is generated.
func (s *LinkService) CreateLink(ctx context.Context, originalURL string, customSlug *string) (*model.Link, error) {
	// Validate URL.
	if err := validateURL(originalURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	link := &model.Link{
		OriginalURL: originalURL,
	}

	// Determine slug.
	if customSlug != nil && *customSlug != "" {
		if err := ValidateCustomSlug(*customSlug); err != nil {
			return nil, fmt.Errorf("invalid custom slug: %w", err)
		}
		link.Slug = *customSlug
		if err := s.repo.Create(ctx, link); err != nil {
			if errors.Is(err, repository.ErrSlugExists) {
				return nil, fmt.Errorf("slug %q is already taken", *customSlug)
			}
			return nil, err
		}
		return link, nil
	}

	// Auto-generate slug with collision retry.
	for i := 0; i < maxSlugRetries; i++ {
		link.Slug = GenerateSlug()
		err := s.repo.Create(ctx, link)
		if err == nil {
			return link, nil
		}
		if !errors.Is(err, repository.ErrSlugExists) {
			return nil, err
		}
		// Slug collision — retry with a new slug.
	}

	return nil, fmt.Errorf("failed to generate unique slug after %d attempts", maxSlugRetries)
}

// GetBySlug retrieves a link by its slug.
func (s *LinkService) GetBySlug(ctx context.Context, slug string) (*model.Link, error) {
	return s.repo.GetBySlug(ctx, slug)
}

// GetByID retrieves a link by its ID.
func (s *LinkService) GetByID(ctx context.Context, id int64) (*model.Link, error) {
	return s.repo.GetByID(ctx, id)
}
// DeleteLink deletes a link by ID.
func (s *LinkService) DeleteLink(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// ListLinks returns all links.
func (s *LinkService) ListLinks(ctx context.Context) ([]model.Link, error) {
	return s.repo.ListAll(ctx)
}

// UpdateLink updates a link's original URL and/or slug.
func (s *LinkService) UpdateLink(ctx context.Context, id int64, newURL *string, newSlug *string) (*model.Link, error) {
	link, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if newURL != nil && *newURL != "" {
		if err := validateURL(*newURL); err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}
		link.OriginalURL = *newURL
	}
	if newSlug != nil && *newSlug != "" {
		link.Slug = *newSlug
	}

	if err := s.repo.Update(ctx, link.ID, link.OriginalURL); err != nil {
		return nil, err
	}
	return link, nil
}

// validateURL checks that the URL is a valid HTTP or HTTPS URL.
func validateURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("malformed URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme, got %q", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("URL must have a host")
	}
	return nil
}
