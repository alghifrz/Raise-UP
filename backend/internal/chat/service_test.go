package chat_test

import (
	"context"
	"testing"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/internal/chat"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type memoryStore struct {
	users         map[string]db.GetChatUserByIDRow
	conversations map[string]db.Conversation
	participants  map[string]map[string]db.ConversationParticipant // convID -> userID
	messages      map[string][]db.Message                          // convID -> msgs
	directPairs   map[string]string                                // sortedPair -> convID
	residents     map[string]db.Resident
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		users:         make(map[string]db.GetChatUserByIDRow),
		conversations: make(map[string]db.Conversation),
		participants:  make(map[string]map[string]db.ConversationParticipant),
		messages:      make(map[string][]db.Message),
		directPairs:   make(map[string]string),
		residents:     make(map[string]db.Resident),
	}
}

func (m *memoryStore) residentNameFor(residentID pgtype.UUID) *string {
	if !residentID.Valid {
		return nil
	}
	id, err := uuidutil.ToString(residentID)
	if err != nil {
		return nil
	}
	resident, ok := m.residents[id]
	if !ok {
		return nil
	}
	name := resident.Name
	return &name
}

func (m *memoryStore) addUser(id, name, email string) {
	pgID, _ := uuidutil.FromString(id)
	m.users[id] = db.GetChatUserByIDRow{
		ID:    pgID,
		Name:  name,
		Email: email,
		Role:  db.UserRoleADMINRW,
	}
}

func pairKey(a, b string) string {
	if a < b {
		return a + ":" + b
	}
	return b + ":" + a
}

func (m *memoryStore) GetConversationByID(_ context.Context, id string) (db.Conversation, error) {
	item, ok := m.conversations[id]
	if !ok {
		return db.Conversation{}, chat.ErrNotFound
	}
	return item, nil
}

func (m *memoryStore) FindDirectConversationBetween(_ context.Context, userID, otherUserID string) (db.Conversation, error) {
	id, ok := m.directPairs[pairKey(userID, otherUserID)]
	if !ok {
		return db.Conversation{}, chat.ErrNotFound
	}
	return m.conversations[id], nil
}

func (m *memoryStore) GetParticipant(_ context.Context, conversationID, userID string) (db.ConversationParticipant, error) {
	byUser, ok := m.participants[conversationID]
	if !ok {
		return db.ConversationParticipant{}, chat.ErrForbidden
	}
	item, ok := byUser[userID]
	if !ok {
		return db.ConversationParticipant{}, chat.ErrForbidden
	}
	return item, nil
}

func (m *memoryStore) ListParticipants(_ context.Context, conversationID string) ([]db.ListParticipantsByConversationIDRow, error) {
	byUser := m.participants[conversationID]
	out := make([]db.ListParticipantsByConversationIDRow, 0, len(byUser))
	pgConv, _ := uuidutil.FromString(conversationID)
	for userID := range byUser {
		u := m.users[userID]
		out = append(out, db.ListParticipantsByConversationIDRow{
			ConversationID: pgConv,
			UserID:         u.ID,
			Name:           u.Name,
			Email:          u.Email,
		})
	}
	return out, nil
}

func (m *memoryStore) ListParticipantsByIDs(ctx context.Context, conversationIDs []string) ([]db.ListParticipantsByConversationIDsRow, error) {
	out := make([]db.ListParticipantsByConversationIDsRow, 0)
	for _, id := range conversationIDs {
		rows, err := m.ListParticipants(ctx, id)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			out = append(out, db.ListParticipantsByConversationIDsRow(row))
		}
	}
	return out, nil
}

