import { useQuery } from '@tanstack/react-query'
import { publicSiteSettingsQueryKey } from '../site-settings/hooks'
import {
  getPublicAnnouncement,
  getPublicSiteSettings,
  getPublicVillageProfile,
  listPublicActivities,
  listPublicAnnouncements,
  listPublicGallery,
  listPublicVillageOfficials,
} from './api'

export function usePublicSiteSettings() {
  return useQuery({
    queryKey: publicSiteSettingsQueryKey,
    queryFn: ({ signal }) => getPublicSiteSettings(signal),
  })
}

export function usePublicVillageProfile() {
  return useQuery({
    queryKey: ['public', 'village-profile'],
    queryFn: ({ signal }) => getPublicVillageProfile(signal),
  })
}

export function usePublicVillageOfficials() {
  return useQuery({
    queryKey: ['public', 'village-officials'],
    queryFn: ({ signal }) => listPublicVillageOfficials(signal),
  })
}

export function usePublicAnnouncements() {
  return useQuery({
    queryKey: ['public', 'announcements'],
    queryFn: ({ signal }) => listPublicAnnouncements({ page: 1, page_size: 6 }, signal),
  })
}

export function usePublicAnnouncement(id: string | undefined) {
  return useQuery({
    queryKey: ['public', 'announcements', 'detail', id ?? ''],
    queryFn: ({ signal }) => getPublicAnnouncement(id!, signal),
    enabled: Boolean(id),
  })
}

export function usePublicGallery() {
  return useQuery({
    queryKey: ['public', 'gallery'],
    queryFn: ({ signal }) => listPublicGallery({ page: 1, page_size: 8 }, signal),
  })
}

export function usePublicActivities() {
  const from = jakartaDateOnly(new Date())
  return useQuery({
    queryKey: ['public', 'activities', from],
    queryFn: ({ signal }) => listPublicActivities({ page: 1, page_size: 6, from }, signal),
  })
}

function jakartaDateOnly(date: Date): string {
  const parts = new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Asia/Jakarta',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).formatToParts(date)
  const value = Object.fromEntries(parts.map((part) => [part.type, part.value]))
  return `${value.year}-${value.month}-${value.day}`
}
