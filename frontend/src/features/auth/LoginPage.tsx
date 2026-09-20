import { motion, useReducedMotion } from 'framer-motion'
import { useState, type FormEvent } from 'react'
import { Link, Navigate } from 'react-router-dom'
import { isApiError } from '../../lib/api/types'
import { cn } from '../../lib/utils'
import { Icon } from '../landing/components/ui'
import { landingAssets, landingBrand } from '../landing/data'
import { useAuth } from './useAuth'

function toLoginErrorMessage(error: unknown): string {
  if (isApiError(error)) {
    if (error.status === 401 || error.code === 'INVALID_CREDENTIALS') {
      return 'Email atau kata sandi salah.'
    }
    if (error.code === 'NETWORK_ERROR') {
      return 'Tidak dapat terhubung ke server. Periksa koneksi Anda.'
    }
    if (error.status === 400) {
      return 'Email atau kata sandi tidak valid.'
    }
  }
  return 'Login gagal. Silakan coba lagi.'
}

export function LoginPage() {
  const { login, isAuthenticated, isInitializing } = useAuth()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [fieldErrors, setFieldErrors] = useState<{ email?: string; password?: string }>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const reduce = useReducedMotion()

  if (isInitializing) {
    return (
      <div className="landing-root flex min-h-screen items-center justify-center bg-[var(--lp-surface)]">
        <div className="flex items-center gap-3 text-sm font-semibold text-[var(--lp-primary)]">
          <span className="h-5 w-5 animate-spin rounded-full border-2 border-[var(--lp-secondary)] border-t-transparent" />
          Memuat sesi…
        </div>
      </div>
    )
  }

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const nextErrors: { email?: string; password?: string } = {}
    if (!email.trim()) {
      nextErrors.email = 'Email wajib diisi.'
    }
    if (!password) {
      nextErrors.password = 'Kata sandi wajib diisi.'
    }
    setFieldErrors(nextErrors)
    if (nextErrors.email || nextErrors.password) {
      return
    }

    setSubmitting(true)
    try {
      await login({ email: email.trim(), password })
    } catch (error) {
      setFormError(toLoginErrorMessage(error))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="landing-root min-h-screen bg-[var(--lp-surface)] text-[var(--lp-ink)] antialiased">
      <div className="mx-auto grid min-h-screen max-w-[1280px] lg:grid-cols-12">
        <motion.aside
          className="relative hidden overflow-hidden bg-[var(--lp-primary)] text-[var(--lp-on-primary)] lg:col-span-5 lg:flex lg:flex-col lg:justify-between lg:p-10 xl:p-12"
          initial={reduce ? false : { opacity: 0, x: -24 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.65, ease: [0.22, 1, 0.36, 1] }}
        >
          <div className="pointer-events-none absolute -right-24 -top-24 h-80 w-80 rounded-full bg-[var(--lp-secondary)]/15 blur-3xl" />
          <div className="pointer-events-none absolute -bottom-20 -left-16 h-64 w-64 rounded-full bg-[var(--lp-tertiary)]/40 blur-2xl" />

          <div className="relative z-10">
            <Link to="/" className="inline-flex items-center gap-3">
              <span className="flex h-11 w-11 items-center justify-center overflow-hidden rounded-xl bg-white ring-1 ring-white/15">
                <img src={landingAssets.logo} alt="" className="h-8 w-8 object-contain" />
              </span>
              <span>
                <span className="block text-base font-bold tracking-tight text-[var(--lp-white)]">
                  {landingBrand.siteName}
                </span>
                <span className="block text-xs font-medium text-[var(--lp-on-primary-muted)]">
                  {landingBrand.tagline}
                </span>
              </span>
            </Link>
          </div>

          <div className="relative z-10 space-y-6 py-16">
            <div className="inline-flex items-center gap-2 rounded-full bg-white/10 px-4 py-1.5 backdrop-blur-md">
              <span className="h-2 w-2 animate-pulse rounded-full bg-[var(--lp-secondary)]" />
              <span className="text-[11px] font-bold uppercase tracking-[0.08em] text-[var(--lp-secondary)]">
                Area Admin
              </span>
            </div>
            <h1 className="max-w-sm text-[clamp(2rem,3vw,2.75rem)] font-bold leading-tight tracking-tight text-[var(--lp-white)]">
              Kelola portal warga dengan{' '}
              <span className="text-[var(--lp-secondary)]">aman & transparan.</span>
            </h1>
            <p className="max-w-sm text-sm leading-relaxed text-[var(--lp-on-primary-muted)]">
              {landingBrand.tagline}. Masuk untuk mengelola pengumuman, kegiatan, galeri, dan
              administrasi RW.
            </p>
            <ul className="space-y-3 pt-2">
              {[
                { icon: 'verified_user', label: 'Akses terbatas pengurus' },
                { icon: 'campaign', label: 'Kelola informasi publik' },
                { icon: 'account_balance', label: 'Pencatatan digital tertib' },
              ].map((item) => (
                <li key={item.label} className="flex items-center gap-3 text-sm text-[var(--lp-on-primary)]">
                  <span className="flex h-9 w-9 items-center justify-center rounded-full bg-[var(--lp-secondary)]/20 text-[var(--lp-secondary)]">
                    <Icon name={item.icon} className="text-[18px]" />
                  </span>
                  {item.label}
                </li>
              ))}
            </ul>
          </div>

          <p className="relative z-10 text-xs text-[var(--lp-on-primary-muted)]">
            © {new Date().getFullYear()} {landingBrand.siteName}
          </p>
        </motion.aside>

        <main className="relative flex flex-col justify-center px-5 py-10 sm:px-8 lg:col-span-7 lg:px-14 xl:px-20">
          <div className="pointer-events-none absolute inset-0 overflow-hidden">
            <div className="absolute -right-20 top-10 h-64 w-64 rounded-full bg-[var(--lp-secondary)]/10 blur-3xl" />
            <div className="absolute bottom-10 left-10 h-48 w-48 rounded-full bg-[var(--lp-tertiary)]/10 blur-3xl" />
          </div>

          <motion.div
            className="relative z-10 mx-auto w-full max-w-md"
            initial={reduce ? false : { opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.55, ease: [0.22, 1, 0.36, 1], delay: 0.08 }}
          >
            <div className="mb-8 flex items-center justify-between gap-3 lg:hidden">
              <Link to="/" className="inline-flex items-center gap-2.5">
                <span className="flex h-10 w-10 items-center justify-center overflow-hidden rounded-xl bg-[var(--lp-primary)]">
                  <img src={landingAssets.logo} alt="" className="h-7 w-7 object-contain" />
                </span>
                <span className="text-base font-bold text-[var(--lp-primary)]">{landingBrand.siteName}</span>
              </Link>
              <Link
                to="/"
                className="inline-flex items-center gap-1 rounded-full px-3 py-2 text-xs font-semibold text-[var(--lp-muted)] hover:bg-[var(--lp-surface-low)] hover:text-[var(--lp-primary)]"
              >
                <Icon name="arrow_back" className="text-[16px]" />
                Beranda
              </Link>
            </div>

            <div className="mb-8">
              <p className="text-[11px] font-bold uppercase tracking-[0.16em] text-[var(--lp-tertiary)]">
                Masuk Admin
              </p>
              <h2 className="mt-2 text-[clamp(1.75rem,3vw,2.25rem)] font-bold tracking-tight text-[var(--lp-primary)]">
                Selamat datang kembali
              </h2>
              <p className="mt-2 text-sm leading-relaxed text-[var(--lp-muted)]">
                Gunakan akun administrator untuk mengelola portal {landingBrand.siteName}.
              </p>
            </div>

            <form
              className="rounded-3xl border border-black/[0.05] bg-[var(--lp-white)] p-6 shadow-[0_20px_50px_-24px_rgba(15,37,39,0.35)] sm:p-8"
              onSubmit={handleSubmit}
              noValidate
            >
              <div className="flex flex-col gap-5">
                <div className="flex flex-col gap-1.5">
                  <label htmlFor="email" className="text-sm font-semibold text-[var(--lp-primary)]">
                    Email
                  </label>
                  <div className="relative">
                    <span className="pointer-events-none absolute inset-y-0 left-3.5 flex items-center text-[var(--lp-muted)]">
                      <Icon name="mail" className="text-[20px]" />
                    </span>
                    <input
                      id="email"
                      name="email"
                      type="email"
                      autoComplete="username"
                      value={email}
                      onChange={(event) => setEmail(event.target.value)}
                      disabled={submitting}
                      required
                      placeholder="admin@lebakasri.id"
                      aria-invalid={fieldErrors.email ? true : undefined}
                      aria-describedby={fieldErrors.email ? 'email-error' : undefined}
                      className={cn(
                        'w-full rounded-2xl border bg-[var(--lp-surface)] py-3.5 pr-4 pl-11 text-sm text-[var(--lp-ink)] outline-none transition',
                        'placeholder:text-[var(--lp-muted)]/60',
                        'focus:border-[var(--lp-tertiary)] focus:bg-[var(--lp-white)] focus:ring-4 focus:ring-[var(--lp-tertiary)]/15',
                        'disabled:cursor-not-allowed disabled:opacity-60',
                        fieldErrors.email
                          ? 'border-red-400 focus:border-red-400 focus:ring-red-100'
                          : 'border-transparent',
                      )}
                    />
                  </div>
                  {fieldErrors.email ? (
                    <p id="email-error" className="text-sm text-red-600">
                      {fieldErrors.email}
                    </p>
                  ) : null}
                </div>

                <div className="flex flex-col gap-1.5">
                  <label htmlFor="password" className="text-sm font-semibold text-[var(--lp-primary)]">
                    Kata sandi
                  </label>
                  <div className="relative">
                    <span className="pointer-events-none absolute inset-y-0 left-3.5 flex items-center text-[var(--lp-muted)]">
                      <Icon name="lock" className="text-[20px]" />
                    </span>
                    <input
                      id="password"
                      name="password"
                      type={showPassword ? 'text' : 'password'}
                      autoComplete="current-password"
                      value={password}
                      onChange={(event) => setPassword(event.target.value)}
                      disabled={submitting}
                      required
                      placeholder="••••••••"
                      aria-invalid={fieldErrors.password ? true : undefined}
                      aria-describedby={fieldErrors.password ? 'password-error' : undefined}
                      className={cn(
                        'w-full rounded-2xl border bg-[var(--lp-surface)] py-3.5 pr-12 pl-11 text-sm text-[var(--lp-ink)] outline-none transition',
                        'placeholder:text-[var(--lp-muted)]/60',
                        'focus:border-[var(--lp-tertiary)] focus:bg-[var(--lp-white)] focus:ring-4 focus:ring-[var(--lp-tertiary)]/15',
                        'disabled:cursor-not-allowed disabled:opacity-60',
                        fieldErrors.password
                          ? 'border-red-400 focus:border-red-400 focus:ring-red-100'
                          : 'border-transparent',
                      )}
                    />
                    <button
                      type="button"
                      className="absolute inset-y-0 right-2 flex items-center rounded-xl px-2.5 text-[var(--lp-muted)] hover:text-[var(--lp-primary)]"
                      onClick={() => setShowPassword((value) => !value)}
                      aria-label={showPassword ? 'Sembunyikan kata sandi' : 'Tampilkan kata sandi'}
                    >
                      <Icon name={showPassword ? 'visibility_off' : 'visibility'} className="text-[20px]" />
                    </button>
                  </div>
                  {fieldErrors.password ? (
                    <p id="password-error" className="text-sm text-red-600">
                      {fieldErrors.password}
                    </p>
                  ) : null}
                </div>

                {formError ? (
                  <p
                    className="flex items-start gap-2 rounded-2xl border border-red-100 bg-red-50 px-4 py-3 text-sm text-red-700"
                    role="alert"
                  >
                    <Icon name="error" className="mt-0.5 shrink-0 text-[18px]" />
                    {formError}
                  </p>
                ) : null}

                <motion.button
                  type="submit"
                  disabled={submitting}
                  className="inline-flex w-full items-center justify-center gap-2 rounded-full bg-[var(--lp-secondary)] px-6 py-3.5 text-sm font-bold text-[var(--lp-on-secondary)] shadow-[0_10px_28px_-8px_rgba(210,248,67,0.65)] transition disabled:cursor-not-allowed disabled:opacity-60"
                  whileHover={reduce || submitting ? undefined : { y: -1, scale: 1.01 }}
                  whileTap={reduce || submitting ? undefined : { scale: 0.98 }}
                >
                  {submitting ? (
                    <>
                      <span className="h-4 w-4 animate-spin rounded-full border-2 border-[var(--lp-on-secondary)] border-t-transparent" />
                      Memuat…
                    </>
                  ) : (
                    <>
                      Masuk
                      <Icon name="arrow_forward" className="text-[18px]" />
                    </>
                  )}
                </motion.button>
              </div>
            </form>

            <div className="mt-6 flex flex-wrap items-center justify-between gap-3 text-sm">
              <Link
                to="/"
                className="hidden items-center gap-1.5 font-semibold text-[var(--lp-muted)] transition hover:text-[var(--lp-primary)] lg:inline-flex"
              >
                <Icon name="arrow_back" className="text-[18px]" />
                Kembali ke beranda
              </Link>
              <p className="text-[var(--lp-muted)]">
                Butuh bantuan?{' '}
                <a href="/#kontak" className="font-semibold text-[var(--lp-tertiary)] hover:text-[var(--lp-primary)]">
                  Hubungi pengurus
                </a>
              </p>
            </div>
          </motion.div>
        </main>
      </div>
    </div>
  )
}
