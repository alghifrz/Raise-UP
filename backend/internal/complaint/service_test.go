package complaint_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/internal/complaint"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type memoryStore struct {
	byID      map[string]db.Complaint
	createErr error
	attempts  int
}

func newMemoryStore() *memoryStore {
	return &memoryStore{byID: make(map[string]db.Complaint)}
}

func (m *memoryStore) GetByID(_ context.Context, id string) (db.Complaint, error) {
	item, ok := m.byID[id]
	if !ok {
		return db.Complaint{}, complaint.ErrNotFound
	}
	return item, nil
}

func (m *memoryStore) List(_ context.Context, search string, status *db.ComplaintStatus, urgency *db.ComplaintUrgency, category string, limit, offset int32) ([]db.Complaint, error) {
	items := make([]db.Complaint, 0)
	for _, item := range m.byID {
		if search != "" {
			needle := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(item.Ref), needle) &&
				!strings.Contains(strings.ToLower(item.ResidentName), needle) &&
				!strings.Contains(strings.ToLower(item.Phone), needle) &&
				!strings.Contains(strings.ToLower(item.Message), needle) {
				continue
			}
		}
		if status != nil && item.Status != *status {
			continue
		}
		if urgency != nil && item.Urgency != *urgency {
			continue
		}
		if category != "" && item.Category != category {
			continue
		}
		items = append(items, item)
	}
	if offset >= int32(len(items)) {
		return []db.Complaint{}, nil
	}
	items = items[offset:]
	if int32(len(items)) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (m *memoryStore) Count(_ context.Context, search string, status *db.ComplaintStatus, urgency *db.ComplaintUrgency, category string) (int64, error) {
	var total int64
	for _, item := range m.byID {
		if search != "" {
			needle := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(item.Ref), needle) &&
				!strings.Contains(strings.ToLower(item.ResidentName), needle) &&
				!strings.Contains(strings.ToLower(item.Phone), needle) &&
				!strings.Contains(strings.ToLower(item.Message), needle) {
				continue
			}
		}
		if status != nil && item.Status != *status {
			continue
		}
		if urgency != nil && item.Urgency != *urgency {
			continue
		}
		if category != "" && item.Category != category {
			continue
		}
		total++
	}
	return total, nil
}

func (m *memoryStore) Create(_ context.Context, arg db.CreateComplaintParams) (db.Complaint, error) {
	m.attempts++
	if m.createErr != nil {
		return db.Complaint{}, m.createErr
	}
	id := uuid.New()
	now := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	item := db.Complaint{
		ID:           pgtype.UUID{Bytes: id, Valid: true},
		Ref:          arg.Ref,
		ResidentID:   arg.ResidentID,
		ResidentName: arg.ResidentName,
		Phone:        arg.Phone,
		Block:        arg.Block,
		Category:     arg.Category,
		Urgency:      arg.Urgency,
		Status:       arg.Status,
		Message:      arg.Message,
		ReceivedAt:   now,
		UpdatedAt:    now,
	}
	m.byID[id.String()] = item
	return item, nil
}

func (m *memoryStore) Update(_ context.Context, arg db.UpdateComplaintParams) (db.Complaint, error) {
	id, err := uuidutil.ToString(arg.ID)
	if err != nil {
		return db.Complaint{}, err
	}
	item, ok := m.byID[id]
	if !ok {
		return db.Complaint{}, complaint.ErrNotFound
	}
	item.ResidentID = arg.ResidentID
	item.ResidentName = arg.ResidentName
	item.Phone = arg.Phone
	item.Block = arg.Block
	item.Category = arg.Category
	item.Urgency = arg.Urgency
	item.Message = arg.Message
	item.UpdatedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	m.byID[id] = item
	return item, nil
}

func (m *memoryStore) UpdateStatus(_ context.Context, id string, status db.ComplaintStatus) (db.Complaint, error) {
	item, ok := m.byID[id]
	if !ok {
		return db.Complaint{}, complaint.ErrNotFound
	}
	item.Status = status
	item.UpdatedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	m.byID[id] = item
	return item, nil
}

func (m *memoryStore) Delete(_ context.Context, id string) error {
	if _, ok := m.byID[id]; !ok {
		return complaint.ErrNotFound
	}
	delete(m.byID, id)
	return nil
}

