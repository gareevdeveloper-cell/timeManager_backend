package domain

import (
	"github.com/google/uuid"
)

// ProjectStatus — статус (колонка) канбан-доски проекта.
// Каждый проект имеет свой набор статусов.
type ProjectStatus struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Key       string // уникальный код в рамках проекта (например TODO, IN_PROGRESS)
	Title     string // отображаемое название (например "To Do", "In Progress")
	Order     int    // порядок колонки на доске
}
