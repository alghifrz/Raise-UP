package resident

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
	GetByID(ctx context.Context, id string) (db.Resident, error)
	List(ctx context.Context, search string, gender *db.Gender, limit, offset int32) ([]db.Resident, error)
	Count(ctx context.Context, search string, gender *db.Gender) (int64, error)
	Create(ctx context.Context, name, phone string, gender db.Gender) (db.Resident, error)
	Update(ctx context.Context, id string, name, phone string, gender db.Gender) (db.Resident, error)
	Delete(ctx context.Context, id string) error
}

// Service implements resident use cases.
type Service struct {
	store Store
}

// NewService creates a resident service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// List returns a paginated, optionally filtered resident list.
func (s *Service) List(ctx context.Context, page, pageSize int, search, gender string) (*ListResult, error) {
	params, err := normalizeListParams(page, pageSize, search, gender)
	if err != nil {
		return nil, err
	}

	offset := int32((params.Page - 1) * params.PageSize)
	limit := int32(params.PageSize)

	items, err := s.store.List(ctx, params.Search, params.Gender, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.store.Count(ctx, params.Search, params.Gender)
	if err != nil {
		return nil, err
	}

	residents := make([]Resident, 0, len(items))
	for _, item := range items {
		mapped, err := toAPIResident(item)
		if err != nil {
			return nil, err
		}
		residents = append(residents, mapped)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(params.PageSize) - 1) / int64(params.PageSize))
	}

	return &ListResult{
		Items: residents,
		Meta: ListMeta{
			Page:       params.Page,
			PageSize:   params.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// Get returns a resident by id.
func (s *Service) Get(ctx context.Context, id string) (*Resident, error) {
	if _, err := uuidutil.FromString(id); err != nil {
		return nil, fmt.Errorf("%w: invalid resident id", ErrInvalidRequest)
	}

	item, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	mapped, err := toAPIResident(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Create validates and creates a resident.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Resident, error) {
	name, phone, gender, err := validateCreate(req)
	if err != nil {
		return nil, err
	}

	item, err := s.store.Create(ctx, name, phone, gender)
	if err != nil {
		return nil, err
	}

	mapped, err := toAPIResident(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Update applies a partial update to a resident.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Resident, error) {
	if _, err := uuidutil.FromString(id); err != nil {
		return nil, fmt.Errorf("%w: invalid resident id", ErrInvalidRequest)
	}

	existing, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	name := existing.Name
	phone := existing.Phone
	gender := existing.Gender

	if req.Name != nil {
		name = strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: name is required", ErrInvalidRequest)
		}
	}
	if req.Phone != nil {
		phone = strings.TrimSpace(*req.Phone)
		if phone == "" {
			return nil, fmt.Errorf("%w: phone is required", ErrInvalidRequest)
		}
	}
	if req.Gender != nil {
		parsed, err := parseGender(strings.TrimSpace(*req.Gender))
		if err != nil {
			return nil, err
		}
		gender = parsed
	}

	item, err := s.store.Update(ctx, id, name, phone, gender)
	if err != nil {
		return nil, err
	}

	mapped, err := toAPIResident(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Delete removes a resident when allowed by FK constraints.
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := uuidutil.FromString(id); err != nil {
		return fmt.Errorf("%w: invalid resident id", ErrInvalidRequest)
	}

	if _, err := s.store.GetByID(ctx, id); err != nil {
		return err
	}

	return s.store.Delete(ctx, id)
}

func normalizeListParams(page, pageSize int, search, gender string) (ListParams, error) {
	if page <= 0 {
		page = defaultPage
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	params := ListParams{
		Page:     page,
		PageSize: pageSize,
		Search:   strings.TrimSpace(search),
	}

	gender = strings.TrimSpace(gender)
	if gender != "" {
		parsed, err := parseGender(gender)
		if err != nil {
			return ListParams{}, err
		}
		params.Gender = &parsed
	}

	return params, nil
}

func validateCreate(req CreateRequest) (string, string, db.Gender, error) {
	name := strings.TrimSpace(req.Name)
	phone := strings.TrimSpace(req.Phone)
	genderRaw := strings.TrimSpace(req.Gender)

	if name == "" {
		return "", "", "", fmt.Errorf("%w: name is required", ErrInvalidRequest)
	}
	if phone == "" {
		return "", "", "", fmt.Errorf("%w: phone is required", ErrInvalidRequest)
	}
	if genderRaw == "" {
		return "", "", "", fmt.Errorf("%w: gender is required", ErrInvalidRequest)
	}

	gender, err := parseGender(genderRaw)
	if err != nil {
		return "", "", "", err
	}

	return name, phone, gender, nil
}

func parseGender(value string) (db.Gender, error) {
	switch db.Gender(value) {
	case db.GenderLAKILAKI, db.GenderPEREMPUAN:
		return db.Gender(value), nil
	default:
		return "", fmt.Errorf("%w: gender must be LAKI_LAKI or PEREMPUAN", ErrInvalidRequest)
	}
}

func toAPIResident(item db.Resident) (Resident, error) {
	id, err := uuidutil.ToString(item.ID)
	if err != nil {
		return Resident{}, err
	}

	createdAt, err := formatTimestamptz(item.CreatedAt)
	if err != nil {
		return Resident{}, err
	}
	updatedAt, err := formatTimestamptz(item.UpdatedAt)
	if err != nil {
		return Resident{}, err
	}

	return Resident{
		ID:        id,
		Name:      item.Name,
		Phone:     item.Phone,
		Gender:    item.Gender,
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
