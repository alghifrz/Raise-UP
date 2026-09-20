export function cn(...parts: Array<string | false | null | undefined>): string {
  return parts.filter(Boolean).join(' ')
}

const idrFormatter = new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  minimumFractionDigits: 0,
  maximumFractionDigits: 0,
})

/** Format backend int64 IDR amounts for display. Do not perform money math here. */
export function formatIdr(amount: number): string {
  return idrFormatter.format(amount).replace(/\s/g, '')
}

/**
 * Parse a positive integer rupiah amount from user input.
 * Rejects decimals, signs, and non-digit characters. Does not round.
 */
export function parsePositiveIntegerAmount(
  raw: string,
): { ok: true; value: number } | { ok: false; message: string } {
  const trimmed = raw.trim()
  if (!trimmed) {
    return { ok: false, message: 'Nominal wajib diisi.' }
  }
  if (!/^\d+$/.test(trimmed)) {
    return { ok: false, message: 'Nominal harus bilangan bulat positif tanpa desimal.' }
  }
  const value = Number(trimmed)
  if (!Number.isSafeInteger(value) || value <= 0) {
    return { ok: false, message: 'Nominal harus bilangan bulat positif yang valid.' }
  }
  return { ok: true, value }
}

const dateTimeFormatter = new Intl.DateTimeFormat('id-ID', {
  dateStyle: 'medium',
  timeStyle: 'short',
  timeZone: 'Asia/Jakarta',
})

const dateFormatter = new Intl.DateTimeFormat('id-ID', {
  dateStyle: 'medium',
  timeZone: 'Asia/Jakarta',
})

/** Format RFC3339 timestamps for admin display in Asia/Jakarta. */
export function formatDateTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return dateTimeFormatter.format(date)
}

/** Format RFC3339 / date-only values as a calendar date in Asia/Jakarta. */
export function formatDate(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return dateFormatter.format(date)
}

export const MONTH_NAMES_ID = [
  'Januari',
  'Februari',
  'Maret',
  'April',
  'Mei',
  'Juni',
  'Juli',
  'Agustus',
  'September',
  'Oktober',
  'November',
  'Desember',
] as const

export function formatDuesPeriodLabel(year: number, month: number, half: number): string {
  const monthName = MONTH_NAMES_ID[month - 1]
  if (!monthName) {
    return `${year} / ${month} / Half ${half}`
  }
  return `${monthName} ${year} · Half ${half}`
}

export function formatRoleLabel(role: string): string {
  switch (role) {
    case 'SUPER_ADMIN':
      return 'Super Admin'
    case 'ADMIN_RW':
      return 'Admin RW'
    default:
      return role
  }
}

/** Convert RFC3339 ISO timestamp to `datetime-local` input value (local browser time). */
export function toDatetimeLocalValue(iso: string): string {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

/** Convert `datetime-local` value to RFC3339 for the backend. */
export function fromDatetimeLocalValue(localValue: string): string {
  const date = new Date(localValue)
  if (Number.isNaN(date.getTime())) {
    throw new Error('invalid datetime')
  }
  return date.toISOString()
}
