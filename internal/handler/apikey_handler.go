package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/taverns-red/tavern-url/internal/auth"
	"github.com/taverns-red/tavern-url/internal/service"
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "name is required"})
		return
	}
	if req.OrgID == 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "org_id is required"})
		return
	}

	rawKey, key, err := h.apiKeySvc.CreateKey(r.Context(), userID, req.OrgID, req.Name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to create API key"})
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

	writeJSON(w, http.StatusOK, map[string]string{"message": "key deleted"})
}
