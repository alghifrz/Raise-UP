package finance_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/internal/finance"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type memoryStore struct {
	byID map[string]db.CashTransaction
}

func newMemoryStore() *memoryStore {
	return &memoryStore{byID: make(map[string]db.CashTransaction)}
}

func (m *memoryStore) GetByID(_ context.Context, id string) (db.CashTransaction, error) {
	item, ok := m.byID[id]
	if !ok {
		return db.CashTransaction{}, finance.ErrNotFound
	}
	return item, nil
}

func (m *memoryStore) List(_ context.Context, txType *db.CashTransactionType, category, search string, fromAt, toExclusive *time.Time, limit, offset int32) ([]db.CashTransaction, error) {
	items := make([]db.CashTransaction, 0)
	for _, item := range m.byID {
		if txType != nil && item.Type != *txType {
			continue
		}
		if category != "" && item.Category != category {
			continue
		}
		if search != "" {
			needle := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(item.Title), needle) &&
				!strings.Contains(strings.ToLower(item.Note), needle) &&
				!strings.Contains(strings.ToLower(item.Category), needle) {
				continue
			}
		}
		if fromAt != nil && item.CreatedAt.Time.Before(*fromAt) {
			continue
		}
		if toExclusive != nil && !item.CreatedAt.Time.Before(*toExclusive) {
			continue
		}
		items = append(items, item)
	}
	if offset >= int32(len(items)) {
		return []db.CashTransaction{}, nil
	}
	items = items[offset:]
	if int32(len(items)) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (m *memoryStore) Count(ctx context.Context, txType *db.CashTransactionType, category, search string, fromAt, toExclusive *time.Time) (int64, error) {
	items, err := m.List(ctx, txType, category, search, fromAt, toExclusive, 100000, 0)
	if err != nil {
		return 0, err
	}
	return int64(len(items)), nil
}

