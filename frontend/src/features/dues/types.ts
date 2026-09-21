import type { PaginationMeta } from '../../lib/api/types'
import { MONTH_NAMES_ID } from '../../lib/utils'

export type DuesPeriod = {
  id: string
  year: number
  month: number
  half: number
  amount: number
  created_at: string
}

export type DuesPayment = {
  id: string
  period_id: string
  resident_id: string
  resident_name: string
  amount: number
  paid_at: string
  created_at: string
}

export type DuesPeriodSummary = {
  period_id: string
  year: number
  month: number
  half: number
  expected_amount: number
  resident_count: number
  paid_count: number
  unpaid_count: number
  expected_total: number
  collected_total: number
  outstanding_total: number
}

export type DuesReminderResult = {
  total: number
  sent: number
  failed: number
}

export type DuesPaymentStatusValue = 'PAID' | 'UNPAID'

export type ResidentPaymentStatus = {
  resident_id: string
  resident_name: string
  phone: string
  status: DuesPaymentStatusValue
  payment_id: string | null
  amount: number | null
  paid_at: string | null
}

export type DuesPeriodFilters = {
  page: number
  page_size: number
  year: string
  month: string
  half: string
}

export type DuesPeriodStatusFilters = {
  page: number
  page_size: number
  status: DuesPaymentStatusValue | ''
  search: string
}

export type DuesPeriodListResponse = {
  items: DuesPeriod[]
  meta: PaginationMeta
}

export type DuesStatusListResponse = {
  items: ResidentPaymentStatus[]
  meta: PaginationMeta
}

export type CreateDuesPeriodRequest = {
  year: number
  month: number
  half: number
  amount: number
}

export type UpdateDuesPeriodRequest = {
  year?: number
  month?: number
  half?: number
  amount?: number
}

export type CreateDuesPaymentRequest = {
  period_id: string
  resident_id: string
  amount: number
  paid_at: string
}

export type UpdateDuesPaymentRequest = {
  period_id?: string
  resident_id?: string
  amount?: number
  paid_at?: string
}

export const MONTH_OPTIONS = MONTH_NAMES_ID.map((label, index) => ({
  value: String(index + 1),
  label,
}))

export const HALF_OPTIONS = [
  { value: '1', label: 'Half 1' },
  { value: '2', label: 'Half 2' },
]

export const DUES_STATUS_OPTIONS: Array<{ value: DuesPaymentStatusValue; label: string }> = [
  { value: 'PAID', label: 'Lunas' },
  { value: 'UNPAID', label: 'Belum Lunas' },
]

export function formatDuesPaymentStatus(status: DuesPaymentStatusValue): string {
  return DUES_STATUS_OPTIONS.find((item) => item.value === status)?.label ?? status
}
