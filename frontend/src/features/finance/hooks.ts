import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { dashboardSummaryQueryKey } from '../dashboard/hooks'
import {
  createFinanceTransaction,
  deleteFinanceTransaction,
  fetchFinanceSummary,
  getFinanceTransaction,
  listFinanceTransactions,
  updateFinanceTransaction,
} from './api'
import type {
  CreateFinanceTransactionRequest,
  FinanceFilters,
  FinanceSummaryFilters,
  UpdateFinanceTransactionRequest,
} from './types'

export const financeQueryKey = ['finance'] as const

export function financeTransactionsQueryKey(filters: FinanceFilters) {
  return [...financeQueryKey, 'transactions', filters] as const
}

export function financeTransactionDetailQueryKey(id: string) {
  return [...financeQueryKey, 'transaction', id] as const
}

export function financeSummaryQueryKey(filters: FinanceSummaryFilters) {
  return [...financeQueryKey, 'summary', filters] as const
}

export function useFinanceTransactions(filters: FinanceFilters) {
  return useQuery({
    queryKey: financeTransactionsQueryKey(filters),
    queryFn: ({ signal }) => listFinanceTransactions(filters, signal),
    placeholderData: (previous) => previous,
  })
}

export function useFinanceTransaction(id: string | undefined) {
  return useQuery({
    queryKey: financeTransactionDetailQueryKey(id ?? ''),
    queryFn: ({ signal }) => getFinanceTransaction(id!, signal),
    enabled: Boolean(id),
  })
}

export function useFinanceSummary(filters: FinanceSummaryFilters) {
  return useQuery({
    queryKey: financeSummaryQueryKey(filters),
    queryFn: ({ signal }) => fetchFinanceSummary(filters, signal),
  })
}

function useInvalidateFinance() {
  const queryClient = useQueryClient()

  return async (transactionId?: string) => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: financeQueryKey }),
      queryClient.invalidateQueries({ queryKey: dashboardSummaryQueryKey }),
      transactionId
        ? queryClient.invalidateQueries({
            queryKey: financeTransactionDetailQueryKey(transactionId),
          })
        : Promise.resolve(),
    ])
  }
}

export function useCreateFinanceTransaction() {
  const invalidate = useInvalidateFinance()

  return useMutation({
    mutationFn: (payload: CreateFinanceTransactionRequest) => createFinanceTransaction(payload),
    onSuccess: async () => {
      await invalidate()
    },
  })
}

export function useUpdateFinanceTransaction() {
  const invalidate = useInvalidateFinance()

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateFinanceTransactionRequest }) =>
      updateFinanceTransaction(id, payload),
    onSuccess: async (item) => {
      await invalidate(item.id)
    },
  })
}

export function useDeleteFinanceTransaction() {
  const invalidate = useInvalidateFinance()

  return useMutation({
    mutationFn: (id: string) => deleteFinanceTransaction(id),
    onSuccess: async (_data, id) => {
      await invalidate(id)
    },
  })
}
