package gallery

import (
	"context"
	"errors"
	"fmt"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides gallery persistence via sqlc.
type Repository struct {
	q *db.Queries
}

// NewRepository creates a gallery repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

// GetByID returns a gallery item by UUID string.
func (r *Repository) GetByID(ctx context.Context, id string) (db.GalleryItem, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.GalleryItem{}, ErrInvalidRequest
	}
	item, err := r.q.GetGalleryItemByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GalleryItem{}, ErrNotFound
		}
		return db.GalleryItem{}, fmt.Errorf("get gallery item: %w", err)
	}
	return item, nil
}

// List returns a filtered page of gallery items.
func (r *Repository) List(ctx context.Context, search string, limit, offset int32) ([]db.GalleryItem, error) {
	items, err := r.q.ListGalleryItemsFiltered(ctx, db.ListGalleryItemsFilteredParams{
		Search:     search,
		PageLimit:  limit,
		PageOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list gallery items: %w", err)
	}
	return items, nil
}

// Count returns the total matching gallery items.
func (r *Repository) Count(ctx context.Context, search string) (int64, error) {
	total, err := r.q.CountGalleryItemsFiltered(ctx, search)
	if err != nil {
		return 0, fmt.Errorf("count gallery items: %w", err)
	}
	return total, nil
}

// Create inserts a gallery item.
func (r *Repository) Create(ctx context.Context, arg db.CreateGalleryItemParams) (db.GalleryItem, error) {
	item, err := r.q.CreateGalleryItem(ctx, arg)
	if err != nil {
		return db.GalleryItem{}, fmt.Errorf("create gallery item: %w", err)
	}
	return item, nil
}

// Update patches a gallery item.
func (r *Repository) Update(ctx context.Context, arg db.UpdateGalleryItemParams) (db.GalleryItem, error) {
	item, err := r.q.UpdateGalleryItem(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GalleryItem{}, ErrNotFound
		}
		return db.GalleryItem{}, fmt.Errorf("update gallery item: %w", err)
	}
	return item, nil
}

// Delete removes a gallery item by id.
func (r *Repository) Delete(ctx context.Context, id string) error {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return ErrInvalidRequest
	}
	n, err := r.q.DeleteGalleryItem(ctx, pgID)
	if err != nil {
		return fmt.Errorf("delete gallery item: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
