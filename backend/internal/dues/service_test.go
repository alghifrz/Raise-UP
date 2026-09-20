package dues_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/internal/dues"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type memoryStore struct {
	periods   map[string]db.DuesPeriod
	payments  map[string]db.GetDuesPaymentByIDRow
	residents map[string]db.Resident

	createPeriodErr  error
	updatePeriodErr  error
	createPaymentErr error
	updatePaymentErr error
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		periods:   make(map[string]db.DuesPeriod),
		payments:  make(map[string]db.GetDuesPaymentByIDRow),
		residents: make(map[string]db.Resident),
	}
}

func (m *memoryStore) GetPeriodByID(_ context.Context, id string) (db.DuesPeriod, error) {
	item, ok := m.periods[id]
	if !ok {
		return db.DuesPeriod{}, dues.ErrPeriodNotFound
	}
	return item, nil
}

func (m *memoryStore) ListPeriods(_ context.Context, year, month, half *int32, limit, offset int32) ([]db.DuesPeriod, error) {
	items := make([]db.DuesPeriod, 0)
	for _, item := range m.periods {
		if year != nil && item.Year != *year {
			continue
		}
		if month != nil && item.Month != *month {
			continue
		}
		if half != nil && item.Half != *half {
			continue
		}
		items = append(items, item)
	}
	return pageSlice(items, limit, offset), nil
}

func (m *memoryStore) CountPeriods(_ context.Context, year, month, half *int32) (int64, error) {
	items, _ := m.ListPeriods(context.Background(), year, month, half, 100000, 0)
	return int64(len(items)), nil
}

func (m *memoryStore) CreatePeriod(_ context.Context, year, month, half int32, amount int64) (db.DuesPeriod, error) {
	if m.createPeriodErr != nil {
		return db.DuesPeriod{}, m.createPeriodErr
	}
	for _, existing := range m.periods {
		if existing.Year == year && existing.Month == month && existing.Half == half {
			return db.DuesPeriod{}, dues.ErrPeriodAlreadyExists
		}
	}
	id := uuid.New()
	item := db.DuesPeriod{
		ID:        pgUUID(id),
		Year:      year,
		Month:     month,
		Half:      half,
		Amount:    amount,
		CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}
	m.periods[id.String()] = item
	return item, nil
}

func (m *memoryStore) UpdatePeriod(_ context.Context, arg db.UpdateDuesPeriodParams) (db.DuesPeriod, error) {
	if m.updatePeriodErr != nil {
		return db.DuesPeriod{}, m.updatePeriodErr
	}
	id, err := uuidutil.ToString(arg.ID)
	if err != nil {
		return db.DuesPeriod{}, err
	}
	item, ok := m.periods[id]
	if !ok {
		return db.DuesPeriod{}, dues.ErrPeriodNotFound
	}
	year, month, half := item.Year, item.Month, item.Half
	if arg.Year != nil {
		year = *arg.Year
	}
	if arg.Month != nil {
		month = *arg.Month
	}
	if arg.Half != nil {
		half = *arg.Half
	}
	for otherID, existing := range m.periods {
		if otherID == id {
			continue
		}
		if existing.Year == year && existing.Month == month && existing.Half == half {
			return db.DuesPeriod{}, dues.ErrPeriodAlreadyExists
		}
	}
	if arg.Year != nil {
		item.Year = *arg.Year
	}
	if arg.Month != nil {
		item.Month = *arg.Month
	}
	if arg.Half != nil {
		item.Half = *arg.Half
	}
	if arg.Amount != nil {
		item.Amount = *arg.Amount
	}
	m.periods[id] = item
	return item, nil
}

func (m *memoryStore) DeletePeriod(_ context.Context, id string) error {
	if _, ok := m.periods[id]; !ok {
		return dues.ErrPeriodNotFound
	}
	delete(m.periods, id)
	for paymentID, payment := range m.payments {
		periodID, _ := uuidutil.ToString(payment.PeriodID)
		if periodID == id {
			delete(m.payments, paymentID)
		}
	}
	return nil
}

func (m *memoryStore) GetPaymentByID(_ context.Context, id string) (db.GetDuesPaymentByIDRow, error) {
	item, ok := m.payments[id]
	if !ok {
		return db.GetDuesPaymentByIDRow{}, dues.ErrPaymentNotFound
	}
	return item, nil
}

