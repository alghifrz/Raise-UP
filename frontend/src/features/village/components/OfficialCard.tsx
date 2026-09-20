import { Button } from '../../../components/ui/Button'
import { ImageWithFallback } from '../../../components/ui/ImageWithFallback'
import type { VillageOfficial } from '../types'

type OfficialCardProps = {
  official: VillageOfficial
  onEdit: (official: VillageOfficial) => void
  onDelete: (official: VillageOfficial) => void
}

function initials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')
}

export function OfficialCard({ official, onEdit, onDelete }: OfficialCardProps) {
  return (
    <article className="rounded-2xl border border-[var(--color-line)] bg-[var(--color-panel)] p-4 shadow-sm">
      <div className="flex items-start gap-3">
        {official.photo_url ? (
          <ImageWithFallback
            src={official.photo_url}
            alt={official.name}
            className="h-16 w-16 shrink-0 rounded-full"
            fallbackLabel={initials(official.name) || '—'}
          />
        ) : (
          <div
            className="flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-[var(--color-accent-soft)] text-sm font-semibold text-[var(--color-muted)]"
            aria-hidden
          >
            {initials(official.name) || '—'}
          </div>
        )}
        <div className="min-w-0 flex-1">
          <h3 className="truncate font-semibold text-[var(--color-ink)]">{official.name}</h3>
          <p className="text-sm text-[var(--color-muted)]">{official.position}</p>
          <p className="mt-1 text-xs text-[var(--color-muted)]">Urutan {official.sort_order}</p>
          <div className="mt-3 flex gap-2">
            <Button
              type="button"
              variant="secondary"
              className="px-2 py-1"
              onClick={() => onEdit(official)}
            >
              Edit
            </Button>
            <Button
              type="button"
              variant="danger"
              className="px-2 py-1"
              onClick={() => onDelete(official)}
            >
              Hapus
            </Button>
          </div>
        </div>
      </div>
    </article>
  )
}
