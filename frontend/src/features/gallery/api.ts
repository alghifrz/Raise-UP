import { apiRequest, apiRequestWithMeta, toQueryString } from '../../lib/api/client'
import type { PaginationMeta } from '../../lib/api/types'
import type {
  CreateGalleryRequest,
  GalleryFilters,
  GalleryItem,
  GalleryListResponse,
  UpdateGalleryRequest,
} from './types'

export function listGalleryItems(
  filters: GalleryFilters,
  signal?: AbortSignal,
): Promise<GalleryListResponse> {
  const qs = toQueryString({
    page: filters.page,
    page_size: filters.page_size,
    search: filters.search || undefined,
  })

  return apiRequestWithMeta<GalleryItem[], PaginationMeta>(`/api/v1/gallery${qs}`, {
    signal,
  }).then((result) => ({
    items: result.data,
    meta: result.meta,
  }))
}

export function getGalleryItem(id: string, signal?: AbortSignal): Promise<GalleryItem> {
  return apiRequest<GalleryItem>(`/api/v1/gallery/${id}`, { signal })
}

export function createGalleryItem(payload: CreateGalleryRequest): Promise<GalleryItem> {
  return apiRequest<GalleryItem>('/api/v1/gallery', {
    method: 'POST',
    body: payload,
  })
}

export function updateGalleryItem(id: string, payload: UpdateGalleryRequest): Promise<GalleryItem> {
  return apiRequest<GalleryItem>(`/api/v1/gallery/${id}`, {
    method: 'PATCH',
    body: payload,
  })
}

export function deleteGalleryItem(id: string): Promise<void> {
  return apiRequest<void>(`/api/v1/gallery/${id}`, {
    method: 'DELETE',
  })
}
