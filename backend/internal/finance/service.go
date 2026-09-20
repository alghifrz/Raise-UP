package finance

import (
	"context"
	"fmt"
	"strings"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// Store is the persistence interface used by Service.
type Store interface {
	GetByID(ctx context.Context, id string) (db.CashTransaction, error)
	List(ctx context.Context, txType *db.CashTransactionType, category, search string, fromAt, toExclusive *time.Time, limit, offset int32) ([]db.CashTransaction, error)
	Count(ctx context.Context, txType *db.CashTransactionType, category, search string, fromAt, toExclusive *time.Time) (int64, error)
	Create(ctx context.Context, arg db.CreateCashTransactionParams) (db.CashTransaction, error)
	Update(ctx context.Context, arg db.UpdateCashTransactionParams) (db.CashTransaction, error)
	Delete(ctx context.Context, id string) error
	Summary(ctx context.Context, fromAt, toExclusive *time.Time) (db.GetCashTransactionSummaryRow, error)
}

// Service implements finance use cases.
type Service struct {
	store Store
}

// NewService creates a finance service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// List returns a paginated filtered transaction list.
func (s *Service) List(ctx context.Context, page, pageSize int, txType, category, search, from, to string) (*ListResult, error) {
	params, err := normalizeListParams(page, pageSize, txType, category, search, from, to)
	if err != nil {
		return nil, err
	}

	offset := int32((params.page - 1) * params.pageSize)
	limit := int32(params.pageSize)

	items, err := s.store.List(ctx, params.txType, params.category, params.search, params.fromAt, params.toExclusive, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.store.Count(ctx, params.txType, params.category, params.search, params.fromAt, params.toExclusive)
	if err != nil {
		return nil, err
	}

	out := make([]Transaction, 0, len(items))
	for _, item := range items {
		mapped, err := toAPITransaction(item)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(params.pageSize) - 1) / int64(params.pageSize))
	}

	return &ListResult{
		Items: out,
		Meta: ListMeta{
			Page:       params.page,
			PageSize:   params.pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// Get returns a transaction by id.
func (s *Service) Get(ctx context.Context, id string) (*Transaction, error) {
	if _, err := uuidutil.FromString(id); err != nil {
		return nil, fmt.Errorf("%w: invalid transaction id", ErrInvalidRequest)
	}
	item, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	mapped, err := toAPITransaction(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Create validates and inserts a transaction.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Transaction, error) {
	txType, title, amount, category, note, err := validateCreate(req)
	if err != nil {
		return nil, err
	}

	item, err := s.store.Create(ctx, db.CreateCashTransactionParams{
		Type:     txType,
		Title:    title,
		Amount:   amount,
		Category: category,
		Note:     note,
	})
	if err != nil {
		return nil, err
	}

	mapped, err := toAPITransaction(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Update applies a partial update.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Transaction, error) {
	if _, err := uuidutil.FromString(id); err != nil {
		return nil, fmt.Errorf("%w: invalid transaction id", ErrInvalidRequest)
	}

	existing, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	txType := existing.Type
	title := existing.Title
	amount := existing.Amount
	category := existing.Category
	note := existing.Note

	if req.Type != nil {
		parsed, err := parseType(strings.TrimSpace(*req.Type))
		if err != nil {
			return nil, err
		}
		txType = parsed
	}
	if req.Title != nil {
		title = strings.TrimSpace(*req.Title)
		if title == "" {
			return nil, fmt.Errorf("%w: title is required", ErrInvalidRequest)
		}
	}
	if req.Amount != nil {
		if int64(*req.Amount) <= 0 {
			return nil, fmt.Errorf("%w: amount must be greater than zero", ErrInvalidRequest)
		}
		amount = int64(*req.Amount)
	}
	if req.Category != nil {
		category = strings.TrimSpace(*req.Category)
		if category == "" {
			return nil, fmt.Errorf("%w: category is required", ErrInvalidRequest)
		}
	}
	if req.Note != nil {
		note = strings.TrimSpace(*req.Note)
	}

	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid transaction id", ErrInvalidRequest)
	}

	item, err := s.store.Update(ctx, db.UpdateCashTransactionParams{
		ID:       pgID,
		Type:     txType,
		Title:    title,
		Amount:   amount,
		Category: category,
		Note:     note,
	})
	if err != nil {
		return nil, err
	}

	mapped, err := toAPITransaction(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Delete permanently removes a transaction.
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := uuidutil.FromString(id); err != nil {
		return fmt.Errorf("%w: invalid transaction id", ErrInvalidRequest)
	}
	if _, err := s.store.GetByID(ctx, id); err != nil {
		return err
	}
	return s.store.Delete(ctx, id)
}

// Summary returns aggregated finance totals for an optional date range.
func (s *Service) Summary(ctx context.Context, from, to string) (*Summary, error) {
	fromAt, toExclusive, err := DayRange(strings.TrimSpace(from), strings.TrimSpace(to))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidRequest, err.Error())
	}

	row, err := s.store.Summary(ctx, fromAt, toExclusive)
	if err != nil {
		return nil, err
	}

	return &Summary{
		TotalIncome:  row.TotalIncome,
		TotalExpense: row.TotalExpense,
		Balance:      row.TotalIncome - row.TotalExpense,
	}, nil
}

type listParams struct {
	page        int
	pageSize    int
	txType      *db.CashTransactionType
	category    string
	search      string
	fromAt      *time.Time
	toExclusive *time.Time
}

func normalizeListParams(page, pageSize int, txType, category, search, from, to string) (listParams, error) {
	if page <= 0 {
		page = defaultPage
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	params := listParams{
		page:     page,
		pageSize: pageSize,
		category: strings.TrimSpace(category),
		search:   strings.TrimSpace(search),
	}

	if txType = strings.TrimSpace(txType); txType != "" {
		parsed, err := parseType(txType)
		if err != nil {
			return listParams{}, err
		}
		params.txType = &parsed
	}

	fromAt, toExclusive, err := DayRange(strings.TrimSpace(from), strings.TrimSpace(to))
	if err != nil {
		return listParams{}, fmt.Errorf("%w: %s", ErrInvalidRequest, err.Error())
	}
	params.fromAt = fromAt
	params.toExclusive = toExclusive

	return params, nil
}

func validateCreate(req CreateRequest) (db.CashTransactionType, string, int64, string, string, error) {
	title := strings.TrimSpace(req.Title)
	category := strings.TrimSpace(req.Category)
	note := strings.TrimSpace(req.Note)
	typeRaw := strings.TrimSpace(req.Type)

	if typeRaw == "" {
		return "", "", 0, "", "", fmt.Errorf("%w: type is required", ErrInvalidRequest)
	}
	txType, err := parseType(typeRaw)
	if err != nil {
		return "", "", 0, "", "", err
	}
	if title == "" {
		return "", "", 0, "", "", fmt.Errorf("%w: title is required", ErrInvalidRequest)
	}
	if category == "" {
		return "", "", 0, "", "", fmt.Errorf("%w: category is required", ErrInvalidRequest)
	}
	if int64(req.Amount) <= 0 {
		return "", "", 0, "", "", fmt.Errorf("%w: amount must be greater than zero", ErrInvalidRequest)
	}

	return txType, title, int64(req.Amount), category, note, nil
}

func parseType(value string) (db.CashTransactionType, error) {
	switch db.CashTransactionType(value) {
	case db.CashTransactionTypeINCOME, db.CashTransactionTypeEXPENSE:
		return db.CashTransactionType(value), nil
	default:
		return "", fmt.Errorf("%w: type must be INCOME or EXPENSE", ErrInvalidRequest)
	}
}

func toAPITransaction(item db.CashTransaction) (Transaction, error) {
	id, err := uuidutil.ToString(item.ID)
	if err != nil {
		return Transaction{}, err
	}
	createdAt, err := formatTimestamptz(item.CreatedAt)
	if err != nil {
		return Transaction{}, err
	}
	updatedAt, err := formatTimestamptz(item.UpdatedAt)
	if err != nil {
		return Transaction{}, err
	}
	return Transaction{
		ID:        id,
		Type:      item.Type,
		Title:     item.Title,
		Amount:    item.Amount,
		Category:  item.Category,
		Note:      item.Note,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func formatTimestamptz(ts pgtype.Timestamptz) (string, error) {
	if !ts.Valid {
		return "", fmt.Errorf("invalid timestamp")
	}
	return ts.Time.UTC().Format(time.RFC3339), nil
}
