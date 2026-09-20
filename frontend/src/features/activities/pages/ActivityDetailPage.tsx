import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Button } from '../../../components/ui/Button'
import { Card } from '../../../components/ui/Card'
import { ConfirmDialog, Modal } from '../../../components/ui/Modal'
import { ErrorState } from '../../../components/ui/ErrorState'
import { LoadingState } from '../../../components/ui/LoadingState'
import { PageHeader } from '../../../components/ui/PageHeader'
import { formatDateTime } from '../../../lib/utils'
import { ActivityForm } from '../components/ActivityForm'
import { toActivityErrorMessage } from '../errors'
import { useActivity, useDeleteActivity, useUpdateActivity } from '../hooks'
import { formatReminderSummary } from '../types'

export function ActivityDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data, isLoading, isError, refetch, error } = useActivity(id)
  const updateMutation = useUpdateActivity()
  const deleteMutation = useDeleteActivity()

  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [feedback, setFeedback] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  if (isLoading) {
    return <LoadingState label="Memuat detail kegiatan…" />
  }

  if (isError || !data) {
    return (
      <div className="space-y-4">
        <PageHeader title="Detail Kegiatan" />
        <ErrorState
          title="Gagal memuat kegiatan."
          message={toActivityErrorMessage(error)}
          onRetry={() => {
            void refetch()
          }}
        />
        <Button type="button" variant="secondary" onClick={() => void navigate('/activities')}>
          Kembali
        </Button>
      </div>
    )
  }

  const activity = data

  return (
    <div className="space-y-4">
      <PageHeader
        title={activity.name}
        description="Detail kegiatan dan konfigurasi reminder."
        actions={
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="secondary" onClick={() => void navigate('/activities')}>
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
              <dt className="text-[var(--color-muted)]">Tanggal</dt>
              <dd className="mt-1 font-medium">{formatDateTime(activity.date)}</dd>
            </div>
            <div>
              <dt className="text-[var(--color-muted)]">Reminder</dt>
              <dd className="mt-1 font-medium">{formatReminderSummary(activity)}</dd>
            </div>
            <div>
              <dt className="text-[var(--color-muted)]">Dibuat</dt>
              <dd className="mt-1">{formatDateTime(activity.created_at)}</dd>
            </div>
            <div>
              <dt className="text-[var(--color-muted)]">Diperbarui</dt>
              <dd className="mt-1">{formatDateTime(activity.updated_at)}</dd>
            </div>
          </dl>
        </Card>

        <Card title="Detail" className="lg:col-span-2">
          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-medium text-[var(--color-muted)]">Deskripsi</h3>
              <p className="mt-1 whitespace-pre-wrap text-sm text-[var(--color-ink)]">
                {activity.description || '—'}
              </p>
            </div>
            <div className="rounded-2xl border border-[var(--color-line)] bg-[var(--color-accent-soft)] px-3 py-3">
              <h3 className="text-sm font-medium text-[var(--color-ink)]">Konfigurasi reminder</h3>
              <dl className="mt-2 grid gap-2 text-sm sm:grid-cols-2">
                <div>
                  <dt className="text-[var(--color-muted)]">Hari sebelum</dt>
                  <dd className="font-medium">{activity.reminder_days_before}</dd>
                </div>
                <div>
                  <dt className="text-[var(--color-muted)]">Jam</dt>
                  <dd className="font-medium">{activity.reminder_time ?? '—'}</dd>
                </div>
                <div className="sm:col-span-2">
                  <dt className="text-[var(--color-muted)]">Pesan</dt>
                  <dd className="mt-1 whitespace-pre-wrap">{activity.reminder_message || '—'}</dd>
                </div>
              </dl>
            </div>
          </div>
        </Card>
      </div>

      <Modal
        open={editOpen}
        title="Edit Kegiatan"
        description="Perbarui detail kegiatan. Field operasional reminder dikelola server."
        onClose={() => setEditOpen(false)}
        className="w-[min(100%-2rem,40rem)]"
      >
        <ActivityForm
          key={activity.id}
          mode="edit"
          initial={activity}
          submitLabel="Simpan perubahan"
          onCancel={() => setEditOpen(false)}
          onSubmitUpdate={async (payload) => {
            await updateMutation.mutateAsync({ id: activity.id, payload })
            setEditOpen(false)
            setFeedback('Kegiatan berhasil diperbarui.')
          }}
        />
      </Modal>

      <ConfirmDialog
        open={deleteOpen}
        title="Hapus kegiatan?"
        message={`Kegiatan "${activity.name}" akan dihapus secara permanen.`}
        confirmLabel="Hapus"
        loading={deleteMutation.isPending}
        onCancel={() => setDeleteOpen(false)}
        onConfirm={() => {
          void (async () => {
            try {
              await deleteMutation.mutateAsync(activity.id)
              void navigate('/activities', { replace: true })
            } catch (err) {
              setActionError(toActivityErrorMessage(err))
              setDeleteOpen(false)
            }
          })()
        }}
      />
    </div>
  )
}
