package dues

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides dues persistence via sqlc.
type Repository struct {
	q *db.Queries
}

// NewRepository creates a dues repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

// GetPeriodByID returns a dues period by UUID string.
func (r *Repository) GetPeriodByID(ctx context.Context, id string) (db.DuesPeriod, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.DuesPeriod{}, ErrInvalidRequest
	}
	item, err := r.q.GetDuesPeriodByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.DuesPeriod{}, ErrPeriodNotFound
		}
		return db.DuesPeriod{}, fmt.Errorf("get dues period: %w", err)
	}
	return item, nil
}

// ListPeriods returns a filtered page of dues periods.
func (r *Repository) ListPeriods(ctx context.Context, year, month, half *int32, limit, offset int32) ([]db.DuesPeriod, error) {
	items, err := r.q.ListDuesPeriodsFiltered(ctx, db.ListDuesPeriodsFilteredParams{
		Year:       year,
		Month:      month,
		Half:       half,
		PageLimit:  limit,
		PageOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list dues periods: %w", err)
	}
	return items, nil
}

// CountPeriods returns the total matching dues periods.
func (r *Repository) CountPeriods(ctx context.Context, year, month, half *int32) (int64, error) {
	total, err := r.q.CountDuesPeriodsFiltered(ctx, db.CountDuesPeriodsFilteredParams{
		Year:  year,
		Month: month,
		Half:  half,
	})
	if err != nil {
		return 0, fmt.Errorf("count dues periods: %w", err)
	}
	return total, nil
}

// CreatePeriod inserts a dues period.
func (r *Repository) CreatePeriod(ctx context.Context, year, month, half int32, amount int64) (db.DuesPeriod, error) {
	item, err := r.q.CreateDuesPeriod(ctx, db.CreateDuesPeriodParams{
		Year:   year,
		Month:  month,
		Half:   half,
		Amount: amount,
	})
	if err != nil {
		if mapped := mapPeriodWriteError(err); mapped != nil {
			return db.DuesPeriod{}, mapped
		}
		return db.DuesPeriod{}, fmt.Errorf("create dues period: %w", err)
	}
	return item, nil
}

// UpdatePeriod updates a dues period.
func (r *Repository) UpdatePeriod(ctx context.Context, arg db.UpdateDuesPeriodParams) (db.DuesPeriod, error) {
	item, err := r.q.UpdateDuesPeriod(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.DuesPeriod{}, ErrPeriodNotFound
		}
		if mapped := mapPeriodWriteError(err); mapped != nil {
			return db.DuesPeriod{}, mapped
		}
		return db.DuesPeriod{}, fmt.Errorf("update dues period: %w", err)
	}
	return item, nil
}

// DeletePeriod deletes a dues period (cascades payments).
func (r *Repository) DeletePeriod(ctx context.Context, id string) error {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return ErrInvalidRequest
	}
	n, err := r.q.DeleteDuesPeriod(ctx, pgID)
	if err != nil {
		return fmt.Errorf("delete dues period: %w", err)
	}
	if n == 0 {
		return ErrPeriodNotFound
	}
	return nil
}

// GetPaymentByID returns a dues payment with resident name.
func (r *Repository) GetPaymentByID(ctx context.Context, id string) (db.GetDuesPaymentByIDRow, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.GetDuesPaymentByIDRow{}, ErrInvalidRequest
	}
	item, err := r.q.GetDuesPaymentByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetDuesPaymentByIDRow{}, ErrPaymentNotFound
		}
		return db.GetDuesPaymentByIDRow{}, fmt.Errorf("get dues payment: %w", err)
	}
	return item, nil
}

