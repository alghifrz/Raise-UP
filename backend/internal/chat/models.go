package chat

import db "github.com/diuk/raiseup/db/generated"

// CreateRequest is the JSON body for POST /api/v1/conversations.
//
// DIRECT: participant_id
// GROUP: title + participant_ids
// WHATSAPP: phone (and optional contact_name / resident_id)
type CreateRequest struct {
	Type           string   `json:"type"`
	ParticipantID  string   `json:"participant_id"`
	Title          string   `json:"title"`
	ParticipantIDs []string `json:"participant_ids"`
	Phone          string   `json:"phone"`
	ContactName    string   `json:"contact_name"`
	ResidentID     string   `json:"resident_id"`
}

// SendMessageRequest is the JSON body for POST /api/v1/conversations/:id/messages.
type SendMessageRequest struct {
	Body string `json:"body"`
}

// ListMeta is pagination metadata.
type ListMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ConversationListResult is the service result for listing conversations.
type ConversationListResult struct {
	Items []Conversation
	Meta  ListMeta
}

// MessageListResult is the service result for listing messages.
type MessageListResult struct {
	Items []Message
	Meta  ListMeta
}

// ContactListResult is the service result for listing chat contacts.
type ContactListResult struct {
	Items []Contact
	Meta  ListMeta
}

// Contact is a user available for starting a chat.
type Contact struct {
	ID    string      `json:"id"`
	Email string      `json:"email"`
	Name  string      `json:"name"`
	Role  db.UserRole `json:"role"`
}

// Participant is a conversation member (API-safe).
type Participant struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Conversation is the API-safe conversation representation for inbox + detail.
type Conversation struct {
	ID                 string              `json:"id"`
	Type               db.ConversationType `json:"type"`
	Title              *string             `json:"title"`
	DisplayName        string              `json:"display_name"`
	CreatedBy          *string             `json:"created_by"`
	WaContactPhone     *string             `json:"wa_contact_phone"`
	WaContactName      *string             `json:"wa_contact_name"`
	ResidentID         *string             `json:"resident_id"`
	ResidentName       *string             `json:"resident_name"`
	LastMessageAt      *string             `json:"last_message_at"`
	LastMessagePreview string              `json:"last_message_preview"`
	UnreadCount        int64               `json:"unread_count"`
	Participants       []Participant       `json:"participants"`
	CreatedAt          string              `json:"created_at"`
	UpdatedAt          string              `json:"updated_at"`
}

// Message is the API-safe message representation.
type Message struct {
	ID             string                `json:"id"`
	ConversationID string                `json:"conversation_id"`
	SenderID       *string               `json:"sender_id"`
	SenderKind     db.MessageSenderKind  `json:"sender_kind"`
	Body           string                `json:"body"`
	WaMessageID    *string               `json:"wa_message_id"`
	WaStatus       *string               `json:"wa_status"`
	CreatedAt      string                `json:"created_at"`
}

// ReadResult is returned after marking a conversation as read.
type ReadResult struct {
	ConversationID string `json:"conversation_id"`
	LastReadAt     string `json:"last_read_at"`
}
