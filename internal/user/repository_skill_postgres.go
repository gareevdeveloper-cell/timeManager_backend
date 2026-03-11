package user

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresSkillRepository — реализация SkillRepository для PostgreSQL.
type PostgresSkillRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresSkillRepository создаёт репозиторий скиллов.
func NewPostgresSkillRepository(pool *pgxpool.Pool) *PostgresSkillRepository {
	return &PostgresSkillRepository{pool: pool}
}

// GetByUserID возвращает названия скиллов пользователя.
func (r *PostgresSkillRepository) GetByUserID(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT s.name FROM user_skills us
		JOIN skills s ON s.id = us.skill_id
		WHERE us.user_id = $1
		ORDER BY s.name
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

// SetForUser заменяет скиллы пользователя на переданный список.
func (r *PostgresSkillRepository) SetForUser(ctx context.Context, userID string, skillNames []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM user_skills WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	for _, name := range skillNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		skillID, err := r.getOrCreate(ctx, tx, name)
		if err != nil {
			return err
		}
		if skillID == "" {
			continue
		}
		_, err = tx.Exec(ctx, `INSERT INTO user_skills (user_id, skill_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, skillID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresSkillRepository) getOrCreate(ctx context.Context, tx pgx.Tx, name string) (string, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM skills WHERE name = $1`, name).Scan(&id)
	if err == nil {
		return id.String(), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	id = uuid.New()
	_, err = tx.Exec(ctx, `INSERT INTO skills (id, name) VALUES ($1, $2) ON CONFLICT (name) DO NOTHING`, id, name)
	if err != nil {
		return "", err
	}

	var finalID uuid.UUID
	err = tx.QueryRow(ctx, `SELECT id FROM skills WHERE name = $1`, name).Scan(&finalID)
	if err != nil {
		return "", err
	}
	return finalID.String(), nil
}
