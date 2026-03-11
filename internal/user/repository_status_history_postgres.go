package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresWorkStatusHistoryRepository — реализация WorkStatusHistoryRepository.
type PostgresWorkStatusHistoryRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresWorkStatusHistoryRepository создаёт репозиторий.
func NewPostgresWorkStatusHistoryRepository(pool *pgxpool.Pool) *PostgresWorkStatusHistoryRepository {
	return &PostgresWorkStatusHistoryRepository{pool: pool}
}

// Create добавляет запись в историю статусов.
func (r *PostgresWorkStatusHistoryRepository) Create(ctx context.Context, userID, workStatus, changedBy string) error {
	query := `
		INSERT INTO user_work_status_history (id, user_id, work_status, changed_by)
		VALUES ($1, $2, $3::work_status_enum, $4)
	`
	id := uuid.New()
	var changedByArg interface{}
	if changedBy != "" {
		changedByArg = changedBy
	} else {
		changedByArg = nil
	}
	_, err := r.pool.Exec(ctx, query, id, userID, workStatus, changedByArg)
	return err
}

// GetByUserID возвращает историю статусов пользователя.
func (r *PostgresWorkStatusHistoryRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]WorkStatusHistoryEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, user_id, work_status::text, changed_at::text, changed_by::text
		FROM user_work_status_history
		WHERE user_id = $1
		ORDER BY changed_at DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []WorkStatusHistoryEntry
	for rows.Next() {
		var e WorkStatusHistoryEntry
		var changedBy *string
		if err := rows.Scan(&e.ID, &e.UserID, &e.WorkStatus, &e.ChangedAt, &changedBy); err != nil {
			return nil, err
		}
		if changedBy != nil {
			e.ChangedBy = *changedBy
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
