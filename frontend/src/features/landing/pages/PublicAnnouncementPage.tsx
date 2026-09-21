import { Link, useParams } from 'react-router-dom'
import { formatDateTime } from '../../../lib/utils'
import { LandingNav } from '../components/LandingNav'
import { LandingFooter } from '../components/Sections'
import { LandingContainer } from '../components/ui'
import { landingBrand } from '../data'
import { usePublicAnnouncement, usePublicSiteSettings } from '../hooks'

export function PublicAnnouncementPage() {
  const { id } = useParams<{ id: string }>()
  const announcementQuery = usePublicAnnouncement(id)
  const settingsQuery = usePublicSiteSettings()
  const settings = settingsQuery.data
  const siteName = settings?.site_name?.trim() || landingBrand.siteName
  const contactHref = settings?.whatsapp_url?.trim() || '/#kontak'

  return (
    <div className="landing-root min-h-screen bg-[var(--lp-surface)] antialiased">
      <LandingNav siteName={siteName} contactHref={contactHref} />

      <main className="pt-28 pb-12 sm:pt-32 sm:pb-16">
        <LandingContainer className="max-w-4xl">
          {announcementQuery.isLoading ? (
            <p className="rounded-3xl bg-[var(--lp-white)] p-8 text-sm text-[var(--lp-muted)] shadow-sm">
              Memuat pengumuman…
            </p>
          ) : announcementQuery.isError || !announcementQuery.data ? (
            <div className="rounded-3xl bg-[var(--lp-white)] p-8 shadow-sm">
              <h1 className="text-2xl font-bold text-[var(--lp-primary)]">
                Pengumuman tidak ditemukan
              </h1>
              <p className="mt-2 text-sm text-[var(--lp-muted)]">
                Pengumuman mungkin tidak tersedia atau belum diterbitkan untuk publik.
              </p>
            </div>
          ) : (
            <div>
              <Link
                to="/#program"
                className="mb-5 inline-flex text-sm font-semibold text-[var(--lp-green)] hover:text-[var(--lp-primary)]"
              >
                ← Kembali ke pengumuman
              </Link>
              <article className="overflow-hidden rounded-3xl bg-[var(--lp-white)] shadow-sm">
                {announcementQuery.data.thumbnail_url ? (
                  <img
                    src={announcementQuery.data.thumbnail_url}
                    alt=""
                    className="max-h-[28rem] w-full object-cover"
                  />
                ) : null}
                <div className="p-6 sm:p-10">
                  <p className="text-xs font-bold uppercase tracking-wider text-[var(--lp-green)]">
                    {announcementQuery.data.published_at
                      ? formatDateTime(announcementQuery.data.published_at)
                      : 'Pengumuman publik'}
                  </p>
                  <h1 className="mt-3 text-3xl font-bold leading-tight text-[var(--lp-primary)] sm:text-4xl">
                    {announcementQuery.data.title}
                  </h1>
                  <div className="mt-8 whitespace-pre-wrap text-base leading-8 text-[var(--lp-muted)]">
                    {announcementQuery.data.body}
                  </div>
                </div>
              </article>
            </div>
          )}
        </LandingContainer>
      </main>

      <LandingFooter settings={settings} />
    </div>
  )
}
