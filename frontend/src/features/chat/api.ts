import { apiRequest, apiRequestWithMeta, toQueryString } from '../../lib/api/client'
import type { PaginationMeta } from '../../lib/api/types'
import type {
  ChatContact,
  ChatMessage,
  ContactFilters,
  ContactListResponse,
  Conversation,
  ConversationFilters,
  ConversationListResponse,
  CreateConversationRequest,
  MessageFilters,
  MessageListResponse,
  ReadResult,
  SendMessageRequest,
  WhatsAppNotificationRequest,
  WhatsAppNotificationResult,
  WhatsAppStatus,
} from './types'

export function listConversations(
  filters: ConversationFilters,
  signal?: AbortSignal,
): Promise<ConversationListResponse> {
  const qs = toQueryString({
    page: filters.page,
    page_size: filters.page_size,
    search: filters.search || undefined,
  })

  return apiRequestWithMeta<Conversation[], PaginationMeta>(`/api/v1/conversations${qs}`, {
    signal,
  }).then((result) => ({
    items: result.data,
    meta: result.meta,
  }))
}

export function getConversation(id: string, signal?: AbortSignal): Promise<Conversation> {
  return apiRequest<Conversation>(`/api/v1/conversations/${id}`, { signal })
}

export function createConversation(payload: CreateConversationRequest): Promise<Conversation> {
  return apiRequest<Conversation>('/api/v1/conversations', {
    method: 'POST',
    body: payload,
  })
}

export function listMessages(
  conversationId: string,
  filters: MessageFilters,
  signal?: AbortSignal,
): Promise<MessageListResponse> {
  const qs = toQueryString({
    page: filters.page,
    page_size: filters.page_size,
  })

  return apiRequestWithMeta<ChatMessage[], PaginationMeta>(
    `/api/v1/conversations/${conversationId}/messages${qs}`,
    { signal },
  ).then((result) => ({
    items: result.data,
    meta: result.meta,
  }))
}

export function sendMessage(
  conversationId: string,
  payload: SendMessageRequest,
): Promise<ChatMessage> {
  return apiRequest<ChatMessage>(`/api/v1/conversations/${conversationId}/messages`, {
    method: 'POST',
    body: payload,
  })
}

export function markConversationRead(conversationId: string): Promise<ReadResult> {
  return apiRequest<ReadResult>(`/api/v1/conversations/${conversationId}/read`, {
    method: 'POST',
  })
}

export function listChatContacts(
  filters: ContactFilters,
  signal?: AbortSignal,
): Promise<ContactListResponse> {
  const qs = toQueryString({
    page: filters.page,
    page_size: filters.page_size,
    search: filters.search || undefined,
  })

  return apiRequestWithMeta<ChatContact[], PaginationMeta>(
    `/api/v1/conversations/contacts${qs}`,
    { signal },
  ).then((result) => ({
    items: result.data,
    meta: result.meta,
  }))
}

export function getWhatsAppStatus(signal?: AbortSignal): Promise<WhatsAppStatus> {
  return apiRequest<WhatsAppStatus>('/api/v1/whatsapp/status', { signal })
}

export function sendWhatsAppNotification(
  payload: WhatsAppNotificationRequest,
): Promise<WhatsAppNotificationResult> {
  return apiRequest<WhatsAppNotificationResult>('/api/v1/whatsapp/notifications', {
    method: 'POST',
    body: payload,
  })
}
