package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/taverns-red/tavern-url/internal/service"
)

// ExportHandler handles CSV export requests.
type ExportHandler struct {
	analyticsSvc *service.AnalyticsService
}

// NewExportHandler creates a new ExportHandler.
func NewExportHandler(analyticsSvc *service.AnalyticsService) *ExportHandler {
	return &ExportHandler{analyticsSvc: analyticsSvc}
}

// ExportCSV handles GET /api/v1/links/{id}/analytics/export
func (h *ExportHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"analytics-%d.csv\"", id))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Daily clicks.
	writer.Write([]string{"Section: Daily Clicks"})
	writer.Write([]string{"Date", "Clicks"})
	for _, day := range summary.ByDay {
		writer.Write([]string{day.Date, fmt.Sprintf("%d", day.Count)})
	}
	writer.Write([]string{""})

	// Countries.
	writer.Write([]string{"Section: Countries"})
	writer.Write([]string{"Country", "Clicks"})
	for _, c := range summary.ByCountry {
		writer.Write([]string{c.Category, fmt.Sprintf("%d", c.Count)})
	}
	writer.Write([]string{""})

	// Devices.
	writer.Write([]string{"Section: Devices"})
	writer.Write([]string{"Device", "Clicks"})
	for _, d := range summary.ByDevice {
		writer.Write([]string{d.Category, fmt.Sprintf("%d", d.Count)})
	}
	writer.Write([]string{""})

	// Referrers.
	writer.Write([]string{"Section: Referrers"})
	writer.Write([]string{"Referrer", "Clicks"})
	for _, ref := range summary.ByReferrer {
		writer.Write([]string{ref.Category, fmt.Sprintf("%d", ref.Count)})
	}
}
