import { apiRequest } from '../../lib/api/client'
import type { SiteSettings, UpdateSiteSettingsRequest } from './types'

export function getSiteSettings(signal?: AbortSignal): Promise<SiteSettings> {
  return apiRequest<SiteSettings>('/api/v1/site-settings', { signal })
}

export function updateSiteSettings(payload: UpdateSiteSettingsRequest): Promise<SiteSettings> {
  return apiRequest<SiteSettings>('/api/v1/site-settings', {
    method: 'PATCH',
    body: payload,
  })
}
