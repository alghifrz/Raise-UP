import { useQuery } from '@tanstack/react-query'
import {
  getPublicSiteSettings,
  getPublicVillageProfile,
  listPublicActivities,
  listPublicAnnouncements,
  listPublicGallery,
  listPublicVillageOfficials,
} from './api'

export function usePublicSiteSettings() {
  return useQuery({
    queryKey: ['public', 'site-settings'],
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

export function usePublicGallery() {
  return useQuery({
    queryKey: ['public', 'gallery'],
    queryFn: ({ signal }) => listPublicGallery({ page: 1, page_size: 8 }, signal),
  })
}

export function usePublicActivities() {
  const from = new Date().toISOString()
  return useQuery({
    queryKey: ['public', 'activities', from.slice(0, 10)],
    queryFn: ({ signal }) => listPublicActivities({ page: 1, page_size: 6, from }, signal),
  })
}
