import type { ReactNode } from 'react'
import { cn } from '../../lib/utils'

type EmptyStateProps = {
  title: string
  description?: string
  action?: ReactNode
  className?: string
}

export function EmptyState({ title, description, action, className }: EmptyStateProps) {
  return (
    <div
      className={cn(
        'flex min-h-40 flex-col items-center justify-center gap-2 rounded-3xl border border-dashed border-[var(--color-line)] bg-[var(--color-accent-soft)]/60 px-6 py-10 text-center',
        className,
      )}
    >
      <h2 className="text-base font-bold text-[var(--color-ink)]">{title}</h2>
      {description ? <p className="max-w-md text-sm text-[var(--color-muted)]">{description}</p> : null}
      {action}
    </div>
  )
}
