import { apiRequest, apiRequestWithMeta, toQueryString } from '../../lib/api/client'
import type { PaginationMeta } from '../../lib/api/types'
import type {
  Announcement,
  AnnouncementFilters,
  AnnouncementListResponse,
  AnnouncementSummary,
  CreateAnnouncementRequest,
  UpdateAnnouncementRequest,
  UpdateAnnouncementStatusRequest,
} from './types'

export function listAnnouncements(
  filters: AnnouncementFilters,
  signal?: AbortSignal,
): Promise<AnnouncementListResponse> {
  const qs = toQueryString({
    page: filters.page,
    page_size: filters.page_size,
    search: filters.search || undefined,
    status: filters.status || undefined,
    visibility: filters.visibility || undefined,
    category: filters.category || undefined,
  })

  return apiRequestWithMeta<AnnouncementSummary[], PaginationMeta>(
    `/api/v1/announcements${qs}`,
    { signal },
  ).then((result) => ({
    items: result.data,
    meta: result.meta,
  }))
}

export function getAnnouncement(id: string, signal?: AbortSignal): Promise<Announcement> {
  return apiRequest<Announcement>(`/api/v1/announcements/${id}`, { signal })
}

export function createAnnouncement(payload: CreateAnnouncementRequest): Promise<Announcement> {
  return apiRequest<Announcement>('/api/v1/announcements', {
    method: 'POST',
    body: payload,
  })
}

export function updateAnnouncement(
  id: string,
  payload: UpdateAnnouncementRequest,
): Promise<Announcement> {
  return apiRequest<Announcement>(`/api/v1/announcements/${id}`, {
    method: 'PATCH',
    body: payload,
  })
}

export function updateAnnouncementStatus(
  id: string,
  payload: UpdateAnnouncementStatusRequest,
): Promise<Announcement> {
  return apiRequest<Announcement>(`/api/v1/announcements/${id}/status`, {
    method: 'PATCH',
    body: payload,
  })
}

export function deleteAnnouncement(id: string): Promise<void> {
  return apiRequest<void>(`/api/v1/announcements/${id}`, {
    method: 'DELETE',
  })
}
