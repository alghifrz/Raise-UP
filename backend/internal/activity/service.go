package activity

import (
	"context"
	"fmt"
	"regexp"
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
)

var reminderTimePattern = regexp.MustCompile(`^\d{2}:\d{2}$`)

// Store is the persistence interface used by Service.
type Store interface {
	GetByID(ctx context.Context, id string) (db.Activity, error)
	List(ctx context.Context, search string, fromAt, toExclusive *time.Time, limit, offset int32) ([]db.Activity, error)
	Count(ctx context.Context, search string, fromAt, toExclusive *time.Time) (int64, error)
	Create(ctx context.Context, arg db.CreateActivityParams) (db.Activity, error)
	Update(ctx context.Context, arg db.UpdateActivityParams) (db.Activity, error)
	Delete(ctx context.Context, id string) error
}

// Service implements activity use cases.
type Service struct {
	store Store
}

// NewService creates an activity service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// List returns a paginated filtered activity list.
func (s *Service) List(ctx context.Context, page, pageSize int, search, from, to string) (*ListResult, error) {
	params, err := normalizeListParams(page, pageSize, search, from, to)
	if err != nil {
		return nil, err
	}

	offset := int32((params.page - 1) * params.pageSize)
	limit := int32(params.pageSize)

	items, err := s.store.List(ctx, params.search, params.fromAt, params.toExclusive, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.store.Count(ctx, params.search, params.fromAt, params.toExclusive)
	if err != nil {
		return nil, err
	}

	out := make([]Activity, 0, len(items))
	for _, item := range items {
		mapped, err := toAPIActivity(item)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}

	return &ListResult{
		Items: out,
		Meta:  listMeta(params.page, params.pageSize, total),
	}, nil
}

// Get returns an activity by id.
func (s *Service) Get(ctx context.Context, id string) (*Activity, error) {
	if _, err := uuidutil.FromString(id); err != nil {
		return nil, fmt.Errorf("%w: invalid activity id", ErrInvalidRequest)
	}
	item, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	mapped, err := toAPIActivity(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Create validates and inserts an activity.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Activity, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidRequest)
	}
	description := strings.TrimSpace(req.Description)

	date, err := parseActivityDate(req.Date)
	if err != nil {
		return nil, err
	}

	daysBefore := int32(0)
	if req.ReminderDaysBefore != nil {
		if *req.ReminderDaysBefore < 0 {
			return nil, fmt.Errorf("%w: reminder_days_before must be >= 0", ErrInvalidRequest)
		}
		daysBefore = *req.ReminderDaysBefore
	}

	reminderTime, err := parseOptionalReminderTime(req.ReminderTime)
	if err != nil {
		return nil, err
	}
	reminderMessage := strings.TrimSpace(req.ReminderMessage)

	item, err := s.store.Create(ctx, db.CreateActivityParams{
		Name:               name,
		Description:        description,
		Date:               pgtype.Timestamptz{Time: date.UTC(), Valid: true},
		ReminderDaysBefore: daysBefore,
		ReminderTime:       reminderTime,
		ReminderMessage:    reminderMessage,
	})
	if err != nil {
		return nil, err
	}

	mapped, err := toAPIActivity(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Update partially updates an activity.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Activity, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid activity id", ErrInvalidRequest)
	}

	if _, err := s.store.GetByID(ctx, id); err != nil {
		return nil, err
	}

	arg := db.UpdateActivityParams{ID: pgID}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidRequest)
		}
		arg.Name = &name
	}
	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		arg.Description = &description
	}
	if req.Date != nil {
		date, err := parseActivityDate(*req.Date)
		if err != nil {
			return nil, err
		}
		arg.Date = pgtype.Timestamptz{Time: date.UTC(), Valid: true}
	}
	if req.ReminderDaysBefore != nil {
		if *req.ReminderDaysBefore < 0 {
			return nil, fmt.Errorf("%w: reminder_days_before must be >= 0", ErrInvalidRequest)
		}
		arg.ReminderDaysBefore = req.ReminderDaysBefore
	}
	if req.ReminderTime != nil {
		reminderTime, err := parseOptionalReminderTime(*req.ReminderTime)
		if err != nil {
			return nil, err
		}
		if !reminderTime.Valid {
			return nil, fmt.Errorf("%w: reminder_time must use HH:MM 24-hour format", ErrInvalidRequest)
		}
		arg.ReminderTime = reminderTime
	}
	if req.ReminderMessage != nil {
		message := strings.TrimSpace(*req.ReminderMessage)
		arg.ReminderMessage = &message
	}

	arg.ReminderConfigChanged = req.Date != nil ||
		req.ReminderDaysBefore != nil ||
		req.ReminderTime != nil ||
		req.ReminderMessage != nil

	item, err := s.store.Update(ctx, arg)
	if err != nil {
		return nil, err
	}
	mapped, err := toAPIActivity(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Delete removes an activity.
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := uuidutil.FromString(id); err != nil {
		return fmt.Errorf("%w: invalid activity id", ErrInvalidRequest)
	}
	return s.store.Delete(ctx, id)
}

