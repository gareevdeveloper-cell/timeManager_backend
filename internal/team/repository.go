package team

import (
	"context"

	"github.com/google/uuid"

	"testik/internal/domain"
)

// OrgRepository — минимальный интерфейс для проверки членства в организации.
type OrgRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error)
	IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
}

// Repository — интерфейс доступа к командам.
type Repository interface {
	Create(ctx context.Context, t *domain.Team) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Team, error)
	GetByOrganization(ctx context.Context, orgID uuid.UUID) ([]*domain.Team, error)
	Update(ctx context.Context, t *domain.Team) error
	UpdateAvatarURL(ctx context.Context, teamID uuid.UUID, avatarURL string) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error
	RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error
	IsMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error)
	GetMembers(ctx context.Context, teamID uuid.UUID) ([]*domain.MemberWithRole, error)
}
