import type { PaginationMeta } from '../../lib/api/types'

export type AnnouncementVisibility = 'PUBLIC' | 'PRIVATE'
export type AnnouncementStatus = 'DRAFT' | 'PUBLISHED'

export type AnnouncementSummary = {
  id: string
  title: string
  excerpt: string
  body: string
  category: string
  visibility: AnnouncementVisibility
  status: AnnouncementStatus
  thumbnail_url: string | null
  author_id: string
  published_at: string | null
  created_at: string
  updated_at: string
}

export type Announcement = AnnouncementSummary & {
  recipient_ids: string[]
  delivery?: {
    total: number
    sent: number
    failed: number
  }
}

export type AnnouncementFilters = {
  page: number
  page_size: number
  search: string
  status: AnnouncementStatus | ''
  visibility: AnnouncementVisibility | ''
  category: string
}

export type AnnouncementListResponse = {
  items: AnnouncementSummary[]
  meta: PaginationMeta
}

export type CreateAnnouncementRequest = {
  title: string
  body: string
  visibility: AnnouncementVisibility
  thumbnail_url?: string | null
  recipient_ids: string[]
}

export type UpdateAnnouncementRequest = {
  title?: string
  body?: string
  visibility?: AnnouncementVisibility
  thumbnail_url?: string | null
  recipient_ids?: string[]
}

export type UpdateAnnouncementStatusRequest = {
  status: AnnouncementStatus
}

export const ANNOUNCEMENT_VISIBILITY_OPTIONS: Array<{
  value: AnnouncementVisibility
  label: string
}> = [
  { value: 'PUBLIC', label: 'Publik' },
  { value: 'PRIVATE', label: 'Privat' },
]

export const ANNOUNCEMENT_STATUS_OPTIONS: Array<{
  value: AnnouncementStatus
  label: string
}> = [
  { value: 'DRAFT', label: 'Draft' },
  { value: 'PUBLISHED', label: 'Terbit' },
]

export function formatAnnouncementVisibility(value: AnnouncementVisibility): string {
  return ANNOUNCEMENT_VISIBILITY_OPTIONS.find((item) => item.value === value)?.label ?? value
}

export function formatAnnouncementStatus(value: AnnouncementStatus): string {
  return ANNOUNCEMENT_STATUS_OPTIONS.find((item) => item.value === value)?.label ?? value
}

export function formatAuthorLabel(authorId: string): string {
  if (authorId.length <= 8) {
    return authorId
  }
  return `${authorId.slice(0, 8)}…`
}