func (m *memoryStore) ListMyConversations(_ context.Context, userID, _ string, limit, offset int32) ([]db.ListMyConversationsRow, error) {
	out := make([]db.ListMyConversationsRow, 0)
	for convID, byUser := range m.participants {
		if _, ok := byUser[userID]; !ok {
			continue
		}
		c := m.conversations[convID]
		out = append(out, db.ListMyConversationsRow{
			ID:                 c.ID,
			Type:               c.Type,
			Title:              c.Title,
			CreatedBy:          c.CreatedBy,
			WaContactPhone:     c.WaContactPhone,
			WaContactName:      c.WaContactName,
			ResidentID:         c.ResidentID,
			ResidentName:       m.residentNameFor(c.ResidentID),
			LastMessageAt:      c.LastMessageAt,
			LastMessagePreview: c.LastMessagePreview,
			CreatedAt:          c.CreatedAt,
			UpdatedAt:          c.UpdatedAt,
			MyLastReadAt:       byUser[userID].LastReadAt,
			UnreadCount:        0,
		})
	}
	if offset >= int32(len(out)) {
		return []db.ListMyConversationsRow{}, nil
	}
	end := offset + limit
	if end > int32(len(out)) {
		end = int32(len(out))
	}
	return out[offset:end], nil
}

func (m *memoryStore) CountMyConversations(_ context.Context, userID, _ string) (int64, error) {
	var total int64
	for _, byUser := range m.participants {
		if _, ok := byUser[userID]; ok {
			total++
		}
	}
	return total, nil
}

