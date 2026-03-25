package project

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"testik/internal/domain"
)

// TaskEventPublisher — публикация событий задач в WebSocket (опционально).
type TaskEventPublisher interface {
	PublishTaskChanged(projectID, taskID uuid.UUID, action string)
	PublishTaskStatusChanged(projectID, taskID uuid.UUID, statusKey string, statusID uuid.UUID)
}

// Дефолтные статусы при создании проекта.
var defaultStatuses = []struct{ key, title string }{
	{"TODO", "To Do"},
	{"IN_PROGRESS", "In Progress"},
	{"IN_REVIEW", "In Review"},
	{"DONE", "Done"},
}

// Service — сервис проектов и задач.
type Service struct {
	projectRepo       ProjectRepository
	projectStatusRepo ProjectStatusRepository
	taskRepo          TaskRepository
	teamRepo          TeamRepository
	taskPub           TaskEventPublisher
}

// NewService создаёт сервис.
func NewService(projectRepo ProjectRepository, projectStatusRepo ProjectStatusRepository, taskRepo TaskRepository, teamRepo TeamRepository, taskPub TaskEventPublisher) *Service {
	return &Service{
		projectRepo:       projectRepo,
		projectStatusRepo: projectStatusRepo,
		taskRepo:          taskRepo,
		teamRepo:          teamRepo,
		taskPub:           taskPub,
	}
}

var keyRegex = regexp.MustCompile(`^[A-Z][A-Z0-9]{0,49}$`)

// CreateProject создаёт проект. team_id опционален: с ним — проект в команде (создатель должен быть членом), без — личный проект.
// key должен быть уникальным и соответствовать формату (A-Z, A-Z0-9, 1-50 символов).
func (s *Service) CreateProject(ctx context.Context, teamID uuid.UUID, key string, name, description string, ownerID uuid.UUID) (*domain.Project, error) {
	if teamID != uuid.Nil {
		team, err := s.teamRepo.GetByID(ctx, teamID)
		if err != nil || team == nil {
			return nil, ErrTeamNotFound
		}
		member, err := s.teamRepo.IsMember(ctx, teamID, ownerID)
		if err != nil {
			return nil, fmt.Errorf("check team membership: %w", err)
		}
		if !member {
			return nil, ErrUserNotInTeam
		}
	}

	key = strings.ToUpper(strings.TrimSpace(key))
	if key == "" || !keyRegex.MatchString(key) {
		return nil, fmt.Errorf("%w: must be 1-50 chars, start with letter, alphanumeric uppercase (e.g. APP)", ErrInvalidKey)
	}
	exists, err := s.projectRepo.ExistsByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("check key: %w", err)
	}
	if exists {
		return nil, ErrKeyAlreadyExists
	}

	now := time.Now()
	p := &domain.Project{
		ID:             uuid.New(),
		Key:            key,
		Name:           strings.TrimSpace(name),
		Description:    strings.TrimSpace(description),
		TeamID:         teamID,
		OwnerID:        ownerID,
		NextTaskNumber: 1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.projectRepo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	// Владелец автоматически получает роль administrator в project_members
	if err := s.projectRepo.AddMember(ctx, p.ID, ownerID, domain.RoleAdministrator); err != nil {
		return nil, fmt.Errorf("add owner as member: %w", err)
	}
	// Дефолтные статусы (колонки) для канбан-доски
	for i, st := range defaultStatuses {
		ps := &domain.ProjectStatus{
			ID:        uuid.New(),
			ProjectID: p.ID,
			Key:       st.key,
			Title:     st.title,
			Order:     i,
		}
		if err := s.projectStatusRepo.Create(ctx, ps); err != nil {
			return nil, fmt.Errorf("create default status: %w", err)
		}
	}
	return p, nil
}

// hasProjectAccess проверяет доступ пользователя к проекту (владелец или член команды).
func (s *Service) hasProjectAccess(ctx context.Context, p *domain.Project, userID uuid.UUID) (bool, error) {
	if p.OwnerID == userID {
		return true, nil
	}
	if p.TeamID == uuid.Nil {
		return false, nil
	}
	return s.teamRepo.IsMember(ctx, p.TeamID, userID)
}

// GetProject возвращает проект по ID. Проверяет доступ (владелец или член команды).
func (s *Service) GetProject(ctx context.Context, projectID, userID uuid.UUID) (*domain.Project, error) {
	p, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil || p == nil {
		return nil, ErrProjectNotFound
	}
	access, err := s.hasProjectAccess(ctx, p, userID)
	if err != nil || !access {
		return nil, ErrForbidden
	}
	return p, nil
}

