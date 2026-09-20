package village_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/internal/village"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type memoryStore struct {
	profile   *db.VillageProfile
	officials map[string]db.VillageOfficial
}

func newStoreWithProfile() *memoryStore {
	id := uuid.New()
	return &memoryStore{
		profile: &db.VillageProfile{
			ID:        pgtype.UUID{Bytes: id, Valid: true},
			History:   "old history",
			Vision:    "old vision",
			Mission:   "old mission",
			UpdatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		},
		officials: make(map[string]db.VillageOfficial),
	}
}

func (m *memoryStore) GetProfile(_ context.Context) (db.VillageProfile, error) {
	if m.profile == nil {
		return db.VillageProfile{}, village.ErrProfileNotFound
	}
	return *m.profile, nil
}

func (m *memoryStore) UpdateProfile(_ context.Context, arg db.UpdateVillageProfileParams) (db.VillageProfile, error) {
	if m.profile == nil {
		return db.VillageProfile{}, village.ErrProfileNotFound
	}
	item := *m.profile
	if arg.History != nil {
		item.History = *arg.History
	}
	if arg.Vision != nil {
		item.Vision = *arg.Vision
	}
	if arg.Mission != nil {
		item.Mission = *arg.Mission
	}
	item.UpdatedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	m.profile = &item
	return item, nil
}

func (m *memoryStore) GetOfficialByID(_ context.Context, id string) (db.VillageOfficial, error) {
	item, ok := m.officials[id]
	if !ok {
		return db.VillageOfficial{}, village.ErrOfficialNotFound
	}
	return item, nil
}

func (m *memoryStore) ListOfficials(_ context.Context, profileID pgtype.UUID) ([]db.VillageOfficial, error) {
	items := make([]db.VillageOfficial, 0)
	for _, item := range m.officials {
		if item.ProfileID == profileID {
			items = append(items, item)
		}
	}
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].SortOrder < items[i].SortOrder ||
				(items[j].SortOrder == items[i].SortOrder && bytesLess(items[j].ID.Bytes[:], items[i].ID.Bytes[:])) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	return items, nil
}

func bytesLess(a, b []byte) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return true
		}
		if a[i] > b[i] {
			return false
		}
	}
	return len(a) < len(b)
}

func (m *memoryStore) CreateOfficial(_ context.Context, arg db.CreateVillageOfficialParams) (db.VillageOfficial, error) {
	id := uuid.New()
	item := db.VillageOfficial{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		ProfileID: arg.ProfileID,
		Name:      arg.Name,
		Position:  arg.Position,
		PhotoUrl:  arg.PhotoUrl,
		SortOrder: arg.SortOrder,
	}
	m.officials[id.String()] = item
	return item, nil
}

func (m *memoryStore) UpdateOfficial(_ context.Context, arg db.UpdateVillageOfficialParams) (db.VillageOfficial, error) {
	id, err := uuidutil.ToString(arg.ID)
	if err != nil {
		return db.VillageOfficial{}, err
	}
	item, ok := m.officials[id]
	if !ok {
		return db.VillageOfficial{}, village.ErrOfficialNotFound
	}
	if arg.Name != nil {
		item.Name = *arg.Name
	}
	if arg.Position != nil {
		item.Position = *arg.Position
	}
	if arg.PhotoUrl != nil {
		item.PhotoUrl = arg.PhotoUrl
	}
	if arg.SortOrder != nil {
		item.SortOrder = *arg.SortOrder
	}
	m.officials[id] = item
	return item, nil
}

func (m *memoryStore) DeleteOfficial(_ context.Context, id string) error {
	if _, ok := m.officials[id]; !ok {
		return village.ErrOfficialNotFound
	}
	delete(m.officials, id)
	return nil
}

