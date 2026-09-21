import { AnimatePresence, motion, useReducedMotion } from 'framer-motion'
import { useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { cn } from '../../../lib/utils'
import type { AuthContextValue } from '../../auth/auth-context'
import { useAuth } from '../../auth/useAuth'
import { landingAssets, landingBrand, landingNavLinks } from '../data'
import { Icon, LandingContainer } from './ui'

type LandingNavProps = {
  siteName: string
  contactHref: string
}

export function LandingNav({ siteName, contactHref }: LandingNavProps) {
  const location = useLocation()
  const isLandingPage = location.pathname === '/'
  const auth: AuthContextValue = useAuth()
  const { isAuthenticated } = auth
  const [scrolled, setScrolled] = useState(false)
  const [open, setOpen] = useState(false)
  const [active, setActive] = useState('#beranda')
  const reduce = useReducedMotion()
  const authHref = isAuthenticated ? '/dashboard' : '/login'
  const authLabel = isAuthenticated ? 'Dashboard' : 'Masuk'

  useEffect(() => {
    const onScroll = () => {
      setScrolled(window.scrollY > 16)
      const sections = landingNavLinks.map((l) => l.href.slice(1))
      for (const id of [...sections].reverse()) {
        const el = document.getElementById(id)
        if (el && el.getBoundingClientRect().top <= 120) {
          setActive(`#${id}`)
          break
        }
      }
    }
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  useEffect(() => {
    if (!open) {
      return
    }
    const prev = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      document.body.style.overflow = prev
    }
  }, [open])

  return (
    <>
      <motion.header
        className="fixed inset-x-0 top-0 z-50"
        initial={reduce ? false : { y: -24, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        transition={{ duration: 0.55, ease: [0.22, 1, 0.36, 1] }}
      >
        <div
          className={cn(
            'transition-[padding,background] duration-300',
            scrolled ? 'bg-[color-mix(in_srgb,var(--lp-surface)_72%,transparent)] px-3 pt-3 backdrop-blur-xl sm:px-4' : 'px-0 pt-0',
          )}
        >
          <LandingContainer
            className={cn(
              'flex items-center justify-between gap-3 transition-all duration-300',
              scrolled
                ? 'h-[4.25rem] rounded-2xl border border-black/[0.06] bg-[color-mix(in_srgb,var(--lp-white)_88%,transparent)] px-4 shadow-[0_10px_40px_-12px_rgba(15,37,39,0.18)] sm:px-5'
                : 'h-20 border border-transparent bg-transparent px-0',
            )}
          >
            <a href={isLandingPage ? '#beranda' : '/'} className="group flex min-w-0 items-center gap-3">
              <span className="relative flex h-11 w-11 shrink-0 items-center justify-center overflow-hidden">
                <img
                  src={landingAssets.logo}
                  alt=""
                  className="h-8 w-8 object-cover transition-transform duration-300 group-hover:scale-105"
                />
              </span>
              <span className="min-w-0">
                <span className="block truncate text-[15px] font-bold tracking-tight text-[var(--lp-primary)] sm:text-base">
                  {siteName}
                </span>
                <span className="hidden truncate text-[11px] font-medium text-[var(--lp-muted)] sm:block">
                  {landingBrand.tagline}
                </span>
              </span>
            </a>

            <nav
              className="relative hidden items-center rounded-full bg-[var(--lp-surface-low)]/80 p-1 lg:flex"
              aria-label="Navigasi utama"
            >
              {landingNavLinks.map((link) => {
                const isActive = isLandingPage && active === link.href
                return (
                  <a
                    key={link.href}
                    href={isLandingPage ? link.href : `/${link.href}`}
                    className={cn(
                      'relative z-10 rounded-full px-3.5 py-2 text-[13px] font-semibold transition-colors xl:px-4',
                      isActive ? 'text-[var(--lp-primary)]' : 'text-[var(--lp-muted)] hover:text-[var(--lp-ink)]',
                    )}
                  >
                    {isActive ? (
                      <motion.span
                        layoutId={reduce ? undefined : 'nav-active-pill'}
                        className="absolute inset-0 -z-10 rounded-full bg-[var(--lp-white)] shadow-[0_1px_4px_rgba(15,37,39,0.08)]"
                        transition={{ type: 'spring', stiffness: 420, damping: 34 }}
                      />
                    ) : null}
                    {link.label}
                  </a>
                )
              })}
            </nav>

            <div className="flex items-center gap-2 sm:gap-2.5">
              <Link
                to={authHref}
                className="hidden items-center gap-2 rounded-full px-3 py-2 text-[13px] font-semibold text-[var(--lp-muted)] transition-colors hover:bg-[var(--lp-surface-low)] hover:text-[var(--lp-primary)] md:inline-flex"
              >
                <span className="flex h-7 w-7 items-center justify-center rounded-full bg-[var(--lp-primary)] text-[var(--lp-on-primary)]">
                  <Icon name="person" className="text-[16px]" />
                </span>
                {authLabel}
              </Link>

              <motion.a
                href={contactHref}
                className="inline-flex items-center gap-1.5 rounded-full bg-[var(--lp-gold)] px-4 py-2.5 text-[13px] font-bold text-[var(--lp-on-gold)] shadow-[0_8px_20px_-6px_rgba(210,248,67,0.55)] sm:gap-2 sm:px-5"
                whileHover={reduce ? undefined : { y: -1, scale: 1.02 }}
                whileTap={reduce ? undefined : { scale: 0.98 }}
              >
                <span className="hidden sm:inline">Hubungi Kami</span>
                <span className="sm:hidden">Kontak</span>
                <Icon name="arrow_forward" className="text-[17px]" />
              </motion.a>

              <motion.button
                type="button"
                className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-black/5 bg-[var(--lp-white)] text-[var(--lp-primary)] shadow-sm lg:hidden"
                aria-expanded={open}
                aria-controls="landing-mobile-nav"
                aria-label={open ? 'Tutup menu' : 'Buka menu'}
                onClick={() => setOpen((v) => !v)}
                whileTap={reduce ? undefined : { scale: 0.94 }}
              >
                <Icon name={open ? 'close' : 'menu'} className="text-[22px]" />
              </motion.button>
            </div>
          </LandingContainer>
        </div>
      </motion.header>

      <AnimatePresence>
        {open ? (
          <>
            <motion.button
              type="button"
              aria-label="Tutup menu"
              className="fixed inset-0 z-40 bg-[var(--lp-primary)]/35 backdrop-blur-[2px] lg:hidden"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={() => setOpen(false)}
            />
            <motion.div
              id="landing-mobile-nav"
              className="fixed inset-x-3 top-[4.75rem] z-50 overflow-hidden rounded-3xl border border-black/5 bg-[var(--lp-white)] shadow-[0_24px_60px_-20px_rgba(15,37,39,0.35)] lg:hidden"
              initial={reduce ? false : { opacity: 0, y: -12, scale: 0.98 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={reduce ? undefined : { opacity: 0, y: -8, scale: 0.98 }}
              transition={{ duration: 0.28, ease: [0.22, 1, 0.36, 1] }}
            >
              <div className="border-b border-black/5 px-5 py-4">
                <p className="text-xs font-bold uppercase tracking-[0.14em] text-[var(--lp-green)]">Menu</p>
                <p className="mt-1 text-sm text-[var(--lp-muted)]">Jelajahi portal {siteName}</p>
              </div>
              <nav className="flex flex-col gap-1 p-3" aria-label="Navigasi mobile">
                {landingNavLinks.map((link, index) => {
                  const isActive = isLandingPage && active === link.href
                  return (
                    <motion.a
                      key={link.href}
                      href={isLandingPage ? link.href : `/${link.href}`}
                      className={cn(
                        'flex items-center justify-between rounded-2xl px-4 py-3.5 text-[15px] font-semibold transition-colors',
                        isActive
                          ? 'bg-[var(--lp-surface-low)] text-[var(--lp-primary)]'
                          : 'text-[var(--lp-ink)] hover:bg-[var(--lp-surface-low)]/70',
                      )}
                      initial={reduce ? false : { opacity: 0, x: -8 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: 0.04 + index * 0.03 }}
                      onClick={() => setOpen(false)}
                    >
                      {link.label}
                      {isActive ? (
                        <span className="h-2 w-2 rounded-full bg-[var(--lp-gold)]" />
                      ) : (
                        <Icon name="chevron_right" className="text-[18px] text-[var(--lp-muted)]" />
                      )}
                    </motion.a>
                  )
                })}
              </nav>
              <div className="grid gap-2 border-t border-black/5 p-4 sm:grid-cols-2">
                <Link
                  to={authHref}
                  className="inline-flex items-center justify-center gap-2 rounded-full border border-black/8 bg-[var(--lp-surface-low)] px-4 py-3 text-sm font-semibold text-[var(--lp-primary)]"
                  onClick={() => setOpen(false)}
                >
                  <Icon name="person" className="text-[18px]" />
                  {authLabel}
                </Link>
                <a
                  href={contactHref}
                  className="inline-flex items-center justify-center gap-2 rounded-full bg-[var(--lp-gold)] px-4 py-3 text-sm font-bold text-[var(--lp-on-gold)]"
                  onClick={() => setOpen(false)}
                >
                  Hubungi Kami
                  <Icon name="arrow_forward" className="text-[18px]" />
                </a>
              </div>
            </motion.div>
          </>
        ) : null}
      </AnimatePresence>
    </>
  )
}
