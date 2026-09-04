package service

import (
	"context"
	"strings"

	"github.com/imageforge/imageforge/ent/user"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/repository"
)

// AuthService handles registration and login. It owns password hashing and
// JWT issuance but never persists plaintext secrets.
type AuthService struct {
	users    *repository.UserRepository
	jwt      *JWTService
	password *Password
}

// NewAuthService builds an AuthService.
func NewAuthService(users *repository.UserRepository, jwt *JWTService, password *Password) *AuthService {
	return &AuthService{users: users, jwt: jwt, password: password}
}

// RegisterInput is the registration request.
type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

// AuthResult is the response envelope for auth success.
type AuthResult struct {
	Token string `json:"token"`
	User  struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
		Role  string `json:"role"`
	} `json:"user"`
}

// Register creates a new user and returns an access token.
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(in.Email))
	if email == "" || len(in.Password) < 8 {
		return nil, errors.New(errors.ErrInvalidRequest, "email and password (≥8 chars) required")
	}

	// Uniqueness check.
	existing, _ := s.users.GetByEmail(ctx, email)
	if existing != nil {
		return nil, errors.New(errors.ErrConflict, "email already registered")
	}

	hash, err := s.password.Hash(in.Password)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to process password", err)
	}

	created, err := s.users.Create(ctx, repository.CreateUserParams{
		Email:        email,
		PasswordHash: hash,
		Name:         strings.TrimSpace(in.Name),
		Role:         user.RoleUser,
	})
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to create user", err)
	}

	return s.resultWithToken(created.ID, created.Email, created.Name, string(created.Role), created.TokenVersion)
}

// LoginInput is the login request.
type LoginInput struct {
	Email    string
	Password string
}

// Login verifies credentials and returns an access token.
func (s *AuthService) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(in.Email))
	created, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		// Same error for unknown email and bad password to avoid user enumeration.
		return nil, errors.New(errors.ErrUnauthorized, "invalid email or password")
	}
	if created.Status != user.StatusActive {
		return nil, errors.New(errors.ErrForbidden, "account is not active")
	}

	if err := s.password.Verify(in.Password, created.PasswordHash); err != nil {
		return nil, errors.New(errors.ErrUnauthorized, "invalid email or password")
	}

	// Best-effort last-login update; failure must not block login.
	_ = s.users.UpdateLastLogin(ctx, created.ID)

	return s.resultWithToken(created.ID, created.Email, created.Name, string(created.Role), created.TokenVersion)
}

func (s *AuthService) resultWithToken(id int64, email, name, role string, tokenVersion int) (*AuthResult, error) {
	token, err := s.jwt.Generate(id, role, tokenVersion)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to issue token", err)
	}
	out := &AuthResult{Token: token}
	out.User.ID = id
	out.User.Email = email
	out.User.Name = name
	out.User.Role = role
	return out, nil
}
