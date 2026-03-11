package project

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"

	"testik/internal/domain"
)

// PostgresProjectRepository — реализация ProjectRepository для PostgreSQL.
type PostgresProjectRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresProjectRepository создаёт репозиторий проектов.
func NewPostgresProjectRepository(pool *pgxpool.Pool) *PostgresProjectRepository {
	return &PostgresProjectRepository{pool: pool}
}

// Create сохраняет проект.
func (r *PostgresProjectRepository) Create(ctx context.Context, p *domain.Project) error {
	query := `
		INSERT INTO projects (id, key, name, description, team_id, owner_id, next_task_number, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		p.ID, p.Key, p.Name, nullIfEmpty(p.Description), nullUUIDIfNil(p.TeamID), p.OwnerID,
		p.NextTaskNumber, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

// GetByID возвращает проект по ID.
func (r *PostgresProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	query := `
		SELECT id, key, name, COALESCE(description, ''), team_id, owner_id, next_task_number, created_at, updated_at
		FROM projects WHERE id = $1
	`
	var p domain.Project
	var teamID *uuid.UUID
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Key, &p.Name, &p.Description, &teamID, &p.OwnerID,
		&p.NextTaskNumber, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if teamID != nil {
		p.TeamID = *teamID
	}
	return &p, nil
}

// GetByKey возвращает проект по key.
func (r *PostgresProjectRepository) GetByKey(ctx context.Context, key string) (*domain.Project, error) {
	query := `
		SELECT id, key, name, COALESCE(description, ''), team_id, owner_id, next_task_number, created_at, updated_at
		FROM projects WHERE key = $1
	`
	var p domain.Project
	var teamID *uuid.UUID
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&p.ID, &p.Key, &p.Name, &p.Description, &teamID, &p.OwnerID,
		&p.NextTaskNumber, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if teamID != nil {
		p.TeamID = *teamID
	}
	return &p, nil
}

// ExistsByKey проверяет существование проекта с данным key.
func (r *PostgresProjectRepository) ExistsByKey(ctx context.Context, key string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM projects WHERE key = $1)`, key).Scan(&exists)
	return exists, err
}

