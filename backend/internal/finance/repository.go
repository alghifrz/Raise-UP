package finance

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides cash transaction persistence via sqlc.
type Repository struct {
	q *db.Queries
}

// NewRepository creates a finance repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

// GetByID returns a transaction by UUID string.
func (r *Repository) GetByID(ctx context.Context, id string) (db.CashTransaction, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.CashTransaction{}, ErrInvalidRequest
	}
	item, err := r.q.GetCashTransactionByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.CashTransaction{}, ErrNotFound
		}
		return db.CashTransaction{}, fmt.Errorf("get cash transaction: %w", err)
	}
	return item, nil
}

// List returns a filtered page of transactions.
func (r *Repository) List(
	ctx context.Context,
	txType *db.CashTransactionType,
	category, search string,
	fromAt, toExclusive *time.Time,
	limit, offset int32,
) ([]db.CashTransaction, error) {
	items, err := r.q.ListCashTransactionsFiltered(ctx, db.ListCashTransactionsFilteredParams{
		Type:        toNullType(txType),
		Category:    category,
		Search:      search,
		FromAt:      toTimestamptz(fromAt),
		ToExclusive: toTimestamptz(toExclusive),
		LimitCount:  limit,
		OffsetCount: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list cash transactions: %w", err)
	}
	return items, nil
}

// Count returns the total matching transactions.
func (r *Repository) Count(
	ctx context.Context,
	txType *db.CashTransactionType,
	category, search string,
	fromAt, toExclusive *time.Time,
) (int64, error) {
	total, err := r.q.CountCashTransactionsFiltered(ctx, db.CountCashTransactionsFilteredParams{
		Type:        toNullType(txType),
		Category:    category,
		Search:      search,
		FromAt:      toTimestamptz(fromAt),
		ToExclusive: toTimestamptz(toExclusive),
	})
	if err != nil {
		return 0, fmt.Errorf("count cash transactions: %w", err)
	}
	return total, nil
}

// Create inserts a cash transaction.
func (r *Repository) Create(ctx context.Context, arg db.CreateCashTransactionParams) (db.CashTransaction, error) {
	item, err := r.q.CreateCashTransaction(ctx, arg)
	if err != nil {
		return db.CashTransaction{}, fmt.Errorf("create cash transaction: %w", err)
	}
	return item, nil
}

// Update replaces transaction fields.
func (r *Repository) Update(ctx context.Context, arg db.UpdateCashTransactionParams) (db.CashTransaction, error) {
	item, err := r.q.UpdateCashTransaction(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.CashTransaction{}, ErrNotFound
		}
		return db.CashTransaction{}, fmt.Errorf("update cash transaction: %w", err)
	}
	return item, nil
}

// Delete removes a transaction by id.
func (r *Repository) Delete(ctx context.Context, id string) error {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return ErrInvalidRequest
	}
	if err := r.q.DeleteCashTransaction(ctx, pgID); err != nil {
		return fmt.Errorf("delete cash transaction: %w", err)
	}
	return nil
}

// Summary returns aggregated income/expense totals.
func (r *Repository) Summary(ctx context.Context, fromAt, toExclusive *time.Time) (db.GetCashTransactionSummaryRow, error) {
	row, err := r.q.GetCashTransactionSummary(ctx, db.GetCashTransactionSummaryParams{
		FromAt:      toTimestamptz(fromAt),
		ToExclusive: toTimestamptz(toExclusive),
	})
	if err != nil {
		return db.GetCashTransactionSummaryRow{}, fmt.Errorf("get cash summary: %w", err)
	}
	return row, nil
}

func toNullType(txType *db.CashTransactionType) db.NullCashTransactionType {
	if txType == nil {
		return db.NullCashTransactionType{}
	}
	return db.NullCashTransactionType{CashTransactionType: *txType, Valid: true}
}

func toTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}
