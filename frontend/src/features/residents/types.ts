import type { PaginationMeta } from '../../lib/api/types'

export type ResidentGender = 'LAKI_LAKI' | 'PEREMPUAN'

export type Resident = {
  id: string
  name: string
  phone: string
  gender: ResidentGender
  created_at: string
  updated_at: string
}

export type ResidentFilters = {
  page: number
  page_size: number
  search: string
  gender: ResidentGender | ''
}

export type ResidentListResponse = {
  items: Resident[]
  meta: PaginationMeta
}

export type CreateResidentRequest = {
  name: string
  phone: string
  gender: ResidentGender
}

export type UpdateResidentRequest = {
  name?: string
  phone?: string
  gender?: ResidentGender
}

export const RESIDENT_GENDER_OPTIONS: Array<{ value: ResidentGender; label: string }> = [
  { value: 'LAKI_LAKI', label: 'Laki-laki' },
  { value: 'PEREMPUAN', label: 'Perempuan' },
]

export function formatResidentGender(gender: ResidentGender): string {
  return RESIDENT_GENDER_OPTIONS.find((item) => item.value === gender)?.label ?? gender
}
