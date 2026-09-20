import type { ReactNode } from 'react'
import { cn } from '../../lib/utils'

type Column<T> = {
  key: string
  header: string
  className?: string
  render: (row: T) => ReactNode
}

type DataTableProps<T> = {
  columns: Column<T>[]
  rows: T[]
  rowKey: (row: T) => string
  className?: string
}

export function DataTable<T>({ columns, rows, rowKey, className }: DataTableProps<T>) {
  return (
    <div
      className={cn(
        'overflow-x-auto rounded-3xl border border-[var(--color-line)] bg-[var(--color-panel)] shadow-[0_10px_30px_-20px_rgba(15,37,39,0.2)]',
        className,
      )}
    >
      <table className="min-w-full divide-y divide-[var(--color-line)] text-left text-sm">
        <thead className="bg-[var(--color-accent-soft)]">
          <tr>
            {columns.map((column) => (
              <th
                key={column.key}
                scope="col"
                className={cn(
                  'whitespace-nowrap px-4 py-3 text-xs font-bold uppercase tracking-wide text-[var(--color-muted)]',
                  column.className,
                )}
              >
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-[var(--color-line)]">
          {rows.map((row) => (
            <tr key={rowKey(row)} className="transition-colors hover:bg-[var(--color-accent-soft)]/50">
              {columns.map((column) => (
                <td
                  key={column.key}
                  className={cn('whitespace-nowrap px-4 py-3 text-[var(--color-ink)]', column.className)}
                >
                  {column.render(row)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
