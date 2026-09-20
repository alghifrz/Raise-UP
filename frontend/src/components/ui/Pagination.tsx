import { Button } from './Button'
import type { PaginationMeta } from '../../lib/api/types'

type PaginationProps = {
  meta: PaginationMeta
  onPageChange: (page: number) => void
  disabled?: boolean
}

export function Pagination({ meta, onPageChange, disabled = false }: PaginationProps) {
  const totalPages = Math.max(meta.total_pages, 1)
  const canPrev = meta.page > 1
  const canNext = meta.page < totalPages && meta.total_pages > 0

  return (
    <div className="flex flex-wrap items-center justify-between gap-3 border-t border-[var(--color-line)] pt-4">
      <p className="text-sm text-[var(--color-muted)]">
        Halaman {meta.page} dari {totalPages} · {meta.total} data
      </p>
      <div className="flex gap-2">
        <Button
          type="button"
          variant="secondary"
          disabled={disabled || !canPrev}
          onClick={() => onPageChange(meta.page - 1)}
        >
          Sebelumnya
        </Button>
        <Button
          type="button"
          variant="secondary"
          disabled={disabled || !canNext}
          onClick={() => onPageChange(meta.page + 1)}
        >
          Berikutnya
        </Button>
      </div>
    </div>
  )
}
