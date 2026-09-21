package activity_test

import (
	"context"
	"strings"
	"testing"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/internal/activity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type reminderStore struct {
	jobs   []activity.ReminderJob
	marked []pgtype.UUID
}

func (s *reminderStore) ListDueReminders(context.Context, time.Time, int) ([]activity.ReminderJob, error) {
	return s.jobs, nil
}

func (s *reminderStore) MarkReminderSent(_ context.Context, id pgtype.UUID, _, _ time.Time) error {
	s.marked = append(s.marked, id)
	return nil
}

type reminderResidents struct {
	items []db.Resident
}

func (r *reminderResidents) List(_ context.Context, _ string, _ *db.Gender, limit, offset int32) ([]db.Resident, error) {
	if int(offset) >= len(r.items) {
		return []db.Resident{}, nil
	}
	end := int(offset + limit)
	if end > len(r.items) {
		end = len(r.items)
	}
	return r.items[offset:end], nil
}

type reminderMessenger struct {
	sent []string
}

func (m *reminderMessenger) Enabled() bool { return true }

func (m *reminderMessenger) SendText(_ context.Context, phone, body string) (string, error) {
	m.sent = append(m.sent, phone+"|"+body)
	return uuid.NewString(), nil
}

func TestReminderSchedulerSendsAndMarks(t *testing.T) {
	id := uuid.New()
	store := &reminderStore{jobs: []activity.ReminderJob{
		{
			ID:      pgtype.UUID{Bytes: id, Valid: true},
			Name:    "Kerja Bakti",
			Date:    time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC),
			Message: "Mohon membawa alat kebersihan.",
			DueAt:   time.Date(2026, 9, 26, 2, 0, 0, 0, time.UTC),
		},
	}}
	residents := &reminderResidents{items: []db.Resident{
		{Name: "A", Phone: "081111111111"},
		{Name: "B", Phone: "082222222222"},
	}}
	messenger := &reminderMessenger{}
	scheduler := activity.NewReminderScheduler(store, residents, messenger, nil)

	if err := scheduler.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	if len(messenger.sent) != 2 {
		t.Fatalf("sent = %d, want 2", len(messenger.sent))
	}
	if len(store.marked) != 1 {
		t.Fatalf("marked = %d, want 1", len(store.marked))
	}
	if !strings.Contains(messenger.sent[0], "Kerja Bakti") ||
		!strings.Contains(messenger.sent[0], "Mohon membawa alat") {
		t.Fatalf("unexpected reminder: %q", messenger.sent[0])
	}
}
