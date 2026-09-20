package auth_test

import (
	"context"
	"testing"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/internal/auth"
	jwtutil "github.com/diuk/raiseup/pkg/jwt"
	"github.com/diuk/raiseup/pkg/password"
	"github.com/diuk/raiseup/pkg/uuidutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type memoryStore struct {
	byEmail map[string]db.User
	byID    map[string]db.User
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		byEmail: make(map[string]db.User),
		byID:    make(map[string]db.User),
	}
}

func (m *memoryStore) GetByEmail(_ context.Context, email string) (db.User, error) {
	user, ok := m.byEmail[email]
	if !ok {
		return db.User{}, auth.ErrNotFound
	}
	return user, nil
}

func (m *memoryStore) GetByID(_ context.Context, id string) (db.User, error) {
	user, ok := m.byID[id]
	if !ok {
		return db.User{}, auth.ErrNotFound
	}
	return user, nil
}

func (m *memoryStore) Create(_ context.Context, email, name, passwordHash string, role db.UserRole) (db.User, error) {
	id := uuid.New()
	user := db.User{
		ID:           pgtype.UUID{Bytes: id, Valid: true},
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		Role:         role,
	}
	m.byEmail[email] = user
	m.byID[id.String()] = user
	return user, nil
}

func (m *memoryStore) Update(_ context.Context, user db.User) (db.User, error) {
	id, err := uuidutil.ToString(user.ID)
	if err != nil {
		return db.User{}, err
	}
	m.byEmail[user.Email] = user
	m.byID[id] = user
	return user, nil
}

func seedUser(t *testing.T, store *memoryStore, email, plain, name string, role db.UserRole) db.User {
	t.Helper()
	hash, err := password.Hash(plain)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	user, err := store.Create(context.Background(), email, name, hash, role)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	return user
}

func TestLoginValidCredentials(t *testing.T) {
	store := newMemoryStore()
	seedUser(t, store, "admin@example.com", "password123", "Admin", db.UserRoleADMINRW)
	service := auth.NewService(store, jwtutil.NewManager("test-secret-at-least-16", time.Hour))

	result, err := service.Login(context.Background(), "admin@example.com", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.AccessToken == "" || result.TokenType != "Bearer" {
		t.Fatalf("unexpected token response: %+v", result)
	}
	if result.User.Email != "admin@example.com" || result.User.Role != db.UserRoleADMINRW {
		t.Fatalf("unexpected user: %+v", result.User)
	}
	if result.User.ID == "" {
		t.Fatal("expected user id")
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	store := newMemoryStore()
	seedUser(t, store, "admin@example.com", "password123", "Admin", db.UserRoleADMINRW)
	service := auth.NewService(store, jwtutil.NewManager("test-secret-at-least-16", time.Hour))

	_, err := service.Login(context.Background(), "admin@example.com", "wrong-password")
	if err != auth.ErrInvalidCredentials {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLoginNonexistentUser(t *testing.T) {
	store := newMemoryStore()
	service := auth.NewService(store, jwtutil.NewManager("test-secret-at-least-16", time.Hour))

	_, err := service.Login(context.Background(), "missing@example.com", "password123")
	if err != auth.ErrInvalidCredentials {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLoginEmptyCredentials(t *testing.T) {
	store := newMemoryStore()
	service := auth.NewService(store, jwtutil.NewManager("test-secret-at-least-16", time.Hour))

	_, err := service.Login(context.Background(), "", "")
	if err != auth.ErrInvalidRequest {
		t.Fatalf("Login() error = %v, want ErrInvalidRequest", err)
	}
}

func TestMe(t *testing.T) {
	store := newMemoryStore()
	user := seedUser(t, store, "admin@example.com", "password123", "Admin", db.UserRoleSUPERADMIN)
	service := auth.NewService(store, jwtutil.NewManager("test-secret-at-least-16", time.Hour))

	id, err := uuidutil.ToString(user.ID)
	if err != nil {
		t.Fatalf("ToString() error = %v", err)
	}

	got, err := service.Me(context.Background(), id)
	if err != nil {
		t.Fatalf("Me() error = %v", err)
	}
	if got.Email != "admin@example.com" || got.Role != db.UserRoleSUPERADMIN {
		t.Fatalf("unexpected user: %+v", got)
	}
}
