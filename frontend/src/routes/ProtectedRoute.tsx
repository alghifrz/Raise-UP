import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { LoadingState } from '../components/ui/LoadingState'
import { useAuth } from '../features/auth/useAuth'

export function ProtectedRoute() {
  const { isAuthenticated, isInitializing } = useAuth()
  const location = useLocation()

  if (isInitializing) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <LoadingState label="Memuat sesi…" />
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  return <Outlet />
}