func (m *memoryStore) Create(_ context.Context, arg db.CreateCashTransactionParams) (db.CashTransaction, error) {
	id := uuid.New()
	now := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	item := db.CashTransaction{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		Type:      arg.Type,
		Title:     arg.Title,
		Amount:    arg.Amount,
		Category:  arg.Category,
		Note:      arg.Note,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.byID[id.String()] = item
	return item, nil
}

func (m *memoryStore) Update(_ context.Context, arg db.UpdateCashTransactionParams) (db.CashTransaction, error) {
	id, err := uuidutil.ToString(arg.ID)
	if err != nil {
		return db.CashTransaction{}, err
	}
	item, ok := m.byID[id]
	if !ok {
		return db.CashTransaction{}, finance.ErrNotFound
	}
	item.Type = arg.Type
	item.Title = arg.Title
	item.Amount = arg.Amount
	item.Category = arg.Category
	item.Note = arg.Note
	item.UpdatedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	m.byID[id] = item
	return item, nil
}

func (m *memoryStore) Delete(_ context.Context, id string) error {
	if _, ok := m.byID[id]; !ok {
		return finance.ErrNotFound
	}
	delete(m.byID, id)
	return nil
}

func (m *memoryStore) Summary(_ context.Context, fromAt, toExclusive *time.Time) (db.GetCashTransactionSummaryRow, error) {
	var income, expense int64
	for _, item := range m.byID {
		if fromAt != nil && item.CreatedAt.Time.Before(*fromAt) {
			continue
		}
		if toExclusive != nil && !item.CreatedAt.Time.Before(*toExclusive) {
			continue
		}
		switch item.Type {
		case db.CashTransactionTypeINCOME:
			income += item.Amount
		case db.CashTransactionTypeEXPENSE:
			expense += item.Amount
		}
	}
	return db.GetCashTransactionSummaryRow{TotalIncome: income, TotalExpense: expense}, nil
}

func TestCreateValidIncomeExpense(t *testing.T) {
	service := finance.NewService(newMemoryStore())
	income, err := service.Create(context.Background(), finance.CreateRequest{
		Type: "INCOME", Title: "  Iuran  ", Amount: 500000, Category: " IURAN ", Note: " note ",
	})
	if err != nil {
		t.Fatalf("Create income error = %v", err)
	}
	if income.Type != db.CashTransactionTypeINCOME || income.Amount != 500000 || income.Title != "Iuran" {
		t.Fatalf("unexpected income: %+v", income)
	}

	expense, err := service.Create(context.Background(), finance.CreateRequest{
		Type: "EXPENSE", Title: "Ops", Amount: 150000, Category: "OPERASIONAL",
	})
	if err != nil || expense.Type != db.CashTransactionTypeEXPENSE {
		t.Fatalf("Create expense error = %v result=%+v", err, expense)
	}
}

func TestCreateValidation(t *testing.T) {
	service := finance.NewService(newMemoryStore())
	cases := []finance.CreateRequest{
		{Title: "t", Amount: 1, Category: "c"},
		{Type: "OTHER", Title: "t", Amount: 1, Category: "c"},
		{Type: "INCOME", Amount: 1, Category: "c"},
		{Type: "INCOME", Title: "t", Amount: 1},
		{Type: "INCOME", Title: "t", Amount: 0, Category: "c"},
		{Type: "INCOME", Title: "t", Amount: -1, Category: "c"},
	}
	for _, req := range cases {
		if _, err := service.Create(context.Background(), req); !errors.Is(err, finance.ErrInvalidRequest) {
			t.Fatalf("Create(%+v) error = %v", req, err)
		}
	}
}

func TestAmountRejectsDecimalJSON(t *testing.T) {
	var amount finance.Amount
	if err := json.Unmarshal([]byte(`500000.50`), &amount); err == nil {
		t.Fatal("expected decimal amount to fail")
	}
	if err := json.Unmarshal([]byte(`500000`), &amount); err != nil || amount != 500000 {
		t.Fatalf("integer amount failed: %v value=%d", err, amount)
	}
}

func TestUpdatePartial(t *testing.T) {
	service := finance.NewService(newMemoryStore())
	created, _ := service.Create(context.Background(), finance.CreateRequest{
		Type: "INCOME", Title: "Old", Amount: 1000, Category: "IURAN",
	})
	title := "New"
	amount := finance.Amount(2000)
	updated, err := service.Update(context.Background(), created.ID, finance.UpdateRequest{
		Title: &title, Amount: &amount,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Title != "New" || updated.Amount != 2000 || updated.Category != "IURAN" {
		t.Fatalf("unexpected update: %+v", updated)
	}

	badAmount := finance.Amount(0)
	if _, err := service.Update(context.Background(), created.ID, finance.UpdateRequest{Amount: &badAmount}); !errors.Is(err, finance.ErrInvalidRequest) {
		t.Fatalf("expected invalid amount, got %v", err)
	}
	empty := "  "
	if _, err := service.Update(context.Background(), created.ID, finance.UpdateRequest{Title: &empty}); !errors.Is(err, finance.ErrInvalidRequest) {
		t.Fatalf("expected empty title error, got %v", err)
	}
	badType := "OTHER"
	if _, err := service.Update(context.Background(), created.ID, finance.UpdateRequest{Type: &badType}); !errors.Is(err, finance.ErrInvalidRequest) {
		t.Fatalf("expected invalid type, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	service := finance.NewService(newMemoryStore())
	created, _ := service.Create(context.Background(), finance.CreateRequest{
		Type: "EXPENSE", Title: "x", Amount: 1, Category: "c",
	})
	if err := service.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if err := service.Delete(context.Background(), created.ID); !errors.Is(err, finance.ErrNotFound) {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestListFiltersAndDates(t *testing.T) {
	store := newMemoryStore()
	service := finance.NewService(store)

	income, _ := service.Create(context.Background(), finance.CreateRequest{
		Type: "INCOME", Title: "Iuran September", Amount: 500000, Category: "IURAN", Note: "blok A",
	})
	_, _ = service.Create(context.Background(), finance.CreateRequest{
		Type: "EXPENSE", Title: "ATK", Amount: 100000, Category: "OPERASIONAL",
	})

	// Force created_at into known Jakarta days for date filter tests.
	item := store.byID[income.ID]
	item.CreatedAt = pgtype.Timestamptz{Time: time.Date(2026, 9, 15, 10, 0, 0, 0, finance.Jakarta).UTC(), Valid: true}
	store.byID[income.ID] = item

	defaults, err := service.List(context.Background(), 0, 0, "", "", "", "", "")
	if err != nil || defaults.Meta.Page != 1 || defaults.Meta.PageSize != 20 || defaults.Meta.Total != 2 {
		t.Fatalf("defaults = %+v err=%v", defaults, err)
	}
	limited, _ := service.List(context.Background(), 1, 1000, "", "", "", "", "")
	if limited.Meta.PageSize != 100 {
		t.Fatalf("page_size = %d", limited.Meta.PageSize)
	}
	byType, _ := service.List(context.Background(), 1, 20, "INCOME", "", "", "", "")
	if byType.Meta.Total != 1 {
		t.Fatalf("type total = %d", byType.Meta.Total)
	}
	byCategory, _ := service.List(context.Background(), 1, 20, "", "OPERASIONAL", "", "", "")
	if byCategory.Meta.Total != 1 {
		t.Fatalf("category total = %d", byCategory.Meta.Total)
	}
	bySearch, _ := service.List(context.Background(), 1, 20, "", "", "blok", "", "")
	if bySearch.Meta.Total != 1 {
		t.Fatalf("search total = %d", bySearch.Meta.Total)
	}
	byFrom, _ := service.List(context.Background(), 1, 20, "", "", "", "2026-09-15", "")
	if byFrom.Meta.Total < 1 {
		t.Fatalf("from filter total = %d", byFrom.Meta.Total)
	}
	byRange, _ := service.List(context.Background(), 1, 20, "", "", "", "2026-09-01", "2026-09-30")
	if byRange.Meta.Total < 1 {
		t.Fatalf("range total = %d", byRange.Meta.Total)
	}
	if _, err := service.List(context.Background(), 1, 20, "", "", "", "bad", ""); !errors.Is(err, finance.ErrInvalidRequest) {
		t.Fatalf("invalid from error = %v", err)
	}
	if _, err := service.List(context.Background(), 1, 20, "", "", "", "2026-09-30", "2026-09-01"); !errors.Is(err, finance.ErrInvalidRequest) {
		t.Fatalf("from>to error = %v", err)
	}
}

func TestSummary(t *testing.T) {
	store := newMemoryStore()
	service := finance.NewService(store)

	empty, err := service.Summary(context.Background(), "", "")
	if err != nil || empty.TotalIncome != 0 || empty.TotalExpense != 0 || empty.Balance != 0 {
		t.Fatalf("empty summary = %+v err=%v", empty, err)
	}

	inc, _ := service.Create(context.Background(), finance.CreateRequest{Type: "INCOME", Title: "a", Amount: 500000, Category: "IURAN"})
	exp, _ := service.Create(context.Background(), finance.CreateRequest{Type: "EXPENSE", Title: "b", Amount: 200000, Category: "OPS"})
	_ = inc
	_ = exp

	sum, err := service.Summary(context.Background(), "", "")
	if err != nil || sum.TotalIncome != 500000 || sum.TotalExpense != 200000 || sum.Balance != 300000 {
		t.Fatalf("summary = %+v err=%v", sum, err)
	}

	item := store.byID[inc.ID]
	item.CreatedAt = pgtype.Timestamptz{Time: time.Date(2026, 9, 10, 8, 0, 0, 0, finance.Jakarta).UTC(), Valid: true}
	store.byID[inc.ID] = item
	item = store.byID[exp.ID]
	item.CreatedAt = pgtype.Timestamptz{Time: time.Date(2026, 10, 5, 8, 0, 0, 0, finance.Jakarta).UTC(), Valid: true}
	store.byID[exp.ID] = item

	sept, _ := service.Summary(context.Background(), "2026-09-01", "2026-09-30")
	if sept.TotalIncome != 500000 || sept.TotalExpense != 0 || sept.Balance != 500000 {
		t.Fatalf("sept summary = %+v", sept)
	}
	if _, err := service.Summary(context.Background(), "2026-10-01", "2026-09-01"); !errors.Is(err, finance.ErrInvalidRequest) {
		t.Fatalf("from>to summary error = %v", err)
	}
}

func TestDayRangeJakartaSemantics(t *testing.T) {
	fromAt, toExclusive, err := finance.DayRange("2026-09-01", "2026-09-30")
	if err != nil {
		t.Fatalf("DayRange() error = %v", err)
	}
	if fromAt.Location().String() != finance.Jakarta.String() {
		t.Fatalf("from location = %s", fromAt.Location())
	}
	if fromAt.Hour() != 0 || fromAt.Day() != 1 {
		t.Fatalf("unexpected from = %v", fromAt)
	}
	if toExclusive.Day() != 1 || toExclusive.Month() != time.October {
		t.Fatalf("unexpected toExclusive = %v", toExclusive)
	}
}
