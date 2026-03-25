package user

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"testik/internal/domain"
	"testik/pkg/response"
	"testik/pkg/storage"
)

// Handler — HTTP-обработчики профиля пользователя.
type Handler struct {
	service *Service
}

// NewHandler создаёт handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Me godoc
// @Summary Текущий пользователь
// @Description Возвращает данные авторизованного пользователя. Требует Bearer token.
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} user.UserResponse
// @Failure 401 {object} user.ErrorResponse
// @Failure 404 {object} user.ErrorResponse
// @Router /api/v1/users/me [get]
func (h *Handler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	u, err := h.service.GetUserByID(c.Request.Context(), userID.(string))
	if err != nil {
		response.Error(c, http.StatusNotFound, "not_found", "user not found")
		return
	}

	skills, _ := h.service.GetUserSkills(c.Request.Context(), userID.(string))
	if skills == nil {
		skills = []string{}
	}

	response.Data(c, http.StatusOK, buildUserResponse(u, skills))
}

// ListMyTasks godoc
// @Summary Мои задачи (исполнитель)
// @Description Задачи, где текущий пользователь — исполнитель. Поле in_work: true только у задачи, выбранной как «текущая в работе» (не более одной).
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} user.ErrorResponse
// @Failure 500 {object} user.ErrorResponse
// @Router /api/v1/users/me/tasks [get]
func (h *Handler) ListMyTasks(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	items, err := h.service.ListMyAssigneeTasks(c.Request.Context(), userID.(string))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to list tasks")
		return
	}

	tasks := make([]gin.H, 0, len(items))
	for _, it := range items {
		th := taskToH(it.Task)
		th["in_work"] = it.InWork
		tasks = append(tasks, th)
	}
	response.Data(c, http.StatusOK, gin.H{"tasks": tasks})
}

// SetCurrentTask godoc
// @Summary Установить текущую задачу в работе
// @Description Задача должна быть назначена на пользователя (assignee). Одновременно «в работе» только одна задача. task_id: null или "" — сброс.
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body SetCurrentTaskRequest true "task_id — UUID или null"
// @Success 200 {object} user.UserResponse
// @Failure 400 {object} user.ErrorResponse
// @Failure 401 {object} user.ErrorResponse
// @Failure 403 {object} user.ErrorResponse
// @Failure 404 {object} user.ErrorResponse
// @Failure 500 {object} user.ErrorResponse
// @Router /api/v1/users/me/current-task [put]
func (h *Handler) SetCurrentTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	var req SetCurrentTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	var taskID *uuid.UUID
	if req.TaskID != nil && *req.TaskID != "" {
		parsed, err := uuid.Parse(*req.TaskID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "validation_error", "invalid task_id")
			return
		}
		taskID = &parsed
	}

	u, err := h.service.SetCurrentTask(c.Request.Context(), userID.(string), taskID)
	if err != nil {
		switch {
		case errors.Is(err, ErrCurrentTaskNotFound):
			response.Error(c, http.StatusNotFound, "not_found", "task not found")
		case errors.Is(err, ErrCurrentTaskNotAssignee):
			response.Error(c, http.StatusForbidden, "forbidden", "task is not assigned to you")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to set current task")
		}
		return
	}

	skills, _ := h.service.GetUserSkills(c.Request.Context(), userID.(string))
	if skills == nil {
		skills = []string{}
	}
	response.Data(c, http.StatusOK, buildUserResponse(u, skills))
}

// UpdateProfile godoc
// @Summary Обновить профиль текущего пользователя
// @Description Обновляет firstname, lastname, birthday (YYYY-MM-DD или RFC3339; пустая строка сбрасывает дату), about, position, skills. Скиллы создаются в БД при первом добавлении.
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body UpdateProfileRequest true "Данные профиля"
// @Success 200 {object} user.UserResponse
// @Failure 400 {object} user.ErrorResponse
// @Failure 401 {object} user.ErrorResponse
// @Failure 500 {object} user.ErrorResponse
// @Router /api/v1/users/me [patch]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	u, err := h.service.GetUserByID(c.Request.Context(), userID.(string))
	if err != nil {
		response.Error(c, http.StatusNotFound, "not_found", "user not found")
		return
	}

	firstName, lastName := u.FirstName, u.LastName
	if req.FirstName != nil {
		firstName = *req.FirstName
	}
	if req.LastName != nil {
		lastName = *req.LastName
	}

	var birthday *time.Time
	if req.Birthday != nil {
		if *req.Birthday == "" {
			birthday = nil
		} else {
			parsed, err := parseBirthdayString(*req.Birthday)
			if err != nil {
				response.Error(c, http.StatusBadRequest, "validation_error", "invalid birthday (use YYYY-MM-DD or RFC3339)")
				return
			}
			birthday = parsed
		}
	} else {
		birthday = u.Birthday
	}

	about, position := u.About, u.Position
	if req.About != nil {
		about = *req.About
	}
	if req.Position != nil {
		position = *req.Position
	}

	in := UpdateProfileInput{
		FirstName: firstName,
		LastName:  lastName,
		Birthday:  birthday,
		About:     about,
		Position:  position,
		Skills:    req.Skills,
	}
	u, err = h.service.UpdateProfile(c.Request.Context(), userID.(string), in)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to update profile")
		return
	}

	skills, _ := h.service.GetUserSkills(c.Request.Context(), userID.(string))
	if skills == nil {
		skills = []string{}
	}

	response.Data(c, http.StatusOK, buildUserResponse(u, skills))
}

