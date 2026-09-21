import type { InputHTMLAttributes } from 'react'
import { cn } from '../../lib/utils'

type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string
  error?: string
  hint?: string
}

export function Input({ id, label, error, hint, className, ...props }: InputProps) {
  const inputId = id ?? props.name
  const descriptionId = inputId ? `${inputId}-${error ? 'error' : 'hint'}` : undefined

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
        aria-describedby={error || hint ? descriptionId : undefined}
        {...props}
      />
      {error ? (
        <p id={descriptionId} className="text-sm text-[var(--color-danger)]">
          {error}
        </p>
      ) : hint ? (
        <p id={descriptionId} className="text-xs text-[var(--color-muted)]">
          {hint}
        </p>
      ) : null}
    </div>
  )
}
