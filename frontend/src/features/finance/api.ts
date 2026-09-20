import { apiRequest, apiRequestWithMeta, toQueryString } from '../../lib/api/client'
import type { PaginationMeta } from '../../lib/api/types'
import type {
  CreateFinanceTransactionRequest,
  FinanceFilters,
  FinanceSummary,
  FinanceSummaryFilters,
  FinanceTransaction,
  FinanceTransactionListResponse,
  UpdateFinanceTransactionRequest,
} from './types'

export function listFinanceTransactions(
  filters: FinanceFilters,
  signal?: AbortSignal,
): Promise<FinanceTransactionListResponse> {
  const qs = toQueryString({
    page: filters.page,
    page_size: filters.page_size,
    search: filters.search || undefined,
    type: filters.type || undefined,
    category: filters.category || undefined,
    from: filters.from || undefined,
    to: filters.to || undefined,
  })

  return apiRequestWithMeta<FinanceTransaction[], PaginationMeta>(
    `/api/v1/finance/transactions${qs}`,
    { signal },
  ).then((result) => ({
    items: result.data,
    meta: result.meta,
  }))
}

export function getFinanceTransaction(
  id: string,
  signal?: AbortSignal,
): Promise<FinanceTransaction> {
  return apiRequest<FinanceTransaction>(`/api/v1/finance/transactions/${id}`, { signal })
}

export function createFinanceTransaction(
  payload: CreateFinanceTransactionRequest,
): Promise<FinanceTransaction> {
  return apiRequest<FinanceTransaction>('/api/v1/finance/transactions', {
    method: 'POST',
    body: payload,
  })
}

export function updateFinanceTransaction(
  id: string,
  payload: UpdateFinanceTransactionRequest,
): Promise<FinanceTransaction> {
  return apiRequest<FinanceTransaction>(`/api/v1/finance/transactions/${id}`, {
    method: 'PATCH',
    body: payload,
  })
}

export function deleteFinanceTransaction(id: string): Promise<void> {
  return apiRequest<void>(`/api/v1/finance/transactions/${id}`, {
    method: 'DELETE',
  })
}

export function fetchFinanceSummary(
  filters: FinanceSummaryFilters,
  signal?: AbortSignal,
): Promise<FinanceSummary> {
  const qs = toQueryString({
    from: filters.from || undefined,
    to: filters.to || undefined,
  })
  return apiRequest<FinanceSummary>(`/api/v1/finance/summary${qs}`, { signal })
}
