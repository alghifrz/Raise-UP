import { AnimatePresence, motion, useReducedMotion } from 'framer-motion'
import { useMemo, useState, type ReactNode } from 'react'
import { formatDate, formatDateTime } from '../../../lib/utils'
import type { Activity } from '../../activities/types'
import type { AnnouncementSummary } from '../../announcements/types'
import type { GalleryItem } from '../../gallery/types'
import type { SiteSettings } from '../../site-settings/types'
import type { VillageOfficial, VillageProfile } from '../../village/types'
import {
  landingAbout,
  landingAgenda,
  landingAssets,
  landingBrand,
  landingFooter,
  landingGallery,
  landingLocation,
  landingMosaicTiles,
  landingPartners,
  landingPrograms,
  landingTransparency,
} from '../data'
import { useRevealVariants } from './useRevealVariants'
import { Icon, LandingContainer, SafeImage } from './ui'

const ease = [0.22, 1, 0.36, 1] as const

function RevealBlock({
  children,
  className,
  id,
}: {
  children: ReactNode
  className?: string
  id?: string
}) {
  const reduce = useReducedMotion()
  return (
    <motion.section
      id={id}
      className={className}
      initial={reduce ? false : { opacity: 0, y: 36 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, amount: 0.15 }}
      transition={{ duration: 0.65, ease }}
    >
      {children}
    </motion.section>
  )
}

export function PartnersSection() {
  const { container, item } = useRevealVariants()
  return (
    <RevealBlock className="py-8">
      <LandingContainer>
        <div className="space-y-6 rounded-3xl bg-[var(--lp-surface-low)] px-6 py-8 text-center lg:px-12">
          <p className="text-sm font-semibold uppercase tracking-wider text-[var(--lp-muted)]">
            {landingPartners.title}
          </p>
          <motion.div
            className="flex flex-wrap items-center justify-center gap-8 lg:gap-14"
            variants={container}
            initial="hidden"
            whileInView="visible"
            viewport={{ once: true }}
          >
            {landingPartners.items.map((partner) => (
              <motion.div
                key={partner.label}
                variants={item}
                className="flex items-center gap-2 text-lg font-semibold text-[var(--lp-primary)]"
              >
                <Icon name={partner.icon} className="text-[26px] text-[var(--lp-green)]" />
                <span>{partner.label}</span>
              </motion.div>
            ))}
          </motion.div>
        </div>
      </LandingContainer>
    </RevealBlock>
  )
}

export function AboutMosaicSection({
  profile,
  gallery,
}: {
  profile: VillageProfile | undefined
  gallery: GalleryItem[] | undefined
}) {
  const reduce = useReducedMotion()
  const tiles = useMemo(() => {
    const photos = gallery?.slice(0, 4) ?? []
    return landingMosaicTiles.map((tile, index) => ({
      label: tile.label,
      icon: tile.icon,
      image: photos[index]?.image_url ?? tile.image,
    }))
  }, [gallery])

  const blurb =
    profile?.vision?.trim() || profile?.history?.trim() || landingAbout.fallbackBlurb

  return (
    <RevealBlock id="tentang" className="scroll-mt-24 py-10">
      <LandingContainer>
        <div className="relative space-y-10 overflow-hidden rounded-3xl bg-[var(--lp-primary-container)] p-8 text-[var(--lp-on-primary)] shadow-xl sm:p-12">
          <div className="pointer-events-none absolute -right-32 bottom-0 h-96 w-96 rounded-full bg-[var(--lp-gold)]/5 blur-3xl" />

          <div className="relative z-10 flex flex-col justify-between gap-6 lg:flex-row lg:items-end">
            <div className="max-w-2xl space-y-3">
              <span className="text-[11px] font-bold uppercase tracking-wider text-[var(--lp-gold)]">
                {landingAbout.eyebrow}
              </span>
              <h2 className="text-[clamp(1.75rem,3vw,2.5rem)] font-bold leading-tight text-[var(--lp-white)]">
                {landingAbout.titleBefore}{' '}
                <span className="text-[var(--lp-gold)]">{landingAbout.titleHighlight}</span>
              </h2>
            </div>
            <div className="max-w-md space-y-4">
              <p className="whitespace-pre-wrap text-sm leading-relaxed text-[var(--lp-on-primary-muted)] line-clamp-5">
                {blurb}
              </p>
              <a
                href={landingAbout.cta.href}
                className="inline-flex items-center gap-2 rounded-full px-5 py-2.5 text-sm font-semibold text-[var(--lp-white)] transition-colors hover:bg-white/10"
              >
                {landingAbout.cta.label}
                <Icon name="arrow_forward" className="text-[16px] text-[var(--lp-gold)]" />
              </a>
            </div>
          </div>

          <div className="relative z-10 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
            {tiles.map((tile, index) => (
              <motion.div
                key={tile.label}
                className={`group relative h-80 overflow-hidden rounded-2xl shadow-md ${index % 2 === 1 ? 'lg:translate-y-4' : ''}`}
                initial={reduce ? false : { opacity: 0, y: 24 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: index * 0.08, duration: 0.55, ease }}
                whileHover={reduce ? undefined : { y: -4 }}
              >
                <SafeImage
                  src={tile.image}
                  alt={tile.label}
                  className="absolute inset-0 h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
                  fallback={
                    <div className="absolute inset-0 flex items-center justify-center bg-[var(--lp-primary)]">
                      <Icon name={tile.icon} className="text-6xl text-[var(--lp-gold)]/40" />
                    </div>
                  }
                />
                <div className="absolute inset-0 bg-gradient-to-t from-[var(--lp-primary)]/90 via-[var(--lp-primary)]/20 to-transparent" />
                <div className="absolute inset-x-4 bottom-4">
                  <span className="inline-block rounded-full bg-[color-mix(in_srgb,var(--lp-surface)_90%,transparent)] px-3 py-1.5 text-[11px] font-semibold text-[var(--lp-primary)] backdrop-blur-md">
                    {tile.label}
                  </span>
                </div>
              </motion.div>
            ))}
          </div>
        </div>
      </LandingContainer>
    </RevealBlock>
  )
}

