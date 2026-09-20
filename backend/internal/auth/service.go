package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	db "github.com/diuk/raiseup/db/generated"
	jwtutil "github.com/diuk/raiseup/pkg/jwt"
	"github.com/diuk/raiseup/pkg/password"
	"github.com/diuk/raiseup/pkg/uuidutil"
)

// Dummy hash used to reduce timing differences when a user is not found.
// Valid bcrypt hash so CompareHashAndPassword performs a full comparison.
var timingSafeDummyHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

// UserStore is the persistence interface used by Service.
type UserStore interface {
	GetByEmail(ctx context.Context, email string) (db.User, error)
	GetByID(ctx context.Context, id string) (db.User, error)
	Create(ctx context.Context, email, name, passwordHash string, role db.UserRole) (db.User, error)
	Update(ctx context.Context, user db.User) (db.User, error)
}

// Service implements authentication use cases.
type Service struct {
	store  UserStore
	tokens *jwtutil.Manager
}

// NewService creates an auth service.
func NewService(store UserStore, tokens *jwtutil.Manager) *Service {
	return &Service{store: store, tokens: tokens}
}

// Login authenticates a user and returns an access token.
func (s *Service) Login(ctx context.Context, email, plainPassword string) (*LoginResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	plainPassword = strings.TrimSpace(plainPassword)

	if email == "" || plainPassword == "" {
		return nil, ErrInvalidRequest
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, ErrInvalidRequest
	}

	user, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			_ = password.Verify(timingSafeDummyHash, plainPassword)
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !password.Verify(user.PasswordHash, plainPassword) {
		return nil, ErrInvalidCredentials
	}

	userID, err := uuidutil.ToString(user.ID)
	if err != nil {
		return nil, err
	}

	token, _, err := s.tokens.Generate(userID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.tokens.ExpiresIn().Seconds()),
		User: PublicUser{
			ID:    userID,
			Email: user.Email,
			Name:  user.Name,
			Role:  user.Role,
		},
	}, nil
}

// Me returns the public profile for an authenticated user id.
func (s *Service) Me(ctx context.Context, userID string) (*PublicUser, error) {
	user, err := s.store.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrUnauthorized
		}
		return nil, err
	}

	id, err := uuidutil.ToString(user.ID)
	if err != nil {
		return nil, err
	}

	return &PublicUser{
		ID:    id,
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	}, nil
}

// SeedAdmin creates or updates an administrator user for development bootstrap.
func (s *Service) SeedAdmin(ctx context.Context, email, name, plainPassword string, role db.UserRole) (PublicUser, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	name = strings.TrimSpace(name)
	plainPassword = strings.TrimSpace(plainPassword)

	if email == "" || name == "" || plainPassword == "" {
		return PublicUser{}, ErrInvalidRequest
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return PublicUser{}, ErrInvalidRequest
	}
	if len(plainPassword) < 8 {
		return PublicUser{}, ErrInvalidRequest
	}
	if role != db.UserRoleSUPERADMIN && role != db.UserRoleADMINRW {
		return PublicUser{}, ErrInvalidRequest
	}

	hash, err := password.Hash(plainPassword)
	if err != nil {
		return PublicUser{}, err
	}

	existing, err := s.store.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return PublicUser{}, err
	}

	var user db.User
	if errors.Is(err, ErrNotFound) {
		user, err = s.store.Create(ctx, email, name, hash, role)
		if err != nil {
			return PublicUser{}, err
		}
	} else {
		existing.Name = name
		existing.PasswordHash = hash
		existing.Role = role
		user, err = s.store.Update(ctx, existing)
		if err != nil {
			return PublicUser{}, err
		}
	}

	id, err := uuidutil.ToString(user.ID)
	if err != nil {
		return PublicUser{}, err
	}

	return PublicUser{
		ID:    id,
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	}, nil
}
