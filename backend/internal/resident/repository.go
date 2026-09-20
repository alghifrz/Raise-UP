package resident

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

// Repository provides resident persistence via sqlc.
type Repository struct {
	q *db.Queries
}

// NewRepository creates a resident repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

// GetByID returns a resident by UUID string.
func (r *Repository) GetByID(ctx context.Context, id string) (db.Resident, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.Resident{}, ErrInvalidRequest
	}

	resident, err := r.q.GetResidentByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Resident{}, ErrNotFound
		}
		return db.Resident{}, fmt.Errorf("get resident by id: %w", err)
	}
	return resident, nil
}

// List returns a filtered page of residents.
func (r *Repository) List(ctx context.Context, search string, gender *db.Gender, limit, offset int32) ([]db.Resident, error) {
	items, err := r.q.ListResidentsFiltered(ctx, db.ListResidentsFilteredParams{
		Search:      search,
		Gender:      toNullGender(gender),
		LimitCount:  limit,
		OffsetCount: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list residents: %w", err)
	}
	return items, nil
}

// Count returns the total number of residents matching filters.
func (r *Repository) Count(ctx context.Context, search string, gender *db.Gender) (int64, error) {
	total, err := r.q.CountResidentsFiltered(ctx, db.CountResidentsFilteredParams{
		Search: search,
		Gender: toNullGender(gender),
	})
	if err != nil {
		return 0, fmt.Errorf("count residents: %w", err)
	}
	return total, nil
}

// Create inserts a resident.
func (r *Repository) Create(ctx context.Context, name, phone string, gender db.Gender) (db.Resident, error) {
	resident, err := r.q.CreateResident(ctx, db.CreateResidentParams{
		Name:   name,
		Phone:  phone,
		Gender: gender,
	})
	if err != nil {
		return db.Resident{}, fmt.Errorf("create resident: %w", err)
	}
	return resident, nil
}

// Update replaces resident fields.
func (r *Repository) Update(ctx context.Context, id string, name, phone string, gender db.Gender) (db.Resident, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.Resident{}, ErrInvalidRequest
	}

	resident, err := r.q.UpdateResident(ctx, db.UpdateResidentParams{
		ID:     pgID,
		Name:   name,
		Phone:  phone,
		Gender: gender,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Resident{}, ErrNotFound
		}
		return db.Resident{}, fmt.Errorf("update resident: %w", err)
	}
	return resident, nil
}

// Delete removes a resident by id.
func (r *Repository) Delete(ctx context.Context, id string) error {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return ErrInvalidRequest
	}

	err = r.q.DeleteResident(ctx, pgID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrDeleteConflict
		}
		return fmt.Errorf("delete resident: %w", err)
	}
	return nil
}

func toNullGender(gender *db.Gender) db.NullGender {
	if gender == nil {
		return db.NullGender{}
	}
	return db.NullGender{Gender: *gender, Valid: true}
}
