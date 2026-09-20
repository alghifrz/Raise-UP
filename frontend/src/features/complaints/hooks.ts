import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { dashboardSummaryQueryKey } from '../dashboard/hooks'
import {
  createComplaint,
  deleteComplaint,
  getComplaint,
  listComplaints,
  updateComplaint,
  updateComplaintStatus,
} from './api'
import type {
  ComplaintFilters,
  CreateComplaintRequest,
  UpdateComplaintRequest,
  UpdateComplaintStatusRequest,
} from './types'

export const complaintsQueryKey = ['complaints'] as const

export function complaintsListQueryKey(filters: ComplaintFilters) {
  return [...complaintsQueryKey, 'list', filters] as const
}

export function complaintDetailQueryKey(id: string) {
  return [...complaintsQueryKey, 'detail', id] as const
}

export function useComplaints(filters: ComplaintFilters) {
  return useQuery({
    queryKey: complaintsListQueryKey(filters),
    queryFn: ({ signal }) => listComplaints(filters, signal),
    placeholderData: (previous) => previous,
  })
}

export function useComplaint(id: string | undefined) {
  return useQuery({
    queryKey: complaintDetailQueryKey(id ?? ''),
    queryFn: ({ signal }) => getComplaint(id!, signal),
    enabled: Boolean(id),
  })
}

function useInvalidateComplaints() {
  const queryClient = useQueryClient()

  return async (complaintId?: string) => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: complaintsQueryKey }),
      queryClient.invalidateQueries({ queryKey: dashboardSummaryQueryKey }),
      complaintId
        ? queryClient.invalidateQueries({ queryKey: complaintDetailQueryKey(complaintId) })
        : Promise.resolve(),
    ])
  }
}

export function useCreateComplaint() {
  const invalidate = useInvalidateComplaints()

  return useMutation({
    mutationFn: (payload: CreateComplaintRequest) => createComplaint(payload),
    onSuccess: async () => {
      await invalidate()
    },
  })
}

export function useUpdateComplaint() {
  const invalidate = useInvalidateComplaints()

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateComplaintRequest }) =>
      updateComplaint(id, payload),
    onSuccess: async (complaint) => {
      await invalidate(complaint.id)
    },
  })
}

export function useUpdateComplaintStatus() {
  const invalidate = useInvalidateComplaints()

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateComplaintStatusRequest }) =>
      updateComplaintStatus(id, payload),
    onSuccess: async (complaint) => {
      await invalidate(complaint.id)
    },
  })
}

export function useDeleteComplaint() {
  const invalidate = useInvalidateComplaints()

  return useMutation({
    mutationFn: (id: string) => deleteComplaint(id),
    onSuccess: async (_data, id) => {
      await invalidate(id)
    },
  })
}
