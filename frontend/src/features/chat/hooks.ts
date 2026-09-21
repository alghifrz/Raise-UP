import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createConversation,
  getConversation,
  getWhatsAppStatus,
  listChatContacts,
  listConversations,
  listMessages,
  markConversationRead,
  sendMessage,
  sendWhatsAppNotification,
} from './api'
import type {
  ChatMessage,
  ContactFilters,
  ConversationFilters,
  CreateConversationRequest,
  MessageFilters,
  MessageListResponse,
  SendMessageRequest,
  WhatsAppNotificationRequest,
} from './types'

export const chatQueryKey = ['chat'] as const

/** Inbox / badge poll — keep list + unread badge on the same cadence. */
const POLL_INBOX_MS = 2_500
/** Open thread poll — slightly faster so inbound bubbles appear sooner. */
const POLL_THREAD_MS = 2_000

export function conversationsListQueryKey(filters: ConversationFilters) {
  return [...chatQueryKey, 'conversations', filters] as const
}

export function conversationDetailQueryKey(id: string) {
  return [...chatQueryKey, 'conversation', id] as const
}

export function messagesListQueryKey(conversationId: string, filters: MessageFilters) {
  return [...chatQueryKey, 'messages', conversationId, filters] as const
}

export function contactsListQueryKey(filters: ContactFilters) {
  return [...chatQueryKey, 'contacts', filters] as const
}

export function whatsappStatusQueryKey() {
  return [...chatQueryKey, 'whatsapp-status'] as const
}

/**
 * Shared filters for sidebar unread badge and the default (no-search) inbox.
 * Same key → one network request, badge and list stay in sync.
 */
export const chatUnreadFilters: ConversationFilters = {
  page: 1,
  page_size: 100,
  search: '',
}

export function chatUnreadTotalQueryKey() {
  return conversationsListQueryKey(chatUnreadFilters)
}

export function useConversations(filters: ConversationFilters) {
  return useQuery({
    queryKey: conversationsListQueryKey(filters),
    queryFn: ({ signal }) => listConversations(filters, signal),
    placeholderData: (previous) => previous,
    refetchInterval: POLL_INBOX_MS,
    refetchOnWindowFocus: true,
    staleTime: 0,
  })
}

/** Total unread across conversations — shares cache with the default inbox list. */
export function useChatUnreadTotal() {
  return useQuery({
    queryKey: conversationsListQueryKey(chatUnreadFilters),
    queryFn: ({ signal }) => listConversations(chatUnreadFilters, signal),
    select: (result) => result.items.reduce((sum, item) => sum + item.unread_count, 0),
    refetchInterval: POLL_INBOX_MS,
    refetchOnWindowFocus: true,
    staleTime: 0,
  })
}

export function useConversation(id: string | undefined) {
  return useQuery({
    queryKey: conversationDetailQueryKey(id ?? ''),
    queryFn: ({ signal }) => getConversation(id!, signal),
    enabled: Boolean(id),
    refetchInterval: POLL_THREAD_MS,
    refetchOnWindowFocus: true,
    staleTime: 0,
  })
}

export function useMessages(conversationId: string | undefined, filters: MessageFilters) {
  return useQuery({
    queryKey: messagesListQueryKey(conversationId ?? '', filters),
    queryFn: ({ signal }) => listMessages(conversationId!, filters, signal),
    enabled: Boolean(conversationId),
    placeholderData: (previous) => previous,
    refetchInterval: POLL_THREAD_MS,
    refetchOnWindowFocus: true,
    staleTime: 0,
  })
}

export function useChatContacts(filters: ContactFilters, enabled = true) {
  return useQuery({
    queryKey: contactsListQueryKey(filters),
    queryFn: ({ signal }) => listChatContacts(filters, signal),
    enabled,
    placeholderData: (previous) => previous,
  })
}

export function useWhatsAppStatus() {
  return useQuery({
    queryKey: whatsappStatusQueryKey(),
    queryFn: ({ signal }) => getWhatsAppStatus(signal),
  })
}

function useInvalidateChat() {
  const queryClient = useQueryClient()

  return async (conversationId?: string) => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: chatQueryKey }),
      conversationId
        ? queryClient.invalidateQueries({ queryKey: conversationDetailQueryKey(conversationId) })
        : Promise.resolve(),
    ])
  }
}

export function useCreateConversation() {
  const invalidate = useInvalidateChat()

  return useMutation({
    mutationFn: (payload: CreateConversationRequest) => createConversation(payload),
    onSuccess: async (conversation) => {
      await invalidate(conversation.id)
    },
  })
}

export function useSendMessage(conversationId: string | undefined) {
  const invalidate = useInvalidateChat()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (payload: SendMessageRequest) => {
      if (!conversationId) {
        throw new Error('conversation id required')
      }
      return sendMessage(conversationId, payload)
    },
    onSuccess: async (message: ChatMessage) => {
      if (conversationId) {
        const key = messagesListQueryKey(conversationId, { page: 1, page_size: 100 })
        queryClient.setQueryData(key, (previous: MessageListResponse | undefined) => {
          if (!previous) {
            return previous
          }
          if (previous.items.some((item) => item.id === message.id)) {
            return previous
          }
          return {
            ...previous,
            items: [message, ...previous.items],
          }
        })
      }
      await invalidate(conversationId)
    },
  })
}

export function useMarkConversationRead() {
  const invalidate = useInvalidateChat()

  return useMutation({
    mutationFn: (conversationId: string) => markConversationRead(conversationId),
    onSuccess: async (result) => {
      await invalidate(result.conversation_id)
    },
  })
}

export function useSendWhatsAppNotification() {
  return useMutation({
    mutationFn: (payload: WhatsAppNotificationRequest) => sendWhatsAppNotification(payload),
  })
}