type listParams struct {
	page, pageSize      int
	search              string
	fromAt, toExclusive *time.Time
}

func normalizeListParams(page, pageSize int, search, from, to string) (*listParams, error) {
	if page < 1 {
		return nil, fmt.Errorf("%w: page must be >= 1", ErrInvalidRequest)
	}
	if pageSize < 1 {
		return nil, fmt.Errorf("%w: page_size must be >= 1", ErrInvalidRequest)
	}
	if pageSize > maxPageSize {
		return nil, fmt.Errorf("%w: page_size must be <= %d", ErrInvalidRequest, maxPageSize)
	}

	fromAt, toExclusive, err := DayRange(from, to)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidRequest, err.Error())
	}

	return &listParams{
		page:        page,
		pageSize:    pageSize,
		search:      strings.TrimSpace(search),
		fromAt:      fromAt,
		toExclusive: toExclusive,
	}, nil
}

func parseActivityDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, fmt.Errorf("%w: date is required", ErrInvalidRequest)
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: date must be RFC3339", ErrInvalidRequest)
	}
	return t, nil
}

func parseOptionalReminderTime(raw string) (pgtype.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pgtype.Time{}, nil
	}
	if !reminderTimePattern.MatchString(raw) {
		return pgtype.Time{}, fmt.Errorf("%w: reminder_time must use HH:MM 24-hour format", ErrInvalidRequest)
	}
	hour, err := strconv.Atoi(raw[0:2])
	if err != nil {
		return pgtype.Time{}, fmt.Errorf("%w: reminder_time must use HH:MM 24-hour format", ErrInvalidRequest)
	}
	minute, err := strconv.Atoi(raw[3:5])
	if err != nil {
		return pgtype.Time{}, fmt.Errorf("%w: reminder_time must use HH:MM 24-hour format", ErrInvalidRequest)
	}
	if hour > 23 || minute > 59 {
		return pgtype.Time{}, fmt.Errorf("%w: reminder_time must use HH:MM 24-hour format", ErrInvalidRequest)
	}
	return pgtype.Time{
		Microseconds: int64(hour*3600+minute*60) * 1_000_000,
		Valid:        true,
	}, nil
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

func toAPIActivity(item db.Activity) (Activity, error) {
	id, err := uuidutil.ToString(item.ID)
	if err != nil {
		return Activity{}, fmt.Errorf("map activity id: %w", err)
	}
	if !item.Date.Valid || !item.CreatedAt.Valid || !item.UpdatedAt.Valid {
		return Activity{}, fmt.Errorf("map activity timestamps: invalid timestamp")
	}

	out := Activity{
		ID:                      id,
		Name:                    item.Name,
		Description:             item.Description,
		Date:                    item.Date.Time.UTC().Format(time.RFC3339),
		ReminderDaysBefore:      item.ReminderDaysBefore,
		ReminderTime:            formatReminderTime(item.ReminderTime),
		ReminderMessage:         item.ReminderMessage,
		ReminderScheduleVersion: item.ReminderScheduleVersion,
		ReminderN8nExecutionID:  item.ReminderN8nExecutionID,
		ReminderScheduledAt:     formatOptionalTimestamptz(item.ReminderScheduledAt),
		ReminderSentAt:          formatOptionalTimestamptz(item.ReminderSentAt),
		CreatedAt:               item.CreatedAt.Time.UTC().Format(time.RFC3339),
		UpdatedAt:               item.UpdatedAt.Time.UTC().Format(time.RFC3339),
	}
	return out, nil
}

func formatReminderTime(t pgtype.Time) *string {
	if !t.Valid {
		return nil
	}
	totalSeconds := t.Microseconds / 1_000_000
	hour := totalSeconds / 3600
	minute := (totalSeconds % 3600) / 60
	formatted := fmt.Sprintf("%02d:%02d", hour, minute)
	return &formatted
}

func formatOptionalTimestamptz(t pgtype.Timestamptz) *string {
	if !t.Valid {
		return nil
	}
	formatted := t.Time.UTC().Format(time.RFC3339)
	return &formatted
}
