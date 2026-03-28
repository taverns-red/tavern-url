package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/taverns-red/tavern-url/internal/repository"
	"github.com/taverns-red/tavern-url/internal/service"
	"github.com/taverns-red/tavern-url/templates"
)

// LinkHandler handles HTTP requests for link operations.
type LinkHandler struct {
	svc          *service.LinkService
	analyticsSvc *service.AnalyticsService
	baseURL      string
}

// NewLinkHandler creates a new LinkHandler.
func NewLinkHandler(svc *service.LinkService, analyticsSvc *service.AnalyticsService, baseURL string) *LinkHandler {
	return &LinkHandler{svc: svc, analyticsSvc: analyticsSvc, baseURL: baseURL}
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
	isForm := strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")

	if isForm {
		if err := r.ParseForm(); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid form data"})
			return
		}
		req.URL = r.FormValue("url")
		slug := r.FormValue("slug")
		if slug != "" {
			req.Slug = &slug
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
			return
		}
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

	// For HTMX requests, re-render the full link list.
	if isForm {
		links, _ := h.svc.ListLinks(r.Context())
		_ = link // used above
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		// Render a simple HTML snippet showing the new link.
		fmt.Fprintf(w, `<div id="link-list"><div style="display:flex;flex-direction:column;gap:var(--space-3)">`)
		for _, l := range links {
			fmt.Fprintf(w, `<div class="card" style="display:flex;align-items:center;gap:var(--space-4);padding:var(--space-4) var(--space-6)"><div style="flex:1;min-width:0"><div style="display:flex;align-items:center;gap:var(--space-2);margin-bottom:var(--space-1)"><a href="%s/%s" target="_blank" style="font-weight:var(--font-weight-semibold);font-size:var(--font-size-base)">%s/%s</a></div><p style="color:var(--color-text-secondary);font-size:var(--font-size-sm);overflow:hidden;text-overflow:ellipsis;white-space:nowrap">↳ %s</p></div><div style="display:flex;align-items:center;gap:var(--space-2)"><span class="text-sm text-muted">%s</span><button class="btn btn-ghost btn-danger" style="padding:var(--space-1) var(--space-2);min-height:auto;font-size:var(--font-size-xs)" hx-delete="/api/v1/links/%d" hx-target="#link-list" hx-swap="outerHTML" hx-confirm="Delete this link?">Delete</button></div></div>`,
				h.baseURL, l.Slug, h.baseURL, l.Slug, l.OriginalURL, l.CreatedAt.Format("Jan 2, 2006"), l.ID)
		}
		fmt.Fprintf(w, `</div></div>`)
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

	// Check expiration and click limits.
	if link.IsExpired() || link.IsExhausted() {
		http.Error(w, "this link has expired or reached its click limit", http.StatusGone)
		return
	}

	// Record click asynchronously (non-blocking).
	if h.analyticsSvc != nil {
		h.analyticsSvc.RecordClick(r.Context(), link.ID, r)
	}

	http.Redirect(w, r, link.OriginalURL, http.StatusFound)
}

// List handles GET /api/v1/links (optional ?q= search filter)
func (h *LinkHandler) List(w http.ResponseWriter, r *http.Request) {
	links, err := h.svc.ListLinks(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	// Filter by search query if provided.
	query := r.URL.Query().Get("q")
	var filtered []createLinkResponse
	for _, l := range links {
		if query != "" && !containsCI(l.Slug, query) && !containsCI(l.OriginalURL, query) {
			continue
		}
		filtered = append(filtered, createLinkResponse{
			ID:          l.ID,
			Slug:        l.Slug,
			ShortURL:    h.baseURL + "/" + l.Slug,
			OriginalURL: l.OriginalURL,
		})
	}
	if filtered == nil {
		filtered = []createLinkResponse{}
	}
	writeJSON(w, http.StatusOK, filtered)
}

type updateLinkRequest struct {
	URL  *string `json:"url,omitempty"`
	Slug *string `json:"slug,omitempty"`
}

// Update handles PUT /api/v1/links/{id}
func (h *LinkHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id := parseID(idStr)
	if id == 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid link ID"})
		return
	}

	var req updateLinkRequest
	isForm := strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")

	if isForm {
		if err := r.ParseForm(); err != nil {
			writeFormError(w, http.StatusBadRequest, "Invalid form data.")
			return
		}
		url := r.FormValue("url")
		slug := r.FormValue("slug")
		if url != "" {
			req.URL = &url
		}
		if slug != "" {
			req.Slug = &slug
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
			return
		}
	}

	link, err := h.svc.UpdateLink(r.Context(), id, req.URL, req.Slug)
	if err != nil {
		if isForm {
			msg := "Failed to update link."
			if errors.Is(err, repository.ErrLinkNotFound) {
				msg = "Link not found."
			} else {
				msg = err.Error()
			}
			writeFormError(w, http.StatusBadRequest, msg)
			return
		}
		if errors.Is(err, repository.ErrLinkNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "link not found"})
			return
		}
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	// HTMX: return updated link list.
	if isForm {
		links, _ := h.svc.ListLinks(r.Context())
		templates.LinkList(links, h.baseURL).Render(r.Context(), w)
		return
	}

	writeJSON(w, http.StatusOK, createLinkResponse{
		ID:          link.ID,
		Slug:        link.Slug,
		ShortURL:    h.baseURL + "/" + link.Slug,
		OriginalURL: link.OriginalURL,
	})
}

type bulkCreateRequest struct {
	URLs []string `json:"urls"`
}

// BulkCreate handles POST /api/v1/links/bulk
func (h *LinkHandler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	var req bulkCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}

	if len(req.URLs) == 0 || len(req.URLs) > 100 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "provide 1-100 URLs"})
		return
	}

	var resp []createLinkResponse
	for _, u := range req.URLs {
		link, err := h.svc.CreateLink(r.Context(), u, nil)
		if err != nil {
			resp = append(resp, createLinkResponse{OriginalURL: u})
			continue
		}
		resp = append(resp, createLinkResponse{
			ID:          link.ID,
			Slug:        link.Slug,
			ShortURL:    h.baseURL + "/" + link.Slug,
			OriginalURL: link.OriginalURL,
		})
	}
	writeJSON(w, http.StatusCreated, resp)
}

// Delete handles DELETE /api/v1/links/{id}
func (h *LinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := parseID(chi.URLParam(r, "id"))
	if id == 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid link ID"})
		return
	}

	if err := h.svc.DeleteLink(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrLinkNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "link not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	// For HTMX requests, re-render the link list using the template.
	if r.Header.Get("HX-Request") == "true" {
		links, _ := h.svc.ListLinks(r.Context())
		templates.LinkList(links, h.baseURL).Render(r.Context(), w)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "link deleted"})
}

// parseID parses a string ID to int64 (returns 0 on invalid).
func parseID(s string) int64 {
	var id int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		id = id*10 + int64(c-'0')
	}
	return id
}

// containsCI performs case-insensitive substring search.
func containsCI(s, substr string) bool {
	return contains(strings.ToLower(s), strings.ToLower(substr))
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
