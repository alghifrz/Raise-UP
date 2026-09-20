package dashboard_test

import (
	"context"
	"errors"
	"testing"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/internal/dashboard"
)

type memoryStore struct {
	residents     db.GetDashboardResidentSummaryRow
	complaints    db.GetDashboardComplaintSummaryRow
	finance       db.GetDashboardFinanceSummaryRow
	dues          db.GetDashboardCurrentDuesSummaryRow
	announcements db.GetDashboardAnnouncementSummaryRow
	activities    db.GetDashboardActivitySummaryRow

	residentsErr     error
	complaintsErr    error
	financeErr       error
	duesErr          error
	announcementsErr error
	activitiesErr    error
}

func (m *memoryStore) GetResidentSummary(context.Context) (db.GetDashboardResidentSummaryRow, error) {
	if m.residentsErr != nil {
		return db.GetDashboardResidentSummaryRow{}, m.residentsErr
	}
	return m.residents, nil
}
func (m *memoryStore) GetComplaintSummary(context.Context) (db.GetDashboardComplaintSummaryRow, error) {
	if m.complaintsErr != nil {
		return db.GetDashboardComplaintSummaryRow{}, m.complaintsErr
	}
	return m.complaints, nil
}
func (m *memoryStore) GetFinanceSummary(context.Context) (db.GetDashboardFinanceSummaryRow, error) {
	if m.financeErr != nil {
		return db.GetDashboardFinanceSummaryRow{}, m.financeErr
	}
	return m.finance, nil
}
func (m *memoryStore) GetCurrentDuesSummary(context.Context) (db.GetDashboardCurrentDuesSummaryRow, error) {
	if m.duesErr != nil {
		return db.GetDashboardCurrentDuesSummaryRow{}, m.duesErr
	}
	return m.dues, nil
}
func (m *memoryStore) GetAnnouncementSummary(context.Context) (db.GetDashboardAnnouncementSummaryRow, error) {
	if m.announcementsErr != nil {
		return db.GetDashboardAnnouncementSummaryRow{}, m.announcementsErr
	}
	return m.announcements, nil
}
func (m *memoryStore) GetActivitySummary(context.Context) (db.GetDashboardActivitySummaryRow, error) {
	if m.activitiesErr != nil {
		return db.GetDashboardActivitySummaryRow{}, m.activitiesErr
	}
	return m.activities, nil
}

func int32Ptr(v int32) *int32 { return &v }

func populatedStore() *memoryStore {
	return &memoryStore{
		residents:     db.GetDashboardResidentSummaryRow{Total: 125, Male: 64, Female: 61},
		complaints:    db.GetDashboardComplaintSummaryRow{Total: 18, Baru: 3, Diproses: 5, Selesai: 8, Ditolak: 2},
		finance:       db.GetDashboardFinanceSummaryRow{Income: 12500000, Expense: 4300000, Balance: 8200000},
		dues:          db.GetDashboardCurrentDuesSummaryRow{Year: int32Ptr(2026), Month: int32Ptr(9), Half: int32Ptr(2), ResidentCount: 125, PaidCount: 100, UnpaidCount: 25, ExpectedTotal: 6250000, CollectedTotal: 5000000, OutstandingTotal: 1250000},
		announcements: db.GetDashboardAnnouncementSummaryRow{Draft: 2, Published: 8},
		activities:    db.GetDashboardActivitySummaryRow{Upcoming: 4, Total: 12},
	}
}

func TestSummaryPopulated(t *testing.T) {
	svc := dashboard.NewService(populatedStore())
	summary, err := svc.Summary(context.Background())
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Residents.Total != 125 || summary.Residents.Male != 64 || summary.Residents.Female != 61 {
		t.Fatalf("residents: %+v", summary.Residents)
	}
	if summary.Complaints.Total != 18 || summary.Complaints.Baru != 3 || summary.Complaints.Diproses != 5 ||
		summary.Complaints.Selesai != 8 || summary.Complaints.Ditolak != 2 {
		t.Fatalf("complaints: %+v", summary.Complaints)
	}
	if summary.Finance.Income != 12500000 || summary.Finance.Expense != 4300000 || summary.Finance.Balance != 8200000 {
		t.Fatalf("finance: %+v", summary.Finance)
	}
	if summary.Dues.Period == nil || summary.Dues.Period.Year != 2026 || summary.Dues.Period.Month != 9 || summary.Dues.Period.Half != 2 {
		t.Fatalf("dues period: %+v", summary.Dues.Period)
	}
	if summary.Dues.PaidCount != 100 || summary.Dues.UnpaidCount != 25 || summary.Dues.OutstandingTotal != 1250000 {
		t.Fatalf("dues metrics: %+v", summary.Dues)
	}
	if summary.Announcements.Draft != 2 || summary.Announcements.Published != 8 {
		t.Fatalf("announcements: %+v", summary.Announcements)
	}
	if summary.Activities.Upcoming != 4 || summary.Activities.Total != 12 {
		t.Fatalf("activities: %+v", summary.Activities)
	}
}

func TestSummaryEmpty(t *testing.T) {
	svc := dashboard.NewService(&memoryStore{})
	summary, err := svc.Summary(context.Background())
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Residents != (dashboard.ResidentSummary{}) {
		t.Fatalf("expected zero residents: %+v", summary.Residents)
	}
	if summary.Complaints != (dashboard.ComplaintSummary{}) {
		t.Fatalf("expected zero complaints: %+v", summary.Complaints)
	}
	if summary.Finance != (dashboard.FinanceSummary{}) {
		t.Fatalf("expected zero finance: %+v", summary.Finance)
	}
	if summary.Dues.Period != nil || summary.Dues.ResidentCount != 0 || summary.Dues.ExpectedTotal != 0 {
		t.Fatalf("expected null period and zero dues: %+v", summary.Dues)
	}
	if summary.Announcements != (dashboard.AnnouncementSummary{}) {
		t.Fatalf("expected zero announcements: %+v", summary.Announcements)
	}
	if summary.Activities != (dashboard.ActivitySummary{}) {
		t.Fatalf("expected zero activities: %+v", summary.Activities)
	}
}

