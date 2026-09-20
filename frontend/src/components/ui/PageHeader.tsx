import type { ReactNode } from 'react'

type PageHeaderProps = {
  title: string
  description?: string
  actions?: ReactNode
}

export function PageHeader({ title, description, actions }: PageHeaderProps) {
  return (
    <header className="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h1 className="text-[clamp(1.5rem,2.5vw,1.875rem)] font-bold tracking-tight text-[var(--color-ink)]">
          {title}
        </h1>
        {description ? <p className="mt-1 text-sm text-[var(--color-muted)]">{description}</p> : null}
      </div>
      {actions ? <div className="flex flex-wrap items-center gap-2">{actions}</div> : null}
    </header>
  )
}
