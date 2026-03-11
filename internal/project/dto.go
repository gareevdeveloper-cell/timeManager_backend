package project

// CreateProjectRequest — тело запроса создания проекта.
// team_id опционален — без него создаётся личный проект без привязки к команде.
type CreateProjectRequest struct {
	TeamID      string `json:"team_id" binding:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Key         string `json:"key" binding:"required,min=1,max=50" example:"APP"`
	Name        string `json:"name" binding:"required,min=1,max=255" example:"My App"`
	Description string `json:"description" example:"Описание проекта"`
}

// CreateTaskRequest — тело запроса создания задачи.
type CreateTaskRequest struct {
	Title      string  `json:"title" binding:"required,min=1,max=500" example:"Реализовать API"`
	Description string `json:"description" example:"Описание задачи"`
	Priority   string  `json:"priority" example:"MEDIUM"`
	AssigneeID *string `json:"assignee_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	DueDate    *string `json:"due_date" example:"2025-03-15T12:00:00Z"`
}

// UpdateTaskRequest — тело запроса обновления задачи (partial).
type UpdateTaskRequest struct {
	Title      *string `json:"title"`
	Description *string `json:"description"`
	Status    *string `json:"status"`
	Priority  *string `json:"priority"`
	AssigneeID *string `json:"assignee_id"`
	DueDate   *string `json:"due_date"`
	Order     *int    `json:"order"`
}

// CreateStatusRequest — тело запроса создания статуса.
type CreateStatusRequest struct {
	Key   string `json:"key" binding:"required,min=1,max=50" example:"TODO"`
	Title string `json:"title" binding:"required,min=1,max=255" example:"To Do"`
	Order int    `json:"order" example:"0"`
}

// UpdateStatusRequest — тело запроса обновления статуса (partial).
type UpdateStatusRequest struct {
	Key   string `json:"key" example:"TODO"`
	Title string `json:"title" example:"To Do"`
	Order *int   `json:"order"`
}

// ErrorResponse — стандартный формат ошибки (для swagger).
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
