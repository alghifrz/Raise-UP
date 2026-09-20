import { Button } from './Button'
import { cn } from '../../lib/utils'

type ErrorStateProps = {
  title?: string
  message?: string
  onRetry?: () => void
  className?: string
}

export function ErrorState({
  title = 'Terjadi kesalahan',
  message = 'Silakan coba lagi.',
  onRetry,
  className,
}: ErrorStateProps) {
  return (
    <div
      className={cn(
        'flex min-h-40 flex-col items-center justify-center gap-3 rounded-xl border border-red-100 bg-[var(--color-danger-soft)] p-6 text-center',
        className,
      )}
      role="alert"
    >
      <h2 className="text-base font-semibold text-[var(--color-danger)]">{title}</h2>
      <p className="max-w-md text-sm text-red-800/80">{message}</p>
      {onRetry ? (
        <Button type="button" variant="secondary" onClick={onRetry}>
          Coba lagi
        </Button>
      ) : null}
    </div>
  )
}
