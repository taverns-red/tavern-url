package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/taverns-red/tavern-url/internal/repository"
	"github.com/taverns-red/tavern-url/internal/service"
)

// LinkHandler handles HTTP requests for link operations.
type LinkHandler struct {
	svc     *service.LinkService
	baseURL string
}

// NewLinkHandler creates a new LinkHandler.
func NewLinkHandler(svc *service.LinkService, baseURL string) *LinkHandler {
	return &LinkHandler{svc: svc, baseURL: baseURL}
}

// createLinkRequest is the JSON body for creating a short link.
type createLinkRequest struct {
	URL  string  `json:"url"`
	Slug *string `json:"slug,omitempty"`
}

// createLinkResponse is the JSON response after creating a short link.
type createLinkResponse struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// errorResponse is the standard JSON error response.
type errorResponse struct {
	Error string `json:"error"`
}

// Create handles POST /api/v1/links
func (h *LinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}

	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "url is required"})
		return
	}

	link, err := h.svc.CreateLink(r.Context(), req.URL, req.Slug)
	if err != nil {
		if errors.Is(err, repository.ErrSlugExists) || containsMsg(err, "already taken") {
			writeJSON(w, http.StatusConflict, errorResponse{Error: err.Error()})
			return
		}
		if containsMsg(err, "invalid") {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, createLinkResponse{
		ID:          link.ID,
		Slug:        link.Slug,
		ShortURL:    h.baseURL + "/" + link.Slug,
		OriginalURL: link.OriginalURL,
	})
}

// Redirect handles GET /{slug}
func (h *LinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	link, err := h.svc.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, repository.ErrLinkNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, link.OriginalURL, http.StatusFound)
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// containsMsg checks if an error message contains a substring.
func containsMsg(err error, msg string) bool {
	return err != nil && len(err.Error()) > 0 && contains(err.Error(), msg)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
