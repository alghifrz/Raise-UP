import { useMemo, useState } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../features/auth/useAuth'
import { cn } from '../lib/utils'
import { PageShell } from './PageShell'
import { adminNavItems } from './nav'
import { resolveNavLabel } from './nav-utils'
import { Sidebar } from './Sidebar'
import { Topbar } from './Topbar'

export function AdminLayout() {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const location = useLocation()
  const navigate = useNavigate()
  const { logout } = useAuth()
  const [navPath, setNavPath] = useState(location.pathname)

  if (navPath !== location.pathname) {
    setNavPath(location.pathname)
    setSidebarOpen(false)
  }

  const isChat = location.pathname === '/chat' || location.pathname.startsWith('/chat/')

  const title = useMemo(
    () => resolveNavLabel(location.pathname, adminNavItems),
    [location.pathname],
  )

  function handleLogout() {
    logout()
    void navigate('/login', { replace: true })
  }

  return (
    <div
      className={cn(
        'flex h-dvh overflow-hidden bg-[var(--color-surface)] text-[var(--color-ink)]',
      )}
    >
      <Sidebar open={sidebarOpen} onClose={() => setSidebarOpen(false)} />
      <div className="flex min-h-0 min-w-0 flex-1 flex-col">
        <Topbar
          title={title}
          menuOpen={sidebarOpen}
          onMenuClick={() => setSidebarOpen(true)}
          onLogout={handleLogout}
        />
        <main
          id="admin-main"
          className={cn(
            'flex min-h-0 flex-1 flex-col',
            isChat
              ? 'overflow-hidden p-0 lg:px-6 lg:py-4'
              : 'overflow-y-auto px-4 py-6 sm:px-6 lg:px-8',
          )}
        >
          {isChat ? (
            <Outlet />
          ) : (
            <PageShell>
              <Outlet />
            </PageShell>
          )}
        </main>
      </div>
    </div>
  )
}
