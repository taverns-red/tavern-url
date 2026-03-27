package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCredentials is returned when email/password don't match.
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrWeakPassword is returned when the password is too short.
	ErrWeakPassword = errors.New("password must be at least 8 characters")

	// ErrInvalidEmail is returned when the email format is invalid.
	ErrInvalidEmail = errors.New("invalid email format")
)

const bcryptCost = 12

// contextKey is a private type for context keys to prevent collisions.
type contextKey string

const userContextKey contextKey = "user"

// Service handles authentication business logic.
type Service struct {
	users repository.UserRepository
}

// NewService creates a new auth Service.
func NewService(users repository.UserRepository) *Service {
	return &Service{users: users}
}

// Register creates a new user with the given email, name, and password.
func (s *Service) Register(ctx context.Context, email, name, password string) (*model.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	name = strings.TrimSpace(name)

	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if len(password) < 8 {
		return nil, ErrWeakPassword
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Email:        email,
		Name:         name,
		PasswordHash: string(hash),
	}

	if err := s.users.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrEmailExists) {
			return nil, fmt.Errorf("email %q is already registered", email)
		}
		return nil, err
	}

	return user, nil
}

// Login verifies credentials and returns the user if valid.
func (s *Service) Login(ctx context.Context, email, password string) (*model.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// GetUserByID retrieves a user by ID (for session rehydration).
func (s *Service) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	return s.users.GetByID(ctx, id)
}

// UserFromContext extracts the authenticated user from the request context.
func UserFromContext(ctx context.Context) *model.User {
	user, _ := ctx.Value(userContextKey).(*model.User)
	return user
}

// ContextWithUser adds a user to the context.
func ContextWithUser(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// RequireAuth is middleware that requires an authenticated session.
func RequireAuth(sessionStore SessionStore, authSvc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := sessionStore.GetUserID(r)
			if err != nil || userID == 0 {
				http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
				return
			}

			user, err := authSvc.GetUserByID(r.Context(), userID)
			if err != nil {
				http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
				return
			}

			ctx := ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateEmail does basic email format validation.
func validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}
	atIdx := strings.Index(email, "@")
	if atIdx < 1 {
		return ErrInvalidEmail
	}
	domain := email[atIdx+1:]
	if !strings.Contains(domain, ".") || len(domain) < 3 {
		return ErrInvalidEmail
	}
	return nil
}
