package dashboard

import (
	"context"

	db "github.com/diuk/raiseup/db/generated"
)

// Store is the persistence interface used by Service.
type Store interface {
	GetResidentSummary(ctx context.Context) (db.GetDashboardResidentSummaryRow, error)
	GetComplaintSummary(ctx context.Context) (db.GetDashboardComplaintSummaryRow, error)
	GetFinanceSummary(ctx context.Context) (db.GetDashboardFinanceSummaryRow, error)
	GetCurrentDuesSummary(ctx context.Context) (db.GetDashboardCurrentDuesSummaryRow, error)
	GetAnnouncementSummary(ctx context.Context) (db.GetDashboardAnnouncementSummaryRow, error)
	GetActivitySummary(ctx context.Context) (db.GetDashboardActivitySummaryRow, error)
}

// Service implements dashboard aggregation use cases.
type Service struct {
	store Store
}

// NewService creates a dashboard service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// Summary assembles the dashboard aggregation response.
func (s *Service) Summary(ctx context.Context) (*Summary, error) {
	residents, err := s.store.GetResidentSummary(ctx)
	if err != nil {
		return nil, err
	}
	complaints, err := s.store.GetComplaintSummary(ctx)
	if err != nil {
		return nil, err
	}
	finance, err := s.store.GetFinanceSummary(ctx)
	if err != nil {
		return nil, err
	}
	dues, err := s.store.GetCurrentDuesSummary(ctx)
	if err != nil {
		return nil, err
	}
	announcements, err := s.store.GetAnnouncementSummary(ctx)
	if err != nil {
		return nil, err
	}
	activities, err := s.store.GetActivitySummary(ctx)
	if err != nil {
		return nil, err
	}

	return &Summary{
		Residents: ResidentSummary{
			Total:  residents.Total,
			Male:   residents.Male,
			Female: residents.Female,
		},
		Complaints: ComplaintSummary{
			Total:    complaints.Total,
			Baru:     complaints.Baru,
			Diproses: complaints.Diproses,
			Selesai:  complaints.Selesai,
			Ditolak:  complaints.Ditolak,
		},
		Finance: FinanceSummary{
			Income:  finance.Income,
			Expense: finance.Expense,
			Balance: finance.Balance,
		},
		Dues:          mapDuesSummary(dues),
		Announcements: AnnouncementSummary{Draft: announcements.Draft, Published: announcements.Published},
		Activities:    ActivitySummary{Upcoming: activities.Upcoming, Total: activities.Total},
	}, nil
}

func mapDuesSummary(row db.GetDashboardCurrentDuesSummaryRow) DuesSummary {
	out := DuesSummary{
		ResidentCount:    row.ResidentCount,
		PaidCount:        row.PaidCount,
		UnpaidCount:      row.UnpaidCount,
		ExpectedTotal:    row.ExpectedTotal,
		CollectedTotal:   row.CollectedTotal,
		OutstandingTotal: row.OutstandingTotal,
	}
	if row.Year != nil && row.Month != nil && row.Half != nil {
		out.Period = &DuesPeriodRef{
			Year:  *row.Year,
			Month: *row.Month,
			Half:  *row.Half,
		}
	}
	return out
}
