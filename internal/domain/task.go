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
	Status      string     // TODO, IN_PROGRESS, IN_REVIEW, DONE
	Priority    string     // LOW, MEDIUM, HIGH, CRITICAL
	AssigneeID  *uuid.UUID // nullable
	ReporterID  uuid.UUID
	DueDate     *time.Time // nullable
	Order       int        // позиция в колонке (для сортировки)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