export function ProgramsSection({ items }: { items: AnnouncementSummary[] | undefined }) {
  const reduce = useReducedMotion()
  const [page, setPage] = useState(0)
  const pageSize = 3
  const list = items ?? []
  const totalPages = Math.max(1, Math.ceil(list.length / pageSize))
  const slice = list.slice(page * pageSize, page * pageSize + pageSize)

  function prev() {
    setPage((p) => (p - 1 + totalPages) % totalPages)
  }
  function next() {
    setPage((p) => (p + 1) % totalPages)
  }

  return (
    <RevealBlock id="program" className="scroll-mt-24 py-12">
      <LandingContainer>
        <div className="mb-8 flex flex-row items-center justify-between gap-4">
          <div>
            <span className="text-[11px] font-bold uppercase tracking-wider text-[var(--lp-green)]">
              {landingPrograms.eyebrow}
            </span>
            <h2 className="text-[clamp(1.75rem,3vw,2.5rem)] font-bold text-[var(--lp-ink)]">
              {landingPrograms.title}
            </h2>
          </div>
          <div className="flex items-center gap-2">
            <motion.button
              type="button"
              aria-label="Sebelumnya"
              className="flex h-11 w-11 items-center justify-center rounded-full bg-[var(--lp-surface-high)] text-[var(--lp-primary)] shadow-sm"
              onClick={prev}
              whileTap={reduce ? undefined : { scale: 0.92 }}
              disabled={list.length <= pageSize}
            >
              <Icon name="arrow_back" className="text-[20px]" />
            </motion.button>
            <motion.button
              type="button"
              aria-label="Berikutnya"
              className="flex h-11 w-11 items-center justify-center rounded-full bg-[var(--lp-primary)] text-[var(--lp-on-primary)] shadow-sm"
              onClick={next}
              whileTap={reduce ? undefined : { scale: 0.92 }}
              disabled={list.length <= pageSize}
            >
              <Icon name="arrow_forward" className="text-[20px]" />
            </motion.button>
          </div>
        </div>

        {!list.length ? (
          <p className="rounded-3xl bg-[var(--lp-white)] p-8 text-sm text-[var(--lp-muted)] shadow-sm">
            {landingPrograms.empty}
          </p>
        ) : (
          <AnimatePresence mode="wait">
            <motion.div
              key={page}
              className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3"
              initial={reduce ? false : { opacity: 0, x: 24 }}
              animate={{ opacity: 1, x: 0 }}
              exit={reduce ? undefined : { opacity: 0, x: -24 }}
              transition={{ duration: 0.35, ease }}
            >
              {slice.map((item, index) => (
                <motion.article
                  key={item.id}
                  className="flex flex-col justify-between rounded-3xl bg-[var(--lp-white)] p-5 shadow-sm transition-shadow hover:shadow-lg"
                  initial={reduce ? false : { opacity: 0, y: 16 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: index * 0.06, duration: 0.45 }}
                  whileHover={reduce ? undefined : { y: -4 }}
                >
                  <div className="space-y-4">
                    <div className="relative h-56 overflow-hidden rounded-2xl bg-[var(--lp-surface-low)]">
                      {item.thumbnail_url ? (
                        <img
                          src={item.thumbnail_url}
                          alt=""
                          className="h-full w-full object-cover"
                        />
                      ) : (
                        <div className="flex h-full items-center justify-center">
                          <Icon name="campaign" className="text-5xl text-[var(--lp-green)]/40" />
                        </div>
                      )}
                      <div className="absolute left-3 top-3 rounded-full bg-[var(--lp-gold)] px-3 py-1 text-[11px] font-bold text-[var(--lp-on-gold)] shadow-md">
                        {item.category || 'Info'}
                      </div>
                    </div>
                    <div className="space-y-2">
                      <span className="text-[11px] font-bold uppercase tracking-wider text-[var(--lp-green)]">
                        {item.published_at ? formatDate(item.published_at) : 'Terbit'}
                      </span>
                      <h3 className="text-lg font-bold text-[var(--lp-primary)] transition-colors hover:text-[var(--lp-green)]">
                        {item.title}
                      </h3>
                      <p className="line-clamp-2 text-xs leading-relaxed text-[var(--lp-muted)]">
                        {item.excerpt || item.body}
                      </p>
                    </div>
                  </div>
                  <div className="mt-6 flex items-center justify-between border-t border-[var(--lp-surface-low)] pt-4 text-xs text-[var(--lp-muted)]">
                    <span className="flex items-center gap-1">
                      <Icon name="public" className="text-[14px] text-[var(--lp-green)]" />
                      Publik
                    </span>
                    <a href={`#pengumuman-${item.id}`} className="font-semibold text-[var(--lp-primary)]">
                      Baca
                    </a>
                  </div>
                </motion.article>
              ))}
            </motion.div>
          </AnimatePresence>
        )}

        {list.length > 0 ? (
          <div className="mt-8 space-y-3" id="pengumuman-detail">
            {list.slice(0, 3).map((item) => (
              <details
                key={`detail-${item.id}`}
                id={`pengumuman-${item.id}`}
                className="group rounded-2xl bg-[var(--lp-white)] p-5 shadow-sm open:shadow-md"
              >
                <summary className="cursor-pointer list-none font-semibold text-[var(--lp-primary)] marker:content-none">
                  <span className="flex items-center justify-between gap-3">
                    {item.title}
                    <Icon name="expand_more" className="transition group-open:rotate-180" />
                  </span>
                </summary>
                <p className="mt-3 whitespace-pre-wrap text-sm leading-relaxed text-[var(--lp-muted)]">
                  {item.body || item.excerpt}
                </p>
              </details>
            ))}
          </div>
        ) : null}
      </LandingContainer>
    </RevealBlock>
  )
}

