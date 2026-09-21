import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '../../../components/ui/Button'
import { DataTable } from '../../../components/ui/DataTable'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorState } from '../../../components/ui/ErrorState'
import { Input } from '../../../components/ui/Input'
import { LoadingState } from '../../../components/ui/LoadingState'
import { Modal } from '../../../components/ui/Modal'
import { PageHeader } from '../../../components/ui/PageHeader'
import { Pagination } from '../../../components/ui/Pagination'
import { Select } from '../../../components/ui/Select'
import { useDebouncedValue } from '../../../lib/hooks/useDebouncedValue'
import { formatDateTime } from '../../../lib/utils'
import { AnnouncementForm } from '../components/AnnouncementForm'
import {
  AnnouncementStatusBadge,
  AnnouncementVisibilityBadge,
} from '../components/AnnouncementBadges'
import { useAnnouncements, useCreateAnnouncement } from '../hooks'
import type {
  AnnouncementFilters,
  AnnouncementStatus,
  AnnouncementVisibility,
} from '../types'
import {
  ANNOUNCEMENT_STATUS_OPTIONS,
  ANNOUNCEMENT_VISIBILITY_OPTIONS,
  formatAuthorLabel,
} from '../types'

const DEFAULT_PAGE_SIZE = 20

export function AnnouncementsPage() {
  const [searchInput, setSearchInput] = useState('')
  const debouncedSearch = useDebouncedValue(searchInput, 300)
  const [status, setStatus] = useState<AnnouncementStatus | ''>('')
  const [visibility, setVisibility] = useState<AnnouncementVisibility | ''>('')
  const [page, setPage] = useState(1)
  const [createOpen, setCreateOpen] = useState(false)
  const [feedback, setFeedback] = useState<string | null>(null)

  const filters = useMemo<AnnouncementFilters>(
    () => ({
      page,
      page_size: DEFAULT_PAGE_SIZE,
      search: debouncedSearch.trim(),
      status,
      visibility,
      category: '',
    }),
    [page, debouncedSearch, status, visibility],
  )

  const { data, isLoading, isError, isFetching, refetch } = useAnnouncements(filters)
  const createMutation = useCreateAnnouncement()

  const rows = data?.items ?? []
  const meta = data?.meta

  return (
    <div className="space-y-4">
      <PageHeader
        title="Pengumuman"
        description="Kelola pengumuman publik dan privat."
        actions={
          <Button type="button" onClick={() => setCreateOpen(true)}>
            Buat Pengumuman
          </Button>
        }
      />

      {feedback ? (
        <p className="rounded-md border border-emerald-100 bg-emerald-50 px-3 py-2 text-sm text-emerald-800" role="status">
          {feedback}
        </p>
      ) : null}

      <div className="grid gap-3 rounded-2xl border border-[var(--color-line)] bg-[var(--color-panel)] p-4 sm:grid-cols-3">
        <Input
          name="search"
          label="Cari"
          placeholder="Judul atau isi"
          value={searchInput}
          onChange={(event) => {
            setSearchInput(event.target.value)
            setPage(1)
          }}
        />
        <Select
          name="statusFilter"
          label="Status"
          value={status}
          onChange={(event) => {
            setStatus(event.target.value as AnnouncementStatus | '')
            setPage(1)
          }}
          options={ANNOUNCEMENT_STATUS_OPTIONS}
          placeholder="Semua"
        />
        <Select
          name="visibilityFilter"
          label="Visibilitas"
          value={visibility}
          onChange={(event) => {
            setVisibility(event.target.value as AnnouncementVisibility | '')
            setPage(1)
          }}
          options={ANNOUNCEMENT_VISIBILITY_OPTIONS}
          placeholder="Semua"
        />
      </div>

      {isLoading ? <LoadingState label="Memuat pengumuman…" /> : null}

      {isError ? (
        <ErrorState
          title="Gagal memuat pengumuman."
          message="Periksa koneksi atau coba lagi."
          onRetry={() => {
            void refetch()
          }}
        />
      ) : null}

      {!isLoading && !isError && rows.length === 0 ? (
        <EmptyState
          title="Belum ada pengumuman"
          description="Buat pengumuman baru atau ubah filter pencarian."
          action={
            <Button type="button" onClick={() => setCreateOpen(true)}>
              Buat Pengumuman
            </Button>
          }
        />
      ) : null}

      {!isLoading && !isError && rows.length > 0 && meta ? (
        <div className="space-y-3">
          {isFetching ? (
            <p className="text-xs text-[var(--color-muted)]" aria-live="polite">
              Memperbarui…
            </p>
          ) : null}
          <DataTable
            rows={rows}
            rowKey={(row) => row.id}
            columns={[
              {
                key: 'title',
                header: 'Judul',
                render: (row) => (
                  <Link
                    className="font-medium text-[var(--color-accent)] hover:underline"
                    to={`/announcements/${row.id}`}
                  >
                    {row.title}
                  </Link>
                ),
              },
              {
                key: 'visibility',
                header: 'Visibilitas',
                render: (row) => <AnnouncementVisibilityBadge visibility={row.visibility} />,
              },
              {
                key: 'status',
                header: 'Status',
                render: (row) => <AnnouncementStatusBadge status={row.status} />,
              },
              {
                key: 'author',
                header: 'Penulis',
                render: (row) => (
                  <span title={row.author_id}>{formatAuthorLabel(row.author_id)}</span>
                ),
              },
              {
                key: 'published_at',
                header: 'Dipublikasikan',
                render: (row) => (row.published_at ? formatDateTime(row.published_at) : '—'),
              },
              {
                key: 'actions',
                header: 'Aksi',
                render: (row) => (
                  <Link
                    to={`/announcements/${row.id}`}
                    className="inline-flex items-center justify-center rounded-full border border-[var(--color-line)] bg-[var(--color-panel)] px-3 py-1 text-sm font-semibold text-[var(--color-ink)] hover:bg-[var(--color-accent-soft)]"
                  >
                    Detail
                  </Link>
                ),
              },
            ]}
          />
          <Pagination meta={meta} onPageChange={setPage} disabled={isFetching} />
        </div>
      ) : null}

      <Modal
        open={createOpen}
        title="Buat Pengumuman"
        description="Pengumuman baru dimulai sebagai draft."
        onClose={() => setCreateOpen(false)}
        className="w-[min(100%-2rem,42rem)]"
      >
        <AnnouncementForm
          mode="create"
          submitLabel="Simpan draft"
          onCancel={() => setCreateOpen(false)}
          onSubmitCreate={async (payload) => {
            await createMutation.mutateAsync(payload)
            setCreateOpen(false)
            setFeedback('Pengumuman berhasil dibuat sebagai draft.')
          }}
        />
      </Modal>
    </div>
  )
}
