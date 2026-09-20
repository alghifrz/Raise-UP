import { apiRequest, apiRequestWithMeta, toQueryString } from '../../lib/api/client'
import type { PaginationMeta } from '../../lib/api/types'
import type { Activity } from '../activities/types'
import type { Announcement, AnnouncementSummary } from '../announcements/types'
import type { GalleryItem } from '../gallery/types'
import type { SiteSettings } from '../site-settings/types'
import type { VillageOfficial, VillageProfile } from '../village/types'

export function getPublicSiteSettings(signal?: AbortSignal): Promise<SiteSettings> {
  return apiRequest<SiteSettings>('/api/v1/public/site-settings', { auth: false, signal })
}

export function getPublicVillageProfile(signal?: AbortSignal): Promise<VillageProfile> {
  return apiRequest<VillageProfile>('/api/v1/public/village-profile', { auth: false, signal })
}

export function listPublicVillageOfficials(signal?: AbortSignal): Promise<VillageOfficial[]> {
  return apiRequest<VillageOfficial[]>('/api/v1/public/village-profile/officials', {
    auth: false,
    signal,
  })
}

export function listPublicAnnouncements(
  params: { page?: number; page_size?: number } = {},
  signal?: AbortSignal,
): Promise<{ items: AnnouncementSummary[]; meta: PaginationMeta }> {
  const qs = toQueryString({
    page: params.page ?? 1,
    page_size: params.page_size ?? 6,
  })
  return apiRequestWithMeta<AnnouncementSummary[], PaginationMeta>(
    `/api/v1/public/announcements${qs}`,
    { auth: false, signal },
  ).then((result) => ({ items: result.data, meta: result.meta }))
}

export function getPublicAnnouncement(
  id: string,
  signal?: AbortSignal,
): Promise<Announcement> {
  return apiRequest<Announcement>(`/api/v1/public/announcements/${id}`, {
    auth: false,
    signal,
  })
}

export function listPublicGallery(
  params: { page?: number; page_size?: number } = {},
  signal?: AbortSignal,
): Promise<{ items: GalleryItem[]; meta: PaginationMeta }> {
  const qs = toQueryString({
    page: params.page ?? 1,
    page_size: params.page_size ?? 8,
  })
  return apiRequestWithMeta<GalleryItem[], PaginationMeta>(`/api/v1/public/gallery${qs}`, {
    auth: false,
    signal,
  }).then((result) => ({ items: result.data, meta: result.meta }))
}

export function listPublicActivities(
  params: { page?: number; page_size?: number; from?: string } = {},
  signal?: AbortSignal,
): Promise<{ items: Activity[]; meta: PaginationMeta }> {
  const qs = toQueryString({
    page: params.page ?? 1,
    page_size: params.page_size ?? 6,
    from: params.from,
  })
  return apiRequestWithMeta<Activity[], PaginationMeta>(`/api/v1/public/activities${qs}`, {
    auth: false,
    signal,
  }).then((result) => ({ items: result.data, meta: result.meta }))
}
