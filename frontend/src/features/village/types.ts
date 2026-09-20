export type VillageProfile = {
  id: string
  history: string
  vision: string
  mission: string
  updated_at: string
}

export type VillageOfficial = {
  id: string
  profile_id: string
  name: string
  position: string
  photo_url: string | null
  sort_order: number
}

export type UpdateVillageProfileRequest = {
  history?: string
  vision?: string
  mission?: string
}

export type CreateVillageOfficialRequest = {
  name: string
  position: string
  photo_url?: string
  sort_order?: number
}

export type UpdateVillageOfficialRequest = {
  name?: string
  position?: string
  photo_url?: string
  sort_order?: number
}
