package dues

import (
	"context"
	"fmt"
	"strconv"
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
	minYear         = 2000
	maxYear         = 2100
)

// Store is the persistence interface used by Service.
type Store interface {
	GetPeriodByID(ctx context.Context, id string) (db.DuesPeriod, error)
	ListPeriods(ctx context.Context, year, month, half *int32, limit, offset int32) ([]db.DuesPeriod, error)
	CountPeriods(ctx context.Context, year, month, half *int32) (int64, error)
	CreatePeriod(ctx context.Context, year, month, half int32, amount int64) (db.DuesPeriod, error)
	UpdatePeriod(ctx context.Context, arg db.UpdateDuesPeriodParams) (db.DuesPeriod, error)
	DeletePeriod(ctx context.Context, id string) error

	GetPaymentByID(ctx context.Context, id string) (db.GetDuesPaymentByIDRow, error)
	ListPayments(ctx context.Context, periodID, residentID *string, search string, fromAt, toExclusive *time.Time, limit, offset int32) ([]db.ListDuesPaymentsFilteredRow, error)
	CountPayments(ctx context.Context, periodID, residentID *string, search string, fromAt, toExclusive *time.Time) (int64, error)
	CreatePayment(ctx context.Context, arg db.CreateDuesPaymentParams) (db.DuesPayment, error)
	UpdatePayment(ctx context.Context, arg db.UpdateDuesPaymentParams) (db.DuesPayment, error)
	DeletePayment(ctx context.Context, id string) error

	GetPeriodSummary(ctx context.Context, periodID string) (db.GetDuesPeriodSummaryRow, error)
	ListResidentPaymentStatus(ctx context.Context, periodID, search, statusFilter string, limit, offset int32) ([]db.ListDuesResidentPaymentStatusRow, error)
	CountResidentPaymentStatus(ctx context.Context, periodID, search, statusFilter string) (int64, error)
}

// ReminderMessenger sends and records a SYSTEM WhatsApp message.
type ReminderMessenger interface {
	Enabled() bool
	SendText(ctx context.Context, phone, body string) (messageID string, err error)
}

// Service implements dues use cases.
type Service struct {
	store     Store
	residents ResidentReader
	messenger ReminderMessenger
}

// NewService creates a dues service.
func NewService(store Store, residents ResidentReader) *Service {
	return &Service{store: store, residents: residents}
}

// SetMessenger enables manual unpaid-payment WhatsApp reminders.
func (s *Service) SetMessenger(messenger ReminderMessenger) {
	s.messenger = messenger
}

