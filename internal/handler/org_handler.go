package handler

import (
	"encoding/json"
	"net/http"

	"github.com/taverns-red/tavern-url/internal/auth"
	"github.com/taverns-red/tavern-url/internal/service"
)

// OrgHandler handles org HTTP requests.
type OrgHandler struct {
	orgSvc *service.OrgService
}

// NewOrgHandler creates a new OrgHandler.
func NewOrgHandler(orgSvc *service.OrgService) *OrgHandler {
	return &OrgHandler{orgSvc: orgSvc}
}

type createOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type orgResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Create handles POST /api/v1/orgs
func (h *OrgHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	var req createOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}

	if req.Name == "" || req.Slug == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "name and slug are required"})
		return
	}

	org, err := h.orgSvc.CreateOrg(r.Context(), req.Name, req.Slug, user.ID)
	if err != nil {
		if containsMsg(err, "already taken") {
			writeJSON(w, http.StatusConflict, errorResponse{Error: err.Error()})
			return
		}
		if containsMsg(err, "required") || containsMsg(err, "must be") {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, orgResponse{ID: org.ID, Name: org.Name, Slug: org.Slug})
}

// List handles GET /api/v1/orgs
func (h *OrgHandler) List(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "authentication required"})
		return
	}

	orgs, err := h.orgSvc.ListUserOrgs(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	var resp []orgResponse
	for _, o := range orgs {
		resp = append(resp, orgResponse{ID: o.ID, Name: o.Name, Slug: o.Slug})
	}
	if resp == nil {
		resp = []orgResponse{} // Return empty array instead of null.
	}
	writeJSON(w, http.StatusOK, resp)
}

// GoogleLogin handles GET /auth/google/login
type GoogleLoginHandler struct {
	provider     *auth.GoogleProvider
	sessionStore auth.SessionStore
}

// NewGoogleLoginHandler creates a new GoogleLoginHandler.
func NewGoogleLoginHandler(provider *auth.GoogleProvider, sessionStore auth.SessionStore) *GoogleLoginHandler {
	return &GoogleLoginHandler{provider: provider, sessionStore: sessionStore}
}

// Login redirects to Google's consent screen.
func (h *GoogleLoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	// In production, use a proper CSRF state token.
	url := h.provider.LoginURL("state-placeholder")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Callback handles the OAuth callback.
func (h *GoogleLoginHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "missing authorization code"})
		return
	}

	user, err := h.provider.HandleCallback(r.Context(), code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "OAuth authentication failed"})
		return
	}

	if err := h.sessionStore.SetUserID(w, r, user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to create session"})
		return
	}

	// Redirect to dashboard after login.
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
