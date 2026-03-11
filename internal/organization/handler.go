package organization

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"testik/internal/domain"
	"testik/pkg/response"
	"testik/pkg/storage"
)

// Handler — HTTP-обработчики организаций.
type Handler struct {
	service *Service
}

// NewHandler создаёт handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetByID godoc
// @Summary Получить организацию по ID
// @Description Возвращает организацию. Доступ только для членов организации.
// @Tags organizations
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID организации"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} organization.ErrorResponse
// @Failure 404 {object} organization.ErrorResponse
// @Failure 500 {object} organization.ErrorResponse
// @Router /api/v1/organizations/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid organization id")
		return
	}

	o, err := h.service.GetByID(c.Request.Context(), orgID, userID)
	if err != nil {
		if err == ErrNotFound {
			response.Error(c, http.StatusNotFound, "not_found", "organization not found")
		} else {
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to get organization")
		}
		return
	}

	response.Data(c, http.StatusOK, orgToResponse(o))
}

// ListMyOrganizations godoc
// @Summary Список моих организаций
// @Description Возвращает организации, в которых текущий пользователь является участником.
// @Tags organizations
// @Produce json
// @Security BearerAuth
// @Success 200 {object} organization.ListResponse
// @Failure 401 {object} organization.ErrorResponse
// @Failure 500 {object} organization.ErrorResponse
// @Router /api/v1/organizations [get]
func (h *Handler) ListMyOrganizations(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	orgs, err := h.service.ListMyOrganizations(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to list organizations")
		return
	}

	items := make([]gin.H, 0, len(orgs))
	for _, o := range orgs {
		items = append(items, orgToResponse(o))
	}
	response.Data(c, http.StatusOK, gin.H{"organizations": items})
}

func orgToResponse(o *domain.Organization) gin.H {
	res := gin.H{
		"id":         o.ID.String(),
		"name":       o.Name,
		"slug":       o.Slug,
		"owner_id":   o.OwnerID.String(),
		"status":     o.Status,
		"created_at": o.CreatedAt,
		"updated_at": o.UpdatedAt,
	}
	if o.AvatarURL != "" {
		res["avatar_url"] = storage.AvatarURLForResponse(o.AvatarURL)
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
// @Summary Создать организацию
// @Description Создаёт организацию от имени текущего пользователя. Slug генерируется автоматически из name. Поддерживает JSON или multipart/form-data (name + опционально avatar).
// @Tags organizations
// @Accept json
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param body body CreateRequest true "Название организации (JSON) или multipart: name, avatar"
// @Success 201 {object} organization.CreateResponse
// @Failure 400 {object} organization.ErrorResponse
// @Failure 401 {object} organization.ErrorResponse
// @Failure 500 {object} organization.ErrorResponse
// @Router /api/v1/organizations [post]
func (h *Handler) Create(c *gin.Context) {
	ownerID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	var name string
	var avatar *AvatarInput

	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		name = c.PostForm("name")
		if name == "" {
			response.Error(c, http.StatusBadRequest, "validation_error", "name is required")
			return
		}
		file, header, err := c.Request.FormFile("avatar")
		if err == nil {
			defer file.Close()
			avatar = &AvatarInput{
				Reader:      file,
				Size:       header.Size,
				ContentType: header.Header.Get("Content-Type"),
			}
		}
	} else {
		var req CreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
			return
		}
		name = req.Name
	}

	o, err := h.service.Create(c.Request.Context(), name, ownerID, avatar)
	if err != nil {
		if err == ErrInvalidAvatar {
			response.Error(c, http.StatusBadRequest, "validation_error", "invalid avatar file (max 5MB, jpeg/png/webp/gif)")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to create organization")
		return
	}

	response.Data(c, http.StatusCreated, orgToResponse(o))
}

// Update godoc
// @Summary Обновить организацию
// @Description Обновляет название и опционально аватарку. Slug остаётся неизменным. Поддерживает JSON или multipart (name + опционально avatar).
// @Tags organizations
// @Accept json
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID организации"
// @Param body body UpdateRequest true "Новое название (JSON) или multipart: name, avatar"
// @Success 200 {object} organization.CreateResponse
// @Failure 400 {object} organization.ErrorResponse
// @Failure 401 {object} organization.ErrorResponse
// @Failure 403 {object} organization.ErrorResponse
// @Failure 404 {object} organization.ErrorResponse
// @Failure 500 {object} organization.ErrorResponse
// @Router /api/v1/organizations/{id} [patch]
func (h *Handler) Update(c *gin.Context) {
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

	var name string
	var avatar *AvatarInput

	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		name = c.PostForm("name")
		if name == "" {
			response.Error(c, http.StatusBadRequest, "validation_error", "name is required")
			return
		}
		file, header, err := c.Request.FormFile("avatar")
		if err == nil {
			defer file.Close()
			avatar = &AvatarInput{
				Reader:      file,
				Size:       header.Size,
				ContentType: header.Header.Get("Content-Type"),
			}
		}
	} else {
		var req UpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
			return
		}
		name = req.Name
	}

	o, err := h.service.Update(c.Request.Context(), orgID, name, avatar)
	if err != nil {
		switch err {
		case ErrNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "organization not found")
		case ErrArchived:
			response.Error(c, http.StatusForbidden, "archived", "cannot update archived organization")
		case ErrInvalidAvatar:
			response.Error(c, http.StatusBadRequest, "validation_error", "invalid avatar file (max 5MB, jpeg/png/webp/gif)")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to update organization")
		}
		return
	}

	response.Data(c, http.StatusOK, orgToResponse(o))
}

