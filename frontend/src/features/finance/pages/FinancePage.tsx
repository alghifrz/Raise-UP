import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '../../../components/ui/Button'
import { Card } from '../../../components/ui/Card'
import { DataTable } from '../../../components/ui/DataTable'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorState } from '../../../components/ui/ErrorState'
import { Input } from '../../../components/ui/Input'
import { LoadingState } from '../../../components/ui/LoadingState'
import { ConfirmDialog, Modal } from '../../../components/ui/Modal'
import { PageHeader } from '../../../components/ui/PageHeader'
import { Pagination } from '../../../components/ui/Pagination'
import { Select } from '../../../components/ui/Select'
import { useDebouncedValue } from '../../../lib/hooks/useDebouncedValue'
import { formatDateTime, formatIdr } from '../../../lib/utils'
import { TransactionForm } from '../components/TransactionForm'
import { TransactionTypeBadge } from '../components/TransactionTypeBadge'
import { toFinanceErrorMessage } from '../errors'
import {
  useCreateFinanceTransaction,
  useDeleteFinanceTransaction,
  useFinanceSummary,
  useFinanceTransactions,
  useUpdateFinanceTransaction,
} from '../hooks'
import type { FinanceFilters, FinanceTransaction, TransactionType } from '../types'
import { TRANSACTION_TYPE_OPTIONS } from '../types'

const DEFAULT_PAGE_SIZE = 20