func (m *memoryStore) ListPayments(_ context.Context, periodID, residentID *string, search string, fromAt, toExclusive *time.Time, limit, offset int32) ([]db.ListDuesPaymentsFilteredRow, error) {
	items := make([]db.ListDuesPaymentsFilteredRow, 0)
	for _, item := range m.payments {
		pID, _ := uuidutil.ToString(item.PeriodID)
		rID, _ := uuidutil.ToString(item.ResidentID)
		if periodID != nil && pID != *periodID {
			continue
		}
		if residentID != nil && rID != *residentID {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(item.ResidentName), strings.ToLower(search)) {
			continue
		}
		if fromAt != nil && item.PaidAt.Time.Before(*fromAt) {
			continue
		}
		if toExclusive != nil && !item.PaidAt.Time.Before(*toExclusive) {
			continue
		}
		items = append(items, db.ListDuesPaymentsFilteredRow{
			ID: item.ID, PeriodID: item.PeriodID, ResidentID: item.ResidentID,
			ResidentName: item.ResidentName, Amount: item.Amount,
			PaidAt: item.PaidAt, CreatedAt: item.CreatedAt,
		})
	}
	return pageSlice(items, limit, offset), nil
}

func (m *memoryStore) CountPayments(ctx context.Context, periodID, residentID *string, search string, fromAt, toExclusive *time.Time) (int64, error) {
	items, _ := m.ListPayments(ctx, periodID, residentID, search, fromAt, toExclusive, 100000, 0)
	return int64(len(items)), nil
}

func (m *memoryStore) CreatePayment(_ context.Context, arg db.CreateDuesPaymentParams) (db.DuesPayment, error) {
	if m.createPaymentErr != nil {
		return db.DuesPayment{}, m.createPaymentErr
	}
	periodID, _ := uuidutil.ToString(arg.PeriodID)
	residentID, _ := uuidutil.ToString(arg.ResidentID)
	for _, existing := range m.payments {
		ep, _ := uuidutil.ToString(existing.PeriodID)
		er, _ := uuidutil.ToString(existing.ResidentID)
		if ep == periodID && er == residentID {
			return db.DuesPayment{}, dues.ErrPaymentAlreadyExists
		}
	}
	resident, ok := m.residents[residentID]
	if !ok {
		return db.DuesPayment{}, dues.ErrResidentNotFound
	}
	id := uuid.New()
	row := db.GetDuesPaymentByIDRow{
		ID:           pgUUID(id),
		PeriodID:     arg.PeriodID,
		ResidentID:   arg.ResidentID,
		ResidentName: resident.Name,
		Amount:       arg.Amount,
		PaidAt:       arg.PaidAt,
		CreatedAt:    pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}
	m.payments[id.String()] = row
	return db.DuesPayment{
		ID: row.ID, PeriodID: row.PeriodID, ResidentID: row.ResidentID,
		Amount: row.Amount, PaidAt: row.PaidAt, CreatedAt: row.CreatedAt,
	}, nil
}

func (m *memoryStore) UpdatePayment(_ context.Context, arg db.UpdateDuesPaymentParams) (db.DuesPayment, error) {
	if m.updatePaymentErr != nil {
		return db.DuesPayment{}, m.updatePaymentErr
	}
	id, err := uuidutil.ToString(arg.ID)
	if err != nil {
		return db.DuesPayment{}, err
	}
	item, ok := m.payments[id]
	if !ok {
		return db.DuesPayment{}, dues.ErrPaymentNotFound
	}
	periodID := item.PeriodID
	residentID := item.ResidentID
	if arg.PeriodID.Valid {
		periodID = arg.PeriodID
	}
	if arg.ResidentID.Valid {
		residentID = arg.ResidentID
	}
	pStr, _ := uuidutil.ToString(periodID)
	rStr, _ := uuidutil.ToString(residentID)
	for otherID, existing := range m.payments {
		if otherID == id {
			continue
		}
		ep, _ := uuidutil.ToString(existing.PeriodID)
		er, _ := uuidutil.ToString(existing.ResidentID)
		if ep == pStr && er == rStr {
			return db.DuesPayment{}, dues.ErrPaymentAlreadyExists
		}
	}
	item.PeriodID = periodID
	item.ResidentID = residentID
	if arg.Amount != nil {
		item.Amount = *arg.Amount
	}
	if arg.PaidAt.Valid {
		item.PaidAt = arg.PaidAt
	}
	if resident, ok := m.residents[rStr]; ok {
		item.ResidentName = resident.Name
	}
	m.payments[id] = item
	return db.DuesPayment{
		ID: item.ID, PeriodID: item.PeriodID, ResidentID: item.ResidentID,
		Amount: item.Amount, PaidAt: item.PaidAt, CreatedAt: item.CreatedAt,
	}, nil
}

