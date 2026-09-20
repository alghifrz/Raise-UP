import { useState, type FormEvent } from 'react'
import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { Textarea } from '../../../components/ui/Textarea'
import { fromDatetimeLocalValue, isValidReminderTime, toDatetimeLocalValue } from '../date'
import { toActivityErrorMessage } from '../errors'
import type { Activity, CreateActivityRequest, UpdateActivityRequest } from '../types'

type Mode = 'create' | 'edit'

type ActivityFormProps = {
  mode: Mode
  initial?: Activity
  submitLabel: string
  onSubmitCreate?: (payload: CreateActivityRequest) => Promise<void>
  onSubmitUpdate?: (payload: UpdateActivityRequest) => Promise<void>
  onCancel: () => void
}

type FieldErrors = {
  name?: string
  date?: string
  reminder_days_before?: string
  reminder_time?: string
}

export function ActivityForm({
  mode,
  initial,
  submitLabel,
  onSubmitCreate,
  onSubmitUpdate,
  onCancel,
}: ActivityFormProps) {
  const [name, setName] = useState(initial?.name ?? '')
  const [description, setDescription] = useState(initial?.description ?? '')
  const [dateLocal, setDateLocal] = useState(
    initial?.date ? toDatetimeLocalValue(initial.date) : '',
  )
  const [reminderDaysBefore, setReminderDaysBefore] = useState(
    String(initial?.reminder_days_before ?? 0),
  )
  const [reminderTime, setReminderTime] = useState(initial?.reminder_time ?? '')
  const [reminderMessage, setReminderMessage] = useState(initial?.reminder_message ?? '')
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const errors: FieldErrors = {}
    if (!name.trim()) {
      errors.name = 'Nama kegiatan wajib diisi.'
    }
    if (!dateLocal) {
      errors.date = 'Tanggal kegiatan wajib diisi.'
    }

    let rfcDate = ''
    if (dateLocal) {
      try {
        rfcDate = fromDatetimeLocalValue(dateLocal)
      } catch {
        errors.date = 'Tanggal kegiatan tidak valid.'
      }
    }

    const days = Number.parseInt(reminderDaysBefore, 10)
    if (Number.isNaN(days) || days < 0) {
      errors.reminder_days_before = 'Hari sebelum kegiatan harus angka ≥ 0.'
    }

    const trimmedTime = reminderTime.trim()
    if (trimmedTime && !isValidReminderTime(trimmedTime)) {
      errors.reminder_time = 'Jam reminder harus berformat HH:MM (24 jam).'
    }

    setFieldErrors(errors)
    if (Object.keys(errors).length > 0) {
      return
    }

    setSubmitting(true)
    try {
      if (mode === 'create' && onSubmitCreate) {
        const payload: CreateActivityRequest = {
          name: name.trim(),
          description: description.trim(),
          date: rfcDate,
          reminder_days_before: days,
          reminder_message: reminderMessage.trim(),
        }
        if (trimmedTime) {
          payload.reminder_time = trimmedTime
        }
        await onSubmitCreate(payload)
      }

      if (mode === 'edit' && onSubmitUpdate) {
        const payload: UpdateActivityRequest = {
          name: name.trim(),
          description: description.trim(),
          date: rfcDate,
          reminder_days_before: days,
          reminder_message: reminderMessage.trim(),
        }
        // Backend rejects empty reminder_time on update; only send when valid HH:MM.
        if (trimmedTime) {
          payload.reminder_time = trimmedTime
        }
        await onSubmitUpdate(payload)
      }
    } catch (error) {
      setFormError(toActivityErrorMessage(error))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
      <Input
        name="name"
        label="Nama kegiatan"
        value={name}
        onChange={(event) => setName(event.target.value)}
        error={fieldErrors.name}
        disabled={submitting}
        required
      />
      <Textarea
        name="description"
        label="Deskripsi"
        value={description}
        onChange={(event) => setDescription(event.target.value)}
        disabled={submitting}
      />
      <Input
        name="date"
        label="Tanggal & waktu"
        type="datetime-local"
        value={dateLocal}
        onChange={(event) => setDateLocal(event.target.value)}
        error={fieldErrors.date}
        disabled={submitting}
        required
      />

      <fieldset className="space-y-3 rounded-lg border border-[var(--color-line)] p-4">
        <legend className="px-1 text-sm font-medium text-[var(--color-ink)]">Reminder</legend>
        <p className="text-xs text-[var(--color-muted)]">
          Konfigurasi pengingat kegiatan. Penjadwalan otomatis belum dijalankan di fase ini.
        </p>
        <Input
          name="reminder_days_before"
          label="Hari sebelum kegiatan"
          type="number"
          min={0}
          value={reminderDaysBefore}
          onChange={(event) => setReminderDaysBefore(event.target.value)}
          error={fieldErrors.reminder_days_before}
          disabled={submitting}
        />
        <Input
          name="reminder_time"
          label="Jam reminder (HH:MM)"
          placeholder="09:00"
          value={reminderTime}
          onChange={(event) => setReminderTime(event.target.value)}
          error={fieldErrors.reminder_time}
          disabled={submitting}
        />
        <Textarea
          name="reminder_message"
          label="Pesan reminder"
          value={reminderMessage}
          onChange={(event) => setReminderMessage(event.target.value)}
          disabled={submitting}
        />
      </fieldset>

      {formError ? (
        <p
          className="rounded-2xl border border-red-100 bg-[var(--color-danger-soft)] px-3 py-2 text-sm text-[var(--color-danger)]"
          role="alert"
        >
          {formError}
        </p>
      ) : null}

      <div className="flex justify-end gap-2">
        <Button type="button" variant="secondary" onClick={onCancel} disabled={submitting}>
          Batal
        </Button>
        <Button type="submit" loading={submitting}>
          {submitLabel}
        </Button>
      </div>
    </form>
  )
}
