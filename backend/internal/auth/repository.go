package auth

import (
	"context"
	"errors"
	"fmt"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides user persistence for authentication.
type Repository struct {
	q *db.Queries
}

// NewRepository creates an auth repository backed by sqlc.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

// GetByEmail returns a user by email.
func (r *Repository) GetByEmail(ctx context.Context, email string) (db.User, error) {
	user, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, ErrNotFound
		}
		return db.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

// GetByID returns a user by UUID string.
func (r *Repository) GetByID(ctx context.Context, id string) (db.User, error) {
	pgID, err := uuidutil.FromString(id)
	if err != nil {
		return db.User{}, ErrNotFound
	}

	user, err := r.q.GetUserByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, ErrNotFound
		}
		return db.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

// Create inserts a new user.
func (r *Repository) Create(ctx context.Context, email, name, passwordHash string, role db.UserRole) (db.User, error) {
	user, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		Role:         role,
	})
	if err != nil {
		return db.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// UpdatePasswordHash updates a user's credentials hash and optional profile fields.
func (r *Repository) Update(ctx context.Context, user db.User) (db.User, error) {
	updated, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:           user.ID,
		Email:        user.Email,
		Name:         user.Name,
		PasswordHash: user.PasswordHash,
		Role:         user.Role,
	})
	if err != nil {
		return db.User{}, fmt.Errorf("update user: %w", err)
	}
	return updated, nil
}
