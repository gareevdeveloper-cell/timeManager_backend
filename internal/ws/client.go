package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Client — WebSocket-клиент.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID string
	send   chan []byte
	channels map[string]struct{}
}

// readPump читает сообщения от клиента (подписки на каналы).
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws: read error: %v", err)
			}
			return
		}
		c.handleMessage(msg)
	}
}

// handleMessage обрабатывает входящее сообщение (subscribe/unsubscribe).
func (c *Client) handleMessage(msg []byte) {
	var req ClientMessage
	if err := json.Unmarshal(msg, &req); err != nil {
		return
	}
	switch req.Type {
	case "subscribe":
		if ch, ok := req.Channel.(string); ok && ch != "" {
			c.hub.Subscribe(c, ch)
		}
	case "unsubscribe":
		if ch, ok := req.Channel.(string); ok && ch != "" {
			c.hub.Unsubscribe(c, ch)
		}
	}
}

// writePump отправляет сообщения клиенту.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ClientMessage — сообщение от клиента (subscribe/unsubscribe).
type ClientMessage struct {
	Type    string      `json:"type"`
	Channel interface{} `json:"channel"`
}
