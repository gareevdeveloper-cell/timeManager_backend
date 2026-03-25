package domain

import (
	"time"

	"github.com/google/uuid"
)

// Task — доменная модель задачи (issue) в проекте.
type Task struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	Key         string     // например APP-1, APP-2 (уникален в системе)
	Title       string
	Description string
	StatusID    uuid.UUID  // FK project_statuses — колонка канбана
	Status      string     // ключ статуса (дублирует project_statuses.key для API/фильтров)
	Type        string     // BUG, TASK, STORY
	Priority    string     // LOW, MEDIUM, HIGH, CRITICAL
	AssigneeID  *uuid.UUID // nullable
	ReporterID  uuid.UUID  // author/reporter — создатель задачи
	DueDate     *time.Time // nullable
	Tags        []string   // метки/теги
	ResultURL   string     // ссылка на результат (может быть пустой)
	Order       int        // позиция в колонке (для сортировки)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
