package server

import (
	"github.com/gin-gonic/gin"

	"testik/internal/auth"
	"testik/internal/http/middleware"
	"testik/internal/user"
)

// registerAuthRoutes регистрирует маршруты аутентификации.
func registerAuthRoutes(v1 *gin.RouterGroup, authHandler *auth.Handler, authService *auth.AuthService, userHandler *user.Handler) {
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.GET("/:provider/redirect", authHandler.OAuthRedirect)
		authGroup.GET("/:provider/callback", authHandler.OAuthCallback)
	}

	protected := v1.Group("")
	protected.Use(middleware.Auth(authService))
	{
		protected.GET("/users/me", userHandler.Me)
		protected.PATCH("/users/me", userHandler.UpdateProfile)
		protected.GET("/users/me/tasks", userHandler.ListMyTasks)
		protected.PUT("/users/me/current-task", userHandler.SetCurrentTask)
		protected.PUT("/users/me/avatar", userHandler.SetAvatar)
		protected.PUT("/users/me/work-status", userHandler.UpdateWorkStatus)
		protected.GET("/users/me/work-status/history", userHandler.GetWorkStatusHistory)
	}
}
