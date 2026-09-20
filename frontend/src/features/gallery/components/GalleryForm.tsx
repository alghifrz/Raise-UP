import { useState, type FormEvent } from 'react'
import { Button } from '../../../components/ui/Button'
import { ImageWithFallback } from '../../../components/ui/ImageWithFallback'
import { Input } from '../../../components/ui/Input'
import { Textarea } from '../../../components/ui/Textarea'
import { toGalleryErrorMessage } from '../errors'
import type { CreateGalleryRequest, GalleryItem, UpdateGalleryRequest } from '../types'

type Mode = 'create' | 'edit'

type GalleryFormProps = {
  mode: Mode
  initial?: GalleryItem
  submitLabel: string
  onSubmitCreate?: (payload: CreateGalleryRequest) => Promise<void>
  onSubmitUpdate?: (payload: UpdateGalleryRequest) => Promise<void>
  onCancel: () => void
}

type FieldErrors = {
  image_url?: string
  sort_order?: string
}

export function GalleryForm({
  mode,
  initial,
  submitLabel,
  onSubmitCreate,
  onSubmitUpdate,
  onCancel,
}: GalleryFormProps) {
  const [imageUrl, setImageUrl] = useState(initial?.image_url ?? '')
  const [storagePath, setStoragePath] = useState(initial?.storage_path ?? '')
  const [caption, setCaption] = useState(initial?.caption ?? '')
  const [sortOrder, setSortOrder] = useState(String(initial?.sort_order ?? 0))
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const errors: FieldErrors = {}
    if (!imageUrl.trim()) {
      errors.image_url = 'Gambar wajib memiliki URL.'
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
          image_url: imageUrl.trim(),
          storage_path: storagePath.trim(),
          caption: caption.trim(),
          sort_order: sort,
        })
      }
      if (mode === 'edit' && onSubmitUpdate) {
        await onSubmitUpdate({
          image_url: imageUrl.trim(),
          storage_path: storagePath.trim(),
          caption: caption.trim(),
          sort_order: sort,
        })
      }
    } catch (error) {
      setFormError(toGalleryErrorMessage(error))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
      <Input
        name="image_url"
        label="URL Gambar"
        value={imageUrl}
        onChange={(event) => setImageUrl(event.target.value)}
        error={fieldErrors.image_url}
        disabled={submitting}
        required
      />

      {imageUrl.trim() ? (
        <div>
          <p className="mb-1.5 text-sm font-medium text-[var(--color-ink)]">Preview</p>
          <ImageWithFallback
            src={imageUrl.trim()}
            alt={caption.trim() || 'Preview gambar'}
            className="h-40 w-full rounded-md"
            fallbackLabel="Preview tidak tersedia"
          />
        </div>
      ) : null}

      <Input
        name="storage_path"
        label="Storage Path (opsional)"
        value={storagePath}
        onChange={(event) => setStoragePath(event.target.value)}
        disabled={submitting}
      />
      <Textarea
        name="caption"
        label="Caption"
        value={caption}
        onChange={(event) => setCaption(event.target.value)}
        disabled={submitting}
      />
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
