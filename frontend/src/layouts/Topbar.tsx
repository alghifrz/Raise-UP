import { Link } from 'react-router-dom'
import { useAuth } from '../features/auth/useAuth'
import { formatRoleLabel } from '../lib/utils'
import { landingBrand } from '../features/landing/data'

type TopbarProps = {
  title: string
  menuOpen: boolean
  onMenuClick: () => void
  onLogout: () => void
}

export function Topbar({ title, menuOpen, onMenuClick, onLogout }: TopbarProps) {
  const { user } = useAuth()
  const initial = (user?.name ?? 'A').trim().charAt(0).toUpperCase() || 'A'

  return (
    <header className="sticky top-0 z-20 shrink-0 border-b border-[var(--color-line)] bg-[var(--color-panel)]/95 backdrop-blur-md">
      <div className="flex h-[4.5rem] items-center justify-between gap-3 px-4 sm:px-6">
        <div className="flex min-w-0 items-center gap-3">
          <button
            type="button"
            className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-[var(--color-accent)] text-white shadow-[0_8px_18px_-10px_rgba(13,59,63,0.65)] transition hover:bg-[var(--color-accent-hover)] lg:hidden"
            onClick={onMenuClick}
            aria-label="Buka menu navigasi"
            aria-controls="admin-sidebar"
            aria-expanded={menuOpen}
          >
            <span className="material-symbols-outlined text-[22px]" aria-hidden>
              menu
            </span>
          </button>

          <div className="flex min-w-0 items-center gap-3">
            <span
              className="hidden h-9 w-1.5 shrink-0 rounded-full bg-[var(--color-secondary)] sm:block"
              aria-hidden
            />
            <div className="min-w-0">
              <p className="truncate text-lg font-bold tracking-tight text-[var(--color-ink)]">
                {title}
              </p>
              <p className="truncate text-[11px] font-medium tracking-wide text-[var(--color-muted)] uppercase">
                {landingBrand.siteName} · Admin
              </p>
            </div>
          </div>
        </div>

        <div className="flex shrink-0 items-center gap-2 sm:gap-2.5">
          <Link
            to="/"
            className="inline-flex h-10 items-center gap-1.5 rounded-2xl border border-[var(--color-line)] bg-[var(--color-surface)] px-3 text-xs font-semibold text-[var(--color-tertiary)] transition hover:border-[var(--color-tertiary)]/30 hover:bg-[var(--color-accent-soft)]"
            title="Buka portal publik"
          >
            <span className="material-symbols-outlined text-[18px]" aria-hidden>
              public
            </span>
            <span className="hidden md:inline">Portal</span>
          </Link>

          <div className="flex min-w-0 items-center gap-2.5 rounded-2xl border border-[var(--color-line)] bg-[var(--color-surface)] py-1.5 pr-3 pl-1.5">
            <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-[var(--color-accent)] text-xs font-bold text-[var(--color-secondary)] ring-2 ring-[var(--color-secondary)]/40">
              {initial}
            </span>
            <div className="hidden min-w-0 sm:block">
              <p className="max-w-[9.5rem] truncate text-sm font-bold leading-tight text-[var(--color-ink)]">
                {user?.name ?? '—'}
              </p>
              {user ? (
                <p className="truncate text-[10px] font-semibold tracking-wide text-[var(--color-muted)] uppercase">
                  {formatRoleLabel(user.role)}
                </p>
              ) : null}
            </div>
          </div>

          <button
            type="button"
            onClick={onLogout}
            className="inline-flex h-10 items-center gap-1.5 rounded-2xl bg-[var(--color-accent)] px-3.5 text-xs font-bold text-white shadow-[0_8px_18px_-10px_rgba(13,59,63,0.65)] transition hover:bg-[var(--color-accent-hover)]"
            aria-label="Keluar"
          >
            <span className="material-symbols-outlined text-[18px]" aria-hidden>
              logout
            </span>
            <span className="hidden sm:inline">Keluar</span>
          </button>
        </div>
      </div>
    </header>
  )
}
