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
  return apiRequest<GalleryItem>('/api/v1/gallery/upload', {
    method: 'POST',
    body: toGalleryFormData(payload.image, payload.caption, payload.sort_order),
  })
}

export function updateGalleryItem(id: string, payload: UpdateGalleryRequest): Promise<GalleryItem> {
  if (payload.image) {
    return apiRequest<GalleryItem>(`/api/v1/gallery/${id}/image`, {
      method: 'PUT',
      body: toGalleryFormData(payload.image, payload.caption, payload.sort_order),
    })
  }
  return apiRequest<GalleryItem>(`/api/v1/gallery/${id}`, {
    method: 'PATCH',
    body: {
      caption: payload.caption,
      sort_order: payload.sort_order,
    },
  })
}

export function deleteGalleryItem(id: string): Promise<void> {
  return apiRequest<void>(`/api/v1/gallery/${id}`, {
    method: 'DELETE',
  })
}

function toGalleryFormData(
  image: File,
  caption?: string,
  sortOrder?: number,
): FormData {
  const form = new FormData()
  form.set('image', image)
  if (caption !== undefined) {
    form.set('caption', caption)
  }
  if (sortOrder !== undefined) {
    form.set('sort_order', String(sortOrder))
  }
  return form
}
