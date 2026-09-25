package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100

	maxMessageBodyLen = 4000
	maxPreviewLen     = 120
	maxGroupTitleLen  = 100
	maxGroupMembers   = 50
)

// Store is the persistence interface used by Service.
type Store interface {
	GetConversationByID(ctx context.Context, id string) (db.Conversation, error)
	GetWhatsAppConversationByPhone(ctx context.Context, phone string) (db.Conversation, error)
	FindDirectConversationBetween(ctx context.Context, userID, otherUserID string) (db.Conversation, error)
	GetParticipant(ctx context.Context, conversationID, userID string) (db.ConversationParticipant, error)
	EnsureParticipant(ctx context.Context, conversationID, userID string, lastReadAt pgtype.Timestamptz) error
	ListParticipants(ctx context.Context, conversationID string) ([]db.ListParticipantsByConversationIDRow, error)
	ListParticipantsByIDs(ctx context.Context, conversationIDs []string) ([]db.ListParticipantsByConversationIDsRow, error)
	ListMyConversations(ctx context.Context, userID, search string, limit, offset int32) ([]db.ListMyConversationsRow, error)
	CountMyConversations(ctx context.Context, userID, search string) (int64, error)
	CreateDirectConversation(ctx context.Context, createdBy, otherUserID string, joinedAt pgtype.Timestamptz) (db.Conversation, error)
	CreateGroupConversation(ctx context.Context, createdBy, title string, participantIDs []string, joinedAt pgtype.Timestamptz) (db.Conversation, error)
	CreateWhatsAppConversation(ctx context.Context, createdBy, phone string, contactName *string, residentID pgtype.UUID) (db.Conversation, error)
	UpdateWhatsAppProfile(ctx context.Context, conversationID string, contactName *string, residentID pgtype.UUID) (db.Conversation, error)
	CreateMessage(ctx context.Context, conversationID string, senderID pgtype.UUID, senderKind db.MessageSenderKind, body string, waMessageID *string, waStatus *string, sentAt pgtype.Timestamptz, preview string) (db.Message, db.Conversation, error)
	GetMessageByWaMessageID(ctx context.Context, waMessageID string) (db.Message, error)
	UpdateMessageWaStatus(ctx context.Context, waMessageID, status string) (db.Message, error)
	ListMessages(ctx context.Context, conversationID string, limit, offset int32) ([]db.Message, error)
	CountMessages(ctx context.Context, conversationID string) (int64, error)
	CountUnread(ctx context.Context, conversationID, userID string) (int64, error)
	GetMaxOtherLastReadAt(ctx context.Context, conversationID, userID string) (pgtype.Timestamptz, error)
	MarkRead(ctx context.Context, conversationID, userID string, readAt pgtype.Timestamptz) (db.ConversationParticipant, error)
	ListContacts(ctx context.Context, userID, search string, limit, offset int32) ([]db.ListChatContactsRow, error)
	CountContacts(ctx context.Context, userID, search string) (int64, error)
	GetChatUserByID(ctx context.Context, id string) (db.GetChatUserByIDRow, error)
	FindResidentByPhones(ctx context.Context, phones []string) (db.FindResidentByPhoneCandidatesRow, error)
	GetResidentByID(ctx context.Context, id string) (db.Resident, error)
}

// OutboundMessenger sends WhatsApp Cloud API messages and read receipts.
type OutboundMessenger interface {
	Enabled() bool
	SendText(ctx context.Context, to, body string) (waMessageID string, err error)
	MarkAsRead(ctx context.Context, waMessageID string) error
}

// PhoneNormalizer normalizes phone numbers for WhatsApp.
type PhoneNormalizer interface {
	Normalize(raw string) (digits string, err error)
	MatchCandidates(digits string) []string
}

// Service implements chat use cases.
type Service struct {
	store     Store
	messenger OutboundMessenger
	phones    PhoneNormalizer
	now       func() time.Time
}

