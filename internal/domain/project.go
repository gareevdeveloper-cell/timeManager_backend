package domain

import (
	"time"

	"github.com/google/uuid"
)

// Project — доменная модель проекта (см. DOMAIN.md).
// Проект принадлежит команде. В одной команде может быть несколько проектов.
type Project struct {
	ID              uuid.UUID
	Key             string    // уникальный короткий код (например, APP)
	Name            string
	Description     string
	TeamID          uuid.UUID // команда, к которой принадлежит проект
	OwnerID         uuid.UUID
	NextTaskNumber  int       // счётчик для генерации key задач (APP-1, APP-2, ...)
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TaskStatus — фиксированные статусы задач (колонки канбана).
const (
	TaskStatusTODO       = "TODO"
	TaskStatusInProgress = "IN_PROGRESS"
	TaskStatusInReview   = "IN_REVIEW"
	TaskStatusDone       = "DONE"
)

// TaskPriority — допустимые приоритеты задач.
const (
	TaskPriorityLow      = "LOW"
	TaskPriorityMedium   = "MEDIUM"
	TaskPriorityHigh     = "HIGH"
	TaskPriorityCritical = "CRITICAL"
)

// TaskType — допустимые типы задач.
const (
	TaskTypeBug   = "BUG"
	TaskTypeTask  = "TASK"
	TaskTypeStory = "STORY"
)

// ValidTaskStatuses — все допустимые статусы.
var ValidTaskStatuses = []string{
	TaskStatusTODO, TaskStatusInProgress, TaskStatusInReview, TaskStatusDone,
}

// ValidTaskPriorities — все допустимые приоритеты.
var ValidTaskPriorities = []string{
	TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh, TaskPriorityCritical,
}

// ValidTaskTypes — все допустимые типы задач.
var ValidTaskTypes = []string{
	TaskTypeBug, TaskTypeTask, TaskTypeStory,
}
