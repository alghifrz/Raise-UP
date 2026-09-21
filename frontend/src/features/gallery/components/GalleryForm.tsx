import { useEffect, useState, type FormEvent } from 'react'
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
  image?: string
  sort_order?: string
}

const MAX_IMAGE_BYTES = 10 * 1024 * 1024
const ALLOWED_IMAGE_TYPES = new Set(['image/jpeg', 'image/png', 'image/webp'])

export function GalleryForm({
  mode,
  initial,
  submitLabel,
  onSubmitCreate,
  onSubmitUpdate,
  onCancel,
}: GalleryFormProps) {
  const [image, setImage] = useState<File | null>(null)
  const [previewUrl, setPreviewUrl] = useState(initial?.image_url ?? '')
  const [caption, setCaption] = useState(initial?.caption ?? '')
  const [sortOrder, setSortOrder] = useState(String(initial?.sort_order ?? 0))
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!image) {
      setPreviewUrl(initial?.image_url ?? '')
      return
    }
    const objectUrl = URL.createObjectURL(image)
    setPreviewUrl(objectUrl)
    return () => URL.revokeObjectURL(objectUrl)
  }, [image, initial?.image_url])

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const errors: FieldErrors = {}
    if (mode === 'create' && !image) {
      errors.image = 'Pilih foto yang akan diunggah.'
    } else if (image && !ALLOWED_IMAGE_TYPES.has(image.type)) {
      errors.image = 'Format foto harus JPG, PNG, atau WebP.'
    } else if (image && image.size > MAX_IMAGE_BYTES) {
      errors.image = 'Ukuran foto maksimal 10 MB.'
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
        if (!image) {
          return
        }
        await onSubmitCreate({
          image,
          caption: caption.trim(),
          sort_order: sort,
        })
      }
      if (mode === 'edit' && onSubmitUpdate) {
        await onSubmitUpdate({
          image: image ?? undefined,
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
      <div>
        <label className="mb-1.5 block text-sm font-medium text-[var(--color-ink)]" htmlFor="gallery-image">
          {mode === 'create' ? 'Foto' : 'Ganti foto (opsional)'}
        </label>
        <input
          id="gallery-image"
          name="image"
          type="file"
          accept="image/jpeg,image/png,image/webp"
          onChange={(event) => {
            setImage(event.target.files?.[0] ?? null)
            setFieldErrors((current) => ({ ...current, image: undefined }))
          }}
          disabled={submitting}
          required={mode === 'create'}
          className="block w-full rounded-xl border border-[var(--color-line)] bg-white px-3 py-2 text-sm text-[var(--color-ink)] file:mr-3 file:rounded-lg file:border-0 file:bg-[var(--color-primary-soft)] file:px-3 file:py-1.5 file:font-medium file:text-[var(--color-primary)]"
        />
        <p className="mt-1 text-xs text-[var(--color-muted)]">JPG, PNG, atau WebP. Maksimal 10 MB.</p>
        {fieldErrors.image ? (
          <p className="mt-1 text-sm text-[var(--color-danger)]" role="alert">
            {fieldErrors.image}
          </p>
        ) : null}
      </div>

      {previewUrl ? (
        <div>
          <p className="mb-1.5 text-sm font-medium text-[var(--color-ink)]">Preview</p>
          <ImageWithFallback
            src={previewUrl}
            alt={caption.trim() || 'Preview gambar'}
            className="h-40 w-full rounded-md"
            fallbackLabel="Preview tidak tersedia"
          />
        </div>
      ) : null}
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
