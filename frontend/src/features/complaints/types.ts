import type { PaginationMeta } from '../../lib/api/types'

export type ComplaintStatus = 'BARU' | 'DIPROSES' | 'SELESAI' | 'DITOLAK'
export type ComplaintUrgency = 'PRIORITY' | 'MEDIUM' | 'NORMAL'

export type Complaint = {
  id: string
  ref: string
  resident_id: string | null
  resident_name: string
  phone: string
  block: string
  category: string
  urgency: ComplaintUrgency
  status: ComplaintStatus
  message: string
  received_at: string
  updated_at: string
}

export type ComplaintFilters = {
  page: number
  page_size: number
  search: string
  status: ComplaintStatus | ''
  urgency: ComplaintUrgency | ''
  category: string
}

export type ComplaintListResponse = {
  items: Complaint[]
  meta: PaginationMeta
}

export type CreateComplaintRequest = {
  resident_id?: string
  resident_name?: string
  phone?: string
  block: string
  category: string
  urgency: ComplaintUrgency
  message: string
}

export type UpdateComplaintRequest = {
  resident_id?: string | null
  block?: string
  category?: string
  urgency?: ComplaintUrgency
  message?: string
}

export type UpdateComplaintStatusRequest = {
  status: ComplaintStatus
}

export const COMPLAINT_STATUS_OPTIONS: Array<{ value: ComplaintStatus; label: string }> = [
  { value: 'BARU', label: 'Baru' },
  { value: 'DIPROSES', label: 'Diproses' },
  { value: 'SELESAI', label: 'Selesai' },
  { value: 'DITOLAK', label: 'Ditolak' },
]

export const COMPLAINT_URGENCY_OPTIONS: Array<{ value: ComplaintUrgency; label: string }> = [
  { value: 'PRIORITY', label: 'Prioritas' },
  { value: 'MEDIUM', label: 'Sedang' },
  { value: 'NORMAL', label: 'Normal' },
]

export function formatComplaintStatus(status: ComplaintStatus): string {
  return COMPLAINT_STATUS_OPTIONS.find((item) => item.value === status)?.label ?? status
}

export function formatComplaintUrgency(urgency: ComplaintUrgency): string {
  return COMPLAINT_URGENCY_OPTIONS.find((item) => item.value === urgency)?.label ?? urgency
}

export function nextComplaintStatusActions(
  status: ComplaintStatus,
): Array<{ status: ComplaintStatus; label: string }> {
  switch (status) {
    case 'BARU':
      return [
        { status: 'DIPROSES', label: 'Proses' },
        { status: 'DITOLAK', label: 'Tolak' },
      ]
    case 'DIPROSES':
      return [
        { status: 'SELESAI', label: 'Selesaikan' },
        { status: 'DITOLAK', label: 'Tolak' },
      ]
    default:
      return []
  }
}
