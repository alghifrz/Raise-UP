import { apiRequest, apiRequestWithMeta, toQueryString } from '../../lib/api/client'
import type { PaginationMeta } from '../../lib/api/types'
import type {
  CreateDuesPaymentRequest,
  CreateDuesPeriodRequest,
  DuesPayment,
  DuesPeriod,
  DuesPeriodFilters,
  DuesPeriodListResponse,
  DuesReminderResult,
  DuesPeriodStatusFilters,
  DuesPeriodSummary,
  DuesStatusListResponse,
  ResidentPaymentStatus,
  UpdateDuesPaymentRequest,
  UpdateDuesPeriodRequest,
} from './types'

export function listDuesPeriods(
  filters: DuesPeriodFilters,
  signal?: AbortSignal,
): Promise<DuesPeriodListResponse> {
  const qs = toQueryString({
    page: filters.page,
    page_size: filters.page_size,
    year: filters.year || undefined,
    month: filters.month || undefined,
    half: filters.half || undefined,
  })

  return apiRequestWithMeta<DuesPeriod[], PaginationMeta>(`/api/v1/dues/periods${qs}`, {
    signal,
  }).then((result) => ({
    items: result.data,
    meta: result.meta,
  }))
}

export function getDuesPeriod(id: string, signal?: AbortSignal): Promise<DuesPeriod> {
  return apiRequest<DuesPeriod>(`/api/v1/dues/periods/${id}`, { signal })
}

export function createDuesPeriod(payload: CreateDuesPeriodRequest): Promise<DuesPeriod> {
  return apiRequest<DuesPeriod>('/api/v1/dues/periods', {
    method: 'POST',
    body: payload,
  })
}

export function updateDuesPeriod(
  id: string,
  payload: UpdateDuesPeriodRequest,
): Promise<DuesPeriod> {
  return apiRequest<DuesPeriod>(`/api/v1/dues/periods/${id}`, {
    method: 'PATCH',
    body: payload,
  })
}

export function deleteDuesPeriod(id: string): Promise<void> {
  return apiRequest<void>(`/api/v1/dues/periods/${id}`, {
    method: 'DELETE',
  })
}

export function fetchDuesPeriodSummary(
  id: string,
  signal?: AbortSignal,
): Promise<DuesPeriodSummary> {
  return apiRequest<DuesPeriodSummary>(`/api/v1/dues/periods/${id}/summary`, { signal })
}

export function sendUnpaidDuesReminders(id: string): Promise<DuesReminderResult> {
  return apiRequest<DuesReminderResult>(`/api/v1/dues/periods/${id}/remind-unpaid`, {
    method: 'POST',
  })
}

export function listDuesPeriodStatus(
  id: string,
  filters: DuesPeriodStatusFilters,
  signal?: AbortSignal,
): Promise<DuesStatusListResponse> {
  const qs = toQueryString({
    page: filters.page,
    page_size: filters.page_size,
    status: filters.status || undefined,
    search: filters.search || undefined,
  })

  return apiRequestWithMeta<ResidentPaymentStatus[], PaginationMeta>(
    `/api/v1/dues/periods/${id}/status${qs}`,
    { signal },
  ).then((result) => ({
    items: result.data,
    meta: result.meta,
  }))
}

export function getDuesPayment(id: string, signal?: AbortSignal): Promise<DuesPayment> {
  return apiRequest<DuesPayment>(`/api/v1/dues/payments/${id}`, { signal })
}

export function createDuesPayment(payload: CreateDuesPaymentRequest): Promise<DuesPayment> {
  return apiRequest<DuesPayment>('/api/v1/dues/payments', {
    method: 'POST',
    body: payload,
  })
}

export function updateDuesPayment(
  id: string,
  payload: UpdateDuesPaymentRequest,
): Promise<DuesPayment> {
  return apiRequest<DuesPayment>(`/api/v1/dues/payments/${id}`, {
    method: 'PATCH',
    body: payload,
  })
}

export function deleteDuesPayment(id: string): Promise<void> {
  return apiRequest<void>(`/api/v1/dues/payments/${id}`, {
    method: 'DELETE',
  })
}
