package server

import (
	"github.com/gin-gonic/gin"

	"testik/internal/ws"
)

// registerWSRoutes регистрирует WebSocket endpoint.
func registerWSRoutes(v1 *gin.RouterGroup, wsHandler *ws.Handler) {
	v1.GET("/ws", wsHandler.ServeWS)
}
