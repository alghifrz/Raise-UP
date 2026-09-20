import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Button } from '../../../components/ui/Button'
import { Card } from '../../../components/ui/Card'
import { ConfirmDialog, Modal } from '../../../components/ui/Modal'
import { ErrorState } from '../../../components/ui/ErrorState'
import { LoadingState } from '../../../components/ui/LoadingState'
import { PageHeader } from '../../../components/ui/PageHeader'
import { formatDateTime, formatIdr } from '../../../lib/utils'
import { TransactionForm } from '../components/TransactionForm'
import { TransactionTypeBadge } from '../components/TransactionTypeBadge'
import { toFinanceErrorMessage } from '../errors'
import {
  useDeleteFinanceTransaction,
  useFinanceTransaction,
  useUpdateFinanceTransaction,
} from '../hooks'

export function FinanceTransactionDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data, isLoading, isError, refetch, error } = useFinanceTransaction(id)
  const updateMutation = useUpdateFinanceTransaction()
  const deleteMutation = useDeleteFinanceTransaction()

  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [feedback, setFeedback] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  if (isLoading) {
    return <LoadingState label="Memuat detail transaksi…" />
  }

  if (isError || !data) {
    return (
      <div className="space-y-4">
        <PageHeader title="Detail Transaksi" />
        <ErrorState
          title="Gagal memuat data keuangan."
          message={toFinanceErrorMessage(error)}
          onRetry={() => {
            void refetch()
          }}
        />
        <Button type="button" variant="secondary" onClick={() => void navigate('/finance')}>
          Kembali
        </Button>
      </div>
    )
  }

  const transaction = data

  return (
    <div className="space-y-4">
      <PageHeader
        title={transaction.title}
        description="Detail transaksi keuangan."
        actions={
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="secondary" onClick={() => void navigate('/finance')}>
              Kembali
            </Button>
            <Button type="button" variant="secondary" onClick={() => setEditOpen(true)}>
              Edit
            </Button>
            <Button type="button" variant="danger" onClick={() => setDeleteOpen(true)}>
              Hapus
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

      <Card title="Informasi transaksi">
        <dl className="grid gap-4 sm:grid-cols-2">
          <div>
            <dt className="text-sm text-[var(--color-muted)]">Jenis</dt>
            <dd className="mt-1">
              <TransactionTypeBadge type={transaction.type} />
            </dd>
          </div>
          <div>
            <dt className="text-sm text-[var(--color-muted)]">Nominal</dt>
            <dd className="mt-1 text-xl font-semibold tabular-nums">{formatIdr(transaction.amount)}</dd>
          </div>
          <div>
            <dt className="text-sm text-[var(--color-muted)]">Kategori</dt>
            <dd className="mt-1 font-medium">{transaction.category}</dd>
          </div>
          <div>
            <dt className="text-sm text-[var(--color-muted)]">Tanggal</dt>
            <dd className="mt-1">{formatDateTime(transaction.created_at)}</dd>
          </div>
          <div className="sm:col-span-2">
            <dt className="text-sm text-[var(--color-muted)]">Catatan</dt>
            <dd className="mt-1 whitespace-pre-wrap text-sm">{transaction.note || '—'}</dd>
          </div>
          <div>
            <dt className="text-sm text-[var(--color-muted)]">Dibuat</dt>
            <dd className="mt-1 text-sm">{formatDateTime(transaction.created_at)}</dd>
          </div>
          <div>
            <dt className="text-sm text-[var(--color-muted)]">Diperbarui</dt>
            <dd className="mt-1 text-sm">{formatDateTime(transaction.updated_at)}</dd>
          </div>
        </dl>
      </Card>

      <Modal open={editOpen} title="Edit Transaksi" onClose={() => setEditOpen(false)}>
        <TransactionForm
          key={transaction.id}
          mode="edit"
          initial={transaction}
          submitLabel="Simpan perubahan"
          onCancel={() => setEditOpen(false)}
          onSubmitUpdate={async (payload) => {
            await updateMutation.mutateAsync({ id: transaction.id, payload })
            setEditOpen(false)
            setFeedback('Transaksi berhasil diperbarui.')
          }}
        />
      </Modal>

      <ConfirmDialog
        open={deleteOpen}
        title="Hapus transaksi?"
        message={`Transaksi "${transaction.title}" akan dihapus secara permanen.`}
        confirmLabel="Hapus"
        loading={deleteMutation.isPending}
        onCancel={() => setDeleteOpen(false)}
        onConfirm={() => {
          void (async () => {
            try {
              await deleteMutation.mutateAsync(transaction.id)
              void navigate('/finance', { replace: true })
            } catch (err) {
              setActionError(toFinanceErrorMessage(err))
              setDeleteOpen(false)
            }
          })()
        }}
      />
    </div>
  )
}
