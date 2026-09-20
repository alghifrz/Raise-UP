package sitesettings_test

import (
	"context"
	"errors"
	"testing"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	sitesettings "github.com/diuk/raiseup/internal/site_settings"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type memoryStore struct {
	item *db.SiteSetting
}

func (m *memoryStore) Get(_ context.Context) (db.SiteSetting, error) {
	if m.item == nil {
		return db.SiteSetting{}, sitesettings.ErrNotFound
	}
	return *m.item, nil
}

func (m *memoryStore) Update(_ context.Context, arg db.UpdateSiteSettingsParams) (db.SiteSetting, error) {
	if m.item == nil {
		return db.SiteSetting{}, sitesettings.ErrNotFound
	}
	item := *m.item
	if arg.SiteName != nil {
		item.SiteName = *arg.SiteName
	}
	if arg.Tagline != nil {
		item.Tagline = *arg.Tagline
	}
	if arg.ChairmanName != nil {
		item.ChairmanName = *arg.ChairmanName
	}
	if arg.Address != nil {
		item.Address = *arg.Address
	}
	if arg.MapsUrl != nil {
		item.MapsUrl = arg.MapsUrl
	}
	item.UpdatedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	m.item = &item
	return item, nil
}

func seedSettings() *memoryStore {
	id := uuid.New()
	return &memoryStore{item: &db.SiteSetting{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		SiteName:  "RAISE UP",
		Tagline:   "Portal",
		UpdatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}}
}

func TestGetAndUpdateSettings(t *testing.T) {
	store := seedSettings()
	svc := sitesettings.NewService(store)

	got, err := svc.Get(context.Background())
	if err != nil || got.SiteName != "RAISE UP" {
		t.Fatalf("get: %v %+v", err, got)
	}

	name := "  RAISE UP RW  "
	maps := " https://maps.example "
	updated, err := svc.Update(context.Background(), sitesettings.UpdateRequest{
		SiteName: &name,
		MapsURL:  &maps,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.SiteName != "RAISE UP RW" || updated.MapsURL == nil || *updated.MapsURL != "https://maps.example" {
		t.Fatalf("unexpected update: %+v", updated)
	}
}

func TestMissingSingleton(t *testing.T) {
	svc := sitesettings.NewService(&memoryStore{})
	_, err := svc.Get(context.Background())
	if !errors.Is(err, sitesettings.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	name := "x"
	_, err = svc.Update(context.Background(), sitesettings.UpdateRequest{SiteName: &name})
	if !errors.Is(err, sitesettings.ErrNotFound) {
		t.Fatalf("expected not found on update, got %v", err)
	}
}
