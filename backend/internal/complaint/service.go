package complaint

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// Store is the persistence interface used by Service.
type Store interface {
	GetByID(ctx context.Context, id string) (db.Complaint, error)
	List(ctx context.Context, search string, status *db.ComplaintStatus, urgency *db.ComplaintUrgency, category string, limit, offset int32) ([]db.Complaint, error)
	Count(ctx context.Context, search string, status *db.ComplaintStatus, urgency *db.ComplaintUrgency, category string) (int64, error)
	Create(ctx context.Context, arg db.CreateComplaintParams) (db.Complaint, error)
	Update(ctx context.Context, arg db.UpdateComplaintParams) (db.Complaint, error)
	UpdateStatus(ctx context.Context, id string, status db.ComplaintStatus) (db.Complaint, error)
	Delete(ctx context.Context, id string) error
}

// ResidentReader loads residents for snapshot resolution.
type ResidentReader interface {
	GetByID(ctx context.Context, id string) (db.Resident, error)
}

// Service implements complaint use cases.
type Service struct {
	store     Store
	residents ResidentReader
	now       func() time.Time
	generate  func(time.Time) (string, error)
}

// NewService creates a complaint service.
func NewService(store Store, residents ResidentReader) *Service {
	return &Service{
		store:     store,
		residents: residents,
		now:       func() time.Time { return time.Now().UTC() },
		generate:  GenerateRef,
	}
}

