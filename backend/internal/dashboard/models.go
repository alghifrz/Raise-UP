package dashboard

// Summary is the aggregated dashboard payload.
type Summary struct {
	Residents     ResidentSummary     `json:"residents"`
	Complaints    ComplaintSummary    `json:"complaints"`
	Finance       FinanceSummary      `json:"finance"`
	Dues          DuesSummary         `json:"dues"`
	Announcements AnnouncementSummary `json:"announcements"`
	Activities    ActivitySummary     `json:"activities"`
}

// ResidentSummary aggregates resident counts by gender.
type ResidentSummary struct {
	Total  int64 `json:"total"`
	Male   int64 `json:"male"`
	Female int64 `json:"female"`
}

// ComplaintSummary aggregates complaints by status.
type ComplaintSummary struct {
	Total    int64 `json:"total"`
	Baru     int64 `json:"baru"`
	Diproses int64 `json:"diproses"`
	Selesai  int64 `json:"selesai"`
	Ditolak  int64 `json:"ditolak"`
}

// FinanceSummary aggregates cash ledger totals (IDR int64).
type FinanceSummary struct {
	Income  int64 `json:"income"`
	Expense int64 `json:"expense"`
	Balance int64 `json:"balance"`
}

// DuesPeriodRef identifies the current dues period (calendar order).
type DuesPeriodRef struct {
	Year  int32 `json:"year"`
	Month int32 `json:"month"`
	Half  int32 `json:"half"`
}

// DuesSummary aggregates collection status for the current dues period.
// Period is null when no dues period exists.
type DuesSummary struct {
	Period           *DuesPeriodRef `json:"period"`
	ResidentCount    int64          `json:"resident_count"`
	PaidCount        int64          `json:"paid_count"`
	UnpaidCount      int64          `json:"unpaid_count"`
	ExpectedTotal    int64          `json:"expected_total"`
	CollectedTotal   int64          `json:"collected_total"`
	OutstandingTotal int64          `json:"outstanding_total"`
}

// AnnouncementSummary aggregates announcements by status.
type AnnouncementSummary struct {
	Draft     int64 `json:"draft"`
	Published int64 `json:"published"`
}

// ActivitySummary aggregates activity totals and upcoming count.
type ActivitySummary struct {
	Upcoming int64 `json:"upcoming"`
	Total    int64 `json:"total"`
}