func (m *memoryStore) DeletePayment(_ context.Context, id string) error {
	if _, ok := m.payments[id]; !ok {
		return dues.ErrPaymentNotFound
	}
	delete(m.payments, id)
	return nil
}

func (m *memoryStore) GetPeriodSummary(_ context.Context, periodID string) (db.GetDuesPeriodSummaryRow, error) {
	period, ok := m.periods[periodID]
	if !ok {
		return db.GetDuesPeriodSummaryRow{}, dues.ErrPeriodNotFound
	}
	var residentCount, paidCount, collected int64
	for _, r := range m.residents {
		residentCount++
		rID, _ := uuidutil.ToString(r.ID)
		for _, p := range m.payments {
			pp, _ := uuidutil.ToString(p.PeriodID)
			pr, _ := uuidutil.ToString(p.ResidentID)
			if pp == periodID && pr == rID {
				paidCount++
				collected += p.Amount
				break
			}
		}
	}
	return db.GetDuesPeriodSummaryRow{
		PeriodID:         period.ID,
		Year:             period.Year,
		Month:            period.Month,
		Half:             period.Half,
		ExpectedAmount:   period.Amount,
		ResidentCount:    residentCount,
		PaidCount:        paidCount,
		UnpaidCount:      residentCount - paidCount,
		ExpectedTotal:    residentCount * period.Amount,
		CollectedTotal:   collected,
		OutstandingTotal: residentCount*period.Amount - collected,
	}, nil
}

func (m *memoryStore) ListResidentPaymentStatus(_ context.Context, periodID, search, statusFilter string, limit, offset int32) ([]db.ListDuesResidentPaymentStatusRow, error) {
	items := make([]db.ListDuesResidentPaymentStatusRow, 0)
	for _, r := range m.residents {
		if search != "" && !strings.Contains(strings.ToLower(r.Name), strings.ToLower(search)) {
			continue
		}
		rID, _ := uuidutil.ToString(r.ID)
		row := db.ListDuesResidentPaymentStatusRow{
			ResidentID:   r.ID,
			ResidentName: r.Name,
			Phone:        r.Phone,
			Status:       dues.PaymentStatusUnpaid,
		}
		for _, p := range m.payments {
			pp, _ := uuidutil.ToString(p.PeriodID)
			pr, _ := uuidutil.ToString(p.ResidentID)
			if pp == periodID && pr == rID {
				row.Status = dues.PaymentStatusPaid
				row.PaymentID = p.ID
				amount := p.Amount
				row.Amount = &amount
				row.PaidAt = p.PaidAt
				break
			}
		}
		if statusFilter != "" && row.Status != statusFilter {
			continue
		}
		items = append(items, row)
	}
	return pageSlice(items, limit, offset), nil
}

func (m *memoryStore) CountResidentPaymentStatus(ctx context.Context, periodID, search, statusFilter string) (int64, error) {
	items, _ := m.ListResidentPaymentStatus(ctx, periodID, search, statusFilter, 100000, 0)
	return int64(len(items)), nil
}

type memoryResidents struct {
	byID map[string]db.Resident
}

