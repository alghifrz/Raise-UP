import { useEffect } from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import { useChatUnreadTotal } from '../features/chat/hooks'
import { landingAssets, landingBrand } from '../features/landing/data'
import { cn } from '../lib/utils'
import { adminNavItems } from './nav'
import { isNavItemActive } from './nav-utils'

type SidebarProps = {
  open: boolean
  onClose: () => void
}

function formatUnreadBadge(count: number): string {
  return count > 99 ? '99+' : String(count)
}

export function Sidebar({ open, onClose }: SidebarProps) {
  const location = useLocation()
  const unreadQuery = useChatUnreadTotal()
  const unreadTotal = unreadQuery.data ?? 0

  useEffect(() => {
    if (!open) {
      return
    }

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [open, onClose])

  return (
    <>
      <button
        type="button"
        className={cn(
          'fixed inset-0 z-30 bg-[var(--color-ink)]/40 transition-opacity lg:hidden',
          open ? 'opacity-100' : 'pointer-events-none opacity-0',
        )}
        onClick={onClose}
        aria-label="Tutup menu navigasi"
        tabIndex={open ? 0 : -1}
      />

      <aside
        id="admin-sidebar"
        className={cn(
          'fixed inset-y-0 left-0 z-40 flex h-dvh w-[17rem] flex-col bg-[var(--color-accent)] text-[var(--color-surface)] transition-transform',
          open ? 'translate-x-0' : '-translate-x-full lg:translate-x-0',
        )}
        aria-label="Navigasi admin"
      >
        <div className="flex h-[4.5rem] shrink-0 items-center gap-3 border-b border-white/10 px-5">
          <NavLink to="/dashboard" className="flex min-w-0 items-center gap-3" onClick={onClose}>
            <span className="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-xl bg-white/10 ring-1 ring-white/15">
              <img src={landingAssets.logo} alt="" className="h-7 w-7 object-contain" />
            </span>
            <span className="min-w-0">
              <span className="block truncate text-sm font-bold tracking-tight text-white">
                {landingBrand.siteName}
              </span>
              <span className="block truncate text-[11px] font-medium text-white/55">
                Admin · {landingBrand.tagline}
              </span>
            </span>
          </NavLink>
        </div>

        <nav className="min-h-0 flex-1 overflow-y-auto px-3 py-4">
          <p className="mb-2 px-3 text-[10px] font-bold uppercase tracking-[0.16em] text-white/40">
            Menu
          </p>
          <ul className="flex flex-col gap-1">
            {adminNavItems.map((item) => {
              const active = isNavItemActive(location.pathname, item.to)
              const showUnread = item.to === '/chat' && unreadTotal > 0
              return (
                <li key={item.to}>
                  <NavLink
                    to={item.to}
                    end={item.to === '/dashboard'}
                    onClick={onClose}
                    aria-current={active ? 'page' : undefined}
                    aria-label={
                      showUnread
                        ? `${item.label}, ${unreadTotal} pesan belum dibaca`
                        : undefined
                    }
                    className={cn(
                      'group flex items-center gap-3 rounded-2xl px-3 py-2.5 text-sm font-semibold transition-colors',
                      active
                        ? 'bg-[var(--color-secondary)] text-[var(--color-secondary-ink)] shadow-[0_8px_20px_-10px_rgba(210,248,67,0.8)]'
                        : 'text-white/75 hover:bg-white/8 hover:text-white',
                    )}
                  >
                    <span
                      className={cn(
                        'material-symbols-outlined text-[20px]',
                        active ? 'text-[var(--color-secondary-ink)]' : 'text-white/55 group-hover:text-white/80',
                      )}
                      aria-hidden
                    >
                      {item.icon}
                    </span>
                    <span className="min-w-0 flex-1 truncate">{item.label}</span>
                    {showUnread ? (
                      <span
                        className={cn(
                          'inline-flex min-w-5 shrink-0 items-center justify-center rounded-full px-1.5 py-0.5 text-[10px] font-bold',
                          active
                            ? 'bg-[var(--color-accent)] text-white'
                            : 'bg-[var(--color-secondary)] text-[var(--color-secondary-ink)]',
                        )}
                        aria-hidden
                      >
                        {formatUnreadBadge(unreadTotal)}
                      </span>
                    ) : null}
                  </NavLink>
                </li>
              )
            })}
          </ul>
        </nav>

        <div className="shrink-0 border-t border-white/10 p-4">
          <NavLink
            to="/"
            onClick={onClose}
            className="flex items-center gap-2 rounded-2xl bg-white/8 px-3 py-2.5 text-xs font-semibold text-white/70 transition hover:bg-white/12 hover:text-white"
          >
            <span className="material-symbols-outlined text-[18px]" aria-hidden>
              public
            </span>
            Lihat landing page
          </NavLink>
        </div>
      </aside>

      {/* Keeps main content clear of the fixed sidebar on desktop */}
      <div className="hidden w-[17rem] shrink-0 lg:block" aria-hidden />
    </>
  )
}
