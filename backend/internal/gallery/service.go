package gallery

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/google/uuid"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// Store is the persistence interface used by Service.
type Store interface {
	GetByID(ctx context.Context, id string) (db.GalleryItem, error)
	List(ctx context.Context, search string, limit, offset int32) ([]db.GalleryItem, error)
	Count(ctx context.Context, search string) (int64, error)
	Create(ctx context.Context, arg db.CreateGalleryItemParams) (db.GalleryItem, error)
	Update(ctx context.Context, arg db.UpdateGalleryItemParams) (db.GalleryItem, error)
	Delete(ctx context.Context, id string) error
}

// ObjectStorage is the subset of object storage used by the gallery service.
type ObjectStorage interface {
	Upload(ctx context.Context, bucket, objectPath, contentType string, body io.Reader) (string, error)
	Delete(ctx context.Context, bucket, objectPath string) error
}

// Service implements gallery use cases.
type Service struct {
	store         Store
	objectStorage ObjectStorage
	bucket        string
}

// NewService creates a gallery service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// SetObjectStorage enables managed gallery uploads.
func (s *Service) SetObjectStorage(objectStorage ObjectStorage, bucket string) {
	s.objectStorage = objectStorage
	s.bucket = strings.TrimSpace(bucket)
}

// List returns a paginated filtered gallery list.
func (s *Service) List(ctx context.Context, page, pageSize int, search string) (*ListResult, error) {
	page, pageSize, err := normalizePage(page, pageSize)
	if err != nil {
		return nil, err
	}
	search = strings.TrimSpace(search)
	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	items, err := s.store.List(ctx, search, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.store.Count(ctx, search)
	if err != nil {
		return nil, err
	}

	out := make([]Item, 0, len(items))
	for _, item := range items {
		mapped, err := toAPIItem(item)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}

	return &ListResult{Items: out, Meta: listMeta(page, pageSize, total)}, nil
}

// Get returns a gallery item by id.
func (s *Service) Get(ctx context.Context, id string) (*Item, error) {
	if _, err := uuidutil.FromString(id); err != nil {
		return nil, fmt.Errorf("%w: invalid gallery item id", ErrInvalidRequest)
	}
	item, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	mapped, err := toAPIItem(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Create validates and inserts a gallery item.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Item, error) {
	imageURL := strings.TrimSpace(req.ImageURL)
	if imageURL == "" {
		return nil, fmt.Errorf("%w: image_url is required", ErrInvalidRequest)
	}
	sortOrder := int32(0)
	if req.SortOrder != nil {
		if *req.SortOrder < 0 {
			return nil, fmt.Errorf("%w: sort_order must be >= 0", ErrInvalidRequest)
		}
		sortOrder = *req.SortOrder
	}

	item, err := s.store.Create(ctx, db.CreateGalleryItemParams{
		ImageUrl:    imageURL,
		StoragePath: strings.TrimSpace(req.StoragePath),
		Caption:     strings.TrimSpace(req.Caption),
		SortOrder:   sortOrder,
	})
	if err != nil {
		return nil, err
	}
	mapped, err := toAPIItem(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// CreateUploaded validates, stores, and creates a gallery item for an image upload.
func (s *Service) CreateUploaded(ctx context.Context, req UploadRequest) (*Item, error) {
	contentType, extension, err := validateImage(req.Data)
	if err != nil {
		return nil, err
	}
	if s.objectStorage == nil || s.bucket == "" {
		return nil, ErrStorageUnavailable
	}
	sortOrder, err := validatedSortOrder(req.SortOrder)
	if err != nil {
		return nil, err
	}

	objectPath := newObjectPath(extension)
	publicURL, err := s.objectStorage.Upload(
		ctx, s.bucket, objectPath, contentType, bytes.NewReader(req.Data),
	)
	if err != nil {
		return nil, fmt.Errorf("upload gallery image: %w", err)
	}

	item, err := s.store.Create(ctx, db.CreateGalleryItemParams{
		ImageUrl:    publicURL,
		StoragePath: objectPath,
		Caption:     strings.TrimSpace(req.Caption),
		SortOrder:   sortOrder,
	})
	if err != nil {
		_ = s.objectStorage.Delete(context.WithoutCancel(ctx), s.bucket, objectPath)
		return nil, err
	}
	mapped, err := toAPIItem(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Update partially updates a gallery item.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Item, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid gallery item id", ErrInvalidRequest)
	}
	if _, err := s.store.GetByID(ctx, id); err != nil {
		return nil, err
	}

	arg := db.UpdateGalleryItemParams{ID: pgID}
	if req.ImageURL != nil {
		imageURL := strings.TrimSpace(*req.ImageURL)
		if imageURL == "" {
			return nil, fmt.Errorf("%w: image_url cannot be empty", ErrInvalidRequest)
		}
		arg.ImageUrl = &imageURL
	}
	if req.StoragePath != nil {
		storagePath := strings.TrimSpace(*req.StoragePath)
		arg.StoragePath = &storagePath
	}
	if req.Caption != nil {
		caption := strings.TrimSpace(*req.Caption)
		arg.Caption = &caption
	}
	if req.SortOrder != nil {
		if *req.SortOrder < 0 {
			return nil, fmt.Errorf("%w: sort_order must be >= 0", ErrInvalidRequest)
		}
		arg.SortOrder = req.SortOrder
	}

	item, err := s.store.Update(ctx, arg)
	if err != nil {
		return nil, err
	}
	mapped, err := toAPIItem(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// ReplaceImage uploads a new object and swaps it into an existing gallery item.
func (s *Service) ReplaceImage(ctx context.Context, id string, req ReplaceImageRequest) (*Item, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid gallery item id", ErrInvalidRequest)
	}
	existing, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	contentType, extension, err := validateImage(req.Data)
	if err != nil {
		return nil, err
	}
	if s.objectStorage == nil || s.bucket == "" {
		return nil, ErrStorageUnavailable
	}

	objectPath := newObjectPath(extension)
	publicURL, err := s.objectStorage.Upload(
		ctx, s.bucket, objectPath, contentType, bytes.NewReader(req.Data),
	)
	if err != nil {
		return nil, fmt.Errorf("upload replacement gallery image: %w", err)
	}

	arg := db.UpdateGalleryItemParams{
		ID:          pgID,
		ImageUrl:    &publicURL,
		StoragePath: &objectPath,
	}
	if req.UpdateCaption {
		caption := strings.TrimSpace(req.Caption)
		arg.Caption = &caption
	}
	if req.UpdateSortOrder {
		sortOrder, validateErr := validatedSortOrder(req.SortOrder)
		if validateErr != nil {
			_ = s.objectStorage.Delete(context.WithoutCancel(ctx), s.bucket, objectPath)
			return nil, validateErr
		}
		arg.SortOrder = &sortOrder
	}

	item, err := s.store.Update(ctx, arg)
	if err != nil {
		_ = s.objectStorage.Delete(context.WithoutCancel(ctx), s.bucket, objectPath)
		return nil, err
	}
	if existing.StoragePath != "" {
		_ = s.objectStorage.Delete(context.WithoutCancel(ctx), s.bucket, existing.StoragePath)
	}
	mapped, err := toAPIItem(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Delete removes a gallery item.
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := uuidutil.FromString(id); err != nil {
		return fmt.Errorf("%w: invalid gallery item id", ErrInvalidRequest)
	}
	item, err := s.store.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if item.StoragePath != "" {
		if s.objectStorage == nil || s.bucket == "" {
			return ErrStorageUnavailable
		}
		if err := s.objectStorage.Delete(ctx, s.bucket, item.StoragePath); err != nil {
			return fmt.Errorf("delete gallery image: %w", err)
		}
	}
	return s.store.Delete(ctx, id)
}

func validatedSortOrder(value *int32) (int32, error) {
	if value == nil {
		return 0, nil
	}
	if *value < 0 {
		return 0, fmt.Errorf("%w: sort_order must be >= 0", ErrInvalidRequest)
	}
	return *value, nil
}

func validateImage(data []byte) (contentType, extension string, err error) {
	if len(data) == 0 {
		return "", "", fmt.Errorf("%w: image file is required", ErrInvalidRequest)
	}
	switch detected := http.DetectContentType(data); detected {
	case "image/jpeg":
		return detected, ".jpg", nil
	case "image/png":
		return detected, ".png", nil
	case "image/webp":
		return detected, ".webp", nil
	default:
		return "", "", fmt.Errorf(
			"%w: image must be JPEG, PNG, or WebP", ErrInvalidRequest,
		)
	}
}

func newObjectPath(extension string) string {
	now := time.Now().UTC()
	return path.Join(
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", int(now.Month())),
		uuid.NewString()+extension,
	)
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

func listMeta(page, pageSize int, total int64) ListMeta {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return ListMeta{Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages}
}

func toAPIItem(item db.GalleryItem) (Item, error) {
	id, err := uuidutil.ToString(item.ID)
	if err != nil {
		return Item{}, fmt.Errorf("map gallery item id: %w", err)
	}
	if !item.CreatedAt.Valid {
		return Item{}, fmt.Errorf("map gallery item created_at: invalid timestamp")
	}
	return Item{
		ID:          id,
		ImageURL:    item.ImageUrl,
		StoragePath: item.StoragePath,
		Caption:     item.Caption,
		SortOrder:   item.SortOrder,
		CreatedAt:   item.CreatedAt.Time.UTC().Format(time.RFC3339),
	}, nil
}