// ListPayments returns a filtered page of dues payments.
func (r *Repository) ListPayments(
	ctx context.Context,
	periodID, residentID *string,
	search string,
	fromAt, toExclusive *time.Time,
	limit, offset int32,
) ([]db.ListDuesPaymentsFilteredRow, error) {
	periodUUID, err := optionalUUID(periodID)
	if err != nil {
		return nil, ErrInvalidRequest
	}
	residentUUID, err := optionalUUID(residentID)
	if err != nil {
		return nil, ErrInvalidRequest
	}

	items, err := r.q.ListDuesPaymentsFiltered(ctx, db.ListDuesPaymentsFilteredParams{
		PeriodID:   periodUUID,
		ResidentID: residentUUID,
		Search:     search,
		FromAt:     toTimestamptz(fromAt),
		ToAt:       toTimestamptz(toExclusive),
		PageLimit:  limit,
		PageOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list dues payments: %w", err)
	}
	return items, nil
}

// CountPayments returns the total matching dues payments.
func (r *Repository) CountPayments(
	ctx context.Context,
	periodID, residentID *string,
	search string,
	fromAt, toExclusive *time.Time,
) (int64, error) {
	periodUUID, err := optionalUUID(periodID)
	if err != nil {
		return 0, ErrInvalidRequest
	}
	residentUUID, err := optionalUUID(residentID)
	if err != nil {
		return 0, ErrInvalidRequest
	}

	total, err := r.q.CountDuesPaymentsFiltered(ctx, db.CountDuesPaymentsFilteredParams{
		PeriodID:   periodUUID,
		ResidentID: residentUUID,
		Search:     search,
		FromAt:     toTimestamptz(fromAt),
		ToAt:       toTimestamptz(toExclusive),
	})
	if err != nil {
		return 0, fmt.Errorf("count dues payments: %w", err)
	}
	return total, nil
}

// CreatePayment inserts a dues payment.
func (r *Repository) CreatePayment(ctx context.Context, arg db.CreateDuesPaymentParams) (db.DuesPayment, error) {
	item, err := r.q.CreateDuesPayment(ctx, arg)
	if err != nil {
		if mapped := mapPaymentWriteError(err); mapped != nil {
			return db.DuesPayment{}, mapped
		}
		return db.DuesPayment{}, fmt.Errorf("create dues payment: %w", err)
	}
	return item, nil
}

// UpdatePayment updates a dues payment.
func (r *Repository) UpdatePayment(ctx context.Context, arg db.UpdateDuesPaymentParams) (db.DuesPayment, error) {
	item, err := r.q.UpdateDuesPayment(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.DuesPayment{}, ErrPaymentNotFound
		}
		if mapped := mapPaymentWriteError(err); mapped != nil {
			return db.DuesPayment{}, mapped
		}
		return db.DuesPayment{}, fmt.Errorf("update dues payment: %w", err)
	}
	return item, nil
}

// DeletePayment deletes a dues payment.
func (r *Repository) DeletePayment(ctx context.Context, id string) error {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return ErrInvalidRequest
	}
	n, err := r.q.DeleteDuesPayment(ctx, pgID)
	if err != nil {
		return fmt.Errorf("delete dues payment: %w", err)
	}
	if n == 0 {
		return ErrPaymentNotFound
	}
	return nil
}

// GetPeriodSummary returns aggregated period collection stats.
func (r *Repository) GetPeriodSummary(ctx context.Context, periodID string) (db.GetDuesPeriodSummaryRow, error) {
	pgID, err := uuidutil.FromString(periodID)
	if err != nil {
		return db.GetDuesPeriodSummaryRow{}, ErrInvalidRequest
	}
	item, err := r.q.GetDuesPeriodSummary(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetDuesPeriodSummaryRow{}, ErrPeriodNotFound
		}
		return db.GetDuesPeriodSummaryRow{}, fmt.Errorf("get dues period summary: %w", err)
	}
	return item, nil
}

// ListResidentPaymentStatus returns derived PAID/UNPAID rows for a period.
func (r *Repository) ListResidentPaymentStatus(
	ctx context.Context,
	periodID, search, statusFilter string,
	limit, offset int32,
) ([]db.ListDuesResidentPaymentStatusRow, error) {
	pgID, err := uuidutil.FromString(periodID)
	if err != nil {
		return nil, ErrInvalidRequest
	}
	items, err := r.q.ListDuesResidentPaymentStatus(ctx, db.ListDuesResidentPaymentStatusParams{
		PeriodID:     pgID,
		Search:       search,
		StatusFilter: statusFilter,
		PageLimit:    limit,
		PageOffset:   offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list dues resident payment status: %w", err)
	}
	return items, nil
}

// CountResidentPaymentStatus returns the total matching status rows.
func (r *Repository) CountResidentPaymentStatus(ctx context.Context, periodID, search, statusFilter string) (int64, error) {
	pgID, err := uuidutil.FromString(periodID)
	if err != nil {
		return 0, ErrInvalidRequest
	}
	total, err := r.q.CountDuesResidentPaymentStatus(ctx, db.CountDuesResidentPaymentStatusParams{
		PeriodID:     pgID,
		Search:       search,
		StatusFilter: statusFilter,
	})
	if err != nil {
		return 0, fmt.Errorf("count dues resident payment status: %w", err)
	}
	return total, nil
}

func mapPeriodWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	if pgErr.Code == "23505" {
		return ErrPeriodAlreadyExists
	}
	return nil
}

func mapPaymentWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	switch pgErr.Code {
	case "23505":
		return ErrPaymentAlreadyExists
	case "23503":
		switch pgErr.ConstraintName {
		case "dues_payments_period_id_fkey":
			return ErrPeriodNotFound
		case "dues_payments_resident_id_fkey":
			return ErrResidentNotFound
		}
	}
	return nil
}

func optionalUUID(id *string) (pgtype.UUID, error) {
	if id == nil || *id == "" {
		return pgtype.UUID{}, nil
	}
	return uuidutil.FromString(*id)
}

func toTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}
