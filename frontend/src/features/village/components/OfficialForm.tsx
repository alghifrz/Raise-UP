import { useState, type FormEvent } from 'react'
import { Button } from '../../../components/ui/Button'
import { ImageWithFallback } from '../../../components/ui/ImageWithFallback'
import { Input } from '../../../components/ui/Input'
import type {
  CreateVillageOfficialRequest,
  UpdateVillageOfficialRequest,
  VillageOfficial,
} from '../types'
import { toVillageErrorMessage } from '../errors'

type Mode = 'create' | 'edit'

type OfficialFormProps = {
  mode: Mode
  initial?: VillageOfficial
  submitLabel: string
  onSubmitCreate?: (payload: CreateVillageOfficialRequest) => Promise<void>
  onSubmitUpdate?: (payload: UpdateVillageOfficialRequest) => Promise<void>
  onCancel: () => void
}

type FieldErrors = {
  name?: string
  position?: string
  sort_order?: string
}

export function OfficialForm({
  mode,
  initial,
  submitLabel,
  onSubmitCreate,
  onSubmitUpdate,
  onCancel,
}: OfficialFormProps) {
  const [name, setName] = useState(initial?.name ?? '')
  const [position, setPosition] = useState(initial?.position ?? '')
  const [photoUrl, setPhotoUrl] = useState(initial?.photo_url ?? '')
  const [sortOrder, setSortOrder] = useState(String(initial?.sort_order ?? 0))
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const errors: FieldErrors = {}
    if (!name.trim()) {
      errors.name = 'Nama pengurus wajib diisi.'
    }
    if (!position.trim()) {
      errors.position = 'Jabatan pengurus wajib diisi.'
    }
    const sort = Number.parseInt(sortOrder, 10)
    if (!/^\d+$/.test(sortOrder.trim()) || Number.isNaN(sort) || sort < 0) {
      errors.sort_order = 'Urutan harus bilangan bulat ≥ 0.'
    }

    setFieldErrors(errors)
    if (Object.keys(errors).length > 0) {
      return
    }

    setSubmitting(true)
    try {
      if (mode === 'create' && onSubmitCreate) {
        await onSubmitCreate({
          name: name.trim(),
          position: position.trim(),
          photo_url: photoUrl.trim(),
          sort_order: sort,
        })
      }
      if (mode === 'edit' && onSubmitUpdate) {
        await onSubmitUpdate({
          name: name.trim(),
          position: position.trim(),
          photo_url: photoUrl.trim(),
          sort_order: sort,
        })
      }
    } catch (error) {
      setFormError(toVillageErrorMessage(error))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
      <Input
        name="name"
        label="Nama"
        value={name}
        onChange={(event) => setName(event.target.value)}
        error={fieldErrors.name}
        disabled={submitting}
        required
      />
      <Input
        name="position"
        label="Jabatan"
        value={position}
        onChange={(event) => setPosition(event.target.value)}
        error={fieldErrors.position}
        disabled={submitting}
        required
      />
      <Input
        name="photo_url"
        label="URL Foto (opsional)"
        value={photoUrl}
        onChange={(event) => setPhotoUrl(event.target.value)}
        disabled={submitting}
      />
      {photoUrl.trim() ? (
        <ImageWithFallback
          src={photoUrl.trim()}
          alt={name.trim() || 'Preview foto pengurus'}
          className="h-32 w-32 rounded-md"
        />
      ) : null}
      <Input
        name="sort_order"
        label="Urutan"
        inputMode="numeric"
        value={sortOrder}
        onChange={(event) => setSortOrder(event.target.value)}
        error={fieldErrors.sort_order}
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