export function AgendaSection({ items }: { items: Activity[] | undefined }) {
  const reduce = useReducedMotion()
  return (
    <RevealBlock id="agenda" className="scroll-mt-24 py-8">
      <LandingContainer>
        <div className="mb-6">
          <span className="text-[11px] font-bold uppercase tracking-wider text-[var(--lp-green)]">
            {landingAgenda.eyebrow}
          </span>
          <h2 className="text-[clamp(1.75rem,3vw,2.5rem)] font-bold text-[var(--lp-ink)]">
            {landingAgenda.title}
          </h2>
        </div>
        {!items?.length ? (
          <p className="text-sm text-[var(--lp-muted)]">{landingAgenda.empty}</p>
        ) : (
          <ol className="relative space-y-0 border-l-2 border-[var(--lp-gold)]/50 pl-6">
            {items.map((activity, index) => (
              <motion.li
                key={activity.id}
                className="relative pb-8 last:pb-0"
                initial={reduce ? false : { opacity: 0, x: -12 }}
                whileInView={{ opacity: 1, x: 0 }}
                viewport={{ once: true }}
                transition={{ delay: index * 0.06, duration: 0.45 }}
              >
                <span className="absolute -left-[1.85rem] top-1.5 h-3 w-3 rounded-full bg-[var(--lp-gold)] shadow-[0_0_0_4px_rgba(210,248,67,0.22)]" />
                <time className="text-xs font-bold uppercase tracking-wider text-[var(--lp-green)]">
                  {formatDateTime(activity.date)}
                </time>
                <h3 className="mt-1 text-lg font-bold text-[var(--lp-primary)]">{activity.name}</h3>
                {activity.description ? (
                  <p className="mt-1 max-w-2xl text-sm text-[var(--lp-muted)]">{activity.description}</p>
                ) : null}
              </motion.li>
            ))}
          </ol>
        )}
      </LandingContainer>
    </RevealBlock>
  )
}

