import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getSiteSettings, updateSiteSettings } from './api'
import type { UpdateSiteSettingsRequest } from './types'

export const siteSettingsQueryKey = ['site-settings'] as const

export function useSiteSettings() {
  return useQuery({
    queryKey: siteSettingsQueryKey,
    queryFn: ({ signal }) => getSiteSettings(signal),
  })
}

export function useUpdateSiteSettings() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (payload: UpdateSiteSettingsRequest) => updateSiteSettings(payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: siteSettingsQueryKey })
    },
  })
}
