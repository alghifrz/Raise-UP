import { useState, type FormEvent } from 'react'
import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { Select } from '../../../components/ui/Select'
import { Textarea } from '../../../components/ui/Textarea'
import { toComplaintErrorMessage } from '../errors'
import type {
  Complaint,
  ComplaintUrgency,
  CreateComplaintRequest,
  UpdateComplaintRequest,
} from '../types'
import { COMPLAINT_URGENCY_OPTIONS } from '../types'
import { ResidentPicker, type ResidentOption } from './ResidentPicker'

type Mode = 'create' | 'edit'

type ComplaintFormProps = {
  mode: Mode
  initial?: Complaint
  submitLabel: string
  onSubmitCreate?: (payload: CreateComplaintRequest) => Promise<void>
  onSubmitUpdate?: (payload: UpdateComplaintRequest) => Promise<void>
  onCancel: () => void
}

type FieldErrors = {
  resident_name?: string
  phone?: string
  block?: string
  category?: string
  urgency?: string
  message?: string
  linkMode?: string
}

export function ComplaintForm({
  mode,
  initial,
  submitLabel,
  onSubmitCreate,
  onSubmitUpdate,
  onCancel,
}: ComplaintFormProps) {
  const [linkMode, setLinkMode] = useState<'linked' | 'manual'>(
    mode === 'edit' ? (initial?.resident_id ? 'linked' : 'manual') : 'linked',
  )
  const [selectedResident, setSelectedResident] = useState<ResidentOption | null>(
    initial?.resident_id
      ? {
          id: initial.resident_id,
          name: initial.resident_name,
          phone: initial.phone,
        }
      : null,
  )
  const [residentName, setResidentName] = useState(initial?.resident_name ?? '')
  const [phone, setPhone] = useState(initial?.phone ?? '')
  const [block, setBlock] = useState(initial?.block ?? '')
  const [category, setCategory] = useState(initial?.category ?? '')
  const [urgency, setUrgency] = useState<ComplaintUrgency | ''>(initial?.urgency ?? 'NORMAL')
  const [message, setMessage] = useState(initial?.message ?? '')
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const errors: FieldErrors = {}
    if (!category.trim()) {
      errors.category = 'Kategori wajib diisi.'
    }
    if (!message.trim()) {
      errors.message = 'Pesan wajib diisi.'
    }
    if (!urgency) {
      errors.urgency = 'Urgensi wajib dipilih.'
    }

    if (linkMode === 'linked') {
      if (mode === 'create' && !selectedResident) {
        errors.linkMode = 'Pilih warga atau gunakan mode tanpa tautan warga.'
      }
    } else if (mode === 'create') {
      if (!residentName.trim()) {
        errors.resident_name = 'Nama warga wajib diisi.'
      }
      if (!phone.trim()) {
        errors.phone = 'Nomor telepon wajib diisi.'
      }
    }

    setFieldErrors(errors)
    if (Object.keys(errors).length > 0) {
      return
    }

    setSubmitting(true)
    try {
      if (mode === 'create' && onSubmitCreate) {
        const payload: CreateComplaintRequest = {
          block: block.trim(),
          category: category.trim(),
          urgency: urgency as ComplaintUrgency,
          message: message.trim(),
        }

        if (linkMode === 'linked' && selectedResident) {
          payload.resident_id = selectedResident.id
        } else {
          payload.resident_name = residentName.trim()
          payload.phone = phone.trim()
        }

        await onSubmitCreate(payload)
      }

      if (mode === 'edit' && onSubmitUpdate) {
        const payload: UpdateComplaintRequest = {
          block: block.trim(),
          category: category.trim(),
          urgency: urgency as ComplaintUrgency,
          message: message.trim(),
        }

        if (linkMode === 'linked') {
          payload.resident_id = selectedResident?.id ?? initial?.resident_id ?? null
          if (!payload.resident_id) {
            setFieldErrors({ linkMode: 'Pilih warga yang akan ditautkan.' })
            setSubmitting(false)
            return
          }
        } else {
          payload.resident_id = null
        }

        await onSubmitUpdate(payload)
      }
    } catch (error) {
      setFormError(toComplaintErrorMessage(error))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
      <fieldset className="space-y-3">
        <legend className="text-sm font-medium text-[var(--color-ink)]">Sumber data warga</legend>
        <div className="flex flex-wrap gap-4 text-sm">
          <label className="inline-flex items-center gap-2">
            <input
              type="radio"
              name="linkMode"
              checked={linkMode === 'linked'}
              onChange={() => setLinkMode('linked')}
              disabled={submitting}
            />
            Tautkan ke warga
          </label>
          <label className="inline-flex items-center gap-2">
            <input
              type="radio"
              name="linkMode"
              checked={linkMode === 'manual'}
              onChange={() => setLinkMode('manual')}
              disabled={submitting}
            />
            Tanpa tautan warga
          </label>
        </div>
        {fieldErrors.linkMode ? (
          <p className="text-sm text-[var(--color-danger)]">{fieldErrors.linkMode}</p>
        ) : null}
      </fieldset>

      {linkMode === 'linked' ? (
        <ResidentPicker selected={selectedResident} onSelect={setSelectedResident} disabled={submitting} />
      ) : mode === 'create' ? (
        <>
          <Input
            name="resident_name"
            label="Nama warga"
            value={residentName}
            onChange={(event) => setResidentName(event.target.value)}
            error={fieldErrors.resident_name}
            disabled={submitting}
            required
          />
          <Input
            name="phone"
            label="No. Telepon"
            value={phone}
            onChange={(event) => setPhone(event.target.value)}
            error={fieldErrors.phone}
            disabled={submitting}
            required
          />
        </>
      ) : (
        <p className="rounded-2xl border border-[var(--color-line)] bg-[var(--color-accent-soft)] px-3 py-2 text-sm text-[var(--color-muted)]">
          Snapshot nama/telepon tetap disimpan: {initial?.resident_name} · {initial?.phone}
        </p>
      )}

      <Input
        name="block"
        label="Blok"
        value={block}
        onChange={(event) => setBlock(event.target.value)}
        error={fieldErrors.block}
        disabled={submitting}
      />
      <Input
        name="category"
        label="Kategori"
        value={category}
        onChange={(event) => setCategory(event.target.value)}
        error={fieldErrors.category}
        disabled={submitting}
        required
      />
      <Select
        name="urgency"
        label="Urgensi"
        value={urgency}
        onChange={(event) => setUrgency(event.target.value as ComplaintUrgency | '')}
        options={COMPLAINT_URGENCY_OPTIONS}
        error={fieldErrors.urgency}
        disabled={submitting}
        required
      />
      <Textarea
        name="message"
        label="Pesan"
        value={message}
        onChange={(event) => setMessage(event.target.value)}
        error={fieldErrors.message}
        disabled={submitting}
        required
      />

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
