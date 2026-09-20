import type { PaginationMeta } from '../../lib/api/types'

export type ConversationType = 'DIRECT' | 'GROUP' | 'WHATSAPP'
export type MessageSenderKind = 'USER' | 'CONTACT' | 'SYSTEM'

export type ChatParticipant = {
  id: string
  name: string
  email: string
}

export type Conversation = {
  id: string
  type: ConversationType
  title: string | null
  display_name: string
  created_by: string | null
  wa_contact_phone: string | null
  wa_contact_name: string | null
  resident_id: string | null
  resident_name: string | null
  last_message_at: string | null
  last_message_preview: string
  unread_count: number
  participants: ChatParticipant[]
  created_at: string
  updated_at: string
}

export type ChatMessage = {
  id: string
  conversation_id: string
  sender_id: string | null
  sender_kind: MessageSenderKind
  body: string
  wa_message_id: string | null
  wa_status: string | null
  created_at: string
}

export type ChatContact = {
  id: string
  email: string
  name: string
  role: string
}

export type ConversationFilters = {
  page: number
  page_size: number
  search: string
}

export type MessageFilters = {
  page: number
  page_size: number
}

export type ContactFilters = {
  page: number
  page_size: number
  search: string
}

export type ConversationListResponse = {
  items: Conversation[]
  meta: PaginationMeta
}

export type MessageListResponse = {
  items: ChatMessage[]
  meta: PaginationMeta
}

export type ContactListResponse = {
  items: ChatContact[]
  meta: PaginationMeta
}

export type CreateDirectConversationRequest = {
  type: 'DIRECT'
  participant_id: string
}

export type CreateGroupConversationRequest = {
  type: 'GROUP'
  title: string
  participant_ids: string[]
}

export type CreateWhatsAppConversationRequest = {
  type: 'WHATSAPP'
  phone: string
  contact_name?: string
  resident_id?: string
}

export type CreateConversationRequest =
  | CreateDirectConversationRequest
  | CreateGroupConversationRequest
  | CreateWhatsAppConversationRequest

export type SendMessageRequest = {
  body: string
}

export type ReadResult = {
  conversation_id: string
  last_read_at: string
}

export type WhatsAppNotificationRequest = {
  to: string
  type: 'text' | 'template'
  body?: string
  template_name?: string
  language?: string
}

export type WhatsAppNotificationResult = {
  to: string
  message_id: string
  type: string
}

export type WhatsAppStatus = {
  enabled: boolean
}

export function initialsFromName(name: string): string {
  const cleaned = name.replace(/\s*\([^)]*\)\s*$/, '').trim()
  const parts = cleaned.split(/\s+/).filter(Boolean)
  if (parts.length === 0) {
    return '?'
  }
  if (parts.length === 1) {
    return parts[0]!.slice(0, 2).toUpperCase()
  }
  return `${parts[0]!.charAt(0)}${parts[1]!.charAt(0)}`.toUpperCase()
}

const chatTimeFormatter = new Intl.DateTimeFormat('id-ID', {
  hour: '2-digit',
  minute: '2-digit',
  timeZone: 'Asia/Jakarta',
})

const chatDayFormatter = new Intl.DateTimeFormat('id-ID', {
  weekday: 'short',
  day: 'numeric',
  month: 'short',
  timeZone: 'Asia/Jakarta',
})

/** Compact inbox timestamp: time today, otherwise short date. */
export function formatChatListTime(value: string | null): string {
  if (!value) {
    return ''
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  const now = new Date()
  const sameDay =
    date.getFullYear() === now.getFullYear() &&
    date.getMonth() === now.getMonth() &&
    date.getDate() === now.getDate()

  return sameDay ? chatTimeFormatter.format(date) : chatDayFormatter.format(date)
}

export function formatMessageTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return chatTimeFormatter.format(date)
}
