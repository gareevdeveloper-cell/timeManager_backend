package team

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"testik/internal/domain"
	"testik/pkg/response"
	"testik/pkg/storage"
)

// Handler — HTTP-обработчики команд.
type Handler struct {
	service *Service
}

// NewHandler создаёт handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func teamToResponse(t *domain.Team) gin.H {
	res := gin.H{
		"id":              t.ID.String(),
		"name":            t.Name,
		"description":     t.Description,
		"organization_id": t.OrganizationID.String(),
		"creator_id":      t.CreatorID.String(),
		"created_at":      t.CreatedAt,
	}
	if t.AvatarURL != "" {
		res["avatar_url"] = storage.AvatarURLForResponse(t.AvatarURL)
	}
	return res
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

// Create godoc
// @Summary Создать команду
// @Description Создаёт команду в организации. Создатель должен быть членом организации. Поддерживает JSON или multipart (name, description, organization_id + опционально avatar).
// @Tags teams
// @Accept json
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param body body CreateRequest true "Данные команды"
// @Success 201 {object} team.TeamResponse
// @Failure 400 {object} team.ErrorResponse
// @Failure 401 {object} team.ErrorResponse
// @Failure 403 {object} team.ErrorResponse
// @Failure 404 {object} team.ErrorResponse
// @Failure 500 {object} team.ErrorResponse
// @Router /api/v1/teams [post]
func (h *Handler) Create(c *gin.Context) {
	creatorID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	var name, description, orgIDStr string
	var avatar *storage.AvatarInput

	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		name = c.PostForm("name")
		description = c.PostForm("description")
		orgIDStr = c.PostForm("organization_id")
		if name == "" || orgIDStr == "" {
			response.Error(c, http.StatusBadRequest, "validation_error", "name and organization_id are required")
			return
		}
		file, header, err := c.Request.FormFile("avatar")
		if err == nil {
			defer file.Close()
			avatar = &storage.AvatarInput{Reader: file, Size: header.Size, ContentType: header.Header.Get("Content-Type")}
		}
	} else {
		var req CreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
			return
		}
		name, description, orgIDStr = req.Name, req.Description, req.OrganizationID
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid organization_id")
		return
	}

	t, err := h.service.Create(c.Request.Context(), name, description, orgID, creatorID, avatar)
	if err != nil {
		switch err {
		case ErrOrgNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "organization not found")
		case ErrUserNotInOrg:
			response.Error(c, http.StatusForbidden, "forbidden", "you must be a member of the organization to create a team")
		case ErrInvalidAvatar:
			response.Error(c, http.StatusBadRequest, "validation_error", "invalid avatar file (max 5MB, jpeg/png/webp/gif)")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to create team")
		}
		return
	}

	response.Data(c, http.StatusCreated, teamToResponse(t))
}

// GetByID godoc
// @Summary Получить команду по ID
// @Tags teams
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID команды"
// @Success 200 {object} team.TeamResponse
// @Failure 401 {object} team.ErrorResponse
// @Failure 404 {object} team.ErrorResponse
// @Router /api/v1/teams/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	_, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid team id")
		return
	}

	t, err := h.service.GetByID(c.Request.Context(), teamID)
	if err != nil || t == nil {
		response.Error(c, http.StatusNotFound, "not_found", "team not found")
		return
	}

	response.Data(c, http.StatusOK, teamToResponse(t))
}

// GetByOrganization godoc
// @Summary Получить команды организации
// @Tags teams
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID организации"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} team.ErrorResponse
// @Router /api/v1/organizations/{id}/teams [get]
func (h *Handler) GetByOrganization(c *gin.Context) {
	_, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid organization id")
		return
	}

	teams, err := h.service.GetByOrganization(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to get teams")
		return
	}

	items := make([]gin.H, 0, len(teams))
	for _, t := range teams {
		items = append(items, teamToResponse(t))
	}
	response.Data(c, http.StatusOK, gin.H{"teams": items})
}

// Update godoc
// @Summary Обновить команду
// @Description Поддерживает JSON или multipart (name, description + опционально avatar).
// @Tags teams
// @Accept json
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID команды"
// @Param body body UpdateRequest true "Данные для обновления"
// @Success 200 {object} team.TeamResponse
// @Failure 400 {object} team.ErrorResponse
// @Failure 401 {object} team.ErrorResponse
// @Failure 404 {object} team.ErrorResponse
// @Failure 500 {object} team.ErrorResponse
// @Router /api/v1/teams/{id} [patch]
func (h *Handler) Update(c *gin.Context) {
	_, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid team id")
		return
	}

	var name, description string
	var avatar *storage.AvatarInput

	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		name = c.PostForm("name")
		description = c.PostForm("description")
		if name == "" {
			response.Error(c, http.StatusBadRequest, "validation_error", "name is required")
			return
		}
		file, header, err := c.Request.FormFile("avatar")
		if err == nil {
			defer file.Close()
			avatar = &storage.AvatarInput{Reader: file, Size: header.Size, ContentType: header.Header.Get("Content-Type")}
		}
	} else {
		var req UpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
			return
		}
		name, description = req.Name, req.Description
	}

	t, err := h.service.Update(c.Request.Context(), teamID, name, description, avatar)
	if err != nil {
		switch err {
		case ErrNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "team not found")
		case ErrInvalidAvatar:
			response.Error(c, http.StatusBadRequest, "validation_error", "invalid avatar file (max 5MB, jpeg/png/webp/gif)")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to update team")
		}
		return
	}

	response.Data(c, http.StatusOK, teamToResponse(t))
}

