package auth

import (
	"encoding/gob"
	"net/http"

	"github.com/gorilla/sessions"
)

func init() {
	// Register types for gob encoding in session store.
	gob.Register(int64(0))
}

const (
	sessionName  = "tavern-session"
	sessionKey   = "user_id"
	maxAge       = 86400 * 7 // 7 days
)

// SessionStore wraps gorilla/sessions for session management.
type SessionStore struct {
	store *sessions.CookieStore
}

// NewSessionStore creates a new cookie-based session store.
func NewSessionStore(secret string) SessionStore {
	store := sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // Set to true in production behind HTTPS
	}
	return SessionStore{store: store}
}

// SetUserID stores the user ID in the session.
func (s SessionStore) SetUserID(w http.ResponseWriter, r *http.Request, userID int64) error {
	session, err := s.store.Get(r, sessionName)
	if err != nil {
		return err
	}
	session.Values[sessionKey] = userID
	return session.Save(r, w)
}

// GetUserID retrieves the user ID from the session.
func (s SessionStore) GetUserID(r *http.Request) (int64, error) {
	session, err := s.store.Get(r, sessionName)
	if err != nil {
		return 0, err
	}
	userID, ok := session.Values[sessionKey].(int64)
	if !ok {
		return 0, nil
	}
	return userID, nil
}

// Clear removes the session.
func (s SessionStore) Clear(w http.ResponseWriter, r *http.Request) error {
	session, err := s.store.Get(r, sessionName)
	if err != nil {
		return err
	}
	session.Options.MaxAge = -1
	return session.Save(r, w)
}
