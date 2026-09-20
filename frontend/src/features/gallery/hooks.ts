import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createGalleryItem,
  deleteGalleryItem,
  getGalleryItem,
  listGalleryItems,
  updateGalleryItem,
} from './api'
import type { CreateGalleryRequest, GalleryFilters, UpdateGalleryRequest } from './types'

export const galleryQueryKey = ['gallery'] as const

export function galleryListQueryKey(filters: GalleryFilters) {
  return [...galleryQueryKey, 'list', filters] as const
}

export function galleryDetailQueryKey(id: string) {
  return [...galleryQueryKey, 'detail', id] as const
}

export function useGalleryItems(filters: GalleryFilters) {
  return useQuery({
    queryKey: galleryListQueryKey(filters),
    queryFn: ({ signal }) => listGalleryItems(filters, signal),
    placeholderData: (previous) => previous,
  })
}

export function useGalleryItem(id: string | undefined) {
  return useQuery({
    queryKey: galleryDetailQueryKey(id ?? ''),
    queryFn: ({ signal }) => getGalleryItem(id!, signal),
    enabled: Boolean(id),
  })
}

function useInvalidateGallery() {
  const queryClient = useQueryClient()
  return async () => {
    await queryClient.invalidateQueries({ queryKey: galleryQueryKey })
  }
}

export function useCreateGalleryItem() {
  const invalidate = useInvalidateGallery()
  return useMutation({
    mutationFn: (payload: CreateGalleryRequest) => createGalleryItem(payload),
    onSuccess: async () => {
      await invalidate()
    },
  })
}

export function useUpdateGalleryItem() {
  const invalidate = useInvalidateGallery()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateGalleryRequest }) =>
      updateGalleryItem(id, payload),
    onSuccess: async () => {
      await invalidate()
    },
  })
}

export function useDeleteGalleryItem() {
  const invalidate = useInvalidateGallery()
  return useMutation({
    mutationFn: (id: string) => deleteGalleryItem(id),
    onSuccess: async () => {
      await invalidate()
    },
  })
}
