package ws

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu      sync.Mutex
	clients map[string]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[*websocket.Conn]struct{})}
}

func (h *Hub) Register(sessionID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[sessionID]; !ok {
		h.clients[sessionID] = make(map[*websocket.Conn]struct{})
	}
	h.clients[sessionID][conn] = struct{}{}
}

func (h *Hub) Unregister(sessionID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[sessionID]; ok {
		delete(clients, conn)
		if len(clients) == 0 {
			delete(h.clients, sessionID)
		}
	}
	_ = conn.Close()
}

func (h *Hub) Broadcast(sessionID string, payload any) {
	h.mu.Lock()
	clients, ok := h.clients[sessionID]
	if !ok {
		h.mu.Unlock()
		return
	}

	conns := make([]*websocket.Conn, 0, len(clients))
	for conn := range clients {
		conns = append(conns, conn)
	}
	h.mu.Unlock()

	for _, conn := range conns {
		_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := conn.WriteJSON(payload); err != nil {
			log.Printf("ws write failed for session=%s: %v", sessionID, err)
			h.Unregister(sessionID, conn)
		}
	}
}
