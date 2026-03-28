package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/repository"
)

// GoogleOAuthConfig holds the configuration for Google OAuth.
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

const defaultUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

// GoogleProvider manages Google OAuth 2.0 authentication.
type GoogleProvider struct {
	config      *oauth2.Config
	users       repository.UserRepository
	userInfoURL string // configurable for testing
}

// NewGoogleProvider creates a new GoogleProvider.
func NewGoogleProvider(cfg GoogleOAuthConfig, users repository.UserRepository) *GoogleProvider {
	return &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		users:       users,
		userInfoURL: defaultUserInfoURL,
	}
}

// googleUserInfo is the response from Google's userinfo endpoint.
type googleUserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// LoginURL returns the URL to redirect users to Google's consent screen.
func (g *GoogleProvider) LoginURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// HandleCallback processes the OAuth callback and returns or creates a user.
func (g *GoogleProvider) HandleCallback(ctx context.Context, code string) (*model.User, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	client := g.config.Client(ctx, token)
	resp, err := client.Get(g.userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info: %w", err)
	}

	var info googleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	// Check if user already exists.
	user, err := g.users.GetByEmail(ctx, info.Email)
	if err == nil {
		return user, nil // Existing user — login.
	}

	if err != repository.ErrUserNotFound {
		return nil, err
	}

	// New user — create with a random password hash (they'll use OAuth).
	user = &model.User{
		Email:        info.Email,
		Name:         info.Name,
		PasswordHash: "oauth:google", // Sentinel value — not a valid bcrypt hash.
	}
	if err := g.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
