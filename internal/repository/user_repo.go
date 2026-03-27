package repository

import (
	"context"
	"errors"

	"github.com/taverns-red/tavern-url/internal/model"
)

// ErrEmailExists is returned when an email is already registered.
var ErrEmailExists = errors.New("email already exists")

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")

// UserRepository defines the interface for user persistence.
type UserRepository interface {
	// Create persists a new user. Returns ErrEmailExists if the email is taken.
	Create(ctx context.Context, user *model.User) error

	// GetByEmail retrieves a user by email. Returns ErrUserNotFound if not found.
	GetByEmail(ctx context.Context, email string) (*model.User, error)

	// GetByID retrieves a user by ID. Returns ErrUserNotFound if not found.
	GetByID(ctx context.Context, id int64) (*model.User, error)
}
