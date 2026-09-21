import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { dashboardSummaryQueryKey } from '../dashboard/hooks'
import {
  createAnnouncement,
  deleteAnnouncement,
  getAnnouncement,
  listAnnouncements,
  updateAnnouncement,
  updateAnnouncementStatus,
} from './api'
import type {
  AnnouncementFilters,
  CreateAnnouncementRequest,
  UpdateAnnouncementRequest,
} from './types'

export const announcementsQueryKey = ['announcements'] as const
const publicAnnouncementsQueryKey = ['public', 'announcements'] as const

export function announcementsListQueryKey(filters: AnnouncementFilters) {
  return [...announcementsQueryKey, 'list', filters] as const
}

export function announcementDetailQueryKey(id: string) {
  return [...announcementsQueryKey, 'detail', id] as const
}

export function useAnnouncements(filters: AnnouncementFilters) {
  return useQuery({
    queryKey: announcementsListQueryKey(filters),
    queryFn: ({ signal }) => listAnnouncements(filters, signal),
    placeholderData: (previous) => previous,
  })
}

export function useAnnouncement(id: string | undefined) {
  return useQuery({
    queryKey: announcementDetailQueryKey(id ?? ''),
    queryFn: ({ signal }) => getAnnouncement(id!, signal),
    enabled: Boolean(id),
  })
}

function useInvalidateAnnouncements() {
  const queryClient = useQueryClient()

  return async (announcementId?: string) => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: announcementsQueryKey }),
      queryClient.invalidateQueries({ queryKey: publicAnnouncementsQueryKey }),
      queryClient.invalidateQueries({ queryKey: dashboardSummaryQueryKey }),
      announcementId
        ? queryClient.invalidateQueries({ queryKey: announcementDetailQueryKey(announcementId) })
        : Promise.resolve(),
    ])
  }
}

export function useCreateAnnouncement() {
  const invalidate = useInvalidateAnnouncements()

  return useMutation({
    mutationFn: (payload: CreateAnnouncementRequest) => createAnnouncement(payload),
    onSuccess: async () => {
      await invalidate()
    },
  })
}

export function useUpdateAnnouncement() {
  const invalidate = useInvalidateAnnouncements()

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateAnnouncementRequest }) =>
      updateAnnouncement(id, payload),
    onSuccess: async (item) => {
      await invalidate(item.id)
    },
  })
}

export function usePublishAnnouncement() {
  const invalidate = useInvalidateAnnouncements()

  return useMutation({
    mutationFn: (id: string) =>
      updateAnnouncementStatus(id, { status: 'PUBLISHED' }),
    onSuccess: async (item) => {
      await invalidate(item.id)
    },
  })
}

export function useDeleteAnnouncement() {
  const invalidate = useInvalidateAnnouncements()

  return useMutation({
    mutationFn: (id: string) => deleteAnnouncement(id),
    onSuccess: async (_data, id) => {
      await invalidate(id)
    },
  })
}
