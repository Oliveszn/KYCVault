package websocket

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type client struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (c *client) send(payload any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(payload)
}

// Hub keeps track of all connected users
type Hub struct {
	mu sync.RWMutex
	// clients map[uuid.UUID][]*websocket.Conn // one user can have multiple tabs open
	clients map[uuid.UUID][]*client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uuid.UUID][]*client),
	}
}

func (h *Hub) Register(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// h.clients[userID] = append(h.clients[userID], conn)
	c := &client{conn: conn}
	h.clients[userID] = append(h.clients[userID], c)
}

func (h *Hub) Unregister(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.clients[userID]

	for i, c := range conns {
		if c.conn == conn {
			h.clients[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

func (h *Hub) SendToUser(userID uuid.UUID, payload any) {
	h.mu.RLock()
	conns := append([]*client(nil), h.clients[userID]...)
	h.mu.RUnlock()

	for _, c := range conns {
		if err := c.send(payload); err != nil {
			go h.Unregister(userID, c.conn)
		}
	}
}
