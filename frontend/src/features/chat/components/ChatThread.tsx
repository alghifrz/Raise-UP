import { useEffect, useMemo, useRef, useState, type FormEvent, type KeyboardEvent } from 'react'
import { Button } from '../../../components/ui/Button'
import { LoadingState } from '../../../components/ui/LoadingState'
import { cn } from '../../../lib/utils'
import type { ChatMessage, Conversation } from '../types'
import { formatMessageTime, initialsFromName } from '../types'

type ChatThreadProps = {
  conversation: Conversation | undefined
  messages: ChatMessage[]
  currentUserId: string | undefined
  loading?: boolean
  sending?: boolean
  errorMessage?: string | null
  onSend: (body: string) => Promise<void> | void
  onBack?: () => void
}

export function ChatThread({
  conversation,
  messages,
  currentUserId,
  loading = false,
  sending = false,
  errorMessage,
  onSend,
  onBack,
}: ChatThreadProps) {
  const [draft, setDraft] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)
  const listRef = useRef<HTMLDivElement>(null)

  const chronological = useMemo(
    () => [...messages].sort((a, b) => a.created_at.localeCompare(b.created_at)),
    [messages],
  )

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' })
  }, [chronological.length, conversation?.id])

  async function handleSubmit(event?: FormEvent) {
    event?.preventDefault()
    const body = draft.trim()
    if (!body || sending) {
      return
    }
    setDraft('')
    await onSend(body)
  }

  function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      void handleSubmit()
    }
  }

  if (!conversation) {
    return (
      <div className="flex h-full min-h-0 flex-col items-center justify-center bg-[var(--color-surface)] px-6 text-center">
        <span
          className="mb-3 flex h-16 w-16 items-center justify-center rounded-full bg-[var(--color-accent-soft)] text-[var(--color-accent)]"
          aria-hidden
        >
          <span className="material-symbols-outlined text-[32px]">forum</span>
        </span>
        <h2 className="text-lg font-bold text-[var(--color-ink)]">Pilih percakapan</h2>
        <p className="mt-1 max-w-sm text-sm text-[var(--color-muted)]">
          Pilih chat di sebelah kiri, atau mulai percakapan baru dengan admin lain.
        </p>
      </div>
    )
  }

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden bg-[var(--color-surface)]">
      <header className="flex shrink-0 items-center gap-3 border-b border-[var(--color-line)] bg-[var(--color-panel)] px-3 py-3 sm:px-4">
        {onBack ? (
          <button
            type="button"
            onClick={onBack}
            className="inline-flex h-10 w-10 items-center justify-center rounded-full text-[var(--color-ink)] transition hover:bg-[var(--color-accent-soft)] lg:hidden"
            aria-label="Kembali ke daftar chat"
          >
            <span className="material-symbols-outlined" aria-hidden>
              arrow_back
            </span>
          </button>
        ) : null}

        <span
          className={cn(
            'flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-sm font-bold text-white',
            conversation.type === 'GROUP'
              ? 'bg-[var(--color-tertiary)]'
              : conversation.type === 'WHATSAPP'
                ? 'bg-[#128C7E]'
                : 'bg-[var(--color-accent)]',
          )}
          aria-hidden
        >
          {conversation.type === 'GROUP' ? (
            <span className="material-symbols-outlined text-[20px]">groups</span>
          ) : conversation.type === 'WHATSAPP' ? (
            <span className="material-symbols-outlined text-[20px]">chat</span>
          ) : (
            initialsFromName(conversation.display_name)
          )}
        </span>

        <div className="min-w-0 flex-1">
          <h2 className="truncate text-sm font-bold text-[var(--color-ink)]">
            {conversation.display_name}
          </h2>
          <p className="truncate text-xs text-[var(--color-muted)]">
            {conversation.type === 'GROUP'
              ? `${conversation.participants.length} anggota`
              : conversation.type === 'WHATSAPP'
                ? conversation.wa_contact_phone
                  ? `WhatsApp · ${conversation.wa_contact_phone}`
                  : 'WhatsApp'
                : conversation.participants.find((p) => p.id !== currentUserId)?.email ??
                  'Chat langsung'}
          </p>
        </div>
      </header>

      <div
        ref={listRef}
        className="min-h-0 flex-1 space-y-2 overflow-y-auto px-3 py-4 sm:px-5"
        aria-live="polite"
      >
        {loading && chronological.length === 0 ? (
          <LoadingState label="Memuat pesan…" className="min-h-48" />
        ) : null}

        {!loading && chronological.length === 0 ? (
          <p className="py-10 text-center text-sm text-[var(--color-muted)]">
            Belum ada pesan. Mulai percakapan di bawah.
          </p>
        ) : null}

        {chronological.map((message) => {
          // SYSTEM (bot) + pesan admin sendiri = sisi sistem (kanan)
          const mine =
            message.sender_kind === 'SYSTEM' ||
            (message.sender_kind === 'USER' &&
              Boolean(currentUserId) &&
              message.sender_id === currentUserId)
          const read =
            message.is_read ||
            (conversation.type === 'WHATSAPP' &&
              typeof message.wa_status === 'string' &&
              message.wa_status.toLowerCase() === 'read')
          return (
            <div
              key={message.id}
              className={cn('flex', mine ? 'justify-end' : 'justify-start')}
            >
              <div
                className={cn(
                  'max-w-[min(100%,28rem)] rounded-2xl px-3.5 py-2 shadow-sm',
                  mine
                    ? 'rounded-br-md bg-[var(--color-secondary)] text-[var(--color-secondary-ink)]'
                    : 'rounded-bl-md bg-[#DCF8C6] text-[var(--color-ink)]',
                )}
              >
                {message.sender_kind === 'SYSTEM' ? (
                  <p className="mb-0.5 text-[11px] font-bold text-[var(--color-secondary-ink)]/80">
                    Bot
                  </p>
                ) : null}
                {!mine && conversation.type === 'GROUP' && message.sender_kind === 'USER' ? (
                  <p className="mb-0.5 text-[11px] font-bold text-[var(--color-tertiary)]">
                    {conversation.participants.find((p) => p.id === message.sender_id)?.name ??
                      'Admin'}
                  </p>
                ) : null}
                <p className="whitespace-pre-wrap text-sm leading-relaxed">{message.body}</p>
                <p
                  className={cn(
                    'mt-1 flex items-center justify-end gap-1 text-[10px]',
                    mine ? 'text-[var(--color-secondary-ink)]/70' : 'text-[var(--color-muted)]',
                  )}
                >
                  <span>{formatMessageTime(message.created_at)}</span>
                  {mine ? (
                    <span
                      className={cn(
                        'material-symbols-outlined text-[14px] leading-none',
                        read ? 'text-[#53BDEB]' : 'opacity-70',
                      )}
                      aria-label={read ? 'Dibaca' : 'Terkirim'}
                    >
                      done_all
                    </span>
                  ) : null}
                </p>
              </div>
            </div>
          )
        })}
        <div ref={bottomRef} />
      </div>

      {errorMessage ? (
        <p className="shrink-0 px-4 pb-2 text-sm text-[var(--color-danger)]" role="alert">
          {errorMessage}
        </p>
      ) : null}

      <form
        onSubmit={(event) => void handleSubmit(event)}
        className="shrink-0 border-t border-[var(--color-line)] bg-[var(--color-panel)] px-3 py-3 sm:px-4"
      >
        <div className="flex items-end gap-2">
          <label className="sr-only" htmlFor="chat-composer">
            Tulis pesan
          </label>
          <textarea
            id="chat-composer"
            rows={1}
            value={draft}
            onChange={(event) => setDraft(event.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Tulis pesan…"
            className="max-h-32 min-h-[2.75rem] flex-1 resize-y rounded-2xl border border-[var(--color-line)] bg-[var(--color-surface)] px-3.5 py-2.5 text-sm text-[var(--color-ink)] outline-none transition placeholder:text-[var(--color-muted)]/60 focus:border-[var(--color-tertiary)] focus:bg-[var(--color-panel)] focus:ring-4 focus:ring-[var(--color-tertiary)]/15"
            disabled={sending}
          />
          <Button
            type="submit"
            loading={sending}
            disabled={!draft.trim()}
            className="shrink-0 rounded-full px-4"
            aria-label="Kirim pesan"
          >
            <span className="material-symbols-outlined text-[20px]" aria-hidden>
              send
            </span>
          </Button>
        </div>
      </form>
    </div>
  )
}
