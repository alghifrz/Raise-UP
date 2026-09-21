import { useEffect, useMemo, useRef, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useNavigate, useParams } from 'react-router-dom'
import { InlineAlert } from '../../../components/ui/InlineAlert'
import { useDebouncedValue } from '../../../lib/hooks/useDebouncedValue'
import { cn } from '../../../lib/utils'
import { useAuth } from '../../auth/useAuth'
import { ChatThread } from '../components/ChatThread'
import { ConversationList } from '../components/ConversationList'
import { NewChatModal } from '../components/NewChatModal'
import { NotifyModal } from '../components/NotifyModal'
import { toChatErrorMessage } from '../errors'
import {
  chatUnreadFilters,
  messagesListQueryKey,
  useConversation,
  useConversations,
  useMarkConversationRead,
  useMessages,
  useSendMessage,
  useWhatsAppStatus,
} from '../hooks'
import type { ConversationFilters, MessageFilters } from '../types'

export function ChatPage() {
  const { conversationId } = useParams<{ conversationId?: string }>()
  const navigate = useNavigate()
  const { user } = useAuth()
  const queryClient = useQueryClient()

  const [searchInput, setSearchInput] = useState('')
  const debouncedSearch = useDebouncedValue(searchInput, 300)
  const [newChatOpen, setNewChatOpen] = useState(false)
  const [notifyOpen, setNotifyOpen] = useState(false)
  const [sendError, setSendError] = useState<string | null>(null)

  // Empty search must match chatUnreadFilters so badge + list share one cache entry.
  const conversationFilters = useMemo<ConversationFilters>(() => {
    const search = debouncedSearch.trim()
    if (!search) {
      return chatUnreadFilters
    }
    return {
      page: 1,
      page_size: chatUnreadFilters.page_size,
      search,
    }
  }, [debouncedSearch])

  const messageFilters = useMemo<MessageFilters>(
    () => ({
      page: 1,
      page_size: 100,
    }),
    [],
  )

  const conversationsQuery = useConversations(conversationFilters)
  const conversationQuery = useConversation(conversationId)
  const messagesQuery = useMessages(conversationId, messageFilters)
  const sendMutation = useSendMessage(conversationId)
  const markReadMutation = useMarkConversationRead()
  const waStatusQuery = useWhatsAppStatus()

  const conversations = conversationsQuery.data?.items ?? []
  const messages = messagesQuery.data?.items ?? []
  const whatsappEnabled = Boolean(waStatusQuery.data?.enabled)

  const activeInboxRow = conversations.find((c) => c.id === conversationId)
  const inboxSyncKey = activeInboxRow
    ? `${activeInboxRow.last_message_at ?? ''}:${activeInboxRow.unread_count}:${activeInboxRow.last_message_preview}`
    : ''
  const lastInboxSyncKey = useRef('')

  useEffect(() => {
    if (!conversationId || !inboxSyncKey) {
      return
    }
    if (lastInboxSyncKey.current === inboxSyncKey) {
      return
    }
    const isFirst = lastInboxSyncKey.current === ''
    lastInboxSyncKey.current = inboxSyncKey
    if (isFirst) {
      return
    }
    void queryClient.invalidateQueries({
      queryKey: messagesListQueryKey(conversationId, messageFilters),
    })
  }, [conversationId, inboxSyncKey, messageFilters, queryClient])

  useEffect(() => {
    lastInboxSyncKey.current = ''
  }, [conversationId])

  useEffect(() => {
    if (!conversationId) {
      return
    }
    const item = conversations.find((c) => c.id === conversationId)
    if (!item || item.unread_count <= 0) {
      return
    }
    markReadMutation.mutate(conversationId)
    // Intentionally omit markReadMutation from deps to avoid re-trigger loops.
    // eslint-disable-next-line react-hooks/exhaustive-deps -- mark-on-open only
  }, [conversationId, conversations])

  function selectConversation(id: string) {
    setSendError(null)
    void navigate(`/chat/${id}`)
  }

  function clearConversation() {
    setSendError(null)
    void navigate('/chat')
  }

  async function handleSend(body: string) {
    setSendError(null)
    try {
      await sendMutation.mutateAsync({ body })
    } catch (error) {
      setSendError(toChatErrorMessage(error))
    }
  }

  const showThreadOnMobile = Boolean(conversationId)
  const listError = conversationsQuery.isError
    ? toChatErrorMessage(conversationsQuery.error)
    : null

  return (
    <div className="flex h-full min-h-0 flex-1 overflow-hidden rounded-none border-0 bg-[var(--color-panel)] lg:rounded-3xl lg:border lg:border-[var(--color-line)] lg:shadow-sm">
      <aside
        className={cn(
          'h-full min-h-0 w-full shrink-0 overflow-hidden lg:w-[22rem] xl:w-[24rem]',
          showThreadOnMobile ? 'hidden lg:flex lg:flex-col' : 'flex flex-col',
        )}
      >
        {listError ? (
          <div className="border-b border-[var(--color-line)] p-3">
            <InlineAlert tone="danger">{listError}</InlineAlert>
          </div>
        ) : null}
        <ConversationList
          items={conversations}
          activeId={conversationId}
          search={searchInput}
          onSearchChange={(value) => {
            setSearchInput(value)
          }}
          onSelect={selectConversation}
          onNewChat={() => setNewChatOpen(true)}
          onNotify={() => setNotifyOpen(true)}
          whatsappEnabled={whatsappEnabled}
          loading={conversationsQuery.isLoading}
        />
      </aside>

      <section
        className={cn(
          'h-full min-h-0 min-w-0 flex-1 overflow-hidden',
          showThreadOnMobile ? 'flex flex-col' : 'hidden lg:flex lg:flex-col',
        )}
      >
        {conversationId && conversationQuery.isLoading && !conversationQuery.data ? (
          <div className="flex h-full min-h-0 items-center justify-center bg-[var(--color-surface)]">
            <p className="text-sm text-[var(--color-muted)]">Memuat percakapan…</p>
          </div>
        ) : (
          <ChatThread
            conversation={conversationQuery.data}
            messages={messages}
            currentUserId={user?.id}
            loading={Boolean(conversationId) && messagesQuery.isLoading}
            sending={sendMutation.isPending}
            errorMessage={
              sendError ??
              (conversationQuery.isError ? toChatErrorMessage(conversationQuery.error) : null)
            }
            onSend={handleSend}
            onBack={clearConversation}
          />
        )}
      </section>

      <NewChatModal
        open={newChatOpen}
        onClose={() => setNewChatOpen(false)}
        onCreated={(id) => {
          setNewChatOpen(false)
          selectConversation(id)
        }}
      />

      <NotifyModal
        open={notifyOpen}
        onClose={() => setNotifyOpen(false)}
        enabled={whatsappEnabled}
      />
    </div>
  )
}
