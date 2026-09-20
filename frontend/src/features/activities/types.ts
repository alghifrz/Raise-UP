import type { PaginationMeta } from '../../lib/api/types'

export type Activity = {
  id: string
  name: string
  description: string
  date: string
  reminder_days_before: number
  reminder_time: string | null
  reminder_message: string
  reminder_scheduled_at: string | null
  reminder_schedule_version: number
  reminder_n8n_execution_id: string | null
  reminder_sent_at: string | null
  created_at: string
  updated_at: string
}

export type ActivityFilters = {
  page: number
  page_size: number
  search: string
  from: string
  to: string
}

export type ActivityListResponse = {
  items: Activity[]
  meta: PaginationMeta
}

export type CreateActivityRequest = {
  name: string
  description: string
  date: string
  reminder_days_before?: number
  reminder_time?: string
  reminder_message?: string
}

export type UpdateActivityRequest = {
  name?: string
  description?: string
  date?: string
  reminder_days_before?: number
  reminder_time?: string
  reminder_message?: string
}

export function formatReminderSummary(activity: Activity): string {
  const days = activity.reminder_days_before
  const time = activity.reminder_time
  if (!time && !activity.reminder_message && days === 0) {
    return 'Tidak dikonfigurasi'
  }
  const parts = [`H-${days}`]
  if (time) {
    parts.push(time)
  }
  return parts.join(' · ')
}
