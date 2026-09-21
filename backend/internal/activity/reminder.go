package activity

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultReminderInterval = time.Minute
	defaultReminderBatch    = 20
	residentBatchSize       = int32(100)
)

// ReminderJob is a due activity reminder ready for delivery.
type ReminderJob struct {
	ID          pgtype.UUID
	Name        string
	Description string
	Date        time.Time
	Message     string
	DueAt       time.Time
}

// ReminderStore persists scheduler state.
type ReminderStore interface {
	ListDueReminders(ctx context.Context, now time.Time, limit int) ([]ReminderJob, error)
	MarkReminderSent(ctx context.Context, id pgtype.UUID, dueAt, sentAt time.Time) error
}

// ResidentLister supplies reminder recipients.
type ResidentLister interface {
	List(ctx context.Context, search string, gender *db.Gender, limit, offset int32) ([]db.Resident, error)
}

// ReminderMessenger sends and records a SYSTEM WhatsApp message.
type ReminderMessenger interface {
	Enabled() bool
	SendText(ctx context.Context, phone, body string) (messageID string, err error)
}

// ReminderScheduler polls due reminders and broadcasts them to all residents.
type ReminderScheduler struct {
	store     ReminderStore
	residents ResidentLister
	messenger ReminderMessenger
	log       *slog.Logger
	interval  time.Duration
	now       func() time.Time
}

// NewReminderScheduler creates the native activity reminder scheduler.
func NewReminderScheduler(
	store ReminderStore,
	residents ResidentLister,
	messenger ReminderMessenger,
	log *slog.Logger,
) *ReminderScheduler {
	if log == nil {
		log = slog.Default()
	}
	return &ReminderScheduler{
		store:     store,
		residents: residents,
		messenger: messenger,
		log:       log,
		interval:  defaultReminderInterval,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

// Run processes due reminders immediately and then once per interval until canceled.
func (s *ReminderScheduler) Run(ctx context.Context) {
	if s == nil {
		return
	}
	s.runAndLog(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runAndLog(ctx)
		}
	}
}

func (s *ReminderScheduler) runAndLog(ctx context.Context) {
	if err := s.RunOnce(ctx); err != nil && ctx.Err() == nil {
		s.log.Error("activity reminder scheduler failed", "error", err)
	}
}

// RunOnce processes one batch of due reminders.
func (s *ReminderScheduler) RunOnce(ctx context.Context) error {
	if s.store == nil || s.residents == nil || s.messenger == nil || !s.messenger.Enabled() {
		return nil
	}

	now := s.now()
	jobs, err := s.store.ListDueReminders(ctx, now, defaultReminderBatch)
	if err != nil {
		return err
	}
	if len(jobs) == 0 {
		return nil
	}

	recipients, err := s.listAllResidents(ctx)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		sent, failed := 0, 0
		body := formatReminder(job)
		for _, resident := range recipients {
			if strings.TrimSpace(resident.Phone) == "" {
				failed++
				continue
			}
			if _, err := s.messenger.SendText(ctx, resident.Phone, body); err != nil {
				failed++
				s.log.Warn(
					"activity reminder WhatsApp send failed",
					"activity", job.Name,
					"resident_id", resident.ID,
					"error", err,
				)
				continue
			}
			sent++
		}

		if err := s.store.MarkReminderSent(ctx, job.ID, job.DueAt, s.now()); err != nil {
			return err
		}
		s.log.Info(
			"activity reminder sent",
			"activity", job.Name,
			"recipients", len(recipients),
			"sent", sent,
			"failed", failed,
		)
	}
	return nil
}

func (s *ReminderScheduler) listAllResidents(ctx context.Context) ([]db.Resident, error) {
	out := make([]db.Resident, 0)
	for offset := int32(0); ; offset += residentBatchSize {
		items, err := s.residents.List(ctx, "", nil, residentBatchSize, offset)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
		if len(items) < int(residentBatchSize) {
			return out, nil
		}
	}
}

func formatReminder(job ReminderJob) string {
	date := job.Date.In(Jakarta).Format("02-01-2006")
	message := strings.TrimSpace(job.Message)
	if message == "" {
		message = strings.TrimSpace(job.Description)
	}
	if message == "" {
		message = "Jangan lupa untuk mengikuti kegiatan ini."
	}
	return fmt.Sprintf(
		"⏰ *Pengingat Kegiatan RT*\n\n*%s*\n🗓️ %s\n\n%s",
		strings.TrimSpace(job.Name),
		date,
		message,
	)
}
