import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '../../../components/ui/Button'
import { DataTable } from '../../../components/ui/DataTable'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorState } from '../../../components/ui/ErrorState'
import { Input } from '../../../components/ui/Input'
import { LoadingState } from '../../../components/ui/LoadingState'
import { ConfirmDialog, Modal } from '../../../components/ui/Modal'
import { PageHeader } from '../../../components/ui/PageHeader'
import { Pagination } from '../../../components/ui/Pagination'
import { Select } from '../../../components/ui/Select'
import { formatDateTime, formatDuesPeriodLabel, formatIdr } from '../../../lib/utils'
import { PeriodForm } from '../components/PeriodForm'
import { toDuesErrorMessage } from '../errors'
import {
  useCreateDuesPeriod,
  useDeleteDuesPeriod,
  useDuesPeriods,
  useUpdateDuesPeriod,
} from '../hooks'
import type { DuesPeriod, DuesPeriodFilters } from '../types'
import { HALF_OPTIONS, MONTH_OPTIONS } from '../types'

const DEFAULT_PAGE_SIZE = 20

export function DuesPage() {
  const [year, setYear] = useState('')
  const [month, setMonth] = useState('')
  const [half, setHalf] = useState('')
  const [page, setPage] = useState(1)
  const [createOpen, setCreateOpen] = useState(false)
  const [editing, setEditing] = useState<DuesPeriod | null>(null)
  const [deleting, setDeleting] = useState<DuesPeriod | null>(null)
  const [feedback, setFeedback] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const filters = useMemo<DuesPeriodFilters>(
    () => ({
      page,
      page_size: DEFAULT_PAGE_SIZE,
      year,
      month,
      half,
    }),
    [page, year, month, half],
  )

  const { data, isLoading, isError, isFetching, refetch } = useDuesPeriods(filters)
  const createMutation = useCreateDuesPeriod()
  const updateMutation = useUpdateDuesPeriod()
  const deleteMutation = useDeleteDuesPeriod()

  const rows = data?.items ?? []
  const meta = data?.meta

  return (
    <div className="space-y-4">
      <PageHeader
        title="Iuran"
        description="Kelola periode dan pembayaran iuran warga."
        actions={
          <Button type="button" onClick={() => setCreateOpen(true)}>
            Tambah Periode
          </Button>
        }
      />

      {feedback ? (
        <p className="rounded-md border border-emerald-100 bg-emerald-50 px-3 py-2 text-sm text-emerald-800" role="status">
          {feedback}
        </p>
      ) : null}

      <section className="space-y-3">
        <h2 className="text-lg font-semibold text-[var(--color-ink)]">Periode Iuran</h2>

        <div className="grid gap-3 rounded-2xl border border-[var(--color-line)] bg-[var(--color-panel)] p-4 sm:grid-cols-3">
          <Input
            name="yearFilter"
            label="Tahun"
            inputMode="numeric"
            placeholder="2026"
            value={year}
            onChange={(event) => {
              setYear(event.target.value)
              setPage(1)
            }}
          />
          <Select
            name="monthFilter"
            label="Bulan"
            value={month}
            onChange={(event) => {
              setMonth(event.target.value)
              setPage(1)
            }}
            options={MONTH_OPTIONS}
            placeholder="Semua"
          />
          <Select
            name="halfFilter"
            label="Half"
            value={half}
            onChange={(event) => {
              setHalf(event.target.value)
              setPage(1)
            }}
            options={HALF_OPTIONS}
            placeholder="Semua"
          />
        </div>

        {isLoading ? <LoadingState label="Memuat periode iuran…" /> : null}

        {isError ? (
          <ErrorState
            title="Gagal memuat periode iuran."
            message="Periksa koneksi atau coba lagi."
            onRetry={() => {
              void refetch()
            }}
          />
        ) : null}

        {!isLoading && !isError && rows.length === 0 ? (
          <EmptyState
            title="Belum ada periode iuran."
            description="Tambahkan periode baru untuk mulai mencatat pembayaran."
            action={
              <Button type="button" onClick={() => setCreateOpen(true)}>
                Tambah Periode
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
                  key: 'period',
                  header: 'Periode',
                  render: (row) => (
                    <Link
                      className="font-medium text-[var(--color-accent)] hover:underline"
                      to={`/dues/periods/${row.id}`}
                    >
                      {formatDuesPeriodLabel(row.year, row.month, row.half)}
                    </Link>
                  ),
                },
                {
                  key: 'amount',
                  header: 'Nominal',
                  render: (row) => (
                    <span className="tabular-nums font-medium">{formatIdr(row.amount)}</span>
                  ),
                },
                {
                  key: 'created_at',
                  header: 'Dibuat',
                  render: (row) => formatDateTime(row.created_at),
                },
                {
                  key: 'actions',
                  header: 'Aksi',
                  render: (row) => (
                    <div className="flex gap-2">
                      <Link
                        to={`/dues/periods/${row.id}`}
                        className="inline-flex items-center justify-center rounded-full border border-[var(--color-line)] bg-[var(--color-panel)] px-3 py-1 text-sm font-semibold hover:bg-[var(--color-accent-soft)]"
                      >
                        Detail
                      </Link>
                      <Button
                        type="button"
                        variant="secondary"
                        className="px-2 py-1"
                        onClick={() => setEditing(row)}
                      >
                        Edit
                      </Button>
                      <Button
                        type="button"
                        variant="danger"
                        className="px-2 py-1"
                        onClick={() => {
                          setDeleteError(null)
                          setDeleting(row)
                        }}
                      >
                        Hapus
                      </Button>
                    </div>
                  ),
                },
              ]}
            />
            <Pagination meta={meta} onPageChange={setPage} disabled={isFetching} />
          </div>
        ) : null}
      </section>

      <Modal
        open={createOpen}
        title="Tambah Periode"
        description="Buat periode iuran baru (tahun + bulan + half)."
        onClose={() => setCreateOpen(false)}
      >
        <PeriodForm
          mode="create"
          submitLabel="Simpan"
          onCancel={() => setCreateOpen(false)}
          onSubmitCreate={async (payload) => {
            await createMutation.mutateAsync(payload)
            setCreateOpen(false)
            setFeedback('Periode iuran berhasil ditambahkan.')
          }}
        />
      </Modal>

      <Modal open={Boolean(editing)} title="Edit Periode" onClose={() => setEditing(null)}>
        {editing ? (
          <PeriodForm
            key={editing.id}
            mode="edit"
            initial={editing}
            submitLabel="Simpan perubahan"
            onCancel={() => setEditing(null)}
            onSubmitUpdate={async (payload) => {
              await updateMutation.mutateAsync({ id: editing.id, payload })
              setEditing(null)
              setFeedback('Periode iuran berhasil diperbarui.')
            }}
          />
        ) : null}
      </Modal>

      <ConfirmDialog
        open={Boolean(deleting)}
        title="Hapus periode iuran?"
        message="Menghapus periode ini juga akan menghapus data pembayaran terkait. Tindakan ini tidak dapat dibatalkan."
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
              setFeedback('Periode iuran berhasil dihapus.')
            } catch (error) {
              setDeleteError(toDuesErrorMessage(error))
            }
          })()
        }}
      />

      {deleteError ? (
        <p
          className="rounded-2xl border border-red-100 bg-[var(--color-danger-soft)] px-3 py-2 text-sm text-[var(--color-danger)]"
          role="alert"
        >
          {deleteError}
        </p>
      ) : null}
    </div>
  )
}
