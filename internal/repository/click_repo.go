package repository

import (
	"context"
	"time"

	"github.com/taverns-red/tavern-url/internal/model"
)

// ClickRepository defines the interface for click event persistence.
type ClickRepository interface {
	// Record inserts a new click event.
	Record(ctx context.Context, event *model.ClickEvent) error

	// GetSummary returns aggregate analytics for a link within a date range.
	GetSummary(ctx context.Context, linkID int64, from, to time.Time) (*model.ClickSummary, error)

	// GetTotalClicks returns the total click count for a link.
	GetTotalClicks(ctx context.Context, linkID int64) (int64, error)
}