// UpdateWorkStatus godoc
// @Summary Обновить рабочий статус текущего пользователя
// @Description Обновляет рабочий статус: working, resting, lunch, vacation, sick_leave, business_trip. Запись добавляется в историю.
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body UpdateWorkStatusRequest true "Рабочий статус"
// @Success 200 {object} user.UserResponse
// @Failure 400 {object} user.ErrorResponse
// @Failure 401 {object} user.ErrorResponse
// @Failure 500 {object} user.ErrorResponse
// @Router /api/v1/users/me/work-status [put]
func (h *Handler) UpdateWorkStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	var req UpdateWorkStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	u, err := h.service.UpdateWorkStatus(c.Request.Context(), userID.(string), req.WorkStatus, userID.(string))
	if err != nil {
		if err == ErrInvalidWorkStatus {
			response.Error(c, http.StatusBadRequest, "validation_error", "invalid work_status: use working, resting, lunch, vacation, sick_leave, business_trip")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to update work status")
		return
	}

	skills, _ := h.service.GetUserSkills(c.Request.Context(), userID.(string))
	if skills == nil {
		skills = []string{}
	}

	response.Data(c, http.StatusOK, buildUserResponse(u, skills))
}

// GetWorkStatusHistory godoc
// @Summary История изменений рабочего статуса
// @Description Возвращает историю изменений рабочего статуса текущего пользователя.
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Лимит записей (по умолчанию 50)"
// @Success 200 {object} user.WorkStatusHistoryResponse
// @Failure 401 {object} user.ErrorResponse
// @Router /api/v1/users/me/work-status/history [get]
func (h *Handler) GetWorkStatusHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, err := parseInt(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	entries, err := h.service.GetWorkStatusHistory(c.Request.Context(), userID.(string), limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to get history")
		return
	}
	if entries == nil {
		entries = []WorkStatusHistoryEntry{}
	}

	response.Data(c, http.StatusOK, gin.H{"history": entries})
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// SetAvatar godoc
// @Summary Установить аватарку текущего пользователя
// @Description multipart/form-data с полем avatar (jpeg/png/webp/gif, max 5MB).
// @Tags user
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "Файл изображения"
// @Success 200 {object} user.UserResponse
// @Failure 400 {object} user.ErrorResponse
// @Failure 401 {object} user.ErrorResponse
// @Failure 500 {object} user.ErrorResponse
// @Router /api/v1/users/me/avatar [put]
func (h *Handler) SetAvatar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "avatar file is required")
		return
	}
	defer file.Close()

	avatar := &storage.AvatarInput{
		Reader:      file,
		Size:        header.Size,
		ContentType: header.Header.Get("Content-Type"),
	}

	u, err := h.service.SetAvatar(c.Request.Context(), userID.(string), avatar)
	if err != nil {
		if err == storage.ErrInvalidAvatar {
			response.Error(c, http.StatusBadRequest, "validation_error", "invalid avatar file (max 5MB, jpeg/png/webp/gif)")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to set avatar")
		return
	}

	skills, _ := h.service.GetUserSkills(c.Request.Context(), userID.(string))
	if skills == nil {
		skills = []string{}
	}

	response.Data(c, http.StatusOK, buildUserResponse(u, skills))
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

func parseBirthdayString(s string) (*time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return &t, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, err
	}
	utc := t.UTC()
	trunc := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
	return &trunc, nil
}

func buildUserResponse(u *domain.User, skills []string) gin.H {
	res := gin.H{
		"id":          u.ID.String(),
		"email":       u.Email,
		"firstname":   u.FirstName,
		"lastname":    u.LastName,
		"about":       u.About,
		"position":    u.Position,
		"skills":      skills,
		"role":        u.Role,
		"status":      u.Status,
		"work_status": u.WorkStatus,
		"created_at":  u.CreatedAt,
	}
	if u.Birthday != nil {
		ds := u.Birthday.UTC().Format("2006-01-02")
		res["birthday"] = ds
	}
	if u.CurrentTaskID != nil {
		res["current_task_id"] = u.CurrentTaskID.String()
	}
	if u.AvatarURL != "" {
		res["avatar_url"] = storage.AvatarURLForResponse(u.AvatarURL)
	}
	return res
}
