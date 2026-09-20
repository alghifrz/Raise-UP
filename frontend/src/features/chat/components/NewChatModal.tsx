import { useMemo, useState } from 'react'
import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { InlineAlert } from '../../../components/ui/InlineAlert'
import { LoadingState } from '../../../components/ui/LoadingState'
import { Modal } from '../../../components/ui/Modal'
import { useDebouncedValue } from '../../../lib/hooks/useDebouncedValue'
import { cn, formatRoleLabel } from '../../../lib/utils'
import { toChatErrorMessage } from '../errors'
import { useChatContacts, useCreateConversation } from '../hooks'
import type { ChatContact, CreateConversationRequest } from '../types'
import { initialsFromName } from '../types'

type NewChatModalProps = {
  open: boolean
  onClose: () => void
  onCreated: (conversationId: string) => void
}

type Mode = 'direct' | 'group' | 'whatsapp'

export function NewChatModal({ open, onClose, onCreated }: NewChatModalProps) {
  const [mode, setMode] = useState<Mode>('direct')
  const [searchInput, setSearchInput] = useState('')
  const debouncedSearch = useDebouncedValue(searchInput, 300)
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [groupTitle, setGroupTitle] = useState('')
  const [waPhone, setWaPhone] = useState('')
  const [waName, setWaName] = useState('')
  const [feedback, setFeedback] = useState<string | null>(null)

  const filters = useMemo(
    () => ({
      page: 1,
      page_size: 50,
      search: debouncedSearch.trim(),
    }),
    [debouncedSearch],
  )

  const { data, isLoading, isError, refetch } = useChatContacts(filters, open && mode !== 'whatsapp')
  const createMutation = useCreateConversation()

  const contacts = data?.items ?? []

  function reset() {
    setMode('direct')
    setSearchInput('')
    setSelectedIds([])
    setGroupTitle('')
    setWaPhone('')
    setWaName('')
    setFeedback(null)
  }

  function handleClose() {
    reset()
    onClose()
  }

  function toggleContact(contact: ChatContact) {
    setFeedback(null)
    if (mode === 'direct') {
      setSelectedIds([contact.id])
      return
    }
    setSelectedIds((prev) =>
      prev.includes(contact.id) ? prev.filter((id) => id !== contact.id) : [...prev, contact.id],
    )
  }

  async function handleSubmit() {
    setFeedback(null)

    let payload: CreateConversationRequest
    if (mode === 'direct') {
      const participantId = selectedIds[0]
      if (!participantId) {
        setFeedback('Pilih satu admin untuk memulai chat.')
        return
      }
      payload = { type: 'DIRECT', participant_id: participantId }
    } else if (mode === 'group') {
      const title = groupTitle.trim()
      if (!title) {
        setFeedback('Judul grup wajib diisi.')
        return
      }
      if (selectedIds.length === 0) {
        setFeedback('Pilih minimal satu anggota grup.')
        return
      }
      payload = { type: 'GROUP', title, participant_ids: selectedIds }
    } else {
      const phone = waPhone.trim()
      if (!phone) {
        setFeedback('Nomor WhatsApp wajib diisi.')
        return
      }
      payload = {
        type: 'WHATSAPP',
        phone,
        contact_name: waName.trim() || undefined,
      }
    }

    try {
      const conversation = await createMutation.mutateAsync(payload)
      reset()
      onCreated(conversation.id)
    } catch (error) {
      setFeedback(toChatErrorMessage(error))
    }
  }

  const modes: Array<{ id: Mode; label: string }> = [
    { id: 'direct', label: 'Admin' },
    { id: 'group', label: 'Grup' },
    { id: 'whatsapp', label: 'WhatsApp' },
  ]

  return (
    <Modal
      open={open}
      title="Chat baru"
      description="Mulai chat admin, grup, atau thread WhatsApp warga."
      onClose={handleClose}
      className="w-[min(100%-2rem,28rem)]"
    >
      <div className="mb-4 flex gap-1 rounded-2xl bg-[var(--color-surface)] p-1">
        {modes.map((item) => (
          <button
            key={item.id}
            type="button"
            onClick={() => {
              setMode(item.id)
              setFeedback(null)
              if (item.id === 'direct') {
                setSelectedIds((prev) => (prev[0] ? [prev[0]] : []))
              }
            }}
            className={cn(
              'flex-1 rounded-xl px-2 py-2 text-xs font-semibold transition sm:text-sm',
              mode === item.id
                ? 'bg-[var(--color-panel)] text-[var(--color-ink)] shadow-sm'
                : 'text-[var(--color-muted)] hover:text-[var(--color-ink)]',
            )}
          >
            {item.label}
          </button>
        ))}
      </div>

      {mode === 'group' ? (
        <div className="mb-3">
          <Input
            label="Judul grup"
            name="group_title"
            value={groupTitle}
            onChange={(event) => setGroupTitle(event.target.value)}
            placeholder="Contoh: Pengurus Blok A"
            required
          />
        </div>
      ) : null}

      {mode === 'whatsapp' ? (
        <div className="space-y-3">
          <Input
            label="Nomor WhatsApp"
            name="wa_phone"
            value={waPhone}
            onChange={(event) => setWaPhone(event.target.value)}
            placeholder="08… atau 628…"
            required
          />
          <Input
            label="Nama kontak (opsional)"
            name="wa_name"
            value={waName}
            onChange={(event) => setWaName(event.target.value)}
            placeholder="Nama warga"
          />
        </div>
      ) : (
        <>
          <Input
            label="Cari admin"
            name="contact_search"
            value={searchInput}
            onChange={(event) => setSearchInput(event.target.value)}
            placeholder="Nama atau email…"
          />

          <div className="mt-3 max-h-64 min-h-40 overflow-y-auto rounded-2xl border border-[var(--color-line)]">
            {isLoading ? <LoadingState label="Memuat kontak…" className="min-h-40" /> : null}
            {isError ? (
              <div className="space-y-2 p-4 text-center">
                <p className="text-sm text-[var(--color-danger)]">Gagal memuat kontak.</p>
                <Button type="button" variant="secondary" onClick={() => void refetch()}>
                  Coba lagi
                </Button>
              </div>
            ) : null}
            {!isLoading && !isError && contacts.length === 0 ? (
              <p className="px-4 py-8 text-center text-sm text-[var(--color-muted)]">
                Tidak ada admin lain.
              </p>
            ) : null}
            {!isLoading && !isError
              ? contacts.map((contact) => {
                  const selected = selectedIds.includes(contact.id)
                  return (
                    <button
                      key={contact.id}
                      type="button"
                      onClick={() => toggleContact(contact)}
                      className={cn(
                        'flex w-full items-center gap-3 border-b border-[var(--color-line)]/70 px-3 py-2.5 text-left last:border-b-0',
                        selected ? 'bg-[var(--color-accent-soft)]' : 'hover:bg-[var(--color-surface)]',
                      )}
                    >
                      <span
                        className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-[var(--color-accent)] text-xs font-bold text-white"
                        aria-hidden
                      >
                        {initialsFromName(contact.name)}
                      </span>
                      <span className="min-w-0 flex-1">
                        <span className="block truncate text-sm font-semibold text-[var(--color-ink)]">
                          {contact.name}
                        </span>
                        <span className="block truncate text-xs text-[var(--color-muted)]">
                          {contact.email} · {formatRoleLabel(contact.role)}
                        </span>
                      </span>
                      {selected ? (
                        <span
                          className="material-symbols-outlined text-[var(--color-accent)]"
                          aria-hidden
                        >
                          check_circle
                        </span>
                      ) : null}
                    </button>
                  )
                })
              : null}
          </div>
        </>
      )}

      {feedback ? (
        <div className="mt-3">
          <InlineAlert tone="danger">{feedback}</InlineAlert>
        </div>
      ) : null}

      <div className="mt-4 flex justify-end gap-2">
        <Button type="button" variant="secondary" onClick={handleClose}>
          Batal
        </Button>
        <Button
          type="button"
          loading={createMutation.isPending}
          onClick={() => void handleSubmit()}
        >
          {mode === 'direct' ? 'Mulai chat' : mode === 'group' ? 'Buat grup' : 'Buka WhatsApp'}
        </Button>
      </div>
    </Modal>
  )
}
