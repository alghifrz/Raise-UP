import { useState, type FormEvent } from 'react'
import { Button } from '../../../components/ui/Button'
import { Card } from '../../../components/ui/Card'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorState } from '../../../components/ui/ErrorState'
import { InlineAlert } from '../../../components/ui/InlineAlert'
import { LoadingState } from '../../../components/ui/LoadingState'
import { ConfirmDialog, Modal } from '../../../components/ui/Modal'
import { PageHeader } from '../../../components/ui/PageHeader'
import { Textarea } from '../../../components/ui/Textarea'
import { formatDateTime } from '../../../lib/utils'
import { OfficialCard } from '../components/OfficialCard'
import { OfficialForm } from '../components/OfficialForm'
import { toVillageErrorMessage } from '../errors'
import {
  useCreateVillageOfficial,
  useDeleteVillageOfficial,
  useUpdateVillageOfficial,
  useUpdateVillageProfile,
  useVillageOfficials,
  useVillageProfile,
} from '../hooks'
import type { VillageOfficial, VillageProfile } from '../types'

type ProfileFormProps = {
  initial: VillageProfile
}

function ProfileForm({ initial }: ProfileFormProps) {
  const updateProfileMutation = useUpdateVillageProfile()
  const [history, setHistory] = useState(initial.history)
  const [vision, setVision] = useState(initial.vision)
  const [mission, setMission] = useState(initial.mission)
  const [savedAt, setSavedAt] = useState(initial.updated_at)
  const [profileFeedback, setProfileFeedback] = useState<string | null>(null)
  const [profileError, setProfileError] = useState<string | null>(null)

  async function handleSaveProfile(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setProfileError(null)
    setProfileFeedback(null)
    try {
      const updated = await updateProfileMutation.mutateAsync({
        history,
        vision,
        mission,
      })
      setSavedAt(updated.updated_at)
      setProfileFeedback('Profil wilayah berhasil disimpan.')
    } catch (error) {
      setProfileError(toVillageErrorMessage(error))
    }
  }

  return (
    <form className="space-y-4" onSubmit={handleSaveProfile}>
      {profileFeedback ? <InlineAlert>{profileFeedback}</InlineAlert> : null}
      <Card title="Konten profil">
        <div className="space-y-4">
          <Textarea
            name="history"
            label="Sejarah"
            value={history}
            onChange={(event) => setHistory(event.target.value)}
            disabled={updateProfileMutation.isPending}
            className="min-h-40"
          />
          <Textarea
            name="vision"
            label="Visi"
            value={vision}
            onChange={(event) => setVision(event.target.value)}
            disabled={updateProfileMutation.isPending}
          />
          <Textarea
            name="mission"
            label="Misi"
            value={mission}
            onChange={(event) => setMission(event.target.value)}
            disabled={updateProfileMutation.isPending}
          />
          <p className="text-xs text-[var(--color-muted)]">
            Terakhir diperbarui: {formatDateTime(savedAt)}
          </p>
        </div>
      </Card>

      {profileError ? <InlineAlert tone="danger">{profileError}</InlineAlert> : null}

      <div className="flex justify-end">
        <Button type="submit" loading={updateProfileMutation.isPending}>
          Simpan Profil
        </Button>
      </div>
    </form>
  )
}

