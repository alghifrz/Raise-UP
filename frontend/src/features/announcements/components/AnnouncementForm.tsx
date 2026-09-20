import { useEffect, useState, type FormEvent } from 'react'
import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { Select } from '../../../components/ui/Select'
import { Textarea } from '../../../components/ui/Textarea'
import { getResident } from '../../residents/api'
import { toAnnouncementErrorMessage } from '../errors'
import type {
  Announcement,
  AnnouncementVisibility,
  CreateAnnouncementRequest,
  UpdateAnnouncementRequest,
} from '../types'
import { ANNOUNCEMENT_VISIBILITY_OPTIONS } from '../types'
import { MultiResidentPicker, type ResidentOption } from './MultiResidentPicker'

type Mode = 'create' | 'edit'

type AnnouncementFormProps = {
  mode: Mode
  initial?: Announcement
  submitLabel: string
  onSubmitCreate?: (payload: CreateAnnouncementRequest) => Promise<void>
  onSubmitUpdate?: (payload: UpdateAnnouncementRequest) => Promise<void>
  onCancel: () => void
}

type FieldErrors = {
  title?: string
  body?: string
  category?: string
  visibility?: string
  recipients?: string
}

export function AnnouncementForm({
  mode,
  initial,
  submitLabel,
  onSubmitCreate,
  onSubmitUpdate,
  onCancel,
}: AnnouncementFormProps) {
  const [title, setTitle] = useState(initial?.title ?? '')
  const [excerpt, setExcerpt] = useState(initial?.excerpt ?? '')
  const [body, setBody] = useState(initial?.body ?? '')
  const [category, setCategory] = useState(initial?.category ?? '')
  const [visibility, setVisibility] = useState<AnnouncementVisibility | ''>(
    initial?.visibility ?? 'PUBLIC',
  )
  const [thumbnailUrl, setThumbnailUrl] = useState(initial?.thumbnail_url ?? '')
  const [recipients, setRecipients] = useState<ResidentOption[]>([])
  const [recipientsHydrated, setRecipientsHydrated] = useState(
    !(mode === 'edit' && (initial?.recipient_ids.length ?? 0) > 0),
  )
  const [recipientLoadError, setRecipientLoadError] = useState<string | null>(null)
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (mode !== 'edit' || !initial || initial.recipient_ids.length === 0) {
      return
    }

    const recipientIds = initial.recipient_ids
    let cancelled = false

    void (async () => {
      try {
        const items = await Promise.all(recipientIds.map((id) => getResident(id)))
        if (!cancelled) {
          setRecipients(
            items.map((item) => ({
              id: item.id,
              name: item.name,
              phone: item.phone,
            })),
          )
          setRecipientLoadError(null)
        }
      } catch {
        if (!cancelled) {
          setRecipientLoadError('Gagal memuat data penerima.')
          setRecipients(
            recipientIds.map((id) => ({
              id,
              name: id,
              phone: '—',
            })),
          )
        }
      } finally {
        if (!cancelled) {
          setRecipientsHydrated(true)
        }
      }
    })()

    return () => {
      cancelled = true
    }
  }, [mode, initial])

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const errors: FieldErrors = {}
    if (!title.trim()) {
      errors.title = 'Judul pengumuman wajib diisi.'
    }
    if (!body.trim()) {
      errors.body = 'Isi pengumuman wajib diisi.'
    }
    if (!category.trim()) {
      errors.category = 'Kategori wajib diisi.'
    }
    if (!visibility) {
      errors.visibility = 'Visibilitas wajib dipilih.'
    }
    if (visibility === 'PRIVATE' && recipients.length === 0) {
      errors.recipients = 'Pengumuman privat harus memiliki minimal satu penerima.'
    }

    setFieldErrors(errors)
    if (Object.keys(errors).length > 0) {
      return
    }

    setSubmitting(true)
    try {
      const recipientIdsPayload =
        visibility === 'PRIVATE' ? recipients.map((item) => item.id) : []

      if (mode === 'create' && onSubmitCreate) {
        const payload: CreateAnnouncementRequest = {
          title: title.trim(),
          excerpt: excerpt.trim(),
          body: body.trim(),
          category: category.trim(),
          visibility: visibility as AnnouncementVisibility,
          recipient_ids: recipientIdsPayload,
        }
        const trimmedThumb = thumbnailUrl.trim()
        if (trimmedThumb) {
          payload.thumbnail_url = trimmedThumb
        }
        await onSubmitCreate(payload)
      }

      if (mode === 'edit' && onSubmitUpdate) {
        const payload: UpdateAnnouncementRequest = {
          title: title.trim(),
          excerpt: excerpt.trim(),
          body: body.trim(),
          category: category.trim(),
          visibility: visibility as AnnouncementVisibility,
          thumbnail_url: thumbnailUrl.trim() || null,
          recipient_ids: recipientIdsPayload,
        }
        await onSubmitUpdate(payload)
      }
    } catch (error) {
      setFormError(toAnnouncementErrorMessage(error))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
      <Input
        name="title"
        label="Judul"
        value={title}
        onChange={(event) => setTitle(event.target.value)}
        error={fieldErrors.title}
        disabled={submitting}
        required
      />
      <Input
        name="excerpt"
        label="Ringkasan"
        value={excerpt}
        onChange={(event) => setExcerpt(event.target.value)}
        disabled={submitting}
      />
      <Textarea
        name="body"
        label="Isi"
        value={body}
        onChange={(event) => setBody(event.target.value)}
        error={fieldErrors.body}
        disabled={submitting}
        required
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
        name="visibility"
        label="Visibilitas"
        value={visibility}
        onChange={(event) => {
          const next = event.target.value as AnnouncementVisibility | ''
          setVisibility(next)
          if (next === 'PUBLIC') {
            setRecipients([])
          }
        }}
        options={ANNOUNCEMENT_VISIBILITY_OPTIONS}
        error={fieldErrors.visibility}
        disabled={submitting}
        required
      />
      <Input
        name="thumbnail_url"
        label="URL Thumbnail (opsional)"
        value={thumbnailUrl}
        onChange={(event) => setThumbnailUrl(event.target.value)}
        disabled={submitting}
      />

      {visibility === 'PRIVATE' ? (
        <div className="space-y-2">
          <p className="text-sm font-medium text-[var(--color-ink)]">Penerima</p>
          {!recipientsHydrated ? (
            <p className="text-sm text-[var(--color-muted)]">Memuat penerima…</p>
          ) : (
            <>
              {recipientLoadError ? (
                <p className="text-sm text-amber-800">{recipientLoadError}</p>
              ) : null}
              <MultiResidentPicker
                selected={recipients}
                onChange={setRecipients}
                disabled={submitting}
                error={fieldErrors.recipients}
              />
            </>
          )}
        </div>
      ) : null}

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
