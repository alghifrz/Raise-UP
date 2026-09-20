import { useMemo, useState } from 'react'
import { Button } from '../../../components/ui/Button'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorState } from '../../../components/ui/ErrorState'
import { Input } from '../../../components/ui/Input'
import { InlineAlert } from '../../../components/ui/InlineAlert'
import { LoadingState } from '../../../components/ui/LoadingState'
import { ConfirmDialog, Modal } from '../../../components/ui/Modal'
import { PageHeader } from '../../../components/ui/PageHeader'
import { Pagination } from '../../../components/ui/Pagination'
import { useDebouncedValue } from '../../../lib/hooks/useDebouncedValue'
import { GalleryCard } from '../components/GalleryCard'
import { GalleryForm } from '../components/GalleryForm'
import { toGalleryErrorMessage } from '../errors'
import {
  useCreateGalleryItem,
  useDeleteGalleryItem,
  useGalleryItems,
  useUpdateGalleryItem,
} from '../hooks'
import type { GalleryFilters, GalleryItem } from '../types'

const DEFAULT_PAGE_SIZE = 12

export function GalleryPage() {
  const [searchInput, setSearchInput] = useState('')
  const debouncedSearch = useDebouncedValue(searchInput, 300)
  const [page, setPage] = useState(1)
  const [createOpen, setCreateOpen] = useState(false)
  const [editing, setEditing] = useState<GalleryItem | null>(null)
  const [deleting, setDeleting] = useState<GalleryItem | null>(null)
  const [feedback, setFeedback] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const filters = useMemo<GalleryFilters>(
    () => ({
      page,
      page_size: DEFAULT_PAGE_SIZE,
      search: debouncedSearch.trim(),
    }),
    [page, debouncedSearch],
  )

  const { data, isLoading, isError, isFetching, refetch, error } = useGalleryItems(filters)
  const createMutation = useCreateGalleryItem()
  const updateMutation = useUpdateGalleryItem()
  const deleteMutation = useDeleteGalleryItem()

  const items = data?.items ?? []
  const meta = data?.meta

  return (
    <div className="space-y-4">
      <PageHeader
        title="Galeri"
        description="Kelola foto dokumentasi kegiatan dan lingkungan."
        actions={
          <Button type="button" onClick={() => setCreateOpen(true)}>
            Tambah Foto
          </Button>
        }
      />

      {feedback ? <InlineAlert>{feedback}</InlineAlert> : null}

      <div className="rounded-2xl border border-[var(--color-line)] bg-[var(--color-panel)] p-4">
        <Input
          name="search"
          label="Cari"
          placeholder="Caption foto"
          value={searchInput}
          onChange={(event) => {
            setSearchInput(event.target.value)
            setPage(1)
          }}
        />
      </div>

      {isLoading ? <LoadingState label="Memuat galeri…" /> : null}

      {isError ? (
        <ErrorState
          title="Gagal memuat galeri."
          message={toGalleryErrorMessage(error)}
          onRetry={() => {
            void refetch()
          }}
        />
      ) : null}

      {!isLoading && !isError && items.length === 0 ? (
        <EmptyState
          title="Belum ada foto di galeri."
          description="Tambahkan foto baru menggunakan URL gambar."
          action={
            <Button type="button" onClick={() => setCreateOpen(true)}>
              Tambah Foto
            </Button>
          }
        />
      ) : null}

      {!isLoading && !isError && items.length > 0 && meta ? (
        <div className="space-y-4">
          {isFetching ? (
            <p className="text-xs text-[var(--color-muted)]" aria-live="polite">
              Memperbarui…
            </p>
          ) : null}
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {items.map((item) => (
              <GalleryCard
                key={item.id}
                item={item}
                onEdit={setEditing}
                onDelete={(galleryItem) => {
                  setDeleteError(null)
                  setDeleting(galleryItem)
                }}
              />
            ))}
          </div>
          <Pagination meta={meta} onPageChange={setPage} disabled={isFetching} />
        </div>
      ) : null}

      <Modal
        open={createOpen}
        title="Tambah Foto"
        description="Masukkan URL gambar yang dapat diakses publik."
        onClose={() => setCreateOpen(false)}
      >
        <GalleryForm
          mode="create"
          submitLabel="Simpan"
          onCancel={() => setCreateOpen(false)}
          onSubmitCreate={async (payload) => {
            await createMutation.mutateAsync(payload)
            setCreateOpen(false)
            setFeedback('Foto berhasil ditambahkan.')
          }}
        />
      </Modal>

      <Modal open={Boolean(editing)} title="Edit Foto" onClose={() => setEditing(null)}>
        {editing ? (
          <GalleryForm
            key={editing.id}
            mode="edit"
            initial={editing}
            submitLabel="Simpan perubahan"
            onCancel={() => setEditing(null)}
            onSubmitUpdate={async (payload) => {
              await updateMutation.mutateAsync({ id: editing.id, payload })
              setEditing(null)
              setFeedback('Foto berhasil diperbarui.')
            }}
          />
        ) : null}
      </Modal>

      <ConfirmDialog
        open={Boolean(deleting)}
        title="Hapus foto?"
        message="Foto akan dihapus dari galeri secara permanen."
        confirmLabel="Hapus"
        loading={deleteMutation.isPending}
        onCancel={() => {
          setDeleting(null)
          setDeleteError(null)
        }}
        onConfirm={() => {
          if (!deleting) {
            return
          }
          void (async () => {
            try {
              await deleteMutation.mutateAsync(deleting.id)
              setDeleting(null)
              setFeedback('Foto berhasil dihapus.')
            } catch (error) {
              setDeleteError(toGalleryErrorMessage(error))
            }
          })()
        }}
      />

      {deleteError ? <InlineAlert tone="danger">{deleteError}</InlineAlert> : null}
    </div>
  )
}
