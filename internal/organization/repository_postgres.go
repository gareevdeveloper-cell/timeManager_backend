package organization

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

// Create сохраняет организацию.
func (r *PostgresRepository) Create(ctx context.Context, o *domain.Organization) error {
	query := `
		INSERT INTO organizations (id, name, slug, owner_id, status, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query, o.ID, o.Name, o.Slug, o.OwnerID, o.Status, nullIfEmpty(o.AvatarURL), o.CreatedAt, o.UpdatedAt)
	return err
}

// ExistsBySlug проверяет существование организации с данным slug.
func (r *PostgresRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM organizations WHERE slug = $1)`, slug).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// IsUniqueViolation проверяет, является ли ошибка нарушением уникальности.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// GetByID возвращает организацию по ID.
func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	query := `
		SELECT id, name, slug, owner_id, status, COALESCE(avatar_url, ''), created_at, updated_at
		FROM organizations WHERE id = $1
	`
	var o domain.Organization
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.Name, &o.Slug, &o.OwnerID, &o.Status, &o.AvatarURL, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

// Update обновляет организацию (name, avatar_url).
func (r *PostgresRepository) Update(ctx context.Context, o *domain.Organization) error {
	query := `
		UPDATE organizations SET name = $2, avatar_url = $4, updated_at = $3 WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, o.ID, o.Name, o.UpdatedAt, nullIfEmpty(o.AvatarURL))
	return err
}

// UpdateAvatarURL обновляет только URL аватарки.
func (r *PostgresRepository) UpdateAvatarURL(ctx context.Context, orgID uuid.UUID, avatarURL string) error {
	query := `UPDATE organizations SET avatar_url = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, orgID, nullIfEmpty(avatarURL))
	return err
}

// UpdateStatus обновляет статус организации.
func (r *PostgresRepository) UpdateStatus(ctx context.Context, o *domain.Organization) error {
	query := `
		UPDATE organizations SET status = $2, updated_at = $3 WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, o.ID, o.Status, o.UpdatedAt)
	return err
}

// AddMember добавляет пользователя в организацию с указанной ролью.
func (r *PostgresRepository) AddMember(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	query := `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, $3)
	`
	_, err := r.pool.Exec(ctx, query, orgID, userID, role)
	return err
}

// RemoveMember удаляет пользователя из организации.
func (r *PostgresRepository) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	query := `DELETE FROM organization_members WHERE organization_id = $1 AND user_id = $2`
	result, err := r.pool.Exec(ctx, query, orgID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return nil // уже не член
	}
	return nil
}

// IsMember проверяет, является ли пользователь членом организации.
func (r *PostgresRepository) IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM organization_members WHERE organization_id = $1 AND user_id = $2)`,
		orgID, userID,
	).Scan(&exists)
	return exists, err
}

// GetMemberOrganization возвращает ID организации, в которой состоит пользователь, или nil.
func (r *PostgresRepository) GetMemberOrganization(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.pool.QueryRow(ctx,
		`SELECT organization_id FROM organization_members WHERE user_id = $1`,
		userID,
	).Scan(&orgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &orgID, nil
}

// ListByMember возвращает организации, в которых состоит пользователь.
func (r *PostgresRepository) ListByMember(ctx context.Context, userID uuid.UUID) ([]*domain.Organization, error) {
	query := `
		SELECT o.id, o.name, o.slug, o.owner_id, o.status, COALESCE(o.avatar_url, ''), o.created_at, o.updated_at
		FROM organizations o
		INNER JOIN organization_members om ON o.id = om.organization_id
		WHERE om.user_id = $1
		ORDER BY o.created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*domain.Organization
	for rows.Next() {
		var o domain.Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.Slug, &o.OwnerID, &o.Status, &o.AvatarURL, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, &o)
	}
	return orgs, rows.Err()
}

// ListMembers возвращает участников организации с ролями и текущей задачей (LEFT JOIN tasks).
func (r *PostgresRepository) ListMembers(ctx context.Context, orgID uuid.UUID) ([]*domain.MemberWithRole, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.firstname, u.lastname, u.middlename, u.birthday, u.role, u.status, COALESCE(u.work_status::text, 'working'), COALESCE(u.avatar_url, ''), u.created_at, u.updated_at, COALESCE(om.role, 'participant'),
		       ct.id, ct.title, ct.project_id
		FROM users u
		INNER JOIN organization_members om ON om.user_id = u.id
		LEFT JOIN tasks ct ON ct.id = u.current_task_id
		WHERE om.organization_id = $1
		ORDER BY u.firstname, u.lastname
	`
	rows, err := r.pool.Query(ctx, query, orgID)
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
