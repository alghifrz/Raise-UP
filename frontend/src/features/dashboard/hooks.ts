import { useQuery } from '@tanstack/react-query'
import { fetchDashboardSummary } from './api'

export const dashboardSummaryQueryKey = ['dashboard', 'summary'] as const

export function useDashboardSummary() {
  return useQuery({
    queryKey: dashboardSummaryQueryKey,
    queryFn: ({ signal }) => fetchDashboardSummary(signal),
  })
}
