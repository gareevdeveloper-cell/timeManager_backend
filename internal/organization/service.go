package organization

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"testik/internal/domain"
	"testik/pkg/slug"
	"testik/pkg/storage"
)

// Service — сервис организаций.
type Service struct {
	repo    Repository
	storage storage.Storage
}

// NewService создаёт сервис.
func NewService(repo Repository, st storage.Storage) *Service {
	return &Service{repo: repo, storage: st}
}

// AvatarInput — алиас для storage.AvatarInput (обратная совместимость handler).
type AvatarInput = storage.AvatarInput

// Create создаёт организацию. Генерирует slug из name, при коллизии добавляет суффикс -2, -3, ...
// avatar — опциональная аватарка.
func (s *Service) Create(ctx context.Context, name string, ownerID uuid.UUID, avatar *AvatarInput) (*domain.Organization, error) {
	baseSlug := slug.From(name)
	if baseSlug == "" {
		baseSlug = "org"
	}

	var candidate string
	for i := 1; ; i++ {
		if i == 1 {
			candidate = baseSlug
		} else {
			candidate = slug.WithSuffix(baseSlug, i)
		}

		exists, err := s.repo.ExistsBySlug(ctx, candidate)
		if err != nil {
			return nil, fmt.Errorf("check slug: %w", err)
		}
		if !exists {
			break
		}
	}

	now := time.Now()
	o := &domain.Organization{
		ID:        uuid.New(),
		Name:      name,
		Slug:      candidate,
		OwnerID:   ownerID,
		Status:    domain.OrganizationStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, o); err != nil {
		if IsUniqueViolation(err) {
			return nil, fmt.Errorf("slug conflict: %w", ErrSlugConflict)
		}
		return nil, fmt.Errorf("create organization: %w", err)
	}

	// Владелец автоматически становится первым членом с ролью administrator
	if err := s.repo.AddMember(ctx, o.ID, ownerID, domain.RoleAdministrator); err != nil {
		return nil, fmt.Errorf("add owner as member: %w", err)
	}

	if avatar != nil {
		avatarURL, err := s.uploadAvatar(ctx, o.ID, avatar)
		if err != nil {
			if err == storage.ErrInvalidAvatar {
				return nil, ErrInvalidAvatar
			}
			return nil, err
		}
		o.AvatarURL = avatarURL
		if err := s.repo.UpdateAvatarURL(ctx, o.ID, avatarURL); err != nil {
			return nil, fmt.Errorf("save avatar url: %w", err)
		}
	}

	return o, nil
}

func (s *Service) uploadAvatar(ctx context.Context, orgID uuid.UUID, avatar *AvatarInput) (string, error) {
	return storage.UploadAvatar(ctx, s.storage, "organizations", orgID, avatar)
}

// Update обновляет организацию (name, опционально avatar). Slug остаётся неизменным.
func (s *Service) Update(ctx context.Context, orgID uuid.UUID, name string, avatar *AvatarInput) (*domain.Organization, error) {
	o, err := s.repo.GetByID(ctx, orgID)
	if err != nil || o == nil {
		return nil, ErrNotFound
	}
	if o.Status == domain.OrganizationStatusArchived {
		return nil, ErrArchived
	}

	o.Name = name
	o.UpdatedAt = time.Now()

	if avatar != nil {
		avatarURL, err := s.uploadAvatar(ctx, orgID, avatar)
		if err != nil {
			if err == storage.ErrInvalidAvatar {
				return nil, ErrInvalidAvatar
			}
			return nil, err
		}
		o.AvatarURL = avatarURL
	}

	if err := s.repo.Update(ctx, o); err != nil {
		return nil, fmt.Errorf("update organization: %w", err)
	}
	return o, nil
}