// ListPeriods returns a paginated filtered period list.
func (s *Service) ListPeriods(ctx context.Context, page, pageSize int, year, month, half string) (*PeriodListResult, error) {
	params, err := normalizePeriodListParams(page, pageSize, year, month, half)
	if err != nil {
		return nil, err
	}

	offset := int32((params.page - 1) * params.pageSize)
	limit := int32(params.pageSize)

	items, err := s.store.ListPeriods(ctx, params.year, params.month, params.half, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.store.CountPeriods(ctx, params.year, params.month, params.half)
	if err != nil {
		return nil, err
	}

	out := make([]Period, 0, len(items))
	for _, item := range items {
		mapped, err := toAPIPeriod(item)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}

	return &PeriodListResult{
		Items: out,
		Meta:  listMeta(params.page, params.pageSize, total),
	}, nil
}

// GetPeriod returns a period by id.
func (s *Service) GetPeriod(ctx context.Context, id string) (*Period, error) {
	if _, err := uuidutil.FromString(id); err != nil {
		return nil, fmt.Errorf("%w: invalid period id", ErrInvalidRequest)
	}
	item, err := s.store.GetPeriodByID(ctx, id)
	if err != nil {
		return nil, err
	}
	mapped, err := toAPIPeriod(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// CreatePeriod validates and inserts a dues period.
func (s *Service) CreatePeriod(ctx context.Context, req CreatePeriodRequest) (*Period, error) {
	if err := validatePeriodFields(req.Year, req.Month, req.Half, int64(req.Amount)); err != nil {
		return nil, err
	}

	item, err := s.store.CreatePeriod(ctx, req.Year, req.Month, req.Half, int64(req.Amount))
	if err != nil {
		return nil, err
	}
	mapped, err := toAPIPeriod(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// UpdatePeriod partially updates a dues period.
func (s *Service) UpdatePeriod(ctx context.Context, id string, req UpdatePeriodRequest) (*Period, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid period id", ErrInvalidRequest)
	}

	current, err := s.store.GetPeriodByID(ctx, id)
	if err != nil {
		return nil, err
	}

	year := current.Year
	month := current.Month
	half := current.Half
	amount := current.Amount

	arg := db.UpdateDuesPeriodParams{ID: pgID}

	if req.Year != nil {
		year = *req.Year
		arg.Year = req.Year
	}
	if req.Month != nil {
		month = *req.Month
		arg.Month = req.Month
	}
	if req.Half != nil {
		half = *req.Half
		arg.Half = req.Half
	}
	if req.Amount != nil {
		amount = int64(*req.Amount)
		arg.Amount = &amount
	}

	if err := validatePeriodFields(year, month, half, amount); err != nil {
		return nil, err
	}

	item, err := s.store.UpdatePeriod(ctx, arg)
	if err != nil {
		return nil, err
	}
	mapped, err := toAPIPeriod(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// DeletePeriod deletes a period and cascaded payments.
func (s *Service) DeletePeriod(ctx context.Context, id string) error {
	if _, err := uuidutil.FromString(id); err != nil {
		return fmt.Errorf("%w: invalid period id", ErrInvalidRequest)
	}
	return s.store.DeletePeriod(ctx, id)
}

// ListPayments returns a paginated filtered payment list.
func (s *Service) ListPayments(
	ctx context.Context,
	page, pageSize int,
	periodID, residentID, search, from, to string,
) (*PaymentListResult, error) {
	params, err := normalizePaymentListParams(page, pageSize, periodID, residentID, search, from, to)
	if err != nil {
		return nil, err
	}

	offset := int32((params.page - 1) * params.pageSize)
	limit := int32(params.pageSize)

	items, err := s.store.ListPayments(ctx, params.periodID, params.residentID, params.search, params.fromAt, params.toExclusive, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.store.CountPayments(ctx, params.periodID, params.residentID, params.search, params.fromAt, params.toExclusive)
	if err != nil {
		return nil, err
	}

	out := make([]Payment, 0, len(items))
	for _, item := range items {
		mapped, err := toAPIPaymentFromList(item)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}

	return &PaymentListResult{
		Items: out,
		Meta:  listMeta(params.page, params.pageSize, total),
	}, nil
}

// GetPayment returns a payment by id.
func (s *Service) GetPayment(ctx context.Context, id string) (*Payment, error) {
	if _, err := uuidutil.FromString(id); err != nil {
		return nil, fmt.Errorf("%w: invalid payment id", ErrInvalidRequest)
	}
	item, err := s.store.GetPaymentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	mapped, err := toAPIPayment(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// CreatePayment validates and inserts a dues payment.
func (s *Service) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*Payment, error) {
	periodID := strings.TrimSpace(req.PeriodID)
	residentID := strings.TrimSpace(req.ResidentID)
	if periodID == "" {
		return nil, fmt.Errorf("%w: period_id is required", ErrInvalidRequest)
	}
	if residentID == "" {
		return nil, fmt.Errorf("%w: resident_id is required", ErrInvalidRequest)
	}
	periodUUID, err := uuidutil.FromString(periodID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid period_id", ErrInvalidRequest)
	}
	residentUUID, err := uuidutil.FromString(residentID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid resident_id", ErrInvalidRequest)
	}
	if int64(req.Amount) <= 0 {
		return nil, fmt.Errorf("%w: amount must be greater than 0", ErrInvalidRequest)
	}
	paidAt, err := parsePaidAt(req.PaidAt)
	if err != nil {
		return nil, err
	}

	period, err := s.store.GetPeriodByID(ctx, periodID)
	if err != nil {
		return nil, err
	}
	if _, err := s.residents.GetByID(ctx, residentID); err != nil {
		return nil, err
	}
	if int64(req.Amount) != period.Amount {
		return nil, ErrInvalidPaymentAmount
	}

	created, err := s.store.CreatePayment(ctx, db.CreateDuesPaymentParams{
		PeriodID:   periodUUID,
		ResidentID: residentUUID,
		Amount:     int64(req.Amount),
		PaidAt:     pgtype.Timestamptz{Time: paidAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	id, err := uuidutil.ToString(created.ID)
	if err != nil {
		return nil, fmt.Errorf("map payment id: %w", err)
	}
	return s.GetPayment(ctx, id)
}

// UpdatePayment partially updates a dues payment.
func (s *Service) UpdatePayment(ctx context.Context, id string, req UpdatePaymentRequest) (*Payment, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid payment id", ErrInvalidRequest)
	}

	current, err := s.store.GetPaymentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	periodIDStr, err := uuidutil.ToString(current.PeriodID)
	if err != nil {
		return nil, fmt.Errorf("map period id: %w", err)
	}
	residentIDStr, err := uuidutil.ToString(current.ResidentID)
	if err != nil {
		return nil, fmt.Errorf("map resident id: %w", err)
	}
	amount := current.Amount

	arg := db.UpdateDuesPaymentParams{ID: pgID}

	if req.PeriodID != nil {
		periodIDStr = strings.TrimSpace(*req.PeriodID)
		if periodIDStr == "" {
			return nil, fmt.Errorf("%w: period_id cannot be empty", ErrInvalidRequest)
		}
		periodUUID, err := uuidutil.FromString(periodIDStr)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid period_id", ErrInvalidRequest)
		}
		arg.PeriodID = periodUUID
	}
	if req.ResidentID != nil {
		residentIDStr = strings.TrimSpace(*req.ResidentID)
		if residentIDStr == "" {
			return nil, fmt.Errorf("%w: resident_id cannot be empty", ErrInvalidRequest)
		}
		residentUUID, err := uuidutil.FromString(residentIDStr)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid resident_id", ErrInvalidRequest)
		}
		arg.ResidentID = residentUUID
	}
	if req.Amount != nil {
		if int64(*req.Amount) <= 0 {
			return nil, fmt.Errorf("%w: amount must be greater than 0", ErrInvalidRequest)
		}
		amount = int64(*req.Amount)
		arg.Amount = &amount
	}
	if req.PaidAt != nil {
		paidAt, err := parsePaidAt(*req.PaidAt)
		if err != nil {
			return nil, err
		}
		arg.PaidAt = pgtype.Timestamptz{Time: paidAt.UTC(), Valid: true}
	}

	period, err := s.store.GetPeriodByID(ctx, periodIDStr)
	if err != nil {
		return nil, err
	}
	if req.ResidentID != nil {
		if _, err := s.residents.GetByID(ctx, residentIDStr); err != nil {
			return nil, err
		}
	}

	// Whenever period_id or amount changes, payment amount must equal the target period amount.
	if req.PeriodID != nil || req.Amount != nil {
		if amount != period.Amount {
			return nil, ErrInvalidPaymentAmount
		}
	}

	if _, err := s.store.UpdatePayment(ctx, arg); err != nil {
		return nil, err
	}
	return s.GetPayment(ctx, id)
}

// DeletePayment deletes a dues payment.
func (s *Service) DeletePayment(ctx context.Context, id string) error {
	if _, err := uuidutil.FromString(id); err != nil {
		return fmt.Errorf("%w: invalid payment id", ErrInvalidRequest)
	}
	return s.store.DeletePayment(ctx, id)
}

// PeriodSummary returns aggregated collection stats for a period.
func (s *Service) PeriodSummary(ctx context.Context, periodID string) (*PeriodSummary, error) {
	if _, err := uuidutil.FromString(periodID); err != nil {
		return nil, fmt.Errorf("%w: invalid period id", ErrInvalidRequest)
	}
	item, err := s.store.GetPeriodSummary(ctx, periodID)
	if err != nil {
		return nil, err
	}
	id, err := uuidutil.ToString(item.PeriodID)
	if err != nil {
		return nil, fmt.Errorf("map period id: %w", err)
	}
	return &PeriodSummary{
		PeriodID:         id,
		Year:             item.Year,
		Month:            item.Month,
		Half:             item.Half,
		ExpectedAmount:   item.ExpectedAmount,
		ResidentCount:    item.ResidentCount,
		PaidCount:        item.PaidCount,
		UnpaidCount:      item.UnpaidCount,
		ExpectedTotal:    item.ExpectedTotal,
		CollectedTotal:   item.CollectedTotal,
		OutstandingTotal: item.OutstandingTotal,
	}, nil
}

// SendUnpaidReminders sends one WhatsApp reminder to every unpaid resident.
func (s *Service) SendUnpaidReminders(ctx context.Context, periodID string) (*ReminderResult, error) {
	if _, err := uuidutil.FromString(periodID); err != nil {
		return nil, fmt.Errorf("%w: invalid period id", ErrInvalidRequest)
	}
	period, err := s.store.GetPeriodByID(ctx, periodID)
	if err != nil {
		return nil, err
	}
	if s.messenger == nil || !s.messenger.Enabled() {
		return nil, ErrReminderUnavailable
	}

	const batchSize = int32(100)
	result := &ReminderResult{}
	for offset := int32(0); ; offset += batchSize {
		items, err := s.store.ListResidentPaymentStatus(
			ctx,
			periodID,
			"",
			PaymentStatusUnpaid,
			batchSize,
			offset,
		)
		if err != nil {
			return nil, err
		}

		for _, item := range items {
			result.Total++
			if strings.TrimSpace(item.Phone) == "" {
				result.Failed++
				continue
			}
			message := formatUnpaidReminder(period, item.ResidentName)
			if _, err := s.messenger.SendText(ctx, item.Phone, message); err != nil {
				result.Failed++
				continue
			}
			result.Sent++
		}

		if len(items) < int(batchSize) {
			break
		}
	}
	return result, nil
}

func formatUnpaidReminder(period db.DuesPeriod, residentName string) string {
	months := [...]string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	month := strconv.Itoa(int(period.Month))
	if period.Month >= 1 && period.Month <= 12 {
		month = months[period.Month-1]
	}
	return fmt.Sprintf(
		"Halo %s,\n\n⏰ *Pengingat Iuran RT*\n"+
			"Periode: %s %d (Tahap %d)\n"+
			"Nominal: %s\n\n"+
			"Pembayaran Anda belum tercatat. Mohon segera melakukan pembayaran. Terima kasih.",
		strings.TrimSpace(residentName),
		month,
		period.Year,
		period.Half,
		formatReminderAmount(period.Amount),
	)
}

func formatReminderAmount(amount int64) string {
	raw := strconv.FormatInt(amount, 10)
	var b strings.Builder
	for i, r := range raw {
		if i > 0 && (len(raw)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}
	return "Rp" + b.String()
}

// ListResidentPaymentStatus returns derived PAID/UNPAID rows for residents in a period.
func (s *Service) ListResidentPaymentStatus(
	ctx context.Context,
	periodID string,
	page, pageSize int,
	status, search string,
) (*StatusListResult, error) {
	if _, err := uuidutil.FromString(periodID); err != nil {
		return nil, fmt.Errorf("%w: invalid period id", ErrInvalidRequest)
	}
	if _, err := s.store.GetPeriodByID(ctx, periodID); err != nil {
		return nil, err
	}

	params, err := normalizeStatusListParams(page, pageSize, status, search)
	if err != nil {
		return nil, err
	}

	offset := int32((params.page - 1) * params.pageSize)
	limit := int32(params.pageSize)

	items, err := s.store.ListResidentPaymentStatus(ctx, periodID, params.search, params.status, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.store.CountResidentPaymentStatus(ctx, periodID, params.search, params.status)
	if err != nil {
		return nil, err
	}

	out := make([]ResidentPaymentStatus, 0, len(items))
	for _, item := range items {
		mapped, err := toAPIStatus(item)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}

	return &StatusListResult{
		Items: out,
		Meta:  listMeta(params.page, params.pageSize, total),
	}, nil
}

type periodListParams struct {
	page, pageSize    int
	year, month, half *int32
}

type paymentListParams struct {
	page, pageSize       int
	periodID, residentID *string
	search               string
	fromAt, toExclusive  *time.Time
}

type statusListParams struct {
	page, pageSize int
	status, search string
}

func normalizePeriodListParams(page, pageSize int, year, month, half string) (*periodListParams, error) {
	page, pageSize, err := normalizePage(page, pageSize)
	if err != nil {
		return nil, err
	}

	var yearPtr, monthPtr, halfPtr *int32
	if year != "" {
		v, err := parseInt32Field(year, "year")
		if err != nil {
			return nil, err
		}
		if err := validateYear(v); err != nil {
			return nil, err
		}
		yearPtr = &v
	}
	if month != "" {
		v, err := parseInt32Field(month, "month")
		if err != nil {
			return nil, err
		}
		if err := validateMonth(v); err != nil {
			return nil, err
		}
		monthPtr = &v
	}
	if half != "" {
		v, err := parseInt32Field(half, "half")
		if err != nil {
			return nil, err
		}
		if err := validateHalf(v); err != nil {
			return nil, err
		}
		halfPtr = &v
	}

	return &periodListParams{
		page:     page,
		pageSize: pageSize,
		year:     yearPtr,
		month:    monthPtr,
		half:     halfPtr,
	}, nil
}

func normalizePaymentListParams(page, pageSize int, periodID, residentID, search, from, to string) (*paymentListParams, error) {
	page, pageSize, err := normalizePage(page, pageSize)
	if err != nil {
		return nil, err
	}

	var periodPtr, residentPtr *string
	if periodID = strings.TrimSpace(periodID); periodID != "" {
		if _, err := uuidutil.FromString(periodID); err != nil {
			return nil, fmt.Errorf("%w: invalid period_id", ErrInvalidRequest)
		}
		periodPtr = &periodID
	}
	if residentID = strings.TrimSpace(residentID); residentID != "" {
		if _, err := uuidutil.FromString(residentID); err != nil {
			return nil, fmt.Errorf("%w: invalid resident_id", ErrInvalidRequest)
		}
		residentPtr = &residentID
	}

	fromAt, toExclusive, err := DayRange(from, to)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidRequest, err.Error())
	}

	return &paymentListParams{
		page:        page,
		pageSize:    pageSize,
		periodID:    periodPtr,
		residentID:  residentPtr,
		search:      strings.TrimSpace(search),
		fromAt:      fromAt,
		toExclusive: toExclusive,
	}, nil
}

func normalizeStatusListParams(page, pageSize int, status, search string) (*statusListParams, error) {
	page, pageSize, err := normalizePage(page, pageSize)
	if err != nil {
		return nil, err
	}
	status = strings.ToUpper(strings.TrimSpace(status))
	switch status {
	case "", PaymentStatusPaid, PaymentStatusUnpaid:
	default:
		return nil, fmt.Errorf("%w: status must be PAID or UNPAID", ErrInvalidRequest)
	}
	return &statusListParams{
		page:     page,
		pageSize: pageSize,
		status:   status,
		search:   strings.TrimSpace(search),
	}, nil
}

func normalizePage(page, pageSize int) (int, int, error) {
	if page < 1 {
		return 0, 0, fmt.Errorf("%w: page must be >= 1", ErrInvalidRequest)
	}
	if pageSize < 1 {
		return 0, 0, fmt.Errorf("%w: page_size must be >= 1", ErrInvalidRequest)
	}
	if pageSize > maxPageSize {
		return 0, 0, fmt.Errorf("%w: page_size must be <= %d", ErrInvalidRequest, maxPageSize)
	}
	return page, pageSize, nil
}

func validatePeriodFields(year, month, half int32, amount int64) error {
	if err := validateYear(year); err != nil {
		return err
	}
	if err := validateMonth(month); err != nil {
		return err
	}
	if err := validateHalf(half); err != nil {
		return err
	}
	if amount <= 0 {
		return fmt.Errorf("%w: amount must be greater than 0", ErrInvalidRequest)
	}
	return nil
}

func validateYear(year int32) error {
	if year < minYear || year > maxYear {
		return fmt.Errorf("%w: year must be between %d and %d", ErrInvalidRequest, minYear, maxYear)
	}
	return nil
}

func validateMonth(month int32) error {
	if month < 1 || month > 12 {
		return fmt.Errorf("%w: month must be between 1 and 12", ErrInvalidRequest)
	}
	return nil
}

func validateHalf(half int32) error {
	if half != 1 && half != 2 {
		return fmt.Errorf("%w: half must be 1 or 2", ErrInvalidRequest)
	}
	return nil
}

func parseInt32Field(raw, field string) (int32, error) {
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid %s", ErrInvalidRequest, field)
	}
	return int32(v), nil
}

func parsePaidAt(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, fmt.Errorf("%w: paid_at is required", ErrInvalidRequest)
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: paid_at must be RFC3339", ErrInvalidRequest)
	}
	return t, nil
}

func listMeta(page, pageSize int, total int64) ListMeta {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return ListMeta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

func toAPIPeriod(item db.DuesPeriod) (Period, error) {
	id, err := uuidutil.ToString(item.ID)
	if err != nil {
		return Period{}, fmt.Errorf("map period id: %w", err)
	}
	if !item.CreatedAt.Valid {
		return Period{}, fmt.Errorf("map period created_at: invalid timestamp")
	}
	return Period{
		ID:        id,
		Year:      item.Year,
		Month:     item.Month,
		Half:      item.Half,
		Amount:    item.Amount,
		CreatedAt: item.CreatedAt.Time.UTC().Format(time.RFC3339),
	}, nil
}

func toAPIPayment(item db.GetDuesPaymentByIDRow) (Payment, error) {
	id, err := uuidutil.ToString(item.ID)
	if err != nil {
		return Payment{}, fmt.Errorf("map payment id: %w", err)
	}
	periodID, err := uuidutil.ToString(item.PeriodID)
	if err != nil {
		return Payment{}, fmt.Errorf("map period id: %w", err)
	}
	residentID, err := uuidutil.ToString(item.ResidentID)
	if err != nil {
		return Payment{}, fmt.Errorf("map resident id: %w", err)
	}
	if !item.PaidAt.Valid || !item.CreatedAt.Valid {
		return Payment{}, fmt.Errorf("map payment timestamps: invalid timestamp")
	}
	return Payment{
		ID:           id,
		PeriodID:     periodID,
		ResidentID:   residentID,
		ResidentName: item.ResidentName,
		Amount:       item.Amount,
		PaidAt:       item.PaidAt.Time.UTC().Format(time.RFC3339),
		CreatedAt:    item.CreatedAt.Time.UTC().Format(time.RFC3339),
	}, nil
}

func toAPIPaymentFromList(item db.ListDuesPaymentsFilteredRow) (Payment, error) {
	return toAPIPayment(db.GetDuesPaymentByIDRow{
		ID:           item.ID,
		PeriodID:     item.PeriodID,
		ResidentID:   item.ResidentID,
		ResidentName: item.ResidentName,
		Amount:       item.Amount,
		PaidAt:       item.PaidAt,
		CreatedAt:    item.CreatedAt,
	})
}

func toAPIStatus(item db.ListDuesResidentPaymentStatusRow) (ResidentPaymentStatus, error) {
	residentID, err := uuidutil.ToString(item.ResidentID)
	if err != nil {
		return ResidentPaymentStatus{}, fmt.Errorf("map resident id: %w", err)
	}

	out := ResidentPaymentStatus{
		ResidentID:   residentID,
		ResidentName: item.ResidentName,
		Phone:        item.Phone,
		Status:       item.Status,
	}

	if item.Status == PaymentStatusPaid && item.PaymentID.Valid {
		paymentID, err := uuidutil.ToString(item.PaymentID)
		if err != nil {
			return ResidentPaymentStatus{}, fmt.Errorf("map payment id: %w", err)
		}
		out.PaymentID = &paymentID
		out.Amount = item.Amount
		if item.PaidAt.Valid {
			paidAt := item.PaidAt.Time.UTC().Format(time.RFC3339)
			out.PaidAt = &paidAt
		}
	}

	return out, nil
}
