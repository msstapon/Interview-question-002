package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"example.com/interview-question-002/internal/domain"
	"example.com/interview-question-002/pkg/hash"
	"example.com/interview-question-002/pkg/jwt"
)

type AuthUsecase struct {
	users  domain.UserRepository
	tokens *jwt.Manager
}

func NewAuth(users domain.UserRepository, tokens *jwt.Manager) *AuthUsecase {
	return &AuthUsecase{users: users, tokens: tokens}
}

// Register hashes the password with Argon2id and creates the user. Returns
// domain.ErrConflict if the username is already taken.
func (a *AuthUsecase) Register(ctx context.Context, username, password string) (*domain.User, error) {
	username = strings.TrimSpace(username)

	// Fast-path existence check; the DB unique constraint is the real guard against races.
	if _, err := a.users.GetByUsername(ctx, username); err == nil {
		return nil, domain.ErrConflict
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	h, err := hash.Hash(password)
	if err != nil {
		return nil, err
	}
	u := &domain.User{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: h,
	}
	if err := a.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// TokenPair is the successful login result.
type TokenPair struct {
	AccessToken string
	ExpiresIn   int
	User        *domain.User
}

// Login verifies the password and issues an RS256 access token. Returns
// domain.ErrInvalidCredentials for both unknown user and wrong password.
func (a *AuthUsecase) Login(ctx context.Context, username, password string) (*TokenPair, error) {
	u, err := a.users.GetByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	ok, err := hash.Verify(password, u.PasswordHash)
	if err != nil || !ok {
		return nil, domain.ErrInvalidCredentials
	}
	tok, err := a.tokens.Issue(u.ID, u.Username)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken: tok.Token,
		ExpiresIn:   int(a.tokens.AccessTTL().Seconds()),
		User:        u,
	}, nil
}

// GetUser fetches a user by ID (used by GET /me after JWT validation).
func (a *AuthUsecase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return a.users.GetByID(ctx, id)
}
