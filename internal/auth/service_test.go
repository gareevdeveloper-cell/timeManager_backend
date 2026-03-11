package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"testik/internal/domain"
)

type mockUserRepo struct {
	users map[string]*domain.User
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID.String() == id {
			return u, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (m *mockUserRepo) GetByOAuthID(ctx context.Context, provider, providerID string) (*domain.User, error) {
	for _, u := range m.users {
		if u.OAuthProvider == provider && u.OAuthProviderID == providerID {
			return u, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func TestAuthService_Register(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*domain.User)}
	svc := NewAuthService(repo, AuthServiceConfig{
		JWTSecret: "test-secret-key-at-least-32-characters",
		ExpiresIn: time.Hour,
	}, nil)

	ctx := context.Background()
	u, err := svc.Register(ctx, "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if u.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", u.Email)
	}
	if u.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*domain.User)}
	svc := NewAuthService(repo, AuthServiceConfig{
		JWTSecret: "test-secret",
		ExpiresIn: time.Hour,
	}, nil)
	ctx := context.Background()

	_, err := svc.Register(ctx, "dup@example.com", "password123")
	if err != nil {
		t.Fatalf("first Register: %v", err)
	}
	_, err = svc.Register(ctx, "dup@example.com", "password456")
	if err != ErrUserAlreadyExists {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}
