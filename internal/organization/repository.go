package organization

import (
	"context"

	"github.com/google/uuid"

	"testik/internal/domain"
)

// Repository — интерфейс доступа к организациям.
type Repository interface {
	Create(ctx context.Context, o *domain.Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error)
	Update(ctx context.Context, o *domain.Organization) error
	UpdateStatus(ctx context.Context, o *domain.Organization) error
	UpdateAvatarURL(ctx context.Context, orgID uuid.UUID, avatarURL string) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	AddMember(ctx context.Context, orgID, userID uuid.UUID, role string) error
	RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error
	IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
	GetMemberOrganization(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error)
	ListByMember(ctx context.Context, userID uuid.UUID) ([]*domain.Organization, error)
	ListMembers(ctx context.Context, orgID uuid.UUID) ([]*domain.MemberWithRole, error)
}
