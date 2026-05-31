package ws

import (
	"log/slog"
	"sync"
	"time"

	"realtime-session-coordination/backend/internal/logging"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu      sync.Mutex
	clients map[string]map[*websocket.Conn]struct{}
	logger  *slog.Logger
}

func NewHub(logger *slog.Logger) *Hub {
	if logger == nil {
		logger = logging.Default()
	}
	logger = logger.With("component", "ws_hub")

	return &Hub{
		clients: make(map[string]map[*websocket.Conn]struct{}),
		logger:  logger,
	}
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
	h.BroadcastWithRequestID(sessionID, payload, "")
}

func (h *Hub) BroadcastWithRequestID(sessionID string, payload any, requestID string) {
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
			if requestID != "" {
				h.logger.Error("ws_write_failed", "session_id", sessionID, "request_id", requestID, "error", err)
			} else {
				h.logger.Error("ws_write_failed", "session_id", sessionID, "error", err)
			}
			h.Unregister(sessionID, conn)
		}
	}
}
