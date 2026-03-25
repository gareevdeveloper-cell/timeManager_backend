package project

import (
	"context"

	"github.com/google/uuid"

	"testik/internal/domain"
)

// ProjectRepository — интерфейс доступа к проектам.
type ProjectRepository interface {
	Create(ctx context.Context, p *domain.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	GetByKey(ctx context.Context, key string) (*domain.Project, error)
	ExistsByKey(ctx context.Context, key string) (bool, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Project, error)
	ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]*domain.Project, error)
	ListAccessibleByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Project, error)
	IncrementNextTaskNumber(ctx context.Context, projectID uuid.UUID) (int, error)
	AddMember(ctx context.Context, projectID, userID uuid.UUID, role string) error
	ListMembers(ctx context.Context, projectID uuid.UUID) ([]*domain.MemberWithRole, error)
}

// TeamRepository — минимальный интерфейс для проверки членства в команде.
type TeamRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Team, error)
	IsMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error)
}

// ProjectStatusRepository — интерфейс доступа к статусам проекта.
type ProjectStatusRepository interface {
	Create(ctx context.Context, s *domain.ProjectStatus) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ProjectStatus, error)
	GetByProjectAndKey(ctx context.Context, projectID uuid.UUID, key string) (*domain.ProjectStatus, error)
	ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]*domain.ProjectStatus, error)
	Update(ctx context.Context, s *domain.ProjectStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteWithMove(ctx context.Context, deleteID, moveToID uuid.UUID) error
	ExistsByKey(ctx context.Context, projectID uuid.UUID, key string) (bool, error)
	ExistsByTitle(ctx context.Context, projectID uuid.UUID, title string, excludeID *uuid.UUID) (bool, error)
	CountByProject(ctx context.Context, projectID uuid.UUID) (int, error)
	GetFirstByProject(ctx context.Context, projectID uuid.UUID) (*domain.ProjectStatus, error)
}

// TaskRepository — интерфейс доступа к задачам.
type TaskRepository interface {
	Create(ctx context.Context, t *domain.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	GetByKey(ctx context.Context, key string) (*domain.Task, error)
	ListByProjectID(ctx context.Context, projectID uuid.UUID, filter TaskListFilter) ([]*domain.Task, error)
	ListByProjectIDGroupedByStatus(ctx context.Context, projectID uuid.UUID, filter TaskListFilter) (map[string][]*domain.Task, error)
	ListByAssigneeID(ctx context.Context, assigneeID uuid.UUID) ([]*domain.Task, error)
	Update(ctx context.Context, t *domain.Task) error
	UpdateTaskStatusKeyByStatusID(ctx context.Context, statusID uuid.UUID, newKey string) error
	Delete(ctx context.Context, id uuid.UUID) error
}
