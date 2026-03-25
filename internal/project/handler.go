package project

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"testik/internal/domain"
	"testik/pkg/response"
	"testik/pkg/storage"
)

// Handler — HTTP-обработчики проектов и задач.
type Handler struct {
	service *Service
}

// NewHandler создаёт handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func getCurrentUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(userID.(string))
	if err != nil {
		return uuid.Nil, false
	}
	return parsed, true
}

// parseTaskListFilter читает query-параметры списка задач. Второе значение — сообщение об ошибке валидации (если непустое).
func parseTaskListFilter(c *gin.Context) (TaskListFilter, string) {
	f := TaskListFilter{
		Status: c.Query("status"),
		Title:  strings.TrimSpace(c.Query("title")),
		Type:   strings.TrimSpace(c.Query("type")),
	}
	if aid := strings.TrimSpace(c.Query("assignee_id")); aid != "" {
		u, err := uuid.Parse(aid)
		if err != nil {
			return TaskListFilter{}, "invalid assignee_id"
		}
		f.AssigneeID = &u
	}
	if df := strings.TrimSpace(c.Query("due_from")); df != "" {
		t, err := time.Parse(time.RFC3339, df)
		if err != nil {
			return TaskListFilter{}, "invalid due_from (use RFC3339)"
		}
		f.DueFrom = &t
	}
	if dt := strings.TrimSpace(c.Query("due_to")); dt != "" {
		t, err := time.Parse(time.RFC3339, dt)
		if err != nil {
			return TaskListFilter{}, "invalid due_to (use RFC3339)"
		}
		f.DueTo = &t
	}
	return f, ""
}

func projectToH(p *domain.Project) gin.H {
	h := gin.H{
		"id": p.ID.String(), "key": p.Key, "name": p.Name, "description": p.Description,
		"owner_id": p.OwnerID.String(), "created_at": p.CreatedAt, "updated_at": p.UpdatedAt,
	}
	if p.TeamID != uuid.Nil {
		h["team_id"] = p.TeamID.String()
	} else {
		h["team_id"] = nil
	}
	return h
}

func taskToH(t *domain.Task) gin.H {
	h := gin.H{
		"id": t.ID.String(), "project_id": t.ProjectID.String(), "key": t.Key,
		"title": t.Title, "description": t.Description, "status": t.Status, "status_id": t.StatusID.String(), "type": t.Type, "priority": t.Priority,
		"reporter_id": t.ReporterID.String(), "author_id": t.ReporterID.String(), "order": t.Order,
		"tags": t.Tags, "created_at": t.CreatedAt, "updated_at": t.UpdatedAt,
	}
	if t.AssigneeID != nil {
		h["assignee_id"] = t.AssigneeID.String()
	} else {
		h["assignee_id"] = nil
	}
	if t.DueDate != nil {
		h["due_date"] = t.DueDate.Format("2006-01-02T15:04:05Z07:00")
	} else {
		h["due_date"] = nil
	}
	if t.ResultURL != "" {
		h["result_url"] = t.ResultURL
	} else {
		h["result_url"] = nil
	}
	if t.Tags == nil {
		h["tags"] = []string{}
	}
	return h
}

// CreateProject godoc
// @Summary Создать проект
// @Description Создаёт проект. key — уникальный короткий код (например, APP).
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateProjectRequest true "Данные проекта"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} project.ErrorResponse
// @Failure 401 {object} project.ErrorResponse
// @Failure 409 {object} project.ErrorResponse
// @Failure 500 {object} project.ErrorResponse
// @Router /api/v1/projects [post]
func (h *Handler) CreateProject(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	var teamID uuid.UUID
	if req.TeamID != "" {
		parsed, err := uuid.Parse(req.TeamID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "validation_error", "invalid team_id")
			return
		}
		teamID = parsed
	}

	p, err := h.service.CreateProject(c.Request.Context(), teamID, req.Key, req.Name, req.Description, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrKeyAlreadyExists):
			response.Error(c, http.StatusConflict, "key_exists", "project key already exists")
		case errors.Is(err, ErrInvalidKey):
			response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		case errors.Is(err, ErrTeamNotFound):
			response.Error(c, http.StatusNotFound, "not_found", "team not found")
		case errors.Is(err, ErrUserNotInTeam):
			response.Error(c, http.StatusForbidden, "forbidden", "user is not a member of the team")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to create project")
		}
		return
	}

	response.Data(c, http.StatusCreated, projectToH(p))
}

