import { cn } from '../../lib/utils'

type LoadingStateProps = {
  label?: string
  className?: string
}

export function LoadingState({ label = 'Memuat…', className }: LoadingStateProps) {
  return (
    <div
      className={cn(
        'flex min-h-40 flex-col items-center justify-center gap-3 text-[var(--color-muted)]',
        className,
      )}
      role="status"
      aria-live="polite"
    >
      <div
        className="h-8 w-8 animate-spin rounded-full border-2 border-[var(--color-secondary)] border-t-transparent"
        aria-hidden
      />
      <p className="text-sm font-medium">{label}</p>
    </div>
  )
}

export function Skeleton({ className }: { className?: string }) {
  return (
    <div
      className={cn('animate-pulse rounded-2xl bg-[var(--color-accent-soft)]', className)}
      aria-hidden
    />
  )
}
