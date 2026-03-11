package team

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"testik/internal/domain"
	"testik/pkg/storage"
)

// Service — сервис команд.
type Service struct {
	repo     Repository
	orgRepo  OrgRepository
	storage  storage.Storage
}

// NewService создаёт сервис.
func NewService(repo Repository, orgRepo OrgRepository, st storage.Storage) *Service {
	return &Service{repo: repo, orgRepo: orgRepo, storage: st}
}

// Create создаёт команду. Создатель должен быть членом организации. avatar — опционально.
func (s *Service) Create(ctx context.Context, name, description string, organizationID, creatorID uuid.UUID, avatar *storage.AvatarInput) (*domain.Team, error) {
	org, err := s.orgRepo.GetByID(ctx, organizationID)
	if err != nil || org == nil {
		return nil, ErrOrgNotFound
	}
	if org.Status == domain.OrganizationStatusArchived {
		return nil, ErrOrgNotFound
	}

	member, err := s.orgRepo.IsMember(ctx, organizationID, creatorID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, ErrUserNotInOrg
	}

	now := time.Now()
	t := &domain.Team{
		ID:             uuid.New(),
		Name:           name,
		Description:    description,
		OrganizationID: organizationID,
		CreatorID:      creatorID,
		CreatedAt:      now,
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("create team: %w", err)
	}

	// Создатель автоматически становится участником команды с ролью administrator
	if err := s.repo.AddMember(ctx, t.ID, creatorID, domain.RoleAdministrator); err != nil {
		return nil, fmt.Errorf("add creator as member: %w", err)
	}

	if avatar != nil {
		avatarURL, err := storage.UploadAvatar(ctx, s.storage, "teams", t.ID, avatar)
		if err != nil {
			if err == storage.ErrInvalidAvatar {
				return nil, ErrInvalidAvatar
			}
			return nil, err
		}
		t.AvatarURL = avatarURL
		if err := s.repo.UpdateAvatarURL(ctx, t.ID, avatarURL); err != nil {
			return nil, fmt.Errorf("save avatar url: %w", err)
		}
	}
	return t, nil
}

// GetByID возвращает команду по ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Team, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByOrganization возвращает команды организации.
func (s *Service) GetByOrganization(ctx context.Context, orgID uuid.UUID) ([]*domain.Team, error) {
	return s.repo.GetByOrganization(ctx, orgID)
}

// Update обновляет команду. avatar — опционально.
func (s *Service) Update(ctx context.Context, teamID uuid.UUID, name, description string, avatar *storage.AvatarInput) (*domain.Team, error) {
	t, err := s.repo.GetByID(ctx, teamID)
	if err != nil || t == nil {
		return nil, ErrNotFound
	}

	t.Name = name
	t.Description = description
	if avatar != nil {
		avatarURL, err := storage.UploadAvatar(ctx, s.storage, "teams", teamID, avatar)
		if err != nil {
			if err == storage.ErrInvalidAvatar {
				return nil, ErrInvalidAvatar
			}
			return nil, err
		}
		t.AvatarURL = avatarURL
	}
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("update team: %w", err)
	}
	return t, nil
}

// SetAvatar устанавливает аватарку команды. Вызывающий должен быть членом организации команды.
func (s *Service) SetAvatar(ctx context.Context, teamID uuid.UUID, userID uuid.UUID, avatar *storage.AvatarInput) (*domain.Team, error) {
	t, err := s.repo.GetByID(ctx, teamID)
	if err != nil || t == nil {
		return nil, ErrNotFound
	}
	member, err := s.orgRepo.IsMember(ctx, t.OrganizationID, userID)
	if err != nil || !member {
		return nil, ErrNotFound
	}
	avatarURL, err := storage.UploadAvatar(ctx, s.storage, "teams", teamID, avatar)
	if err != nil {
		if err == storage.ErrInvalidAvatar {
			return nil, ErrInvalidAvatar
		}
		return nil, err
	}
	if err := s.repo.UpdateAvatarURL(ctx, teamID, avatarURL); err != nil {
		return nil, fmt.Errorf("save avatar url: %w", err)
	}
	t.AvatarURL = avatarURL
	return t, nil
}

// Delete удаляет команду.
func (s *Service) Delete(ctx context.Context, teamID uuid.UUID) error {
	t, err := s.repo.GetByID(ctx, teamID)
	if err != nil || t == nil {
		return ErrNotFound
	}
	return s.repo.Delete(ctx, teamID)
}

// AddMember добавляет пользователя в команду с указанной ролью. Пользователь должен быть в организации команды.
func (s *Service) AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error {
	t, err := s.repo.GetByID(ctx, teamID)
	if err != nil || t == nil {
		return ErrNotFound
	}

	member, err := s.orgRepo.IsMember(ctx, t.OrganizationID, userID)
	if err != nil {
		return err
	}
	if !member {
		return ErrUserNotInOrg
	}

	inTeam, err := s.repo.IsMember(ctx, teamID, userID)
	if err != nil {
		return err
	}
	if inTeam {
		return ErrUserAlreadyInTeam
	}

	if role == "" {
		role = domain.RoleParticipant
	}
	if !containsRole(domain.ValidMemberRoles, role) {
		role = domain.RoleParticipant
	}
	return s.repo.AddMember(ctx, teamID, userID, role)
}

func containsRole(slice []string, v string) bool {
	for _, s := range slice {
		if s == v {
			return true
		}
	}
	return false
}

// RemoveMember удаляет пользователя из команды.
func (s *Service) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	t, err := s.repo.GetByID(ctx, teamID)
	if err != nil || t == nil {
		return ErrNotFound
	}

	inTeam, err := s.repo.IsMember(ctx, teamID, userID)
	if err != nil {
		return err
	}
	if !inTeam {
		return ErrUserNotInTeam
	}

	return s.repo.RemoveMember(ctx, teamID, userID)
}

// GetMembers возвращает список участников команды с ролями.
func (s *Service) GetMembers(ctx context.Context, teamID uuid.UUID) ([]*domain.MemberWithRole, error) {
	t, err := s.repo.GetByID(ctx, teamID)
	if err != nil || t == nil {
		return nil, ErrNotFound
	}
	return s.repo.GetMembers(ctx, teamID)
}
