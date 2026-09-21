import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getSiteSettings, updateSiteSettings } from './api'
import type { UpdateSiteSettingsRequest } from './types'

export const siteSettingsQueryKey = ['site-settings'] as const
export const publicSiteSettingsQueryKey = ['public', 'site-settings'] as const

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
    onSuccess: async (updated) => {
      // Keep the public landing page in sync immediately after an admin update.
      queryClient.setQueryData(publicSiteSettingsQueryKey, updated)
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: siteSettingsQueryKey }),
        queryClient.invalidateQueries({ queryKey: publicSiteSettingsQueryKey }),
      ])
    },
  })
}
