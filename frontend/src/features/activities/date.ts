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

/** Convert RFC3339 to a date input value using the Jakarta calendar date. */
export function toDateInputValue(iso: string): string {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  const parts = new Intl.DateTimeFormat('en-US', {
    timeZone: 'Asia/Jakarta',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).formatToParts(date)
  const value = (type: Intl.DateTimeFormatPartTypes) =>
    parts.find((part) => part.type === type)?.value ?? ''
  return `${value('year')}-${value('month')}-${value('day')}`
}

/** Convert a date input value to Jakarta midnight as RFC3339/UTC. */
export function fromDateInputValue(value: string): string {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) {
    throw new Error('invalid date')
  }
  const date = new Date(`${value}T00:00:00+07:00`)
  if (Number.isNaN(date.getTime())) {
    throw new Error('invalid date')
  }
  return date.toISOString()
}

const REMINDER_TIME_PATTERN = /^([01]\d|2[0-3]):([0-5]\d)$/

export function isValidReminderTime(value: string): boolean {
  return REMINDER_TIME_PATTERN.test(value)
}
