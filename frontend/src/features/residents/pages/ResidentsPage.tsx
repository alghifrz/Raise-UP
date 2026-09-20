import { useMemo, useState } from 'react'
import { Button } from '../../../components/ui/Button'
import { DataTable } from '../../../components/ui/DataTable'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorState } from '../../../components/ui/ErrorState'
import { InlineAlert } from '../../../components/ui/InlineAlert'
import { Input } from '../../../components/ui/Input'
import { LoadingState } from '../../../components/ui/LoadingState'
import { ConfirmDialog, Modal } from '../../../components/ui/Modal'
import { PageHeader } from '../../../components/ui/PageHeader'
import { Pagination } from '../../../components/ui/Pagination'
import { Select } from '../../../components/ui/Select'
import { useDebouncedValue } from '../../../lib/hooks/useDebouncedValue'
import { formatDateTime } from '../../../lib/utils'
import { ImportResidentsModal } from '../components/ImportResidentsModal'
import { ResidentForm } from '../components/ResidentForm'
import { toResidentErrorMessage } from '../errors'
import {
  useCreateResident,
  useDeleteResident,
  useInvalidateResidentsAfterImport,
  useResidents,
  useUpdateResident,
} from '../hooks'
import type { Resident, ResidentFilters, ResidentGender } from '../types'
import { formatResidentGender, RESIDENT_GENDER_OPTIONS } from '../types'

const DEFAULT_PAGE_SIZE = 20

export function ResidentsPage() {
  const [searchInput, setSearchInput] = useState('')
  const debouncedSearch = useDebouncedValue(searchInput, 300)
  const [gender, setGender] = useState<ResidentGender | ''>('')
  const [page, setPage] = useState(1)
  const [createOpen, setCreateOpen] = useState(false)
  const [importOpen, setImportOpen] = useState(false)
  const [editing, setEditing] = useState<Resident | null>(null)
  const [deleting, setDeleting] = useState<Resident | null>(null)
  const [feedback, setFeedback] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const filters = useMemo<ResidentFilters>(
    () => ({
      page,
      page_size: DEFAULT_PAGE_SIZE,
      search: debouncedSearch.trim(),
      gender,
    }),
    [page, debouncedSearch, gender],
  )

  const { data, isLoading, isError, isFetching, refetch } = useResidents(filters)
  const createMutation = useCreateResident()
  const updateMutation = useUpdateResident()
  const deleteMutation = useDeleteResident()
  const invalidateAfterImport = useInvalidateResidentsAfterImport()

  const rows = data?.items ?? []
  const meta = data?.meta

  return (
    <div className="space-y-4">
      <PageHeader
        title="Warga"
        description="Kelola data warga RW."
        actions={
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="secondary" onClick={() => setImportOpen(true)}>
              Import Excel
            </Button>
            <Button type="button" onClick={() => setCreateOpen(true)}>
              Tambah Warga
            </Button>
          </div>
        }
      />

      {feedback ? <InlineAlert>{feedback}</InlineAlert> : null}

      <div className="grid gap-3 rounded-2xl border border-[var(--color-line)] bg-[var(--color-panel)] p-4 sm:grid-cols-2">
        <Input
          name="search"
          label="Cari"
          placeholder="Nama atau telepon"
          value={searchInput}
          onChange={(event) => {
            setSearchInput(event.target.value)
            setPage(1)
          }}
        />
        <Select
          name="genderFilter"
          label="Jenis Kelamin"
          value={gender}
          onChange={(event) => {
            setGender(event.target.value as ResidentGender | '')
            setPage(1)
          }}
          options={RESIDENT_GENDER_OPTIONS}
          placeholder="Semua"
        />
      </div>

      {isLoading ? <LoadingState label="Memuat data warga…" /> : null}

      {isError ? (
        <ErrorState
          title="Gagal memuat data warga."
          message="Periksa koneksi atau coba lagi."
          onRetry={() => {
            void refetch()
          }}
        />
      ) : null}

      {!isLoading && !isError && rows.length === 0 ? (
        <EmptyState
          title="Belum ada data warga"
          description="Tambahkan warga baru, import Excel, atau ubah filter pencarian."
          action={
            <div className="flex flex-wrap justify-center gap-2">
              <Button type="button" variant="secondary" onClick={() => setImportOpen(true)}>
                Import Excel
              </Button>
              <Button type="button" onClick={() => setCreateOpen(true)}>
                Tambah Warga
              </Button>
            </div>
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
              { key: 'name', header: 'Nama', render: (row) => row.name },
              { key: 'phone', header: 'No. Telepon', render: (row) => row.phone },
              {
                key: 'gender',
                header: 'Jenis Kelamin',
                render: (row) => formatResidentGender(row.gender),
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

      <Modal
        open={createOpen}
        title="Tambah Warga"
        description="Isi data warga baru."
        onClose={() => setCreateOpen(false)}
      >
        <ResidentForm
          submitLabel="Simpan"
          onCancel={() => setCreateOpen(false)}
          onSubmit={async (payload) => {
            await createMutation.mutateAsync(payload)
            setCreateOpen(false)
            setFeedback('Warga berhasil ditambahkan.')
          }}
        />
      </Modal>

      <Modal
        open={importOpen}
        title="Import Warga dari Excel"
        description="Unduh template, isi data, lalu unggah file spreadsheet."
        onClose={() => setImportOpen(false)}
        className="w-[min(100%-2rem,36rem)]"
      >
        <ImportResidentsModal
          onCancel={() => setImportOpen(false)}
          onImported={({ success, failed }) => {
            void invalidateAfterImport()
            if (success > 0 && failed === 0) {
              setFeedback(`${success} warga berhasil diimpor.`)
            } else if (success > 0) {
              setFeedback(`${success} warga berhasil diimpor, ${failed} gagal.`)
            }
          }}
        />
      </Modal>

      <Modal
        open={Boolean(editing)}
        title="Edit Warga"
        description="Perbarui data warga."
        onClose={() => setEditing(null)}
      >
        {editing ? (
          <ResidentForm
            key={editing.id}
            initial={editing}
            submitLabel="Simpan perubahan"
            onCancel={() => setEditing(null)}
            onSubmit={async (payload) => {
              await updateMutation.mutateAsync({ id: editing.id, payload })
              setEditing(null)
              setFeedback('Data warga berhasil diperbarui.')
            }}
          />
        ) : null}
      </Modal>

      <ConfirmDialog
        open={Boolean(deleting)}
        title="Hapus warga?"
        message="Menghapus warga dapat ditolak jika data masih digunakan oleh pengaduan, iuran, atau data terkait lainnya."
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
              setDeleteError(null)
              setFeedback('Warga berhasil dihapus.')
            } catch (error) {
              setDeleteError(toResidentErrorMessage(error))
            }
          })()
        }}
      />

      {deleteError ? <InlineAlert tone="danger">{deleteError}</InlineAlert> : null}
    </div>
  )
}
