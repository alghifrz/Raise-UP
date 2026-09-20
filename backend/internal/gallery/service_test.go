package gallery_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/internal/gallery"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type memoryStore struct {
	byID map[string]db.GalleryItem
}

func newMemoryStore() *memoryStore {
	return &memoryStore{byID: make(map[string]db.GalleryItem)}
}

func (m *memoryStore) GetByID(_ context.Context, id string) (db.GalleryItem, error) {
	item, ok := m.byID[id]
	if !ok {
		return db.GalleryItem{}, gallery.ErrNotFound
	}
	return item, nil
}

func (m *memoryStore) List(_ context.Context, search string, limit, offset int32) ([]db.GalleryItem, error) {
	items := make([]db.GalleryItem, 0)
	for _, item := range m.byID {
		if search != "" {
			needle := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(item.Caption), needle) &&
				!strings.Contains(strings.ToLower(item.ImageUrl), needle) {
				continue
			}
		}
		items = append(items, item)
	}
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].SortOrder < items[i].SortOrder ||
				(items[j].SortOrder == items[i].SortOrder && items[j].CreatedAt.Time.After(items[i].CreatedAt.Time)) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	if offset >= int32(len(items)) {
		return []db.GalleryItem{}, nil
	}
	items = items[offset:]
	if int32(len(items)) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (m *memoryStore) Count(ctx context.Context, search string) (int64, error) {
	items, _ := m.List(ctx, search, 100000, 0)
	return int64(len(items)), nil
}

func (m *memoryStore) Create(_ context.Context, arg db.CreateGalleryItemParams) (db.GalleryItem, error) {
	id := uuid.New()
	item := db.GalleryItem{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		ImageUrl:    arg.ImageUrl,
		StoragePath: arg.StoragePath,
		Caption:     arg.Caption,
		SortOrder:   arg.SortOrder,
		CreatedAt:   pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}
	m.byID[id.String()] = item
	return item, nil
}

func (m *memoryStore) Update(_ context.Context, arg db.UpdateGalleryItemParams) (db.GalleryItem, error) {
	id, err := uuidutil.ToString(arg.ID)
	if err != nil {
		return db.GalleryItem{}, err
	}
	item, ok := m.byID[id]
	if !ok {
		return db.GalleryItem{}, gallery.ErrNotFound
	}
	if arg.ImageUrl != nil {
		item.ImageUrl = *arg.ImageUrl
	}
	if arg.StoragePath != nil {
		item.StoragePath = *arg.StoragePath
	}
	if arg.Caption != nil {
		item.Caption = *arg.Caption
	}
	if arg.SortOrder != nil {
		item.SortOrder = *arg.SortOrder
	}
	m.byID[id] = item
	return item, nil
}

func (m *memoryStore) Delete(_ context.Context, id string) error {
	if _, ok := m.byID[id]; !ok {
		return gallery.ErrNotFound
	}
	delete(m.byID, id)
	return nil
}

func TestCreateValidAndOptional(t *testing.T) {
	svc := gallery.NewService(newMemoryStore())
	order := int32(1)
	item, err := svc.Create(context.Background(), gallery.CreateRequest{
		ImageURL:    " https://example.com/image.jpg ",
		StoragePath: " gallery/2026/image.jpg ",
		Caption:     " Kerja bakti warga ",
		SortOrder:   &order,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if item.ImageURL != "https://example.com/image.jpg" || item.SortOrder != 1 || item.Caption != "Kerja bakti warga" {
		t.Fatalf("unexpected item: %+v", item)
	}

	item2, err := svc.Create(context.Background(), gallery.CreateRequest{ImageURL: "https://example.com/a.jpg"})
	if err != nil {
		t.Fatalf("create optional: %v", err)
	}
	if item2.SortOrder != 0 || item2.StoragePath != "" || item2.Caption != "" {
		t.Fatalf("unexpected defaults: %+v", item2)
	}
}

func TestCreateValidation(t *testing.T) {
	svc := gallery.NewService(newMemoryStore())
	_, err := svc.Create(context.Background(), gallery.CreateRequest{})
	if !errors.Is(err, gallery.ErrInvalidRequest) {
		t.Fatalf("missing image_url: %v", err)
	}
	_, err = svc.Create(context.Background(), gallery.CreateRequest{ImageURL: "   "})
	if !errors.Is(err, gallery.ErrInvalidRequest) {
		t.Fatalf("empty image_url: %v", err)
	}
	neg := int32(-1)
	_, err = svc.Create(context.Background(), gallery.CreateRequest{ImageURL: "https://x", SortOrder: &neg})
	if !errors.Is(err, gallery.ErrInvalidRequest) {
		t.Fatalf("negative sort: %v", err)
	}
}

func TestGetListUpdateDelete(t *testing.T) {
	svc := gallery.NewService(newMemoryStore())
	a, _ := svc.Create(context.Background(), gallery.CreateRequest{ImageURL: "https://a.jpg", Caption: "Alpha"})
	_, _ = svc.Create(context.Background(), gallery.CreateRequest{ImageURL: "https://b.jpg", Caption: "Beta kerja"})

	got, err := svc.Get(context.Background(), a.ID)
	if err != nil || got.ID != a.ID {
		t.Fatalf("get: %v %+v", err, got)
	}

	list, err := svc.List(context.Background(), 1, 1, "")
	if err != nil || list.Meta.Total != 2 || len(list.Items) != 1 {
		t.Fatalf("pagination: %+v err=%v", list, err)
	}
	search, err := svc.List(context.Background(), 1, 20, "kerja")
	if err != nil || search.Meta.Total != 1 {
		t.Fatalf("search: %+v err=%v", search, err)
	}

	caption := "Updated"
	updated, err := svc.Update(context.Background(), a.ID, gallery.UpdateRequest{Caption: &caption})
	if err != nil || updated.Caption != "Updated" {
		t.Fatalf("update: %v %+v", err, updated)
	}

	if err := svc.Delete(context.Background(), a.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err = svc.Get(context.Background(), a.ID)
	if !errors.Is(err, gallery.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	_, err = svc.Get(context.Background(), "bad-id")
	if !errors.Is(err, gallery.ErrInvalidRequest) {
		t.Fatalf("expected invalid uuid, got %v", err)
	}
	err = svc.Delete(context.Background(), uuid.NewString())
	if !errors.Is(err, gallery.ErrNotFound) {
		t.Fatalf("expected missing delete, got %v", err)
	}
}