// ListProjects godoc
// @Summary Список проектов
// @Description Возвращает проекты, доступные пользователю (владелец или член команды).
// @Tags projects
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} project.ErrorResponse
// @Failure 500 {object} project.ErrorResponse
// @Router /api/v1/projects [get]
func (h *Handler) ListProjects(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	projects, err := h.service.ListProjects(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to list projects")
		return
	}

	items := make([]gin.H, 0, len(projects))
	for _, p := range projects {
		items = append(items, projectToH(p))
	}
	response.Data(c, http.StatusOK, gin.H{"projects": items})
}

// GetProject godoc
// @Summary Получить проект
// @Description Возвращает проект по ID.
// @Tags projects
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} project.ErrorResponse
// @Failure 403 {object} project.ErrorResponse
// @Failure 404 {object} project.ErrorResponse
// @Failure 500 {object} project.ErrorResponse
// @Router /api/v1/projects/{projectId} [get]
func (h *Handler) GetProject(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid project id")
		return
	}

	p, err := h.service.GetProject(c.Request.Context(), projectID, userID)
	if err != nil {
		if err == ErrProjectNotFound {
			response.Error(c, http.StatusNotFound, "not_found", "project not found")
			return
		}
		if err == ErrForbidden {
			response.Error(c, http.StatusForbidden, "forbidden", "access denied to project")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to get project")
		return
	}

	response.Data(c, http.StatusOK, projectToH(p))
}

// GetProjectMembers godoc
// @Summary Участники проекта
// @Description Возвращает участников проекта с ролями. У каждого: current_task_id и current_task {id, title, project_id} — текущая задача в работе (если есть). Доступ: владелец или член команды.
// @Tags projects
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} project.ErrorResponse
// @Failure 403 {object} project.ErrorResponse
// @Failure 404 {object} project.ErrorResponse
// @Failure 500 {object} project.ErrorResponse
// @Router /api/v1/projects/{projectId}/members [get]
func (h *Handler) GetProjectMembers(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid project id")
		return
	}

	members, err := h.service.ListProjectMembers(c.Request.Context(), projectID, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			response.Error(c, http.StatusNotFound, "not_found", "project not found")
		case errors.Is(err, ErrForbidden):
			response.Error(c, http.StatusForbidden, "forbidden", "access denied")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to list members")
		}
		return
	}

	items := make([]gin.H, 0, len(members))
	for _, m := range members {
		items = append(items, memberWithRoleToH(m))
	}
	response.Data(c, http.StatusOK, gin.H{"members": items})
}

func memberWithRoleToH(m *domain.MemberWithRole) gin.H {
	u := m.User
	h := gin.H{
		"id":         u.ID.String(),
		"email":      u.Email,
		"firstname":  u.FirstName,
		"lastname":   u.LastName,
		"middlename": u.MiddleName,
		"role":       m.Role,
	}
	if u.AvatarURL != "" {
		h["avatar_url"] = storage.AvatarURLForResponse(u.AvatarURL)
	}
	if m.CurrentTask != nil {
		h["current_task_id"] = m.CurrentTask.ID.String()
		h["current_task"] = gin.H{
			"id":         m.CurrentTask.ID.String(),
			"title":      m.CurrentTask.Title,
			"project_id": m.CurrentTask.ProjectID.String(),
		}
	} else {
		h["current_task_id"] = nil
		h["current_task"] = nil
	}
	return h
}

// ListStatuses godoc
// @Summary Статусы (колонки) проекта
// @Description Возвращает динамические статусы проекта для построения канбан-доски.
// @Tags projects
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} project.ErrorResponse
// @Failure 403 {object} project.ErrorResponse
// @Failure 404 {object} project.ErrorResponse
// @Failure 500 {object} project.ErrorResponse
// @Router /api/v1/projects/{projectId}/statuses [get]
func (h *Handler) ListStatuses(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid project id")
		return
	}

	statuses, err := h.service.ListStatuses(c.Request.Context(), projectID, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			response.Error(c, http.StatusNotFound, "not_found", "project not found")
		case errors.Is(err, ErrForbidden):
			response.Error(c, http.StatusForbidden, "forbidden", "access denied")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to list statuses")
		}
		return
	}

	items := make([]gin.H, 0, len(statuses))
	for _, s := range statuses {
		items = append(items, gin.H{
			"id":         s.ID.String(),
			"project_id": s.ProjectID.String(),
			"key":        s.Key,
			"title":     s.Title,
			"order":     s.Order,
		})
	}
	response.Data(c, http.StatusOK, gin.H{"statuses": items})
}