// ListByOwner возвращает проекты владельца.
func (r *PostgresProjectRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Project, error) {
	query := `
		SELECT id, key, name, COALESCE(description, ''), team_id, owner_id, next_task_number, created_at, updated_at
		FROM projects WHERE owner_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Project
	for rows.Next() {
		var p domain.Project
		var teamID *uuid.UUID
		if err := rows.Scan(&p.ID, &p.Key, &p.Name, &p.Description, &teamID, &p.OwnerID, &p.NextTaskNumber, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if teamID != nil {
			p.TeamID = *teamID
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

// ListByTeamID возвращает проекты команды.
func (r *PostgresProjectRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]*domain.Project, error) {
	query := `
		SELECT id, key, name, COALESCE(description, ''), team_id, owner_id, next_task_number, created_at, updated_at
		FROM projects WHERE team_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Project
	for rows.Next() {
		var p domain.Project
		var tid *uuid.UUID
		if err := rows.Scan(&p.ID, &p.Key, &p.Name, &p.Description, &tid, &p.OwnerID, &p.NextTaskNumber, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if tid != nil {
			p.TeamID = *tid
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

// ListAccessibleByUser возвращает проекты, доступные пользователю (владелец или член команды).
func (r *PostgresProjectRepository) ListAccessibleByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Project, error) {
	query := `
		SELECT DISTINCT p.id, p.key, p.name, COALESCE(p.description, ''), p.team_id, p.owner_id, p.next_task_number, p.created_at, p.updated_at
		FROM projects p
		LEFT JOIN team_members tm ON p.team_id = tm.team_id AND tm.user_id = $1
		WHERE p.owner_id = $1 OR (p.team_id IS NOT NULL AND tm.user_id = $1)
		ORDER BY p.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Project
	for rows.Next() {
		var p domain.Project
		var tid *uuid.UUID
		if err := rows.Scan(&p.ID, &p.Key, &p.Name, &p.Description, &tid, &p.OwnerID, &p.NextTaskNumber, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if tid != nil {
			p.TeamID = *tid
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

// AddMember добавляет участника проекта с указанной ролью.
func (r *PostgresProjectRepository) AddMember(ctx context.Context, projectID, userID uuid.UUID, role string) error {
	query := `INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT (project_id, user_id) DO UPDATE SET role = $3`
	_, err := r.pool.Exec(ctx, query, projectID, userID, role)
	return err
}

// ListMembers возвращает участников проекта с ролями.
func (r *PostgresProjectRepository) ListMembers(ctx context.Context, projectID uuid.UUID) ([]*domain.MemberWithRole, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.firstname, u.lastname, u.middlename, u.birthday, u.role, u.status, COALESCE(u.avatar_url, ''), u.created_at, u.updated_at, COALESCE(pm.role, 'participant')
		FROM users u
		INNER JOIN project_members pm ON pm.user_id = u.id
		WHERE pm.project_id = $1
		ORDER BY u.firstname, u.lastname
	`
	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.MemberWithRole
	for rows.Next() {
		var u domain.User
		var memberRole string
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.MiddleName,
			&u.Birthday, &u.Role, &u.Status, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt, &memberRole); err != nil {
			return nil, err
		}
		members = append(members, &domain.MemberWithRole{User: &u, Role: memberRole})
	}
	return members, rows.Err()
}

// IncrementNextTaskNumber атомарно инкрементирует счётчик и возвращает новое значение.
func (r *PostgresProjectRepository) IncrementNextTaskNumber(ctx context.Context, projectID uuid.UUID) (int, error) {
	query := `
		UPDATE projects SET next_task_number = next_task_number + 1, updated_at = NOW()
		WHERE id = $1
		RETURNING next_task_number
	`
	var next int
	err := r.pool.QueryRow(ctx, query, projectID).Scan(&next)
	return next, err
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullUUIDIfNil(id uuid.UUID) interface{} {
	if id == uuid.Nil {
		return nil
	}
	return id
}

// PostgresProjectStatusRepository — реализация ProjectStatusRepository для PostgreSQL.
type PostgresProjectStatusRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresProjectStatusRepository создаёт репозиторий статусов проектов.
func NewPostgresProjectStatusRepository(pool *pgxpool.Pool) *PostgresProjectStatusRepository {
	return &PostgresProjectStatusRepository{pool: pool}
}

func (r *PostgresProjectStatusRepository) Create(ctx context.Context, s *domain.ProjectStatus) error {
	query := `INSERT INTO project_statuses (id, project_id, key, title, "order") VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pool.Exec(ctx, query, s.ID, s.ProjectID, s.Key, s.Title, s.Order)
	return err
}

func (r *PostgresProjectStatusRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProjectStatus, error) {
	query := `SELECT id, project_id, key, title, "order" FROM project_statuses WHERE id = $1`
	var s domain.ProjectStatus
	err := r.pool.QueryRow(ctx, query, id).Scan(&s.ID, &s.ProjectID, &s.Key, &s.Title, &s.Order)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *PostgresProjectStatusRepository) ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]*domain.ProjectStatus, error) {
	query := `SELECT id, project_id, key, title, "order" FROM project_statuses WHERE project_id = $1 ORDER BY "order" ASC, key ASC`
	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.ProjectStatus
	for rows.Next() {
		var s domain.ProjectStatus
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Key, &s.Title, &s.Order); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *PostgresProjectStatusRepository) Update(ctx context.Context, s *domain.ProjectStatus) error {
	query := `UPDATE project_statuses SET key = $2, title = $3, "order" = $4 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, s.ID, s.Key, s.Title, s.Order)
	return err
}

func (r *PostgresProjectStatusRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM project_statuses WHERE id = $1`, id)
	return err
}

func (r *PostgresProjectStatusRepository) ExistsByKey(ctx context.Context, projectID uuid.UUID, key string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM project_statuses WHERE project_id = $1 AND key = $2)`,
		projectID, key,
	).Scan(&exists)
	return exists, err
}

func (r *PostgresProjectStatusRepository) GetFirstByProject(ctx context.Context, projectID uuid.UUID) (*domain.ProjectStatus, error) {
	query := `SELECT id, project_id, key, title, "order" FROM project_statuses WHERE project_id = $1 ORDER BY "order" ASC, key ASC LIMIT 1`
	var s domain.ProjectStatus
	err := r.pool.QueryRow(ctx, query, projectID).Scan(&s.ID, &s.ProjectID, &s.Key, &s.Title, &s.Order)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// PostgresTaskRepository — реализация TaskRepository для PostgreSQL.
type PostgresTaskRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresTaskRepository создаёт репозиторий задач.
func NewPostgresTaskRepository(pool *pgxpool.Pool) *PostgresTaskRepository {
	return &PostgresTaskRepository{pool: pool}
}

// Create сохраняет задачу.
func (r *PostgresTaskRepository) Create(ctx context.Context, t *domain.Task) error {
	query := `
		INSERT INTO tasks (id, project_id, key, title, description, status, priority, assignee_id, reporter_id, due_date, "order", created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.pool.Exec(ctx, query,
		t.ID, t.ProjectID, t.Key, t.Title, nullIfEmpty(t.Description),
		t.Status, t.Priority, t.AssigneeID, t.ReporterID, t.DueDate,
		t.Order, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

// GetByID возвращает задачу по ID.
func (r *PostgresTaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	query := `
		SELECT id, project_id, key, title, COALESCE(description, ''), status, priority,
		       assignee_id, reporter_id, due_date, "order", created_at, updated_at
		FROM tasks WHERE id = $1
	`
	var t domain.Task
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.ProjectID, &t.Key, &t.Title, &t.Description,
		&t.Status, &t.Priority, &t.AssigneeID, &t.ReporterID, &t.DueDate,
		&t.Order, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// GetByKey возвращает задачу по key.
func (r *PostgresTaskRepository) GetByKey(ctx context.Context, key string) (*domain.Task, error) {
	query := `
		SELECT id, project_id, key, title, COALESCE(description, ''), status, priority,
		       assignee_id, reporter_id, due_date, "order", created_at, updated_at
		FROM tasks WHERE key = $1
	`
	var t domain.Task
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&t.ID, &t.ProjectID, &t.Key, &t.Title, &t.Description,
		&t.Status, &t.Priority, &t.AssigneeID, &t.ReporterID, &t.DueDate,
		&t.Order, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// ListByProjectID возвращает задачи проекта, опционально фильтруя по статусу.
func (r *PostgresTaskRepository) ListByProjectID(ctx context.Context, projectID uuid.UUID, status string) ([]*domain.Task, error) {
	var rows pgx.Rows
	var err error
	if status != "" {
		rows, err = r.pool.Query(ctx, `
			SELECT id, project_id, key, title, COALESCE(description, ''), status, priority,
			       assignee_id, reporter_id, due_date, "order", created_at, updated_at
			FROM tasks WHERE project_id = $1 AND status = $2
			ORDER BY "order" ASC, created_at ASC
		`, projectID, status)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, project_id, key, title, COALESCE(description, ''), status, priority,
			       assignee_id, reporter_id, due_date, "order", created_at, updated_at
			FROM tasks WHERE project_id = $1
			ORDER BY status ASC, "order" ASC, created_at ASC
		`, projectID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Key, &t.Title, &t.Description,
			&t.Status, &t.Priority, &t.AssigneeID, &t.ReporterID, &t.DueDate,
			&t.Order, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, rows.Err()
}

// ListByProjectIDGroupedByStatus возвращает задачи проекта, сгруппированные по статусу.
func (r *PostgresTaskRepository) ListByProjectIDGroupedByStatus(ctx context.Context, projectID uuid.UUID) (map[string][]*domain.Task, error) {
	tasks, err := r.ListByProjectID(ctx, projectID, "")
	if err != nil {
		return nil, err
	}
	result := map[string][]*domain.Task{
		domain.TaskStatusTODO:       nil,
		domain.TaskStatusInProgress: nil,
		domain.TaskStatusInReview:   nil,
		domain.TaskStatusDone:      nil,
	}
	for _, t := range tasks {
		result[t.Status] = append(result[t.Status], t)
	}
	return result, nil
}

// Update обновляет задачу.
func (r *PostgresTaskRepository) Update(ctx context.Context, t *domain.Task) error {
	query := `
		UPDATE tasks SET
			title = $2, description = $3, status = $4, priority = $5,
			assignee_id = $6, due_date = $7, "order" = $8, updated_at = $9
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		t.ID, t.Title, nullIfEmpty(t.Description), t.Status, t.Priority,
		t.AssigneeID, t.DueDate, t.Order, t.UpdatedAt,
	)
	return err
}

// Delete удаляет задачу.
func (r *PostgresTaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	return err
}
