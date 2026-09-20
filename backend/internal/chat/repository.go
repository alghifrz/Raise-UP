package chat

import (
	"context"
	"errors"
	"fmt"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides chat persistence via sqlc, including transactions.
type Repository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

// NewRepository creates a chat repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, q: db.New(pool)}
}

// GetConversationByID returns a conversation by UUID string.
func (r *Repository) GetConversationByID(ctx context.Context, id string) (db.Conversation, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.Conversation{}, ErrInvalidRequest
	}
	item, err := r.q.GetConversationByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Conversation{}, ErrNotFound
		}
		return db.Conversation{}, fmt.Errorf("get conversation by id: %w", err)
	}
	return item, nil
}

// GetWhatsAppConversationByPhone returns a WHATSAPP thread by E.164 digits.
func (r *Repository) GetWhatsAppConversationByPhone(ctx context.Context, phone string) (db.Conversation, error) {
	item, err := r.q.GetWhatsAppConversationByPhone(ctx, &phone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Conversation{}, ErrNotFound
		}
		return db.Conversation{}, fmt.Errorf("get whatsapp conversation: %w", err)
	}
	return item, nil
}

// FindDirectConversationBetween returns an existing DIRECT conversation between two users.
func (r *Repository) FindDirectConversationBetween(ctx context.Context, userID, otherUserID string) (db.Conversation, error) {
	a, err := uuidutil.FromString(userID)
	if err != nil {
		return db.Conversation{}, ErrInvalidRequest
	}
	b, err := uuidutil.FromString(otherUserID)
	if err != nil {
		return db.Conversation{}, ErrInvalidRequest
	}
	item, err := r.q.FindDirectConversationBetween(ctx, db.FindDirectConversationBetweenParams{
		UserID:   a,
		UserID_2: b,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Conversation{}, ErrNotFound
		}
		return db.Conversation{}, fmt.Errorf("find direct conversation: %w", err)
	}
	return item, nil
}

// GetParticipant returns a participant row if the user belongs to the conversation.
func (r *Repository) GetParticipant(ctx context.Context, conversationID, userID string) (db.ConversationParticipant, error) {
	pgConversationID, err := uuidutil.FromString(conversationID)
	if err != nil {
		return db.ConversationParticipant{}, ErrInvalidRequest
	}
	pgUserID, err := uuidutil.FromString(userID)
	if err != nil {
		return db.ConversationParticipant{}, ErrInvalidRequest
	}
	item, err := r.q.GetConversationParticipant(ctx, db.GetConversationParticipantParams{
		ConversationID: pgConversationID,
		UserID:         pgUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.ConversationParticipant{}, ErrForbidden
		}
		return db.ConversationParticipant{}, fmt.Errorf("get conversation participant: %w", err)
	}
	return item, nil
}

// EnsureParticipant adds the user to the conversation if missing.
func (r *Repository) EnsureParticipant(ctx context.Context, conversationID, userID string, lastReadAt pgtype.Timestamptz) error {
	pgConversationID, err := uuidutil.FromString(conversationID)
	if err != nil {
		return ErrInvalidRequest
	}
	pgUserID, err := uuidutil.FromString(userID)
	if err != nil {
		return ErrInvalidRequest
	}
	if err := r.q.EnsureConversationParticipant(ctx, db.EnsureConversationParticipantParams{
		ConversationID: pgConversationID,
		UserID:         pgUserID,
		LastReadAt:     lastReadAt,
	}); err != nil {
		return fmt.Errorf("ensure participant: %w", err)
	}
	return nil
}

// ListParticipants returns participants for one conversation.
func (r *Repository) ListParticipants(ctx context.Context, conversationID string) ([]db.ListParticipantsByConversationIDRow, error) {
	pgID, err := uuidutil.FromString(conversationID)
	if err != nil {
		return nil, ErrInvalidRequest
	}
	items, err := r.q.ListParticipantsByConversationID(ctx, pgID)
	if err != nil {
		return nil, fmt.Errorf("list participants: %w", err)
	}
	return items, nil
}

// ListParticipantsByIDs returns participants for many conversations.
func (r *Repository) ListParticipantsByIDs(ctx context.Context, conversationIDs []string) ([]db.ListParticipantsByConversationIDsRow, error) {
	ids := make([]pgtype.UUID, 0, len(conversationIDs))
	for _, id := range conversationIDs {
		pgID, err := uuidutil.FromString(id)
		if err != nil {
			return nil, ErrInvalidRequest
		}
		ids = append(ids, pgID)
	}
	if len(ids) == 0 {
		return []db.ListParticipantsByConversationIDsRow{}, nil
	}
	items, err := r.q.ListParticipantsByConversationIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("list participants by ids: %w", err)
	}
	return items, nil
}

