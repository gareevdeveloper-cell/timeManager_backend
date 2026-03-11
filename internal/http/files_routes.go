package server

import (
	"github.com/gin-gonic/gin"

	"testik/internal/files"
)

// registerFilesRoutes регистрирует маршруты для доступа к файлам из хранилища.
func registerFilesRoutes(v1 *gin.RouterGroup, filesHandler *files.Handler) {
	v1.GET("/files/*path", filesHandler.GetFile)
}
