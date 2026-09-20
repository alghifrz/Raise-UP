import { Component, type ErrorInfo, type ReactNode } from 'react'
import { landingAssets, landingBrand } from '../features/landing/data'
import { Button } from './ui/Button'

type Props = {
  children: ReactNode
}

type State = {
  hasError: boolean
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false }

  static getDerivedStateFromError(): State {
    return { hasError: true }
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    if (import.meta.env.DEV) {
      console.error('ErrorBoundary caught', error, info.componentStack)
    }
  }

  private handleReload = () => {
    window.location.assign('/dashboard')
  }

  private handleRetry = () => {
    this.setState({ hasError: false })
  }

  render() {
    if (!this.state.hasError) {
      return this.props.children
    }

    return (
      <div className="flex min-h-screen items-center justify-center bg-[var(--color-surface)] px-4 py-10">
        <div
          className="w-full max-w-md rounded-3xl border border-[var(--color-line)] bg-[var(--color-panel)] p-8 text-center shadow-[0_20px_50px_-24px_rgba(15,37,39,0.3)]"
          role="alert"
        >
          <img src={landingAssets.logo} alt="" className="mx-auto h-12 w-12 object-contain" />
          <p className="mt-4 text-xs font-bold uppercase tracking-[0.16em] text-[var(--color-tertiary)]">
            {landingBrand.siteName}
          </p>
          <h1 className="mt-2 text-xl font-bold text-[var(--color-ink)]">Terjadi kesalahan</h1>
          <p className="mt-2 text-sm text-[var(--color-muted)]">
            Aplikasi mengalami kesalahan tak terduga. Silakan coba lagi atau muat ulang halaman.
          </p>
          <div className="mt-6 flex flex-col gap-2 sm:flex-row sm:justify-center">
            <Button type="button" variant="secondary" onClick={this.handleRetry}>
              Coba lagi
            </Button>
            <Button type="button" onClick={this.handleReload}>
              Muat ulang
            </Button>
          </div>
        </div>
      </div>
    )
  }
}
