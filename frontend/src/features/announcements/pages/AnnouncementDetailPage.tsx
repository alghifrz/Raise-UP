import { useQueries } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { Button } from '../../../components/ui/Button'
import { Card } from '../../../components/ui/Card'
import { ConfirmDialog, Modal } from '../../../components/ui/Modal'
import { ErrorState } from '../../../components/ui/ErrorState'
import { LoadingState } from '../../../components/ui/LoadingState'
import { PageHeader } from '../../../components/ui/PageHeader'
import { formatDateTime } from '../../../lib/utils'
import { getResident } from '../../residents/api'
import { residentDetailQueryKey } from '../../residents/hooks'
import { AnnouncementForm } from '../components/AnnouncementForm'
import {
  AnnouncementStatusBadge,
  AnnouncementVisibilityBadge,
} from '../components/AnnouncementBadges'
import { toAnnouncementErrorMessage } from '../errors'
import {
  useAnnouncement,
  useDeleteAnnouncement,
  usePublishAnnouncement,
  useUpdateAnnouncement,
} from '../hooks'
import { formatAuthorLabel } from '../types'

export function AnnouncementDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data, isLoading, isError, refetch, error } = useAnnouncement(id)
  const updateMutation = useUpdateAnnouncement()
  const publishMutation = usePublishAnnouncement()
  const deleteMutation = useDeleteAnnouncement()

  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [feedback, setFeedback] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const recipientQueries = useQueries({
    queries: (data?.recipient_ids ?? []).map((recipientId) => ({
      queryKey: residentDetailQueryKey(recipientId),
      queryFn: ({ signal }: { signal?: AbortSignal }) => getResident(recipientId, signal),
      enabled: Boolean(data && data.recipient_ids.length > 0),
    })),
  })

  if (isLoading) {
    return <LoadingState label="Memuat detail pengumuman…" />
  }

  if (isError || !data) {
    return (
      <div className="space-y-4">
        <PageHeader title="Detail Pengumuman" />
        <ErrorState
          title="Gagal memuat pengumuman."
          message={toAnnouncementErrorMessage(error)}
          onRetry={() => {
            void refetch()
          }}
        />
        <Button type="button" variant="secondary" onClick={() => void navigate('/announcements')}>
          Kembali
        </Button>
      </div>
    )
  }

  const announcement = data
  const canPublish = announcement.status === 'DRAFT'

  async function handlePublish() {
    setActionError(null)
    try {
      const published = await publishMutation.mutateAsync(announcement.id)
      const delivery = published.delivery
      if (!delivery || delivery.total === 0) {
        setFeedback('Pengumuman berhasil diterbitkan. Tidak ada penerima WhatsApp.')
      } else if (delivery.failed === 0) {
        setFeedback(
          `Pengumuman berhasil diterbitkan dan dikirim ke ${delivery.sent} penerima WhatsApp.`,
        )
      } else {
        setFeedback(
          `Pengumuman diterbitkan. WhatsApp terkirim ${delivery.sent} dari ${delivery.total} penerima (${delivery.failed} gagal).`,
        )
      }
    } catch (err) {
      setActionError(toAnnouncementErrorMessage(err))
    }
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title={announcement.title}
        description="Detail dan tindak lanjut pengumuman."
        actions={
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="secondary" onClick={() => void navigate('/announcements')}>
              Kembali
            </Button>
            <Button type="button" variant="secondary" onClick={() => setEditOpen(true)}>
              Edit
            </Button>
            {canPublish ? (
              <Button
                type="button"
                loading={publishMutation.isPending}
                onClick={() => {
                  void handlePublish()
                }}
              >
                Terbitkan
              </Button>
            ) : null}
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
                <AnnouncementStatusBadge status={announcement.status} />
              </dd>
            </div>
            <div>
              <dt className="text-[var(--color-muted)]">Visibilitas</dt>
              <dd className="mt-1">
                <AnnouncementVisibilityBadge visibility={announcement.visibility} />
              </dd>
            </div>
            <div>
              <dt className="text-[var(--color-muted)]">Penulis</dt>
              <dd className="mt-1 font-medium" title={announcement.author_id}>
                {formatAuthorLabel(announcement.author_id)}
              </dd>
            </div>
            <div>
              <dt className="text-[var(--color-muted)]">Dipublikasikan</dt>
              <dd className="mt-1">
                {announcement.published_at ? formatDateTime(announcement.published_at) : '—'}
              </dd>
            </div>
            <div>
              <dt className="text-[var(--color-muted)]">Dibuat</dt>
              <dd className="mt-1">{formatDateTime(announcement.created_at)}</dd>
            </div>
            <div>
              <dt className="text-[var(--color-muted)]">Diperbarui</dt>
              <dd className="mt-1">{formatDateTime(announcement.updated_at)}</dd>
            </div>
          </dl>
        </Card>

        <Card title="Konten" className="lg:col-span-2">
          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-medium text-[var(--color-muted)]">Isi</h3>
              <p className="mt-1 whitespace-pre-wrap rounded-2xl border border-[var(--color-line)] bg-[var(--color-accent-soft)] px-3 py-3 text-sm">
                {announcement.body}
              </p>
            </div>
            {announcement.thumbnail_url ? (
              <div>
                <h3 className="text-sm font-medium text-[var(--color-muted)]">Thumbnail</h3>
                <a
                  className="mt-1 inline-block text-sm text-[var(--color-accent)] hover:underline"
                  href={announcement.thumbnail_url}
                  target="_blank"
                  rel="noreferrer"
                >
                  {announcement.thumbnail_url}
                </a>
              </div>
            ) : null}
          </div>
        </Card>
      </div>

      <Card title="Penerima">
        {announcement.recipient_ids.length === 0 ? (
          <p className="text-sm text-[var(--color-muted)]">Tidak ada penerima tercatat.</p>
        ) : (
          <ul className="space-y-2">
            {announcement.recipient_ids.map((recipientId, index) => {
              const resident = recipientQueries[index]?.data
              return (
                <li
                  key={recipientId}
                  className="rounded-md border border-[var(--color-line)] px-3 py-2 text-sm"
                >
                  {resident ? (
                    <span>
                      {resident.name} · {resident.phone}
                    </span>
                  ) : (
                    <span title={recipientId}>
                      {recipientQueries[index]?.isLoading
                        ? 'Memuat…'
                        : `Warga ${formatAuthorLabel(recipientId)}`}
                    </span>
                  )}{' '}
                  <Link className="text-[var(--color-accent)] hover:underline" to="/residents">
                    Lihat warga
                  </Link>
                </li>
              )
            })}
          </ul>
        )}
      </Card>

      <Modal
        open={editOpen}
        title="Edit Pengumuman"
        description="Status diterbitkan melalui tombol Terbitkan, bukan form edit."
        onClose={() => setEditOpen(false)}
        className="w-[min(100%-2rem,42rem)]"
      >
        <AnnouncementForm
          key={announcement.id}
          mode="edit"
          initial={announcement}
          submitLabel="Simpan perubahan"
          onCancel={() => setEditOpen(false)}
          onSubmitUpdate={async (payload) => {
            await updateMutation.mutateAsync({ id: announcement.id, payload })
            setEditOpen(false)
            setFeedback('Pengumuman berhasil diperbarui.')
          }}
        />
      </Modal>

      <ConfirmDialog
        open={deleteOpen}
        title="Hapus pengumuman?"
        message={`Pengumuman "${announcement.title}" akan dihapus secara permanen.`}
        confirmLabel="Hapus"
        loading={deleteMutation.isPending}
        onCancel={() => setDeleteOpen(false)}
        onConfirm={() => {
          void (async () => {
            try {
              await deleteMutation.mutateAsync(announcement.id)
              void navigate('/announcements', { replace: true })
            } catch (err) {
              setActionError(toAnnouncementErrorMessage(err))
              setDeleteOpen(false)
            }
          })()
        }}
      />
    </div>
  )
}