func (m *memoryResidents) GetByID(_ context.Context, id string) (db.Resident, error) {
	item, ok := m.byID[id]
	if !ok {
		return db.Resident{}, dues.ErrResidentNotFound
	}
	return item, nil
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pageSlice[T any](items []T, limit, offset int32) []T {
	if offset >= int32(len(items)) {
		return []T{}
	}
	items = items[offset:]
	if int32(len(items)) > limit {
		items = items[:limit]
	}
	return items
}

func addResident(store *memoryStore, name, phone string) string {
	id := uuid.New()
	item := db.Resident{
		ID:    pgUUID(id),
		Name:  name,
		Phone: phone,
	}
	store.residents[id.String()] = item
	return id.String()
}

func newService(store *memoryStore) *dues.Service {
	return dues.NewService(store, &memoryResidents{byID: store.residents})
}

func TestCreatePeriodValid(t *testing.T) {
	svc := newService(newMemoryStore())
	item, err := svc.CreatePeriod(context.Background(), dues.CreatePeriodRequest{
		Year: 2026, Month: 9, Half: 1, Amount: 50000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Year != 2026 || item.Month != 9 || item.Half != 1 || item.Amount != 50000 {
		t.Fatalf("unexpected period: %+v", item)
	}
}

func TestCreatePeriodInvalidFields(t *testing.T) {
	svc := newService(newMemoryStore())
	cases := []struct {
		name string
		req  dues.CreatePeriodRequest
	}{
		{"year", dues.CreatePeriodRequest{Year: 1999, Month: 9, Half: 1, Amount: 50000}},
		{"month", dues.CreatePeriodRequest{Year: 2026, Month: 13, Half: 1, Amount: 50000}},
		{"half", dues.CreatePeriodRequest{Year: 2026, Month: 9, Half: 3, Amount: 50000}},
		{"amount", dues.CreatePeriodRequest{Year: 2026, Month: 9, Half: 1, Amount: 0}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreatePeriod(context.Background(), tc.req)
			if !errors.Is(err, dues.ErrInvalidRequest) {
				t.Fatalf("expected ErrInvalidRequest, got %v", err)
			}
		})
	}
}

func TestCreatePeriodDuplicate(t *testing.T) {
	store := newMemoryStore()
	svc := newService(store)
	_, err := svc.CreatePeriod(context.Background(), dues.CreatePeriodRequest{
		Year: 2026, Month: 9, Half: 1, Amount: 50000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.CreatePeriod(context.Background(), dues.CreatePeriodRequest{
		Year: 2026, Month: 9, Half: 1, Amount: 60000,
	})
	if !errors.Is(err, dues.ErrPeriodAlreadyExists) {
		t.Fatalf("expected ErrPeriodAlreadyExists, got %v", err)
	}
}

func TestUpdatePeriodAndDuplicate(t *testing.T) {
	store := newMemoryStore()
	svc := newService(store)
	a, err := svc.CreatePeriod(context.Background(), dues.CreatePeriodRequest{
		Year: 2026, Month: 9, Half: 1, Amount: 50000,
	})
	if err != nil {
		t.Fatalf("create a: %v", err)
	}
	_, err = svc.CreatePeriod(context.Background(), dues.CreatePeriodRequest{
		Year: 2026, Month: 9, Half: 2, Amount: 50000,
	})
	if err != nil {
		t.Fatalf("create b: %v", err)
	}

	half := int32(2)
	amount := dues.Amount(55000)
	_, err = svc.UpdatePeriod(context.Background(), a.ID, dues.UpdatePeriodRequest{Amount: &amount})
	if err != nil {
		t.Fatalf("update amount: %v", err)
	}
	_, err = svc.UpdatePeriod(context.Background(), a.ID, dues.UpdatePeriodRequest{Half: &half})
	if !errors.Is(err, dues.ErrPeriodAlreadyExists) {
		t.Fatalf("expected duplicate, got %v", err)
	}
}

func TestDeletePeriodNotFound(t *testing.T) {
	svc := newService(newMemoryStore())
	err := svc.DeletePeriod(context.Background(), uuid.NewString())
	if !errors.Is(err, dues.ErrPeriodNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestListPeriodsPaginationAndFilter(t *testing.T) {
	store := newMemoryStore()
	svc := newService(store)
	for half := int32(1); half <= 2; half++ {
		_, err := svc.CreatePeriod(context.Background(), dues.CreatePeriodRequest{
			Year: 2026, Month: 9, Half: half, Amount: 50000,
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
	}
	result, err := svc.ListPeriods(context.Background(), 1, 1, "2026", "9", "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.Meta.Total != 2 || len(result.Items) != 1 || result.Meta.TotalPages != 2 {
		t.Fatalf("unexpected pagination: items=%d meta=%+v", len(result.Items), result.Meta)
	}
}

func TestCreatePaymentValidAndRules(t *testing.T) {
	store := newMemoryStore()
	svc := newService(store)
	period, err := svc.CreatePeriod(context.Background(), dues.CreatePeriodRequest{
		Year: 2026, Month: 9, Half: 1, Amount: 50000,
	})
	if err != nil {
		t.Fatalf("period: %v", err)
	}
	residentID := addResident(store, "Budi", "0811")

	payment, err := svc.CreatePayment(context.Background(), dues.CreatePaymentRequest{
		PeriodID: period.ID, ResidentID: residentID, Amount: 50000,
		PaidAt: "2026-09-20T09:00:00+07:00",
	})
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if payment.ResidentName != "Budi" || payment.Amount != 50000 {
		t.Fatalf("unexpected payment: %+v", payment)
	}

	_, err = svc.CreatePayment(context.Background(), dues.CreatePaymentRequest{
		PeriodID: period.ID, ResidentID: residentID, Amount: 50000,
		PaidAt: "2026-09-20T09:00:00+07:00",
	})
	if !errors.Is(err, dues.ErrPaymentAlreadyExists) {
		t.Fatalf("expected duplicate, got %v", err)
	}

	other := addResident(store, "Ani", "0812")
	_, err = svc.CreatePayment(context.Background(), dues.CreatePaymentRequest{
		PeriodID: period.ID, ResidentID: other, Amount: 30000,
		PaidAt: "2026-09-20T09:00:00+07:00",
	})
	if !errors.Is(err, dues.ErrInvalidPaymentAmount) {
		t.Fatalf("expected amount mismatch, got %v", err)
	}

	_, err = svc.CreatePayment(context.Background(), dues.CreatePaymentRequest{
		PeriodID: period.ID, ResidentID: uuid.NewString(), Amount: 50000,
		PaidAt: "2026-09-20T09:00:00+07:00",
	})
	if !errors.Is(err, dues.ErrResidentNotFound) {
		t.Fatalf("expected resident not found, got %v", err)
	}

	_, err = svc.CreatePayment(context.Background(), dues.CreatePaymentRequest{
		PeriodID: uuid.NewString(), ResidentID: other, Amount: 50000,
		PaidAt: "2026-09-20T09:00:00+07:00",
	})
	if !errors.Is(err, dues.ErrPeriodNotFound) {
		t.Fatalf("expected period not found, got %v", err)
	}

	_, err = svc.CreatePayment(context.Background(), dues.CreatePaymentRequest{
		PeriodID: period.ID, ResidentID: other, Amount: 0,
		PaidAt: "2026-09-20T09:00:00+07:00",
	})
	if !errors.Is(err, dues.ErrInvalidRequest) {
		t.Fatalf("expected invalid amount, got %v", err)
	}

	_, err = svc.CreatePayment(context.Background(), dues.CreatePaymentRequest{
		PeriodID: period.ID, ResidentID: other, Amount: 50000,
		PaidAt: "not-a-timestamp",
	})
	if !errors.Is(err, dues.ErrInvalidRequest) {
		t.Fatalf("expected invalid paid_at, got %v", err)
	}
}

func TestUpdatePaymentAndDuplicate(t *testing.T) {
	store := newMemoryStore()
	svc := newService(store)
	period, _ := svc.CreatePeriod(context.Background(), dues.CreatePeriodRequest{
		Year: 2026, Month: 9, Half: 1, Amount: 50000,
	})
	r1 := addResident(store, "Budi", "0811")
	r2 := addResident(store, "Ani", "0812")

	p1, err := svc.CreatePayment(context.Background(), dues.CreatePaymentRequest{
		PeriodID: period.ID, ResidentID: r1, Amount: 50000,
		PaidAt: "2026-09-20T09:00:00+07:00",
	})
	if err != nil {
		t.Fatalf("p1: %v", err)
	}
	_, err = svc.CreatePayment(context.Background(), dues.CreatePaymentRequest{
		PeriodID: period.ID, ResidentID: r2, Amount: 50000,
		PaidAt: "2026-09-20T10:00:00+07:00",
	})
	if err != nil {
		t.Fatalf("p2: %v", err)
	}

	paidAt := "2026-09-21T09:00:00+07:00"
	updated, err := svc.UpdatePayment(context.Background(), p1.ID, dues.UpdatePaymentRequest{PaidAt: &paidAt})
	if err != nil {
		t.Fatalf("update paid_at: %v", err)
	}
	if !strings.HasPrefix(updated.PaidAt, "2026-09-21") {
		t.Fatalf("unexpected paid_at: %s", updated.PaidAt)
	}

	_, err = svc.UpdatePayment(context.Background(), p1.ID, dues.UpdatePaymentRequest{ResidentID: &r2})
	if !errors.Is(err, dues.ErrPaymentAlreadyExists) {
		t.Fatalf("expected duplicate on update, got %v", err)
	}

	badAmount := dues.Amount(30000)
	_, err = svc.UpdatePayment(context.Background(), p1.ID, dues.UpdatePaymentRequest{Amount: &badAmount})
	if !errors.Is(err, dues.ErrInvalidPaymentAmount) {
		t.Fatalf("expected amount mismatch, got %v", err)
	}
}

func TestDeletePayment(t *testing.T) {
	store := newMemoryStore()
	svc := newService(store)
	period, _ := svc.CreatePeriod(context.Background(), dues.CreatePeriodRequest{
		Year: 2026, Month: 9, Half: 1, Amount: 50000,
	})
	r1 := addResident(store, "Budi", "0811")
	payment, _ := svc.CreatePayment(context.Background(), dues.CreatePaymentRequest{
		PeriodID: period.ID, ResidentID: r1, Amount: 50000,
		PaidAt: "2026-09-20T09:00:00+07:00",
	})
	if err := svc.DeletePayment(context.Background(), payment.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err := svc.GetPayment(context.Background(), payment.ID)
	if !errors.Is(err, dues.ErrPaymentNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestListPaymentFilters(t *testing.T) {
	store := newMemoryStore()
	svc := newService(store)
	period, _ := svc.CreatePeriod(context.Background(), dues.CreatePeriodRequest{
		Year: 2026, Month: 9, Half: 1, Amount: 50000,
	})
	r1 := addResident(store, "Budi Santoso", "0811")
	r2 := addResident(store, "Ani Putri", "0812")
	_, _ = svc.CreatePayment(context.Background(), dues.CreatePaymentRequest{
		PeriodID: period.ID, ResidentID: r1, Amount: 50000,
		PaidAt: "2026-09-20T09:00:00+07:00",
	})
	_, _ = svc.CreatePayment(context.Background(), dues.CreatePaymentRequest{
		PeriodID: period.ID, ResidentID: r2, Amount: 50000,
		PaidAt: "2026-09-21T09:00:00+07:00",
	})

	result, err := svc.ListPayments(context.Background(), 1, 20, period.ID, "", "budi", "2026-09-20", "2026-09-20")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.Meta.Total != 1 || result.Items[0].ResidentName != "Budi Santoso" {
		t.Fatalf("unexpected filter result: %+v", result)
	}
}

func TestPeriodSummaryAggregates(t *testing.T) {
	store := newMemoryStore()
	svc := newService(store)
	period, _ := svc.CreatePeriod(context.Background(), dues.CreatePeriodRequest{
		Year: 2026, Month: 9, Half: 1, Amount: 50000,
	})

	summary, err := svc.PeriodSummary(context.Background(), period.ID)
	if err != nil {
		t.Fatalf("zero residents: %v", err)
	}
	if summary.ResidentCount != 0 || summary.ExpectedTotal != 0 || summary.CollectedTotal != 0 {
		t.Fatalf("expected zeros, got %+v", summary)
	}

	r1 := addResident(store, "A", "1")
	r2 := addResident(store, "B", "2")
	summary, err = svc.PeriodSummary(context.Background(), period.ID)
	if err != nil {
		t.Fatalf("all unpaid: %v", err)
	}
	if summary.ResidentCount != 2 || summary.PaidCount != 0 || summary.UnpaidCount != 2 ||
		summary.ExpectedTotal != 100000 || summary.CollectedTotal != 0 || summary.OutstandingTotal != 100000 {
		t.Fatalf("unexpected unpaid summary: %+v", summary)
	}

	_, _ = svc.CreatePayment(context.Background(), dues.CreatePaymentRequest{
		PeriodID: period.ID, ResidentID: r1, Amount: 50000,
		PaidAt: "2026-09-20T09:00:00+07:00",
	})
	summary, err = svc.PeriodSummary(context.Background(), period.ID)
	if err != nil {
		t.Fatalf("partial: %v", err)
	}
	if summary.PaidCount != 1 || summary.UnpaidCount != 1 || summary.CollectedTotal != 50000 || summary.OutstandingTotal != 50000 {
		t.Fatalf("unexpected partial summary: %+v", summary)
	}

	_, _ = svc.CreatePayment(context.Background(), dues.CreatePaymentRequest{
		PeriodID: period.ID, ResidentID: r2, Amount: 50000,
		PaidAt: "2026-09-20T10:00:00+07:00",
	})
	summary, err = svc.PeriodSummary(context.Background(), period.ID)
	if err != nil {
		t.Fatalf("all paid: %v", err)
	}
	if summary.PaidCount != 2 || summary.UnpaidCount != 0 || summary.CollectedTotal != 100000 || summary.OutstandingTotal != 0 {
		t.Fatalf("unexpected all-paid summary: %+v", summary)
	}
}

func TestResidentPaymentStatus(t *testing.T) {
	store := newMemoryStore()
	svc := newService(store)
	period, _ := svc.CreatePeriod(context.Background(), dues.CreatePeriodRequest{
		Year: 2026, Month: 9, Half: 1, Amount: 50000,
	})
	r1 := addResident(store, "Budi", "0811")
	_ = addResident(store, "Ani", "0812")
	_, _ = svc.CreatePayment(context.Background(), dues.CreatePaymentRequest{
		PeriodID: period.ID, ResidentID: r1, Amount: 50000,
		PaidAt: "2026-09-20T09:00:00+07:00",
	})

	all, err := svc.ListResidentPaymentStatus(context.Background(), period.ID, 1, 20, "", "")
	if err != nil {
		t.Fatalf("all: %v", err)
	}
	if all.Meta.Total != 2 {
		t.Fatalf("expected 2, got %d", all.Meta.Total)
	}

	paid, err := svc.ListResidentPaymentStatus(context.Background(), period.ID, 1, 20, "PAID", "")
	if err != nil {
		t.Fatalf("paid: %v", err)
	}
	if paid.Meta.Total != 1 || paid.Items[0].Status != dues.PaymentStatusPaid || paid.Items[0].PaymentID == nil {
		t.Fatalf("unexpected paid: %+v", paid)
	}

	unpaid, err := svc.ListResidentPaymentStatus(context.Background(), period.ID, 1, 20, "UNPAID", "")
	if err != nil {
		t.Fatalf("unpaid: %v", err)
	}
	if unpaid.Meta.Total != 1 || unpaid.Items[0].Status != dues.PaymentStatusUnpaid || unpaid.Items[0].PaymentID != nil {
		t.Fatalf("unexpected unpaid: %+v", unpaid)
	}

	search, err := svc.ListResidentPaymentStatus(context.Background(), period.ID, 1, 20, "", "ani")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if search.Meta.Total != 1 || search.Items[0].ResidentName != "Ani" {
		t.Fatalf("unexpected search: %+v", search)
	}

	page, err := svc.ListResidentPaymentStatus(context.Background(), period.ID, 1, 1, "", "")
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if len(page.Items) != 1 || page.Meta.TotalPages != 2 {
		t.Fatalf("unexpected pagination: %+v", page)
	}
}

func TestAmountRejectsFloatJSON(t *testing.T) {
	var req dues.CreatePeriodRequest
	err := json.Unmarshal([]byte(`{"year":2026,"month":9,"half":1,"amount":50000.5}`), &req)
	if err == nil {
		t.Fatal("expected float amount to fail")
	}
}

func TestDayRangeJakartaSemantics(t *testing.T) {
	fromAt, toExclusive, err := dues.DayRange("2026-09-20", "2026-09-20")
	if err != nil {
		t.Fatalf("day range: %v", err)
	}
	if fromAt == nil || toExclusive == nil {
		t.Fatal("expected both bounds")
	}
	if fromAt.Location().String() != dues.Jakarta.String() {
		t.Fatalf("expected Jakarta location, got %s", fromAt.Location())
	}
	if toExclusive.Sub(*fromAt) != 24*time.Hour {
		t.Fatalf("expected 24h window, got %s", toExclusive.Sub(*fromAt))
	}
}
