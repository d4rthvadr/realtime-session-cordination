package api

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"realtime-session-coordination/backend/internal/session"
	"realtime-session-coordination/backend/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	manager  *session.Manager
	hub      *ws.Hub
	upgrader websocket.Upgrader
}

func NewHandler(manager *session.Manager, hub *ws.Hub) *Handler {
	return &Handler{
		manager: manager,
		hub:     hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/sessions", h.createSession)
		apiV1.GET("/sessions", h.listSessions)
		apiV1.GET("/sessions/:id", h.getSession)
		apiV1.POST("/sessions/:id/start", h.startSession)
		apiV1.POST("/sessions/:id/pause", h.pauseSession)
		apiV1.POST("/sessions/:id/resume", h.resumeSession)
		apiV1.POST("/sessions/:id/end", h.endSession)
		apiV1.POST("/sessions/:id/adjust-time", h.adjustTime)
	}

	router.GET("/ws/sessions/:id", h.sessionSocket)
}

func (h *Handler) createSession(c *gin.Context) {
	var input session.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	snap, token, err := h.manager.Create(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session":      snap,
		"controlToken": token,
		"viewerPath":   "/sessions/" + snap.ID,
	})
}

func (h *Handler) getSession(c *gin.Context) {
	snap, err := h.manager.GetSnapshot(c.Param("id"))
	if err != nil {
		h.writeDomainErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"session": snap})
}

func (h *Handler) listSessions(c *gin.Context) {
	snapshots, err := h.manager.ListSnapshots()
	if err != nil {
		h.writeDomainErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": snapshots})
}

func (h *Handler) startSession(c *gin.Context) {
	id := c.Param("id")
	if !h.authorizeControl(c, id) {
		return
	}

	event, err := h.manager.Start(id)
	if err != nil {
		h.writeDomainErr(c, err)
		return
	}
	c.JSON(http.StatusOK, event)
	h.hub.Broadcast(id, event)
}

func (h *Handler) pauseSession(c *gin.Context) {
	id := c.Param("id")
	if !h.authorizeControl(c, id) {
		return
	}

	event, err := h.manager.Pause(id)
	if err != nil {
		h.writeDomainErr(c, err)
		return
	}
	c.JSON(http.StatusOK, event)
	h.hub.Broadcast(id, event)
}

func (h *Handler) resumeSession(c *gin.Context) {
	id := c.Param("id")
	if !h.authorizeControl(c, id) {
		return
	}

	event, err := h.manager.Resume(id)
	if err != nil {
		h.writeDomainErr(c, err)
		return
	}
	c.JSON(http.StatusOK, event)
	h.hub.Broadcast(id, event)
}

func (h *Handler) endSession(c *gin.Context) {
	id := c.Param("id")
	if !h.authorizeControl(c, id) {
		return
	}

	event, err := h.manager.End(id)
	if err != nil {
		h.writeDomainErr(c, err)
		return
	}
	c.JSON(http.StatusOK, event)
	h.hub.Broadcast(id, event)
}

type adjustTimeBody struct {
	DeltaSeconds int `json:"deltaSeconds"`
}

func (h *Handler) adjustTime(c *gin.Context) {
	id := c.Param("id")
	if !h.authorizeControl(c, id) {
		return
	}

	var body adjustTimeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	event, err := h.manager.AdjustTime(id, body.DeltaSeconds)
	if err != nil {
		h.writeDomainErr(c, err)
		return
	}
	c.JSON(http.StatusOK, event)
	h.hub.Broadcast(id, event)
}

func (h *Handler) sessionSocket(c *gin.Context) {
	sessionID := c.Param("id")
	if !h.manager.SessionExists(sessionID) {
		c.JSON(http.StatusNotFound, gin.H{"error": session.ErrNotFound.Error()})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws upgrade failed: %v", err)
		return
	}

	h.hub.Register(sessionID, conn)
	defer h.hub.Unregister(sessionID, conn)

	snap, err := h.manager.GetSnapshot(sessionID)
	if err == nil {
		h.hub.Broadcast(sessionID, session.Event{Type: "SESSION_SNAPSHOT", Session: snap})
	}

	for {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *Handler) authorizeControl(c *gin.Context, sessionID string) bool {
	token := strings.TrimSpace(c.GetHeader("X-Control-Token"))
	if token == "" {
		token = strings.TrimSpace(c.Query("controlToken"))
	}

	err := h.manager.ValidateControlToken(sessionID, token)
	if err == nil {
		return true
	}

	h.writeDomainErr(c, err)
	return false
}

func (h *Handler) writeDomainErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, session.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, session.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, session.ErrInvalidTransition):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func CORSMiddleware() gin.HandlerFunc {
	allowed := os.Getenv("CORS_ALLOW_ORIGIN")
	if allowed == "" {
		allowed = "*"
	}

	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowed)
		c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-Control-Token")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
