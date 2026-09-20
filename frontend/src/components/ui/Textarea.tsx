import type { TextareaHTMLAttributes } from 'react'
import { cn } from '../../lib/utils'

type TextareaProps = TextareaHTMLAttributes<HTMLTextAreaElement> & {
  label: string
  error?: string
}

export function Textarea({ id, label, error, className, ...props }: TextareaProps) {
  const textareaId = id ?? props.name

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={textareaId} className="text-sm font-semibold text-[var(--color-ink)]">
        {label}
        {props.required ? (
          <span className="text-[var(--color-danger)]" aria-hidden="true">
            {' '}
            *
          </span>
        ) : null}
      </label>
      <textarea
        id={textareaId}
        className={cn(
          'min-h-28 rounded-2xl border border-[var(--color-line)] bg-[var(--color-surface)] px-3.5 py-2.5 text-sm text-[var(--color-ink)] outline-none transition',
          'placeholder:text-[var(--color-muted)]/60',
          'focus:border-[var(--color-tertiary)] focus:bg-[var(--color-panel)] focus:ring-4 focus:ring-[var(--color-tertiary)]/15',
          'disabled:cursor-not-allowed disabled:opacity-60',
          error && 'border-[var(--color-danger)]',
          className,
        )}
        aria-invalid={error ? true : undefined}
        aria-describedby={error && textareaId ? `${textareaId}-error` : undefined}
        {...props}
      />
      {error ? (
        <p id={textareaId ? `${textareaId}-error` : undefined} className="text-sm text-[var(--color-danger)]">
          {error}
        </p>
      ) : null}
    </div>
  )
}
