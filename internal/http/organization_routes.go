package server

import (
	"github.com/gin-gonic/gin"

	"testik/internal/auth"
	"testik/internal/http/middleware"
	"testik/internal/organization"
)

// registerOrganizationRoutes регистрирует маршруты организаций.
func registerOrganizationRoutes(v1 *gin.RouterGroup, orgHandler *organization.Handler, authService *auth.AuthService) {
	protected := v1.Group("/organizations")
	protected.Use(middleware.Auth(authService))
	{
		protected.GET("", orgHandler.ListMyOrganizations)
		protected.POST("", orgHandler.Create)
		protected.GET("/:id", orgHandler.GetByID)
		protected.GET("/:id/members", orgHandler.GetMembers)
		protected.PATCH("/:id", orgHandler.Update)
		protected.PUT("/:id/avatar", orgHandler.SetAvatar)
		protected.POST("/:id/archive", orgHandler.Archive)
		protected.POST("/:id/members", orgHandler.AddMember)
		protected.DELETE("/:id/members/:user_id", orgHandler.RemoveMember)
	}
}