type memoryResidents struct {
	byID map[string]db.Resident
}

func newMemoryResidents() *memoryResidents {
	return &memoryResidents{byID: make(map[string]db.Resident)}
}

func (m *memoryResidents) GetByID(_ context.Context, id string) (db.Resident, error) {
	item, ok := m.byID[id]
	if !ok {
		return db.Resident{}, complaint.ErrResidentNotFound
	}
	return item, nil
}

func (m *memoryResidents) add(name, phone string) db.Resident {
	id := uuid.New()
	item := db.Resident{
		ID:    pgtype.UUID{Bytes: id, Valid: true},
		Name:  name,
		Phone: phone,
	}
	m.byID[id.String()] = item
	return item
}

func newService(store *memoryStore, residents *memoryResidents) *complaint.Service {
	return complaint.NewService(store, residents)
}

func TestCreateValidDefaultsUrgency(t *testing.T) {
	service := newService(newMemoryStore(), newMemoryResidents())
	got, err := service.Create(context.Background(), complaint.CreateRequest{
		ResidentName: "  Budi  ",
		Phone:        " 08123 ",
		Category:     " Kebersihan ",
		Message:      " Ada sampah ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got.Urgency != db.ComplaintUrgencyNORMAL || got.Status != db.ComplaintStatusBARU {
		t.Fatalf("unexpected defaults: %+v", got)
	}
	if got.ResidentName != "Budi" || got.Phone != "08123" || got.Category != "Kebersihan" {
		t.Fatalf("unexpected normalization: %+v", got)
	}
	if !strings.HasPrefix(got.Ref, "CMP-") {
		t.Fatalf("unexpected ref: %s", got.Ref)
	}
}

func TestCreateInvalidUrgency(t *testing.T) {
	service := newService(newMemoryStore(), newMemoryResidents())
	_, err := service.Create(context.Background(), complaint.CreateRequest{
		ResidentName: "Budi",
		Phone:        "08123",
		Category:     "Umum",
		Urgency:      "HIGH",
		Message:      "test",
	})
	if !errors.Is(err, complaint.ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
}

func TestCreateMissingFields(t *testing.T) {
	service := newService(newMemoryStore(), newMemoryResidents())
	cases := []complaint.CreateRequest{
		{Phone: "081", Category: "A", Message: "m"},
		{ResidentName: "Budi", Category: "A", Message: "m"},
		{ResidentName: "Budi", Phone: "081", Message: "m"},
		{ResidentName: "Budi", Phone: "081", Category: "A"},
	}
	for _, req := range cases {
		if _, err := service.Create(context.Background(), req); !errors.Is(err, complaint.ErrInvalidRequest) {
			t.Fatalf("Create(%+v) error = %v, want ErrInvalidRequest", req, err)
		}
	}
}

func TestCreateResidentNotFound(t *testing.T) {
	id := uuid.NewString()
	service := newService(newMemoryStore(), newMemoryResidents())
	_, err := service.Create(context.Background(), complaint.CreateRequest{
		ResidentID:   &id,
		ResidentName: "ignored",
		Phone:        "ignored",
		Category:     "Umum",
		Message:      "test",
	})
	if !errors.Is(err, complaint.ErrResidentNotFound) {
		t.Fatalf("error = %v, want ErrResidentNotFound", err)
	}
}

func TestCreateResidentSnapshotAuthoritative(t *testing.T) {
	residents := newMemoryResidents()
	resident := residents.add("Official Name", "081111")
	id, _ := uuidutil.ToString(resident.ID)
	service := newService(newMemoryStore(), residents)

	got, err := service.Create(context.Background(), complaint.CreateRequest{
		ResidentID:   &id,
		ResidentName: "Wrong Name",
		Phone:        "000",
		Category:     "Umum",
		Message:      "test",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got.ResidentName != "Official Name" || got.Phone != "081111" {
		t.Fatalf("snapshot not authoritative: %+v", got)
	}
	if got.ResidentID == nil || *got.ResidentID != id {
		t.Fatalf("unexpected resident_id: %+v", got.ResidentID)
	}
}

func TestUpdatePartialAndResidentReassignment(t *testing.T) {
	store := newMemoryStore()
	residents := newMemoryResidents()
	service := newService(store, residents)

	created, err := service.Create(context.Background(), complaint.CreateRequest{
		ResidentName: "Anon",
		Phone:        "08000",
		Category:     "Umum",
		Message:      "awal",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	msg := "updated"
	updated, err := service.Update(context.Background(), created.ID, complaint.UpdateRequest{Message: &msg})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Message != "updated" || updated.ResidentName != "Anon" {
		t.Fatalf("unexpected partial update: %+v", updated)
	}

	resident := residents.add("Linked", "08999")
	rid, _ := uuidutil.ToString(resident.ID)
	reassigned, err := service.Update(context.Background(), created.ID, complaint.UpdateRequest{
		ResidentID: complaint.OptionalUUID{Present: true, Valid: true, Value: rid},
	})
	if err != nil {
		t.Fatalf("Update() reassignment error = %v", err)
	}
	if reassigned.ResidentName != "Linked" || reassigned.Phone != "08999" {
		t.Fatalf("snapshot not refreshed: %+v", reassigned)
	}

	cleared, err := service.Update(context.Background(), created.ID, complaint.UpdateRequest{
		ResidentID: complaint.OptionalUUID{Present: true, Valid: false},
	})
	if err != nil {
		t.Fatalf("Update() clear error = %v", err)
	}
	if cleared.ResidentID != nil {
		t.Fatal("expected resident_id null")
	}
	if cleared.ResidentName != "Linked" || cleared.Phone != "08999" {
		t.Fatalf("snapshot should be preserved: %+v", cleared)
	}
}

func TestUpdateInvalidResidentID(t *testing.T) {
	service := newService(newMemoryStore(), newMemoryResidents())
	created, err := service.Create(context.Background(), complaint.CreateRequest{
		ResidentName: "Anon",
		Phone:        "08000",
		Category:     "Umum",
		Message:      "awal",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	_, err = service.Update(context.Background(), created.ID, complaint.UpdateRequest{
		ResidentID: complaint.OptionalUUID{Present: true, Valid: true, Value: "bad"},
	})
	if !errors.Is(err, complaint.ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
}

func TestStatusTransitions(t *testing.T) {
	service := newService(newMemoryStore(), newMemoryResidents())
	created, err := service.Create(context.Background(), complaint.CreateRequest{
		ResidentName: "Anon",
		Phone:        "08000",
		Category:     "Umum",
		Message:      "awal",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	assertTransition := func(fromID, status string, want db.ComplaintStatus) {
		t.Helper()
		got, err := service.UpdateStatus(context.Background(), fromID, status)
		if err != nil {
			t.Fatalf("UpdateStatus(%s) error = %v", status, err)
		}
		if got.Status != want {
			t.Fatalf("status = %s, want %s", got.Status, want)
		}
	}

	assertTransition(created.ID, "DIPROSES", db.ComplaintStatusDIPROSES)
	assertTransition(created.ID, "SELESAI", db.ComplaintStatusSELESAI)
	if _, err := service.UpdateStatus(context.Background(), created.ID, "DIPROSES"); !errors.Is(err, complaint.ErrInvalidStatusTransition) {
		t.Fatalf("SELESAI transition error = %v", err)
	}

	rejected, _ := service.Create(context.Background(), complaint.CreateRequest{
		ResidentName: "Anon", Phone: "08000", Category: "Umum", Message: "x",
	})
	assertTransition(rejected.ID, "DITOLAK", db.ComplaintStatusDITOLAK)
	if _, err := service.UpdateStatus(context.Background(), rejected.ID, "BARU"); !errors.Is(err, complaint.ErrInvalidStatusTransition) {
		t.Fatalf("DITOLAK transition error = %v", err)
	}

	processed, _ := service.Create(context.Background(), complaint.CreateRequest{
		ResidentName: "Anon", Phone: "08000", Category: "Umum", Message: "y",
	})
	assertTransition(processed.ID, "DIPROSES", db.ComplaintStatusDIPROSES)
	assertTransition(processed.ID, "DITOLAK", db.ComplaintStatusDITOLAK)

	baru, _ := service.Create(context.Background(), complaint.CreateRequest{
		ResidentName: "Anon", Phone: "08000", Category: "Umum", Message: "z",
	})
	assertTransition(baru.ID, "DITOLAK", db.ComplaintStatusDITOLAK)

	if _, err := service.UpdateStatus(context.Background(), baru.ID, "UNKNOWN"); !errors.Is(err, complaint.ErrInvalidRequest) {
		t.Fatalf("invalid status error = %v", err)
	}
}

func TestDeleteSuccessAndNotFound(t *testing.T) {
	service := newService(newMemoryStore(), newMemoryResidents())
	created, err := service.Create(context.Background(), complaint.CreateRequest{
		ResidentName: "Anon", Phone: "08000", Category: "Umum", Message: "x",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := service.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if err := service.Delete(context.Background(), created.ID); !errors.Is(err, complaint.ErrNotFound) {
		t.Fatalf("Delete() error = %v, want ErrNotFound", err)
	}
}

func TestListFiltersAndPagination(t *testing.T) {
	store := newMemoryStore()
	service := newService(store, newMemoryResidents())
	_, _ = service.Create(context.Background(), complaint.CreateRequest{
		ResidentName: "Budi", Phone: "08111", Category: "Kebersihan", Urgency: "PRIORITY", Message: "sampah",
	})
	_, _ = service.Create(context.Background(), complaint.CreateRequest{
		ResidentName: "Siti", Phone: "08222", Category: "Keamanan", Urgency: "NORMAL", Message: "malam",
	})

	defaults, err := service.List(context.Background(), 0, 0, "", "", "", "")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if defaults.Meta.Page != 1 || defaults.Meta.PageSize != 20 || defaults.Meta.Total != 2 {
		t.Fatalf("unexpected defaults: %+v", defaults.Meta)
	}

	limited, err := service.List(context.Background(), 1, 500, "", "", "", "")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if limited.Meta.PageSize != 100 {
		t.Fatalf("page_size = %d, want 100", limited.Meta.PageSize)
	}

	bySearch, err := service.List(context.Background(), 1, 20, "sampah", "", "", "")
	if err != nil || bySearch.Meta.Total != 1 || bySearch.Items[0].ResidentName != "Budi" {
		t.Fatalf("search failed: %+v err=%v", bySearch, err)
	}
	byStatus, err := service.List(context.Background(), 1, 20, "", "BARU", "", "")
	if err != nil || byStatus.Meta.Total != 2 {
		t.Fatalf("status filter failed: %+v err=%v", byStatus, err)
	}
	byUrgency, err := service.List(context.Background(), 1, 20, "", "", "PRIORITY", "")
	if err != nil || byUrgency.Meta.Total != 1 {
		t.Fatalf("urgency filter failed: %+v err=%v", byUrgency, err)
	}
	byCategory, err := service.List(context.Background(), 1, 20, "", "", "", "Keamanan")
	if err != nil || byCategory.Meta.Total != 1 || byCategory.Items[0].Category != "Keamanan" {
		t.Fatalf("category filter failed: %+v err=%v", byCategory, err)
	}
}

func TestGenerateRefFormat(t *testing.T) {
	now := time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)
	ref, err := complaint.GenerateRef(now)
	if err != nil {
		t.Fatalf("GenerateRef() error = %v", err)
	}
	if !strings.HasPrefix(ref, "CMP-20260920-") {
		t.Fatalf("ref = %s, want CMP-20260920-XXXX", ref)
	}
	parts := strings.Split(ref, "-")
	if len(parts) != 3 || len(parts[2]) != 4 {
		t.Fatalf("unexpected ref shape: %s", ref)
	}
}

func TestCreateRetriesOnUniqueViolation(t *testing.T) {
	store := newMemoryStore()
	store.createErr = &pgconn.PgError{Code: "23505"}
	service := complaint.NewService(store, newMemoryResidents())

	_, err := service.Create(context.Background(), complaint.CreateRequest{
		ResidentName: "Anon", Phone: "08000", Category: "Umum", Message: "x",
	})
	if err == nil {
		t.Fatal("expected error after exhausted retries")
	}
	if store.attempts < 5 {
		t.Fatalf("attempts = %d, want >= 5", store.attempts)
	}
}
