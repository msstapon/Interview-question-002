package usecase

import (
	"context"
	"testing"

	"example.com/interview-question-002/internal/domain"
	"example.com/interview-question-002/pkg/hash"
)

// mockUserRepo is an in-memory domain.UserRepository for unit tests.
type mockUserRepo struct {
	byUsername map[string]*domain.User
}

func newMockRepo() *mockUserRepo {
	return &mockUserRepo{byUsername: map[string]*domain.User{}}
}

func (m *mockUserRepo) Create(_ context.Context, u *domain.User) error {
	if _, ok := m.byUsername[u.Username]; ok {
		return domain.ErrConflict
	}
	m.byUsername[u.Username] = u
	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	for _, u := range m.byUsername {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockUserRepo) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	if u, ok := m.byUsername[username]; ok {
		return u, nil
	}
	return nil, domain.ErrNotFound
}

func TestRegister_Success(t *testing.T) {
	uc := NewAuth(newMockRepo(), nil)
	u, err := uc.Register(context.Background(), "alice", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID == "" {
		t.Fatal("expected user id to be set")
	}
	// Password must be stored hashed (argon2id), never plaintext.
	if u.PasswordHash == "password123" {
		t.Fatal("password stored as plaintext")
	}
	ok, err := hash.Verify("password123", u.PasswordHash)
	if err != nil || !ok {
		t.Fatalf("stored hash does not verify: ok=%v err=%v", ok, err)
	}
}

func TestRegister_Duplicate(t *testing.T) {
	repo := newMockRepo()
	uc := NewAuth(repo, nil)
	if _, err := uc.Register(context.Background(), "alice", "password123"); err != nil {
		t.Fatalf("first register failed: %v", err)
	}
	_, err := uc.Register(context.Background(), "alice", "another-pass")
	if err != domain.ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockRepo()
	h, _ := hash.Hash("correct-horse")
	repo.byUsername["bob"] = &domain.User{ID: "id-bob", Username: "bob", PasswordHash: h}

	uc := NewAuth(repo, nil) // tokens unused: invalid path returns before Issue
	_, err := uc.Login(context.Background(), "bob", "wrong-pass")
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	uc := NewAuth(newMockRepo(), nil)
	_, err := uc.Login(context.Background(), "ghost", "whatever")
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
