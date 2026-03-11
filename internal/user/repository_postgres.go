package user

import (
	"context"

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
			role, status, COALESCE(work_status::text, 'working'), COALESCE(avatar_url, ''), created_at, updated_at
		FROM users WHERE id = $1
	`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.OAuthProvider, &u.OAuthProviderID,
		&u.FirstName, &u.LastName, &u.MiddleName, &u.Birthday, &u.About, &u.Position,
		&u.Role, &u.Status, &u.WorkStatus, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateProfile обновляет профиль пользователя (about, position).
func (r *PostgresUserRepository) UpdateProfile(ctx context.Context, userID string, about, position string) error {
	query := `UPDATE users SET about = $2, position = $3, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID, nullIfEmpty(about), nullIfEmpty(position))
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