// ListMyConversations returns inbox rows for a user (including all WHATSAPP threads).
func (r *Repository) ListMyConversations(ctx context.Context, userID, search string, limit, offset int32) ([]db.ListMyConversationsRow, error) {
	pgUserID, err := uuidutil.FromString(userID)
	if err != nil {
		return nil, ErrInvalidRequest
	}
	items, err := r.q.ListMyConversations(ctx, db.ListMyConversationsParams{
		UserID:      pgUserID,
		Search:      search,
		LimitCount:  limit,
		OffsetCount: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list my conversations: %w", err)
	}
	return items, nil
}

// CountMyConversations returns total inbox conversations for a user.
func (r *Repository) CountMyConversations(ctx context.Context, userID, search string) (int64, error) {
	pgUserID, err := uuidutil.FromString(userID)
	if err != nil {
		return 0, ErrInvalidRequest
	}
	total, err := r.q.CountMyConversations(ctx, db.CountMyConversationsParams{
		UserID: pgUserID,
		Search: search,
	})
	if err != nil {
		return 0, fmt.Errorf("count my conversations: %w", err)
	}
	return total, nil
}

// CreateDirectConversation creates a DIRECT conversation with both participants.
func (r *Repository) CreateDirectConversation(ctx context.Context, createdBy, otherUserID string, joinedAt pgtype.Timestamptz) (db.Conversation, error) {
	createdByID, err := uuidutil.FromString(createdBy)
	if err != nil {
		return db.Conversation{}, ErrInvalidRequest
	}
	otherID, err := uuidutil.FromString(otherUserID)
	if err != nil {
		return db.Conversation{}, ErrInvalidRequest
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return db.Conversation{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)
	item, err := q.CreateConversation(ctx, db.CreateConversationParams{
		Type:           db.ConversationTypeDIRECT,
		Title:          nil,
		CreatedBy:      createdByID,
		WaContactPhone: nil,
		WaContactName:  nil,
		ResidentID:     pgtype.UUID{},
	})
	if err != nil {
		return db.Conversation{}, fmt.Errorf("create direct conversation: %w", err)
	}

	for _, userID := range []pgtype.UUID{createdByID, otherID} {
		if err := q.AddConversationParticipant(ctx, db.AddConversationParticipantParams{
			ConversationID: item.ID,
			UserID:         userID,
			LastReadAt:     joinedAt,
		}); err != nil {
			return db.Conversation{}, fmt.Errorf("add direct participant: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Conversation{}, fmt.Errorf("commit direct conversation: %w", err)
	}
	return item, nil
}

// CreateGroupConversation creates a GROUP conversation with participants.
func (r *Repository) CreateGroupConversation(ctx context.Context, createdBy, title string, participantIDs []string, joinedAt pgtype.Timestamptz) (db.Conversation, error) {
	createdByID, err := uuidutil.FromString(createdBy)
	if err != nil {
		return db.Conversation{}, ErrInvalidRequest
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return db.Conversation{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)
	titleCopy := title
	item, err := q.CreateConversation(ctx, db.CreateConversationParams{
		Type:           db.ConversationTypeGROUP,
		Title:          &titleCopy,
		CreatedBy:      createdByID,
		WaContactPhone: nil,
		WaContactName:  nil,
		ResidentID:     pgtype.UUID{},
	})
	if err != nil {
		return db.Conversation{}, fmt.Errorf("create group conversation: %w", err)
	}

	seen := map[string]struct{}{createdBy: {}}
	ids := []string{createdBy}
	for _, id := range participantIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	for _, id := range ids {
		pgUserID, err := uuidutil.FromString(id)
		if err != nil {
			return db.Conversation{}, ErrInvalidRequest
		}
		if err := q.AddConversationParticipant(ctx, db.AddConversationParticipantParams{
			ConversationID: item.ID,
			UserID:         pgUserID,
			LastReadAt:     joinedAt,
		}); err != nil {
			return db.Conversation{}, fmt.Errorf("add group participant: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Conversation{}, fmt.Errorf("commit group conversation: %w", err)
	}
	return item, nil
}

// CreateWhatsAppConversation creates a WHATSAPP thread for a phone contact.
func (r *Repository) CreateWhatsAppConversation(
	ctx context.Context,
	createdBy string,
	phone string,
	contactName *string,
	residentID pgtype.UUID,
) (db.Conversation, error) {
	var createdByID pgtype.UUID
	if createdBy != "" {
		id, err := uuidutil.FromString(createdBy)
		if err != nil {
			return db.Conversation{}, ErrInvalidRequest
		}
		createdByID = id
	}

	phoneCopy := phone
	item, err := r.q.CreateConversation(ctx, db.CreateConversationParams{
		Type:           db.ConversationTypeWHATSAPP,
		Title:          nil,
		CreatedBy:      createdByID,
		WaContactPhone: &phoneCopy,
		WaContactName:  contactName,
		ResidentID:     residentID,
	})
	if err != nil {
		return db.Conversation{}, fmt.Errorf("create whatsapp conversation: %w", err)
	}
	return item, nil
}

// UpdateWhatsAppProfile updates contact name / resident link.
func (r *Repository) UpdateWhatsAppProfile(ctx context.Context, conversationID string, contactName *string, residentID pgtype.UUID) (db.Conversation, error) {
	pgID, err := uuidutil.FromString(conversationID)
	if err != nil {
		return db.Conversation{}, ErrInvalidRequest
	}
	item, err := r.q.UpdateWhatsAppContactProfile(ctx, db.UpdateWhatsAppContactProfileParams{
		ID:            pgID,
		WaContactName: contactName,
		ResidentID:    residentID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Conversation{}, ErrNotFound
		}
		return db.Conversation{}, fmt.Errorf("update whatsapp profile: %w", err)
	}
	return item, nil
}

// CreateMessage inserts a message, updates conversation preview, and marks sender as read when USER.
func (r *Repository) CreateMessage(
	ctx context.Context,
	conversationID string,
	senderID pgtype.UUID,
	senderKind db.MessageSenderKind,
	body string,
	waMessageID *string,
	waStatus *string,
	sentAt pgtype.Timestamptz,
	preview string,
) (db.Message, db.Conversation, error) {
	pgConversationID, err := uuidutil.FromString(conversationID)
	if err != nil {
		return db.Message{}, db.Conversation{}, ErrInvalidRequest
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return db.Message{}, db.Conversation{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)
	msg, err := q.CreateMessage(ctx, db.CreateMessageParams{
		ConversationID: pgConversationID,
		SenderID:       senderID,
		SenderKind:     senderKind,
		Body:           body,
		WaMessageID:    waMessageID,
		WaStatus:       waStatus,
	})
	if err != nil {
		return db.Message{}, db.Conversation{}, fmt.Errorf("create message: %w", err)
	}

	conversation, err := q.TouchConversationLastMessage(ctx, db.TouchConversationLastMessageParams{
		ID:                 pgConversationID,
		LastMessageAt:      sentAt,
		LastMessagePreview: preview,
	})
	if err != nil {
		return db.Message{}, db.Conversation{}, fmt.Errorf("touch conversation: %w", err)
	}

	if senderKind == db.MessageSenderKindUSER && senderID.Valid {
		if _, err := q.MarkConversationRead(ctx, db.MarkConversationReadParams{
			ConversationID: pgConversationID,
			UserID:         senderID,
			LastReadAt:     sentAt,
		}); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return db.Message{}, db.Conversation{}, fmt.Errorf("mark sender read: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Message{}, db.Conversation{}, fmt.Errorf("commit message: %w", err)
	}
	return msg, conversation, nil
}

// GetMessageByWaMessageID looks up a message by Meta id.
func (r *Repository) GetMessageByWaMessageID(ctx context.Context, waMessageID string) (db.Message, error) {
	item, err := r.q.GetMessageByWaMessageID(ctx, &waMessageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Message{}, ErrNotFound
		}
		return db.Message{}, fmt.Errorf("get message by wa id: %w", err)
	}
	return item, nil
}

// UpdateMessageWaStatus updates delivery status by Meta message id.
func (r *Repository) UpdateMessageWaStatus(ctx context.Context, waMessageID, status string) (db.Message, error) {
	item, err := r.q.UpdateMessageWaStatus(ctx, db.UpdateMessageWaStatusParams{
		WaMessageID: &waMessageID,
		WaStatus:    &status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Message{}, ErrNotFound
		}
		return db.Message{}, fmt.Errorf("update wa status: %w", err)
	}
	return item, nil
}

// ListMessages returns a page of messages for a conversation (newest first).
func (r *Repository) ListMessages(ctx context.Context, conversationID string, limit, offset int32) ([]db.Message, error) {
	pgID, err := uuidutil.FromString(conversationID)
	if err != nil {
		return nil, ErrInvalidRequest
	}
	items, err := r.q.ListMessagesByConversation(ctx, db.ListMessagesByConversationParams{
		ConversationID: pgID,
		LimitCount:     limit,
		OffsetCount:    offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	return items, nil
}

// CountMessages returns total messages in a conversation.
func (r *Repository) CountMessages(ctx context.Context, conversationID string) (int64, error) {
	pgID, err := uuidutil.FromString(conversationID)
	if err != nil {
		return 0, ErrInvalidRequest
	}
	total, err := r.q.CountMessagesByConversation(ctx, pgID)
	if err != nil {
		return 0, fmt.Errorf("count messages: %w", err)
	}
	return total, nil
}

// CountUnread returns unread messages for a user in a conversation.
func (r *Repository) CountUnread(ctx context.Context, conversationID, userID string) (int64, error) {
	pgConversationID, err := uuidutil.FromString(conversationID)
	if err != nil {
		return 0, ErrInvalidRequest
	}
	pgUserID, err := uuidutil.FromString(userID)
	if err != nil {
		return 0, ErrInvalidRequest
	}
	total, err := r.q.CountUnreadMessages(ctx, db.CountUnreadMessagesParams{
		ConversationID: pgConversationID,
		UserID:         pgUserID,
	})
	if err != nil {
		return 0, fmt.Errorf("count unread messages: %w", err)
	}
	return total, nil
}

// MarkRead updates the user's last_read_at for a conversation.
func (r *Repository) MarkRead(ctx context.Context, conversationID, userID string, readAt pgtype.Timestamptz) (db.ConversationParticipant, error) {
	pgConversationID, err := uuidutil.FromString(conversationID)
	if err != nil {
		return db.ConversationParticipant{}, ErrInvalidRequest
	}
	pgUserID, err := uuidutil.FromString(userID)
	if err != nil {
		return db.ConversationParticipant{}, ErrInvalidRequest
	}
	item, err := r.q.MarkConversationRead(ctx, db.MarkConversationReadParams{
		ConversationID: pgConversationID,
		UserID:         pgUserID,
		LastReadAt:     readAt,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.ConversationParticipant{}, ErrForbidden
		}
		return db.ConversationParticipant{}, fmt.Errorf("mark conversation read: %w", err)
	}
	return item, nil
}

// ListContacts returns other users available for chat.
func (r *Repository) ListContacts(ctx context.Context, userID, search string, limit, offset int32) ([]db.ListChatContactsRow, error) {
	pgUserID, err := uuidutil.FromString(userID)
	if err != nil {
		return nil, ErrInvalidRequest
	}
	items, err := r.q.ListChatContacts(ctx, db.ListChatContactsParams{
		UserID:      pgUserID,
		Search:      search,
		LimitCount:  limit,
		OffsetCount: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list chat contacts: %w", err)
	}
	return items, nil
}

// CountContacts returns total other users available for chat.
func (r *Repository) CountContacts(ctx context.Context, userID, search string) (int64, error) {
	pgUserID, err := uuidutil.FromString(userID)
	if err != nil {
		return 0, ErrInvalidRequest
	}
	total, err := r.q.CountChatContacts(ctx, db.CountChatContactsParams{
		UserID: pgUserID,
		Search: search,
	})
	if err != nil {
		return 0, fmt.Errorf("count chat contacts: %w", err)
	}
	return total, nil
}

// GetChatUserByID returns a user summary for chat validation.
func (r *Repository) GetChatUserByID(ctx context.Context, id string) (db.GetChatUserByIDRow, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.GetChatUserByIDRow{}, ErrInvalidRequest
	}
	item, err := r.q.GetChatUserByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetChatUserByIDRow{}, ErrUserNotFound
		}
		return db.GetChatUserByIDRow{}, fmt.Errorf("get chat user: %w", err)
	}
	return item, nil
}

// FindResidentByPhones returns the first resident matching any phone candidate.
func (r *Repository) FindResidentByPhones(ctx context.Context, phones []string) (db.FindResidentByPhoneCandidatesRow, error) {
	if len(phones) == 0 {
		return db.FindResidentByPhoneCandidatesRow{}, ErrNotFound
	}
	item, err := r.q.FindResidentByPhoneCandidates(ctx, phones)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.FindResidentByPhoneCandidatesRow{}, ErrNotFound
		}
		return db.FindResidentByPhoneCandidatesRow{}, fmt.Errorf("find resident by phones: %w", err)
	}
	return item, nil
}

// GetResidentByID returns a resident by UUID string.
func (r *Repository) GetResidentByID(ctx context.Context, id string) (db.Resident, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.Resident{}, ErrInvalidRequest
	}
	item, err := r.q.GetResidentByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Resident{}, ErrNotFound
		}
		return db.Resident{}, fmt.Errorf("get resident by id: %w", err)
	}
	return item, nil
}
