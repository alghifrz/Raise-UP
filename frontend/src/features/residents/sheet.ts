import * as XLSX from 'xlsx'
import type { CreateResidentRequest, ResidentGender } from './types'

export const RESIDENT_SHEET_HEADERS = ['nama', 'telepon', 'jenis_kelamin'] as const

export const RESIDENT_SHEET_EXAMPLE_ROWS: Array<[string, string, string]> = [
  ['Budi Santoso', '081234567890', 'Laki-laki'],
  ['Siti Aminah', '081298765432', 'Perempuan'],
]

export const MAX_RESIDENT_IMPORT_ROWS = 500

export type ParsedResidentRow = {
  rowNumber: number
  name: string
  phone: string
  genderRaw: string
  gender: ResidentGender | null
  error: string | null
}

export type ResidentImportResult = {
  rowNumber: number
  name: string
  phone: string
  ok: boolean
  message: string
}

function normalizeHeader(value: unknown): string {
  return String(value ?? '')
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '_')
}

function cellText(value: unknown): string {
  if (value == null) {
    return ''
  }
  if (typeof value === 'number' && Number.isFinite(value)) {
    return String(Math.trunc(value))
  }
  return String(value).trim()
}

export function parseResidentGenderCell(raw: string): ResidentGender | null {
  const normalized = raw.trim().toLowerCase().replace(/[\s-]+/g, '_')
  if (
    normalized === 'laki_laki' ||
    normalized === 'laki' ||
    normalized === 'l' ||
    normalized === 'pria' ||
    normalized === 'male'
  ) {
    return 'LAKI_LAKI'
  }
  if (
    normalized === 'perempuan' ||
    normalized === 'p' ||
    normalized === 'wanita' ||
    normalized === 'female'
  ) {
    return 'PEREMPUAN'
  }
  return null
}

/** Download an .xlsx template that Excel / Google Sheets can open. */
export function downloadResidentImportTemplate(): void {
  const sheet = XLSX.utils.aoa_to_sheet([
    [...RESIDENT_SHEET_HEADERS],
    ...RESIDENT_SHEET_EXAMPLE_ROWS,
  ])
  sheet['!cols'] = [{ wch: 24 }, { wch: 18 }, { wch: 16 }]

  const workbook = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(workbook, sheet, 'Warga')
  XLSX.writeFile(workbook, 'template-import-warga.xlsx')
}

export async function parseResidentImportFile(file: File): Promise<ParsedResidentRow[]> {
  const buffer = await file.arrayBuffer()
  const workbook = XLSX.read(buffer, { type: 'array' })
  const firstSheetName = workbook.SheetNames[0]
  if (!firstSheetName) {
    throw new Error('File tidak berisi sheet.')
  }

  const sheet = workbook.Sheets[firstSheetName]
  if (!sheet) {
    throw new Error('Sheet pertama tidak dapat dibaca.')
  }

  const rows = XLSX.utils.sheet_to_json<(string | number | null | undefined)[]>(sheet, {
    header: 1,
    defval: '',
    blankrows: false,
  })

  if (rows.length === 0) {
    throw new Error('File kosong.')
  }

  const headerRow = (rows[0] ?? []).map(normalizeHeader)
  const nameIdx = headerRow.findIndex((h) => h === 'nama' || h === 'name')
  const phoneIdx = headerRow.findIndex(
    (h) => h === 'telepon' || h === 'phone' || h === 'no_telepon' || h === 'nomor_telepon',
  )
  const genderIdx = headerRow.findIndex(
    (h) =>
      h === 'jenis_kelamin' ||
      h === 'gender' ||
      h === 'jk' ||
      h === 'kelamin',
  )

  if (nameIdx < 0 || phoneIdx < 0 || genderIdx < 0) {
    throw new Error(
      'Header wajib: nama, telepon, jenis_kelamin. Unduh template untuk format yang benar.',
    )
  }

  const dataRows = rows.slice(1)
  if (dataRows.length > MAX_RESIDENT_IMPORT_ROWS) {
    throw new Error(`Maksimal ${MAX_RESIDENT_IMPORT_ROWS} baris data per import.`)
  }

  return dataRows.map((row, index) => {
    const rowNumber = index + 2
    const name = cellText(row[nameIdx])
    const phone = cellText(row[phoneIdx])
    const genderRaw = cellText(row[genderIdx])
    const gender = parseResidentGenderCell(genderRaw)

    let error: string | null = null
    if (!name) {
      error = 'Nama wajib diisi.'
    } else if (!phone) {
      error = 'Telepon wajib diisi.'
    } else if (!genderRaw) {
      error = 'Jenis kelamin wajib diisi.'
    } else if (!gender) {
      error = 'Jenis kelamin harus Laki-laki atau Perempuan.'
    }

    return {
      rowNumber,
      name,
      phone,
      genderRaw,
      gender,
      error,
    }
  })
}

export function toCreateResidentRequest(row: ParsedResidentRow): CreateResidentRequest | null {
  if (row.error || !row.gender) {
    return null
  }
  return {
    name: row.name,
    phone: row.phone,
    gender: row.gender,
  }
}