export function TransparencySection({ officials }: { officials: VillageOfficial[] | undefined }) {
  const reduce = useReducedMotion()

  return (
    <RevealBlock id="transparansi" className="scroll-mt-24 py-12">
      <LandingContainer>
        <div className="grid grid-cols-1 items-center gap-10 lg:grid-cols-12">
          <div className="space-y-6 lg:col-span-6">
            <span className="text-[11px] font-bold uppercase tracking-wider text-[var(--lp-green)]">
              {landingTransparency.eyebrow}
            </span>
            <h2 className="text-[clamp(1.75rem,3vw,2.5rem)] font-bold leading-tight text-[var(--lp-ink)]">
              {landingTransparency.title}
            </h2>
            <p className="text-base leading-relaxed text-[var(--lp-muted)]">{landingTransparency.body}</p>
            <div className="space-y-4 pt-2">
              {landingTransparency.features.map((feature, index) => (
                <motion.div
                  key={feature.title}
                  className="flex items-start gap-4"
                  initial={reduce ? false : { opacity: 0, x: -16 }}
                  whileInView={{ opacity: 1, x: 0 }}
                  viewport={{ once: true }}
                  transition={{ delay: index * 0.08, duration: 0.45 }}
                >
                  <div className="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-[var(--lp-primary)] text-[var(--lp-on-primary)]">
                    <Icon name={feature.icon} className="text-[20px]" />
                  </div>
                  <div>
                    <div className="text-lg font-bold text-[var(--lp-primary)]">{feature.title}</div>
                    <p className="text-xs leading-relaxed text-[var(--lp-muted)]">{feature.body}</p>
                  </div>
                </motion.div>
              ))}
            </div>
            <div className="pt-2">
              <motion.a
                href={landingTransparency.cta.href}
                className="inline-flex items-center gap-2 rounded-full bg-[var(--lp-gold)] px-7 py-3.5 text-sm font-semibold text-[var(--lp-on-gold)] shadow-md"
                whileHover={reduce ? undefined : { scale: 1.02 }}
                whileTap={reduce ? undefined : { scale: 0.98 }}
              >
                {landingTransparency.cta.label}
                <Icon name="arrow_forward" className="text-[18px]" />
              </motion.a>
            </div>
          </div>

          <div className="relative flex items-center justify-center overflow-hidden rounded-3xl bg-[var(--lp-surface-low)] p-6 sm:p-10 lg:col-span-6">
            <svg
              className="pointer-events-none absolute h-full w-full opacity-10"
              fill="none"
              viewBox="0 0 400 400"
              aria-hidden
            >
              <circle cx="200" cy="200" r="160" stroke="currentColor" strokeWidth="2" />
              <circle cx="200" cy="200" r="100" stroke="currentColor" strokeWidth="2" />
            </svg>
            <motion.div
              className="relative z-10 w-full max-w-sm rounded-[2.5rem] bg-[var(--lp-primary-container)] p-4 text-[var(--lp-on-primary)] shadow-2xl"
              initial={reduce ? false : { opacity: 0, y: 30, rotate: -2 }}
              whileInView={{ opacity: 1, y: 0, rotate: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.7, ease }}
            >
              <div className="flex items-center justify-between px-3 py-2 text-xs text-[var(--lp-on-primary-muted)]">
                <span>09:41</span>
                <div className="flex items-center gap-1.5">
                  <Icon name="signal_cellular_4_bar" className="text-[14px]" />
                  <Icon name="wifi" className="text-[14px]" />
                  <Icon name="battery_full" className="text-[14px]" />
                </div>
              </div>
              <div className="space-y-4 rounded-3xl bg-[var(--lp-white)] p-4 text-[var(--lp-ink)]">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="text-xs text-[var(--lp-muted)]">{landingTransparency.mockupGreeting}</div>
                    <div className="text-sm font-bold text-[var(--lp-primary)]">
                      {landingTransparency.mockupUser}
                    </div>
                  </div>
                  <div className="flex h-8 w-8 items-center justify-center rounded-full bg-[var(--lp-gold)] text-xs font-bold text-[var(--lp-on-gold)]">
                    {landingTransparency.mockupBadge}
                  </div>
                </div>
                <div className="space-y-2 rounded-2xl bg-[var(--lp-primary)] p-4 text-[var(--lp-on-primary)]">
                  <div className="flex items-center justify-between text-xs text-[var(--lp-on-primary-muted)]">
                    <span>Status Portal</span>
                    <span className="rounded-full bg-[var(--lp-gold)] px-2 py-0.5 text-[10px] font-bold text-[var(--lp-on-gold)]">
                      Live
                    </span>
                  </div>
                  <div className="text-xl font-extrabold text-[var(--lp-gold)]">Transparansi Aktif</div>
                  <div className="flex items-center justify-between pt-1 text-[11px] text-[var(--lp-on-primary-muted)]">
                    <span>Pengurus: {officials?.length ?? 0}</span>
                    <span>Publik terbuka</span>
                  </div>
                </div>
                <div className="space-y-2 pt-1">
                  <div className="flex items-center justify-between text-xs font-bold text-[var(--lp-primary)]">
                    <span>Aktivitas Terbaru</span>
                    <span className="text-[11px] text-[var(--lp-green)]">Lihat Semua</span>
                  </div>
                  <div className="flex items-center gap-3 rounded-xl bg-[var(--lp-surface-low)] p-2">
                    <span className="flex h-8 w-8 items-center justify-center rounded-full bg-[var(--lp-gold)]/30 text-[var(--lp-green)]">
                      <Icon name="campaign" className="text-[16px]" />
                    </span>
                    <div className="min-w-0 flex-1">
                      <div className="truncate text-xs font-bold text-[var(--lp-primary)]">Pengumuman publik</div>
                      <div className="text-[10px] text-[var(--lp-muted)]">Portal warga • baru saja</div>
                    </div>
                  </div>
                  <div className="flex items-center gap-3 rounded-xl bg-[var(--lp-surface-low)] p-2">
                    <span className="flex h-8 w-8 items-center justify-center rounded-full bg-[var(--lp-surface-high)] text-[var(--lp-primary)]">
                      <Icon name="event" className="text-[16px]" />
                    </span>
                    <div className="min-w-0 flex-1">
                      <div className="truncate text-xs font-bold text-[var(--lp-primary)]">Agenda kegiatan</div>
                      <div className="text-[10px] text-[var(--lp-muted)]">Jadwal bersama</div>
                    </div>
                  </div>
                </div>
              </div>
            </motion.div>
          </div>
        </div>
      </LandingContainer>
    </RevealBlock>
  )
}