// ListProjects возвращает проекты, доступные пользователю (владелец или член команды).
func (s *Service) ListProjects(ctx context.Context, userID uuid.UUID) ([]*domain.Project, error) {
	return s.projectRepo.ListAccessibleByUser(ctx, userID)
}

// ListProjectMembers возвращает участников проекта с ролями. Доступ: владелец или член команды.
func (s *Service) ListProjectMembers(ctx context.Context, projectID, userID uuid.UUID) ([]*domain.MemberWithRole, error) {
	p, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil || p == nil {
		return nil, ErrProjectNotFound
	}
	access, err := s.hasProjectAccess(ctx, p, userID)
	if err != nil || !access {
		return nil, ErrForbidden
	}
	return s.projectRepo.ListMembers(ctx, projectID)
}

// ListProjectsByTeam возвращает проекты команды. Пользователь должен быть членом команды.
func (s *Service) ListProjectsByTeam(ctx context.Context, teamID, userID uuid.UUID) ([]*domain.Project, error) {
	t, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil || t == nil {
		return nil, ErrTeamNotFound
	}
	member, err := s.teamRepo.IsMember(ctx, teamID, userID)
	if err != nil || !member {
		return nil, ErrForbidden
	}
	return s.projectRepo.ListByTeamID(ctx, teamID)
}

// CreateTask создаёт задачу. key генерируется как PROJECTKEY-N. Доступ: владелец или член команды.
func (s *Service) CreateTask(ctx context.Context, projectID, reporterID uuid.UUID, req CreateTaskRequest) (*domain.Task, error) {
	p, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil || p == nil {
		return nil, ErrProjectNotFound
	}
	access, err := s.hasProjectAccess(ctx, p, reporterID)
	if err != nil || !access {
		return nil, ErrForbidden
	}

	firstStatus, err := s.projectStatusRepo.GetFirstByProject(ctx, projectID)
	if err != nil || firstStatus == nil {
		return nil, fmt.Errorf("project has no statuses: %w", ErrProjectNotFound)
	}

	nextNum, err := s.projectRepo.IncrementNextTaskNumber(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("increment counter: %w", err)
	}
	key := fmt.Sprintf("%s-%d", p.Key, nextNum)

	priority := domain.TaskPriorityMedium
	if req.Priority != "" {
		if !contains(domain.ValidTaskPriorities, req.Priority) {
			return nil, ErrInvalidPriority
		}
		priority = req.Priority
	}

	taskType := domain.TaskTypeTask
	if req.Type != "" {
		if !contains(domain.ValidTaskTypes, req.Type) {
			return nil, ErrInvalidType
		}
		taskType = req.Type
	}

	var assigneeID *uuid.UUID
	if req.AssigneeID != nil && *req.AssigneeID != "" {
		u, parseErr := uuid.Parse(*req.AssigneeID)
		if parseErr != nil {
			return nil, fmt.Errorf("%w: invalid assignee_id", ErrInvalidRequest)
		}
		assigneeID = &u
	}

	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		t, parseErr := time.Parse(time.RFC3339, *req.DueDate)
		if parseErr != nil {
			return nil, fmt.Errorf("%w: invalid due_date format (use RFC3339)", ErrInvalidRequest)
		}
		dueDate = &t
	}

	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	now := time.Now()
	t := &domain.Task{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Key:         key,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		StatusID:    firstStatus.ID,
		Status:      firstStatus.Key,
		Type:        taskType,
		Priority:    priority,
		AssigneeID:  assigneeID,
		ReporterID:  reporterID,
		DueDate:     dueDate,
		Tags:        tags,
		ResultURL:   strings.TrimSpace(req.ResultURL),
		Order:       0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.taskRepo.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	if s.taskPub != nil {
		s.taskPub.PublishTaskChanged(projectID, t.ID, "created")
		s.taskPub.PublishTaskStatusChanged(projectID, t.ID, t.Status, t.StatusID)
	}
	return t, nil
}

// GetTask возвращает задачу по ID. Проверяет доступ через проект.
func (s *Service) GetTask(ctx context.Context, taskID, userID uuid.UUID) (*domain.Task, error) {
	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil || t == nil {
		return nil, ErrTaskNotFound
	}
	p, err := s.projectRepo.GetByID(ctx, t.ProjectID)
	if err != nil || p == nil {
		return nil, ErrTaskNotFound
	}
	if p.OwnerID != userID {
		return nil, ErrForbidden
	}
	return t, nil
}

