package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/repository"
)

// testPool returns a pgxpool.Pool connected to the test database.
// Skips the test if DATABASE_URL is not set.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set — skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// cleanupUser removes a test user by email.
func cleanupUser(t *testing.T, pool *pgxpool.Pool, email string) {
	t.Helper()
	_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email)
}

// cleanupLink removes a test link by slug.
func cleanupLink(t *testing.T, pool *pgxpool.Pool, slug string) {
	t.Helper()
	_, _ = pool.Exec(context.Background(), "DELETE FROM links WHERE slug = $1", slug)
}

// --- PgUserRepository Tests ---

func TestPgUserRepo_CreateAndGetByEmail(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewPgUserRepository(pool)
	ctx := context.Background()

	email := "integration-test-user@tavern-test.com"
	cleanupUser(t, pool, email)
	t.Cleanup(func() { cleanupUser(t, pool, email) })

	user := &model.User{Email: email, Name: "Integration Test", PasswordHash: "hash123"}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if user.ID == 0 {
		t.Error("expected user to have ID after create")
	}
	if user.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	// GetByEmail.
	got, err := repo.GetByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("expected ID %d, got %d", user.ID, got.ID)
	}
	if got.Email != email {
		t.Errorf("expected email %q, got %q", email, got.Email)
	}
}

func TestPgUserRepo_CreateDuplicate(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewPgUserRepository(pool)
	ctx := context.Background()

	email := "integration-dup@tavern-test.com"
	cleanupUser(t, pool, email)
	t.Cleanup(func() { cleanupUser(t, pool, email) })

	user := &model.User{Email: email, Name: "First", PasswordHash: "hash"}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	user2 := &model.User{Email: email, Name: "Duplicate", PasswordHash: "hash2"}
	err := repo.Create(ctx, user2)
	if err == nil {
		t.Error("expected duplicate email error")
	}
	if err != repository.ErrEmailExists {
		t.Errorf("expected ErrEmailExists, got: %v", err)
	}
}

func TestPgUserRepo_GetByID(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewPgUserRepository(pool)
	ctx := context.Background()

	email := "integration-byid@tavern-test.com"
	cleanupUser(t, pool, email)
	t.Cleanup(func() { cleanupUser(t, pool, email) })

	user := &model.User{Email: email, Name: "ByID Test", PasswordHash: "hash"}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Email != email {
		t.Errorf("expected email %q, got %q", email, got.Email)
	}
}

func TestPgUserRepo_GetByID_NotFound(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewPgUserRepository(pool)

	_, err := repo.GetByID(context.Background(), 999999)
	if err != repository.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}
}

func TestPgUserRepo_GetByEmail_NotFound(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewPgUserRepository(pool)

	_, err := repo.GetByEmail(context.Background(), "nonexistent@tavern-test.com")
	if err != repository.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}
}

// --- PgLinkRepository Tests ---

func TestPgLinkRepo_CreateAndGetBySlug(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewPgLinkRepository(pool)
	ctx := context.Background()

	slug := "integ-test-slug"
	cleanupLink(t, pool, slug)
	t.Cleanup(func() { cleanupLink(t, pool, slug) })

	link := &model.Link{Slug: slug, OriginalURL: "https://example.com/integ-test"}
	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if link.ID == 0 {
		t.Error("expected link to have ID")
	}

	got, err := repo.GetBySlug(ctx, slug)
	if err != nil {
		t.Fatalf("GetBySlug failed: %v", err)
	}
	if got.OriginalURL != "https://example.com/integ-test" {
		t.Errorf("expected original URL, got %q", got.OriginalURL)
	}
}

func TestPgLinkRepo_CreateDuplicate(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewPgLinkRepository(pool)
	ctx := context.Background()

	slug := "integ-dup-slug"
	cleanupLink(t, pool, slug)
	t.Cleanup(func() { cleanupLink(t, pool, slug) })

	link := &model.Link{Slug: slug, OriginalURL: "https://example.com/1"}
	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	link2 := &model.Link{Slug: slug, OriginalURL: "https://example.com/2"}
	err := repo.Create(ctx, link2)
	if err != repository.ErrSlugExists {
		t.Errorf("expected ErrSlugExists, got: %v", err)
	}
}

func TestPgLinkRepo_GetByID(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewPgLinkRepository(pool)
	ctx := context.Background()

	slug := "integ-byid-link"
	cleanupLink(t, pool, slug)
	t.Cleanup(func() { cleanupLink(t, pool, slug) })

	link := &model.Link{Slug: slug, OriginalURL: "https://example.com/byid"}
	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := repo.GetByID(ctx, link.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Slug != slug {
		t.Errorf("expected slug %q, got %q", slug, got.Slug)
	}
}

func TestPgLinkRepo_GetByID_NotFound(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewPgLinkRepository(pool)

	_, err := repo.GetByID(context.Background(), 999999)
	if err != repository.ErrLinkNotFound {
		t.Errorf("expected ErrLinkNotFound, got: %v", err)
	}
}

func TestPgLinkRepo_ListAll(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewPgLinkRepository(pool)
	ctx := context.Background()

	slug := "integ-list-link"
	cleanupLink(t, pool, slug)
	t.Cleanup(func() { cleanupLink(t, pool, slug) })

	link := &model.Link{Slug: slug, OriginalURL: "https://example.com/list"}
	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	links, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(links) == 0 {
		t.Error("expected at least one link")
	}

	var found bool
	for _, l := range links {
		if l.Slug == slug {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find slug %q in list", slug)
	}
}

func TestPgLinkRepo_Update(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewPgLinkRepository(pool)
	ctx := context.Background()

	slug := "integ-update-link"
	cleanupLink(t, pool, slug)
	t.Cleanup(func() { cleanupLink(t, pool, slug) })

	link := &model.Link{Slug: slug, OriginalURL: "https://old.com"}
	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := repo.Update(ctx, link.ID, "https://new.com"); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := repo.GetByID(ctx, link.ID)
	if err != nil {
		t.Fatalf("GetByID after update failed: %v", err)
	}
	if got.OriginalURL != "https://new.com" {
		t.Errorf("expected updated URL, got %q", got.OriginalURL)
	}
}

func TestPgLinkRepo_Update_NotFound(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewPgLinkRepository(pool)

	err := repo.Update(context.Background(), 999999, "https://nope.com")
	if err != repository.ErrLinkNotFound {
		t.Errorf("expected ErrLinkNotFound, got: %v", err)
	}
}

func TestPgLinkRepo_Delete(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewPgLinkRepository(pool)
	ctx := context.Background()

	slug := "integ-delete-link"
	cleanupLink(t, pool, slug)
	t.Cleanup(func() { cleanupLink(t, pool, slug) })

	link := &model.Link{Slug: slug, OriginalURL: "https://delete.com"}
	if err := repo.Create(ctx, link); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := repo.Delete(ctx, link.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := repo.GetByID(ctx, link.ID)
	if err != repository.ErrLinkNotFound {
		t.Errorf("expected ErrLinkNotFound after delete, got: %v", err)
	}
}

func TestPgLinkRepo_Delete_NotFound(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewPgLinkRepository(pool)

	err := repo.Delete(context.Background(), 999999)
	if err != repository.ErrLinkNotFound {
		t.Errorf("expected ErrLinkNotFound, got: %v", err)
	}
}
