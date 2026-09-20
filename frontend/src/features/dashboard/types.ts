export type ResidentSummary = {
  total: number
  male: number
  female: number
}

export type ComplaintSummary = {
  total: number
  baru: number
  diproses: number
  selesai: number
  ditolak: number
}

export type FinanceSummary = {
  income: number
  expense: number
  balance: number
}

export type DuesPeriodSummary = {
  year: number
  month: number
  half: number
}

export type DuesSummary = {
  period: DuesPeriodSummary | null
  resident_count: number
  paid_count: number
  unpaid_count: number
  expected_total: number
  collected_total: number
  outstanding_total: number
}

export type AnnouncementSummary = {
  draft: number
  published: number
}

export type ActivitySummary = {
  upcoming: number
  total: number
}

export type DashboardSummary = {
  residents: ResidentSummary
  complaints: ComplaintSummary
  finance: FinanceSummary
  dues: DuesSummary
  announcements: AnnouncementSummary
  activities: ActivitySummary
}
