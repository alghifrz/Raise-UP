package resident_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/internal/resident"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type memoryStore struct {
	byID map[string]db.Resident
}

func newMemoryStore() *memoryStore {
	return &memoryStore{byID: make(map[string]db.Resident)}
}

func (m *memoryStore) GetByID(_ context.Context, id string) (db.Resident, error) {
	item, ok := m.byID[id]
	if !ok {
		return db.Resident{}, resident.ErrNotFound
	}
	return item, nil
}

func (m *memoryStore) List(_ context.Context, search string, gender *db.Gender, limit, offset int32) ([]db.Resident, error) {
	items := make([]db.Resident, 0)
	for _, item := range m.byID {
		if search != "" {
			needle := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(item.Name), needle) &&
				!strings.Contains(strings.ToLower(item.Phone), needle) {
				continue
			}
		}
		if gender != nil && item.Gender != *gender {
			continue
		}
		items = append(items, item)
	}

	if offset >= int32(len(items)) {
		return []db.Resident{}, nil
	}
	items = items[offset:]
	if int32(len(items)) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (m *memoryStore) Count(_ context.Context, search string, gender *db.Gender) (int64, error) {
	var total int64
	for _, item := range m.byID {
		if search != "" {
			needle := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(item.Name), needle) &&
				!strings.Contains(strings.ToLower(item.Phone), needle) {
				continue
			}
		}
		if gender != nil && item.Gender != *gender {
			continue
		}
		total++
	}
	return total, nil
}

