package team

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"

	"testik/internal/domain"
)

// PostgresRepository — реализация Repository для PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository создаёт репозиторий.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// Create сохраняет команду.
func (r *PostgresRepository) Create(ctx context.Context, t *domain.Team) error {
	query := `
		INSERT INTO teams (id, name, description, organization_id, creator_id, avatar_url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query, t.ID, t.Name, t.Description, t.OrganizationID, t.CreatorID, nullIfEmpty(t.AvatarURL), t.CreatedAt)
	return err
}

// GetByID возвращает команду по ID.
func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Team, error) {
	query := `
		SELECT id, name, description, organization_id, creator_id, COALESCE(avatar_url, ''), created_at
		FROM teams WHERE id = $1
	`
	var t domain.Team
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.Description, &t.OrganizationID, &t.CreatorID, &t.AvatarURL, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// GetByOrganization возвращает команды организации.
func (r *PostgresRepository) GetByOrganization(ctx context.Context, orgID uuid.UUID) ([]*domain.Team, error) {
	query := `
		SELECT id, name, description, organization_id, creator_id, COALESCE(avatar_url, ''), created_at
		FROM teams WHERE organization_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*domain.Team
	for rows.Next() {
		var t domain.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.OrganizationID, &t.CreatorID, &t.AvatarURL, &t.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, &t)
	}
	return teams, rows.Err()
}

// Update обновляет команду.
func (r *PostgresRepository) Update(ctx context.Context, t *domain.Team) error {
	query := `UPDATE teams SET name = $2, description = $3, avatar_url = $4 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, t.ID, t.Name, t.Description, nullIfEmpty(t.AvatarURL))
	return err
}

// UpdateAvatarURL обновляет только URL аватарки команды.
func (r *PostgresRepository) UpdateAvatarURL(ctx context.Context, teamID uuid.UUID, avatarURL string) error {
	query := `UPDATE teams SET avatar_url = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, teamID, nullIfEmpty(avatarURL))
	return err
}

// Delete удаляет команду.
func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM teams WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// AddMember добавляет пользователя в команду с указанной ролью.
func (r *PostgresRepository) AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error {
	query := `INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)`
	_, err := r.pool.Exec(ctx, query, teamID, userID, role)
	return err
}

// RemoveMember удаляет пользователя из команды.
func (r *PostgresRepository) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	query := `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, teamID, userID)
	return err
}

// IsMember проверяет, является ли пользователь членом команды.
func (r *PostgresRepository) IsMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2)`,
		teamID, userID,
	).Scan(&exists)
	return exists, err
}

// GetMembers возвращает участников команды с ролями и текущей задачей (LEFT JOIN tasks).
func (r *PostgresRepository) GetMembers(ctx context.Context, teamID uuid.UUID) ([]*domain.MemberWithRole, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.firstname, u.lastname, u.middlename, u.birthday, u.role, u.status, COALESCE(u.work_status::text, 'working'), COALESCE(u.avatar_url, ''), u.created_at, u.updated_at, COALESCE(tm.role, 'participant'),
		       ct.id, ct.title, ct.project_id
		FROM users u
		INNER JOIN team_members tm ON tm.user_id = u.id
		LEFT JOIN tasks ct ON ct.id = u.current_task_id
		WHERE tm.team_id = $1
		ORDER BY u.firstname, u.lastname
	`
	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.MemberWithRole
	for rows.Next() {
		var u domain.User
		var memberRole string
		var ctID, ctProjID *uuid.UUID
		var ctTitle sql.NullString
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.MiddleName,
			&u.Birthday, &u.Role, &u.Status, &u.WorkStatus, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt, &memberRole,
			&ctID, &ctTitle, &ctProjID); err != nil {
			return nil, err
		}
		var ct *domain.MemberCurrentTask
		if ctID != nil && ctProjID != nil {
			title := ""
			if ctTitle.Valid {
				title = ctTitle.String
			}
			ct = &domain.MemberCurrentTask{ID: *ctID, Title: title, ProjectID: *ctProjID}
		}
		members = append(members, &domain.MemberWithRole{User: &u, Role: memberRole, CurrentTask: ct})
	}
	return members, rows.Err()
}