// SetAvatar устанавливает аватарку организации. Вызывающий должен быть членом организации.
func (s *Service) SetAvatar(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, avatar *AvatarInput) (*domain.Organization, error) {
	o, err := s.repo.GetByID(ctx, orgID)
	if err != nil || o == nil {
		return nil, ErrNotFound
	}
	if o.Status == domain.OrganizationStatusArchived {
		return nil, ErrArchived
	}
	member, err := s.repo.IsMember(ctx, orgID, userID)
	if err != nil || !member {
		return nil, ErrNotFound
	}

	avatarURL, err := s.uploadAvatar(ctx, orgID, avatar)
	if err != nil {
		if err == storage.ErrInvalidAvatar {
			return nil, ErrInvalidAvatar
		}
		return nil, err
	}

	if err := s.repo.UpdateAvatarURL(ctx, orgID, avatarURL); err != nil {
		return nil, fmt.Errorf("save avatar url: %w", err)
	}
	o.AvatarURL = avatarURL
	o.UpdatedAt = time.Now()
	return o, nil
}

// Archive переводит организацию в статус archived.
func (s *Service) Archive(ctx context.Context, orgID uuid.UUID) (*domain.Organization, error) {
	o, err := s.repo.GetByID(ctx, orgID)
	if err != nil || o == nil {
		return nil, ErrNotFound
	}
	if o.Status == domain.OrganizationStatusArchived {
		return o, nil // уже архивирована
	}

	o.Status = domain.OrganizationStatusArchived
	o.UpdatedAt = time.Now()
	if err := s.repo.UpdateStatus(ctx, o); err != nil {
		return nil, fmt.Errorf("archive organization: %w", err)
	}
	return o, nil
}

// AddMember добавляет пользователя в организацию с указанной ролью.
// Пользователь может быть только в одной организации (DOMAIN.md).
func (s *Service) AddMember(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	o, err := s.repo.GetByID(ctx, orgID)
	if err != nil || o == nil {
		return ErrNotFound
	}
	if o.Status == domain.OrganizationStatusArchived {
		return ErrArchived
	}

	currentOrg, err := s.repo.GetMemberOrganization(ctx, userID)
	if err != nil {
		return err
	}
	if currentOrg != nil {
		if *currentOrg == orgID {
			return nil // уже в этой организации
		}
		return ErrUserAlreadyInOrg
	}

	if role == "" {
		role = domain.RoleParticipant
	}
	if !containsRole(domain.ValidMemberRoles, role) {
		role = domain.RoleParticipant
	}
	return s.repo.AddMember(ctx, orgID, userID, role)
}

// RemoveMember удаляет пользователя из организации.
func (s *Service) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	o, err := s.repo.GetByID(ctx, orgID)
	if err != nil || o == nil {
		return ErrNotFound
	}
	if o.Status == domain.OrganizationStatusArchived {
		return ErrArchived
	}

	member, err := s.repo.IsMember(ctx, orgID, userID)
	if err != nil {
		return err
	}
	if !member {
		return ErrUserNotInOrg
	}

	return s.repo.RemoveMember(ctx, orgID, userID)
}

func containsRole(slice []string, v string) bool {
	for _, s := range slice {
		if s == v {
			return true
		}
	}
	return false
}

// ListMyOrganizations возвращает организации, в которых состоит пользователь.
func (s *Service) ListMyOrganizations(ctx context.Context, userID uuid.UUID) ([]*domain.Organization, error) {
	return s.repo.ListByMember(ctx, userID)
}

// GetByID возвращает организацию по ID. Доступ только для членов организации.
func (s *Service) GetByID(ctx context.Context, orgID, userID uuid.UUID) (*domain.Organization, error) {
	o, err := s.repo.GetByID(ctx, orgID)
	if err != nil || o == nil {
		return nil, ErrNotFound
	}
	member, err := s.repo.IsMember(ctx, orgID, userID)
	if err != nil || !member {
		return nil, ErrNotFound
	}
	return o, nil
}

// ListMembers возвращает участников организации с ролями. Доступ только для членов организации.
func (s *Service) ListMembers(ctx context.Context, orgID, userID uuid.UUID) ([]*domain.MemberWithRole, error) {
	member, err := s.repo.IsMember(ctx, orgID, userID)
	if err != nil || !member {
		return nil, ErrNotFound
	}
	return s.repo.ListMembers(ctx, orgID)
}
