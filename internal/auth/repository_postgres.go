package auth

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

// Create сохраняет пользователя.
func (r *PostgresUserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, oauth_provider, oauth_provider_id, firstname, lastname, middlename, birthday, about, position, role, status, work_status, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, COALESCE($14::work_status_enum, 'working'), $15, $16, $17)
	`
	var birthday interface{}
	if u.Birthday != nil {
		birthday = u.Birthday
	} else {
		birthday = nil
	}
	workStatus := nullIfEmpty(u.WorkStatus)
	if workStatus == nil {
		workStatus = "working"
	}
	_, err := r.pool.Exec(ctx, query,
		u.ID, u.Email, nullIfEmpty(u.PasswordHash), nullIfEmpty(u.OAuthProvider), nullIfEmpty(u.OAuthProviderID),
		u.FirstName, u.LastName, u.MiddleName, birthday, nullIfEmpty(u.About), nullIfEmpty(u.Position),
		u.Role, u.Status, workStatus, nullIfEmpty(u.AvatarURL), u.CreatedAt, u.UpdatedAt,
	)
	return err
}

// GetByEmail возвращает пользователя по email.
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, COALESCE(password_hash, ''), COALESCE(oauth_provider, ''), COALESCE(oauth_provider_id, ''),
			firstname, lastname, middlename, birthday, COALESCE(about, ''), COALESCE(position, ''),
			role, status, COALESCE(work_status::text, 'working'), COALESCE(avatar_url, ''), current_task_id, created_at, updated_at
		FROM users WHERE email = $1
	`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.OAuthProvider, &u.OAuthProviderID,
		&u.FirstName, &u.LastName, &u.MiddleName, &u.Birthday, &u.About, &u.Position,
		&u.Role, &u.Status, &u.WorkStatus, &u.AvatarURL, &u.CurrentTaskID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
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

// GetByOAuthID возвращает пользователя по OAuth-провайдеру и внешнему ID.
func (r *PostgresUserRepository) GetByOAuthID(ctx context.Context, provider, providerID string) (*domain.User, error) {
	query := `
		SELECT id, email, COALESCE(password_hash, ''), COALESCE(oauth_provider, ''), COALESCE(oauth_provider_id, ''),
			firstname, lastname, middlename, birthday, COALESCE(about, ''), COALESCE(position, ''),
			role, status, COALESCE(work_status::text, 'working'), COALESCE(avatar_url, ''), current_task_id, created_at, updated_at
		FROM users WHERE oauth_provider = $1 AND oauth_provider_id = $2
	`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, provider, providerID).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.OAuthProvider, &u.OAuthProviderID,
		&u.FirstName, &u.LastName, &u.MiddleName, &u.Birthday, &u.About, &u.Position,
		&u.Role, &u.Status, &u.WorkStatus, &u.AvatarURL, &u.CurrentTaskID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

