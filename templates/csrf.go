package templates

import "context"

type contextKey string

// CSRFContextKey is the robust key used to store/retrieve the token.
const CSRFContextKey = contextKey("csrf_token")

func getCSRFToken(ctx context.Context) string {
	if token, ok := ctx.Value(CSRFContextKey).(string); ok {
		return token
	}
	return ""
}
