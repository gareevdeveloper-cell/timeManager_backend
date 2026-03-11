package organization

import "time"

// CreateRequest — тело запроса создания организации.
type CreateRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255" example:"My Company"`
}

// UpdateRequest — тело запроса обновления организации.
type UpdateRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255" example:"My Company Updated"`
}

// AddMemberRequest — тело запроса добавления участника.
type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role   string `json:"role" binding:"omitempty,oneof=administrator participant user" example:"participant"`
}

// CreateResponse — ответ создания организации (для swagger).
type CreateResponse struct {
	Data struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Slug      string    `json:"slug"`
		OwnerID   string    `json:"owner_id"`
		Status    string    `json:"status"`
		AvatarURL string    `json:"avatar_url,omitempty"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"data"`
}

// ListResponse — ответ списка организаций (для swagger).
type ListResponse struct {
	Data struct {
		Organizations []struct {
			ID        string    `json:"id"`
			Name      string    `json:"name"`
			Slug      string    `json:"slug"`
			OwnerID   string    `json:"owner_id"`
			Status    string    `json:"status"`
			AvatarURL string    `json:"avatar_url,omitempty"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"organizations"`
	} `json:"data"`
}

// ErrorResponse — стандартный формат ошибки (для swagger).
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
