import { apiRequest, apiRequestWithMeta, toQueryString } from '../../lib/api/client'
import type { PaginationMeta } from '../../lib/api/types'
import type {
  Activity,
  ActivityFilters,
  ActivityListResponse,
  CreateActivityRequest,
  UpdateActivityRequest,
} from './types'

export function listActivities(
  filters: ActivityFilters,
  signal?: AbortSignal,
): Promise<ActivityListResponse> {
  const qs = toQueryString({
    page: filters.page,
    page_size: filters.page_size,
    search: filters.search || undefined,
    from: filters.from || undefined,
    to: filters.to || undefined,
  })

  return apiRequestWithMeta<Activity[], PaginationMeta>(`/api/v1/activities${qs}`, {
    signal,
  }).then((result) => ({
    items: result.data,
    meta: result.meta,
  }))
}

export function getActivity(id: string, signal?: AbortSignal): Promise<Activity> {
  return apiRequest<Activity>(`/api/v1/activities/${id}`, { signal })
}

export function createActivity(payload: CreateActivityRequest): Promise<Activity> {
  return apiRequest<Activity>('/api/v1/activities', {
    method: 'POST',
    body: payload,
  })
}

export function updateActivity(id: string, payload: UpdateActivityRequest): Promise<Activity> {
  return apiRequest<Activity>(`/api/v1/activities/${id}`, {
    method: 'PATCH',
    body: payload,
  })
}

export function deleteActivity(id: string): Promise<void> {
  return apiRequest<void>(`/api/v1/activities/${id}`, {
    method: 'DELETE',
  })
}
