package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/taverns-red/tavern-url/internal/model"
)

// PgClickRepository implements ClickRepository using PostgreSQL.
type PgClickRepository struct {
	pool *pgxpool.Pool
}

// NewPgClickRepository creates a new PgClickRepository.
func NewPgClickRepository(pool *pgxpool.Pool) *PgClickRepository {
	return &PgClickRepository{pool: pool}
}

// Record inserts a click event.
func (r *PgClickRepository) Record(ctx context.Context, event *model.ClickEvent) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO click_events (link_id, clicked_at, country, device_category, referrer_domain)
		 VALUES ($1, NOW(), $2, $3, $4)`,
		event.LinkID, event.Country, event.DeviceCategory, event.ReferrerDomain,
	)
	return err
}

// GetSummary returns aggregate analytics for a link within a date range.
func (r *PgClickRepository) GetSummary(ctx context.Context, linkID int64, from, to time.Time) (*model.ClickSummary, error) {
	summary := &model.ClickSummary{}

	// Total clicks.
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM click_events WHERE link_id = $1 AND clicked_at BETWEEN $2 AND $3`,
		linkID, from, to,
	).Scan(&summary.TotalClicks)
	if err != nil {
		return nil, err
	}

	// By day.
	rows, err := r.pool.Query(ctx,
		`SELECT DATE(clicked_at) as day, COUNT(*) as count
		 FROM click_events WHERE link_id = $1 AND clicked_at BETWEEN $2 AND $3
		 GROUP BY day ORDER BY day`,
		linkID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var dc model.DayCount
		var day time.Time
		if err := rows.Scan(&day, &dc.Count); err != nil {
			return nil, err
		}
		dc.Date = day.Format("2006-01-02")
		summary.ByDay = append(summary.ByDay, dc)
	}

	// By country.
	summary.ByCountry, err = r.aggregateColumn(ctx, linkID, from, to, "country")
	if err != nil {
		return nil, err
	}

	// By device.
	summary.ByDevice, err = r.aggregateColumn(ctx, linkID, from, to, "device_category")
	if err != nil {
		return nil, err
	}

	// By referrer.
	summary.ByReferrer, err = r.aggregateColumn(ctx, linkID, from, to, "referrer_domain")
	if err != nil {
		return nil, err
	}

	return summary, nil
}

// aggregateColumn groups clicks by a column and returns counts.
func (r *PgClickRepository) aggregateColumn(ctx context.Context, linkID int64, from, to time.Time, column string) ([]model.CategoryCount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+column+`, COUNT(*) as count
		 FROM click_events WHERE link_id = $1 AND clicked_at BETWEEN $2 AND $3
		 GROUP BY `+column+` ORDER BY count DESC LIMIT 20`,
		linkID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.CategoryCount
	for rows.Next() {
		var cc model.CategoryCount
		if err := rows.Scan(&cc.Category, &cc.Count); err != nil {
			return nil, err
		}
		results = append(results, cc)
	}
	return results, rows.Err()
}

// GetTotalClicks returns the total click count for a link.
func (r *PgClickRepository) GetTotalClicks(ctx context.Context, linkID int64) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM click_events WHERE link_id = $1`, linkID,
	).Scan(&count)
	return count, err
}
