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

const REMINDER_TIME_PATTERN = /^([01]\d|2[0-3]):([0-5]\d)$/

export function isValidReminderTime(value: string): boolean {
  return REMINDER_TIME_PATTERN.test(value)
}
