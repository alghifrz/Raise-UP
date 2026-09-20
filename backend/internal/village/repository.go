package village

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

// Repository provides village profile/officials persistence via sqlc.
type Repository struct {
	q *db.Queries
}

// NewRepository creates a village repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

// GetProfile returns the singleton village profile.
func (r *Repository) GetProfile(ctx context.Context) (db.VillageProfile, error) {
	item, err := r.q.GetVillageProfile(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.VillageProfile{}, ErrProfileNotFound
		}
		return db.VillageProfile{}, fmt.Errorf("get village profile: %w", err)
	}
	return item, nil
}

// UpdateProfile patches the singleton village profile by id.
func (r *Repository) UpdateProfile(ctx context.Context, arg db.UpdateVillageProfileParams) (db.VillageProfile, error) {
	item, err := r.q.UpdateVillageProfile(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.VillageProfile{}, ErrProfileNotFound
		}
		return db.VillageProfile{}, fmt.Errorf("update village profile: %w", err)
	}
	return item, nil
}

// GetOfficialByID returns an official by UUID string.
func (r *Repository) GetOfficialByID(ctx context.Context, id string) (db.VillageOfficial, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.VillageOfficial{}, ErrInvalidRequest
	}
	item, err := r.q.GetVillageOfficialByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.VillageOfficial{}, ErrOfficialNotFound
		}
		return db.VillageOfficial{}, fmt.Errorf("get village official: %w", err)
	}
	return item, nil
}

// ListOfficials returns officials for a profile ordered by sort_order, id.
func (r *Repository) ListOfficials(ctx context.Context, profileID pgtype.UUID) ([]db.VillageOfficial, error) {
	items, err := r.q.ListVillageOfficials(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("list village officials: %w", err)
	}
	return items, nil
}

// CreateOfficial inserts a village official.
func (r *Repository) CreateOfficial(ctx context.Context, arg db.CreateVillageOfficialParams) (db.VillageOfficial, error) {
	item, err := r.q.CreateVillageOfficial(ctx, arg)
	if err != nil {
		return db.VillageOfficial{}, fmt.Errorf("create village official: %w", err)
	}
	return item, nil
}

// UpdateOfficial patches a village official.
func (r *Repository) UpdateOfficial(ctx context.Context, arg db.UpdateVillageOfficialParams) (db.VillageOfficial, error) {
	item, err := r.q.UpdateVillageOfficial(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.VillageOfficial{}, ErrOfficialNotFound
		}
		return db.VillageOfficial{}, fmt.Errorf("update village official: %w", err)
	}
	return item, nil
}

// DeleteOfficial removes a village official by id.
func (r *Repository) DeleteOfficial(ctx context.Context, id string) error {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return ErrInvalidRequest
	}
	n, err := r.q.DeleteVillageOfficial(ctx, pgID)
	if err != nil {
		return fmt.Errorf("delete village official: %w", err)
	}
	if n == 0 {
		return ErrOfficialNotFound
	}
	return nil
}
