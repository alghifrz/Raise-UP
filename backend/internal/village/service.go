package village

import (
	"context"
	"fmt"
	"strings"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/jackc/pgx/v5/pgtype"
)

// Store is the persistence interface used by Service.
type Store interface {
	GetProfile(ctx context.Context) (db.VillageProfile, error)
	UpdateProfile(ctx context.Context, arg db.UpdateVillageProfileParams) (db.VillageProfile, error)
	GetOfficialByID(ctx context.Context, id string) (db.VillageOfficial, error)
	ListOfficials(ctx context.Context, profileID pgtype.UUID) ([]db.VillageOfficial, error)
	CreateOfficial(ctx context.Context, arg db.CreateVillageOfficialParams) (db.VillageOfficial, error)
	UpdateOfficial(ctx context.Context, arg db.UpdateVillageOfficialParams) (db.VillageOfficial, error)
	DeleteOfficial(ctx context.Context, id string) error
}

// Service implements village profile and officials use cases.
type Service struct {
	store Store
}

// NewService creates a village service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// GetProfile returns the singleton village profile.
func (s *Service) GetProfile(ctx context.Context) (*Profile, error) {
	item, err := s.store.GetProfile(ctx)
	if err != nil {
		return nil, err
	}
	mapped, err := toAPIProfile(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// UpdateProfile patches the singleton village profile.
func (s *Service) UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*Profile, error) {
	current, err := s.store.GetProfile(ctx)
	if err != nil {
		return nil, err
	}

	arg := db.UpdateVillageProfileParams{ID: current.ID}
	if req.History != nil {
		history := strings.TrimSpace(*req.History)
		arg.History = &history
	}
	if req.Vision != nil {
		vision := strings.TrimSpace(*req.Vision)
		arg.Vision = &vision
	}
	if req.Mission != nil {
		mission := strings.TrimSpace(*req.Mission)
		arg.Mission = &mission
	}

	item, err := s.store.UpdateProfile(ctx, arg)
	if err != nil {
		return nil, err
	}
	mapped, err := toAPIProfile(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// ListOfficials returns officials for the singleton profile.
func (s *Service) ListOfficials(ctx context.Context) ([]Official, error) {
	profile, err := s.store.GetProfile(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.store.ListOfficials(ctx, profile.ID)
	if err != nil {
		return nil, err
	}
	out := make([]Official, 0, len(items))
	for _, item := range items {
		mapped, err := toAPIOfficial(item)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}
	return out, nil
}

// CreateOfficial creates an official under the singleton village profile.
func (s *Service) CreateOfficial(ctx context.Context, req CreateOfficialRequest) (*Official, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidRequest)
	}
	position := strings.TrimSpace(req.Position)
	if position == "" {
		return nil, fmt.Errorf("%w: position is required", ErrInvalidRequest)
	}
	sortOrder := int32(0)
	if req.SortOrder != nil {
		if *req.SortOrder < 0 {
			return nil, fmt.Errorf("%w: sort_order must be >= 0", ErrInvalidRequest)
		}
		sortOrder = *req.SortOrder
	}

	profile, err := s.store.GetProfile(ctx)
	if err != nil {
		return nil, err
	}

	var photoURL *string
	if trimmed := strings.TrimSpace(req.PhotoURL); trimmed != "" {
		photoURL = &trimmed
	}

	item, err := s.store.CreateOfficial(ctx, db.CreateVillageOfficialParams{
		ProfileID: profile.ID,
		Name:      name,
		Position:  position,
		PhotoUrl:  photoURL,
		SortOrder: sortOrder,
	})
	if err != nil {
		return nil, err
	}
	mapped, err := toAPIOfficial(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// UpdateOfficial partially updates an official. profile_id cannot be changed.
func (s *Service) UpdateOfficial(ctx context.Context, id string, req UpdateOfficialRequest) (*Official, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid official id", ErrInvalidRequest)
	}
	if _, err := s.store.GetOfficialByID(ctx, id); err != nil {
		return nil, err
	}

	arg := db.UpdateVillageOfficialParams{ID: pgID}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidRequest)
		}
		arg.Name = &name
	}
	if req.Position != nil {
		position := strings.TrimSpace(*req.Position)
		if position == "" {
			return nil, fmt.Errorf("%w: position cannot be empty", ErrInvalidRequest)
		}
		arg.Position = &position
	}
	if req.PhotoURL != nil {
		photo := strings.TrimSpace(*req.PhotoURL)
		arg.PhotoUrl = &photo
	}
	if req.SortOrder != nil {
		if *req.SortOrder < 0 {
			return nil, fmt.Errorf("%w: sort_order must be >= 0", ErrInvalidRequest)
		}
		arg.SortOrder = req.SortOrder
	}

	item, err := s.store.UpdateOfficial(ctx, arg)
	if err != nil {
		return nil, err
	}
	mapped, err := toAPIOfficial(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// DeleteOfficial removes an official.
func (s *Service) DeleteOfficial(ctx context.Context, id string) error {
	if _, err := uuidutil.FromString(id); err != nil {
		return fmt.Errorf("%w: invalid official id", ErrInvalidRequest)
	}
	return s.store.DeleteOfficial(ctx, id)
}

func toAPIProfile(item db.VillageProfile) (Profile, error) {
	id, err := uuidutil.ToString(item.ID)
	if err != nil {
		return Profile{}, fmt.Errorf("map village profile id: %w", err)
	}
	if !item.UpdatedAt.Valid {
		return Profile{}, fmt.Errorf("map village profile updated_at: invalid timestamp")
	}
	return Profile{
		ID:        id,
		History:   item.History,
		Vision:    item.Vision,
		Mission:   item.Mission,
		UpdatedAt: item.UpdatedAt.Time.UTC().Format(time.RFC3339),
	}, nil
}

func toAPIOfficial(item db.VillageOfficial) (Official, error) {
	id, err := uuidutil.ToString(item.ID)
	if err != nil {
		return Official{}, fmt.Errorf("map village official id: %w", err)
	}
	profileID, err := uuidutil.ToString(item.ProfileID)
	if err != nil {
		return Official{}, fmt.Errorf("map village official profile_id: %w", err)
	}
	return Official{
		ID:        id,
		ProfileID: profileID,
		Name:      item.Name,
		Position:  item.Position,
		PhotoURL:  item.PhotoUrl,
		SortOrder: item.SortOrder,
	}, nil
}
