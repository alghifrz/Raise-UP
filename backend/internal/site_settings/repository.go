package sitesettings

import (
	"context"
	"errors"
	"fmt"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides site settings persistence via sqlc.
type Repository struct {
	q *db.Queries
}

// NewRepository creates a site settings repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

// Get returns the singleton site settings row.
func (r *Repository) Get(ctx context.Context) (db.SiteSetting, error) {
	item, err := r.q.GetSiteSettings(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.SiteSetting{}, ErrNotFound
		}
		return db.SiteSetting{}, fmt.Errorf("get site settings: %w", err)
	}
	return item, nil
}

// Update patches the singleton site settings row by id.
func (r *Repository) Update(ctx context.Context, arg db.UpdateSiteSettingsParams) (db.SiteSetting, error) {
	item, err := r.q.UpdateSiteSettings(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.SiteSetting{}, ErrNotFound
		}
		return db.SiteSetting{}, fmt.Errorf("update site settings: %w", err)
	}
	return item, nil
}
