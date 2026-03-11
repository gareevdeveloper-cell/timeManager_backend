package user

import (
	"context"

	"testik/internal/domain"
)

// UserRepository — интерфейс доступа к пользователям (профиль).
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID string, about, position string) error
	UpdateAvatarURL(ctx context.Context, userID string, avatarURL string) error
	UpdateWorkStatus(ctx context.Context, userID string, workStatus string) error
}

// WorkStatusHistoryRepository — интерфейс доступа к истории статусов.
type WorkStatusHistoryRepository interface {
	Create(ctx context.Context, userID, workStatus, changedBy string) error
	GetByUserID(ctx context.Context, userID string, limit int) ([]WorkStatusHistoryEntry, error)
}

// WorkStatusHistoryEntry — запись в истории статусов.
type WorkStatusHistoryEntry struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	WorkStatus string `json:"work_status"`
	ChangedAt  string `json:"changed_at"`
	ChangedBy  string `json:"changed_by,omitempty"`
}

// SkillRepository — интерфейс доступа к скиллам.
type SkillRepository interface {
	GetByUserID(ctx context.Context, userID string) ([]string, error)
	SetForUser(ctx context.Context, userID string, skillNames []string) error
}
