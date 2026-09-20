package announcement

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
	GetByID(ctx context.Context, id string) (db.Announcement, error)
	ListRecipientIDs(ctx context.Context, announcementID string) ([]string, error)
	List(ctx context.Context, search string, status *db.AnnouncementStatus, visibility *db.AnnouncementVisibility, category string, limit, offset int32) ([]db.Announcement, error)
	Count(ctx context.Context, search string, status *db.AnnouncementStatus, visibility *db.AnnouncementVisibility, category string) (int64, error)
	CreateWithRecipients(ctx context.Context, arg db.CreateAnnouncementParams, recipientIDs []string) (db.Announcement, error)
	UpdateWithRecipients(ctx context.Context, arg db.UpdateAnnouncementParams, recipientIDs []string, replace bool) (db.Announcement, error)
	Publish(ctx context.Context, id string, publishedAt pgtype.Timestamptz) (db.Announcement, error)
	Delete(ctx context.Context, id string) error
}

// ResidentReader loads residents to validate recipient IDs.
type ResidentReader interface {
	GetByID(ctx context.Context, id string) (db.Resident, error)
}

// Service implements announcement use cases.
type Service struct {
	store     Store
	residents ResidentReader
	now       func() time.Time
}

// NewService creates an announcement service.
func NewService(store Store, residents ResidentReader) *Service {
	return &Service{
		store:     store,
		residents: residents,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

// List returns a paginated filtered announcement list (without recipient IDs).
func (s *Service) List(ctx context.Context, page, pageSize int, search, status, visibility, category string) (*ListResult, error) {
	params, err := normalizeListParams(page, pageSize, search, status, visibility, category)
	if err != nil {
		return nil, err
	}

	offset := int32((params.page - 1) * params.pageSize)
	limit := int32(params.pageSize)

	items, err := s.store.List(ctx, params.search, params.status, params.visibility, params.category, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.store.Count(ctx, params.search, params.status, params.visibility, params.category)
	if err != nil {
		return nil, err
	}

	out := make([]AnnouncementSummary, 0, len(items))
	for _, item := range items {
		mapped, err := toAPIAnnouncement(item, nil, false)
		if err != nil {
			return nil, err
		}
		out = append(out, AnnouncementSummary{
			ID:           mapped.ID,
			Title:        mapped.Title,
			Excerpt:      mapped.Excerpt,
			Body:         mapped.Body,
			Category:     mapped.Category,
			Visibility:   mapped.Visibility,
			Status:       mapped.Status,
			ThumbnailURL: mapped.ThumbnailURL,
			AuthorID:     mapped.AuthorID,
			PublishedAt:  mapped.PublishedAt,
			CreatedAt:    mapped.CreatedAt,
			UpdatedAt:    mapped.UpdatedAt,
		})
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

// Get returns an announcement with recipient IDs.
func (s *Service) Get(ctx context.Context, id string) (*Announcement, error) {
	if _, err := uuidutil.FromString(id); err != nil {
		return nil, fmt.Errorf("%w: invalid announcement id", ErrInvalidRequest)
	}

	item, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	recipientIDs := []string{}
	if item.Visibility == db.AnnouncementVisibilityPRIVATE {
		recipientIDs, err = s.store.ListRecipientIDs(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	mapped, err := toAPIAnnouncement(item, recipientIDs, true)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Create creates a DRAFT announcement authored by authorID.
func (s *Service) Create(ctx context.Context, authorID string, req CreateRequest) (*Announcement, error) {
	authorUUID, err := uuidutil.FromString(authorID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid author id", ErrInvalidRequest)
	}

	title, excerpt, body, category, visibility, thumbnail, recipients, err := s.validateCreate(ctx, req)
	if err != nil {
		return nil, err
	}

	item, err := s.store.CreateWithRecipients(ctx, db.CreateAnnouncementParams{
		Title:        title,
		Excerpt:      excerpt,
		Body:         body,
		Category:     category,
		Visibility:   visibility,
		Status:       db.AnnouncementStatusDRAFT,
		ThumbnailUrl: thumbnail,
		AuthorID:     authorUUID,
		PublishedAt:  pgtype.Timestamptz{},
	}, recipients)
	if err != nil {
		return nil, err
	}

	mapped, err := toAPIAnnouncement(item, recipients, true)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Update applies a partial content/visibility/recipient update.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Announcement, error) {
	if _, err := uuidutil.FromString(id); err != nil {
		return nil, fmt.Errorf("%w: invalid announcement id", ErrInvalidRequest)
	}

	existing, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	title := existing.Title
	excerpt := existing.Excerpt
	body := existing.Body
	category := existing.Category
	visibility := existing.Visibility
	thumbnail := existing.ThumbnailUrl

	if req.Title != nil {
		title = strings.TrimSpace(*req.Title)
		if title == "" {
			return nil, fmt.Errorf("%w: title is required", ErrInvalidRequest)
		}
	}
	if req.Excerpt != nil {
		excerpt = strings.TrimSpace(*req.Excerpt)
	}
	if req.Body != nil {
		body = strings.TrimSpace(*req.Body)
		if body == "" {
			return nil, fmt.Errorf("%w: body is required", ErrInvalidRequest)
		}
	}
	if req.Category != nil {
		category = strings.TrimSpace(*req.Category)
		if category == "" {
			return nil, fmt.Errorf("%w: category is required", ErrInvalidRequest)
		}
	}
	if req.Visibility != nil {
		parsed, err := parseVisibility(strings.TrimSpace(*req.Visibility))
		if err != nil {
			return nil, err
		}
		visibility = parsed
	}
	if req.ThumbnailURL != nil {
		trimmed := strings.TrimSpace(*req.ThumbnailURL)
		if trimmed == "" {
			thumbnail = nil
		} else {
			thumbnail = &trimmed
		}
	}

	currentRecipients, err := s.store.ListRecipientIDs(ctx, id)
	if err != nil {
		return nil, err
	}

	replace := false
	recipients := currentRecipients

	switch {
	case visibility == db.AnnouncementVisibilityPUBLIC:
		if req.RecipientIDs != nil && len(*req.RecipientIDs) > 0 {
			return nil, fmt.Errorf("%w: public announcement cannot have recipients", ErrInvalidRequest)
		}
		if existing.Visibility == db.AnnouncementVisibilityPRIVATE || len(currentRecipients) > 0 {
			replace = true
			recipients = []string{}
		}
		if req.RecipientIDs != nil {
			replace = true
			recipients = []string{}
		}
	case visibility == db.AnnouncementVisibilityPRIVATE:
		if req.RecipientIDs != nil {
			normalized, err := s.normalizeRecipients(ctx, *req.RecipientIDs, true)
			if err != nil {
				return nil, err
			}
			replace = true
			recipients = normalized
		} else if existing.Visibility == db.AnnouncementVisibilityPUBLIC {
			return nil, ErrRecipientRequired
		}
	}

	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid announcement id", ErrInvalidRequest)
	}

	item, err := s.store.UpdateWithRecipients(ctx, db.UpdateAnnouncementParams{
		ID:           pgID,
		Title:        title,
		Excerpt:      excerpt,
		Body:         body,
		Category:     category,
		Visibility:   visibility,
		ThumbnailUrl: thumbnail,
	}, recipients, replace)
	if err != nil {
		return nil, err
	}

	if !replace {
		recipients, err = s.store.ListRecipientIDs(ctx, id)
		if err != nil {
			return nil, err
		}
	}
	if item.Visibility == db.AnnouncementVisibilityPUBLIC {
		recipients = []string{}
	}

	mapped, err := toAPIAnnouncement(item, recipients, true)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// UpdateStatus publishes a draft announcement.
func (s *Service) UpdateStatus(ctx context.Context, id, statusRaw string) (*Announcement, error) {
	if _, err := uuidutil.FromString(id); err != nil {
		return nil, fmt.Errorf("%w: invalid announcement id", ErrInvalidRequest)
	}

	next, err := parseStatus(strings.TrimSpace(statusRaw))
	if err != nil {
		return nil, err
	}

	existing, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing.Status != db.AnnouncementStatusDRAFT || next != db.AnnouncementStatusPUBLISHED {
		return nil, ErrInvalidStatusTransition
	}

	publishedAt := pgtype.Timestamptz{Time: s.now(), Valid: true}
	item, err := s.store.Publish(ctx, id, publishedAt)
	if err != nil {
		return nil, err
	}

	recipients := []string{}
	if item.Visibility == db.AnnouncementVisibilityPRIVATE {
		recipients, err = s.store.ListRecipientIDs(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	mapped, err := toAPIAnnouncement(item, recipients, true)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Delete permanently removes an announcement.
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := uuidutil.FromString(id); err != nil {
		return fmt.Errorf("%w: invalid announcement id", ErrInvalidRequest)
	}
	if _, err := s.store.GetByID(ctx, id); err != nil {
		return err
	}
	return s.store.Delete(ctx, id)
}

func (s *Service) validateCreate(ctx context.Context, req CreateRequest) (string, string, string, string, db.AnnouncementVisibility, *string, []string, error) {
	title := strings.TrimSpace(req.Title)
	excerpt := strings.TrimSpace(req.Excerpt)
	body := strings.TrimSpace(req.Body)
	category := strings.TrimSpace(req.Category)
	visibilityRaw := strings.TrimSpace(req.Visibility)

	if title == "" {
		return "", "", "", "", "", nil, nil, fmt.Errorf("%w: title is required", ErrInvalidRequest)
	}
	if body == "" {
		return "", "", "", "", "", nil, nil, fmt.Errorf("%w: body is required", ErrInvalidRequest)
	}
	if category == "" {
		return "", "", "", "", "", nil, nil, fmt.Errorf("%w: category is required", ErrInvalidRequest)
	}
	visibility, err := parseVisibility(visibilityRaw)
	if err != nil {
		return "", "", "", "", "", nil, nil, err
	}

	if visibility == db.AnnouncementVisibilityPUBLIC && len(req.RecipientIDs) > 0 {
		return "", "", "", "", "", nil, nil, fmt.Errorf("%w: public announcement cannot have recipients", ErrInvalidRequest)
	}

	var thumbnail *string
	if req.ThumbnailURL != nil {
		trimmed := strings.TrimSpace(*req.ThumbnailURL)
		if trimmed != "" {
			thumbnail = &trimmed
		}
	}

	requireRecipients := visibility == db.AnnouncementVisibilityPRIVATE
	recipients, err := s.normalizeRecipients(ctx, req.RecipientIDs, requireRecipients)
	if err != nil {
		return "", "", "", "", "", nil, nil, err
	}

	return title, excerpt, body, category, visibility, thumbnail, recipients, nil
}

func (s *Service) normalizeRecipients(ctx context.Context, ids []string, required bool) ([]string, error) {
	if len(ids) == 0 {
		if required {
			return nil, ErrRecipientRequired
		}
		return []string{}, nil
	}

	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		if id == "" {
			return nil, fmt.Errorf("%w: invalid recipient id", ErrInvalidRequest)
		}
		if _, err := uuidutil.FromString(id); err != nil {
			return nil, fmt.Errorf("%w: invalid recipient id", ErrInvalidRequest)
		}
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("%w: duplicate recipient id", ErrInvalidRequest)
		}
		if _, err := s.residents.GetByID(ctx, id); err != nil {
			return nil, err
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

type listParams struct {
	page       int
	pageSize   int
	search     string
	status     *db.AnnouncementStatus
	visibility *db.AnnouncementVisibility
	category   string
}

func normalizeListParams(page, pageSize int, search, status, visibility, category string) (listParams, error) {
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
	if visibility = strings.TrimSpace(visibility); visibility != "" {
		parsed, err := parseVisibility(visibility)
		if err != nil {
			return listParams{}, err
		}
		params.visibility = &parsed
	}

	return params, nil
}

func parseVisibility(value string) (db.AnnouncementVisibility, error) {
	switch db.AnnouncementVisibility(value) {
	case db.AnnouncementVisibilityPUBLIC, db.AnnouncementVisibilityPRIVATE:
		return db.AnnouncementVisibility(value), nil
	default:
		return "", fmt.Errorf("%w: visibility must be PUBLIC or PRIVATE", ErrInvalidRequest)
	}
}

func parseStatus(value string) (db.AnnouncementStatus, error) {
	switch db.AnnouncementStatus(value) {
	case db.AnnouncementStatusDRAFT, db.AnnouncementStatusPUBLISHED:
		return db.AnnouncementStatus(value), nil
	default:
		return "", fmt.Errorf("%w: status must be DRAFT or PUBLISHED", ErrInvalidRequest)
	}
}

func toAPIAnnouncement(item db.Announcement, recipientIDs []string, includeRecipients bool) (Announcement, error) {
	id, err := uuidutil.ToString(item.ID)
	if err != nil {
		return Announcement{}, err
	}
	authorID, err := uuidutil.ToString(item.AuthorID)
	if err != nil {
		return Announcement{}, err
	}
	createdAt, err := formatTimestamptz(item.CreatedAt)
	if err != nil {
		return Announcement{}, err
	}
	updatedAt, err := formatTimestamptz(item.UpdatedAt)
	if err != nil {
		return Announcement{}, err
	}

	var publishedAt *string
	if item.PublishedAt.Valid {
		value := item.PublishedAt.Time.UTC().Format(time.RFC3339)
		publishedAt = &value
	}

	out := Announcement{
		ID:           id,
		Title:        item.Title,
		Excerpt:      item.Excerpt,
		Body:         item.Body,
		Category:     item.Category,
		Visibility:   item.Visibility,
		Status:       item.Status,
		ThumbnailURL: item.ThumbnailUrl,
		AuthorID:     authorID,
		PublishedAt:  publishedAt,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
	if includeRecipients {
		if recipientIDs == nil {
			recipientIDs = []string{}
		}
		out.RecipientIDs = recipientIDs
	}
	return out, nil
}

func formatTimestamptz(ts pgtype.Timestamptz) (string, error) {
	if !ts.Valid {
		return "", fmt.Errorf("invalid timestamp")
	}
	return ts.Time.UTC().Format(time.RFC3339), nil
}
