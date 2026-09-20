package complaint

import (
	"context"
	"errors"
	"fmt"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides complaint persistence via sqlc.
type Repository struct {
	q *db.Queries
}

// NewRepository creates a complaint repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

// GetByID returns a complaint by UUID string.
func (r *Repository) GetByID(ctx context.Context, id string) (db.Complaint, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.Complaint{}, ErrInvalidRequest
	}

	item, err := r.q.GetComplaintByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Complaint{}, ErrNotFound
		}
		return db.Complaint{}, fmt.Errorf("get complaint by id: %w", err)
	}
	return item, nil
}

// List returns a filtered page of complaints.
func (r *Repository) List(
	ctx context.Context,
	search string,
	status *db.ComplaintStatus,
	urgency *db.ComplaintUrgency,
	category string,
	limit, offset int32,
) ([]db.Complaint, error) {
	items, err := r.q.ListComplaintsFiltered(ctx, db.ListComplaintsFilteredParams{
		Search:      search,
		Status:      toNullStatus(status),
		Urgency:     toNullUrgency(urgency),
		Category:    category,
		LimitCount:  limit,
		OffsetCount: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list complaints: %w", err)
	}
	return items, nil
}

// Count returns the total matching complaints.
func (r *Repository) Count(
	ctx context.Context,
	search string,
	status *db.ComplaintStatus,
	urgency *db.ComplaintUrgency,
	category string,
) (int64, error) {
	total, err := r.q.CountComplaintsFiltered(ctx, db.CountComplaintsFilteredParams{
		Search:   search,
		Status:   toNullStatus(status),
		Urgency:  toNullUrgency(urgency),
		Category: category,
	})
	if err != nil {
		return 0, fmt.Errorf("count complaints: %w", err)
	}
	return total, nil
}

// Create inserts a complaint.
func (r *Repository) Create(ctx context.Context, arg db.CreateComplaintParams) (db.Complaint, error) {
	item, err := r.q.CreateComplaint(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return db.Complaint{}, err
		}
		return db.Complaint{}, fmt.Errorf("create complaint: %w", err)
	}
	return item, nil
}

// Update replaces mutable complaint fields.
func (r *Repository) Update(ctx context.Context, arg db.UpdateComplaintParams) (db.Complaint, error) {
	item, err := r.q.UpdateComplaint(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Complaint{}, ErrNotFound
		}
		return db.Complaint{}, fmt.Errorf("update complaint: %w", err)
	}
	return item, nil
}

// UpdateStatus changes only the complaint status.
func (r *Repository) UpdateStatus(ctx context.Context, id string, status db.ComplaintStatus) (db.Complaint, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.Complaint{}, ErrInvalidRequest
	}

	item, err := r.q.UpdateComplaintStatus(ctx, db.UpdateComplaintStatusParams{
		ID:     pgID,
		Status: status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Complaint{}, ErrNotFound
		}
		return db.Complaint{}, fmt.Errorf("update complaint status: %w", err)
	}
	return item, nil
}

// Delete removes a complaint by id.
func (r *Repository) Delete(ctx context.Context, id string) error {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return ErrInvalidRequest
	}
	if err := r.q.DeleteComplaint(ctx, pgID); err != nil {
		return fmt.Errorf("delete complaint: %w", err)
	}
	return nil
}

func toNullStatus(status *db.ComplaintStatus) db.NullComplaintStatus {
	if status == nil {
		return db.NullComplaintStatus{}
	}
	return db.NullComplaintStatus{ComplaintStatus: *status, Valid: true}
}

func toNullUrgency(urgency *db.ComplaintUrgency) db.NullComplaintUrgency {
	if urgency == nil {
		return db.NullComplaintUrgency{}
	}
	return db.NullComplaintUrgency{ComplaintUrgency: *urgency, Valid: true}
}