export function VillagePage() {
  const profileQuery = useVillageProfile()
  const officialsQuery = useVillageOfficials()
  const createOfficialMutation = useCreateVillageOfficial()
  const updateOfficialMutation = useUpdateVillageOfficial()
  const deleteOfficialMutation = useDeleteVillageOfficial()

  const [createOpen, setCreateOpen] = useState(false)
  const [editing, setEditing] = useState<VillageOfficial | null>(null)
  const [deleting, setDeleting] = useState<VillageOfficial | null>(null)
  const [officialFeedback, setOfficialFeedback] = useState<string | null>(null)
  const [officialError, setOfficialError] = useState<string | null>(null)

  return (
    <div className="space-y-6">
      <PageHeader
        title="Profil Wilayah"
        description="Kelola sejarah, visi, misi, dan data pengurus."
      />

      <section className="space-y-3">
        <h2 className="text-lg font-semibold text-[var(--color-ink)]">Profil Wilayah</h2>

        {profileQuery.isLoading ? <LoadingState label="Memuat profil wilayah…" /> : null}

        {profileQuery.isError ? (
          <ErrorState
            title="Gagal memuat profil wilayah."
            message={toVillageErrorMessage(profileQuery.error)}
            onRetry={() => {
              void profileQuery.refetch()
            }}
          />
        ) : null}

        {profileQuery.data ? (
          <ProfileForm key={profileQuery.data.updated_at} initial={profileQuery.data} />
        ) : null}
      </section>

      <section className="space-y-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-lg font-semibold text-[var(--color-ink)]">Pejabat / Pengurus</h2>
          <Button type="button" onClick={() => setCreateOpen(true)}>
            Tambah Pengurus
          </Button>
        </div>

        {officialFeedback ? <InlineAlert>{officialFeedback}</InlineAlert> : null}

        {officialError ? <InlineAlert tone="danger">{officialError}</InlineAlert> : null}

        {officialsQuery.isLoading ? <LoadingState label="Memuat data pengurus…" /> : null}

        {officialsQuery.isError ? (
          <ErrorState
            title="Gagal memuat data pengurus."
            message={toVillageErrorMessage(officialsQuery.error)}
            onRetry={() => {
              void officialsQuery.refetch()
            }}
          />
        ) : null}

        {!officialsQuery.isLoading &&
        !officialsQuery.isError &&
        (officialsQuery.data?.length ?? 0) === 0 ? (
          <EmptyState
            title="Belum ada pengurus yang ditambahkan."
            description="Tambahkan pejabat atau pengurus wilayah."
            action={
              <Button type="button" onClick={() => setCreateOpen(true)}>
                Tambah Pengurus
              </Button>
            }
          />
        ) : null}

        {!officialsQuery.isLoading && !officialsQuery.isError && (officialsQuery.data?.length ?? 0) > 0 ? (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {officialsQuery.data?.map((official) => (
              <OfficialCard
                key={official.id}
                official={official}
                onEdit={setEditing}
                onDelete={(item) => {
                  setOfficialError(null)
                  setDeleting(item)
                }}
              />
            ))}
          </div>
        ) : null}
      </section>

      <Modal open={createOpen} title="Tambah Pengurus" onClose={() => setCreateOpen(false)}>
        <OfficialForm
          mode="create"
          submitLabel="Simpan"
          onCancel={() => setCreateOpen(false)}
          onSubmitCreate={async (payload) => {
            await createOfficialMutation.mutateAsync(payload)
            setCreateOpen(false)
            setOfficialFeedback('Pengurus berhasil ditambahkan.')
          }}
        />
      </Modal>

      <Modal open={Boolean(editing)} title="Edit Pengurus" onClose={() => setEditing(null)}>
        {editing ? (
          <OfficialForm
            key={editing.id}
            mode="edit"
            initial={editing}
            submitLabel="Simpan perubahan"
            onCancel={() => setEditing(null)}
            onSubmitUpdate={async (payload) => {
              await updateOfficialMutation.mutateAsync({ id: editing.id, payload })
              setEditing(null)
              setOfficialFeedback('Pengurus berhasil diperbarui.')
            }}
          />
        ) : null}
      </Modal>

      <ConfirmDialog
        open={Boolean(deleting)}
        title="Hapus pengurus?"
        message={deleting ? `"${deleting.name}" akan dihapus dari daftar pengurus.` : ''}
        confirmLabel="Hapus"
        loading={deleteOfficialMutation.isPending}
        onCancel={() => setDeleting(null)}
        onConfirm={() => {
          if (!deleting) {
            return
          }
          void (async () => {
            try {
              await deleteOfficialMutation.mutateAsync(deleting.id)
              setDeleting(null)
              setOfficialFeedback('Pengurus berhasil dihapus.')
            } catch (error) {
              setOfficialError(toVillageErrorMessage(error))
              setDeleting(null)
            }
          })()
        }}
      />
    </div>
  )
}
