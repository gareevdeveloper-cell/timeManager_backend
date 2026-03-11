package files

import (
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"

	"testik/pkg/storage"
)

// Handler — HTTP-обработчики для доступа к файлам из хранилища.
type Handler struct {
	storage storage.Storage
}

// NewHandler создаёт handler.
func NewHandler(st storage.Storage) *Handler {
	return &Handler{storage: st}
}

// Разрешённые префиксы путей (аватарки и т.п.).
var allowedPrefixes = []string{"users/", "organizations/", "teams/"}

func isPathAllowed(p string) bool {
	p = strings.TrimPrefix(p, "/")
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

// GetFile godoc
// @Summary Получить файл из хранилища
// @Description Возвращает файл (картинку) из MinIO по частичному пути. Путь: users/{id}/avatar.jpg, organizations/{id}/avatar.jpg, teams/{id}/avatar.jpg
// @Tags files
// @Produce octet-stream
// @Param path path string true "Частичный путь к файлу (например users/uuid/avatar.jpg)"
// @Success 200 "Файл"
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/files/{path} [get]
func (h *Handler) GetFile(c *gin.Context) {
	rawPath := c.Param("path")
	if rawPath == "" {
		responseError(c, http.StatusBadRequest, "validation_error", "path is required")
		return
	}
	// Убираем ведущий слэш, нормализуем (Gin передаёт path с ведущим /)
	filePath := strings.TrimPrefix(path.Clean("/"+rawPath), "/")
	if strings.Contains(filePath, "..") {
		responseError(c, http.StatusBadRequest, "validation_error", "invalid path")
		return
	}
	if !isPathAllowed(filePath) {
		responseError(c, http.StatusForbidden, "forbidden", "path not allowed")
		return
	}

	reader, info, err := h.storage.Get(c.Request.Context(), filePath)
	if err != nil {
		responseError(c, http.StatusNotFound, "not_found", "file not found")
		return
	}
	defer reader.Close()

	c.DataFromReader(http.StatusOK, info.Size, info.ContentType, reader, nil)
}

func responseError(c *gin.Context, code int, errCode, message string) {
	c.JSON(code, gin.H{
		"error": gin.H{"code": errCode, "message": message},
	})
}
