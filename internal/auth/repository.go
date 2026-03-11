package auth

import (
	"context"

	"testik/internal/domain"
)

// UserRepository — интерфейс доступа к пользователям (аутентификация).
type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByOAuthID(ctx context.Context, provider, providerID string) (*domain.User, error)
}
