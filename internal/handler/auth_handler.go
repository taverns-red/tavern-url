package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/taverns-red/tavern-url/internal/auth"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authSvc      *auth.Service
	sessionStore auth.SessionStore
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authSvc *auth.Service, sessionStore auth.SessionStore) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, sessionStore: sessionStore}
}

type registerRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// writeFormError returns an HTML error alert for HTMX form submissions.
func writeFormError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, `<div class="alert alert-error">%s</div>`, msg)
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	isForm := strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")

	if isForm {
		if err := r.ParseForm(); err != nil {
			writeFormError(w, http.StatusBadRequest, "Invalid form data.")
			return
		}
		req.Email = r.FormValue("email")
		req.Name = r.FormValue("name")
		req.Password = r.FormValue("password")
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
			return
		}
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		if isForm {
			writeFormError(w, http.StatusBadRequest, "All fields are required.")
			return
		}
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "email, name, and password are required"})
		return
	}

	user, err := h.authSvc.Register(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		if isForm {
			if errors.Is(err, auth.ErrWeakPassword) {
				writeFormError(w, http.StatusBadRequest, "Password is too weak. Use at least 8 characters.")
			} else if errors.Is(err, auth.ErrInvalidEmail) {
				writeFormError(w, http.StatusBadRequest, "Please enter a valid email address.")
			} else if containsMsg(err, "already registered") {
				writeFormError(w, http.StatusConflict, "An account with this email already exists.")
			} else {
				writeFormError(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
			}
			return
		}
		if errors.Is(err, auth.ErrWeakPassword) || errors.Is(err, auth.ErrInvalidEmail) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		if containsMsg(err, "already registered") {
			writeJSON(w, http.StatusConflict, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	// Auto-login after registration.
	if err := h.sessionStore.SetUserID(w, r, user.ID); err != nil {
		if isForm {
			writeFormError(w, http.StatusInternalServerError, "Account created but login failed. Please log in manually.")
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to create session"})
		return
	}

	if isForm {
		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusOK)
		return
	}

	writeJSON(w, http.StatusCreated, userResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	})
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	isForm := strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")

	if isForm {
		if err := r.ParseForm(); err != nil {
			writeFormError(w, http.StatusBadRequest, "Invalid form data.")
			return
		}
		req.Email = r.FormValue("email")
		req.Password = r.FormValue("password")
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
			return
		}
	}

	if req.Email == "" || req.Password == "" {
		if isForm {
			writeFormError(w, http.StatusBadRequest, "Email and password are required.")
			return
		}
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "email and password are required"})
		return
	}

	user, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if isForm {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				writeFormError(w, http.StatusUnauthorized, "Invalid email or password.")
			} else {
				writeFormError(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
			}
			return
		}
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid email or password"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	if err := h.sessionStore.SetUserID(w, r, user.ID); err != nil {
		if isForm {
			writeFormError(w, http.StatusInternalServerError, "Login failed. Please try again.")
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to create session"})
		return
	}

	if isForm {
		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusOK)
		return
	}

	writeJSON(w, http.StatusOK, userResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	})
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.sessionStore.Clear(w, r); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to clear session"})
		return
	}

	// HTMX request — redirect to home.
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// Me handles GET /api/v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "not authenticated"})
		return
	}
	writeJSON(w, http.StatusOK, userResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	})
}
