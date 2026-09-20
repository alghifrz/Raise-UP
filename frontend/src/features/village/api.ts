import { apiRequest } from '../../lib/api/client'
import type {
  CreateVillageOfficialRequest,
  UpdateVillageOfficialRequest,
  UpdateVillageProfileRequest,
  VillageOfficial,
  VillageProfile,
} from './types'

export function getVillageProfile(signal?: AbortSignal): Promise<VillageProfile> {
  return apiRequest<VillageProfile>('/api/v1/village-profile', { signal })
}

export function updateVillageProfile(
  payload: UpdateVillageProfileRequest,
): Promise<VillageProfile> {
  return apiRequest<VillageProfile>('/api/v1/village-profile', {
    method: 'PATCH',
    body: payload,
  })
}

export function listVillageOfficials(signal?: AbortSignal): Promise<VillageOfficial[]> {
  return apiRequest<VillageOfficial[]>('/api/v1/village-profile/officials', { signal })
}

export function createVillageOfficial(
  payload: CreateVillageOfficialRequest,
): Promise<VillageOfficial> {
  return apiRequest<VillageOfficial>('/api/v1/village-profile/officials', {
    method: 'POST',
    body: payload,
  })
}

export function updateVillageOfficial(
  id: string,
  payload: UpdateVillageOfficialRequest,
): Promise<VillageOfficial> {
  return apiRequest<VillageOfficial>(`/api/v1/village-profile/officials/${id}`, {
    method: 'PATCH',
    body: payload,
  })
}

export function deleteVillageOfficial(id: string): Promise<void> {
  return apiRequest<void>(`/api/v1/village-profile/officials/${id}`, {
    method: 'DELETE',
  })
}
