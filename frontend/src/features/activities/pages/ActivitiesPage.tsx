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
import { useDebouncedValue } from '../../../lib/hooks/useDebouncedValue'
import { formatDateTime } from '../../../lib/utils'
import { ActivityForm } from '../components/ActivityForm'
import { useActivities, useCreateActivity } from '../hooks'
import type { ActivityFilters } from '../types'
import { formatReminderSummary } from '../types'

const DEFAULT_PAGE_SIZE = 20

export function ActivitiesPage() {
  const [searchInput, setSearchInput] = useState('')
  const debouncedSearch = useDebouncedValue(searchInput, 300)
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [page, setPage] = useState(1)
  const [createOpen, setCreateOpen] = useState(false)
  const [feedback, setFeedback] = useState<string | null>(null)

  const filters = useMemo<ActivityFilters>(
    () => ({
      page,
      page_size: DEFAULT_PAGE_SIZE,
      search: debouncedSearch.trim(),
      from,
      to,
    }),
    [page, debouncedSearch, from, to],
  )

  const { data, isLoading, isError, isFetching, refetch } = useActivities(filters)
  const createMutation = useCreateActivity()

  const rows = data?.items ?? []
  const meta = data?.meta

  return (
    <div className="space-y-4">
      <PageHeader
        title="Kegiatan"
        description="Kelola agenda dan pengingat kegiatan RW."
        actions={
          <Button type="button" onClick={() => setCreateOpen(true)}>
            Buat Kegiatan
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
          placeholder="Nama atau deskripsi"
          value={searchInput}
          onChange={(event) => {
            setSearchInput(event.target.value)
            setPage(1)
          }}
        />
        <Input
          name="from"
          label="Dari tanggal"
          type="date"
          value={from}
          onChange={(event) => {
            setFrom(event.target.value)
            setPage(1)
          }}
        />
        <Input
          name="to"
          label="Sampai tanggal"
          type="date"
          value={to}
          onChange={(event) => {
            setTo(event.target.value)
            setPage(1)
          }}
        />
      </div>

      {isLoading ? <LoadingState label="Memuat kegiatan…" /> : null}

      {isError ? (
        <ErrorState
          title="Gagal memuat kegiatan."
          message="Periksa koneksi atau coba lagi."
          onRetry={() => {
            void refetch()
          }}
        />
      ) : null}

      {!isLoading && !isError && rows.length === 0 ? (
        <EmptyState
          title="Belum ada kegiatan"
          description="Buat kegiatan baru atau ubah filter pencarian."
          action={
            <Button type="button" onClick={() => setCreateOpen(true)}>
              Buat Kegiatan
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
                key: 'name',
                header: 'Kegiatan',
                render: (row) => (
                  <Link
                    className="font-medium text-[var(--color-accent)] hover:underline"
                    to={`/activities/${row.id}`}
                  >
                    {row.name}
                  </Link>
                ),
              },
              {
                key: 'date',
                header: 'Tanggal',
                render: (row) => formatDateTime(row.date),
              },
              {
                key: 'description',
                header: 'Deskripsi',
                className: 'max-w-xs truncate',
                render: (row) => row.description || '—',
              },
              {
                key: 'reminder',
                header: 'Reminder',
                render: (row) => formatReminderSummary(row),
              },
              {
                key: 'actions',
                header: 'Aksi',
                render: (row) => (
                  <Link
                    to={`/activities/${row.id}`}
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
        title="Buat Kegiatan"
        description="Isi detail kegiatan dan pengingat opsional."
        onClose={() => setCreateOpen(false)}
        className="w-[min(100%-2rem,40rem)]"
      >
        <ActivityForm
          mode="create"
          submitLabel="Simpan"
          onCancel={() => setCreateOpen(false)}
          onSubmitCreate={async (payload) => {
            await createMutation.mutateAsync(payload)
            setCreateOpen(false)
            setFeedback('Kegiatan berhasil ditambahkan.')
          }}
        />
      </Modal>
    </div>
  )
}
