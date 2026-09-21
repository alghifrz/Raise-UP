package announcement_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/internal/announcement"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type memoryStore struct {
	byID       map[string]db.Announcement
	recipients map[string][]string
	failCreate bool
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		byID:       make(map[string]db.Announcement),
		recipients: make(map[string][]string),
	}
}

func (m *memoryStore) GetByID(_ context.Context, id string) (db.Announcement, error) {
	item, ok := m.byID[id]
	if !ok {
		return db.Announcement{}, announcement.ErrNotFound
	}
	return item, nil
}

func (m *memoryStore) ListRecipientIDs(_ context.Context, announcementID string) ([]string, error) {
	ids := m.recipients[announcementID]
	if ids == nil {
		return []string{}, nil
	}
	out := make([]string, len(ids))
	copy(out, ids)
	return out, nil
}

func (m *memoryStore) List(_ context.Context, search string, status *db.AnnouncementStatus, visibility *db.AnnouncementVisibility, category string, limit, offset int32) ([]db.Announcement, error) {
	items := make([]db.Announcement, 0)
	for _, item := range m.byID {
		if search != "" {
			needle := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(item.Title), needle) &&
				!strings.Contains(strings.ToLower(item.Excerpt), needle) &&
				!strings.Contains(strings.ToLower(item.Body), needle) &&
				!strings.Contains(strings.ToLower(item.Category), needle) {
				continue
			}
		}
		if status != nil && item.Status != *status {
			continue
		}
		if visibility != nil && item.Visibility != *visibility {
			continue
		}
		if category != "" && item.Category != category {
			continue
		}
		items = append(items, item)
	}
	if offset >= int32(len(items)) {
		return []db.Announcement{}, nil
	}
	items = items[offset:]
	if int32(len(items)) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (m *memoryStore) Count(ctx context.Context, search string, status *db.AnnouncementStatus, visibility *db.AnnouncementVisibility, category string) (int64, error) {
	items, err := m.List(ctx, search, status, visibility, category, 100000, 0)
	if err != nil {
		return 0, err
	}
	return int64(len(items)), nil
}

