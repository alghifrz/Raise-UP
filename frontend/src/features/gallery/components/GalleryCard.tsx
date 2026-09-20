import { Button } from '../../../components/ui/Button'
import { ImageWithFallback } from '../../../components/ui/ImageWithFallback'
import { formatDateTime } from '../../../lib/utils'
import type { GalleryItem } from '../types'

type GalleryCardProps = {
  item: GalleryItem
  onEdit: (item: GalleryItem) => void
  onDelete: (item: GalleryItem) => void
}

export function GalleryCard({ item, onEdit, onDelete }: GalleryCardProps) {
  return (
    <article className="overflow-hidden rounded-2xl border border-[var(--color-line)] bg-[var(--color-panel)] shadow-sm">
      <ImageWithFallback
        src={item.image_url}
        alt={item.caption || 'Foto galeri'}
        className="aspect-[4/3] w-full"
        fallbackLabel="Gambar tidak tersedia"
      />
      <div className="space-y-3 p-4">
        <div>
          <p className="line-clamp-2 text-sm font-medium text-[var(--color-ink)]">
            {item.caption || 'Tanpa caption'}
          </p>
          <p className="mt-1 text-xs text-[var(--color-muted)]">
            Urutan {item.sort_order} · {formatDateTime(item.created_at)}
          </p>
        </div>
        <div className="flex gap-2">
          <Button type="button" variant="secondary" className="px-2 py-1" onClick={() => onEdit(item)}>
            Edit
          </Button>
          <Button type="button" variant="danger" className="px-2 py-1" onClick={() => onDelete(item)}>
            Hapus
          </Button>
        </div>
      </div>
    </article>
  )
}
