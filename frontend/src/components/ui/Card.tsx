import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '../../lib/utils'

type CardProps = HTMLAttributes<HTMLDivElement> & {
  title?: string
  description?: string
  action?: ReactNode
}

export function Card({ title, description, action, className, children, ...props }: CardProps) {
  return (
    <section
      className={cn(
        'rounded-3xl border border-[var(--color-line)] bg-[var(--color-panel)] p-5 shadow-[0_10px_30px_-20px_rgba(15,37,39,0.25)]',
        className,
      )}
      {...props}
    >
      {(title || action) && (
        <div className="mb-4 flex items-start justify-between gap-3">
          <div>
            {title ? (
              <h2 className="text-base font-bold tracking-tight text-[var(--color-ink)]">{title}</h2>
            ) : null}
            {description ? (
              <p className="mt-1 text-sm text-[var(--color-muted)]">{description}</p>
            ) : null}
          </div>
          {action}
        </div>
      )}
      {children}
    </section>
  )
}
