package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/taverns-red/tavern-url/internal/auth"
	"github.com/taverns-red/tavern-url/internal/service"
	"github.com/taverns-red/tavern-url/templates"
)

// APIKeyHandler handles API key management requests.
type APIKeyHandler struct {
	apiKeySvc *service.APIKeyService
}

// NewAPIKeyHandler creates a new APIKeyHandler.
func NewAPIKeyHandler(apiKeySvc *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeySvc: apiKeySvc}
}

type createKeyRequest struct {
	Name  string `json:"name"`
	OrgID int64  `json:"org_id"`
}

type createKeyResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	KeyPrefix string `json:"key_prefix"`
	RawKey    string `json:"raw_key"` // Shown only once!
}

type keyResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	KeyPrefix string `json:"key_prefix"`
	LastUsed  string `json:"last_used,omitempty"`
	CreatedAt string `json:"created_at"`
}

// Create handles POST /api/v1/keys
func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "not authenticated"})
		return
	}

	var req createKeyRequest
	isForm := strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")

	if isForm {
		if err := r.ParseForm(); err != nil {
			writeFormError(w, http.StatusBadRequest, "Invalid form data.")
			return
		}
		req.Name = r.FormValue("name")
		orgID := r.FormValue("org_id")
		if orgID != "" {
			req.OrgID, _ = strconv.ParseInt(orgID, 10, 64)
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
			return
		}
	}

	if req.Name == "" {
		if isForm {
			writeFormError(w, http.StatusBadRequest, "Key name is required.")
			return
		}
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "name is required"})
		return
	}

	// Default org_id to 0 for personal keys.
	rawKey, key, err := h.apiKeySvc.CreateKey(r.Context(), userID, req.OrgID, req.Name)
	if err != nil {
		if isForm {
			writeFormError(w, http.StatusInternalServerError, "Failed to create API key.")
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to create API key"})
		return
	}

	// HTMX: show the raw key and refresh key list.
	if isForm {
		keys, _ := h.apiKeySvc.ListKeys(r.Context(), userID)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		// Show raw key alert first, then render key list.
		fmt.Fprintf(w, `<div class="alert" style="background:var(--color-success-bg,#d4edda);border:1px solid var(--color-success-border,#c3e6cb);padding:var(--space-4);border-radius:var(--radius-md);margin-bottom:var(--space-4);word-break:break-all"><strong>API Key created!</strong> Copy it now — it won't be shown again:<br/><code style="font-size:var(--font-size-sm)">%s</code></div>`, rawKey)
		_ = key // suppress unused
		templates.APIKeyList(keys).Render(r.Context(), w)
		return
	}

	writeJSON(w, http.StatusCreated, createKeyResponse{
		ID:        key.ID,
		Name:      key.Name,
		KeyPrefix: key.KeyPrefix,
		RawKey:    rawKey,
	})
}

// List handles GET /api/v1/keys
func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "not authenticated"})
		return
	}

	keys, err := h.apiKeySvc.ListKeys(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to list keys"})
		return
	}

	resp := make([]keyResponse, 0, len(keys))
	for _, k := range keys {
		kr := keyResponse{
			ID:        k.ID,
			Name:      k.Name,
			KeyPrefix: k.KeyPrefix,
			CreatedAt: k.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if k.LastUsedAt != nil {
			kr.LastUsed = k.LastUsedAt.Format("2006-01-02T15:04:05Z")
		}
		resp = append(resp, kr)
	}

	writeJSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /api/v1/keys/{id}
func (h *APIKeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "not authenticated"})
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid key ID"})
		return
	}

	if err := h.apiKeySvc.DeleteKey(r.Context(), id, userID); err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "key not found"})
		return
	}

	// HTMX: return updated key list.
	if r.Header.Get("HX-Request") == "true" {
		keys, _ := h.apiKeySvc.ListKeys(r.Context(), userID)
		templates.APIKeyList(keys).Render(r.Context(), w)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "key deleted"})
}
