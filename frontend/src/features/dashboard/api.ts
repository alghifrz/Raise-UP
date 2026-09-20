import { apiRequest } from '../../lib/api/client'
import type { DashboardSummary } from './types'

export function fetchDashboardSummary(signal?: AbortSignal): Promise<DashboardSummary> {
  return apiRequest<DashboardSummary>('/api/v1/dashboard/summary', { signal })
}