func validateTaskListFilter(filter TaskListFilter) error {
	if filter.Type != "" && !contains(domain.ValidTaskTypes, filter.Type) {
		return ErrInvalidType
	}
	if filter.DueFrom != nil && filter.DueTo != nil && filter.DueFrom.After(*filter.DueTo) {
		return fmt.Errorf("%w: due_from must be before or equal to due_to", ErrInvalidRequest)
	}
	return nil
}

// ListTasks возвращает задачи проекта. Проверяет доступ.
func (s *Service) ListTasks(ctx context.Context, projectID, userID uuid.UUID, filter TaskListFilter) ([]*domain.Task, error) {
	if _, err := s.GetProject(ctx, projectID, userID); err != nil {
		return nil, err
	}
	if err := validateTaskListFilter(filter); err != nil {
		return nil, err
	}
	return s.taskRepo.ListByProjectID(ctx, projectID, filter)
}

// UpdateTask частично обновляет задачу.
func (s *Service) UpdateTask(ctx context.Context, taskID, userID uuid.UUID, req UpdateTaskRequest) (*domain.Task, error) {
	t, err := s.GetTask(ctx, taskID, userID)
	if err != nil {
		return nil, err
	}
	prevStatusID := t.StatusID

	if req.Title != nil {
		t.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		t.Description = strings.TrimSpace(*req.Description)
	}
	if req.Status != nil {
		st, err := s.projectStatusRepo.GetByProjectAndKey(ctx, t.ProjectID, *req.Status)
		if err != nil || st == nil {
			return nil, ErrInvalidStatus
		}
		t.Status = st.Key
		t.StatusID = st.ID
	}
	if req.Priority != nil {
		if !contains(domain.ValidTaskPriorities, *req.Priority) {
			return nil, ErrInvalidPriority
		}
		t.Priority = *req.Priority
	}
	if req.AssigneeID != nil {
		if *req.AssigneeID == "" {
			t.AssigneeID = nil
		} else {
			u, parseErr := uuid.Parse(*req.AssigneeID)
			if parseErr != nil {
				return nil, fmt.Errorf("%w: invalid assignee_id", ErrInvalidRequest)
			}
			t.AssigneeID = &u
		}
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			t.DueDate = nil
		} else {
			parsed, parseErr := time.Parse(time.RFC3339, *req.DueDate)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid due_date: %w", parseErr)
			}
			t.DueDate = &parsed
		}
	}
	if req.Order != nil {
		t.Order = *req.Order
	}
	if req.Type != nil {
		if !contains(domain.ValidTaskTypes, *req.Type) {
			return nil, ErrInvalidType
		}
		t.Type = *req.Type
	}
	if req.Tags != nil {
		t.Tags = *req.Tags
		if t.Tags == nil {
			t.Tags = []string{}
		}
	}
	if req.ResultURL != nil {
		t.ResultURL = strings.TrimSpace(*req.ResultURL)
	}

	t.UpdatedAt = time.Now()
	if err := s.taskRepo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}
	if s.taskPub != nil {
		s.taskPub.PublishTaskChanged(t.ProjectID, t.ID, "updated")
		if prevStatusID != t.StatusID {
			s.taskPub.PublishTaskStatusChanged(t.ProjectID, t.ID, t.Status, t.StatusID)
		}
	}
	return t, nil
}

// DeleteTask удаляет задачу.
func (s *Service) DeleteTask(ctx context.Context, taskID, userID uuid.UUID) error {
	t, err := s.GetTask(ctx, taskID, userID)
	if err != nil {
		return err
	}
	projectID := t.ProjectID
	if err := s.taskRepo.Delete(ctx, taskID); err != nil {
		return err
	}
	if s.taskPub != nil {
		s.taskPub.PublishTaskChanged(projectID, taskID, "deleted")
	}
	return nil
}

// BoardColumn — колонка канбан-доски.
type BoardColumn struct {
	Status string         `json:"status"`
	Title  string         `json:"title"`
	Order  int            `json:"order"`
	Tasks  []*domain.Task `json:"tasks"`
}

// GetBoard возвращает канбан-доску для проекта. Колонки строятся из динамических статусов проекта.
// Фильтр такой же, как у ListTasks (assignee_id, title, type, due_from, due_to, status).
func (s *Service) GetBoard(ctx context.Context, projectID, userID uuid.UUID, filter TaskListFilter) ([]BoardColumn, error) {
	if _, err := s.GetProject(ctx, projectID, userID); err != nil {
		return nil, err
	}
	if err := validateTaskListFilter(filter); err != nil {
		return nil, err
	}
	statuses, err := s.projectStatusRepo.ListByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	grouped, err := s.taskRepo.ListByProjectIDGroupedByStatus(ctx, projectID, filter)
	if err != nil {
		return nil, err
	}

	columns := make([]BoardColumn, 0, len(statuses))
	for _, st := range statuses {
		tasks := grouped[st.Key]
		if tasks == nil {
			tasks = []*domain.Task{}
		}
		columns = append(columns, BoardColumn{
			Status: st.Key,
			Title:  st.Title,
			Order:  st.Order,
			Tasks:  tasks,
		})
	}
	return columns, nil
}

