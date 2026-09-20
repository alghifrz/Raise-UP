import type { ReactNode } from 'react'
import { cn } from '../../lib/utils'

type InlineAlertProps = {
  tone?: 'success' | 'danger'
  children: ReactNode
  className?: string
}

export function InlineAlert({ tone = 'success', children, className }: InlineAlertProps) {
  return (
    <p
      className={cn(
        'rounded-2xl border px-4 py-2.5 text-sm font-medium',
        tone === 'success' &&
          'border-[color-mix(in_srgb,var(--color-secondary)_45%,transparent)] bg-[color-mix(in_srgb,var(--color-secondary)_28%,white)] text-[var(--color-accent)]',
        tone === 'danger' &&
          'border-red-100 bg-[var(--color-danger-soft)] text-[var(--color-danger)]',
        className,
      )}
      role={tone === 'danger' ? 'alert' : 'status'}
    >
      {children}
    </p>
  )
}
