package ws

import (
	"encoding/json"

	"github.com/google/uuid"
)

// UserWorkStatusChanged — событие смены статуса пользователя.
const ChannelUserStatus = "user:status"

// UserWorkStatusPayload — payload события смены статуса.
type UserWorkStatusPayload struct {
	Type    string `json:"type"`
	UserID  string `json:"user_id"`
	WorkStatus string `json:"work_status"`
}

// EncodeUserWorkStatusChanged кодирует событие смены статуса.
func EncodeUserWorkStatusChanged(userID, workStatus string) ([]byte, error) {
	return json.Marshal(UserWorkStatusPayload{
		Type:        "user.work_status_changed",
		UserID:      userID,
		WorkStatus:  workStatus,
	})
}

// События задач проекта (каналы {projectId}:taskChange и {projectId}:taskChangeStatus).

const (
	MsgTypeProjectTaskChange        = "project.task_change"
	MsgTypeProjectTaskStatusChanged = "project.task_status_changed"
)

// ProjectTaskChangePayload — задача создана, обновлена (поля) или удалена.
type ProjectTaskChangePayload struct {
	Type      string `json:"type"`
	ProjectID string `json:"project_id"`
	TaskID    string `json:"task_id"`
	Action    string `json:"action"` // created | updated | deleted
}

// ProjectTaskStatusChangePayload — смена колонки/статуса задачи.
type ProjectTaskStatusChangePayload struct {
	Type      string `json:"type"`
	ProjectID string `json:"project_id"`
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	StatusID  string `json:"status_id"`
}

// EncodeProjectTaskChange кодирует событие project.task_change.
func EncodeProjectTaskChange(projectID, taskID uuid.UUID, action string) ([]byte, error) {
	return json.Marshal(ProjectTaskChangePayload{
		Type:      MsgTypeProjectTaskChange,
		ProjectID: projectID.String(),
		TaskID:    taskID.String(),
		Action:    action,
	})
}

// EncodeProjectTaskStatusChange кодирует событие project.task_status_changed.
func EncodeProjectTaskStatusChange(projectID, taskID uuid.UUID, statusKey string, statusID uuid.UUID) ([]byte, error) {
	return json.Marshal(ProjectTaskStatusChangePayload{
		Type:      MsgTypeProjectTaskStatusChanged,
		ProjectID: projectID.String(),
		TaskID:    taskID.String(),
		Status:    statusKey,
		StatusID:  statusID.String(),
	})
}

// ChannelProjectTaskChange — канал «изменение данных задачи».
func ChannelProjectTaskChange(projectID uuid.UUID) string {
	return projectID.String() + ":taskChange"
}

// ChannelProjectTaskChangeStatus — канал «смена статуса (колонки) задачи».
func ChannelProjectTaskChangeStatus(projectID uuid.UUID) string {
	return projectID.String() + ":taskChangeStatus"
}
