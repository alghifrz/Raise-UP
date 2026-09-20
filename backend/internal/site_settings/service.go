package sitesettings

import (
	"context"
	"fmt"
	"strings"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/pkg/uuidutil"
)

// Store is the persistence interface used by Service.
type Store interface {
	Get(ctx context.Context) (db.SiteSetting, error)
	Update(ctx context.Context, arg db.UpdateSiteSettingsParams) (db.SiteSetting, error)
}

// Service implements site settings use cases.
type Service struct {
	store Store
}

// NewService creates a site settings service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// Get returns the singleton settings.
func (s *Service) Get(ctx context.Context) (*Settings, error) {
	item, err := s.store.Get(ctx)
	if err != nil {
		return nil, err
	}
	mapped, err := toAPISettings(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

// Update patches the singleton settings.
func (s *Service) Update(ctx context.Context, req UpdateRequest) (*Settings, error) {
	current, err := s.store.Get(ctx)
	if err != nil {
		return nil, err
	}

	arg := db.UpdateSiteSettingsParams{ID: current.ID}
	arg.SiteName = trimPtr(req.SiteName)
	arg.Tagline = trimPtr(req.Tagline)
	arg.ChairmanName = trimPtr(req.ChairmanName)
	arg.ChairmanRole = trimPtr(req.ChairmanRole)
	arg.ChairmanQuote = trimPtr(req.ChairmanQuote)
	arg.ChairmanPhotoUrl = trimNullablePtr(req.ChairmanPhotoURL)
	arg.ChairmanPhotoStoragePath = trimNullablePtr(req.ChairmanPhotoStoragePath)
	arg.MapTitle = trimPtr(req.MapTitle)
	arg.MapDescription = trimPtr(req.MapDescription)
	arg.MapsUrl = trimNullablePtr(req.MapsURL)
	arg.EmbedUrl = trimNullablePtr(req.EmbedURL)
	arg.Address = trimPtr(req.Address)
	arg.Phone = trimPtr(req.Phone)
	arg.WhatsappUrl = trimNullablePtr(req.WhatsappURL)
	arg.FooterBlurb = trimPtr(req.FooterBlurb)

	item, err := s.store.Update(ctx, arg)
	if err != nil {
		return nil, err
	}
	mapped, err := toAPISettings(item)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

func trimPtr(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	return &trimmed
}

// trimNullablePtr trims optional nullable URL-like fields.
// Empty string after trim is stored as empty string (schema allows NULL, but
// COALESCE with "" is valid and keeps the field explicitly empty).
func trimNullablePtr(v *string) *string {
	return trimPtr(v)
}

func toAPISettings(item db.SiteSetting) (Settings, error) {
	id, err := uuidutil.ToString(item.ID)
	if err != nil {
		return Settings{}, fmt.Errorf("map site settings id: %w", err)
	}
	if !item.UpdatedAt.Valid {
		return Settings{}, fmt.Errorf("map site settings updated_at: invalid timestamp")
	}
	return Settings{
		ID:                       id,
		SiteName:                 item.SiteName,
		Tagline:                  item.Tagline,
		ChairmanName:             item.ChairmanName,
		ChairmanRole:             item.ChairmanRole,
		ChairmanQuote:            item.ChairmanQuote,
		ChairmanPhotoURL:         item.ChairmanPhotoUrl,
		ChairmanPhotoStoragePath: item.ChairmanPhotoStoragePath,
		MapTitle:                 item.MapTitle,
		MapDescription:           item.MapDescription,
		MapsURL:                  item.MapsUrl,
		EmbedURL:                 item.EmbedUrl,
		Address:                  item.Address,
		Phone:                    item.Phone,
		WhatsappURL:              item.WhatsappUrl,
		FooterBlurb:              item.FooterBlurb,
		UpdatedAt:                item.UpdatedAt.Time.UTC().Format(time.RFC3339),
	}, nil
}
