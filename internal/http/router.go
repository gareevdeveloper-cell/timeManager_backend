package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"testik/internal/auth"
	"testik/internal/files"
	"testik/internal/organization"
	"testik/internal/project"
	"testik/internal/team"
	"testik/internal/user"
	"testik/internal/ws"
)

// Router — HTTP-роутер приложения.
type Router struct {
	engine *gin.Engine
}

// RouterDeps — зависимости для роутера.
type RouterDeps struct {
	AuthHandler    *auth.Handler
	AuthService    *auth.AuthService
	UserHandler    *user.Handler
	OrgHandler     *organization.Handler
	TeamHandler    *team.Handler
	ProjectHandler *project.Handler
	FilesHandler   *files.Handler
	WsHandler      *ws.Handler
}

// NewRouter создаёт роутер и регистрирует все маршруты.
func NewRouter(deps RouterDeps) *Router {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	registerAuthRoutes(v1, deps.AuthHandler, deps.AuthService, deps.UserHandler)
	registerFilesRoutes(v1, deps.FilesHandler)
	registerOrganizationRoutes(v1, deps.OrgHandler, deps.AuthService)
	registerTeamRoutes(v1, deps.TeamHandler, deps.ProjectHandler, deps.AuthService)
	registerProjectRoutes(v1, deps.ProjectHandler, deps.AuthService)
	if deps.WsHandler != nil {
		registerWSRoutes(v1, deps.WsHandler)
	}

	return &Router{engine: r}
}

// Engine возвращает *gin.Engine для использования в HTTP-сервере.
func (rt *Router) Engine() *gin.Engine {
	return rt.engine
}
