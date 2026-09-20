package activity_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/internal/activity"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type memoryStore struct {
	byID map[string]db.Activity
}

func newMemoryStore() *memoryStore {
	return &memoryStore{byID: make(map[string]db.Activity)}
}

func (m *memoryStore) GetByID(_ context.Context, id string) (db.Activity, error) {
	item, ok := m.byID[id]
	if !ok {
		return db.Activity{}, activity.ErrNotFound
	}
	return item, nil
}

func (m *memoryStore) List(_ context.Context, search string, fromAt, toExclusive *time.Time, limit, offset int32) ([]db.Activity, error) {
	items := make([]db.Activity, 0)
	for _, item := range m.byID {
		if search != "" {
			needle := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(item.Name), needle) &&
				!strings.Contains(strings.ToLower(item.Description), needle) {
				continue
			}
		}
		if fromAt != nil && item.Date.Time.Before(*fromAt) {
			continue
		}
		if toExclusive != nil && !item.Date.Time.Before(*toExclusive) {
			continue
		}
		items = append(items, item)
	}
	// crude sort by date ASC then created_at DESC
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Date.Time.Before(items[i].Date.Time) ||
				(items[j].Date.Time.Equal(items[i].Date.Time) && items[j].CreatedAt.Time.After(items[i].CreatedAt.Time)) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	if offset >= int32(len(items)) {
		return []db.Activity{}, nil
	}
	items = items[offset:]
	if int32(len(items)) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (m *memoryStore) Count(ctx context.Context, search string, fromAt, toExclusive *time.Time) (int64, error) {
	items, _ := m.List(ctx, search, fromAt, toExclusive, 100000, 0)
	return int64(len(items)), nil
}

func (m *memoryStore) Create(_ context.Context, arg db.CreateActivityParams) (db.Activity, error) {
	id := uuid.New()
	now := time.Now().UTC()
	item := db.Activity{
		ID:                      pgUUID(id),
		Name:                    arg.Name,
		Description:             arg.Description,
		Date:                    arg.Date,
		ReminderDaysBefore:      arg.ReminderDaysBefore,
		ReminderTime:            arg.ReminderTime,
		ReminderMessage:         arg.ReminderMessage,
		ReminderScheduleVersion: 1,
		ReminderScheduledAt:     pgtype.Timestamptz{},
		ReminderN8nExecutionID:  nil,
		ReminderSentAt:          pgtype.Timestamptz{},
		CreatedAt:               pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:               pgtype.Timestamptz{Time: now, Valid: true},
	}
	m.byID[id.String()] = item
	return item, nil
}

func (m *memoryStore) Update(_ context.Context, arg db.UpdateActivityParams) (db.Activity, error) {
	id, err := uuidutil.ToString(arg.ID)
	if err != nil {
		return db.Activity{}, err
	}
	item, ok := m.byID[id]
	if !ok {
		return db.Activity{}, activity.ErrNotFound
	}
	if arg.Name != nil {
		item.Name = *arg.Name
	}
	if arg.Description != nil {
		item.Description = *arg.Description
	}
	if arg.Date.Valid {
		item.Date = arg.Date
	}
	if arg.ReminderDaysBefore != nil {
		item.ReminderDaysBefore = *arg.ReminderDaysBefore
	}
	if arg.ReminderTime.Valid {
		item.ReminderTime = arg.ReminderTime
	}
	if arg.ReminderMessage != nil {
		item.ReminderMessage = *arg.ReminderMessage
	}
	if arg.ReminderConfigChanged {
		item.ReminderScheduleVersion++
		item.ReminderScheduledAt = pgtype.Timestamptz{}
		item.ReminderN8nExecutionID = nil
		item.ReminderSentAt = pgtype.Timestamptz{}
	}
	item.UpdatedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	m.byID[id] = item
	return item, nil
}

