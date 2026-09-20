import { useState, type FormEvent } from 'react'
import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { Select } from '../../../components/ui/Select'
import { toResidentErrorMessage } from '../errors'
import type { CreateResidentRequest, Resident, ResidentGender } from '../types'
import { RESIDENT_GENDER_OPTIONS } from '../types'

type ResidentFormValues = {
  name: string
  phone: string
  gender: ResidentGender | ''
}

type FieldErrors = {
  name?: string
  phone?: string
  gender?: string
}

type ResidentFormProps = {
  initial?: Resident
  submitLabel: string
  onSubmit: (payload: CreateResidentRequest) => Promise<void>
  onCancel: () => void
}

function validate(values: ResidentFormValues): FieldErrors {
  const errors: FieldErrors = {}
  if (!values.name.trim()) {
    errors.name = 'Nama wajib diisi.'
  }
  if (!values.phone.trim()) {
    errors.phone = 'Nomor telepon wajib diisi.'
  }
  if (!values.gender) {
    errors.gender = 'Jenis kelamin wajib dipilih.'
  }
  return errors
}

export function ResidentForm({ initial, submitLabel, onSubmit, onCancel }: ResidentFormProps) {
  const [values, setValues] = useState<ResidentFormValues>({
    name: initial?.name ?? '',
    phone: initial?.phone ?? '',
    gender: initial?.gender ?? '',
  })
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const errors = validate(values)
    setFieldErrors(errors)
    if (errors.name || errors.phone || errors.gender) {
      return
    }

    setSubmitting(true)
    try {
      await onSubmit({
        name: values.name.trim(),
        phone: values.phone.trim(),
        gender: values.gender as ResidentGender,
      })
    } catch (error) {
      setFormError(toResidentErrorMessage(error))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
      <Input
        name="name"
        label="Nama"
        value={values.name}
        onChange={(event) => setValues((current) => ({ ...current, name: event.target.value }))}
        error={fieldErrors.name}
        disabled={submitting}
        required
      />
      <Input
        name="phone"
        label="No. Telepon"
        value={values.phone}
        onChange={(event) => setValues((current) => ({ ...current, phone: event.target.value }))}
        error={fieldErrors.phone}
        disabled={submitting}
        required
      />
      <Select
        name="gender"
        label="Jenis Kelamin"
        value={values.gender}
        onChange={(event) =>
          setValues((current) => ({
            ...current,
            gender: event.target.value as ResidentGender | '',
          }))
        }
        options={RESIDENT_GENDER_OPTIONS}
        placeholder="Pilih jenis kelamin"
        error={fieldErrors.gender}
        disabled={submitting}
        required
      />

      {formError ? (
        <p className="rounded-2xl border border-red-100 bg-[var(--color-danger-soft)] px-3 py-2 text-sm text-[var(--color-danger)]" role="alert">
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
