package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const complaintSessionState = "AWAITING_COMPLAINT"

// SessionStore persists multi-message bot flows across restarts.
type SessionStore interface {
	IsAwaitingComplaint(ctx context.Context, phone string) (bool, error)
	SetAwaitingComplaint(ctx context.Context, phone string) error
	Clear(ctx context.Context, phone string) error
}

// PostgresSessionStore stores bot state by WhatsApp phone.
type PostgresSessionStore struct {
	pool *pgxpool.Pool
}

// NewPostgresSessionStore creates a persistent bot session store.
func NewPostgresSessionStore(pool *pgxpool.Pool) *PostgresSessionStore {
	return &PostgresSessionStore{pool: pool}
}

// IsAwaitingComplaint reports whether the next message should become a complaint.
func (s *PostgresSessionStore) IsAwaitingComplaint(ctx context.Context, phone string) (bool, error) {
	const query = `SELECT state FROM whatsapp_bot_sessions WHERE phone = $1`
	var state string
	err := s.pool.QueryRow(ctx, query, normalizeSessionPhone(phone)).Scan(&state)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("get bot session: %w", err)
	}
	return state == complaintSessionState, nil
}

// SetAwaitingComplaint starts or refreshes the complaint flow.
func (s *PostgresSessionStore) SetAwaitingComplaint(ctx context.Context, phone string) error {
	const query = `
INSERT INTO whatsapp_bot_sessions (phone, state, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (phone) DO UPDATE
SET state = EXCLUDED.state, updated_at = NOW()`
	if _, err := s.pool.Exec(ctx, query, normalizeSessionPhone(phone), complaintSessionState); err != nil {
		return fmt.Errorf("set bot session: %w", err)
	}
	return nil
}

// Clear ends any active flow for a phone.
func (s *PostgresSessionStore) Clear(ctx context.Context, phone string) error {
	if _, err := s.pool.Exec(
		ctx,
		`DELETE FROM whatsapp_bot_sessions WHERE phone = $1`,
		normalizeSessionPhone(phone),
	); err != nil {
		return fmt.Errorf("clear bot session: %w", err)
	}
	return nil
}

func normalizeSessionPhone(phone string) string {
	return strings.TrimSpace(phone)
}
