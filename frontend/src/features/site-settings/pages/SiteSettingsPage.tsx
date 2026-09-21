import { useState, type FormEvent } from 'react'
import { Button } from '../../../components/ui/Button'
import { Card } from '../../../components/ui/Card'
import { ErrorState } from '../../../components/ui/ErrorState'
import { ImageWithFallback } from '../../../components/ui/ImageWithFallback'
import { InlineAlert } from '../../../components/ui/InlineAlert'
import { Input } from '../../../components/ui/Input'
import { LoadingState } from '../../../components/ui/LoadingState'
import { PageHeader } from '../../../components/ui/PageHeader'
import { Textarea } from '../../../components/ui/Textarea'
import { formatDateTime } from '../../../lib/utils'
import { toSiteSettingsErrorMessage } from '../errors'
import { useSiteSettings, useUpdateSiteSettings } from '../hooks'
import type { SiteSettings } from '../types'

/** Empty string clears nullable URL fields (COALESCE); null would leave the prior value. */
function optionalUrl(value: string): string {
  return value.trim()
}

function looksLikeUrl(value: string): boolean {
  if (!value.trim()) {
    return true
  }
  try {
    const parsed = new URL(value.trim())
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

/** Accept either a raw URL or Google Maps' complete iframe snippet. */
function normalizeEmbedUrl(value: string): string | null {
  const input = value.trim()
  if (!input) {
    return ''
  }
  if (looksLikeUrl(input)) {
    return input
  }

  const srcMatch = input.match(/<iframe\b[^>]*\bsrc\s*=\s*(["'])(.*?)\1/i)
  if (!srcMatch) {
    return null
  }
  const src = (srcMatch[2] ?? '').replaceAll('&amp;', '&').trim()
  return looksLikeUrl(src) ? src : null
}

type SiteSettingsFormProps = {
  initial: SiteSettings
}

function SiteSettingsForm({ initial }: SiteSettingsFormProps) {
  const updateMutation = useUpdateSiteSettings()

  const [siteName, setSiteName] = useState(initial.site_name)
  const [tagline, setTagline] = useState(initial.tagline)
  const [chairmanName, setChairmanName] = useState(initial.chairman_name)
  const [chairmanRole, setChairmanRole] = useState(initial.chairman_role)
  const [chairmanQuote, setChairmanQuote] = useState(initial.chairman_quote)
  const [chairmanPhotoUrl, setChairmanPhotoUrl] = useState(initial.chairman_photo_url ?? '')
  const [chairmanPhotoStoragePath, setChairmanPhotoStoragePath] = useState(
    initial.chairman_photo_storage_path ?? '',
  )
  const [mapTitle, setMapTitle] = useState(initial.map_title)
  const [mapDescription, setMapDescription] = useState(initial.map_description)
  const [mapsUrl, setMapsUrl] = useState(initial.maps_url ?? '')
  const [embedUrl, setEmbedUrl] = useState(initial.embed_url ?? '')
  const [address, setAddress] = useState(initial.address)
  const [phone, setPhone] = useState(initial.phone)
  const [whatsappUrl, setWhatsappUrl] = useState(initial.whatsapp_url ?? '')
  const [footerBlurb, setFooterBlurb] = useState(initial.footer_blurb)
  const [formError, setFormError] = useState<string | null>(null)
  const [feedback, setFeedback] = useState<string | null>(null)
  const [savedAt, setSavedAt] = useState(initial.updated_at)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)
    setFeedback(null)

    const normalizedEmbedUrl = normalizeEmbedUrl(embedUrl)
    if (normalizedEmbedUrl === null) {
      setFormError(
        'Kode embed peta tidak valid. Tempel kode <iframe> dari Google Maps atau URL embed-nya.',
      )
      return
    }

    if (
      !looksLikeUrl(chairmanPhotoUrl) ||
      !looksLikeUrl(mapsUrl) ||
      !looksLikeUrl(whatsappUrl)
    ) {
      setFormError('URL yang diisi harus valid (http/https) atau dikosongkan.')
      return
    }

    try {
      const updated = await updateMutation.mutateAsync({
        site_name: siteName.trim(),
        tagline: tagline.trim(),
        chairman_name: chairmanName.trim(),
        chairman_role: chairmanRole.trim(),
        chairman_quote: chairmanQuote.trim(),
        chairman_photo_url: optionalUrl(chairmanPhotoUrl),
        chairman_photo_storage_path: optionalUrl(chairmanPhotoStoragePath),
        map_title: mapTitle.trim(),
        map_description: mapDescription.trim(),
        maps_url: optionalUrl(mapsUrl),
        embed_url: normalizedEmbedUrl,
        address: address.trim(),
        phone: phone.trim(),
        whatsapp_url: optionalUrl(whatsappUrl),
        footer_blurb: footerBlurb.trim(),
      })
      setEmbedUrl(updated.embed_url ?? '')
      setSavedAt(updated.updated_at)
      setFeedback('Pengaturan situs berhasil disimpan.')
    } catch (err) {
      setFormError(toSiteSettingsErrorMessage(err))
    }
  }

  const mapPreviewUrl = normalizeEmbedUrl(embedUrl)

  return (
    <>
      {feedback ? <InlineAlert>{feedback}</InlineAlert> : null}

      <form className="space-y-4" onSubmit={handleSubmit} noValidate>
        <Card title="Identitas Situs">
          <div className="grid gap-4 sm:grid-cols-2">
            <Input
              name="site_name"
              label="Nama situs"
              value={siteName}
              onChange={(event) => setSiteName(event.target.value)}
              disabled={updateMutation.isPending}
            />
            <Input
              name="tagline"
              label="Tagline"
              value={tagline}
              onChange={(event) => setTagline(event.target.value)}
              disabled={updateMutation.isPending}
            />
          </div>
        </Card>

        <Card title="Ketua / Pimpinan">
          <div className="grid gap-4 sm:grid-cols-2">
            <Input
              name="chairman_name"
              label="Nama"
              value={chairmanName}
              onChange={(event) => setChairmanName(event.target.value)}
              disabled={updateMutation.isPending}
            />
            <Input
              name="chairman_role"
              label="Jabatan"
              value={chairmanRole}
              onChange={(event) => setChairmanRole(event.target.value)}
              disabled={updateMutation.isPending}
            />
            <div className="sm:col-span-2">
              <Textarea
                name="chairman_quote"
                label="Kutipan"
                value={chairmanQuote}
                onChange={(event) => setChairmanQuote(event.target.value)}
                disabled={updateMutation.isPending}
              />
            </div>
            <Input
              name="chairman_photo_url"
              label="URL foto"
              value={chairmanPhotoUrl}
              onChange={(event) => setChairmanPhotoUrl(event.target.value)}
              disabled={updateMutation.isPending}
            />
            <Input
              name="chairman_photo_storage_path"
              label="Storage path foto"
              value={chairmanPhotoStoragePath}
              onChange={(event) => setChairmanPhotoStoragePath(event.target.value)}
              disabled={updateMutation.isPending}
            />
            {chairmanPhotoUrl.trim() ? (
              <div className="sm:col-span-2">
                <p className="mb-1.5 text-sm font-medium">Preview foto</p>
                <ImageWithFallback
                  src={chairmanPhotoUrl.trim()}
                  alt={chairmanName.trim() || 'Foto ketua'}
                  className="h-40 w-40 rounded-md"
                />
              </div>
            ) : null}
          </div>
        </Card>

        <Card title="Lokasi">
          <div className="grid gap-4 sm:grid-cols-2">
            <Input
              name="map_title"
              label="Judul peta"
              value={mapTitle}
              onChange={(event) => setMapTitle(event.target.value)}
              disabled={updateMutation.isPending}
            />
            <Input
              name="maps_url"
              label="Link Google Maps"
              value={mapsUrl}
              onChange={(event) => setMapsUrl(event.target.value)}
              disabled={updateMutation.isPending}
              placeholder="https://maps.app.goo.gl/..."
              hint="Buka Google Maps → Bagikan → Salin link. Link ini digunakan oleh tombol petunjuk arah."
            />
            <div className="sm:col-span-2">
              <Textarea
                name="map_description"
                label="Deskripsi peta"
                value={mapDescription}
                onChange={(event) => setMapDescription(event.target.value)}
                disabled={updateMutation.isPending}
              />
            </div>
            <Input
              name="address"
              label="Alamat"
              value={address}
              onChange={(event) => setAddress(event.target.value)}
              disabled={updateMutation.isPending}
            />
            <div className="sm:col-span-2">
              <Textarea
                name="embed_url"
                label="Embed Google Maps"
                value={embedUrl}
                onChange={(event) => setEmbedUrl(event.target.value)}
                disabled={updateMutation.isPending}
                placeholder='<iframe src="https://www.google.com/maps/embed?pb=..." ...></iframe>'
                hint="Google Maps → Bagikan → Sematkan peta → Salin HTML. Tempel seluruh kode iframe di sini; URL-nya akan diambil otomatis."
                className="min-h-32 font-mono text-xs"
              />
            </div>
            {mapPreviewUrl ? (
              <div className="sm:col-span-2">
                <p className="mb-1.5 text-sm font-semibold text-[var(--color-ink)]">
                  Preview peta
                </p>
                <iframe
                  title={mapTitle.trim() || 'Preview lokasi'}
                  src={mapPreviewUrl}
                  className="h-72 w-full rounded-2xl border border-[var(--color-line)]"
                  loading="lazy"
                  referrerPolicy="strict-origin-when-cross-origin"
                  allowFullScreen
                />
              </div>
            ) : null}
          </div>
        </Card>

        <Card title="Kontak">
          <div className="grid gap-4 sm:grid-cols-2">
            <Input
              name="phone"
              label="Telepon"
              value={phone}
              onChange={(event) => setPhone(event.target.value)}
              disabled={updateMutation.isPending}
            />
            <Input
              name="whatsapp_url"
              label="WhatsApp URL"
              value={whatsappUrl}
              onChange={(event) => setWhatsappUrl(event.target.value)}
              disabled={updateMutation.isPending}
            />
          </div>
        </Card>

        <Card title="Footer">
          <Textarea
            name="footer_blurb"
            label="Deskripsi footer"
            value={footerBlurb}
            onChange={(event) => setFooterBlurb(event.target.value)}
            disabled={updateMutation.isPending}
          />
          <p className="mt-3 text-xs text-[var(--color-muted)]">
            Terakhir diperbarui: {formatDateTime(savedAt)}
          </p>
        </Card>

        {formError ? <InlineAlert tone="danger">{formError}</InlineAlert> : null}

        <div className="flex justify-end">
          <Button type="submit" loading={updateMutation.isPending}>
            Simpan Perubahan
          </Button>
        </div>
      </form>
    </>
  )
}

export function SiteSettingsPage() {
  const { data, isLoading, isError, refetch, error } = useSiteSettings()

  if (isLoading) {
    return <LoadingState label="Memuat pengaturan situs…" />
  }

  if (isError || !data) {
    return (
      <div className="space-y-4">
        <PageHeader title="Pengaturan Situs" />
        <ErrorState
          title="Gagal memuat pengaturan situs."
          message={toSiteSettingsErrorMessage(error)}
          onRetry={() => {
            void refetch()
          }}
        />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title="Pengaturan Situs"
        description="Kelola identitas, kontak, dan konten publik situs."
      />
      <SiteSettingsForm key={data.updated_at} initial={data} />
    </div>
  )
}
