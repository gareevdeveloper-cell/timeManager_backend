package server

import (
	"github.com/gin-gonic/gin"

	"testik/internal/auth"
	"testik/internal/http/middleware"
	"testik/internal/project"
)

// registerProjectRoutes регистрирует маршруты проектов и задач.
func registerProjectRoutes(v1 *gin.RouterGroup, projectHandler *project.Handler, authService *auth.AuthService) {
	protected := v1.Group("/projects")
	protected.Use(middleware.Auth(authService))
	{
		protected.GET("", projectHandler.ListProjects)
		protected.POST("", projectHandler.CreateProject)
		protected.GET("/:projectId", projectHandler.GetProject)
		protected.GET("/:projectId/members", projectHandler.GetProjectMembers)
		protected.GET("/:projectId/statuses", projectHandler.ListStatuses)
		protected.POST("/:projectId/statuses", projectHandler.CreateStatus)
		protected.GET("/:projectId/board", projectHandler.GetBoard)
		protected.GET("/:projectId/tasks", projectHandler.ListTasks)
		protected.POST("/:projectId/tasks", projectHandler.CreateTask)
	}

	statuses := v1.Group("/projects/statuses")
	statuses.Use(middleware.Auth(authService))
	{
		statuses.PATCH("/:statusId", projectHandler.UpdateStatus)
		statuses.DELETE("/:statusId", projectHandler.DeleteStatus)
	}

	tasks := v1.Group("/tasks")
	tasks.Use(middleware.Auth(authService))
	{
		tasks.GET("/:taskId", projectHandler.GetTask)
		tasks.PATCH("/:taskId", projectHandler.UpdateTask)
		tasks.DELETE("/:taskId", projectHandler.DeleteTask)
	}
}
