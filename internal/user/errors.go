package user

import "errors"

var (
	ErrCurrentTaskNotFound    = errors.New("task not found")
	ErrCurrentTaskNotAssignee = errors.New("task is not assigned to you")
)
