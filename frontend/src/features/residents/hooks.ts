import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { dashboardSummaryQueryKey } from '../dashboard/hooks'
import {
  createResident,
  deleteResident,
  getResident,
  listResidents,
  updateResident,
} from './api'
import type { CreateResidentRequest, ResidentFilters, UpdateResidentRequest } from './types'

export const residentsQueryKey = ['residents'] as const

export function residentsListQueryKey(filters: ResidentFilters) {
  return [...residentsQueryKey, 'list', filters] as const
}

export function residentDetailQueryKey(id: string) {
  return [...residentsQueryKey, 'detail', id] as const
}

export function useResidents(filters: ResidentFilters) {
  return useQuery({
    queryKey: residentsListQueryKey(filters),
    queryFn: ({ signal }) => listResidents(filters, signal),
    placeholderData: (previous) => previous,
  })
}

export function useResident(id: string | null) {
  return useQuery({
    queryKey: residentDetailQueryKey(id ?? ''),
    queryFn: ({ signal }) => getResident(id!, signal),
    enabled: Boolean(id),
  })
}

function useInvalidateResidents() {
  const queryClient = useQueryClient()

  return async (residentId?: string) => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: residentsQueryKey }),
      queryClient.invalidateQueries({ queryKey: dashboardSummaryQueryKey }),
      residentId
        ? queryClient.invalidateQueries({ queryKey: residentDetailQueryKey(residentId) })
        : Promise.resolve(),
    ])
  }
}

export function useCreateResident() {
  const invalidate = useInvalidateResidents()

  return useMutation({
    mutationFn: (payload: CreateResidentRequest) => createResident(payload),
    onSuccess: async () => {
      await invalidate()
    },
  })
}

export function useUpdateResident() {
  const invalidate = useInvalidateResidents()

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateResidentRequest }) =>
      updateResident(id, payload),
    onSuccess: async (resident) => {
      await invalidate(resident.id)
    },
  })
}

export function useDeleteResident() {
  const invalidate = useInvalidateResidents()

  return useMutation({
    mutationFn: (id: string) => deleteResident(id),
    onSuccess: async (_data, id) => {
      await invalidate(id)
    },
  })
}

/** Invalidate residents + dashboard after a bulk sheet import finishes. */
export function useInvalidateResidentsAfterImport() {
  const invalidate = useInvalidateResidents()
  return invalidate
}
