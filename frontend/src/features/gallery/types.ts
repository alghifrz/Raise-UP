import type { PaginationMeta } from '../../lib/api/types'

export type GalleryItem = {
  id: string
  image_url: string
  storage_path: string
  caption: string
  sort_order: number
  created_at: string
}

export type GalleryFilters = {
  page: number
  page_size: number
  search: string
}

export type GalleryListResponse = {
  items: GalleryItem[]
  meta: PaginationMeta
}

export type CreateGalleryRequest = {
  image: File
  caption?: string
  sort_order?: number
}

export type UpdateGalleryRequest = {
  image?: File
  caption?: string
  sort_order?: number
}
