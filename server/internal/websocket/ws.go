package websocket

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Hub keeps track of all connected users
type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID][]*websocket.Conn // one user can have multiple tabs open
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uuid.UUID][]*websocket.Conn),
	}
}

func (h *Hub) Register(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[userID] = append(h.clients[userID], conn)
}

func (h *Hub) Unregister(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := h.clients[userID]
	for i, c := range conns {
		if c == conn {
			h.clients[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

func (h *Hub) SendToUser(userID uuid.UUID, payload any) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, conn := range h.clients[userID] {
		conn.WriteJSON(payload)
	}
}