// ListStatuses возвращает статусы (колонки) проекта.
func (s *Service) ListStatuses(ctx context.Context, projectID, userID uuid.UUID) ([]*domain.ProjectStatus, error) {
	if _, err := s.GetProject(ctx, projectID, userID); err != nil {
		return nil, err
	}
	return s.projectStatusRepo.ListByProjectID(ctx, projectID)
}

// CreateStatus создаёт статус в проекте.
func (s *Service) CreateStatus(ctx context.Context, projectID, userID uuid.UUID, key, title string, order int) (*domain.ProjectStatus, error) {
	if _, err := s.GetProject(ctx, projectID, userID); err != nil {
		return nil, err
	}
	key = strings.ToUpper(strings.TrimSpace(key))
	if key == "" {
		return nil, ErrInvalidRequest
	}
	exists, err := s.projectStatusRepo.ExistsByKey(ctx, projectID, key)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrStatusKeyExists
	}
	ps := &domain.ProjectStatus{
		ID:        uuid.New(),
		ProjectID: projectID,
		Key:       key,
		Title:     strings.TrimSpace(title),
		Order:     order,
	}
	if ps.Title == "" {
		ps.Title = key
	}
	dupTitle, err := s.projectStatusRepo.ExistsByTitle(ctx, projectID, ps.Title, nil)
	if err != nil {
		return nil, err
	}
	if dupTitle {
		return nil, ErrStatusTitleExists
	}
	if err := s.projectStatusRepo.Create(ctx, ps); err != nil {
		return nil, fmt.Errorf("create status: %w", err)
	}
	return ps, nil
}

// UpdateStatus обновляет статус проекта.
func (s *Service) UpdateStatus(ctx context.Context, statusID, userID uuid.UUID, key, title string, order *int) (*domain.ProjectStatus, error) {
	ps, err := s.projectStatusRepo.GetByID(ctx, statusID)
	if err != nil || ps == nil {
		return nil, ErrStatusNotFound
	}
	if _, err := s.GetProject(ctx, ps.ProjectID, userID); err != nil {
		return nil, err
	}
	oldKey := ps.Key
	if key != "" {
		newKey := strings.ToUpper(strings.TrimSpace(key))
		if newKey != ps.Key {
			exists, err := s.projectStatusRepo.ExistsByKey(ctx, ps.ProjectID, newKey)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, ErrStatusKeyExists
			}
			ps.Key = newKey
		}
	}
	if title != "" {
		nt := strings.TrimSpace(title)
		if nt != ps.Title {
			dup, err := s.projectStatusRepo.ExistsByTitle(ctx, ps.ProjectID, nt, &ps.ID)
			if err != nil {
				return nil, err
			}
			if dup {
				return nil, ErrStatusTitleExists
			}
		}
		ps.Title = strings.TrimSpace(title)
		if ps.Title == "" {
			ps.Title = ps.Key
		}
	}
	if order != nil {
		ps.Order = *order
	}
	if err := s.projectStatusRepo.Update(ctx, ps); err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}
	if ps.Key != oldKey {
		if err := s.taskRepo.UpdateTaskStatusKeyByStatusID(ctx, ps.ID, ps.Key); err != nil {
			return nil, fmt.Errorf("sync task status key: %w", err)
		}
	}
	return ps, nil
}

// DeleteStatus удаляет статус и переносит задачи в колонку moveToID (тот же проект). Нельзя удалить последний статус.
func (s *Service) DeleteStatus(ctx context.Context, statusID, userID, moveToID uuid.UUID) error {
	ps, err := s.projectStatusRepo.GetByID(ctx, statusID)
	if err != nil || ps == nil {
		return ErrStatusNotFound
	}
	if _, err := s.GetProject(ctx, ps.ProjectID, userID); err != nil {
		return err
	}
	return s.projectStatusRepo.DeleteWithMove(ctx, statusID, moveToID)
}

func contains(slice []string, v string) bool {
	for _, s := range slice {
		if s == v {
			return true
		}
	}
	return false
}
