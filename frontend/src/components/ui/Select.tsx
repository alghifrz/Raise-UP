import type { SelectHTMLAttributes } from 'react'
import { cn } from '../../lib/utils'

type SelectOption = {
  value: string
  label: string
}

type SelectProps = SelectHTMLAttributes<HTMLSelectElement> & {
  label: string
  error?: string
  options: SelectOption[]
  placeholder?: string
}

export function Select({
  id,
  label,
  error,
  options,
  placeholder,
  className,
  ...props
}: SelectProps) {
  const selectId = id ?? props.name

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={selectId} className="text-sm font-semibold text-[var(--color-ink)]">
        {label}
        {props.required ? (
          <span className="text-[var(--color-danger)]" aria-hidden="true">
            {' '}
            *
          </span>
        ) : null}
      </label>
      <select
        id={selectId}
        className={cn(
          'rounded-2xl border border-[var(--color-line)] bg-[var(--color-surface)] px-3.5 py-2.5 text-sm text-[var(--color-ink)] outline-none transition',
          'focus:border-[var(--color-tertiary)] focus:bg-[var(--color-panel)] focus:ring-4 focus:ring-[var(--color-tertiary)]/15',
          'disabled:cursor-not-allowed disabled:opacity-60',
          error && 'border-[var(--color-danger)]',
          className,
        )}
        aria-invalid={error ? true : undefined}
        aria-describedby={error && selectId ? `${selectId}-error` : undefined}
        {...props}
      >
        {placeholder ? <option value="">{placeholder}</option> : null}
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      {error ? (
        <p id={selectId ? `${selectId}-error` : undefined} className="text-sm text-[var(--color-danger)]">
          {error}
        </p>
      ) : null}
    </div>
  )
}
