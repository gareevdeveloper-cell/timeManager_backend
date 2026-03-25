package user

import "time"

// UpdateProfileRequest — тело запроса обновления профиля. Все поля опциональны.
type UpdateProfileRequest struct {
	About     *string  `json:"about,omitempty"`
	Position  *string  `json:"position,omitempty"`
	Skills    []string `json:"skills,omitempty"`
	FirstName *string  `json:"firstname,omitempty"`
	LastName  *string  `json:"lastname,omitempty"`
	Birthday  *string  `json:"birthday,omitempty"` // YYYY-MM-DD или RFC3339; пустая строка — сброс даты рождения
}

// UserResponse — ответ с данными пользователя (для swagger).
type UserResponse struct {
	Data struct {
		ID         string    `json:"id"`
		Email      string    `json:"email"`
		FirstName  string    `json:"firstname"`
		LastName   string    `json:"lastname"`
		Birthday   *string   `json:"birthday,omitempty"`
		About      string    `json:"about"`
		Position   string    `json:"position"`
		Skills     []string  `json:"skills"`
		Role       string    `json:"role"`
		Status     string    `json:"status"`
		WorkStatus      string    `json:"work_status"`
		CreatedAt       time.Time `json:"created_at"`
		CurrentTaskID   *string   `json:"current_task_id,omitempty"`
	} `json:"data"`
}

// SetCurrentTaskRequest — тело запроса «текущая задача в работе».
type SetCurrentTaskRequest struct {
	TaskID *string `json:"task_id"` // UUID задачи; null или пустая строка — сброс текущей задачи
}

// UpdateWorkStatusRequest — тело запроса обновления статуса.
type UpdateWorkStatusRequest struct {
	WorkStatus string `json:"work_status" binding:"required"`
}

// WorkStatusHistoryResponse — ответ с историей статусов (для swagger).
type WorkStatusHistoryResponse struct {
	Data struct {
		History []WorkStatusHistoryEntry `json:"history"`
	} `json:"data"`
}

// ErrorResponse — стандартный формат ошибки (для swagger).
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
