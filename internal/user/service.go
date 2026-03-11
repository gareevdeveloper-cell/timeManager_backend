package user

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"testik/internal/domain"
	"testik/pkg/storage"
)

// WorkStatusPublisher — публикация смены статуса в WebSocket (опционально).
type WorkStatusPublisher interface {
	PublishWorkStatusChanged(userID, workStatus string)
}

// Service — сервис управления профилем пользователя.
type Service struct {
	repo         UserRepository
	skillRepo    SkillRepository
	statusHist   WorkStatusHistoryRepository
	storage      storage.Storage
	statusPub    WorkStatusPublisher
}

// NewService создаёт UserService.
func NewService(repo UserRepository, skillRepo SkillRepository, statusHist WorkStatusHistoryRepository, st storage.Storage, statusPub WorkStatusPublisher) *Service {
	return &Service{
		repo:       repo,
		skillRepo:  skillRepo,
		statusHist: statusHist,
		storage:    st,
		statusPub:  statusPub,
	}
}

// ErrInvalidWorkStatus — недопустимый рабочий статус.
var ErrInvalidWorkStatus = errors.New("invalid work status")

// GetUserByID возвращает пользователя по ID.
func (s *Service) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

// GetUserSkills возвращает скиллы пользователя.
func (s *Service) GetUserSkills(ctx context.Context, userID string) ([]string, error) {
	return s.skillRepo.GetByUserID(ctx, userID)
}

// UpdateProfileInput — входные данные для обновления профиля.
type UpdateProfileInput struct {
	About    string
	Position string
	Skills   []string // nil = не менять, [] = очистить, [x,y] = заменить
}

// UpdateProfile обновляет профиль пользователя.
func (s *Service) UpdateProfile(ctx context.Context, userID string, in UpdateProfileInput) (*domain.User, error) {
	if err := s.repo.UpdateProfile(ctx, userID, in.About, in.Position); err != nil {
		return nil, err
	}
	if in.Skills != nil {
		if err := s.skillRepo.SetForUser(ctx, userID, in.Skills); err != nil {
			return nil, err
		}
	}
	return s.repo.GetByID(ctx, userID)
}

// UpdateWorkStatus обновляет рабочий статус пользователя и записывает в историю.
func (s *Service) UpdateWorkStatus(ctx context.Context, userID string, workStatus string, changedBy string) (*domain.User, error) {
	if !domain.ValidWorkStatuses[workStatus] {
		return nil, ErrInvalidWorkStatus
	}
	if err := s.repo.UpdateWorkStatus(ctx, userID, workStatus); err != nil {
		return nil, err
	}
	if s.statusHist != nil {
		_ = s.statusHist.Create(ctx, userID, workStatus, changedBy)
	}
	if s.statusPub != nil {
		s.statusPub.PublishWorkStatusChanged(userID, workStatus)
	}
	return s.repo.GetByID(ctx, userID)
}

// GetWorkStatusHistory возвращает историю изменений статуса пользователя.
func (s *Service) GetWorkStatusHistory(ctx context.Context, userID string, limit int) ([]WorkStatusHistoryEntry, error) {
	if s.statusHist == nil {
		return nil, nil
	}
	return s.statusHist.GetByUserID(ctx, userID, limit)
}

// SetAvatar устанавливает аватарку пользователя.
func (s *Service) SetAvatar(ctx context.Context, userID string, avatar *storage.AvatarInput) (*domain.User, error) {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil || u == nil {
		return nil, err
	}
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	avatarURL, err := storage.UploadAvatar(ctx, s.storage, "users", parsedID, avatar)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateAvatarURL(ctx, userID, avatarURL); err != nil {
		return nil, err
	}
	u.AvatarURL = avatarURL
	return u, nil
}