// NewService creates a chat service.
func NewService(store Store, messenger OutboundMessenger, phones PhoneNormalizer) *Service {
	return &Service{
		store:     store,
		messenger: messenger,
		phones:    phones,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

// ListConversations returns the authenticated user's inbox (plus all WA threads).
func (s *Service) ListConversations(ctx context.Context, userID string, page, pageSize int, search string) (*ConversationListResult, error) {
	if _, err := uuidutil.FromString(userID); err != nil {
		return nil, fmt.Errorf("%w: invalid user id", ErrInvalidRequest)
	}

	page, pageSize = normalizePage(page, pageSize)
	search = strings.TrimSpace(search)
	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	rows, err := s.store.ListMyConversations(ctx, userID, search, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.store.CountMyConversations(ctx, userID, search)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		id, err := uuidutil.ToString(row.ID)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	participantsByConv, err := s.loadParticipantsMap(ctx, ids)
	if err != nil {
		return nil, err
	}

	items := make([]Conversation, 0, len(rows))
	for _, row := range rows {
		id, err := uuidutil.ToString(row.ID)
		if err != nil {
			return nil, err
		}
		mapped, err := toAPIConversationFromListRow(row, userID, participantsByConv[id])
		if err != nil {
			return nil, err
		}
		items = append(items, mapped)
	}

	return &ConversationListResult{Items: items, Meta: buildMeta(page, pageSize, total)}, nil
}

// GetConversation returns one conversation if the user may access it.
func (s *Service) GetConversation(ctx context.Context, userID, conversationID string) (*Conversation, error) {
	item, err := s.store.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeAccess(ctx, userID, item); err != nil {
		return nil, err
	}
	if item.Type == db.ConversationTypeWHATSAPP {
		_ = s.store.EnsureParticipant(ctx, conversationID, userID, pgtype.Timestamptz{})
	}

	participants, err := s.store.ListParticipants(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	apiParticipants, err := mapParticipantRows(participants)
	if err != nil {
		return nil, err
	}
	unread, err := s.store.CountUnread(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	var residentName *string
	if item.ResidentID.Valid {
		residentID, idErr := uuidutil.ToString(item.ResidentID)
		if idErr == nil {
			if resident, resErr := s.store.GetResidentByID(ctx, residentID); resErr == nil {
				name := resident.Name
				residentName = &name
			}
		}
	}
	mapped, err := toAPIConversation(item, userID, apiParticipants, unread, residentName)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// CreateConversation starts a DIRECT, GROUP, or WHATSAPP chat.
func (s *Service) CreateConversation(ctx context.Context, userID string, req CreateRequest) (*Conversation, error) {
	convType := strings.ToUpper(strings.TrimSpace(req.Type))
	now := timestamptz(s.now())

	switch db.ConversationType(convType) {
	case db.ConversationTypeDIRECT:
		return s.createDirect(ctx, userID, strings.TrimSpace(req.ParticipantID), now)
	case db.ConversationTypeGROUP:
		return s.createGroup(ctx, userID, strings.TrimSpace(req.Title), req.ParticipantIDs, now)
	case db.ConversationTypeWHATSAPP:
		return s.createWhatsApp(ctx, userID, strings.TrimSpace(req.Phone), strings.TrimSpace(req.ContactName), strings.TrimSpace(req.ResidentID))
	default:
		return nil, fmt.Errorf("%w: type must be DIRECT, GROUP, or WHATSAPP", ErrInvalidRequest)
	}
}

func (s *Service) createDirect(ctx context.Context, userID, participantID string, now pgtype.Timestamptz) (*Conversation, error) {
	if participantID == "" {
		return nil, fmt.Errorf("%w: participant_id is required", ErrInvalidRequest)
	}
	if _, err := uuidutil.FromString(participantID); err != nil {
		return nil, fmt.Errorf("%w: invalid participant_id", ErrInvalidRequest)
	}
	if participantID == userID {
		return nil, fmt.Errorf("%w: cannot start a direct chat with yourself", ErrInvalidRequest)
	}
	if _, err := s.store.GetChatUserByID(ctx, participantID); err != nil {
		return nil, err
	}

	existing, err := s.store.FindDirectConversationBetween(ctx, userID, participantID)
	if err == nil {
		id, idErr := uuidutil.ToString(existing.ID)
		if idErr != nil {
			return nil, idErr
		}
		return s.GetConversation(ctx, userID, id)
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	created, err := s.store.CreateDirectConversation(ctx, userID, participantID, now)
	if err != nil {
		return nil, err
	}
	id, err := uuidutil.ToString(created.ID)
	if err != nil {
		return nil, err
	}
	return s.GetConversation(ctx, userID, id)
}

func (s *Service) createGroup(ctx context.Context, userID, title string, participantIDs []string, now pgtype.Timestamptz) (*Conversation, error) {
	if title == "" {
		return nil, fmt.Errorf("%w: title is required for GROUP", ErrInvalidRequest)
	}
	if utf8.RuneCountInString(title) > maxGroupTitleLen {
		return nil, fmt.Errorf("%w: title must be at most %d characters", ErrInvalidRequest, maxGroupTitleLen)
	}

	uniqueOthers := make([]string, 0, len(participantIDs))
	seen := map[string]struct{}{userID: {}}
	for _, raw := range participantIDs {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, err := uuidutil.FromString(id); err != nil {
			return nil, fmt.Errorf("%w: invalid participant id", ErrInvalidRequest)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		if _, err := s.store.GetChatUserByID(ctx, id); err != nil {
			return nil, err
		}
		uniqueOthers = append(uniqueOthers, id)
	}

	if len(uniqueOthers) == 0 {
		return nil, fmt.Errorf("%w: at least one other participant is required", ErrInvalidRequest)
	}
	if len(uniqueOthers)+1 > maxGroupMembers {
		return nil, fmt.Errorf("%w: group may have at most %d members", ErrInvalidRequest, maxGroupMembers)
	}

	created, err := s.store.CreateGroupConversation(ctx, userID, title, uniqueOthers, now)
	if err != nil {
		return nil, err
	}
	id, err := uuidutil.ToString(created.ID)
	if err != nil {
		return nil, err
	}
	return s.GetConversation(ctx, userID, id)
}

func (s *Service) createWhatsApp(ctx context.Context, userID, phone, contactName, residentID string) (*Conversation, error) {
	if s.phones == nil {
		return nil, fmt.Errorf("%w: phone normalizer not configured", ErrInvalidRequest)
	}
	digits, err := s.phones.Normalize(phone)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid phone", ErrInvalidRequest)
	}

	existing, err := s.store.GetWhatsAppConversationByPhone(ctx, digits)
	if err == nil {
		id, idErr := uuidutil.ToString(existing.ID)
		if idErr != nil {
			return nil, idErr
		}
		return s.GetConversation(ctx, userID, id)
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	var namePtr *string
	if contactName != "" {
		namePtr = &contactName
	}
	var residentUUID pgtype.UUID
	if residentID != "" {
		pgID, err := uuidutil.FromString(residentID)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid resident_id", ErrInvalidRequest)
		}
		residentUUID = pgID
	} else if row, findErr := s.store.FindResidentByPhones(ctx, s.phones.MatchCandidates(digits)); findErr == nil {
		residentUUID = row.ID
	}

	created, err := s.store.CreateWhatsAppConversation(ctx, userID, digits, namePtr, residentUUID)
	if err != nil {
		return nil, err
	}
	id, err := uuidutil.ToString(created.ID)
	if err != nil {
		return nil, err
	}
	_ = s.store.EnsureParticipant(ctx, id, userID, timestamptz(s.now()))
	return s.GetConversation(ctx, userID, id)
}

// ListMessages returns paginated messages (newest first).
func (s *Service) ListMessages(ctx context.Context, userID, conversationID string, page, pageSize int) (*MessageListResult, error) {
	item, err := s.store.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeAccess(ctx, userID, item); err != nil {
		return nil, err
	}

	page, pageSize = normalizePage(page, pageSize)
	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	rows, err := s.store.ListMessages(ctx, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.store.CountMessages(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	peerReadAt, err := s.store.GetMaxOtherLastReadAt(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}

	items := make([]Message, 0, len(rows))
	for _, row := range rows {
		mapped, err := toAPIMessage(row, messageIsRead(item.Type, row, userID, peerReadAt))
		if err != nil {
			return nil, err
		}
		items = append(items, mapped)
	}
	return &MessageListResult{Items: items, Meta: buildMeta(page, pageSize, total)}, nil
}

// SendMessage posts a message; WHATSAPP threads also send via Cloud API.
func (s *Service) SendMessage(ctx context.Context, userID, conversationID string, req SendMessageRequest) (*Message, error) {
	item, err := s.store.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeAccess(ctx, userID, item); err != nil {
		return nil, err
	}

	body := strings.TrimSpace(req.Body)
	if body == "" {
		return nil, fmt.Errorf("%w: body is required", ErrInvalidRequest)
	}
	if utf8.RuneCountInString(body) > maxMessageBodyLen {
		return nil, fmt.Errorf("%w: body must be at most %d characters", ErrInvalidRequest, maxMessageBodyLen)
	}

	senderID, err := uuidutil.FromString(userID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user id", ErrInvalidRequest)
	}

	if item.Type == db.ConversationTypeWHATSAPP {
		_ = s.store.EnsureParticipant(ctx, conversationID, userID, pgtype.Timestamptz{})
	}

	var waMessageID *string
	var waStatus *string
	if item.Type == db.ConversationTypeWHATSAPP {
		if s.messenger == nil || !s.messenger.Enabled() {
			return nil, fmt.Errorf("%w: WhatsApp is not configured", ErrInvalidRequest)
		}
		if item.WaContactPhone == nil || *item.WaContactPhone == "" {
			return nil, fmt.Errorf("%w: conversation has no WhatsApp phone", ErrInvalidRequest)
		}
		id, sendErr := s.messenger.SendText(ctx, *item.WaContactPhone, body)
		if sendErr != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidRequest, sendErr)
		}
		waMessageID = &id
		status := "accepted"
		waStatus = &status
	}

	sentAt := timestamptz(s.now())
	msg, _, err := s.store.CreateMessage(
		ctx,
		conversationID,
		senderID,
		db.MessageSenderKindUSER,
		body,
		waMessageID,
		waStatus,
		sentAt,
		previewText(body),
	)
	if err != nil {
		return nil, err
	}
	mapped, err := toAPIMessage(msg, false)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// MarkRead marks a conversation as read for the current user.
// For WHATSAPP threads it also notifies Meta so the contact sees blue ticks.
func (s *Service) MarkRead(ctx context.Context, userID, conversationID string) (*ReadResult, error) {
	item, err := s.store.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeAccess(ctx, userID, item); err != nil {
		return nil, err
	}

	readAt := timestamptz(s.now())
	if item.Type == db.ConversationTypeWHATSAPP {
		_ = s.store.EnsureParticipant(ctx, conversationID, userID, readAt)
	}

	participant, err := s.store.MarkRead(ctx, conversationID, userID, readAt)
	if err != nil {
		if errors.Is(err, ErrForbidden) && item.Type == db.ConversationTypeWHATSAPP {
			_ = s.store.EnsureParticipant(ctx, conversationID, userID, readAt)
			participant, err = s.store.MarkRead(ctx, conversationID, userID, readAt)
		}
		if err != nil {
			return nil, err
		}
	}

	if item.Type == db.ConversationTypeWHATSAPP {
		s.notifyWhatsAppRead(ctx, conversationID)
	}

	lastReadAt, err := formatTimestamptz(participant.LastReadAt)
	if err != nil {
		return nil, err
	}
	return &ReadResult{ConversationID: conversationID, LastReadAt: lastReadAt}, nil
}

// notifyWhatsAppRead marks the newest inbound contact message as read on Meta.
func (s *Service) notifyWhatsAppRead(ctx context.Context, conversationID string) {
	if s.messenger == nil || !s.messenger.Enabled() {
		return
	}
	rows, err := s.store.ListMessages(ctx, conversationID, 50, 0)
	if err != nil {
		return
	}
	for _, row := range rows {
		if row.SenderKind != db.MessageSenderKindCONTACT {
			continue
		}
		if row.WaMessageID == nil || strings.TrimSpace(*row.WaMessageID) == "" {
			continue
		}
		_ = s.messenger.MarkAsRead(ctx, strings.TrimSpace(*row.WaMessageID))
		return
	}
}

// ListContacts returns other admin users that can be chatted with.
func (s *Service) ListContacts(ctx context.Context, userID string, page, pageSize int, search string) (*ContactListResult, error) {
	if _, err := uuidutil.FromString(userID); err != nil {
		return nil, fmt.Errorf("%w: invalid user id", ErrInvalidRequest)
	}

	page, pageSize = normalizePage(page, pageSize)
	search = strings.TrimSpace(search)
	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	rows, err := s.store.ListContacts(ctx, userID, search, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.store.CountContacts(ctx, userID, search)
	if err != nil {
		return nil, err
	}

	items := make([]Contact, 0, len(rows))
	for _, row := range rows {
		id, err := uuidutil.ToString(row.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, Contact{ID: id, Email: row.Email, Name: row.Name, Role: row.Role})
	}
	return &ContactListResult{Items: items, Meta: buildMeta(page, pageSize, total)}, nil
}

// IngestWhatsAppInbound upserts a WHATSAPP thread and stores an inbound contact message.
// The bool is true when a new message row was written (false on empty/duplicate).
func (s *Service) IngestWhatsAppInbound(ctx context.Context, phone, contactName, waMessageID, body string) (bool, error) {
	if s.phones == nil {
		return false, fmt.Errorf("%w: phone normalizer not configured", ErrInvalidRequest)
	}
	digits, err := s.phones.Normalize(phone)
	if err != nil {
		return false, fmt.Errorf("%w: invalid phone", ErrInvalidRequest)
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return false, nil
	}
	if waMessageID != "" {
		if _, err := s.store.GetMessageByWaMessageID(ctx, waMessageID); err == nil {
			return false, nil
		} else if !errors.Is(err, ErrNotFound) {
			return false, err
		}
	}

	var namePtr *string
	if strings.TrimSpace(contactName) != "" {
		n := strings.TrimSpace(contactName)
		namePtr = &n
	}

	var residentUUID pgtype.UUID
	if row, findErr := s.store.FindResidentByPhones(ctx, s.phones.MatchCandidates(digits)); findErr == nil {
		residentUUID = row.ID
	}

	conv, err := s.store.GetWhatsAppConversationByPhone(ctx, digits)
	if errors.Is(err, ErrNotFound) {
		conv, err = s.store.CreateWhatsAppConversation(ctx, "", digits, namePtr, residentUUID)
		if err != nil {
			return false, err
		}
	} else if err != nil {
		return false, err
	} else if namePtr != nil || residentUUID.Valid {
		convID, _ := uuidutil.ToString(conv.ID)
		_, _ = s.store.UpdateWhatsAppProfile(ctx, convID, namePtr, residentUUID)
	}

	convID, err := uuidutil.ToString(conv.ID)
	if err != nil {
		return false, err
	}

	var waIDPtr *string
	if waMessageID != "" {
		waIDPtr = &waMessageID
	}
	status := "received"
	sentAt := timestamptz(s.now())
	_, _, err = s.store.CreateMessage(
		ctx,
		convID,
		pgtype.UUID{},
		db.MessageSenderKindCONTACT,
		body,
		waIDPtr,
		&status,
		sentAt,
		previewText(body),
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

// WhatsAppMessageCount returns how many messages exist in the WA thread for phone.
func (s *Service) WhatsAppMessageCount(ctx context.Context, phone string) (int64, error) {
	if s.phones == nil {
		return 0, fmt.Errorf("%w: phone normalizer not configured", ErrInvalidRequest)
	}
	digits, err := s.phones.Normalize(phone)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid phone", ErrInvalidRequest)
	}
	conv, err := s.store.GetWhatsAppConversationByPhone(ctx, digits)
	if errors.Is(err, ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	convID, err := uuidutil.ToString(conv.ID)
	if err != nil {
		return 0, err
	}
	return s.store.CountMessages(ctx, convID)
}

// WhatsAppDisplayName resolves a friendly name for bot replies.
// Registered resident data takes precedence over the WhatsApp profile name.
func (s *Service) WhatsAppDisplayName(ctx context.Context, phone, fallback string) (string, error) {
	if s.phones == nil {
		return strings.TrimSpace(fallback), fmt.Errorf("%w: phone normalizer not configured", ErrInvalidRequest)
	}
	digits, err := s.phones.Normalize(phone)
	if err != nil {
		return strings.TrimSpace(fallback), fmt.Errorf("%w: invalid phone", ErrInvalidRequest)
	}

	if row, findErr := s.store.FindResidentByPhones(ctx, s.phones.MatchCandidates(digits)); findErr == nil && row.Name != "" {
		return strings.TrimSpace(row.Name), nil
	}

	if conv, err := s.store.GetWhatsAppConversationByPhone(ctx, digits); err == nil {
		if conv.WaContactName != nil && strings.TrimSpace(*conv.WaContactName) != "" {
			return strings.TrimSpace(*conv.WaContactName), nil
		}
	}

	if name := strings.TrimSpace(fallback); name != "" {
		return name, nil
	}
	return "Warga", nil
}

// WhatsAppResidentIdentity resolves the registered resident behind a WhatsApp number.
func (s *Service) WhatsAppResidentIdentity(ctx context.Context, phone string) (residentID, name, normalizedPhone string, found bool, err error) {
	if s.phones == nil {
		return "", "", "", false, fmt.Errorf("%w: phone normalizer not configured", ErrInvalidRequest)
	}
	digits, err := s.phones.Normalize(phone)
	if err != nil {
		return "", "", "", false, fmt.Errorf("%w: invalid phone", ErrInvalidRequest)
	}
	row, err := s.store.FindResidentByPhones(ctx, s.phones.MatchCandidates(digits))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", "", digits, false, nil
		}
		return "", "", digits, false, err
	}
	id, err := uuidutil.ToString(row.ID)
	if err != nil {
		return "", "", digits, false, err
	}
	return id, strings.TrimSpace(row.Name), strings.TrimSpace(row.Phone), true, nil
}

// Enabled reports whether system-originated WhatsApp sending is available.
func (s *Service) Enabled() bool {
	return s != nil && s.messenger != nil && s.messenger.Enabled()
}

// SendText sends a SYSTEM WhatsApp message, creating the inbox thread when needed.
// It implements the announcement.OutboundMessenger shape without coupling packages.
func (s *Service) SendText(ctx context.Context, phone, body string) (string, error) {
	if s.phones == nil {
		return "", fmt.Errorf("%w: phone normalizer not configured", ErrInvalidRequest)
	}
	if !s.Enabled() {
		return "", fmt.Errorf("%w: WhatsApp is not configured", ErrInvalidRequest)
	}

	digits, err := s.phones.Normalize(phone)
	if err != nil {
		return "", fmt.Errorf("%w: invalid phone", ErrInvalidRequest)
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return "", fmt.Errorf("%w: body is required", ErrInvalidRequest)
	}
	if utf8.RuneCountInString(body) > maxMessageBodyLen {
		return "", fmt.Errorf("%w: body must be at most %d characters", ErrInvalidRequest, maxMessageBodyLen)
	}

	var namePtr *string
	var residentUUID pgtype.UUID
	if resident, findErr := s.store.FindResidentByPhones(ctx, s.phones.MatchCandidates(digits)); findErr == nil {
		residentUUID = resident.ID
		if name := strings.TrimSpace(resident.Name); name != "" {
			namePtr = &name
		}
	}

	conv, err := s.store.GetWhatsAppConversationByPhone(ctx, digits)
	if errors.Is(err, ErrNotFound) {
		conv, err = s.store.CreateWhatsAppConversation(ctx, "", digits, namePtr, residentUUID)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	} else if namePtr != nil || residentUUID.Valid {
		convID, _ := uuidutil.ToString(conv.ID)
		_, _ = s.store.UpdateWhatsAppProfile(ctx, convID, namePtr, residentUUID)
	}

	convID, err := uuidutil.ToString(conv.ID)
	if err != nil {
		return "", err
	}

	waID, err := s.messenger.SendText(ctx, digits, body)
	if err != nil {
		return "", err
	}
	status := "accepted"
	sentAt := timestamptz(s.now())
	_, _, err = s.store.CreateMessage(
		ctx,
		convID,
		pgtype.UUID{},
		db.MessageSenderKindSYSTEM,
		body,
		&waID,
		&status,
		sentAt,
		previewText(body),
	)
	if err != nil {
		return "", err
	}
	return waID, nil
}

// SendBotReply sends an automated SYSTEM reply over WhatsApp and stores it in the thread.
func (s *Service) SendBotReply(ctx context.Context, phone, body string) error {
	_, err := s.SendText(ctx, phone, body)
	return err
}

// ApplyWhatsAppStatus updates outbound message delivery status.
func (s *Service) ApplyWhatsAppStatus(ctx context.Context, waMessageID, status string) error {
	if strings.TrimSpace(waMessageID) == "" || strings.TrimSpace(status) == "" {
		return nil
	}
	_, err := s.store.UpdateMessageWaStatus(ctx, waMessageID, status)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}

func (s *Service) authorizeAccess(ctx context.Context, userID string, item db.Conversation) error {
	if item.Type == db.ConversationTypeWHATSAPP {
		return nil
	}
	id, err := uuidutil.ToString(item.ID)
	if err != nil {
		return err
	}
	if _, err := s.store.GetParticipant(ctx, id, userID); err != nil {
		return err
	}
	return nil
}

func (s *Service) loadParticipantsMap(ctx context.Context, conversationIDs []string) (map[string][]Participant, error) {
	rows, err := s.store.ListParticipantsByIDs(ctx, conversationIDs)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]Participant, len(conversationIDs))
	for _, row := range rows {
		conversationID, err := uuidutil.ToString(row.ConversationID)
		if err != nil {
			return nil, err
		}
		userID, err := uuidutil.ToString(row.UserID)
		if err != nil {
			return nil, err
		}
		out[conversationID] = append(out[conversationID], Participant{
			ID: userID, Name: row.Name, Email: row.Email,
		})
	}
	return out, nil
}

func mapParticipantRows(rows []db.ListParticipantsByConversationIDRow) ([]Participant, error) {
	out := make([]Participant, 0, len(rows))
	for _, row := range rows {
		userID, err := uuidutil.ToString(row.UserID)
		if err != nil {
			return nil, err
		}
		out = append(out, Participant{ID: userID, Name: row.Name, Email: row.Email})
	}
	return out, nil
}

func toAPIConversationFromListRow(row db.ListMyConversationsRow, currentUserID string, participants []Participant) (Conversation, error) {
	id, err := uuidutil.ToString(row.ID)
	if err != nil {
		return Conversation{}, err
	}
	createdAt, err := formatTimestamptz(row.CreatedAt)
	if err != nil {
		return Conversation{}, err
	}
	updatedAt, err := formatTimestamptz(row.UpdatedAt)
	if err != nil {
		return Conversation{}, err
	}
	var lastMessageAt *string
	if row.LastMessageAt.Valid {
		value, err := formatTimestamptz(row.LastMessageAt)
		if err != nil {
			return Conversation{}, err
		}
		lastMessageAt = &value
	}
	if participants == nil {
		participants = []Participant{}
	}

	var createdBy *string
	if row.CreatedBy.Valid {
		value, err := uuidutil.ToString(row.CreatedBy)
		if err != nil {
			return Conversation{}, err
		}
		createdBy = &value
	}
	var residentID *string
	if row.ResidentID.Valid {
		value, err := uuidutil.ToString(row.ResidentID)
		if err != nil {
			return Conversation{}, err
		}
		residentID = &value
	}
	var residentName *string
	if row.ResidentName != nil {
		trimmed := strings.TrimSpace(*row.ResidentName)
		if trimmed != "" {
			residentName = &trimmed
		}
	}

	return Conversation{
		ID:                 id,
		Type:               row.Type,
		Title:              row.Title,
		DisplayName:        displayName(row.Type, row.Title, row.WaContactName, row.WaContactPhone, residentName, currentUserID, participants),
		CreatedBy:          createdBy,
		WaContactPhone:     row.WaContactPhone,
		WaContactName:      row.WaContactName,
		ResidentID:         residentID,
		ResidentName:       residentName,
		LastMessageAt:      lastMessageAt,
		LastMessagePreview: row.LastMessagePreview,
		UnreadCount:        row.UnreadCount,
		Participants:       participants,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}, nil
}

func toAPIConversation(item db.Conversation, currentUserID string, participants []Participant, unread int64, residentName *string) (Conversation, error) {
	id, err := uuidutil.ToString(item.ID)
	if err != nil {
		return Conversation{}, err
	}
	createdAt, err := formatTimestamptz(item.CreatedAt)
	if err != nil {
		return Conversation{}, err
	}
	updatedAt, err := formatTimestamptz(item.UpdatedAt)
	if err != nil {
		return Conversation{}, err
	}
	var lastMessageAt *string
	if item.LastMessageAt.Valid {
		value, err := formatTimestamptz(item.LastMessageAt)
		if err != nil {
			return Conversation{}, err
		}
		lastMessageAt = &value
	}
	if participants == nil {
		participants = []Participant{}
	}
	var createdBy *string
	if item.CreatedBy.Valid {
		value, err := uuidutil.ToString(item.CreatedBy)
		if err != nil {
			return Conversation{}, err
		}
		createdBy = &value
	}
	var residentID *string
	if item.ResidentID.Valid {
		value, err := uuidutil.ToString(item.ResidentID)
		if err != nil {
			return Conversation{}, err
		}
		residentID = &value
	}
	if residentName != nil {
		trimmed := strings.TrimSpace(*residentName)
		if trimmed == "" {
			residentName = nil
		} else {
			residentName = &trimmed
		}
	}

	return Conversation{
		ID:                 id,
		Type:               item.Type,
		Title:              item.Title,
		DisplayName:        displayName(item.Type, item.Title, item.WaContactName, item.WaContactPhone, residentName, currentUserID, participants),
		CreatedBy:          createdBy,
		WaContactPhone:     item.WaContactPhone,
		WaContactName:      item.WaContactName,
		ResidentID:         residentID,
		ResidentName:       residentName,
		LastMessageAt:      lastMessageAt,
		LastMessagePreview: item.LastMessagePreview,
		UnreadCount:        unread,
		Participants:       participants,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}, nil
}

func toAPIMessage(item db.Message, isRead bool) (Message, error) {
	id, err := uuidutil.ToString(item.ID)
	if err != nil {
		return Message{}, err
	}
	conversationID, err := uuidutil.ToString(item.ConversationID)
	if err != nil {
		return Message{}, err
	}
	createdAt, err := formatTimestamptz(item.CreatedAt)
	if err != nil {
		return Message{}, err
	}
	var senderID *string
	if item.SenderID.Valid {
		value, err := uuidutil.ToString(item.SenderID)
		if err != nil {
			return Message{}, err
		}
		senderID = &value
	}
	return Message{
		ID:             id,
		ConversationID: conversationID,
		SenderID:       senderID,
		SenderKind:     item.SenderKind,
		Body:           item.Body,
		WaMessageID:    item.WaMessageID,
		WaStatus:       item.WaStatus,
		IsRead:         isRead,
		CreatedAt:      createdAt,
	}, nil
}

func messageIsRead(
	convType db.ConversationType,
	msg db.Message,
	viewerID string,
	peerReadAt pgtype.Timestamptz,
) bool {
	if msg.SenderKind != db.MessageSenderKindUSER || !msg.SenderID.Valid {
		return false
	}
	senderID, err := uuidutil.ToString(msg.SenderID)
	if err != nil || senderID != viewerID {
		return false
	}

	// WhatsApp delivery/read comes from Meta status webhooks.
	if convType == db.ConversationTypeWHATSAPP {
		if msg.WaStatus == nil {
			return false
		}
		return strings.EqualFold(*msg.WaStatus, "read")
	}

	if !peerReadAt.Valid || !msg.CreatedAt.Valid {
		return false
	}
	return !msg.CreatedAt.Time.After(peerReadAt.Time)
}

func displayName(
	convType db.ConversationType,
	title *string,
	waName *string,
	waPhone *string,
	residentName *string,
	currentUserID string,
	participants []Participant,
) string {
	switch convType {
	case db.ConversationTypeGROUP:
		if title != nil {
			return *title
		}
		return "Grup"
	case db.ConversationTypeWHATSAPP:
		resident := ""
		if residentName != nil {
			resident = strings.TrimSpace(*residentName)
		}
		wa := ""
		if waName != nil {
			wa = strings.TrimSpace(*waName)
		}
		phone := ""
		if waPhone != nil {
			phone = strings.TrimSpace(*waPhone)
		}

		// Keep the WhatsApp profile name and the residents-table name distinct.
		// If they were previously copied into each other, fall back to the phone.
		primary := wa
		if primary == "" || (resident != "" && strings.EqualFold(primary, resident)) {
			if phone != "" {
				primary = phone
			} else {
				primary = "WhatsApp"
			}
		}
		if resident != "" {
			return primary + " (" + resident + ")"
		}
		return primary + " (belum dikenali)"
	default:
		for _, p := range participants {
			if p.ID != currentUserID {
				return p.Name
			}
		}
		return "Chat"
	}
}

func previewText(body string) string {
	runes := []rune(body)
	if len(runes) <= maxPreviewLen {
		return body
	}
	return string(runes[:maxPreviewLen]) + "…"
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = defaultPage
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

func buildMeta(page, pageSize int, total int64) ListMeta {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return ListMeta{Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages}
}

func timestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}

func formatTimestamptz(ts pgtype.Timestamptz) (string, error) {
	if !ts.Valid {
		return "", fmt.Errorf("invalid timestamp")
	}
	return ts.Time.UTC().Format(time.RFC3339), nil
}
