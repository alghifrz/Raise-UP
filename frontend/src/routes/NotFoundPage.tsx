import { Link } from 'react-router-dom'
import { useAuth } from '../features/auth/useAuth'
import { landingAssets, landingBrand } from '../features/landing/data'

export function NotFoundPage() {
  const { isAuthenticated } = useAuth()
  const homeTo = isAuthenticated ? '/dashboard' : '/'

  return (
    <div className="flex min-h-screen items-center justify-center bg-[var(--color-surface)] px-4 py-10">
      <div className="w-full max-w-md rounded-3xl border border-[var(--color-line)] bg-[var(--color-panel)] p-8 text-center shadow-[0_20px_50px_-24px_rgba(15,37,39,0.3)]">
        <img src={landingAssets.logo} alt="" className="mx-auto h-12 w-12 object-contain" />
        <p className="mt-4 text-xs font-bold uppercase tracking-[0.16em] text-[var(--color-tertiary)]">
          {landingBrand.siteName}
        </p>
        <p className="mt-3 text-5xl font-extrabold tracking-tight text-[var(--color-accent)]">404</p>
        <h1 className="mt-2 text-xl font-bold text-[var(--color-ink)]">Halaman tidak ditemukan</h1>
        <p className="mt-2 text-sm text-[var(--color-muted)]">
          Alamat yang Anda buka tidak tersedia atau telah dipindahkan.
        </p>
        <div className="mt-6">
          <Link
            to={homeTo}
            className="inline-flex items-center justify-center rounded-full bg-[var(--color-secondary)] px-5 py-2.5 text-sm font-bold text-[var(--color-secondary-ink)] shadow-[0_8px_20px_-8px_rgba(210,248,67,0.55)] hover:bg-[var(--color-secondary-dim)]"
          >
            {isAuthenticated ? 'Kembali ke Dashboard' : 'Ke beranda'}
          </Link>
        </div>
      </div>
    </div>
  )
}
