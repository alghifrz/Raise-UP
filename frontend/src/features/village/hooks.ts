import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createVillageOfficial,
  deleteVillageOfficial,
  getVillageProfile,
  listVillageOfficials,
  updateVillageOfficial,
  updateVillageProfile,
} from './api'
import type {
  CreateVillageOfficialRequest,
  UpdateVillageOfficialRequest,
  UpdateVillageProfileRequest,
} from './types'

export const villageQueryKey = ['village'] as const
export const villageProfileQueryKey = [...villageQueryKey, 'profile'] as const
export const villageOfficialsQueryKey = [...villageQueryKey, 'officials'] as const

export function useVillageProfile() {
  return useQuery({
    queryKey: villageProfileQueryKey,
    queryFn: ({ signal }) => getVillageProfile(signal),
  })
}

export function useVillageOfficials() {
  return useQuery({
    queryKey: villageOfficialsQueryKey,
    queryFn: ({ signal }) => listVillageOfficials(signal),
  })
}

export function useUpdateVillageProfile() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (payload: UpdateVillageProfileRequest) => updateVillageProfile(payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: villageQueryKey })
    },
  })
}

function useInvalidateOfficials() {
  const queryClient = useQueryClient()
  return async () => {
    await queryClient.invalidateQueries({ queryKey: villageQueryKey })
  }
}

export function useCreateVillageOfficial() {
  const invalidate = useInvalidateOfficials()
  return useMutation({
    mutationFn: (payload: CreateVillageOfficialRequest) => createVillageOfficial(payload),
    onSuccess: async () => {
      await invalidate()
    },
  })
}

export function useUpdateVillageOfficial() {
  const invalidate = useInvalidateOfficials()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateVillageOfficialRequest }) =>
      updateVillageOfficial(id, payload),
    onSuccess: async () => {
      await invalidate()
    },
  })
}

export function useDeleteVillageOfficial() {
  const invalidate = useInvalidateOfficials()
  return useMutation({
    mutationFn: (id: string) => deleteVillageOfficial(id),
    onSuccess: async () => {
      await invalidate()
    },
  })
}