// Archive godoc
// @Summary Архивировать организацию
// @Description Переводит организацию в статус archived.
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID организации"
// @Success 200 {object} organization.CreateResponse
// @Failure 401 {object} organization.ErrorResponse
// @Failure 403 {object} organization.ErrorResponse
// @Failure 404 {object} organization.ErrorResponse
// @Failure 500 {object} organization.ErrorResponse
// @Router /api/v1/organizations/{id}/archive [post]
func (h *Handler) Archive(c *gin.Context) {
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

	o, err := h.service.Archive(c.Request.Context(), orgID)
	if err != nil {
		switch err {
		case ErrNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "organization not found")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to archive organization")
		}
		return
	}

	response.Data(c, http.StatusOK, orgToResponse(o))
}

// SetAvatar godoc
// @Summary Установить аватарку организации
// @Description Загружает аватарку в MinIO. multipart/form-data с полем avatar (jpeg/png/webp/gif, max 5MB).
// @Tags organizations
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID организации"
// @Param avatar formData file true "Файл изображения"
// @Success 200 {object} organization.CreateResponse
// @Failure 400 {object} organization.ErrorResponse
// @Failure 401 {object} organization.ErrorResponse
// @Failure 403 {object} organization.ErrorResponse
// @Failure 404 {object} organization.ErrorResponse
// @Failure 500 {object} organization.ErrorResponse
// @Router /api/v1/organizations/{id}/avatar [put]
func (h *Handler) SetAvatar(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid organization id")
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "avatar file is required")
		return
	}
	defer file.Close()

	avatar := &AvatarInput{
		Reader:      file,
		Size:       header.Size,
		ContentType: header.Header.Get("Content-Type"),
	}

	o, err := h.service.SetAvatar(c.Request.Context(), orgID, userID, avatar)
	if err != nil {
		switch err {
		case ErrNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "organization not found")
		case ErrArchived:
			response.Error(c, http.StatusForbidden, "archived", "cannot update archived organization")
		case ErrInvalidAvatar:
			response.Error(c, http.StatusBadRequest, "validation_error", "invalid avatar file (max 5MB, jpeg/png/webp/gif)")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to set avatar")
		}
		return
	}

	response.Data(c, http.StatusOK, orgToResponse(o))
}

// AddMember godoc
// @Summary Добавить участника в организацию
// @Description Добавляет пользователя в организацию. Пользователь может быть только в одной организации.
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID организации"
// @Param body body AddMemberRequest true "ID пользователя"
// @Success 204 "No Content"
// @Failure 400 {object} organization.ErrorResponse
// @Failure 401 {object} organization.ErrorResponse
// @Failure 403 {object} organization.ErrorResponse
// @Failure 404 {object} organization.ErrorResponse
// @Failure 409 {object} organization.ErrorResponse
// @Failure 500 {object} organization.ErrorResponse
// @Router /api/v1/organizations/{id}/members [post]
func (h *Handler) AddMember(c *gin.Context) {
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

	err = h.service.AddMember(c.Request.Context(), orgID, userID, req.Role)
	if err != nil {
		switch err {
		case ErrNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "organization not found")
		case ErrArchived:
			response.Error(c, http.StatusForbidden, "archived", "cannot add members to archived organization")
		case ErrUserAlreadyInOrg:
			response.Error(c, http.StatusConflict, "user_already_in_org", "user is already in an organization")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to add member")
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GetMembers godoc
// @Summary Список участников организации
// @Description Возвращает всех пользователей, входящих в организацию. Доступ только для членов организации.
// @Tags organizations
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID организации"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} organization.ErrorResponse
// @Failure 404 {object} organization.ErrorResponse
// @Failure 500 {object} organization.ErrorResponse
// @Router /api/v1/organizations/{id}/members [get]
func (h *Handler) GetMembers(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	orgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid organization id")
		return
	}

	members, err := h.service.ListMembers(c.Request.Context(), orgID, userID)
	if err != nil {
		if err == ErrNotFound {
			response.Error(c, http.StatusNotFound, "not_found", "organization not found")
		} else {
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to list members")
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

// RemoveMember godoc
// @Summary Удалить участника из организации
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID организации"
// @Param user_id path string true "ID пользователя"
// @Success 204 "No Content"
// @Failure 400 {object} organization.ErrorResponse
// @Failure 401 {object} organization.ErrorResponse
// @Failure 403 {object} organization.ErrorResponse
// @Failure 404 {object} organization.ErrorResponse
// @Failure 500 {object} organization.ErrorResponse
// @Router /api/v1/organizations/{id}/members/{user_id} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
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

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid user_id")
		return
	}

	err = h.service.RemoveMember(c.Request.Context(), orgID, userID)
	if err != nil {
		switch err {
		case ErrNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "organization not found")
		case ErrArchived:
			response.Error(c, http.StatusForbidden, "archived", "cannot remove members from archived organization")
		case ErrUserNotInOrg:
			response.Error(c, http.StatusNotFound, "not_found", "user is not a member of this organization")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to remove member")
		}
		return
	}

	c.Status(http.StatusNoContent)
}
