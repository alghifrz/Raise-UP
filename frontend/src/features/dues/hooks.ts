import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { dashboardSummaryQueryKey } from '../dashboard/hooks'
import {
  createDuesPayment,
  createDuesPeriod,
  deleteDuesPayment,
  deleteDuesPeriod,
  fetchDuesPeriodSummary,
  getDuesPayment,
  getDuesPeriod,
  listDuesPeriodStatus,
  listDuesPeriods,
  updateDuesPayment,
  updateDuesPeriod,
} from './api'
import type {
  CreateDuesPaymentRequest,
  CreateDuesPeriodRequest,
  DuesPeriodFilters,
  DuesPeriodStatusFilters,
  UpdateDuesPaymentRequest,
  UpdateDuesPeriodRequest,
} from './types'

export const duesQueryKey = ['dues'] as const

export function duesPeriodsQueryKey(filters: DuesPeriodFilters) {
  return [...duesQueryKey, 'periods', filters] as const
}

export function duesPeriodDetailQueryKey(id: string) {
  return [...duesQueryKey, 'period', id] as const
}

export function duesPeriodSummaryQueryKey(id: string) {
  return [...duesQueryKey, 'period-summary', id] as const
}

export function duesPeriodStatusQueryKey(id: string, filters: DuesPeriodStatusFilters) {
  return [...duesQueryKey, 'period-status', id, filters] as const
}

export function duesPaymentDetailQueryKey(id: string) {
  return [...duesQueryKey, 'payment', id] as const
}

export function useDuesPeriods(filters: DuesPeriodFilters) {
  return useQuery({
    queryKey: duesPeriodsQueryKey(filters),
    queryFn: ({ signal }) => listDuesPeriods(filters, signal),
    placeholderData: (previous) => previous,
  })
}

export function useDuesPeriod(id: string | undefined) {
  return useQuery({
    queryKey: duesPeriodDetailQueryKey(id ?? ''),
    queryFn: ({ signal }) => getDuesPeriod(id!, signal),
    enabled: Boolean(id),
  })
}

export function useDuesPeriodSummary(id: string | undefined) {
  return useQuery({
    queryKey: duesPeriodSummaryQueryKey(id ?? ''),
    queryFn: ({ signal }) => fetchDuesPeriodSummary(id!, signal),
    enabled: Boolean(id),
  })
}

export function useDuesPeriodStatus(id: string | undefined, filters: DuesPeriodStatusFilters) {
  return useQuery({
    queryKey: duesPeriodStatusQueryKey(id ?? '', filters),
    queryFn: ({ signal }) => listDuesPeriodStatus(id!, filters, signal),
    enabled: Boolean(id),
    placeholderData: (previous) => previous,
  })
}

export function useDuesPayment(id: string | undefined) {
  return useQuery({
    queryKey: duesPaymentDetailQueryKey(id ?? ''),
    queryFn: ({ signal }) => getDuesPayment(id!, signal),
    enabled: Boolean(id),
  })
}

function useInvalidateDues() {
  const queryClient = useQueryClient()

  return async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: duesQueryKey }),
      queryClient.invalidateQueries({ queryKey: dashboardSummaryQueryKey }),
    ])
  }
}

export function useCreateDuesPeriod() {
  const invalidate = useInvalidateDues()

  return useMutation({
    mutationFn: (payload: CreateDuesPeriodRequest) => createDuesPeriod(payload),
    onSuccess: async () => {
      await invalidate()
    },
  })
}

export function useUpdateDuesPeriod() {
  const invalidate = useInvalidateDues()

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateDuesPeriodRequest }) =>
      updateDuesPeriod(id, payload),
    onSuccess: async () => {
      await invalidate()
    },
  })
}

export function useDeleteDuesPeriod() {
  const invalidate = useInvalidateDues()

  return useMutation({
    mutationFn: (id: string) => deleteDuesPeriod(id),
    onSuccess: async () => {
      await invalidate()
    },
  })
}

export function useCreateDuesPayment() {
  const invalidate = useInvalidateDues()

  return useMutation({
    mutationFn: (payload: CreateDuesPaymentRequest) => createDuesPayment(payload),
    onSuccess: async () => {
      await invalidate()
    },
  })
}

export function useUpdateDuesPayment() {
  const invalidate = useInvalidateDues()

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateDuesPaymentRequest }) =>
      updateDuesPayment(id, payload),
    onSuccess: async () => {
      await invalidate()
    },
  })
}

export function useDeleteDuesPayment() {
  const invalidate = useInvalidateDues()

  return useMutation({
    mutationFn: (id: string) => deleteDuesPayment(id),
    onSuccess: async () => {
      await invalidate()
    },
  })
}
