package dashboard

import (
	"context"
	"fmt"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides dashboard aggregation persistence via sqlc.
type Repository struct {
	q *db.Queries
}

// NewRepository creates a dashboard repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

// GetResidentSummary returns resident gender aggregates.
func (r *Repository) GetResidentSummary(ctx context.Context) (db.GetDashboardResidentSummaryRow, error) {
	row, err := r.q.GetDashboardResidentSummary(ctx)
	if err != nil {
		return db.GetDashboardResidentSummaryRow{}, fmt.Errorf("dashboard resident summary: %w", err)
	}
	return row, nil
}

// GetComplaintSummary returns complaint status aggregates.
func (r *Repository) GetComplaintSummary(ctx context.Context) (db.GetDashboardComplaintSummaryRow, error) {
	row, err := r.q.GetDashboardComplaintSummary(ctx)
	if err != nil {
		return db.GetDashboardComplaintSummaryRow{}, fmt.Errorf("dashboard complaint summary: %w", err)
	}
	return row, nil
}

// GetFinanceSummary returns cash ledger aggregates.
func (r *Repository) GetFinanceSummary(ctx context.Context) (db.GetDashboardFinanceSummaryRow, error) {
	row, err := r.q.GetDashboardFinanceSummary(ctx)
	if err != nil {
		return db.GetDashboardFinanceSummaryRow{}, fmt.Errorf("dashboard finance summary: %w", err)
	}
	return row, nil
}

// GetCurrentDuesSummary returns aggregates for the latest calendar dues period.
func (r *Repository) GetCurrentDuesSummary(ctx context.Context) (db.GetDashboardCurrentDuesSummaryRow, error) {
	row, err := r.q.GetDashboardCurrentDuesSummary(ctx)
	if err != nil {
		return db.GetDashboardCurrentDuesSummaryRow{}, fmt.Errorf("dashboard dues summary: %w", err)
	}
	return row, nil
}

// GetAnnouncementSummary returns announcement status aggregates.
func (r *Repository) GetAnnouncementSummary(ctx context.Context) (db.GetDashboardAnnouncementSummaryRow, error) {
	row, err := r.q.GetDashboardAnnouncementSummary(ctx)
	if err != nil {
		return db.GetDashboardAnnouncementSummaryRow{}, fmt.Errorf("dashboard announcement summary: %w", err)
	}
	return row, nil
}

// GetActivitySummary returns activity total and upcoming counts.
func (r *Repository) GetActivitySummary(ctx context.Context) (db.GetDashboardActivitySummaryRow, error) {
	row, err := r.q.GetDashboardActivitySummary(ctx)
	if err != nil {
		return db.GetDashboardActivitySummaryRow{}, fmt.Errorf("dashboard activity summary: %w", err)
	}
	return row, nil
}
