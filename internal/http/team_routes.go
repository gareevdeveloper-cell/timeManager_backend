package server

import (
	"github.com/gin-gonic/gin"

	"testik/internal/auth"
	"testik/internal/http/middleware"
	"testik/internal/project"
	"testik/internal/team"
)

// registerTeamRoutes регистрирует маршруты команд.
func registerTeamRoutes(v1 *gin.RouterGroup, teamHandler *team.Handler, projectHandler *project.Handler, authService *auth.AuthService) {
	teams := v1.Group("/teams")
	teams.Use(middleware.Auth(authService))
	{
		teams.POST("", teamHandler.Create)
		teams.GET("/:id", teamHandler.GetByID)
		teams.GET("/:id/projects", projectHandler.GetProjectsByTeam)
		teams.PATCH("/:id", teamHandler.Update)
		teams.PUT("/:id/avatar", teamHandler.SetAvatar)
		teams.DELETE("/:id", teamHandler.Delete)
		teams.POST("/:id/members", teamHandler.AddMember)
		teams.DELETE("/:id/members/:user_id", teamHandler.RemoveMember)
		teams.GET("/:id/members", teamHandler.GetMembers)
	}

	orgs := v1.Group("/organizations")
	orgs.Use(middleware.Auth(authService))
	{
		orgs.GET("/:id/teams", teamHandler.GetByOrganization)
	}
}
