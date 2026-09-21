package activity

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides activity persistence via sqlc.
type Repository struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

// NewRepository creates an activity repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool), pool: pool}
}

// GetByID returns an activity by UUID string.
func (r *Repository) GetByID(ctx context.Context, id string) (db.Activity, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.Activity{}, ErrInvalidRequest
	}
	item, err := r.q.GetActivityByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Activity{}, ErrNotFound
		}
		return db.Activity{}, fmt.Errorf("get activity: %w", err)
	}
	return item, nil
}

// List returns a filtered page of activities.
func (r *Repository) List(
	ctx context.Context,
	search string,
	fromAt, toExclusive *time.Time,
	limit, offset int32,
) ([]db.Activity, error) {
	items, err := r.q.ListActivitiesFiltered(ctx, db.ListActivitiesFilteredParams{
		Search:     search,
		FromAt:     toTimestamptz(fromAt),
		ToAt:       toTimestamptz(toExclusive),
		PageLimit:  limit,
		PageOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list activities: %w", err)
	}
	return items, nil
}

// Count returns the total matching activities.
func (r *Repository) Count(ctx context.Context, search string, fromAt, toExclusive *time.Time) (int64, error) {
	total, err := r.q.CountActivitiesFiltered(ctx, db.CountActivitiesFilteredParams{
		Search: search,
		FromAt: toTimestamptz(fromAt),
		ToAt:   toTimestamptz(toExclusive),
	})
	if err != nil {
		return 0, fmt.Errorf("count activities: %w", err)
	}
	return total, nil
}

// Create inserts an activity with initial reminder version/state.
func (r *Repository) Create(ctx context.Context, arg db.CreateActivityParams) (db.Activity, error) {
	item, err := r.q.CreateActivity(ctx, arg)
	if err != nil {
		return db.Activity{}, fmt.Errorf("create activity: %w", err)
	}
	return item, nil
}

// Update applies a patch, optionally bumping reminder version and clearing operational state.
func (r *Repository) Update(ctx context.Context, arg db.UpdateActivityParams) (db.Activity, error) {
	item, err := r.q.UpdateActivity(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Activity{}, ErrNotFound
		}
		return db.Activity{}, fmt.Errorf("update activity: %w", err)
	}
	return item, nil
}

// Delete removes an activity by id.
func (r *Repository) Delete(ctx context.Context, id string) error {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return ErrInvalidRequest
	}
	n, err := r.q.DeleteActivity(ctx, pgID)
	if err != nil {
		return fmt.Errorf("delete activity: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListDueReminders returns unsent reminders whose Jakarta wall-clock schedule is due.
func (r *Repository) ListDueReminders(ctx context.Context, now time.Time, limit int) ([]ReminderJob, error) {
	const query = `
SELECT
    a.id,
    a.name,
    a.description,
    a.date,
    a.reminder_message,
    COALESCE(
        a.reminder_scheduled_at,
        (
            (
                (a.date AT TIME ZONE 'Asia/Jakarta')::date
                - a.reminder_days_before
                + a.reminder_time
            ) AT TIME ZONE 'Asia/Jakarta'
        )
    ) AS due_at
FROM activities a
WHERE a.reminder_time IS NOT NULL
  AND a.reminder_sent_at IS NULL
  AND a.date >= $1
  AND COALESCE(
        a.reminder_scheduled_at,
        (
            (
                (a.date AT TIME ZONE 'Asia/Jakarta')::date
                - a.reminder_days_before
                + a.reminder_time
            ) AT TIME ZONE 'Asia/Jakarta'
        )
      ) <= $1
ORDER BY due_at ASC
LIMIT $2`

	rows, err := r.pool.Query(ctx, query, now.UTC(), limit)
	if err != nil {
		return nil, fmt.Errorf("list due activity reminders: %w", err)
	}
	defer rows.Close()

	jobs := make([]ReminderJob, 0)
	for rows.Next() {
		var job ReminderJob
		if err := rows.Scan(
			&job.ID,
			&job.Name,
			&job.Description,
			&job.Date,
			&job.Message,
			&job.DueAt,
		); err != nil {
			return nil, fmt.Errorf("scan due activity reminder: %w", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate due activity reminders: %w", err)
	}
	return jobs, nil
}

// MarkReminderSent records the computed schedule and completion timestamp.
func (r *Repository) MarkReminderSent(ctx context.Context, id pgtype.UUID, dueAt, sentAt time.Time) error {
	const query = `
UPDATE activities
SET reminder_scheduled_at = COALESCE(reminder_scheduled_at, $2),
    reminder_sent_at = $3
WHERE id = $1
  AND reminder_sent_at IS NULL`
	if _, err := r.pool.Exec(ctx, query, id, dueAt.UTC(), sentAt.UTC()); err != nil {
		return fmt.Errorf("mark activity reminder sent: %w", err)
	}
	return nil
}

func toTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}
