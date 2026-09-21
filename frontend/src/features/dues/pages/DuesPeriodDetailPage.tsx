import { useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
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
import { formatDateTime, formatDuesPeriodLabel, formatIdr } from '../../../lib/utils'
import { getDuesPayment } from '../api'
import { DuesPaymentStatusBadge } from '../components/DuesPaymentStatusBadge'
import { PaymentForm } from '../components/PaymentForm'
import { PeriodForm } from '../components/PeriodForm'
import { toDuesErrorMessage } from '../errors'
import {
  useCreateDuesPayment,
  useDeleteDuesPayment,
  useDeleteDuesPeriod,
  useDuesPeriod,
  useDuesPeriodStatus,
  useDuesPeriodSummary,
  useSendUnpaidDuesReminders,
  useUpdateDuesPayment,
  useUpdateDuesPeriod,
} from '../hooks'
import type {
  DuesPayment,
  DuesPaymentStatusValue,
  DuesPeriodStatusFilters,
  ResidentPaymentStatus,
} from '../types'
import { DUES_STATUS_OPTIONS } from '../types'

const DEFAULT_PAGE_SIZE = 20

export function DuesPeriodDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const periodQuery = useDuesPeriod(id)
  const summaryQuery = useDuesPeriodSummary(id)

  const [searchInput, setSearchInput] = useState('')
  const debouncedSearch = useDebouncedValue(searchInput, 300)
  const [status, setStatus] = useState<DuesPaymentStatusValue | ''>('')
  const [page, setPage] = useState(1)

  const statusFilters = useMemo<DuesPeriodStatusFilters>(
    () => ({
      page,
      page_size: DEFAULT_PAGE_SIZE,
      status,
      search: debouncedSearch.trim(),
    }),
    [page, status, debouncedSearch],
  )

  const statusQuery = useDuesPeriodStatus(id, statusFilters)
  const updatePeriodMutation = useUpdateDuesPeriod()
  const deletePeriodMutation = useDeleteDuesPeriod()
  const createPaymentMutation = useCreateDuesPayment()
  const updatePaymentMutation = useUpdateDuesPayment()
  const deletePaymentMutation = useDeleteDuesPayment()
  const reminderMutation = useSendUnpaidDuesReminders()

  const [editPeriodOpen, setEditPeriodOpen] = useState(false)
  const [deletePeriodOpen, setDeletePeriodOpen] = useState(false)
  const [paymentOpen, setPaymentOpen] = useState(false)
  const [editingPayment, setEditingPayment] = useState<{
    payment: DuesPayment
  } | null>(null)
  const [deletingPaymentId, setDeletingPaymentId] = useState<string | null>(null)
  const [reminderOpen, setReminderOpen] = useState(false)
  const [feedback, setFeedback] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  if (periodQuery.isLoading) {
    return <LoadingState label="Memuat detail periode…" />
  }

  if (periodQuery.isError || !periodQuery.data) {
    return (
      <div className="space-y-4">
        <PageHeader title="Detail Periode Iuran" />
        <ErrorState
          title="Gagal memuat periode iuran."
          message={toDuesErrorMessage(periodQuery.error)}
          onRetry={() => {
            void periodQuery.refetch()
          }}
        />
        <Button type="button" variant="secondary" onClick={() => void navigate('/dues')}>
          Kembali
        </Button>
      </div>
    )
  }

  const period = periodQuery.data
  const summary = summaryQuery.data
  const statusRows = statusQuery.data?.items ?? []
  const statusMeta = statusQuery.data?.meta

  async function openEditPayment(row: ResidentPaymentStatus) {
    if (!row.payment_id) {
      return
    }
    setActionError(null)
    try {
      const payment = await getDuesPayment(row.payment_id)
      setEditingPayment({ payment })
    } catch (error) {
      setActionError(toDuesErrorMessage(error))
    }
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title={formatDuesPeriodLabel(period.year, period.month, period.half)}
        description="Ringkasan dan status pembayaran periode."
        actions={
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="secondary" onClick={() => void navigate('/dues')}>
              Kembali
            </Button>
            <Button type="button" variant="secondary" onClick={() => setEditPeriodOpen(true)}>
              Edit Periode
            </Button>
            <Button type="button" onClick={() => setPaymentOpen(true)}>
              Tambah Pembayaran
            </Button>
            <Button type="button" variant="danger" onClick={() => setDeletePeriodOpen(true)}>
              Hapus Periode
            </Button>
          </div>
        }
      />

      {feedback ? (
        <p className="rounded-md border border-emerald-100 bg-emerald-50 px-3 py-2 text-sm text-emerald-800" role="status">
          {feedback}
        </p>
      ) : null}

      {actionError ? (
        <p
          className="rounded-2xl border border-red-100 bg-[var(--color-danger-soft)] px-3 py-2 text-sm text-[var(--color-danger)]"
          role="alert"
        >
          {actionError}
        </p>
      ) : null}

      <Card title="Informasi periode">
        <dl className="grid gap-3 sm:grid-cols-4 text-sm">
          <div>
            <dt className="text-[var(--color-muted)]">Periode</dt>
            <dd className="mt-1 font-medium">
              {formatDuesPeriodLabel(period.year, period.month, period.half)}
            </dd>
          </div>
          <div>
            <dt className="text-[var(--color-muted)]">Nominal</dt>
            <dd className="mt-1 font-medium tabular-nums">{formatIdr(period.amount)}</dd>
          </div>
          <div>
            <dt className="text-[var(--color-muted)]">Dibuat</dt>
            <dd className="mt-1">{formatDateTime(period.created_at)}</dd>
          </div>
        </dl>
      </Card>

      {summaryQuery.isLoading ? <LoadingState label="Memuat ringkasan…" className="min-h-24" /> : null}
      {summaryQuery.isError ? (
        <ErrorState
          title="Gagal memuat ringkasan periode."
          onRetry={() => {
            void summaryQuery.refetch()
          }}
        />
      ) : null}
      {summary ? (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <Card title="Warga">
            <p className="text-2xl font-semibold tabular-nums">{summary.resident_count}</p>
            <p className="mt-1 text-xs text-[var(--color-muted)]">
              Lunas {summary.paid_count} · Belum {summary.unpaid_count}
            </p>
          </Card>
          <Card title="Target">
            <p className="text-2xl font-semibold tabular-nums">{formatIdr(summary.expected_total)}</p>
            <p className="mt-1 text-xs text-[var(--color-muted)]">
              Per warga {formatIdr(summary.expected_amount)}
            </p>
          </Card>
          <Card title="Terkumpul">
            <p className="text-2xl font-semibold tabular-nums text-emerald-700">
              {formatIdr(summary.collected_total)}
            </p>
            <p className="mt-1 text-xs text-[var(--color-muted)]">
              Tunggakan {formatIdr(summary.outstanding_total)}
            </p>
          </Card>
        </div>
      ) : null}

      <section className="space-y-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-lg font-semibold text-[var(--color-ink)]">Status Pembayaran</h2>
          <Button
            type="button"
            variant="secondary"
            onClick={() => setReminderOpen(true)}
            disabled={summary?.unpaid_count === 0}
          >
            Kirim Reminder
          </Button>
        </div>

        <div className="grid gap-3 rounded-2xl border border-[var(--color-line)] bg-[var(--color-panel)] p-4 sm:grid-cols-2">
          <Input
            name="statusSearch"
            label="Cari warga"
            placeholder="Nama warga"
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
              setStatus(event.target.value as DuesPaymentStatusValue | '')
              setPage(1)
            }}
            options={DUES_STATUS_OPTIONS}
            placeholder="Semua"
          />
        </div>

        {statusQuery.isLoading ? <LoadingState label="Memuat status pembayaran…" /> : null}

        {statusQuery.isError ? (
          <ErrorState
            title="Gagal memuat status pembayaran."
            onRetry={() => {
              void statusQuery.refetch()
            }}
          />
        ) : null}

        {!statusQuery.isLoading && !statusQuery.isError && statusRows.length === 0 ? (
          <EmptyState
            title="Tidak ada warga yang cocok dengan pencarian."
            description="Ubah filter atau pastikan data warga sudah tersedia."
          />
        ) : null}

        {!statusQuery.isLoading && !statusQuery.isError && statusRows.length > 0 && statusMeta ? (
          <div className="space-y-3">
            <DataTable
              rows={statusRows}
              rowKey={(row) => row.resident_id}
              columns={[
                {
                  key: 'resident',
                  header: 'Warga',
                  render: (row) => row.resident_name,
                },
                {
                  key: 'phone',
                  header: 'No. Telepon',
                  render: (row) => row.phone,
                },
                {
                  key: 'status',
                  header: 'Status',
                  render: (row) => <DuesPaymentStatusBadge status={row.status} />,
                },
                {
                  key: 'amount',
                  header: 'Nominal',
                  render: (row) =>
                    row.amount != null ? (
                      <span className="tabular-nums">{formatIdr(row.amount)}</span>
                    ) : (
                      '—'
                    ),
                },
                {
                  key: 'paid_at',
                  header: 'Tanggal Bayar',
                  render: (row) => (row.paid_at ? formatDateTime(row.paid_at) : '—'),
                },
                {
                  key: 'actions',
                  header: 'Aksi',
                  render: (row) =>
                    row.status === 'PAID' && row.payment_id ? (
                      <div className="flex gap-2">
                        <Button
                          type="button"
                          variant="secondary"
                          className="px-2 py-1"
                          onClick={() => {
                            void openEditPayment(row)
                          }}
                        >
                          Edit
                        </Button>
                        <Button
                          type="button"
                          variant="danger"
                          className="px-2 py-1"
                          onClick={() => {
                            if (row.payment_id) {
                              setDeletingPaymentId(row.payment_id)
                            }
                          }}
                        >
                          Hapus
                        </Button>
                      </div>
                    ) : (
                      <Button
                        type="button"
                        variant="secondary"
                        className="px-2 py-1"
                        onClick={() => setPaymentOpen(true)}
                      >
                        Bayar
                      </Button>
                    ),
                },
              ]}
            />
            <Pagination
              meta={statusMeta}
              onPageChange={setPage}
              disabled={statusQuery.isFetching}
            />
          </div>
        ) : null}
      </section>

      <Modal open={editPeriodOpen} title="Edit Periode" onClose={() => setEditPeriodOpen(false)}>
        <PeriodForm
          key={period.id}
          mode="edit"
          initial={period}
          submitLabel="Simpan perubahan"
          onCancel={() => setEditPeriodOpen(false)}
          onSubmitUpdate={async (payload) => {
            await updatePeriodMutation.mutateAsync({ id: period.id, payload })
            setEditPeriodOpen(false)
            setFeedback('Periode berhasil diperbarui.')
          }}
        />
      </Modal>

      <Modal
        open={paymentOpen}
        title="Tambah Pembayaran"
        description="Nominal harus sama dengan nominal periode."
        onClose={() => setPaymentOpen(false)}
      >
        <PaymentForm
          mode="create"
          period={period}
          submitLabel="Simpan pembayaran"
          onCancel={() => setPaymentOpen(false)}
          onSubmitCreate={async (payload) => {
            await createPaymentMutation.mutateAsync(payload)
            setPaymentOpen(false)
            setFeedback('Pembayaran berhasil ditambahkan.')
          }}
        />
      </Modal>

      <Modal
        open={Boolean(editingPayment)}
        title="Edit Pembayaran"
        onClose={() => setEditingPayment(null)}
      >
        {editingPayment ? (
          <PaymentForm
            key={editingPayment.payment.id}
            mode="edit"
            period={period}
            initial={editingPayment.payment}
            submitLabel="Simpan perubahan"
            onCancel={() => setEditingPayment(null)}
            onSubmitUpdate={async (payload) => {
              await updatePaymentMutation.mutateAsync({
                id: editingPayment.payment.id,
                payload,
              })
              setEditingPayment(null)
              setFeedback('Pembayaran berhasil diperbarui.')
            }}
          />
        ) : null}
      </Modal>

      <ConfirmDialog
        open={reminderOpen}
        title="Kirim reminder iuran?"
        message={`Reminder WhatsApp akan dikirim ke ${summary?.unpaid_count ?? 0} warga yang belum membayar periode ini.`}
        confirmLabel="Kirim reminder"
        loading={reminderMutation.isPending}
        onCancel={() => setReminderOpen(false)}
        onConfirm={() => {
          void (async () => {
            setActionError(null)
            try {
              const result = await reminderMutation.mutateAsync(period.id)
              setReminderOpen(false)
              if (result.total === 0) {
                setFeedback('Semua warga sudah membayar. Tidak ada reminder yang dikirim.')
              } else if (result.failed === 0) {
                setFeedback(`Reminder berhasil dikirim ke ${result.sent} warga.`)
              } else {
                setFeedback(
                  `Reminder terkirim ${result.sent} dari ${result.total} warga (${result.failed} gagal).`,
                )
              }
            } catch (error) {
              setActionError(toDuesErrorMessage(error))
              setReminderOpen(false)
            }
          })()
        }}
      />

      <ConfirmDialog
        open={deletePeriodOpen}
        title="Hapus periode iuran?"
        message="Menghapus periode ini juga akan menghapus data pembayaran terkait. Tindakan ini tidak dapat dibatalkan."
        confirmLabel="Hapus"
        loading={deletePeriodMutation.isPending}
        onCancel={() => setDeletePeriodOpen(false)}
        onConfirm={() => {
          void (async () => {
            try {
              await deletePeriodMutation.mutateAsync(period.id)
              void navigate('/dues', { replace: true })
            } catch (error) {
              setActionError(toDuesErrorMessage(error))
              setDeletePeriodOpen(false)
            }
          })()
        }}
      />

      <ConfirmDialog
        open={Boolean(deletingPaymentId)}
        title="Hapus pembayaran?"
        message="Pembayaran ini akan dihapus dan status warga kembali menjadi belum bayar."
        confirmLabel="Hapus"
        loading={deletePaymentMutation.isPending}
        onCancel={() => setDeletingPaymentId(null)}
        onConfirm={() => {
          if (!deletingPaymentId) {
            return
          }
          void (async () => {
            try {
              await deletePaymentMutation.mutateAsync(deletingPaymentId)
              setDeletingPaymentId(null)
              setFeedback('Pembayaran berhasil dihapus.')
            } catch (error) {
              setActionError(toDuesErrorMessage(error))
              setDeletingPaymentId(null)
            }
          })()
        }}
      />
    </div>
  )
}
