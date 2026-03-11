package user

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

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

// UpdateProfile godoc
// @Summary Обновить профиль текущего пользователя
// @Description Обновляет поля about (о себе), position (должность), skills (скиллы). Скиллы создаются в БД при первом добавлении.
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

	about, position := u.About, u.Position
	if req.About != nil {
		about = *req.About
	}
	if req.Position != nil {
		position = *req.Position
	}

	in := UpdateProfileInput{About: about, Position: position, Skills: req.Skills}
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
	if u.AvatarURL != "" {
		res["avatar_url"] = storage.AvatarURLForResponse(u.AvatarURL)
	}
	return res
}
