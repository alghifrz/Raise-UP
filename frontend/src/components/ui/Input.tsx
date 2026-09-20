import type { InputHTMLAttributes } from 'react'
import { cn } from '../../lib/utils'

type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string
  error?: string
}

export function Input({ id, label, error, className, ...props }: InputProps) {
  const inputId = id ?? props.name

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={inputId} className="text-sm font-semibold text-[var(--color-ink)]">
        {label}
        {props.required ? (
          <span className="text-[var(--color-danger)]" aria-hidden="true">
            {' '}
            *
          </span>
        ) : null}
      </label>
      <input
        id={inputId}
        className={cn(
          'rounded-2xl border border-[var(--color-line)] bg-[var(--color-surface)] px-3.5 py-2.5 text-sm text-[var(--color-ink)] outline-none transition',
          'placeholder:text-[var(--color-muted)]/60',
          'focus:border-[var(--color-tertiary)] focus:bg-[var(--color-panel)] focus:ring-4 focus:ring-[var(--color-tertiary)]/15',
          'disabled:cursor-not-allowed disabled:opacity-60',
          error && 'border-[var(--color-danger)] focus:border-[var(--color-danger)] focus:ring-red-100',
          className,
        )}
        aria-invalid={error ? true : undefined}
        aria-describedby={error && inputId ? `${inputId}-error` : undefined}
        {...props}
      />
      {error ? (
        <p id={inputId ? `${inputId}-error` : undefined} className="text-sm text-[var(--color-danger)]">
          {error}
        </p>
      ) : null}
    </div>
  )
}