func (m *memoryStore) Create(_ context.Context, name, phone string, gender db.Gender) (db.Resident, error) {
	id := uuid.New()
	now := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	item := db.Resident{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		Name:      name,
		Phone:     phone,
		Gender:    gender,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.byID[id.String()] = item
	return item, nil
}

func (m *memoryStore) Update(_ context.Context, id string, name, phone string, gender db.Gender) (db.Resident, error) {
	item, ok := m.byID[id]
	if !ok {
		return db.Resident{}, resident.ErrNotFound
	}
	item.Name = name
	item.Phone = phone
	item.Gender = gender
	item.UpdatedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	m.byID[id] = item
	return item, nil
}

func (m *memoryStore) Delete(_ context.Context, id string) error {
	if _, ok := m.byID[id]; !ok {
		return resident.ErrNotFound
	}
	delete(m.byID, id)
	return nil
}

type conflictStore struct {
	*memoryStore
}

func (c *conflictStore) Delete(_ context.Context, _ string) error {
	return resident.ErrDeleteConflict
}

func TestCreateValid(t *testing.T) {
	service := resident.NewService(newMemoryStore())
	got, err := service.Create(context.Background(), resident.CreateRequest{
		Name:   "  Budi  ",
		Phone:  " 08123 ",
		Gender: "LAKI_LAKI",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got.Name != "Budi" || got.Phone != "08123" || got.Gender != db.GenderLAKILAKI {
		t.Fatalf("unexpected resident: %+v", got)
	}
	if got.ID == "" || got.CreatedAt == "" {
		t.Fatal("expected id and timestamps")
	}
}

func TestCreateMissingName(t *testing.T) {
	service := resident.NewService(newMemoryStore())
	_, err := service.Create(context.Background(), resident.CreateRequest{
		Name:   "   ",
		Phone:  "08123",
		Gender: "LAKI_LAKI",
	})
	if !errors.Is(err, resident.ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
}

func TestCreateMissingPhone(t *testing.T) {
	service := resident.NewService(newMemoryStore())
	_, err := service.Create(context.Background(), resident.CreateRequest{
		Name:   "Budi",
		Phone:  "",
		Gender: "LAKI_LAKI",
	})
	if !errors.Is(err, resident.ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
}

func TestCreateInvalidGender(t *testing.T) {
	service := resident.NewService(newMemoryStore())
	_, err := service.Create(context.Background(), resident.CreateRequest{
		Name:   "Budi",
		Phone:  "08123",
		Gender: "OTHER",
	})
	if !errors.Is(err, resident.ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
}

func TestUpdatePartial(t *testing.T) {
	store := newMemoryStore()
	service := resident.NewService(store)
	created, err := service.Create(context.Background(), resident.CreateRequest{
		Name:   "Budi",
		Phone:  "08123",
		Gender: "LAKI_LAKI",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	name := "Budi Santoso"
	updated, err := service.Update(context.Background(), created.ID, resident.UpdateRequest{Name: &name})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name != "Budi Santoso" || updated.Phone != "08123" || updated.Gender != db.GenderLAKILAKI {
		t.Fatalf("unexpected update result: %+v", updated)
	}
}

func TestUpdateValidGenderAndPhone(t *testing.T) {
	store := newMemoryStore()
	service := resident.NewService(store)
	created, err := service.Create(context.Background(), resident.CreateRequest{
		Name:   "Siti",
		Phone:  "08111",
		Gender: "PEREMPUAN",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	phone := "08999"
	gender := "LAKI_LAKI"
	updated, err := service.Update(context.Background(), created.ID, resident.UpdateRequest{
		Phone:  &phone,
		Gender: &gender,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Phone != "08999" || updated.Gender != db.GenderLAKILAKI || updated.Name != "Siti" {
		t.Fatalf("unexpected update result: %+v", updated)
	}
}

func TestGetInvalidUUID(t *testing.T) {
	service := resident.NewService(newMemoryStore())
	_, err := service.Get(context.Background(), "not-a-uuid")
	if !errors.Is(err, resident.ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
}

func TestGetNotFound(t *testing.T) {
	service := resident.NewService(newMemoryStore())
	_, err := service.Get(context.Background(), uuid.NewString())
	if !errors.Is(err, resident.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestDeleteSuccess(t *testing.T) {
	service := resident.NewService(newMemoryStore())
	created, err := service.Create(context.Background(), resident.CreateRequest{
		Name:   "Budi",
		Phone:  "08123",
		Gender: "LAKI_LAKI",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := service.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	_, err = service.Get(context.Background(), created.ID)
	if !errors.Is(err, resident.ErrNotFound) {
		t.Fatalf("Get() after delete error = %v, want ErrNotFound", err)
	}
}

func TestDeleteConflict(t *testing.T) {
	base := newMemoryStore()
	service := resident.NewService(&conflictStore{memoryStore: base})
	created, err := service.Create(context.Background(), resident.CreateRequest{
		Name:   "Budi",
		Phone:  "08123",
		Gender: "LAKI_LAKI",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	err = service.Delete(context.Background(), created.ID)
	if !errors.Is(err, resident.ErrDeleteConflict) {
		t.Fatalf("Delete() error = %v, want ErrDeleteConflict", err)
	}
}

func TestListPaginationDefaultsAndLimit(t *testing.T) {
	store := newMemoryStore()
	service := resident.NewService(store)
	for i := 0; i < 5; i++ {
		_, err := service.Create(context.Background(), resident.CreateRequest{
			Name:   "Resident",
			Phone:  uuid.NewString()[:8],
			Gender: "LAKI_LAKI",
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	result, err := service.List(context.Background(), 0, 0, "", "")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Meta.Page != 1 || result.Meta.PageSize != 20 {
		t.Fatalf("unexpected defaults: %+v", result.Meta)
	}

	limited, err := service.List(context.Background(), 1, 1000, "", "")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if limited.Meta.PageSize != 100 {
		t.Fatalf("page_size = %d, want 100 max", limited.Meta.PageSize)
	}
}

func TestListSearchAndGenderFilter(t *testing.T) {
	store := newMemoryStore()
	service := resident.NewService(store)
	_, _ = service.Create(context.Background(), resident.CreateRequest{Name: "Budi", Phone: "08111", Gender: "LAKI_LAKI"})
	_, _ = service.Create(context.Background(), resident.CreateRequest{Name: "Siti", Phone: "08222", Gender: "PEREMPUAN"})

	byName, err := service.List(context.Background(), 1, 20, "bud", "")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if byName.Meta.Total != 1 || byName.Items[0].Name != "Budi" {
		t.Fatalf("unexpected search result: %+v", byName)
	}

	byPhone, err := service.List(context.Background(), 1, 20, "0822", "")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if byPhone.Meta.Total != 1 || byPhone.Items[0].Name != "Siti" {
		t.Fatalf("unexpected phone search: %+v", byPhone)
	}

	byGender, err := service.List(context.Background(), 1, 20, "", "PEREMPUAN")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if byGender.Meta.Total != 1 || byGender.Items[0].Gender != db.GenderPEREMPUAN {
		t.Fatalf("unexpected gender filter: %+v", byGender)
	}
}

func TestListInvalidGender(t *testing.T) {
	service := resident.NewService(newMemoryStore())
	_, err := service.List(context.Background(), 1, 20, "", "UNKNOWN")
	if !errors.Is(err, resident.ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
}

func TestUUIDUtilRoundTrip(t *testing.T) {
	id := uuid.New()
	pgID := pgtype.UUID{Bytes: id, Valid: true}
	got, err := uuidutil.ToString(pgID)
	if err != nil {
		t.Fatalf("ToString() error = %v", err)
	}
	if got != id.String() {
		t.Fatalf("ToString() = %s, want %s", got, id.String())
	}
}
