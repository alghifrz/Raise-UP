import { apiRequest, apiRequestWithMeta, toQueryString } from '../../lib/api/client'
import type { PaginationMeta } from '../../lib/api/types'
import type {
  CreateResidentRequest,
  Resident,
  ResidentFilters,
  ResidentListResponse,
  UpdateResidentRequest,
} from './types'

export function listResidents(
  filters: ResidentFilters,
  signal?: AbortSignal,
): Promise<ResidentListResponse> {
  const qs = toQueryString({
    page: filters.page,
    page_size: filters.page_size,
    search: filters.search || undefined,
    gender: filters.gender || undefined,
  })

  return apiRequestWithMeta<Resident[], PaginationMeta>(`/api/v1/residents${qs}`, { signal }).then(
    (result) => ({
      items: result.data,
      meta: result.meta,
    }),
  )
}

export function getResident(id: string, signal?: AbortSignal): Promise<Resident> {
  return apiRequest<Resident>(`/api/v1/residents/${id}`, { signal })
}

export function createResident(payload: CreateResidentRequest): Promise<Resident> {
  return apiRequest<Resident>('/api/v1/residents', {
    method: 'POST',
    body: payload,
  })
}

export function updateResident(id: string, payload: UpdateResidentRequest): Promise<Resident> {
  return apiRequest<Resident>(`/api/v1/residents/${id}`, {
    method: 'PATCH',
    body: payload,
  })
}

export function deleteResident(id: string): Promise<void> {
  return apiRequest<void>(`/api/v1/residents/${id}`, {
    method: 'DELETE',
  })
}