func TestProfileGetUpdateTrim(t *testing.T) {
	store := newStoreWithProfile()
	svc := village.NewService(store)

	got, err := svc.GetProfile(context.Background())
	if err != nil || got.History != "old history" {
		t.Fatalf("get: %v %+v", err, got)
	}

	history := "  new history  "
	updated, err := svc.UpdateProfile(context.Background(), village.UpdateProfileRequest{History: &history})
	if err != nil || updated.History != "new history" {
		t.Fatalf("update: %v %+v", err, updated)
	}

	svcMissing := village.NewService(&memoryStore{officials: map[string]db.VillageOfficial{}})
	_, err = svcMissing.GetProfile(context.Background())
	if !errors.Is(err, village.ErrProfileNotFound) {
		t.Fatalf("expected missing profile, got %v", err)
	}
}

func TestOfficialsCRUD(t *testing.T) {
	store := newStoreWithProfile()
	svc := village.NewService(store)
	profileID, _ := uuidutil.ToString(store.profile.ID)

	order := int32(2)
	created, err := svc.CreateOfficial(context.Background(), village.CreateOfficialRequest{
		Name: " Budi Santoso ", Position: " Ketua RT ", PhotoURL: " https://example.com/budi.jpg ", SortOrder: &order,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Name != "Budi Santoso" || created.Position != "Ketua RT" || created.ProfileID != profileID || created.SortOrder != 2 {
		t.Fatalf("unexpected official: %+v", created)
	}
	if created.PhotoURL == nil || *created.PhotoURL != "https://example.com/budi.jpg" {
		t.Fatalf("unexpected photo: %v", created.PhotoURL)
	}

	_, err = svc.CreateOfficial(context.Background(), village.CreateOfficialRequest{Position: "X"})
	if !errors.Is(err, village.ErrInvalidRequest) {
		t.Fatalf("required name: %v", err)
	}
	_, err = svc.CreateOfficial(context.Background(), village.CreateOfficialRequest{Name: "X"})
	if !errors.Is(err, village.ErrInvalidRequest) {
		t.Fatalf("required position: %v", err)
	}
	neg := int32(-1)
	_, err = svc.CreateOfficial(context.Background(), village.CreateOfficialRequest{Name: "X", Position: "Y", SortOrder: &neg})
	if !errors.Is(err, village.ErrInvalidRequest) {
		t.Fatalf("negative sort: %v", err)
	}

	_, _ = svc.CreateOfficial(context.Background(), village.CreateOfficialRequest{Name: "Ani", Position: "Wakil", SortOrder: int32Ptr(1)})
	list, err := svc.ListOfficials(context.Background())
	if err != nil || len(list) != 2 || list[0].SortOrder != 1 {
		t.Fatalf("list ordering: %+v err=%v", list, err)
	}

	name := "Budi Updated"
	updated, err := svc.UpdateOfficial(context.Background(), created.ID, village.UpdateOfficialRequest{Name: &name})
	if err != nil || updated.Name != "Budi Updated" || updated.ProfileID != profileID {
		t.Fatalf("update: %v %+v", err, updated)
	}

	if err := svc.DeleteOfficial(context.Background(), created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	err = svc.DeleteOfficial(context.Background(), "bad")
	if !errors.Is(err, village.ErrInvalidRequest) {
		t.Fatalf("invalid uuid: %v", err)
	}
	err = svc.DeleteOfficial(context.Background(), uuid.NewString())
	if !errors.Is(err, village.ErrOfficialNotFound) {
		t.Fatalf("missing: %v", err)
	}
}

func TestUpdateOfficialRequestOmitsProfileID(t *testing.T) {
	fields := map[string]struct{}{}
	tpe := reflect.TypeOf(village.UpdateOfficialRequest{})
	for i := 0; i < tpe.NumField(); i++ {
		tag := tpe.Field(i).Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		fields[name] = struct{}{}
	}
	if _, ok := fields["profile_id"]; ok {
		t.Fatal("UpdateOfficialRequest must not expose profile_id")
	}

	raw := []byte(`{"name":"X","position":"Y","profile_id":"should-be-ignored"}`)
	var req village.UpdateOfficialRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// profile_id has nowhere to land on the struct
	if req.Name == nil || *req.Name != "X" {
		t.Fatalf("unexpected req: %+v", req)
	}
}

func int32Ptr(v int32) *int32 { return &v }
