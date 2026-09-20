import { Link } from 'react-router-dom'
import { Button } from '../components/ui/Button'
import { Badge } from '../components/ui/Badge'
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

  return (
    <>
      <header className="fixed top-0 right-0 left-0 z-20 flex h-[4.5rem] items-center justify-between gap-4 border-b border-[var(--color-line)] bg-[color-mix(in_srgb,var(--color-panel)_92%,transparent)] px-4 backdrop-blur-xl sm:px-6 lg:left-[17rem]">
        <div className="flex min-w-0 items-center gap-3">
          <button
            type="button"
            className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-[var(--color-line)] bg-[var(--color-panel)] text-[var(--color-accent)] shadow-sm lg:hidden"
            onClick={onMenuClick}
            aria-label="Buka menu navigasi"
            aria-controls="admin-sidebar"
            aria-expanded={menuOpen}
          >
            <span className="material-symbols-outlined text-[22px]" aria-hidden>
              menu
            </span>
          </button>
          <div className="min-w-0">
            <p className="truncate text-base font-bold tracking-tight text-[var(--color-ink)]">
              {title}
            </p>
            <p className="hidden text-xs font-medium text-[var(--color-muted)] sm:block">
              Panel administrasi {landingBrand.siteName}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2 sm:gap-3">
          <Link
            to="/"
            className="hidden items-center gap-1.5 rounded-full px-3 py-2 text-xs font-semibold text-[var(--color-muted)] transition hover:bg-[var(--color-accent-soft)] hover:text-[var(--color-accent)] md:inline-flex"
          >
            <span className="material-symbols-outlined text-[16px]" aria-hidden>
              public
            </span>
            Portal publik
          </Link>

          <div className="hidden items-center gap-3 rounded-full border border-[var(--color-line)] bg-[var(--color-panel)] py-1.5 pr-3 pl-1.5 sm:flex">
            <span className="flex h-8 w-8 items-center justify-center rounded-full bg-[var(--color-accent)] text-xs font-bold text-[var(--color-surface)]">
              {(user?.name ?? 'A').slice(0, 1).toUpperCase()}
            </span>
            <div className="min-w-0 text-left">
              <p className="max-w-[10rem] truncate text-sm font-semibold text-[var(--color-ink)]">
                {user?.name ?? '—'}
              </p>
              {user ? (
                <div className="mt-0.5">
                  <Badge tone="info">{formatRoleLabel(user.role)}</Badge>
                </div>
              ) : null}
            </div>
          </div>

          <Button type="button" variant="secondary" onClick={onLogout} className="rounded-full px-4">
            <span className="material-symbols-outlined text-[18px]" aria-hidden>
              logout
            </span>
            <span className="hidden sm:inline">Keluar</span>
          </Button>
        </div>
      </header>
      {/* Keeps page content clear of the fixed topbar */}
      <div className="h-[4.5rem] shrink-0" aria-hidden />
    </>
  )
}