// CreateStatus godoc
// @Summary Создать статус (колонку)
// @Description Создаёт новый статус в проекте для канбан-доски.
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Param body body CreateStatusRequest true "Данные статуса"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} project.ErrorResponse
// @Failure 401 {object} project.ErrorResponse
// @Failure 403 {object} project.ErrorResponse
// @Failure 404 {object} project.ErrorResponse
// @Failure 409 {object} project.ErrorResponse
// @Failure 500 {object} project.ErrorResponse
// @Router /api/v1/projects/{projectId}/statuses [post]
func (h *Handler) CreateStatus(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid project id")
		return
	}

	var req CreateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	s, err := h.service.CreateStatus(c.Request.Context(), projectID, userID, req.Key, req.Title, req.Order)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			response.Error(c, http.StatusNotFound, "not_found", "project not found")
		case errors.Is(err, ErrForbidden):
			response.Error(c, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, ErrStatusKeyExists):
			response.Error(c, http.StatusConflict, "status_key_exists", "status with this key already exists")
		case errors.Is(err, ErrStatusTitleExists):
			response.Error(c, http.StatusConflict, "status_title_exists", "status with this title already exists")
		case errors.Is(err, ErrInvalidRequest):
			response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to create status")
		}
		return
	}

	response.Data(c, http.StatusCreated, gin.H{
		"id":         s.ID.String(),
		"project_id": s.ProjectID.String(),
		"key":        s.Key,
		"title":     s.Title,
		"order":     s.Order,
	})
}

// UpdateStatus godoc
// @Summary Обновить статус
// @Description Обновляет статус (колонку) проекта.
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param statusId path string true "ID статуса"
// @Param body body UpdateStatusRequest true "Поля для обновления"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} project.ErrorResponse
// @Failure 401 {object} project.ErrorResponse
// @Failure 403 {object} project.ErrorResponse
// @Failure 404 {object} project.ErrorResponse
// @Failure 409 {object} project.ErrorResponse
// @Failure 500 {object} project.ErrorResponse
// @Router /api/v1/projects/statuses/{statusId} [patch]
func (h *Handler) UpdateStatus(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	statusID, err := uuid.Parse(c.Param("statusId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid status id")
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	s, err := h.service.UpdateStatus(c.Request.Context(), statusID, userID, req.Key, req.Title, req.Order)
	if err != nil {
		switch {
		case errors.Is(err, ErrStatusNotFound):
			response.Error(c, http.StatusNotFound, "not_found", "status not found")
		case errors.Is(err, ErrForbidden):
			response.Error(c, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, ErrStatusKeyExists):
			response.Error(c, http.StatusConflict, "status_key_exists", "status with this key already exists")
		case errors.Is(err, ErrStatusTitleExists):
			response.Error(c, http.StatusConflict, "status_title_exists", "status with this title already exists")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to update status")
		}
		return
	}

	response.Data(c, http.StatusOK, gin.H{
		"id":         s.ID.String(),
		"project_id": s.ProjectID.String(),
		"key":        s.Key,
		"title":     s.Title,
		"order":     s.Order,
	})
}

// DeleteStatus godoc
// @Summary Удалить статус
// @Description Удаляет колонку (статус). Переносит все задачи в колонку move_to_status_id (query). Нельзя удалить последний статус.
// @Tags projects
// @Produce json
// @Security BearerAuth
// @Param statusId path string true "ID удаляемого статуса"
// @Param move_to_status_id query string true "ID статуса того же проекта, куда перенести задачи"
// @Success 204 "No Content"
// @Failure 400 {object} project.ErrorResponse
// @Failure 401 {object} project.ErrorResponse
// @Failure 403 {object} project.ErrorResponse
// @Failure 404 {object} project.ErrorResponse
// @Failure 409 {object} project.ErrorResponse
// @Failure 500 {object} project.ErrorResponse
// @Router /api/v1/projects/statuses/{statusId} [delete]
func (h *Handler) DeleteStatus(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	statusID, err := uuid.Parse(c.Param("statusId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid status id")
		return
	}

	moveTo := c.Query("move_to_status_id")
	if moveTo == "" {
		response.Error(c, http.StatusBadRequest, "validation_error", "move_to_status_id is required")
		return
	}
	moveToID, err := uuid.Parse(moveTo)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid move_to_status_id")
		return
	}

	err = h.service.DeleteStatus(c.Request.Context(), statusID, userID, moveToID)
	if err != nil {
		switch {
		case errors.Is(err, ErrStatusNotFound):
			response.Error(c, http.StatusNotFound, "not_found", "status not found")
		case errors.Is(err, ErrForbidden):
			response.Error(c, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, ErrLastStatusCannotDelete):
			response.Error(c, http.StatusConflict, "last_status", "cannot delete the last status in project")
		case errors.Is(err, ErrInvalidMoveTarget):
			response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to delete status")
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateTask godoc
// @Summary Создать задачу
// @Description Создаёт задачу в проекте. key генерируется автоматически (PROJECTKEY-N).
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Param body body CreateTaskRequest true "Данные задачи"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} project.ErrorResponse
// @Failure 401 {object} project.ErrorResponse
// @Failure 403 {object} project.ErrorResponse
// @Failure 404 {object} project.ErrorResponse
// @Failure 500 {object} project.ErrorResponse
// @Router /api/v1/projects/{projectId}/tasks [post]
func (h *Handler) CreateTask(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid project id")
		return
	}

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	t, err := h.service.CreateTask(c.Request.Context(), projectID, userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			response.Error(c, http.StatusNotFound, "not_found", "project not found")
		case errors.Is(err, ErrForbidden):
			response.Error(c, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, ErrInvalidPriority), errors.Is(err, ErrInvalidRequest), errors.Is(err, ErrInvalidType):
			response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to create task")
		}
		return
	}

	response.Data(c, http.StatusCreated, taskToH(t))
}