func TestDuesPeriodSelectionAndPaymentStates(t *testing.T) {
	cases := []struct {
		name string
		dues db.GetDashboardCurrentDuesSummaryRow
		want dashboard.DuesSummary
	}{
		{
			name: "current period selected",
			dues: db.GetDashboardCurrentDuesSummaryRow{
				Year: int32Ptr(2026), Month: int32Ptr(9), Half: int32Ptr(2),
				ResidentCount: 10, PaidCount: 4, UnpaidCount: 6,
				ExpectedTotal: 500000, CollectedTotal: 200000, OutstandingTotal: 300000,
			},
			want: dashboard.DuesSummary{
				Period:        &dashboard.DuesPeriodRef{Year: 2026, Month: 9, Half: 2},
				ResidentCount: 10, PaidCount: 4, UnpaidCount: 6,
				ExpectedTotal: 500000, CollectedTotal: 200000, OutstandingTotal: 300000,
			},
		},
		{
			name: "all unpaid",
			dues: db.GetDashboardCurrentDuesSummaryRow{
				Year: int32Ptr(2026), Month: int32Ptr(9), Half: int32Ptr(1),
				ResidentCount: 5, PaidCount: 0, UnpaidCount: 5,
				ExpectedTotal: 250000, CollectedTotal: 0, OutstandingTotal: 250000,
			},
			want: dashboard.DuesSummary{
				Period:        &dashboard.DuesPeriodRef{Year: 2026, Month: 9, Half: 1},
				ResidentCount: 5, PaidCount: 0, UnpaidCount: 5,
				ExpectedTotal: 250000, CollectedTotal: 0, OutstandingTotal: 250000,
			},
		},
		{
			name: "all paid",
			dues: db.GetDashboardCurrentDuesSummaryRow{
				Year: int32Ptr(2026), Month: int32Ptr(8), Half: int32Ptr(2),
				ResidentCount: 5, PaidCount: 5, UnpaidCount: 0,
				ExpectedTotal: 250000, CollectedTotal: 250000, OutstandingTotal: 0,
			},
			want: dashboard.DuesSummary{
				Period:        &dashboard.DuesPeriodRef{Year: 2026, Month: 8, Half: 2},
				ResidentCount: 5, PaidCount: 5, UnpaidCount: 0,
				ExpectedTotal: 250000, CollectedTotal: 250000, OutstandingTotal: 0,
			},
		},
		{
			name: "no period",
			dues: db.GetDashboardCurrentDuesSummaryRow{},
			want: dashboard.DuesSummary{},
		},
		{
			name: "period with zero residents",
			dues: db.GetDashboardCurrentDuesSummaryRow{
				Year: int32Ptr(2026), Month: int32Ptr(9), Half: int32Ptr(2),
			},
			want: dashboard.DuesSummary{
				Period: &dashboard.DuesPeriodRef{Year: 2026, Month: 9, Half: 2},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &memoryStore{dues: tc.dues}
			summary, err := dashboard.NewService(store).Summary(context.Background())
			if err != nil {
				t.Fatalf("summary: %v", err)
			}
			got := summary.Dues
			if (got.Period == nil) != (tc.want.Period == nil) {
				t.Fatalf("period nil mismatch: got %+v want %+v", got.Period, tc.want.Period)
			}
			if got.Period != nil && *got.Period != *tc.want.Period {
				t.Fatalf("period: got %+v want %+v", *got.Period, *tc.want.Period)
			}
			got.Period, tc.want.Period = nil, nil
			if got != tc.want {
				t.Fatalf("dues: got %+v want %+v", got, tc.want)
			}
		})
	}
}

func TestFinanceBalance(t *testing.T) {
	store := &memoryStore{
		finance: db.GetDashboardFinanceSummaryRow{Income: 1000, Expense: 400, Balance: 600},
	}
	summary, err := dashboard.NewService(store).Summary(context.Background())
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Finance.Balance != 600 {
		t.Fatalf("balance: %+v", summary.Finance)
	}
}

func TestActivityUpcomingVsTotal(t *testing.T) {
	store := &memoryStore{
		activities: db.GetDashboardActivitySummaryRow{Upcoming: 2, Total: 7},
	}
	summary, err := dashboard.NewService(store).Summary(context.Background())
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Activities.Upcoming != 2 || summary.Activities.Total != 7 {
		t.Fatalf("activities: %+v", summary.Activities)
	}
}

func TestRepositoryErrorPropagation(t *testing.T) {
	boom := errors.New("db down")
	cases := []struct {
		name  string
		store *memoryStore
	}{
		{"residents", &memoryStore{residentsErr: boom}},
		{"complaints", &memoryStore{complaintsErr: boom}},
		{"finance", &memoryStore{financeErr: boom}},
		{"dues", &memoryStore{duesErr: boom}},
		{"announcements", &memoryStore{announcementsErr: boom}},
		{"activities", &memoryStore{activitiesErr: boom}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := dashboard.NewService(tc.store).Summary(context.Background())
			if !errors.Is(err, boom) {
				t.Fatalf("expected %v, got %v", boom, err)
			}
		})
	}
}