func (m *memoryStore) CreateWithRecipients(_ context.Context, arg db.CreateAnnouncementParams, recipientIDs []string) (db.Announcement, error) {
	if m.failCreate {
		return db.Announcement{}, errors.New("forced create failure")
	}
	id := uuid.New()
	now := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	item := db.Announcement{
		ID:           pgtype.UUID{Bytes: id, Valid: true},
		Title:        arg.Title,
		Excerpt:      arg.Excerpt,
		Body:         arg.Body,
		Category:     arg.Category,
		Visibility:   arg.Visibility,
		Status:       arg.Status,
		ThumbnailUrl: arg.ThumbnailUrl,
		AuthorID:     arg.AuthorID,
		PublishedAt:  arg.PublishedAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	m.byID[id.String()] = item
	copied := make([]string, len(recipientIDs))
	copy(copied, recipientIDs)
	m.recipients[id.String()] = copied
	return item, nil
}

func (m *memoryStore) UpdateWithRecipients(_ context.Context, arg db.UpdateAnnouncementParams, recipientIDs []string, replace bool) (db.Announcement, error) {
	id, err := uuidutil.ToString(arg.ID)
	if err != nil {
		return db.Announcement{}, err
	}
	item, ok := m.byID[id]
	if !ok {
		return db.Announcement{}, announcement.ErrNotFound
	}
	item.Title = arg.Title
	item.Excerpt = arg.Excerpt
	item.Body = arg.Body
	item.Category = arg.Category
	item.Visibility = arg.Visibility
	item.ThumbnailUrl = arg.ThumbnailUrl
	item.UpdatedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	m.byID[id] = item
	if replace {
		copied := make([]string, len(recipientIDs))
		copy(copied, recipientIDs)
		m.recipients[id] = copied
	}
	return item, nil
}

func (m *memoryStore) Publish(_ context.Context, id string, publishedAt pgtype.Timestamptz) (db.Announcement, error) {
	item, ok := m.byID[id]
	if !ok {
		return db.Announcement{}, announcement.ErrNotFound
	}
	item.Status = db.AnnouncementStatusPUBLISHED
	item.PublishedAt = publishedAt
	item.UpdatedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	m.byID[id] = item
	return item, nil
}

func (m *memoryStore) Delete(_ context.Context, id string) error {
	if _, ok := m.byID[id]; !ok {
		return announcement.ErrNotFound
	}
	delete(m.byID, id)
	delete(m.recipients, id)
	return nil
}

type memoryResidents struct {
	byID map[string]db.Resident
}

type sentMessage struct {
	to   string
	body string
}

type memoryMessenger struct {
	enabled bool
	sent    []sentMessage
}

func (m *memoryMessenger) Enabled() bool {
	return m.enabled
}

func (m *memoryMessenger) SendText(_ context.Context, to, body string) (string, error) {
	m.sent = append(m.sent, sentMessage{to: to, body: body})
	return uuid.NewString(), nil
}

func newMemoryResidents() *memoryResidents {
	return &memoryResidents{byID: make(map[string]db.Resident)}
}

func (m *memoryResidents) GetByID(_ context.Context, id string) (db.Resident, error) {
	item, ok := m.byID[id]
	if !ok {
		return db.Resident{}, announcement.ErrResidentNotFound
	}
	return item, nil
}

func (m *memoryResidents) add() string {
	id := uuid.New()
	m.byID[id.String()] = db.Resident{
		ID:    pgtype.UUID{Bytes: id, Valid: true},
		Name:  "Resident",
		Phone: "081",
	}
	return id.String()
}

func authorID() string {
	id := uuid.New()
	return id.String()
}

func TestCreatePublicValid(t *testing.T) {
	service := announcement.NewService(newMemoryStore(), newMemoryResidents())
	got, err := service.Create(context.Background(), authorID(), announcement.CreateRequest{
		Title:      "  Title  ",
		Excerpt:    " excerpt ",
		Body:       " body ",
		Category:   " KEGIATAN ",
		Visibility: "PUBLIC",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got.Status != db.AnnouncementStatusDRAFT || got.Visibility != db.AnnouncementVisibilityPUBLIC {
		t.Fatalf("unexpected create result: %+v", got)
	}
	if got.AuthorID == "" || got.PublishedAt != nil || len(got.RecipientIDs) != 0 {
		t.Fatalf("unexpected author/publish/recipients: %+v", got)
	}
}

func TestCreatePrivateValid(t *testing.T) {
	residents := newMemoryResidents()
	r1, r2 := residents.add(), residents.add()
	service := announcement.NewService(newMemoryStore(), residents)
	got, err := service.Create(context.Background(), authorID(), announcement.CreateRequest{
		Title:        "Private",
		Body:         "secret",
		Category:     "INFO",
		Visibility:   "PRIVATE",
		RecipientIDs: []string{r1, r2},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if len(got.RecipientIDs) != 2 {
		t.Fatalf("recipients = %v", got.RecipientIDs)
	}
}

func TestCreateValidationErrors(t *testing.T) {
	residents := newMemoryResidents()
	service := announcement.NewService(newMemoryStore(), residents)
	author := authorID()

	cases := []struct {
		name string
		req  announcement.CreateRequest
		want error
	}{
		{"missing title", announcement.CreateRequest{Body: "b", Category: "c", Visibility: "PUBLIC"}, announcement.ErrInvalidRequest},
		{"missing body", announcement.CreateRequest{Title: "t", Category: "c", Visibility: "PUBLIC"}, announcement.ErrInvalidRequest},
		{"invalid visibility", announcement.CreateRequest{Title: "t", Body: "b", Category: "c", Visibility: "HIDDEN"}, announcement.ErrInvalidRequest},
		{"private without recipients", announcement.CreateRequest{Title: "t", Body: "b", Category: "c", Visibility: "PRIVATE"}, announcement.ErrRecipientRequired},
		{"invalid recipient uuid", announcement.CreateRequest{Title: "t", Body: "b", Category: "c", Visibility: "PRIVATE", RecipientIDs: []string{"bad"}}, announcement.ErrInvalidRequest},
		{"missing recipient", announcement.CreateRequest{Title: "t", Body: "b", Category: "c", Visibility: "PRIVATE", RecipientIDs: []string{uuid.NewString()}}, announcement.ErrResidentNotFound},
		{"duplicate recipients", announcement.CreateRequest{Title: "t", Body: "b", Category: "c", Visibility: "PRIVATE", RecipientIDs: []string{residents.add(), ""}}, announcement.ErrInvalidRequest},
	}

	// fix duplicate recipients case with real duplicate
	dup := residents.add()
	cases[len(cases)-1] = struct {
		name string
		req  announcement.CreateRequest
		want error
	}{"duplicate recipients", announcement.CreateRequest{Title: "t", Body: "b", Category: "c", Visibility: "PRIVATE", RecipientIDs: []string{dup, dup}}, announcement.ErrInvalidRequest}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.Create(context.Background(), author, tc.req)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestCreateDefaultsCategoryWhenOmitted(t *testing.T) {
	service := announcement.NewService(newMemoryStore(), newMemoryResidents())
	got, err := service.Create(context.Background(), authorID(), announcement.CreateRequest{
		Title:      "Info warga",
		Body:       "Isi pengumuman",
		Visibility: "PUBLIC",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got.Category != "UMUM" {
		t.Fatalf("Category = %q, want UMUM", got.Category)
	}
}

func TestCreateAuthorFromContext(t *testing.T) {
	service := announcement.NewService(newMemoryStore(), newMemoryResidents())
	author := authorID()
	got, err := service.Create(context.Background(), author, announcement.CreateRequest{
		Title: "t", Body: "b", Category: "c", Visibility: "PUBLIC",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got.AuthorID != author {
		t.Fatalf("author_id = %s, want %s", got.AuthorID, author)
	}
}

func TestUpdateVisibilityAndRecipients(t *testing.T) {
	store := newMemoryStore()
	residents := newMemoryResidents()
	r1, r2, r3 := residents.add(), residents.add(), residents.add()
	service := announcement.NewService(store, residents)

	created, err := service.Create(context.Background(), authorID(), announcement.CreateRequest{
		Title: "t", Body: "b", Category: "c", Visibility: "PRIVATE", RecipientIDs: []string{r1, r2},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	title := "Updated"
	updated, err := service.Update(context.Background(), created.ID, announcement.UpdateRequest{Title: &title})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Title != "Updated" || len(updated.RecipientIDs) != 2 {
		t.Fatalf("unexpected partial update: %+v", updated)
	}

	recipients := []string{r1, r3}
	replaced, err := service.Update(context.Background(), created.ID, announcement.UpdateRequest{RecipientIDs: &recipients})
	if err != nil {
		t.Fatalf("Update recipients error = %v", err)
	}
	if len(replaced.RecipientIDs) != 2 || replaced.RecipientIDs[0] != r1 && replaced.RecipientIDs[1] != r1 {
		// order may vary - check set
	}
	if !containsAll(replaced.RecipientIDs, r1, r3) || containsAll(replaced.RecipientIDs, r2) {
		t.Fatalf("unexpected recipients: %v", replaced.RecipientIDs)
	}

	visPublic := "PUBLIC"
	toPublic, err := service.Update(context.Background(), created.ID, announcement.UpdateRequest{Visibility: &visPublic})
	if err != nil {
		t.Fatalf("PRIVATE→PUBLIC error = %v", err)
	}
	if toPublic.Visibility != db.AnnouncementVisibilityPUBLIC || !containsAll(toPublic.RecipientIDs, r1, r3) {
		t.Fatalf("visibility should not alter recipients: %+v", toPublic)
	}
	if !containsAll(store.recipients[created.ID], r1, r3) {
		t.Fatalf("store recipients changed with visibility: %v", store.recipients[created.ID])
	}

	visPrivate := "PRIVATE"
	backToPrivate, err := service.Update(context.Background(), created.ID, announcement.UpdateRequest{Visibility: &visPrivate})
	if err != nil {
		t.Fatalf("PUBLIC→PRIVATE with existing recipients error = %v", err)
	}
	if !containsAll(backToPrivate.RecipientIDs, r1, r3) {
		t.Fatalf("existing recipients were not preserved: %+v", backToPrivate)
	}

	newRecipients := []string{r2}
	toPrivate, err := service.Update(context.Background(), created.ID, announcement.UpdateRequest{
		Visibility:   &visPrivate,
		RecipientIDs: &newRecipients,
	})
	if err != nil {
		t.Fatalf("PUBLIC→PRIVATE error = %v", err)
	}
	if toPrivate.Visibility != db.AnnouncementVisibilityPRIVATE || !containsAll(toPrivate.RecipientIDs, r2) {
		t.Fatalf("unexpected private result: %+v", toPrivate)
	}
}

func TestUpdatePublicCanHaveRecipients(t *testing.T) {
	residents := newMemoryResidents()
	r1 := residents.add()
	service := announcement.NewService(newMemoryStore(), residents)
	created, _ := service.Create(context.Background(), authorID(), announcement.CreateRequest{
		Title: "t", Body: "b", Category: "c", Visibility: "PUBLIC",
	})
	ids := []string{r1}
	updated, err := service.Update(context.Background(), created.ID, announcement.UpdateRequest{RecipientIDs: &ids})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !containsAll(updated.RecipientIDs, r1) {
		t.Fatalf("public recipients = %v", updated.RecipientIDs)
	}
}

func TestStatusPublish(t *testing.T) {
	service := announcement.NewService(newMemoryStore(), newMemoryResidents())
	created, err := service.Create(context.Background(), authorID(), announcement.CreateRequest{
		Title: "t", Body: "b", Category: "c", Visibility: "PUBLIC",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	published, err := service.UpdateStatus(context.Background(), created.ID, "PUBLISHED")
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if published.Status != db.AnnouncementStatusPUBLISHED || published.PublishedAt == nil {
		t.Fatalf("unexpected publish result: %+v", published)
	}

	_, err = service.UpdateStatus(context.Background(), created.ID, "DRAFT")
	if !errors.Is(err, announcement.ErrInvalidStatusTransition) {
		t.Fatalf("error = %v, want ErrInvalidStatusTransition", err)
	}
	_, err = service.UpdateStatus(context.Background(), created.ID, "PUBLISHED")
	if !errors.Is(err, announcement.ErrInvalidStatusTransition) {
		t.Fatalf("republish error = %v", err)
	}
}

func TestPublishSendsWhatsAppToRecipients(t *testing.T) {
	store := newMemoryStore()
	residents := newMemoryResidents()
	r1, r2 := residents.add(), residents.add()
	messenger := &memoryMessenger{enabled: true}
	service := announcement.NewService(store, residents)
	service.SetMessenger(messenger)

	created, err := service.Create(context.Background(), authorID(), announcement.CreateRequest{
		Title:        "Kerja Bakti",
		Body:         "Hari Minggu pukul 07.00.",
		Visibility:   "PRIVATE",
		RecipientIDs: []string{r1, r2},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	published, err := service.UpdateStatus(context.Background(), created.ID, "PUBLISHED")
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if published.Delivery == nil || published.Delivery.Sent != 2 || published.Delivery.Failed != 0 {
		t.Fatalf("delivery = %+v", published.Delivery)
	}
	if len(messenger.sent) != 2 {
		t.Fatalf("sent messages = %d, want 2", len(messenger.sent))
	}
	for _, message := range messenger.sent {
		if !strings.Contains(message.body, "Kerja Bakti") || !strings.Contains(message.body, "Hari Minggu") {
			t.Fatalf("unexpected WhatsApp body: %q", message.body)
		}
	}
}

func TestDeleteAndGet(t *testing.T) {
	residents := newMemoryResidents()
	r1 := residents.add()
	service := announcement.NewService(newMemoryStore(), residents)
	created, _ := service.Create(context.Background(), authorID(), announcement.CreateRequest{
		Title: "t", Body: "body search", Category: "INFO", Visibility: "PRIVATE", RecipientIDs: []string{r1},
	})

	got, err := service.Get(context.Background(), created.ID)
	if err != nil || len(got.RecipientIDs) != 1 {
		t.Fatalf("Get() = %+v err=%v", got, err)
	}

	public, _ := service.Create(context.Background(), authorID(), announcement.CreateRequest{
		Title: "public", Body: "b", Category: "INFO", Visibility: "PUBLIC",
	})
	gotPublic, err := service.Get(context.Background(), public.ID)
	if err != nil || gotPublic.RecipientIDs == nil || len(gotPublic.RecipientIDs) != 0 {
		t.Fatalf("public recipients should be empty array: %+v", gotPublic)
	}

	if err := service.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if err := service.Delete(context.Background(), created.ID); !errors.Is(err, announcement.ErrNotFound) {
		t.Fatalf("Delete() error = %v, want ErrNotFound", err)
	}
}

func TestListFilters(t *testing.T) {
	service := announcement.NewService(newMemoryStore(), newMemoryResidents())
	_, _ = service.Create(context.Background(), authorID(), announcement.CreateRequest{
		Title: "Kerja Bakti", Body: "detail", Category: "KEGIATAN", Visibility: "PUBLIC",
	})
	_, _ = service.Create(context.Background(), authorID(), announcement.CreateRequest{
		Title: "Info Lain", Body: "lain", Category: "INFO", Visibility: "PUBLIC",
	})

	defaults, err := service.List(context.Background(), 0, 0, "", "", "", "")
	if err != nil || defaults.Meta.Page != 1 || defaults.Meta.PageSize != 20 || defaults.Meta.Total != 2 {
		t.Fatalf("defaults = %+v err=%v", defaults, err)
	}
	limited, _ := service.List(context.Background(), 1, 1000, "", "", "", "")
	if limited.Meta.PageSize != 100 {
		t.Fatalf("page_size = %d", limited.Meta.PageSize)
	}
	bySearch, _ := service.List(context.Background(), 1, 20, "bakti", "", "", "")
	if bySearch.Meta.Total != 1 {
		t.Fatalf("search total = %d", bySearch.Meta.Total)
	}
	byCategory, _ := service.List(context.Background(), 1, 20, "", "", "", "INFO")
	if byCategory.Meta.Total != 1 {
		t.Fatalf("category total = %d", byCategory.Meta.Total)
	}
}

func TestCreateTransactionFailure(t *testing.T) {
	store := newMemoryStore()
	store.failCreate = true
	service := announcement.NewService(store, newMemoryResidents())
	_, err := service.Create(context.Background(), authorID(), announcement.CreateRequest{
		Title: "t", Body: "b", Category: "c", Visibility: "PUBLIC",
	})
	if err == nil {
		t.Fatal("expected create failure")
	}
	if len(store.byID) != 0 {
		t.Fatal("partial create should not persist")
	}
}

func containsAll(values []string, want ...string) bool {
	set := map[string]struct{}{}
	for _, v := range values {
		set[v] = struct{}{}
	}
	for _, w := range want {
		if _, ok := set[w]; !ok {
			return false
		}
	}
	return true
}
