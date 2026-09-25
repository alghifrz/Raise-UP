package whatsapp

import (
	"context"
	"errors"
	"fmt"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SettingsStore persists the admin chatbot toggle.
type SettingsStore interface {
	BotEnabled(ctx context.Context) (bool, error)
	SetBotEnabled(ctx context.Context, enabled bool) (bool, error)
}

// SettingsRepository reads and writes the whatsapp_settings singleton.
type SettingsRepository struct {
	q *db.Queries
}

// NewSettingsRepository creates a WhatsApp settings repository.
func NewSettingsRepository(pool *pgxpool.Pool) *SettingsRepository {
	return &SettingsRepository{q: db.New(pool)}
}

// BotEnabled returns the persisted chatbot toggle.
func (r *SettingsRepository) BotEnabled(ctx context.Context) (bool, error) {
	item, err := r.q.GetWhatsAppSettings(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return true, nil
		}
		return false, fmt.Errorf("get whatsapp settings: %w", err)
	}
	return item.BotEnabled, nil
}

// SetBotEnabled updates the persisted chatbot toggle.
func (r *SettingsRepository) SetBotEnabled(ctx context.Context, enabled bool) (bool, error) {
	item, err := r.q.UpdateWhatsAppBotEnabled(ctx, enabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrNotFound
		}
		return false, fmt.Errorf("update whatsapp bot enabled: %w", err)
	}
	return item.BotEnabled, nil
}
