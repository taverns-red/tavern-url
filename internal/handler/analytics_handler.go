package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/taverns-red/tavern-url/internal/service"
)

// AnalyticsHandler handles analytics and QR code HTTP requests.
type AnalyticsHandler struct {
	analyticsSvc *service.AnalyticsService
	qrSvc        *service.QRService
	linkSvc      *service.LinkService
	baseURL      string
}

// NewAnalyticsHandler creates a new AnalyticsHandler.
func NewAnalyticsHandler(analyticsSvc *service.AnalyticsService, qrSvc *service.QRService, linkSvc *service.LinkService, baseURL string) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsSvc: analyticsSvc, qrSvc: qrSvc, linkSvc: linkSvc, baseURL: baseURL}
}

// GetSummary handles GET /api/v1/links/{id}/analytics
func (h *AnalyticsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid link ID"})
		return
	}

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	summary, err := h.analyticsSvc.GetSummary(r.Context(), id, days)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to get analytics"})
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// QRCode handles GET /api/v1/links/{id}/qr
func (h *AnalyticsHandler) QRCode(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid link ID"})
		return
	}

	// Get the link to build the short URL.
	link, err := h.linkSvc.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "link not found"})
		return
	}

	shortURL := h.baseURL + "/" + link.Slug

	size := 256
	if s := r.URL.Query().Get("size"); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil {
			size = parsed
		}
	}
	fg := r.URL.Query().Get("fg")
	bg := r.URL.Query().Get("bg")

	data, err := h.qrSvc.GeneratePNG(shortURL, size, fg, bg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to generate QR code"})
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "inline; filename=\"qr-"+link.Slug+".png\"")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
