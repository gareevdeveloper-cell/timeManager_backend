package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"testik/internal/domain"
)

// PostgresUserRepository — реализация UserRepository для PostgreSQL.
type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresUserRepository создаёт репозиторий.
func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// GetByID возвращает пользователя по ID.
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, COALESCE(password_hash, ''), COALESCE(oauth_provider, ''), COALESCE(oauth_provider_id, ''),
			firstname, lastname, middlename, birthday, COALESCE(about, ''), COALESCE(position, ''),
			role, status, COALESCE(work_status::text, 'working'), COALESCE(avatar_url, ''), current_task_id, created_at, updated_at
		FROM users WHERE id = $1
	`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.OAuthProvider, &u.OAuthProviderID,
		&u.FirstName, &u.LastName, &u.MiddleName, &u.Birthday, &u.About, &u.Position,
		&u.Role, &u.Status, &u.WorkStatus, &u.AvatarURL, &u.CurrentTaskID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateProfile обновляет профиль пользователя (имя, фамилия, день рождения, about, position).
func (r *PostgresUserRepository) UpdateProfile(ctx context.Context, userID string, firstname, lastname string, birthday *time.Time, about, position string) error {
	query := `UPDATE users SET firstname = $2, lastname = $3, birthday = $4, about = $5, position = $6, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID, nullIfEmpty(firstname), nullIfEmpty(lastname), birthday, nullIfEmpty(about), nullIfEmpty(position))
	return err
}

// UpdateAvatarURL обновляет URL аватарки пользователя.
func (r *PostgresUserRepository) UpdateAvatarURL(ctx context.Context, userID string, avatarURL string) error {
	query := `UPDATE users SET avatar_url = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID, nullIfEmpty(avatarURL))
	return err
}

// UpdateWorkStatus обновляет рабочий статус пользователя.
func (r *PostgresUserRepository) UpdateWorkStatus(ctx context.Context, userID string, workStatus string) error {
	query := `UPDATE users SET work_status = $2::work_status_enum, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID, workStatus)
	return err
}

// UpdateCurrentTaskID задаёт или сбрасывает текущую задачу «в работе».
func (r *PostgresUserRepository) UpdateCurrentTaskID(ctx context.Context, userID string, taskID *uuid.UUID) error {
	query := `UPDATE users SET current_task_id = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID, taskID)
	return err
}
