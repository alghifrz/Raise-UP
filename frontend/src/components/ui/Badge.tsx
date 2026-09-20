import type { HTMLAttributes } from 'react'
import { cn } from '../../lib/utils'

type BadgeTone = 'neutral' | 'info' | 'success' | 'warning' | 'danger'

type BadgeProps = HTMLAttributes<HTMLSpanElement> & {
  tone?: BadgeTone
}

const toneClasses: Record<BadgeTone, string> = {
  neutral: 'bg-[var(--color-accent-soft)] text-[var(--color-accent)]',
  info: 'bg-[var(--color-accent-soft)] text-[var(--color-tertiary)]',
  success: 'bg-[color-mix(in_srgb,var(--color-secondary)_35%,white)] text-[var(--color-accent)]',
  warning: 'bg-amber-50 text-amber-900',
  danger: 'bg-[var(--color-danger-soft)] text-[var(--color-danger)]',
}

export function Badge({ tone = 'neutral', className, children, ...props }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-bold',
        toneClasses[tone],
        className,
      )}
      {...props}
    >
      {children}
    </span>
  )
}