func (m *memoryStore) Delete(_ context.Context, id string) error {
	if _, ok := m.byID[id]; !ok {
		return activity.ErrNotFound
	}
	delete(m.byID, id)
	return nil
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func seedOperationalState(store *memoryStore, id string) {
	item := store.byID[id]
	execID := "exec-123"
	item.ReminderScheduledAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	item.ReminderN8nExecutionID = &execID
	item.ReminderSentAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	store.byID[id] = item
}

func TestCreateValidActivity(t *testing.T) {
	svc := activity.NewService(newMemoryStore())
	days := int32(2)
	item, err := svc.Create(context.Background(), activity.CreateRequest{
		Name:               "Kerja Bakti RT",
		Description:        "Kerja bakti membersihkan lingkungan RT",
		Date:               "2026-10-10T07:00:00+07:00",
		ReminderDaysBefore: &days,
		ReminderTime:       "19:00",
		ReminderMessage:    "Pengingat: kerja bakti RT akan dilaksanakan 2 hari lagi.",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if item.Name != "Kerja Bakti RT" || item.ReminderDaysBefore != 2 || item.ReminderScheduleVersion != 1 {
		t.Fatalf("unexpected activity: %+v", item)
	}
	if item.ReminderTime == nil || *item.ReminderTime != "19:00" {
		t.Fatalf("unexpected reminder_time: %v", item.ReminderTime)
	}
	if item.ReminderScheduledAt != nil || item.ReminderN8nExecutionID != nil || item.ReminderSentAt != nil {
		t.Fatalf("operational state should be null: %+v", item)
	}
}

func TestCreateValidation(t *testing.T) {
	svc := activity.NewService(newMemoryStore())
	cases := []struct {
		name string
		req  activity.CreateRequest
	}{
		{"missing name", activity.CreateRequest{Date: "2026-10-10T07:00:00+07:00"}},
		{"empty trimmed name", activity.CreateRequest{Name: "   ", Date: "2026-10-10T07:00:00+07:00"}},
		{"missing date", activity.CreateRequest{Name: "Event"}},
		{"invalid date", activity.CreateRequest{Name: "Event", Date: "not-rfc3339"}},
		{"negative days", activity.CreateRequest{Name: "Event", Date: "2026-10-10T07:00:00+07:00", ReminderDaysBefore: int32Ptr(-1)}},
		{"invalid reminder time", activity.CreateRequest{Name: "Event", Date: "2026-10-10T07:00:00+07:00", ReminderTime: "7pm"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tc.req)
			if !errors.Is(err, activity.ErrInvalidRequest) {
				t.Fatalf("expected ErrInvalidRequest, got %v", err)
			}
		})
	}
}

func TestCreateOptionalReminderDefaults(t *testing.T) {
	svc := activity.NewService(newMemoryStore())
	item, err := svc.Create(context.Background(), activity.CreateRequest{
		Name: "Simple",
		Date: "2026-10-10T07:00:00+07:00",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if item.ReminderDaysBefore != 0 || item.ReminderTime != nil || item.ReminderMessage != "" || item.ReminderScheduleVersion != 1 {
		t.Fatalf("unexpected defaults: %+v", item)
	}
}

func TestUpdateNameOnlyDoesNotBumpVersion(t *testing.T) {
	store := newMemoryStore()
	svc := activity.NewService(store)
	created, err := svc.Create(context.Background(), activity.CreateRequest{
		Name: "Old", Date: "2026-10-10T07:00:00+07:00", ReminderTime: "19:00",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	seedOperationalState(store, created.ID)

	name := "New Name"
	updated, err := svc.Update(context.Background(), created.ID, activity.UpdateRequest{Name: &name})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "New Name" || updated.ReminderScheduleVersion != 1 {
		t.Fatalf("unexpected update: %+v", updated)
	}
	if updated.ReminderScheduledAt == nil || updated.ReminderN8nExecutionID == nil || updated.ReminderSentAt == nil {
		t.Fatalf("operational state should remain: %+v", updated)
	}
}

func TestUpdateDescriptionOnlyDoesNotBumpVersion(t *testing.T) {
	store := newMemoryStore()
	svc := activity.NewService(store)
	created, _ := svc.Create(context.Background(), activity.CreateRequest{
		Name: "Event", Date: "2026-10-10T07:00:00+07:00",
	})
	seedOperationalState(store, created.ID)
	desc := "Updated description"
	updated, err := svc.Update(context.Background(), created.ID, activity.UpdateRequest{Description: &desc})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.ReminderScheduleVersion != 1 || updated.ReminderScheduledAt == nil {
		t.Fatalf("version/state should be unchanged: %+v", updated)
	}
}

func TestUpdateReminderConfigIncrementsAndClears(t *testing.T) {
	store := newMemoryStore()
	svc := activity.NewService(store)
	created, _ := svc.Create(context.Background(), activity.CreateRequest{
		Name: "Event", Date: "2026-10-10T07:00:00+07:00", ReminderTime: "19:00",
	})
	seedOperationalState(store, created.ID)

	cases := []struct {
		name string
		req  activity.UpdateRequest
	}{
		{"date", activity.UpdateRequest{Date: strPtr("2026-10-11T08:00:00+07:00")}},
		{"days", activity.UpdateRequest{ReminderDaysBefore: int32Ptr(3)}},
		{"time", activity.UpdateRequest{ReminderTime: strPtr("20:30")}},
		{"message", activity.UpdateRequest{ReminderMessage: strPtr("New reminder")}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store = newMemoryStore()
			svc = activity.NewService(store)
			created, _ = svc.Create(context.Background(), activity.CreateRequest{
				Name: "Event", Date: "2026-10-10T07:00:00+07:00", ReminderTime: "19:00",
			})
			seedOperationalState(store, created.ID)

			updated, err := svc.Update(context.Background(), created.ID, tc.req)
			if err != nil {
				t.Fatalf("update: %v", err)
			}
			if updated.ReminderScheduleVersion != 2 {
				t.Fatalf("expected version 2, got %d", updated.ReminderScheduleVersion)
			}
			if updated.ReminderScheduledAt != nil || updated.ReminderN8nExecutionID != nil || updated.ReminderSentAt != nil {
				t.Fatalf("operational state should be cleared: %+v", updated)
			}
		})
	}
}

func TestUpdateInvalidAndMissing(t *testing.T) {
	svc := activity.NewService(newMemoryStore())
	name := "x"
	_, err := svc.Update(context.Background(), "not-a-uuid", activity.UpdateRequest{Name: &name})
	if !errors.Is(err, activity.ErrInvalidRequest) {
		t.Fatalf("expected invalid uuid, got %v", err)
	}
	_, err = svc.Update(context.Background(), uuid.NewString(), activity.UpdateRequest{Name: &name})
	if !errors.Is(err, activity.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestListPaginationSearchDateRange(t *testing.T) {
	store := newMemoryStore()
	svc := activity.NewService(store)
	_, _ = svc.Create(context.Background(), activity.CreateRequest{
		Name: "Kerja Bakti", Description: "bersih", Date: "2026-10-10T07:00:00+07:00",
	})
	_, _ = svc.Create(context.Background(), activity.CreateRequest{
		Name: "Rapat RW", Description: "bulanan", Date: "2026-10-20T07:00:00+07:00",
	})
	_, _ = svc.Create(context.Background(), activity.CreateRequest{
		Name: "Lain", Description: "kerja lain", Date: "2026-11-01T07:00:00+07:00",
	})

	page, err := svc.List(context.Background(), 1, 1, "", "", "")
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if page.Meta.Total != 3 || len(page.Items) != 1 || page.Meta.TotalPages != 3 {
		t.Fatalf("unexpected pagination: %+v", page.Meta)
	}

	search, err := svc.List(context.Background(), 1, 20, "kerja", "", "")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if search.Meta.Total != 2 {
		t.Fatalf("expected 2 search hits, got %d", search.Meta.Total)
	}

	ranged, err := svc.List(context.Background(), 1, 20, "", "2026-10-01", "2026-10-31")
	if err != nil {
		t.Fatalf("range: %v", err)
	}
	if ranged.Meta.Total != 2 {
		t.Fatalf("expected 2 in October, got %d", ranged.Meta.Total)
	}

	_, err = svc.List(context.Background(), 1, 20, "", "2026-10-31", "2026-10-01")
	if !errors.Is(err, activity.ErrInvalidRequest) {
		t.Fatalf("expected invalid range, got %v", err)
	}

	// ordering: earlier date first
	ordered, err := svc.List(context.Background(), 1, 20, "", "", "")
	if err != nil {
		t.Fatalf("ordered: %v", err)
	}
	if !strings.Contains(ordered.Items[0].Name, "Kerja") {
		t.Fatalf("expected earliest date first, got %+v", ordered.Items[0])
	}
}

func TestDelete(t *testing.T) {
	svc := activity.NewService(newMemoryStore())
	created, _ := svc.Create(context.Background(), activity.CreateRequest{
		Name: "Event", Date: "2026-10-10T07:00:00+07:00",
	})
	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err := svc.Get(context.Background(), created.ID)
	if !errors.Is(err, activity.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	err = svc.Delete(context.Background(), "bad-id")
	if !errors.Is(err, activity.ErrInvalidRequest) {
		t.Fatalf("expected invalid uuid, got %v", err)
	}
	err = svc.Delete(context.Background(), uuid.NewString())
	if !errors.Is(err, activity.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestRequestStructsOmitOperationalFields(t *testing.T) {
	createFields := jsonFieldNames(activity.CreateRequest{})
	updateFields := jsonFieldNames(activity.UpdateRequest{})
	forbidden := []string{
		"reminder_scheduled_at",
		"reminder_schedule_version",
		"reminder_n8n_execution_id",
		"reminder_sent_at",
	}
	for _, field := range forbidden {
		if _, ok := createFields[field]; ok {
			t.Fatalf("CreateRequest must not expose %s", field)
		}
		if _, ok := updateFields[field]; ok {
			t.Fatalf("UpdateRequest must not expose %s", field)
		}
	}
}

func jsonFieldNames(v any) map[string]struct{} {
	out := make(map[string]struct{})
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		out[name] = struct{}{}
	}
	return out
}

func int32Ptr(v int32) *int32 { return &v }
func strPtr(v string) *string { return &v }

func TestRoundTripJSONDoesNotAcceptOperationalCreateFields(t *testing.T) {
	raw := []byte(`{
		"name":"Event",
		"date":"2026-10-10T07:00:00+07:00",
		"reminder_scheduled_at":"2026-10-01T00:00:00Z",
		"reminder_schedule_version":99,
		"reminder_n8n_execution_id":"hack",
		"reminder_sent_at":"2026-10-01T00:00:00Z"
	}`)
	var req activity.CreateRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	svc := activity.NewService(newMemoryStore())
	item, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if item.ReminderScheduleVersion != 1 || item.ReminderScheduledAt != nil || item.ReminderN8nExecutionID != nil {
		t.Fatalf("client operational fields must be ignored: %+v", item)
	}
}
