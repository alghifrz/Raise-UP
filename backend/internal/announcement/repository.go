package announcement

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

// Repository provides announcement persistence via sqlc, including transactions.
type Repository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

// NewRepository creates an announcement repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, q: db.New(pool)}
}

// GetByID returns an announcement by UUID string.
func (r *Repository) GetByID(ctx context.Context, id string) (db.Announcement, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.Announcement{}, ErrInvalidRequest
	}
	item, err := r.q.GetAnnouncementByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Announcement{}, ErrNotFound
		}
		return db.Announcement{}, fmt.Errorf("get announcement by id: %w", err)
	}
	return item, nil
}

// ListRecipientIDs returns recipient resident UUIDs for an announcement.
func (r *Repository) ListRecipientIDs(ctx context.Context, announcementID string) ([]string, error) {
	pgID, err := uuidutil.FromString(announcementID)
	if err != nil {
		return nil, ErrInvalidRequest
	}
	ids, err := r.q.ListAnnouncementRecipientIDs(ctx, pgID)
	if err != nil {
		return nil, fmt.Errorf("list announcement recipients: %w", err)
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		value, err := uuidutil.ToString(id)
		if err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, nil
}

// List returns a filtered page of announcements.
func (r *Repository) List(
	ctx context.Context,
	search string,
	status *db.AnnouncementStatus,
	visibility *db.AnnouncementVisibility,
	category string,
	limit, offset int32,
) ([]db.Announcement, error) {
	items, err := r.q.ListAnnouncementsFiltered(ctx, db.ListAnnouncementsFilteredParams{
		Search:      search,
		Status:      toNullStatus(status),
		Visibility:  toNullVisibility(visibility),
		Category:    category,
		LimitCount:  limit,
		OffsetCount: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list announcements: %w", err)
	}
	return items, nil
}

// Count returns the total matching announcements.
func (r *Repository) Count(
	ctx context.Context,
	search string,
	status *db.AnnouncementStatus,
	visibility *db.AnnouncementVisibility,
	category string,
) (int64, error) {
	total, err := r.q.CountAnnouncementsFiltered(ctx, db.CountAnnouncementsFilteredParams{
		Search:     search,
		Status:     toNullStatus(status),
		Visibility: toNullVisibility(visibility),
		Category:   category,
	})
	if err != nil {
		return 0, fmt.Errorf("count announcements: %w", err)
	}
	return total, nil
}

// CreateWithRecipients inserts an announcement and optional recipients in one transaction.
func (r *Repository) CreateWithRecipients(ctx context.Context, arg db.CreateAnnouncementParams, recipientIDs []string) (db.Announcement, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)
	item, err := q.CreateAnnouncement(ctx, arg)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("create announcement: %w", err)
	}

	if err := replaceRecipients(ctx, q, item.ID, recipientIDs); err != nil {
		return db.Announcement{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Announcement{}, fmt.Errorf("commit tx: %w", err)
	}
	return item, nil
}

// UpdateWithRecipients updates announcement fields and replaces recipients in one transaction.
// When replaceRecipients is false, existing recipients are left untouched.
func (r *Repository) UpdateWithRecipients(
	ctx context.Context,
	arg db.UpdateAnnouncementParams,
	recipientIDs []string,
	replace bool,
) (db.Announcement, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return db.Announcement{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)
	item, err := q.UpdateAnnouncement(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Announcement{}, ErrNotFound
		}
		return db.Announcement{}, fmt.Errorf("update announcement: %w", err)
	}

	if replace {
		if err := replaceRecipients(ctx, q, item.ID, recipientIDs); err != nil {
			return db.Announcement{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Announcement{}, fmt.Errorf("commit tx: %w", err)
	}
	return item, nil
}

// Publish marks an announcement as published.
func (r *Repository) Publish(ctx context.Context, id string, publishedAt pgtype.Timestamptz) (db.Announcement, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.Announcement{}, ErrInvalidRequest
	}
	item, err := r.q.PublishAnnouncement(ctx, db.PublishAnnouncementParams{
		ID:          pgID,
		PublishedAt: publishedAt,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Announcement{}, ErrNotFound
		}
		return db.Announcement{}, fmt.Errorf("publish announcement: %w", err)
	}
	return item, nil
}

// Delete removes an announcement (recipients cascade).
func (r *Repository) Delete(ctx context.Context, id string) error {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return ErrInvalidRequest
	}
	if err := r.q.DeleteAnnouncement(ctx, pgID); err != nil {
		return fmt.Errorf("delete announcement: %w", err)
	}
	return nil
}

func replaceRecipients(ctx context.Context, q *db.Queries, announcementID pgtype.UUID, recipientIDs []string) error {
	if err := q.DeleteAnnouncementRecipients(ctx, announcementID); err != nil {
		return fmt.Errorf("delete announcement recipients: %w", err)
	}
	for _, id := range recipientIDs {
		residentID, err := uuidutil.FromString(id)
		if err != nil {
			return ErrInvalidRequest
		}
		if err := q.AddAnnouncementRecipient(ctx, db.AddAnnouncementRecipientParams{
			AnnouncementID: announcementID,
			ResidentID:     residentID,
		}); err != nil {
			return fmt.Errorf("add announcement recipient: %w", err)
		}
	}
	return nil
}

func toNullStatus(status *db.AnnouncementStatus) db.NullAnnouncementStatus {
	if status == nil {
		return db.NullAnnouncementStatus{}
	}
	return db.NullAnnouncementStatus{AnnouncementStatus: *status, Valid: true}
}

func toNullVisibility(visibility *db.AnnouncementVisibility) db.NullAnnouncementVisibility {
	if visibility == nil {
		return db.NullAnnouncementVisibility{}
	}
	return db.NullAnnouncementVisibility{AnnouncementVisibility: *visibility, Valid: true}
}
