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
import { ComplaintForm } from '../components/ComplaintForm'
import { ComplaintStatusBadge, ComplaintUrgencyBadge } from '../components/ComplaintBadges'
import { useComplaints, useCreateComplaint } from '../hooks'
import type { ComplaintFilters, ComplaintStatus, ComplaintUrgency } from '../types'
import { COMPLAINT_STATUS_OPTIONS, COMPLAINT_URGENCY_OPTIONS } from '../types'

const DEFAULT_PAGE_SIZE = 20

export function ComplaintsPage() {
  const [searchInput, setSearchInput] = useState('')
  const debouncedSearch = useDebouncedValue(searchInput, 300)
  const [status, setStatus] = useState<ComplaintStatus | ''>('')
  const [urgency, setUrgency] = useState<ComplaintUrgency | ''>('')
  const [categoryInput, setCategoryInput] = useState('')
  const debouncedCategory = useDebouncedValue(categoryInput, 300)
  const [page, setPage] = useState(1)
  const [createOpen, setCreateOpen] = useState(false)
  const [feedback, setFeedback] = useState<string | null>(null)

  const filters = useMemo<ComplaintFilters>(
    () => ({
      page,
      page_size: DEFAULT_PAGE_SIZE,
      search: debouncedSearch.trim(),
      status,
      urgency,
      category: debouncedCategory.trim(),
    }),
    [page, debouncedSearch, status, urgency, debouncedCategory],
  )

  const { data, isLoading, isError, isFetching, refetch } = useComplaints(filters)
  const createMutation = useCreateComplaint()

  const rows = data?.items ?? []
  const meta = data?.meta

  return (
    <div className="space-y-4">
      <PageHeader
        title="Pengaduan"
        description="Kelola pengaduan warga."
        actions={
          <Button type="button" onClick={() => setCreateOpen(true)}>
            Tambah Pengaduan
          </Button>
        }
      />

      {feedback ? (
        <p className="rounded-md border border-emerald-100 bg-emerald-50 px-3 py-2 text-sm text-emerald-800" role="status">
          {feedback}
        </p>
      ) : null}

      <div className="grid gap-3 rounded-2xl border border-[var(--color-line)] bg-[var(--color-panel)] p-4 sm:grid-cols-2 lg:grid-cols-4">
        <Input
          name="search"
          label="Cari"
          placeholder="Ref, nama, atau pesan"
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
            setStatus(event.target.value as ComplaintStatus | '')
            setPage(1)
          }}
          options={COMPLAINT_STATUS_OPTIONS}
          placeholder="Semua"
        />
        <Select
          name="urgencyFilter"
          label="Urgensi"
          value={urgency}
          onChange={(event) => {
            setUrgency(event.target.value as ComplaintUrgency | '')
            setPage(1)
          }}
          options={COMPLAINT_URGENCY_OPTIONS}
          placeholder="Semua"
        />
        <Input
          name="categoryFilter"
          label="Kategori"
          placeholder="Mis. Kebersihan"
          value={categoryInput}
          onChange={(event) => {
            setCategoryInput(event.target.value)
            setPage(1)
          }}
        />
      </div>

      {isLoading ? <LoadingState label="Memuat pengaduan…" /> : null}

      {isError ? (
        <ErrorState
          title="Gagal memuat pengaduan."
          message="Periksa koneksi atau coba lagi."
          onRetry={() => {
            void refetch()
          }}
        />
      ) : null}

      {!isLoading && !isError && rows.length === 0 ? (
        <EmptyState
          title="Belum ada pengaduan"
          description="Buat pengaduan baru atau ubah filter pencarian."
          action={
            <Button type="button" onClick={() => setCreateOpen(true)}>
              Tambah Pengaduan
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
                key: 'ref',
                header: 'Ref',
                render: (row) => (
                  <Link
                    className="font-medium text-[var(--color-accent)] hover:underline"
                    to={`/complaints/${row.id}`}
                  >
                    {row.ref}
                  </Link>
                ),
              },
              {
                key: 'resident',
                header: 'Warga',
                render: (row) => row.resident_name,
              },
              {
                key: 'category',
                header: 'Kategori',
                render: (row) => row.category,
              },
              {
                key: 'urgency',
                header: 'Urgensi',
                render: (row) => <ComplaintUrgencyBadge urgency={row.urgency} />,
              },
              {
                key: 'status',
                header: 'Status',
                render: (row) => <ComplaintStatusBadge status={row.status} />,
              },
              {
                key: 'received_at',
                header: 'Diterima',
                render: (row) => formatDateTime(row.received_at),
              },
              {
                key: 'actions',
                header: 'Aksi',
                render: (row) => (
                  <Link
                    to={`/complaints/${row.id}`}
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
        title="Tambah Pengaduan"
        description="Buat pengaduan baru untuk warga."
        onClose={() => setCreateOpen(false)}
        className="w-[min(100%-2rem,40rem)]"
      >
        <ComplaintForm
          mode="create"
          submitLabel="Simpan"
          onCancel={() => setCreateOpen(false)}
          onSubmitCreate={async (payload) => {
            await createMutation.mutateAsync(payload)
            setCreateOpen(false)
            setFeedback('Pengaduan berhasil ditambahkan.')
          }}
        />
      </Modal>
    </div>
  )
}
