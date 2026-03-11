package team

import "time"

// CreateRequest — тело запроса создания команды.
type CreateRequest struct {
	Name           string `json:"name" binding:"required,min=1,max=255" example:"Backend Team"`
	Description    string `json:"description" example:"Команда разработки бэкенда"`
	OrganizationID string `json:"organization_id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// UpdateRequest — тело запроса обновления команды.
type UpdateRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255" example:"Backend Team Updated"`
	Description string `json:"description" example:"Обновлённое описание"`
}

// AddMemberRequest — тело запроса добавления участника.
type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role   string `json:"role" binding:"omitempty,oneof=administrator participant user" example:"participant"`
}

// TeamResponse — ответ с данными команды (для swagger).
type TeamResponse struct {
	Data struct {
		ID             string    `json:"id"`
		Name           string    `json:"name"`
		Description    string    `json:"description"`
		OrganizationID string    `json:"organization_id"`
		CreatorID      string    `json:"creator_id"`
		CreatedAt      time.Time `json:"created_at"`
	} `json:"data"`
}

// ErrorResponse — стандартный формат ошибки (для swagger).
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
