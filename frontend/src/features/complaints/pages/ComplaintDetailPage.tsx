import { useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { Button } from '../../../components/ui/Button'
import { Card } from '../../../components/ui/Card'
import { ConfirmDialog, Modal } from '../../../components/ui/Modal'
import { ErrorState } from '../../../components/ui/ErrorState'
import { LoadingState } from '../../../components/ui/LoadingState'
import { PageHeader } from '../../../components/ui/PageHeader'
import { formatDateTime } from '../../../lib/utils'
import { ComplaintForm } from '../components/ComplaintForm'
import { ComplaintStatusBadge, ComplaintUrgencyBadge } from '../components/ComplaintBadges'
import { toComplaintErrorMessage } from '../errors'
import {
  useComplaint,
  useDeleteComplaint,
  useUpdateComplaint,
  useUpdateComplaintStatus,
} from '../hooks'
import type { ComplaintStatus } from '../types'
import { nextComplaintStatusActions } from '../types'

export function ComplaintDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data, isLoading, isError, refetch, error } = useComplaint(id)
  const updateMutation = useUpdateComplaint()
  const statusMutation = useUpdateComplaintStatus()
  const deleteMutation = useDeleteComplaint()

  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [feedback, setFeedback] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  if (isLoading) {
    return <LoadingState label="Memuat detail pengaduan…" />
  }

  if (isError || !data) {
    return (
      <div className="space-y-4">
        <PageHeader title="Detail Pengaduan" />
        <ErrorState
          title="Gagal memuat pengaduan."
          message={toComplaintErrorMessage(error)}
          onRetry={() => {
            void refetch()
          }}
        />
        <Button type="button" variant="secondary" onClick={() => void navigate('/complaints')}>
          Kembali
        </Button>
      </div>
    )
  }

  const complaint = data
  const statusActions = nextComplaintStatusActions(complaint.status)

  async function handleStatusChange(nextStatus: ComplaintStatus) {
    setActionError(null)
    try {
      await statusMutation.mutateAsync({
        id: complaint.id,
        payload: { status: nextStatus },
      })
      setFeedback('Status pengaduan berhasil diperbarui.')
    } catch (err) {
      setActionError(toComplaintErrorMessage(err))
    }
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title={complaint.ref}
        description="Detail dan tindak lanjut pengaduan."
        actions={
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="secondary" onClick={() => void navigate('/complaints')}>
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

      <div className="grid gap-4 lg:grid-cols-3">
        <Card title="Ringkasan" className="lg:col-span-1">
          <dl className="space-y-3 text-sm">
            <div>
              <dt className="text-[var(--color-muted)]">Status</dt>
              <dd className="mt-1">
                <ComplaintStatusBadge status={complaint.status} />
              </dd>
            </div>
            <div>
              <dt className="text-[var(--color-muted)]">Urgensi</dt>
              <dd className="mt-1">
                <ComplaintUrgencyBadge urgency={complaint.urgency} />
              </dd>
            </div>
            <div>
              <dt className="text-[var(--color-muted)]">Kategori</dt>
              <dd className="mt-1 font-medium">{complaint.category}</dd>
            </div>
            <div>
              <dt className="text-[var(--color-muted)]">Blok</dt>
              <dd className="mt-1 font-medium">{complaint.block || '—'}</dd>
            </div>
            <div>
              <dt className="text-[var(--color-muted)]">Diterima</dt>
              <dd className="mt-1">{formatDateTime(complaint.received_at)}</dd>
            </div>
            <div>
              <dt className="text-[var(--color-muted)]">Diperbarui</dt>
              <dd className="mt-1">{formatDateTime(complaint.updated_at)}</dd>
            </div>
          </dl>
        </Card>

        <Card title="Warga & Pesan" className="lg:col-span-2">
          <dl className="grid gap-4 sm:grid-cols-2">
            <div>
              <dt className="text-sm text-[var(--color-muted)]">Nama</dt>
              <dd className="mt-1 font-medium">{complaint.resident_name}</dd>
            </div>
            <div>
              <dt className="text-sm text-[var(--color-muted)]">Telepon</dt>
              <dd className="mt-1 font-medium">{complaint.phone}</dd>
            </div>
            <div className="sm:col-span-2">
              <dt className="text-sm text-[var(--color-muted)]">Tautan warga</dt>
              <dd className="mt-1 text-sm">
                {complaint.resident_id ? (
                  <span>
                    Tertaut ke warga{' '}
                    <Link className="text-[var(--color-accent)] hover:underline" to="/residents">
                      {complaint.resident_id}
                    </Link>
                  </span>
                ) : (
                  <span className="text-[var(--color-muted)]">
                    Pengaduan tanpa tautan warga (snapshot manual)
                  </span>
                )}
              </dd>
            </div>
            <div className="sm:col-span-2">
              <dt className="text-sm text-[var(--color-muted)]">Pesan</dt>
              <dd className="mt-1 whitespace-pre-wrap rounded-2xl border border-[var(--color-line)] bg-[var(--color-accent-soft)] px-3 py-3 text-sm">
                {complaint.message}
              </dd>
            </div>
          </dl>
        </Card>
      </div>

      <Card title="Tindak lanjut status">
        {statusActions.length === 0 ? (
          <p className="text-sm text-[var(--color-muted)]">
            Status terminal. Tidak ada perubahan status lanjutan yang tersedia di antarmuka ini.
          </p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {statusActions.map((action) => (
              <Button
                key={action.status}
                type="button"
                variant={action.status === 'DITOLAK' ? 'danger' : 'primary'}
                loading={statusMutation.isPending}
                onClick={() => {
                  void handleStatusChange(action.status)
                }}
              >
                {action.label}
              </Button>
            ))}
          </div>
        )}
      </Card>

      <Modal
        open={editOpen}
        title="Edit Pengaduan"
        description="Perbarui detail pengaduan. Status diubah melalui tindakan terpisah."
        onClose={() => setEditOpen(false)}
        className="w-[min(100%-2rem,40rem)]"
      >
        <ComplaintForm
          mode="edit"
          initial={complaint}
          submitLabel="Simpan perubahan"
          onCancel={() => setEditOpen(false)}
          onSubmitUpdate={async (payload) => {
            await updateMutation.mutateAsync({ id: complaint.id, payload })
            setEditOpen(false)
            setFeedback('Pengaduan berhasil diperbarui.')
          }}
        />
      </Modal>

      <ConfirmDialog
        open={deleteOpen}
        title="Hapus pengaduan?"
        message={`Pengaduan ${complaint.ref} akan dihapus secara permanen.`}
        confirmLabel="Hapus"
        loading={deleteMutation.isPending}
        onCancel={() => setDeleteOpen(false)}
        onConfirm={() => {
          void (async () => {
            try {
              await deleteMutation.mutateAsync(complaint.id)
              void navigate('/complaints', { replace: true })
            } catch (err) {
              setActionError(toComplaintErrorMessage(err))
              setDeleteOpen(false)
            }
          })()
        }}
      />
    </div>
  )
}
