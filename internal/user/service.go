package user

import (
	"context"
	"errors"
	"time"

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
	tasks        AssigneeTaskReader
	skillRepo    SkillRepository
	statusHist   WorkStatusHistoryRepository
	storage      storage.Storage
	statusPub    WorkStatusPublisher
}

// NewService создаёт UserService.
func NewService(repo UserRepository, tasks AssigneeTaskReader, skillRepo SkillRepository, statusHist WorkStatusHistoryRepository, st storage.Storage, statusPub WorkStatusPublisher) *Service {
	return &Service{
		repo:       repo,
		tasks:      tasks,
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
	FirstName string
	LastName  string
	Birthday  *time.Time // nil в БД — нет даты; значение — установить дату
	About     string
	Position  string
	Skills    []string // nil = не менять, [] = очистить, [x,y] = заменить
}

// UpdateProfile обновляет профиль пользователя.
func (s *Service) UpdateProfile(ctx context.Context, userID string, in UpdateProfileInput) (*domain.User, error) {
	if err := s.repo.UpdateProfile(ctx, userID, in.FirstName, in.LastName, in.Birthday, in.About, in.Position); err != nil {
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

// AssigneeTaskItem — задача в списке «мои задачи» с отметкой «в работе».
type AssigneeTaskItem struct {
	Task   *domain.Task
	InWork bool
}

// ListMyAssigneeTasks возвращает все задачи, где пользователь — исполнитель.
// InWork == true только у задачи, id которой совпадает с current_task_id пользователя (не больше одной).
func (s *Service) ListMyAssigneeTasks(ctx context.Context, userID string) ([]AssigneeTaskItem, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	tasks, err := s.tasks.ListByAssigneeID(ctx, uid)
	if err != nil {
		return nil, err
	}
	var currentID *uuid.UUID
	if u != nil {
		currentID = u.CurrentTaskID
	}
	out := make([]AssigneeTaskItem, 0, len(tasks))
	for _, t := range tasks {
		inWork := currentID != nil && *currentID == t.ID
		out = append(out, AssigneeTaskItem{Task: t, InWork: inWork})
	}
	return out, nil
}

// SetCurrentTask задаёт задачу «в работе» (только одна; должна быть назначена на пользователя). taskID == nil — сброс.
func (s *Service) SetCurrentTask(ctx context.Context, userID string, taskID *uuid.UUID) (*domain.User, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	if taskID == nil {
		if err := s.repo.UpdateCurrentTaskID(ctx, userID, nil); err != nil {
			return nil, err
		}
		return s.repo.GetByID(ctx, userID)
	}
	t, err := s.tasks.GetByID(ctx, *taskID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrCurrentTaskNotFound
	}
	if t.AssigneeID == nil || *t.AssigneeID != uid {
		return nil, ErrCurrentTaskNotAssignee
	}
	if err := s.repo.UpdateCurrentTaskID(ctx, userID, taskID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, userID)
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