func (m *memoryStore) CreateDirectConversation(_ context.Context, createdBy, otherUserID string, joinedAt pgtype.Timestamptz) (db.Conversation, error) {
	id := uuid.New().String()
	pgID, _ := uuidutil.FromString(id)
	createdByID, _ := uuidutil.FromString(createdBy)
	otherID, _ := uuidutil.FromString(otherUserID)
	now := joinedAt
	item := db.Conversation{
		ID:        pgID,
		Type:      db.ConversationTypeDIRECT,
		CreatedBy: createdByID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.conversations[id] = item
	m.participants[id] = map[string]db.ConversationParticipant{
		createdBy:    {ConversationID: pgID, UserID: createdByID, JoinedAt: now, LastReadAt: joinedAt},
		otherUserID:  {ConversationID: pgID, UserID: otherID, JoinedAt: now, LastReadAt: joinedAt},
	}
	m.directPairs[pairKey(createdBy, otherUserID)] = id
	return item, nil
}

func (m *memoryStore) CreateGroupConversation(_ context.Context, createdBy, title string, participantIDs []string, joinedAt pgtype.Timestamptz) (db.Conversation, error) {
	id := uuid.New().String()
	pgID, _ := uuidutil.FromString(id)
	createdByID, _ := uuidutil.FromString(createdBy)
	titleCopy := title
	now := joinedAt
	item := db.Conversation{
		ID:        pgID,
		Type:      db.ConversationTypeGROUP,
		Title:     &titleCopy,
		CreatedBy: createdByID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.conversations[id] = item
	m.participants[id] = map[string]db.ConversationParticipant{
		createdBy: {ConversationID: pgID, UserID: createdByID, JoinedAt: now, LastReadAt: joinedAt},
	}
	for _, pid := range participantIDs {
		pgUser, _ := uuidutil.FromString(pid)
		m.participants[id][pid] = db.ConversationParticipant{
			ConversationID: pgID,
			UserID:         pgUser,
			JoinedAt:       now,
			LastReadAt:     joinedAt,
		}
	}
	return item, nil
}

func (m *memoryStore) ListMessages(_ context.Context, conversationID string, limit, offset int32) ([]db.Message, error) {
	items := m.messages[conversationID]
	// newest first
	out := make([]db.Message, len(items))
	copy(out, items)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	if offset >= int32(len(out)) {
		return []db.Message{}, nil
	}
	end := offset + limit
	if end > int32(len(out)) {
		end = int32(len(out))
	}
	return out[offset:end], nil
}

func (m *memoryStore) CountMessages(_ context.Context, conversationID string) (int64, error) {
	return int64(len(m.messages[conversationID])), nil
}

func (m *memoryStore) CountUnread(_ context.Context, conversationID, userID string) (int64, error) {
	p, ok := m.participants[conversationID][userID]
	if !ok {
		return 0, chat.ErrForbidden
	}
	var total int64
	for _, msg := range m.messages[conversationID] {
		sender, _ := uuidutil.ToString(msg.SenderID)
		if sender == userID {
			continue
		}
		if !p.LastReadAt.Valid || msg.CreatedAt.Time.After(p.LastReadAt.Time) {
			total++
		}
	}
	return total, nil
}

func (m *memoryStore) MarkRead(_ context.Context, conversationID, userID string, readAt pgtype.Timestamptz) (db.ConversationParticipant, error) {
	p, ok := m.participants[conversationID][userID]
	if !ok {
		return db.ConversationParticipant{}, chat.ErrForbidden
	}
	p.LastReadAt = readAt
	m.participants[conversationID][userID] = p
	return p, nil
}

func (m *memoryStore) ListContacts(_ context.Context, userID, _ string, limit, offset int32) ([]db.ListChatContactsRow, error) {
	out := make([]db.ListChatContactsRow, 0)
	for id, u := range m.users {
		if id == userID {
			continue
		}
		out = append(out, db.ListChatContactsRow(u))
	}
	if offset >= int32(len(out)) {
		return []db.ListChatContactsRow{}, nil
	}
	end := offset + limit
	if end > int32(len(out)) {
		end = int32(len(out))
	}
	return out[offset:end], nil
}

func (m *memoryStore) CountContacts(_ context.Context, userID, _ string) (int64, error) {
	var total int64
	for id := range m.users {
		if id != userID {
			total++
		}
	}
	return total, nil
}

func (m *memoryStore) GetChatUserByID(_ context.Context, id string) (db.GetChatUserByIDRow, error) {
	u, ok := m.users[id]
	if !ok {
		return db.GetChatUserByIDRow{}, chat.ErrUserNotFound
	}
	return u, nil
}

func (m *memoryStore) GetWhatsAppConversationByPhone(context.Context, string) (db.Conversation, error) {
	return db.Conversation{}, chat.ErrNotFound
}

func (m *memoryStore) EnsureParticipant(context.Context, string, string, pgtype.Timestamptz) error {
	return nil
}

func (m *memoryStore) CreateWhatsAppConversation(context.Context, string, string, *string, pgtype.UUID) (db.Conversation, error) {
	return db.Conversation{}, chat.ErrInvalidRequest
}

func (m *memoryStore) UpdateWhatsAppProfile(context.Context, string, *string, pgtype.UUID) (db.Conversation, error) {
	return db.Conversation{}, chat.ErrNotFound
}

func (m *memoryStore) GetMessageByWaMessageID(context.Context, string) (db.Message, error) {
	return db.Message{}, chat.ErrNotFound
}

func (m *memoryStore) UpdateMessageWaStatus(context.Context, string, string) (db.Message, error) {
	return db.Message{}, chat.ErrNotFound
}

func (m *memoryStore) FindResidentByPhones(context.Context, []string) (db.FindResidentByPhoneCandidatesRow, error) {
	return db.FindResidentByPhoneCandidatesRow{}, chat.ErrNotFound
}

func (m *memoryStore) GetResidentByID(_ context.Context, id string) (db.Resident, error) {
	item, ok := m.residents[id]
	if !ok {
		return db.Resident{}, chat.ErrNotFound
	}
	return item, nil
}

func (m *memoryStore) CreateMessage(
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
	_ = senderKind
	_ = waMessageID
	_ = waStatus
	sender, _ := uuidutil.ToString(senderID)
	return m.createMessageLegacy(ctx, conversationID, sender, body, sentAt, preview)
}

func (m *memoryStore) createMessageLegacy(_ context.Context, conversationID, senderID, body string, sentAt pgtype.Timestamptz, preview string) (db.Message, db.Conversation, error) {
	msgID := uuid.New().String()
	pgMsgID, _ := uuidutil.FromString(msgID)
	pgConvID, _ := uuidutil.FromString(conversationID)
	pgSenderID, _ := uuidutil.FromString(senderID)
	msg := db.Message{
		ID:             pgMsgID,
		ConversationID: pgConvID,
		SenderID:       pgSenderID,
		Body:           body,
		CreatedAt:      sentAt,
		SenderKind:     db.MessageSenderKindUSER,
	}
	m.messages[conversationID] = append(m.messages[conversationID], msg)
	conv := m.conversations[conversationID]
	conv.LastMessageAt = sentAt
	conv.LastMessagePreview = preview
	m.conversations[conversationID] = conv
	if p, ok := m.participants[conversationID][senderID]; ok {
		p.LastReadAt = sentAt
		m.participants[conversationID][senderID] = p
	}
	return msg, conv, nil
}

func newTestService(store *memoryStore) *chat.Service {
	return chat.NewService(store, nil, nil)
}

func TestCreateDirectAndSendMessage(t *testing.T) {
	store := newMemoryStore()
	me := uuid.New().String()
	other := uuid.New().String()
	store.addUser(me, "Admin A", "a@example.com")
	store.addUser(other, "Admin B", "b@example.com")

	svc := newTestService(store)
	ctx := context.Background()

	conv, err := svc.CreateConversation(ctx, me, chat.CreateRequest{
		Type:          "DIRECT",
		ParticipantID: other,
	})
	if err != nil {
		t.Fatalf("create direct: %v", err)
	}
	if conv.Type != db.ConversationTypeDIRECT {
		t.Fatalf("expected DIRECT, got %s", conv.Type)
	}
	if conv.DisplayName != "Admin B" {
		t.Fatalf("expected display name Admin B, got %q", conv.DisplayName)
	}

	// Reuse existing direct chat
	again, err := svc.CreateConversation(ctx, me, chat.CreateRequest{
		Type:          "DIRECT",
		ParticipantID: other,
	})
	if err != nil {
		t.Fatalf("reuse direct: %v", err)
	}
	if again.ID != conv.ID {
		t.Fatalf("expected same conversation id")
	}

	time.Sleep(2 * time.Millisecond)

	msg, err := svc.SendMessage(ctx, me, conv.ID, chat.SendMessageRequest{Body: "Halo"})
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	if msg.Body != "Halo" {
		t.Fatalf("unexpected body %q", msg.Body)
	}

	list, err := svc.ListMessages(ctx, other, conv.ID, 1, 20)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected 1 message, got %d", len(list.Items))
	}

	unread, err := svc.GetConversation(ctx, other, conv.ID)
	if err != nil {
		t.Fatalf("get conversation: %v", err)
	}
	if unread.UnreadCount != 1 {
		t.Fatalf("expected unread 1, got %d", unread.UnreadCount)
	}

	if _, err := svc.MarkRead(ctx, other, conv.ID); err != nil {
		t.Fatalf("mark read: %v", err)
	}
	after, err := svc.GetConversation(ctx, other, conv.ID)
	if err != nil {
		t.Fatalf("get after read: %v", err)
	}
	if after.UnreadCount != 0 {
		t.Fatalf("expected unread 0, got %d", after.UnreadCount)
	}
}

func TestCreateGroupValidation(t *testing.T) {
	store := newMemoryStore()
	me := uuid.New().String()
	other := uuid.New().String()
	store.addUser(me, "Admin A", "a@example.com")
	store.addUser(other, "Admin B", "b@example.com")
	svc := newTestService(store)

	_, err := svc.CreateConversation(context.Background(), me, chat.CreateRequest{
		Type:           "GROUP",
		Title:          "Pengurus",
		ParticipantIDs: []string{other},
	})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	_, err = svc.CreateConversation(context.Background(), me, chat.CreateRequest{
		Type:  "GROUP",
		Title: "Kosong",
	})
	if err == nil {
		t.Fatal("expected error for empty participants")
	}
}

func TestForbiddenNonParticipant(t *testing.T) {
	store := newMemoryStore()
	a := uuid.New().String()
	b := uuid.New().String()
	c := uuid.New().String()
	store.addUser(a, "A", "a@example.com")
	store.addUser(b, "B", "b@example.com")
	store.addUser(c, "C", "c@example.com")
	svc := newTestService(store)

	conv, err := svc.CreateConversation(context.Background(), a, chat.CreateRequest{
		Type:          "DIRECT",
		ParticipantID: b,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err = svc.SendMessage(context.Background(), c, conv.ID, chat.SendMessageRequest{Body: "hack"})
	if err == nil {
		t.Fatal("expected forbidden")
	}
}

func TestSendMessageRequiresBody(t *testing.T) {
	store := newMemoryStore()
	a := uuid.New().String()
	b := uuid.New().String()
	store.addUser(a, "A", "a@example.com")
	store.addUser(b, "B", "b@example.com")
	svc := newTestService(store)

	conv, err := svc.CreateConversation(context.Background(), a, chat.CreateRequest{
		Type:          "DIRECT",
		ParticipantID: b,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err = svc.SendMessage(context.Background(), a, conv.ID, chat.SendMessageRequest{Body: "   "})
	if err == nil {
		t.Fatal("expected invalid body")
	}
}

func TestWhatsAppDisplayNameUsesResidentOrUnknown(t *testing.T) {
	store := newMemoryStore()
	adminID := uuid.New().String()
	store.addUser(adminID, "Admin", "admin@example.com")

	residentID := uuid.New()
	residentPg, _ := uuidutil.FromString(residentID.String())
	store.residents[residentID.String()] = db.Resident{
		ID:   residentPg,
		Name: "alghif",
	}

	knownConvID := uuid.New().String()
	knownPg, _ := uuidutil.FromString(knownConvID)
	adminPg, _ := uuidutil.FromString(adminID)
	waName := "User"
	phone := "6281234567890"
	now := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	store.conversations[knownConvID] = db.Conversation{
		ID:             knownPg,
		Type:           db.ConversationTypeWHATSAPP,
		CreatedBy:      adminPg,
		WaContactName:  &waName,
		WaContactPhone: &phone,
		ResidentID:     residentPg,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	store.participants[knownConvID] = map[string]db.ConversationParticipant{
		adminID: {ConversationID: knownPg, UserID: adminPg, LastReadAt: now},
	}

	unknownConvID := uuid.New().String()
	unknownPg, _ := uuidutil.FromString(unknownConvID)
	store.conversations[unknownConvID] = db.Conversation{
		ID:             unknownPg,
		Type:           db.ConversationTypeWHATSAPP,
		CreatedBy:      adminPg,
		WaContactName:  &waName,
		WaContactPhone: &phone,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	store.participants[unknownConvID] = map[string]db.ConversationParticipant{
		adminID: {ConversationID: unknownPg, UserID: adminPg, LastReadAt: now},
	}

	svc := newTestService(store)
	list, err := svc.ListConversations(context.Background(), adminID, 1, 20, "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	byID := map[string]string{}
	for _, item := range list.Items {
		byID[item.ID] = item.DisplayName
	}
	if got := byID[knownConvID]; got != "User (alghif)" {
		t.Fatalf("known resident display name: got %q", got)
	}
	if got := byID[unknownConvID]; got != "User (belum dikenal)" {
		t.Fatalf("unknown resident display name: got %q", got)
	}

	detail, err := svc.GetConversation(context.Background(), adminID, knownConvID)
	if err != nil {
		t.Fatalf("get known: %v", err)
	}
	if detail.DisplayName != "User (alghif)" {
		t.Fatalf("get known display name: got %q", detail.DisplayName)
	}
}