// ListTasks godoc
// @Summary Список задач проекта
// @Description Возвращает задачи проекта. Опциональные query-фильтры: status, assignee_id, title (подстрока), type, due_from, due_to (RFC3339).
// @Tags tasks
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Param status query string false "Фильтр по статусу"
// @Param assignee_id query string false "Фильтр по исполнителю (UUID)"
// @Param title query string false "Подстрока заголовка (без учёта регистра)"
// @Param type query string false "Тип: BUG, TASK, STORY"
// @Param due_from query string false "Срок с (RFC3339), только задачи с due_date >= due_from"
// @Param due_to query string false "Срок по (RFC3339), только задачи с due_date <= due_to"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} project.ErrorResponse
// @Failure 401 {object} project.ErrorResponse
// @Failure 403 {object} project.ErrorResponse
// @Failure 404 {object} project.ErrorResponse
// @Failure 500 {object} project.ErrorResponse
// @Router /api/v1/projects/{projectId}/tasks [get]
func (h *Handler) ListTasks(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid project id")
		return
	}

	filter, badQuery := parseTaskListFilter(c)
	if badQuery != "" {
		response.Error(c, http.StatusBadRequest, "validation_error", badQuery)
		return
	}

	tasks, err := h.service.ListTasks(c.Request.Context(), projectID, userID, filter)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			response.Error(c, http.StatusNotFound, "not_found", "project not found")
		case errors.Is(err, ErrForbidden):
			response.Error(c, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, ErrInvalidType):
			response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		case errors.Is(err, ErrInvalidRequest):
			response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to list tasks")
		}
		return
	}

	items := make([]gin.H, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, taskToH(t))
	}
	response.Data(c, http.StatusOK, gin.H{"tasks": items})
}

// GetTask godoc
// @Summary Получить задачу
// @Description Возвращает задачу по ID.
// @Tags tasks
// @Produce json
// @Security BearerAuth
// @Param taskId path string true "ID задачи"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} project.ErrorResponse
// @Failure 403 {object} project.ErrorResponse
// @Failure 404 {object} project.ErrorResponse
// @Failure 500 {object} project.ErrorResponse
// @Router /api/v1/tasks/{taskId} [get]
func (h *Handler) GetTask(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid task id")
		return
	}

	t, err := h.service.GetTask(c.Request.Context(), taskID, userID)
	if err != nil {
		switch err {
		case ErrTaskNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "task not found")
		case ErrForbidden:
			response.Error(c, http.StatusForbidden, "forbidden", "access denied")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to get task")
		}
		return
	}

	response.Data(c, http.StatusOK, taskToH(t))
}

