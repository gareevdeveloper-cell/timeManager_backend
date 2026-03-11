package ws

import (
	"log"
	"sync"
)

// Hub управляет WebSocket-клиентами и рассылкой сообщений.
type Hub struct {
	mu sync.RWMutex

	// clients по userID для быстрой рассылки конкретному пользователю
	clientsByUser map[string]map[*Client]struct{}

	// clients по каналу (например user:status для статусов всех пользователей)
	clientsByChannel map[string]map[*Client]struct{}

	// регистрация / отмена
	register   chan *Client
	unregister chan *Client

	// broadcast в канал
	broadcast chan *ChannelMessage
}

// ChannelMessage — сообщение для рассылки в канал.
type ChannelMessage struct {
	Channel string
	Payload []byte
}

// NewHub создаёт Hub.
func NewHub() *Hub {
	return &Hub{
		clientsByUser:    make(map[string]map[*Client]struct{}),
		clientsByChannel: make(map[string]map[*Client]struct{}),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		broadcast:        make(chan *ChannelMessage, 256),
	}
}

// Run запускает обработку событий Hub.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if h.clientsByUser[c.userID] == nil {
				h.clientsByUser[c.userID] = make(map[*Client]struct{})
			}
			h.clientsByUser[c.userID][c] = struct{}{}
			h.mu.Unlock()

		case c := <-h.unregister:
			h.mu.Lock()
			if m := h.clientsByUser[c.userID]; m != nil {
				delete(m, c)
				if len(m) == 0 {
					delete(h.clientsByUser, c.userID)
				}
			}
			for ch, m := range h.clientsByChannel {
				if _, ok := m[c]; ok {
					delete(m, c)
					if len(m) == 0 {
						delete(h.clientsByChannel, ch)
					}
				}
			}
			close(c.send)
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.clientsByChannel[msg.Channel]
			if clients == nil {
				h.mu.RUnlock()
				continue
			}
			for c := range clients {
				select {
				case c.send <- msg.Payload:
				default:
					// буфер переполнен — отключаем
					h.mu.RUnlock()
					h.unregister <- c
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast отправляет сообщение всем подписчикам канала.
func (h *Hub) Broadcast(channel string, payload []byte) {
	select {
	case h.broadcast <- &ChannelMessage{Channel: channel, Payload: payload}:
	default:
		log.Printf("ws: broadcast channel full, dropping message to %s", channel)
	}
}

// Subscribe добавляет клиента в канал.
func (h *Hub) Subscribe(c *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clientsByChannel[channel] == nil {
		h.clientsByChannel[channel] = make(map[*Client]struct{})
	}
	h.clientsByChannel[channel][c] = struct{}{}
	if c.channels == nil {
		c.channels = make(map[string]struct{})
	}
	c.channels[channel] = struct{}{}
}

// Unsubscribe удаляет клиента из канала.
func (h *Hub) Unsubscribe(c *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m := h.clientsByChannel[channel]; m != nil {
		delete(m, c)
		if len(m) == 0 {
			delete(h.clientsByChannel, channel)
		}
	}
	delete(c.channels, channel)
}
