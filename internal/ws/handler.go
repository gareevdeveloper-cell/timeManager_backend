package ws

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// TokenValidator — интерфейс валидации JWT.
type TokenValidator interface {
	ValidateToken(tokenString string) (userID string, err error)
}

// Handler — WebSocket handler.
type Handler struct {
	hub    *Hub
	validator TokenValidator
}

// NewHandler создаёт WebSocket handler.
func NewHandler(hub *Hub, validator TokenValidator) *Handler {
	return &Handler{hub: hub, validator: validator}
}

// ServeWS обрабатывает upgrade на WebSocket.
func (h *Handler) ServeWS(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}
	}
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}

	userID, err := h.validator.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws: upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:     h.hub,
		conn:    conn,
		userID:  userID,
		send:    make(chan []byte, 256),
		channels: make(map[string]struct{}),
	}
	h.hub.register <- client

	go client.writePump()
	go client.readPump()
}
