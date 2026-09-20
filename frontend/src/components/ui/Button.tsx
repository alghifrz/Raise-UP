import type { ButtonHTMLAttributes } from 'react'
import { cn } from '../../lib/utils'

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger'

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant
  loading?: boolean
}

const variantClasses: Record<ButtonVariant, string> = {
  primary:
    'bg-[var(--color-secondary)] text-[var(--color-secondary-ink)] hover:bg-[var(--color-secondary-dim)] disabled:bg-[var(--color-line)] disabled:text-[var(--color-muted)] shadow-[0_8px_20px_-8px_rgba(210,248,67,0.55)]',
  secondary:
    'border border-[var(--color-line)] bg-[var(--color-panel)] text-[var(--color-ink)] hover:bg-[var(--color-accent-soft)] disabled:opacity-60',
  ghost:
    'bg-transparent text-[var(--color-ink)] hover:bg-[var(--color-accent-soft)] disabled:opacity-60',
  danger:
    'bg-[var(--color-danger)] text-white hover:bg-red-800 disabled:bg-[var(--color-line)]',
}

export function Button({
  className,
  variant = 'primary',
  loading = false,
  disabled,
  children,
  type = 'button',
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      className={cn(
        'inline-flex items-center justify-center gap-2 rounded-full px-4 py-2.5 text-sm font-semibold transition',
        'disabled:cursor-not-allowed',
        variantClasses[variant],
        className,
      )}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      {...props}
    >
      {loading ? 'Memuat…' : children}
    </button>
  )
}
