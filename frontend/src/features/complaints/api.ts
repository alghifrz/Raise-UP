import { apiRequest, apiRequestWithMeta, toQueryString } from '../../lib/api/client'
import type { PaginationMeta } from '../../lib/api/types'
import type {
  Complaint,
  ComplaintFilters,
  ComplaintListResponse,
  CreateComplaintRequest,
  UpdateComplaintRequest,
  UpdateComplaintStatusRequest,
} from './types'

export function listComplaints(
  filters: ComplaintFilters,
  signal?: AbortSignal,
): Promise<ComplaintListResponse> {
  const qs = toQueryString({
    page: filters.page,
    page_size: filters.page_size,
    search: filters.search || undefined,
    status: filters.status || undefined,
    urgency: filters.urgency || undefined,
    category: filters.category || undefined,
  })

  return apiRequestWithMeta<Complaint[], PaginationMeta>(`/api/v1/complaints${qs}`, { signal }).then(
    (result) => ({
      items: result.data,
      meta: result.meta,
    }),
  )
}

export function getComplaint(id: string, signal?: AbortSignal): Promise<Complaint> {
  return apiRequest<Complaint>(`/api/v1/complaints/${id}`, { signal })
}

export function createComplaint(payload: CreateComplaintRequest): Promise<Complaint> {
  return apiRequest<Complaint>('/api/v1/complaints', {
    method: 'POST',
    body: payload,
  })
}

export function updateComplaint(id: string, payload: UpdateComplaintRequest): Promise<Complaint> {
  return apiRequest<Complaint>(`/api/v1/complaints/${id}`, {
    method: 'PATCH',
    body: payload,
  })
}

export function updateComplaintStatus(
  id: string,
  payload: UpdateComplaintStatusRequest,
): Promise<Complaint> {
  return apiRequest<Complaint>(`/api/v1/complaints/${id}/status`, {
    method: 'PATCH',
    body: payload,
  })
}

export function deleteComplaint(id: string): Promise<void> {
  return apiRequest<void>(`/api/v1/complaints/${id}`, {
    method: 'DELETE',
  })
}