export function GalleryStripSection({ items }: { items: GalleryItem[] | undefined }) {
  const reduce = useReducedMotion()
  const photos = items ?? []

  return (
    <RevealBlock id="galeri" className="scroll-mt-24 py-12">
      <LandingContainer>
        <div className="mb-8 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
          <div>
            <span className="text-[11px] font-bold uppercase tracking-wider text-[var(--lp-green)]">
              {landingGallery.eyebrow}
            </span>
            <h2 className="text-[clamp(1.75rem,3vw,2.5rem)] font-bold text-[var(--lp-ink)]">
              {landingGallery.title}
            </h2>
            <p className="mt-2 max-w-xl text-sm text-[var(--lp-muted)]">{landingGallery.description}</p>
          </div>
          <a
            href={landingGallery.cta.href}
            className="inline-flex items-center gap-2 self-start rounded-full bg-[var(--lp-surface-high)] px-5 py-2.5 text-sm font-semibold text-[var(--lp-primary)] transition-colors hover:bg-[var(--lp-surface-highest)] sm:self-auto"
          >
            {landingGallery.cta.label}
            <Icon name="arrow_forward" className="text-[16px] text-[var(--lp-green)]" />
          </a>
        </div>

        {photos.length === 0 ? (
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
            {landingGallery.placeholders.map((placeholder, index) => (
              <motion.div
                key={placeholder.label}
                className="relative h-64 overflow-hidden rounded-3xl border border-dashed border-[var(--lp-green)]/25 bg-[var(--lp-white)] shadow-sm"
                initial={reduce ? false : { opacity: 0, y: 16 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: index * 0.08, duration: 0.45 }}
              >
                <SafeImage
                  src={placeholder.image}
                  alt={placeholder.label}
                  className="h-full w-full object-cover"
                  fallback={
                    <div className="flex h-full flex-col items-center justify-center">
                      <div className="mb-3 flex h-14 w-14 items-center justify-center rounded-full bg-[var(--lp-surface-low)] text-[var(--lp-green)]">
                        <Icon name={placeholder.icon} className="text-[28px]" />
                      </div>
                      <p className="text-sm font-semibold text-[var(--lp-primary)]">{placeholder.label}</p>
                      <p className="mt-1 text-xs text-[var(--lp-muted)]">{landingGallery.emptyCaption}</p>
                    </div>
                  }
                />
              </motion.div>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
            {photos.map((photo, index) => (
              <motion.figure
                key={photo.id}
                className="group relative overflow-hidden rounded-3xl bg-[var(--lp-white)] shadow-sm"
                initial={reduce ? false : { opacity: 0, y: 16 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: index * 0.04, duration: 0.45 }}
                whileHover={reduce ? undefined : { y: -4 }}
              >
                <div className="relative h-64 overflow-hidden">
                  <img
                    src={photo.image_url}
                    alt={photo.caption || 'Dokumentasi kegiatan'}
                    className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
                  />
                  <div className="absolute inset-0 bg-gradient-to-t from-[var(--lp-primary)]/70 via-transparent to-transparent opacity-80" />
                  <figcaption className="absolute inset-x-4 bottom-4 text-sm font-semibold text-white">
                    {photo.caption || `Dokumentasi ${landingBrand.siteName}`}
                  </figcaption>
                </div>
              </motion.figure>
            ))}
          </div>
        )}
      </LandingContainer>
    </RevealBlock>
  )
}

function toGoogleMapsEmbed(query: string): string {
  const q = encodeURIComponent(query.trim() || landingBrand.mapQuery)
  return `https://maps.google.com/maps?q=${q}&hl=id&z=16&output=embed`
}

function resolveMapEmbedUrl(settings: SiteSettings | undefined): string {
  const embed = settings?.embed_url?.trim()
  if (embed) {
    return embed
  }

  const mapsUrl = settings?.maps_url?.trim()
  if (mapsUrl) {
    if (mapsUrl.includes('/maps/embed') || mapsUrl.includes('output=embed')) {
      return mapsUrl
    }
    try {
      const parsed = new URL(mapsUrl)
      const q =
        parsed.searchParams.get('q') ||
        parsed.searchParams.get('query') ||
        decodedPlaceFromMapsPath(parsed.pathname) ||
        settings?.address?.trim() ||
        landingBrand.mapQuery
      return toGoogleMapsEmbed(q)
    } catch {
      return toGoogleMapsEmbed(settings?.address?.trim() || landingBrand.mapQuery)
    }
  }

  return toGoogleMapsEmbed(settings?.address?.trim() || landingBrand.mapQuery)
}

function decodedPlaceFromMapsPath(pathname: string): string | null {
  const match = pathname.match(/\/place\/([^/]+)/)
  if (!match?.[1]) {
    return null
  }
  try {
    return decodeURIComponent(match[1].replace(/\+/g, ' '))
  } catch {
    return match[1]
  }
}

export function LocationBanner({ settings }: { settings: SiteSettings | undefined }) {
  const reduce = useReducedMotion()
  const mapTitle = settings?.map_title?.trim() || landingLocation.eyebrowFallback
  const mapDescription = settings?.map_description?.trim() || landingLocation.description
  const address = settings?.address?.trim() || landingBrand.fullAddress
  const whatsapp = settings?.whatsapp_url?.trim()
  const mapsUrl =
    settings?.maps_url?.trim() ||
    `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(address)}`
  const embedSrc = resolveMapEmbedUrl(settings)

  return (
    <RevealBlock id="kontak" className="scroll-mt-24 pt-6 pb-16">
      <LandingContainer>
        <div className="relative overflow-hidden rounded-3xl bg-[var(--lp-primary-container)] p-8 text-[var(--lp-on-primary)] shadow-xl sm:p-12">
          <svg
            className="pointer-events-none absolute right-0 top-0 bottom-0 h-full w-full opacity-20 lg:w-1/2"
            fill="none"
            viewBox="0 0 500 300"
            aria-hidden
          >
            <pattern id="grid-dots" width="20" height="20" patternUnits="userSpaceOnUse">
              <circle cx="2" cy="2" r="1.5" fill="#d2f843" />
            </pattern>
            <rect width="500" height="300" fill="url(#grid-dots)" />
            <path
              d="M 50 250 Q 200 100 450 180"
              fill="none"
              stroke="#d2f843"
              strokeDasharray="6 6"
              strokeWidth="2"
            />
          </svg>

          <div className="relative z-10 grid grid-cols-1 items-center gap-8 lg:grid-cols-12">
            <div className="space-y-5 lg:col-span-7">
              <div className="inline-flex items-center gap-2 rounded-full bg-white/10 px-3.5 py-1">
                <Icon name="pin_drop" className="text-[16px] text-[var(--lp-gold)]" />
                <span className="text-[11px] font-bold uppercase tracking-wider text-[var(--lp-gold)]">
                  {mapTitle}
                </span>
              </div>
              <h2 className="text-[clamp(1.75rem,3vw,2.5rem)] font-bold leading-tight text-[var(--lp-white)]">
                {landingLocation.headline}
              </h2>
              <p className="max-w-lg text-sm leading-relaxed text-[var(--lp-on-primary-muted)]">
                {mapDescription}
              </p>
              <p className="flex items-start gap-2 text-sm text-[var(--lp-on-primary-muted)]">
                <Icon name="home_pin" className="mt-0.5 shrink-0 text-[var(--lp-gold)]" />
                <span>{address}</span>
              </p>
              {settings?.phone ? (
                <p className="text-sm text-[var(--lp-on-primary-muted)]">
                  Telepon:{' '}
                  <a
                    href={`tel:${settings.phone.replace(/\s/g, '')}`}
                    className="font-semibold text-[var(--lp-gold)]"
                  >
                    {settings.phone}
                  </a>
                </p>
              ) : null}
              <div className="flex flex-wrap gap-4 pt-2">
                <motion.a
                  href={mapsUrl}
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex items-center gap-2 rounded-full bg-[var(--lp-gold)] px-7 py-3.5 text-sm font-bold text-[var(--lp-on-gold)] shadow-lg"
                  whileHover={reduce ? undefined : { scale: 1.02 }}
                >
                  {landingLocation.mapsCta}
                  <Icon name="arrow_forward" className="text-[18px]" />
                </motion.a>
                {whatsapp ? (
                  <a
                    href={whatsapp}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center gap-2 rounded-full bg-white/10 px-6 py-3.5 text-sm font-semibold text-[var(--lp-white)] transition-colors hover:bg-white/20"
                  >
                    <Icon name="chat" className="text-[18px]" />
                    {landingLocation.whatsappCta}
                  </a>
                ) : (
                  <a
                    href="#kontak"
                    className="inline-flex items-center gap-2 rounded-full bg-white/10 px-6 py-3.5 text-sm font-semibold text-[var(--lp-white)] transition-colors hover:bg-white/20"
                  >
                    <Icon name="chat" className="text-[18px]" />
                    {landingLocation.contactCta}
                  </a>
                )}
              </div>
            </div>

            <div className="flex flex-col items-start justify-center gap-3 lg:col-span-5 lg:items-end">
              {landingLocation.badges.map((badge, index) => (
                <motion.div
                  key={badge.label}
                  className={`flex items-center gap-2.5 rounded-full bg-[var(--lp-white)] px-4 py-2.5 text-[var(--lp-primary)] shadow-lg ${index === 1 ? 'lg:-mr-4' : ''}`}
                  initial={reduce ? false : { opacity: 0, x: 20 }}
                  whileInView={{ opacity: 1, x: 0 }}
                  viewport={{ once: true }}
                  transition={{ delay: 0.1 + index * 0.08, duration: 0.45 }}
                >
                  <Icon name={badge.icon} className="text-[18px] text-[var(--lp-green)]" />
                  <span className="text-sm font-bold">{badge.label}</span>
                </motion.div>
              ))}
            </div>
          </div>

          <motion.div
            className="relative z-10 mt-8 overflow-hidden rounded-2xl border border-white/10 bg-black/20 shadow-lg"
            initial={reduce ? false : { opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.55, ease: [0.22, 1, 0.36, 1] }}
          >
            <iframe
              title={mapTitle}
              src={embedSrc}
              className="h-72 w-full border-0 sm:h-96"
              loading="lazy"
              referrerPolicy="no-referrer-when-downgrade"
              allowFullScreen
            />
          </motion.div>
        </div>
      </LandingContainer>
    </RevealBlock>
  )
}

export function LandingFooter({ settings }: { settings: SiteSettings | undefined }) {
  const year = new Date().getFullYear()
  const name = settings?.site_name?.trim() || landingBrand.siteName
  const blurb = settings?.footer_blurb?.trim() || landingBrand.footerBlurb

  const socialHrefs = {
    whatsapp: settings?.whatsapp_url || '#kontak',
    phone: settings?.phone ? `tel:${settings.phone.replace(/\s/g, '')}` : '#kontak',
  }

  return (
    <footer className="w-full bg-[var(--lp-primary-container)] text-[var(--lp-on-primary)]">
      <LandingContainer className="py-10">
        <div className="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-12">
          <div className="space-y-4 lg:col-span-4">
            <div className="flex items-center gap-3">
              <img src={landingAssets.logo} alt="" className="h-8 w-auto object-contain" />
              <span className="text-lg font-semibold text-[var(--lp-white)]">{name}</span>
            </div>
            <p className="max-w-sm text-sm leading-relaxed text-[var(--lp-on-primary-muted)]">{blurb}</p>
            <div className="flex items-center gap-3 pt-2">
              {landingFooter.social.map((social) => {
                const href = social.href ?? (social.hrefKey ? socialHrefs[social.hrefKey] : '#kontak')
                return (
                  <a
                    key={social.icon}
                    href={href}
                    target={href.startsWith('http') ? '_blank' : undefined}
                    rel={href.startsWith('http') ? 'noreferrer' : undefined}
                    className="flex h-10 w-10 items-center justify-center rounded-full bg-[var(--lp-primary)] text-[var(--lp-white)] transition-colors hover:bg-[var(--lp-gold)] hover:text-[var(--lp-on-gold)]"
                  >
                    <Icon name={social.icon} className="text-[20px]" />
                  </a>
                )
              })}
            </div>
          </div>

          <div className="lg:col-span-2">
            <h4 className="mb-4 text-sm font-semibold uppercase tracking-wider text-[var(--lp-gold)]">
              Tautan Cepat
            </h4>
            <ul className="space-y-3 text-sm text-[var(--lp-on-primary-muted)]">
              {landingFooter.quickLinks.map((link) => (
                <li key={link.href}>
                  <a href={link.href} className="transition-colors hover:text-[var(--lp-white)]">
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          <div className="lg:col-span-2">
            <h4 className="mb-4 text-sm font-semibold uppercase tracking-wider text-[var(--lp-gold)]">
              Layanan
            </h4>
            <ul className="space-y-3 text-sm text-[var(--lp-on-primary-muted)]">
              {landingFooter.serviceLinks.map((link) => (
                <li key={link.href}>
                  <a href={link.href} className="hover:text-[var(--lp-white)]">
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          <div className="space-y-4 lg:col-span-4">
            <h4 className="text-sm font-semibold uppercase tracking-wider text-[var(--lp-gold)]">
              {landingFooter.newsletterTitle}
            </h4>
            <p className="text-xs leading-relaxed text-[var(--lp-on-primary-muted)]">
              {landingFooter.newsletterBody}
            </p>
            <a
              href={landingFooter.newsletterCta.href}
              className="inline-flex items-center gap-2 rounded-full bg-[var(--lp-gold)] px-5 py-3 text-sm font-semibold text-[var(--lp-on-gold)]"
            >
              {landingFooter.newsletterCta.label}
              <Icon name="send" className="text-[16px]" />
            </a>
          </div>
        </div>

        <div className="mt-10 flex flex-col items-center justify-between gap-4 border-t border-white/10 pt-6 text-xs text-[var(--lp-on-primary-muted)] md:flex-row">
          <p>
            © {year} {name}. {landingFooter.copyrightSuffix}
          </p>
          <div className="flex items-center gap-6">
            {landingFooter.values.map((value) => (
              <span key={value}>{value}</span>
            ))}
          </div>
        </div>
      </LandingContainer>
    </footer>
  )
}