// SetAvatar godoc
// @Summary Установить аватарку команды
// @Description multipart/form-data с полем avatar. Только член организации может изменить.
// @Tags teams
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID команды"
// @Param avatar formData file true "Файл изображения"
// @Success 200 {object} team.TeamResponse
// @Failure 400 {object} team.ErrorResponse
// @Failure 401 {object} team.ErrorResponse
// @Failure 404 {object} team.ErrorResponse
// @Failure 500 {object} team.ErrorResponse
// @Router /api/v1/teams/{id}/avatar [put]
func (h *Handler) SetAvatar(c *gin.Context) {
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

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "avatar file is required")
		return
	}
	defer file.Close()

	avatar := &storage.AvatarInput{Reader: file, Size: header.Size, ContentType: header.Header.Get("Content-Type")}

	t, err := h.service.SetAvatar(c.Request.Context(), teamID, userID, avatar)
	if err != nil {
		switch err {
		case ErrNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "team not found")
		case ErrInvalidAvatar:
			response.Error(c, http.StatusBadRequest, "validation_error", "invalid avatar file (max 5MB, jpeg/png/webp/gif)")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to set avatar")
		}
		return
	}

	response.Data(c, http.StatusOK, teamToResponse(t))
}

// Delete godoc
// @Summary Удалить команду
// @Tags teams
// @Security BearerAuth
// @Param id path string true "ID команды"
// @Success 204 "No Content"
// @Failure 401 {object} team.ErrorResponse
// @Failure 404 {object} team.ErrorResponse
// @Failure 500 {object} team.ErrorResponse
// @Router /api/v1/teams/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	_, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid team id")
		return
	}

	err = h.service.Delete(c.Request.Context(), teamID)
	if err != nil {
		switch err {
		case ErrNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "team not found")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to delete team")
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// AddMember godoc
// @Summary Добавить участника в команду
// @Description Добавлять можно только пользователей, состоящих в организации команды.
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID команды"
// @Param body body AddMemberRequest true "ID пользователя"
// @Success 204 "No Content"
// @Failure 400 {object} team.ErrorResponse
// @Failure 401 {object} team.ErrorResponse
// @Failure 403 {object} team.ErrorResponse
// @Failure 404 {object} team.ErrorResponse
// @Failure 409 {object} team.ErrorResponse
// @Failure 500 {object} team.ErrorResponse
// @Router /api/v1/teams/{id}/members [post]
func (h *Handler) AddMember(c *gin.Context) {
	_, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid team id")
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid user_id")
		return
	}

	err = h.service.AddMember(c.Request.Context(), teamID, userID, req.Role)
	if err != nil {
		switch err {
		case ErrNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "team not found")
		case ErrUserNotInOrg:
			response.Error(c, http.StatusForbidden, "forbidden", "user must be a member of the organization")
		case ErrUserAlreadyInTeam:
			response.Error(c, http.StatusConflict, "user_already_in_team", "user is already in the team")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to add member")
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// RemoveMember godoc
// @Summary Удалить участника из команды
// @Tags teams
// @Security BearerAuth
// @Param id path string true "ID команды"
// @Param user_id path string true "ID пользователя"
// @Success 204 "No Content"
// @Failure 401 {object} team.ErrorResponse
// @Failure 404 {object} team.ErrorResponse
// @Failure 500 {object} team.ErrorResponse
// @Router /api/v1/teams/{id}/members/{user_id} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
	_, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid team id")
		return
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid user_id")
		return
	}

	err = h.service.RemoveMember(c.Request.Context(), teamID, userID)
	if err != nil {
		switch err {
		case ErrNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "team not found")
		case ErrUserNotInTeam:
			response.Error(c, http.StatusNotFound, "not_found", "user is not a member of the team")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to remove member")
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GetMembers godoc
// @Summary Получить список участников команды
// @Tags teams
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID команды"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} team.ErrorResponse
// @Failure 404 {object} team.ErrorResponse
// @Failure 500 {object} team.ErrorResponse
// @Router /api/v1/teams/{id}/members [get]
func (h *Handler) GetMembers(c *gin.Context) {
	_, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	teamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid team id")
		return
	}

	members, err := h.service.GetMembers(c.Request.Context(), teamID)
	if err != nil {
		switch err {
		case ErrNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "team not found")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to get members")
		}
		return
	}

	items := make([]gin.H, 0, len(members))
	for _, m := range members {
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
		items = append(items, h)
	}
	response.Data(c, http.StatusOK, gin.H{"members": items})
}