// List returns a paginated filtered complaint list.
func (s *Service) List(ctx context.Context, page, pageSize int, search, status, urgency, category string) (*ListResult, error) {
	params, err := normalizeListParams(page, pageSize, search, status, urgency, category)
	if err != nil {
		return nil, err
	}

	offset := int32((params.page - 1) * params.pageSize)
	limit := int32(params.pageSize)

	items, err := s.store.List(ctx, params.search, params.status, params.urgency, params.category, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.store.Count(ctx, params.search, params.status, params.urgency, params.category)
	if err != nil {
		return nil, err
	}

	out := make([]Complaint, 0, len(items))
	for _, item := range items {
		mapped, err := toAPIComplaint(item)
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

// Get returns a complaint by id.
func (s *Service) Get(ctx context.Context, id string) (*Complaint, error) {
	if _, err := uuidutil.FromString(id); err != nil {
		return nil, fmt.Errorf("%w: invalid complaint id", ErrInvalidRequest)
	}

	item, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	mapped, err := toAPIComplaint(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Create validates input, resolves snapshots, and inserts a complaint.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Complaint, error) {
	residentID, name, phone, block, category, urgency, message, err := s.resolveCreateInput(ctx, req)
	if err != nil {
		return nil, err
	}

	var created db.Complaint
	for attempt := 0; attempt < maxRefAttempts; attempt++ {
		ref, err := s.generate(s.now())
		if err != nil {
			return nil, err
		}

		created, err = s.store.Create(ctx, db.CreateComplaintParams{
			Ref:          ref,
			ResidentID:   residentID,
			ResidentName: name,
			Phone:        phone,
			Block:        block,
			Category:     category,
			Urgency:      urgency,
			Status:       db.ComplaintStatusBARU,
			Message:      message,
		})
		if err == nil {
			mapped, mapErr := toAPIComplaint(created)
			if mapErr != nil {
				return nil, mapErr
			}
			return &mapped, nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			continue
		}
		return nil, err
	}

	return nil, fmt.Errorf("failed to allocate unique complaint reference")
}

// Update applies a partial update. Status cannot be changed here.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Complaint, error) {
	if _, err := uuidutil.FromString(id); err != nil {
		return nil, fmt.Errorf("%w: invalid complaint id", ErrInvalidRequest)
	}

	existing, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	residentID := existing.ResidentID
	name := existing.ResidentName
	phone := existing.Phone
	block := existing.Block
	category := existing.Category
	urgency := existing.Urgency
	message := existing.Message

	if req.ResidentID.Present {
		if !req.ResidentID.Valid {
			// Explicit null: unlink resident, keep historical snapshot.
			residentID = pgtype.UUID{}
		} else {
			resolved, snapName, snapPhone, err := s.resolveResident(ctx, req.ResidentID.Value)
			if err != nil {
				return nil, err
			}
			residentID = resolved
			name = snapName
			phone = snapPhone
		}
	}

	if req.Block != nil {
		block = strings.TrimSpace(*req.Block)
	}
	if req.Category != nil {
		category = strings.TrimSpace(*req.Category)
		if category == "" {
			return nil, fmt.Errorf("%w: category is required", ErrInvalidRequest)
		}
	}
	if req.Urgency != nil {
		parsed, err := parseUrgency(strings.TrimSpace(*req.Urgency))
		if err != nil {
			return nil, err
		}
		urgency = parsed
	}
	if req.Message != nil {
		message = strings.TrimSpace(*req.Message)
		if message == "" {
			return nil, fmt.Errorf("%w: message is required", ErrInvalidRequest)
		}
	}

	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid complaint id", ErrInvalidRequest)
	}

	updated, err := s.store.Update(ctx, db.UpdateComplaintParams{
		ID:           pgID,
		ResidentID:   residentID,
		ResidentName: name,
		Phone:        phone,
		Block:        block,
		Category:     category,
		Urgency:      urgency,
		Message:      message,
	})
	if err != nil {
		return nil, err
	}

	mapped, err := toAPIComplaint(updated)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// UpdateStatus applies an allowed status transition.
func (s *Service) UpdateStatus(ctx context.Context, id, statusRaw string) (*Complaint, error) {
	if _, err := uuidutil.FromString(id); err != nil {
		return nil, fmt.Errorf("%w: invalid complaint id", ErrInvalidRequest)
	}

	next, err := parseStatus(strings.TrimSpace(statusRaw))
	if err != nil {
		return nil, err
	}

	existing, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !isAllowedTransition(existing.Status, next) {
		return nil, ErrInvalidStatusTransition
	}

	updated, err := s.store.UpdateStatus(ctx, id, next)
	if err != nil {
		return nil, err
	}

	mapped, err := toAPIComplaint(updated)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Delete permanently removes a complaint.
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := uuidutil.FromString(id); err != nil {
		return fmt.Errorf("%w: invalid complaint id", ErrInvalidRequest)
	}
	if _, err := s.store.GetByID(ctx, id); err != nil {
		return err
	}
	return s.store.Delete(ctx, id)
}

func (s *Service) resolveCreateInput(ctx context.Context, req CreateRequest) (pgtype.UUID, string, string, string, string, db.ComplaintUrgency, string, error) {
	block := strings.TrimSpace(req.Block)
	category := strings.TrimSpace(req.Category)
	message := strings.TrimSpace(req.Message)
	urgencyRaw := strings.TrimSpace(req.Urgency)

	if category == "" {
		return pgtype.UUID{}, "", "", "", "", "", "", fmt.Errorf("%w: category is required", ErrInvalidRequest)
	}
	if message == "" {
		return pgtype.UUID{}, "", "", "", "", "", "", fmt.Errorf("%w: message is required", ErrInvalidRequest)
	}

	urgency := db.ComplaintUrgencyNORMAL
	if urgencyRaw != "" {
		parsed, err := parseUrgency(urgencyRaw)
		if err != nil {
			return pgtype.UUID{}, "", "", "", "", "", "", err
		}
		urgency = parsed
	}

	if req.ResidentID != nil {
		id := strings.TrimSpace(*req.ResidentID)
		if id == "" {
			return pgtype.UUID{}, "", "", "", "", "", "", fmt.Errorf("%w: invalid resident id", ErrInvalidRequest)
		}
		residentID, name, phone, err := s.resolveResident(ctx, id)
		if err != nil {
			return pgtype.UUID{}, "", "", "", "", "", "", err
		}
		return residentID, name, phone, block, category, urgency, message, nil
	}

	name := strings.TrimSpace(req.ResidentName)
	phone := strings.TrimSpace(req.Phone)
	if name == "" {
		return pgtype.UUID{}, "", "", "", "", "", "", fmt.Errorf("%w: resident_name is required", ErrInvalidRequest)
	}
	if phone == "" {
		return pgtype.UUID{}, "", "", "", "", "", "", fmt.Errorf("%w: phone is required", ErrInvalidRequest)
	}

	return pgtype.UUID{}, name, phone, block, category, urgency, message, nil
}

func (s *Service) resolveResident(ctx context.Context, id string) (pgtype.UUID, string, string, error) {
	if _, err := uuidutil.FromString(id); err != nil {
		return pgtype.UUID{}, "", "", fmt.Errorf("%w: invalid resident id", ErrInvalidRequest)
	}

	resident, err := s.residents.GetByID(ctx, id)
	if err != nil {
		return pgtype.UUID{}, "", "", err
	}

	pgID := resident.ID
	if !pgID.Valid {
		pgID, err = uuidutil.FromString(id)
		if err != nil {
			return pgtype.UUID{}, "", "", fmt.Errorf("%w: invalid resident id", ErrInvalidRequest)
		}
	}

	return pgID, resident.Name, resident.Phone, nil
}

type listParams struct {
	page     int
	pageSize int
	search   string
	status   *db.ComplaintStatus
	urgency  *db.ComplaintUrgency
	category string
}

func normalizeListParams(page, pageSize int, search, status, urgency, category string) (listParams, error) {
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
		search:   strings.TrimSpace(search),
		category: strings.TrimSpace(category),
	}

	if status = strings.TrimSpace(status); status != "" {
		parsed, err := parseStatus(status)
		if err != nil {
			return listParams{}, err
		}
		params.status = &parsed
	}
	if urgency = strings.TrimSpace(urgency); urgency != "" {
		parsed, err := parseUrgency(urgency)
		if err != nil {
			return listParams{}, err
		}
		params.urgency = &parsed
	}

	return params, nil
}

func parseUrgency(value string) (db.ComplaintUrgency, error) {
	switch db.ComplaintUrgency(value) {
	case db.ComplaintUrgencyPRIORITY, db.ComplaintUrgencyMEDIUM, db.ComplaintUrgencyNORMAL:
		return db.ComplaintUrgency(value), nil
	default:
		return "", fmt.Errorf("%w: urgency must be PRIORITY, MEDIUM, or NORMAL", ErrInvalidRequest)
	}
}

func parseStatus(value string) (db.ComplaintStatus, error) {
	switch db.ComplaintStatus(value) {
	case db.ComplaintStatusBARU, db.ComplaintStatusDIPROSES, db.ComplaintStatusSELESAI, db.ComplaintStatusDITOLAK:
		return db.ComplaintStatus(value), nil
	default:
		return "", fmt.Errorf("%w: status must be BARU, DIPROSES, SELESAI, or DITOLAK", ErrInvalidRequest)
	}
}

func isAllowedTransition(from, to db.ComplaintStatus) bool {
	switch from {
	case db.ComplaintStatusBARU:
		return to == db.ComplaintStatusDIPROSES || to == db.ComplaintStatusDITOLAK
	case db.ComplaintStatusDIPROSES:
		return to == db.ComplaintStatusSELESAI || to == db.ComplaintStatusDITOLAK
	default:
		return false
	}
}

func toAPIComplaint(item db.Complaint) (Complaint, error) {
	id, err := uuidutil.ToString(item.ID)
	if err != nil {
		return Complaint{}, err
	}

	var residentID *string
	if item.ResidentID.Valid {
		value, err := uuidutil.ToString(item.ResidentID)
		if err != nil {
			return Complaint{}, err
		}
		residentID = &value
	}

	receivedAt, err := formatTimestamptz(item.ReceivedAt)
	if err != nil {
		return Complaint{}, err
	}
	updatedAt, err := formatTimestamptz(item.UpdatedAt)
	if err != nil {
		return Complaint{}, err
	}

	return Complaint{
		ID:           id,
		Ref:          item.Ref,
		ResidentID:   residentID,
		ResidentName: item.ResidentName,
		Phone:        item.Phone,
		Block:        item.Block,
		Category:     item.Category,
		Urgency:      item.Urgency,
		Status:       item.Status,
		Message:      item.Message,
		ReceivedAt:   receivedAt,
		UpdatedAt:    updatedAt,
	}, nil
}

func formatTimestamptz(ts pgtype.Timestamptz) (string, error) {
	if !ts.Valid {
		return "", fmt.Errorf("invalid timestamp")
	}
	return ts.Time.UTC().Format(time.RFC3339), nil
}