// UpdateTask godoc
// @Summary Обновить задачу
// @Description Частичное обновление задачи (статус, приоритет, assignee, порядок в колонке и т.д.).
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param taskId path string true "ID задачи"
// @Param body body UpdateTaskRequest true "Поля для обновления"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} project.ErrorResponse
// @Failure 401 {object} project.ErrorResponse
// @Failure 403 {object} project.ErrorResponse
// @Failure 404 {object} project.ErrorResponse
// @Failure 500 {object} project.ErrorResponse
// @Router /api/v1/tasks/{taskId} [patch]
func (h *Handler) UpdateTask(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid task id")
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	t, err := h.service.UpdateTask(c.Request.Context(), taskID, userID, req)
	if err != nil {
		switch err {
		case ErrTaskNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "task not found")
		case ErrForbidden:
			response.Error(c, http.StatusForbidden, "forbidden", "access denied")
		case ErrInvalidStatus, ErrInvalidPriority, ErrInvalidType:
			response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		default:
			response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		}
		return
	}

	response.Data(c, http.StatusOK, taskToH(t))
}

// DeleteTask godoc
// @Summary Удалить задачу
// @Description Удаляет задачу.
// @Tags tasks
// @Produce json
// @Security BearerAuth
// @Param taskId path string true "ID задачи"
// @Success 204 "No Content"
// @Failure 401 {object} project.ErrorResponse
// @Failure 403 {object} project.ErrorResponse
// @Failure 404 {object} project.ErrorResponse
// @Failure 500 {object} project.ErrorResponse
// @Router /api/v1/tasks/{taskId} [delete]
func (h *Handler) DeleteTask(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid task id")
		return
	}

	err = h.service.DeleteTask(c.Request.Context(), taskID, userID)
	if err != nil {
		switch err {
		case ErrTaskNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "task not found")
		case ErrForbidden:
			response.Error(c, http.StatusForbidden, "forbidden", "access denied")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to delete task")
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GetBoard godoc
// @Summary Канбан-доска проекта
// @Description Возвращает доску с колонками (динамические статусы) и задачами в каждой. Те же query-фильтры, что у списка задач: status, assignee_id, title, type, due_from, due_to.
// @Tags projects
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Param status query string false "Фильтр по статусу"
// @Param assignee_id query string false "Фильтр по исполнителю (UUID)"
// @Param title query string false "Подстрока заголовка (без учёта регистра)"
// @Param type query string false "Тип: BUG, TASK, STORY"
// @Param due_from query string false "Срок с (RFC3339)"
// @Param due_to query string false "Срок по (RFC3339)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} project.ErrorResponse
// @Failure 401 {object} project.ErrorResponse
// @Failure 403 {object} project.ErrorResponse
// @Failure 404 {object} project.ErrorResponse
// @Failure 500 {object} project.ErrorResponse
// @Router /api/v1/projects/{projectId}/board [get]
func (h *Handler) GetBoard(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid project id")
		return
	}

	filter, badQuery := parseTaskListFilter(c)
	if badQuery != "" {
		response.Error(c, http.StatusBadRequest, "validation_error", badQuery)
		return
	}

	columns, err := h.service.GetBoard(c.Request.Context(), projectID, userID, filter)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			response.Error(c, http.StatusNotFound, "not_found", "project not found")
		case errors.Is(err, ErrForbidden):
			response.Error(c, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, ErrInvalidType):
			response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		case errors.Is(err, ErrInvalidRequest):
			response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to get board")
		}
		return
	}

	colItems := make([]gin.H, 0, len(columns))
	for _, col := range columns {
		tasks := make([]gin.H, 0, len(col.Tasks))
		for _, t := range col.Tasks {
			tasks = append(tasks, taskToH(t))
		}
		colItems = append(colItems, gin.H{
			"status": col.Status,
			"title":  col.Title,
			"order":  col.Order,
			"tasks":  tasks,
		})
	}
	response.Data(c, http.StatusOK, gin.H{"columns": colItems})
}

// GetProjectsByTeam godoc
// @Summary Проекты команды
// @Description Возвращает проекты команды. Только для членов команды.
// @Tags teams
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID команды"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} project.ErrorResponse
// @Failure 403 {object} project.ErrorResponse
// @Failure 404 {object} project.ErrorResponse
// @Failure 500 {object} project.ErrorResponse
// @Router /api/v1/teams/{id}/projects [get]
func (h *Handler) GetProjectsByTeam(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid team id")
		return
	}

	projects, err := h.service.ListProjectsByTeam(c.Request.Context(), teamID, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrTeamNotFound):
			response.Error(c, http.StatusNotFound, "not_found", "team not found")
		case errors.Is(err, ErrForbidden):
			response.Error(c, http.StatusForbidden, "forbidden", "access denied")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to list projects")
		}
		return
	}

	items := make([]gin.H, 0, len(projects))
	for _, p := range projects {
		items = append(items, projectToH(p))
	}
	response.Data(c, http.StatusOK, gin.H{"projects": items})
}
