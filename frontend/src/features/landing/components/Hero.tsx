import { motion, useReducedMotion } from 'framer-motion'
import { cn } from '../../../lib/utils'
import { landingAssets, landingHero } from '../data'
import { Icon, LandingContainer, SafeImage } from './ui'

type HeroSectionProps = {
  siteName: string
  tagline: string
  announcementCount: number
  officialCount: number
  galleryCover?: string | null
}

export function HeroSection({
  siteName,
  tagline,
  announcementCount,
  officialCount,
  galleryCover,
}: HeroSectionProps) {
  const reduce = useReducedMotion()
  const coverSrc = galleryCover || landingAssets.heroCover

  return (
    <section id="beranda" className="scroll-mt-20 pt-6 pb-10">
      <LandingContainer>
        <div className="grid grid-cols-1 items-stretch gap-6 lg:grid-cols-12">
          <motion.div
            className="relative flex flex-col justify-between overflow-hidden rounded-3xl bg-[var(--lp-primary-container)] p-8 text-[var(--lp-on-primary)] shadow-xl sm:p-12 lg:col-span-7"
            initial={reduce ? false : { opacity: 0, y: 28 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.7, ease: [0.22, 1, 0.36, 1] }}
          >
            <div className="pointer-events-none absolute -right-20 -top-20 h-80 w-80 rounded-full bg-[var(--lp-gold)]/10 blur-3xl" />
            <div className="pointer-events-none absolute -bottom-16 -left-16 h-60 w-60 rounded-full bg-[var(--lp-green-light)]/25 blur-2xl" />

            <div className="relative z-10 space-y-6">
              <motion.div
                className="inline-flex items-center gap-2 rounded-full bg-white/10 px-4 py-1.5 backdrop-blur-md"
                initial={reduce ? false : { opacity: 0, x: -12 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.15, duration: 0.5 }}
              >
                <span className="h-2 w-2 animate-pulse rounded-full bg-[var(--lp-gold)]" />
                <span className="text-[11px] font-bold uppercase tracking-[0.08em] text-[var(--lp-gold)]">
                  {landingHero.eyebrow}
                </span>
              </motion.div>

              <h1 className="text-[clamp(2rem,4.5vw,3.5rem)] font-bold leading-[1.1] tracking-tight text-[var(--lp-white)]">
                {landingHero.headlineBefore}{' '}
                <span className="text-[var(--lp-gold)] underline decoration-[var(--lp-gold)]/30 decoration-wavy decoration-2">
                  {landingHero.headlineHighlight}
                </span>{' '}
                {landingHero.headlineAfter}
              </h1>

              <p className="max-w-xl text-base leading-relaxed text-[var(--lp-on-primary-muted)] sm:text-lg">
                {tagline ||
                  `Wadah transparansi administrasi, agenda bersama, dan informasi publik ${siteName} untuk warga yang terhubung.`}
              </p>

              <div className="flex flex-wrap items-center gap-3 pt-1">
                <motion.a
                  href={landingHero.primaryCta.href}
                  className="inline-flex items-center gap-2.5 rounded-full bg-[var(--lp-gold)] px-7 py-3.5 text-sm font-semibold text-[var(--lp-primary)] shadow-[0_8px_24px_-4px_rgba(210,248,67,0.4)] hover:bg-[var(--lp-gold)]/80"
                  whileHover={reduce ? undefined : { x: 2 }}
                  whileTap={reduce ? undefined : { scale: 0.98 }}
                >
                  <span className="text-[var(--lp-primary)]">{landingHero.primaryCta.label}</span>
                  <Icon name="arrow_forward" className="text-[18px] text-[var(--lp-primary)]" />
                </motion.a>
                <a
                  href={landingHero.secondaryCta.href}
                  className="inline-flex items-center gap-2 rounded-full px-6 py-3.5 text-sm font-semibold text-[var(--lp-white)] transition-colors hover:bg-white/10"
                >
                  <Icon name={landingHero.secondaryCta.icon} className="text-[18px] text-[var(--lp-gold)]" />
                  {landingHero.secondaryCta.label}
                </a>
              </div>
            </div>

            <div className="relative z-10 mt-10 grid grid-cols-1 gap-4 pt-8 sm:grid-cols-2">
              <motion.div
                className="flex items-center justify-between rounded-2xl bg-[var(--lp-white)] p-4 text-[var(--lp-ink)] shadow-md sm:p-5"
                initial={reduce ? false : { opacity: 0, y: 16 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.35, duration: 0.55 }}
              >
                <div className="space-y-1">
                  <span className="inline-block rounded-full bg-[var(--lp-surface-low)] px-2.5 py-0.5 text-[11px] font-bold text-[var(--lp-primary)]">
                    {landingHero.statPublic.badge}
                  </span>
                  <div className="text-2xl font-extrabold tracking-tight text-[var(--lp-primary)]">
                    {announcementCount > 0 ? `${announcementCount}+` : '—'}
                  </div>
                  <p className="text-xs text-[var(--lp-muted)]">{landingHero.statPublic.caption}</p>
                </div>
                <div className="flex h-11 w-11 items-center justify-center rounded-full bg-[var(--lp-surface-high)] text-[var(--lp-primary)]">
                  <Icon name={landingHero.statPublic.icon} className="text-[24px]" />
                </div>
              </motion.div>

              <motion.div
                className="flex items-center justify-between rounded-2xl bg-[var(--lp-white)] p-4 text-[var(--lp-ink)] shadow-md sm:p-5"
                initial={reduce ? false : { opacity: 0, y: 16 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.45, duration: 0.55 }}
              >
                <div className="space-y-1">
                  <div className="mb-1 flex items-center -space-x-2">
                    {landingHero.statTeam.avatars.map((letter, i) => (
                      <span
                        key={letter}
                        className={cn(
                          'flex h-7 w-7 items-center justify-center rounded-full text-[11px] font-bold',
                          i === 0 && 'bg-[var(--lp-gold)] text-[var(--lp-on-gold)]',
                          i === 1 && 'bg-[var(--lp-surface-high)] text-[var(--lp-primary)]',
                          i === 2 && 'bg-[var(--lp-green)] text-white',
                        )}
                      >
                        {letter}
                      </span>
                    ))}
                  </div>
                  <div className="text-2xl font-extrabold tracking-tight text-[var(--lp-primary)]">
                    {officialCount > 0 ? `${officialCount}` : '—'}
                  </div>
                  <p className="text-xs text-[var(--lp-muted)]">{landingHero.statTeam.caption}</p>
                </div>
                <span className="self-start rounded-full bg-[var(--lp-gold)]/40 px-2.5 py-1 text-[11px] font-bold text-[var(--lp-on-gold)]">
                  {landingHero.statTeam.badge}
                </span>
              </motion.div>
            </div>
          </motion.div>

          <motion.div
            className="group relative min-h-[420px] overflow-hidden rounded-3xl shadow-xl lg:col-span-5 lg:min-h-[580px]"
            initial={reduce ? false : { opacity: 0, scale: 0.96 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.75, ease: [0.22, 1, 0.36, 1], delay: 0.12 }}
          >
            <SafeImage
              src={coverSrc}
              alt={`Suasana ${siteName}`}
              className="absolute inset-0 h-full w-full object-cover transition-transform duration-700 group-hover:scale-105"
              fallback={
                <div className="absolute inset-0 flex items-center justify-center bg-[var(--lp-primary)]">
                  <img
                    src={landingAssets.logo}
                    alt=""
                    className="w-[55%] max-w-[260px] object-contain opacity-95"
                  />
                </div>
              }
            />
            <div className="absolute inset-0 bg-gradient-to-t from-[var(--lp-primary)]/80 via-transparent to-transparent" />
            <motion.div
              className="absolute inset-x-6 bottom-6 flex items-center justify-between rounded-2xl bg-[color-mix(in_srgb,var(--lp-white)_90%,transparent)] p-4 backdrop-blur-md"
              initial={reduce ? false : { y: 20, opacity: 0 }}
              animate={{ y: 0, opacity: 1 }}
              transition={{ delay: 0.5, duration: 0.55 }}
            >
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-[var(--lp-gold)] text-[var(--lp-on-gold)]">
                  <Icon name={landingHero.coverBadge.icon} className="text-[20px]" />
                </div>
                <div>
                  <div className="text-sm font-semibold text-[var(--lp-primary)]">Suasana {siteName}</div>
                  <p className="text-xs text-[var(--lp-muted)]">{landingHero.coverBadge.subtitle}</p>
                </div>
              </div>
              <Icon name="arrow_outward" className="text-[20px] text-[var(--lp-primary)]" />
            </motion.div>
          </motion.div>
        </div>
      </LandingContainer>
    </section>
  )
}