export function FinancePage() {
  const [searchInput, setSearchInput] = useState('')
  const debouncedSearch = useDebouncedValue(searchInput, 300)
  const [type, setType] = useState<TransactionType | ''>('')
  const [categoryInput, setCategoryInput] = useState('')
  const debouncedCategory = useDebouncedValue(categoryInput, 300)
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [page, setPage] = useState(1)
  const [createOpen, setCreateOpen] = useState(false)
  const [editing, setEditing] = useState<FinanceTransaction | null>(null)
  const [deleting, setDeleting] = useState<FinanceTransaction | null>(null)
  const [feedback, setFeedback] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const filters = useMemo<FinanceFilters>(
    () => ({
      page,
      page_size: DEFAULT_PAGE_SIZE,
      search: debouncedSearch.trim(),
      type,
      category: debouncedCategory.trim(),
      from,
      to,
    }),
    [page, debouncedSearch, type, debouncedCategory, from, to],
  )

  const summaryFilters = useMemo(
    () => ({
      from,
      to,
    }),
    [from, to],
  )

  const transactionsQuery = useFinanceTransactions(filters)
  const summaryQuery = useFinanceSummary(summaryFilters)
  const createMutation = useCreateFinanceTransaction()
  const updateMutation = useUpdateFinanceTransaction()
  const deleteMutation = useDeleteFinanceTransaction()

  const rows = transactionsQuery.data?.items ?? []
  const meta = transactionsQuery.data?.meta
  const summary = summaryQuery.data

  return (
    <div className="space-y-4">
      <PageHeader
        title="Keuangan"
        description="Kelola transaksi pemasukan dan pengeluaran."
        actions={
          <Button type="button" onClick={() => setCreateOpen(true)}>
            Tambah Transaksi
          </Button>
        }
      />

      {feedback ? (
        <p className="rounded-md border border-emerald-100 bg-emerald-50 px-3 py-2 text-sm text-emerald-800" role="status">
          {feedback}
        </p>
      ) : null}

      {summaryQuery.isLoading ? <LoadingState label="Memuat ringkasan keuangan…" className="min-h-24" /> : null}
      {summaryQuery.isError ? (
        <ErrorState
          title="Gagal memuat data keuangan."
          message="Ringkasan tidak dapat diambil."
          onRetry={() => {
            void summaryQuery.refetch()
          }}
        />
      ) : null}
      {summary ? (
        <div className="grid gap-3 sm:grid-cols-3">
          <Card title="Total Pemasukan">
            <p className="text-2xl font-semibold tabular-nums text-emerald-700">
              {formatIdr(summary.total_income)}
            </p>
          </Card>
          <Card title="Total Pengeluaran">
            <p className="text-2xl font-semibold tabular-nums text-red-700">
              {formatIdr(summary.total_expense)}
            </p>
          </Card>
          <Card title="Saldo">
            <p className="text-2xl font-semibold tabular-nums text-[var(--color-ink)]">
              {formatIdr(summary.balance)}
            </p>
          </Card>
        </div>
      ) : null}

      <div className="grid gap-3 rounded-2xl border border-[var(--color-line)] bg-[var(--color-panel)] p-4 sm:grid-cols-2 lg:grid-cols-5">
        <Input
          name="search"
          label="Cari"
          placeholder="Judul atau catatan"
          value={searchInput}
          onChange={(event) => {
            setSearchInput(event.target.value)
            setPage(1)
          }}
        />
        <Select
          name="typeFilter"
          label="Jenis"
          value={type}
          onChange={(event) => {
            setType(event.target.value as TransactionType | '')
            setPage(1)
          }}
          options={TRANSACTION_TYPE_OPTIONS}
          placeholder="Semua"
        />
        <Input
          name="categoryFilter"
          label="Kategori"
          placeholder="Mis. Operasional"
          value={categoryInput}
          onChange={(event) => {
            setCategoryInput(event.target.value)
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

      {transactionsQuery.isLoading ? <LoadingState label="Memuat transaksi…" /> : null}

      {transactionsQuery.isError ? (
        <ErrorState
          title="Gagal memuat data keuangan."
          message="Daftar transaksi tidak dapat diambil."
          onRetry={() => {
            void transactionsQuery.refetch()
          }}
        />
      ) : null}

      {!transactionsQuery.isLoading && !transactionsQuery.isError && rows.length === 0 ? (
        <EmptyState
          title="Belum ada transaksi keuangan."
          description="Tambahkan transaksi baru atau ubah filter pencarian."
          action={
            <Button type="button" onClick={() => setCreateOpen(true)}>
              Tambah Transaksi
            </Button>
          }
        />
      ) : null}

      {!transactionsQuery.isLoading && !transactionsQuery.isError && rows.length > 0 && meta ? (
        <div className="space-y-3">
          {transactionsQuery.isFetching ? (
            <p className="text-xs text-[var(--color-muted)]" aria-live="polite">
              Memperbarui…
            </p>
          ) : null}
          <DataTable
            rows={rows}
            rowKey={(row) => row.id}
            columns={[
              {
                key: 'created_at',
                header: 'Tanggal',
                render: (row) => formatDateTime(row.created_at),
              },
              {
                key: 'type',
                header: 'Jenis',
                render: (row) => <TransactionTypeBadge type={row.type} />,
              },
              {
                key: 'title',
                header: 'Judul',
                render: (row) => (
                  <Link
                    className="font-medium text-[var(--color-accent)] hover:underline"
                    to={`/finance/${row.id}`}
                  >
                    {row.title}
                  </Link>
                ),
              },
              {
                key: 'category',
                header: 'Kategori',
                render: (row) => row.category,
              },
              {
                key: 'amount',
                header: 'Nominal',
                render: (row) => (
                  <span className="tabular-nums font-medium">{formatIdr(row.amount)}</span>
                ),
              },
              {
                key: 'note',
                header: 'Catatan',
                className: 'max-w-xs truncate',
                render: (row) => row.note || '—',
              },
              {
                key: 'actions',
                header: 'Aksi',
                render: (row) => (
                  <div className="flex gap-2">
                    <Link
                      to={`/finance/${row.id}`}
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
          <Pagination
            meta={meta}
            onPageChange={setPage}
            disabled={transactionsQuery.isFetching}
          />
        </div>
      ) : null}

      <Modal
        open={createOpen}
        title="Tambah Transaksi"
        description="Catat pemasukan atau pengeluaran baru."
        onClose={() => setCreateOpen(false)}
      >
        <TransactionForm
          mode="create"
          submitLabel="Simpan"
          onCancel={() => setCreateOpen(false)}
          onSubmitCreate={async (payload) => {
            await createMutation.mutateAsync(payload)
            setCreateOpen(false)
            setFeedback('Transaksi berhasil ditambahkan.')
          }}
        />
      </Modal>

      <Modal
        open={Boolean(editing)}
        title="Edit Transaksi"
        onClose={() => setEditing(null)}
      >
        {editing ? (
          <TransactionForm
            key={editing.id}
            mode="edit"
            initial={editing}
            submitLabel="Simpan perubahan"
            onCancel={() => setEditing(null)}
            onSubmitUpdate={async (payload) => {
              await updateMutation.mutateAsync({ id: editing.id, payload })
              setEditing(null)
              setFeedback('Transaksi berhasil diperbarui.')
            }}
          />
        ) : null}
      </Modal>

      <ConfirmDialog
        open={Boolean(deleting)}
        title="Hapus transaksi?"
        message={deleting ? `Transaksi "${deleting.title}" akan dihapus secara permanen.` : ''}
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
              setFeedback('Transaksi berhasil dihapus.')
            } catch (error) {
              setDeleteError(toFinanceErrorMessage(error))
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
