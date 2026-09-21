import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { dashboardSummaryQueryKey } from '../dashboard/hooks'
import {
  createActivity,
  deleteActivity,
  getActivity,
  listActivities,
  updateActivity,
} from './api'
import type { ActivityFilters, CreateActivityRequest, UpdateActivityRequest } from './types'

export const activitiesQueryKey = ['activities'] as const
const publicActivitiesQueryKey = ['public', 'activities'] as const

export function activitiesListQueryKey(filters: ActivityFilters) {
  return [...activitiesQueryKey, 'list', filters] as const
}

export function activityDetailQueryKey(id: string) {
  return [...activitiesQueryKey, 'detail', id] as const
}

export function useActivities(filters: ActivityFilters) {
  return useQuery({
    queryKey: activitiesListQueryKey(filters),
    queryFn: ({ signal }) => listActivities(filters, signal),
    placeholderData: (previous) => previous,
  })
}

export function useActivity(id: string | undefined) {
  return useQuery({
    queryKey: activityDetailQueryKey(id ?? ''),
    queryFn: ({ signal }) => getActivity(id!, signal),
    enabled: Boolean(id),
  })
}

function useInvalidateActivities() {
  const queryClient = useQueryClient()

  return async (activityId?: string) => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: activitiesQueryKey }),
      queryClient.invalidateQueries({ queryKey: publicActivitiesQueryKey }),
      queryClient.invalidateQueries({ queryKey: dashboardSummaryQueryKey }),
      activityId
        ? queryClient.invalidateQueries({ queryKey: activityDetailQueryKey(activityId) })
        : Promise.resolve(),
    ])
  }
}

export function useCreateActivity() {
  const invalidate = useInvalidateActivities()

  return useMutation({
    mutationFn: (payload: CreateActivityRequest) => createActivity(payload),
    onSuccess: async () => {
      await invalidate()
    },
  })
}

export function useUpdateActivity() {
  const invalidate = useInvalidateActivities()

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateActivityRequest }) =>
      updateActivity(id, payload),
    onSuccess: async (item) => {
      await invalidate(item.id)
    },
  })
}

export function useDeleteActivity() {
  const invalidate = useInvalidateActivities()

  return useMutation({
    mutationFn: (id: string) => deleteActivity(id),
    onSuccess: async (_data, id) => {
      await invalidate(id)
    },
  })
}
