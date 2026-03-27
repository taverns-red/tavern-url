package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/taverns-red/tavern-url/internal/model"
	"github.com/taverns-red/tavern-url/internal/service"
)

// RequireAPIKey returns middleware that authenticates via API key.
// It reads the Authorization: Bearer tvn_... header.
func RequireAPIKey(apiKeySvc *service.APIKeyService, authSvc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawKey := extractBearerToken(r)
			if rawKey == "" {
				http.Error(w, `{"error":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
				return
			}

			key, err := apiKeySvc.Authenticate(r.Context(), rawKey)
			if err != nil {
				http.Error(w, `{"error":"invalid API key"}`, http.StatusUnauthorized)
				return
			}

			// Load full user to maintain context compatibility with session auth.
			user, err := authSvc.GetUserByID(r.Context(), key.UserID)
			if err != nil {
				http.Error(w, `{"error":"user not found"}`, http.StatusUnauthorized)
				return
			}

			ctx := ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuthOrAPIKey returns middleware that accepts either a session cookie or an API key.
// Session is checked first; if not authenticated, falls back to API key.
func RequireAuthOrAPIKey(sessionStore SessionStore, authSvc *Service, apiKeySvc *service.APIKeyService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try session auth first.
			userID, err := sessionStore.GetUserID(r)
			if err == nil && userID > 0 {
				user, err := authSvc.GetUserByID(r.Context(), userID)
				if err == nil {
					ctx := ContextWithUser(r.Context(), user)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// Fall back to API key.
			rawKey := extractBearerToken(r)
			if rawKey == "" {
				http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
				return
			}

			key, err := apiKeySvc.Authenticate(r.Context(), rawKey)
			if err != nil {
				http.Error(w, `{"error":"invalid API key"}`, http.StatusUnauthorized)
				return
			}

			user, err := authSvc.GetUserByID(r.Context(), key.UserID)
			if err != nil {
				http.Error(w, `{"error":"user not found"}`, http.StatusUnauthorized)
				return
			}

			ctx := ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the authenticated user's ID from context.
// Returns 0 if no user is in the context.
func UserIDFromContext(ctx context.Context) int64 {
	user := UserFromContext(ctx)
	if user == nil {
		return 0
	}
	return user.ID
}

// extractBearerToken extracts the token from "Authorization: Bearer <token>".
func extractBearerToken(r *http.Request) string {
	a := r.Header.Get("Authorization")
	if a == "" {
		return ""
	}
	parts := strings.SplitN(a, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// compile-time check: ensure model.User is used
var _ = (*model.User)(nil)
