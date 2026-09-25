import { cn } from '../../../lib/utils'
import type { Conversation } from '../types'
import { formatChatListTime, initialsFromName } from '../types'

type ConversationListProps = {
  items: Conversation[]
  activeId: string | undefined
  search: string
  onSearchChange: (value: string) => void
  onSelect: (id: string) => void
  onNewChat: () => void
  onNotify?: () => void
  whatsappEnabled?: boolean
  botEnabled?: boolean
  botToggling?: boolean
  onToggleBot?: (enabled: boolean) => void
  loading?: boolean
}

export function ConversationList({
  items,
  activeId,
  search,
  onSearchChange,
  onSelect,
  onNewChat,
  onNotify,
  whatsappEnabled = false,
  botEnabled = false,
  botToggling = false,
  onToggleBot,
  loading = false,
}: ConversationListProps) {
  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden border-r border-[var(--color-line)] bg-[var(--color-panel)]">
      <div className="flex shrink-0 items-center justify-between gap-2 border-b border-[var(--color-line)] px-4 py-3">
        <div>
          <h2 className="text-base font-bold tracking-tight text-[var(--color-ink)]">Chat</h2>
          <p className="text-xs text-[var(--color-muted)]">
            Admin & WhatsApp{whatsappEnabled ? '' : ' · WA off'}
            {whatsappEnabled ? (botEnabled ? ' · Bot on' : ' · Bot off') : ''}
          </p>
        </div>
        <div className="flex items-center gap-1.5">
          {onToggleBot ? (
            <button
              type="button"
              role="switch"
              aria-checked={botEnabled}
              aria-label={botEnabled ? 'Matikan chatbot WhatsApp' : 'Nyalakan chatbot WhatsApp'}
              title={
                !whatsappEnabled
                  ? 'WhatsApp belum dikonfigurasi'
                  : botEnabled
                    ? 'Chatbot sedang nyala. Klik untuk mematikan balasan otomatis.'
                    : 'Chatbot sedang mati. Klik untuk menyalakan balasan otomatis.'
              }
              disabled={botToggling}
              onClick={() => onToggleBot(!botEnabled)}
              className={cn(
                'inline-flex h-10 items-center gap-2 rounded-full border px-3 text-xs font-semibold transition',
                botEnabled
                  ? 'border-[#128C7E]/30 bg-[#DCF8C6] text-[#075E54]'
                  : 'border-[var(--color-line)] bg-[var(--color-surface)] text-[var(--color-muted)]',
                'hover:brightness-95 disabled:cursor-not-allowed disabled:opacity-50',
              )}
            >
              <span
                className={cn(
                  'relative inline-flex h-4 w-7 shrink-0 items-center rounded-full transition',
                  botEnabled ? 'bg-[#128C7E]' : 'bg-[var(--color-line)]',
                )}
                aria-hidden
              >
                <span
                  className={cn(
                    'absolute h-3 w-3 rounded-full bg-white shadow transition',
                    botEnabled ? 'left-3.5' : 'left-0.5',
                  )}
                />
              </span>
              Chatbot
            </button>
          ) : null}
          {onNotify ? (
            <button
              type="button"
              onClick={onNotify}
              className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-[var(--color-line)] text-[var(--color-ink)] transition hover:bg-[var(--color-surface)]"
              aria-label="Kirim notifikasi WhatsApp"
              title="Notifikasi WA"
            >
              <span className="material-symbols-outlined text-[20px]" aria-hidden>
                campaign
              </span>
            </button>
          ) : null}
          <button
            type="button"
            onClick={onNewChat}
            className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-[var(--color-secondary)] text-[var(--color-secondary-ink)] shadow-[0_8px_20px_-10px_rgba(210,248,67,0.8)] transition hover:brightness-95"
            aria-label="Mulai chat baru"
            title="Chat baru"
          >
            <span className="material-symbols-outlined text-[22px]" aria-hidden>
              edit_square
            </span>
          </button>
        </div>
      </div>

      <div className="shrink-0 border-b border-[var(--color-line)] px-3 py-2.5">
        <label className="sr-only" htmlFor="chat-search">
          Cari chat
        </label>
        <div className="relative">
          <span
            className="material-symbols-outlined pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-[18px] text-[var(--color-muted)]"
            aria-hidden
          >
            search
          </span>
          <input
            id="chat-search"
            type="search"
            value={search}
            onChange={(event) => onSearchChange(event.target.value)}
            placeholder="Cari nama atau pesan…"
            className="w-full rounded-2xl border border-[var(--color-line)] bg-[var(--color-surface)] py-2.5 pr-3 pl-10 text-sm text-[var(--color-ink)] outline-none transition placeholder:text-[var(--color-muted)]/60 focus:border-[var(--color-tertiary)] focus:bg-[var(--color-panel)] focus:ring-4 focus:ring-[var(--color-tertiary)]/15"
          />
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto" role="list" aria-label="Daftar percakapan">
        {loading && items.length === 0 ? (
          <p className="px-4 py-8 text-center text-sm text-[var(--color-muted)]">Memuat chat…</p>
        ) : null}

        {!loading && items.length === 0 ? (
          <div className="px-4 py-10 text-center">
            <p className="text-sm font-semibold text-[var(--color-ink)]">Belum ada percakapan</p>
            <p className="mt-1 text-xs text-[var(--color-muted)]">
              Mulai chat baru dengan admin lain.
            </p>
          </div>
        ) : null}

        {items.map((item) => {
          const active = item.id === activeId
          const unread = item.unread_count > 0
          return (
            <button
              key={item.id}
              type="button"
              role="listitem"
              onClick={() => onSelect(item.id)}
              className={cn(
                'flex w-full items-start gap-3 border-b border-[var(--color-line)]/70 px-4 py-3 text-left transition',
                active
                  ? 'bg-[var(--color-accent-soft)]'
                  : 'hover:bg-[var(--color-surface)]',
              )}
              aria-current={active ? 'true' : undefined}
            >
              <span
                className={cn(
                  'mt-0.5 flex h-11 w-11 shrink-0 items-center justify-center rounded-full text-sm font-bold',
                  item.type === 'GROUP'
                    ? 'bg-[var(--color-tertiary)] text-white'
                    : item.type === 'WHATSAPP'
                      ? 'bg-[#128C7E] text-white'
                      : 'bg-[var(--color-accent)] text-white',
                )}
                aria-hidden
              >
                {item.type === 'GROUP' ? (
                  <span className="material-symbols-outlined text-[20px]">groups</span>
                ) : item.type === 'WHATSAPP' ? (
                  <span className="material-symbols-outlined text-[20px]">chat</span>
                ) : (
                  initialsFromName(item.display_name)
                )}
              </span>

              <span className="min-w-0 flex-1">
                <span className="flex items-center justify-between gap-2">
                  <span className="flex min-w-0 items-center gap-1.5">
                    <span
                      className={cn(
                        'truncate text-sm',
                        unread ? 'font-bold text-[var(--color-ink)]' : 'font-semibold text-[var(--color-ink)]',
                      )}
                    >
                      {item.display_name}
                    </span>
                    {/* {item.type === 'WHATSAPP' ? (
                      <span className="shrink-0 rounded-md bg-[#DCF8C6] px-1.5 py-0.5 text-[9px] font-bold tracking-wide text-[#075E54]">
                        WA
                      </span>
                    ) : null} */}
                  </span>
                  <span className="shrink-0 text-[11px] text-[var(--color-muted)]">
                    {formatChatListTime(item.last_message_at)}
                  </span>
                </span>
                <span className="mt-0.5 flex items-center justify-between gap-2">
                  <span
                    className={cn(
                      'truncate text-xs',
                      unread ? 'font-semibold text-[var(--color-ink)]' : 'text-[var(--color-muted)]',
                    )}
                  >
                    {item.last_message_preview || 'Belum ada pesan'}
                  </span>
                  {unread ? (
                    <span className="inline-flex min-w-5 shrink-0 items-center justify-center rounded-full bg-[var(--color-secondary)] px-1.5 py-0.5 text-[10px] font-bold text-[var(--color-secondary-ink)]">
                      {item.unread_count > 99 ? '99+' : item.unread_count}
                    </span>
                  ) : null}
                </span>
              </span>
            </button>
          )
        })}
      </div>
    </div>
  )
}
