package api

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"realtime-session-coordination/backend/internal/auth"
	"realtime-session-coordination/backend/internal/logging"
	"realtime-session-coordination/backend/internal/programitem"
	"realtime-session-coordination/backend/internal/session"
	"realtime-session-coordination/backend/internal/user"
	"realtime-session-coordination/backend/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	manager            *session.Manager
	programItemManager *programitem.Manager
	hub                *ws.Hub
	authService        *auth.Service
	logger             *slog.Logger
	upgrader           websocket.Upgrader
}

func NewHandler(manager *session.Manager, programItemManager *programitem.Manager, hub *ws.Hub, authService *auth.Service, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = logging.Default()
	}
	logger = logger.With("component", "api_handler")

	return &Handler{
		manager:            manager,
		programItemManager: programItemManager,
		hub:                hub,
		authService:        authService,
		logger:             logger,
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
		apiV1.POST("/auth/guest", h.createGuest)

		protected := apiV1.Group("")
		protected.Use(h.requireAuth())
		protected.POST("/sessions", h.createSession)
		protected.GET("/sessions", h.listSessions)
		protected.GET("/sessions/:id/program-items", h.listProgramItems)
		protected.POST("/sessions/:id/program-items", h.createProgramItem)
		protected.POST("/sessions/:id/program-items/reorder", h.reorderProgramItems)
		protected.POST("/sessions/:id/start", h.startSession)
		protected.POST("/sessions/:id/pause", h.pauseSession)
		protected.POST("/sessions/:id/resume", h.resumeSession)
		protected.POST("/sessions/:id/end", h.endSession)
		protected.POST("/sessions/:id/adjust-time", h.adjustTime)
		protected.PATCH("/program-items/:itemId", h.updateProgramItem)
		protected.POST("/program-items/:itemId/cancel", h.cancelProgramItem)
		protected.POST("/program-items/:itemId/start", h.startProgramItem)
		protected.POST("/program-items/:itemId/end", h.endProgramItem)

		apiV1.GET("/sessions/:id", h.getSession)
		apiV1.GET("/sessions/:id/current-program-item", h.getCurrentProgramItem)
	}

	router.GET("/ws/sessions/:id", h.sessionSocket)
}

func (h *Handler) createGuest(c *gin.Context) {
	if h.authService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "auth service not configured"})
		return
	}

	u, token, err := h.authService.CreateGuest()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create guest user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user":  user.ToSnapshot(u),
	})
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

func (h *Handler) listProgramItems(c *gin.Context) {
	if h.programItemManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "program item manager not configured"})
		return
	}

	items, err := h.programItemManager.ListSnapshots(c.Param("id"))
	if err != nil {
		h.writeProgramItemErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"programItems": items})
}

func (h *Handler) getCurrentProgramItem(c *gin.Context) {
	if h.programItemManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "program item manager not configured"})
		return
	}

	item, nextItem, err := h.programItemManager.CurrentAndNextSnapshots(c.Param("id"), time.Now().UTC())
	if err != nil {
		h.writeProgramItemErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"programItem": item, "nextProgramItem": nextItem})
}

type createProgramItemBody struct {
	Title                   string         `json:"title"`
	Type                    string         `json:"type"`
	HostName                string         `json:"hostName"`
	ScheduledStart          time.Time      `json:"scheduledStart"`
	ScheduledEnd            time.Time      `json:"scheduledEnd"`
	ExpectedDurationMinutes int            `json:"expectedDurationMinutes"`
	Position                int            `json:"position"`
	Location                string         `json:"location"`
	Metadata                map[string]any `json:"metadata"`
}

func (h *Handler) createProgramItem(c *gin.Context) {
	if h.programItemManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "program item manager not configured"})
		return
	}

	sessionID := c.Param("id")
	if !h.authorizeControl(c, sessionID) {
		return
	}

	var body createProgramItemBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if !programitem.IsAllowedType(body.Type) {
		c.JSON(http.StatusBadRequest, gin.H{"error": programitem.ErrInvalidType.Error()})
		return
	}

	snap, err := h.programItemManager.Create(programitem.CreateInput{
		SessionID:               sessionID,
		Title:                   body.Title,
		Type:                    body.Type,
		HostName:                body.HostName,
		ScheduledStart:          body.ScheduledStart,
		ScheduledEnd:            body.ScheduledEnd,
		ExpectedDurationMinutes: body.ExpectedDurationMinutes,
		Position:                body.Position,
		Location:                body.Location,
		Metadata:                body.Metadata,
	})
	if err != nil {
		h.writeProgramItemErr(c, err)
		return
	}

	h.broadcastProgramItemEvent(sessionID, programitem.Event{
		Type:        programitem.EventCreated,
		SessionID:   sessionID,
		ProgramItem: &snap,
	}, c)

	c.JSON(http.StatusCreated, gin.H{"programItem": snap})
}

type updateProgramItemBody struct {
	Title                   *string         `json:"title"`
	Type                    *string         `json:"type"`
	Status                  *string         `json:"status"`
	HostName                *string         `json:"hostName"`
	ScheduledStart          *time.Time      `json:"scheduledStart"`
	ScheduledEnd            *time.Time      `json:"scheduledEnd"`
	ExpectedDurationMinutes *int            `json:"expectedDurationMinutes"`
	Position                *int            `json:"position"`
	Location                *string         `json:"location"`
	Metadata                *map[string]any `json:"metadata"`
}

func (h *Handler) updateProgramItem(c *gin.Context) {
	if h.programItemManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "program item manager not configured"})
		return
	}

	itemID := c.Param("itemId")
	item, err := h.programItemManager.GetSnapshot(itemID)
	if err != nil {
		h.writeProgramItemErr(c, err)
		return
	}
	if !h.authorizeControl(c, item.SessionID) {
		return
	}

	var body updateProgramItemBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if body.Type != nil && !programitem.IsAllowedType(*body.Type) {
		c.JSON(http.StatusBadRequest, gin.H{"error": programitem.ErrInvalidType.Error()})
		return
	}

	snap, err := h.programItemManager.Update(itemID, programitem.UpdateInput{
		Title:                   body.Title,
		Type:                    body.Type,
		Status:                  body.Status,
		HostName:                body.HostName,
		ScheduledStart:          body.ScheduledStart,
		ScheduledEnd:            body.ScheduledEnd,
		ExpectedDurationMinutes: body.ExpectedDurationMinutes,
		Position:                body.Position,
		Location:                body.Location,
		Metadata:                body.Metadata,
	})
	if err != nil {
		h.writeProgramItemErr(c, err)
		return
	}

	h.broadcastProgramItemEvent(item.SessionID, programitem.Event{
		Type:        programitem.EventUpdated,
		SessionID:   item.SessionID,
		ProgramItem: &snap,
	}, c)

	c.JSON(http.StatusOK, gin.H{"programItem": snap})
}

func (h *Handler) cancelProgramItem(c *gin.Context) {
	if h.programItemManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "program item manager not configured"})
		return
	}

	itemID := c.Param("itemId")
	item, err := h.programItemManager.GetSnapshot(itemID)
	if err != nil {
		h.writeProgramItemErr(c, err)
		return
	}
	if !h.authorizeControl(c, item.SessionID) {
		return
	}

	snap, err := h.programItemManager.Cancel(itemID)
	if err != nil {
		h.writeProgramItemErr(c, err)
		return
	}

	h.broadcastProgramItemEvent(item.SessionID, programitem.Event{
		Type:        programitem.EventCanceled,
		SessionID:   item.SessionID,
		ProgramItem: &snap,
	}, c)

	c.JSON(http.StatusOK, gin.H{"programItem": snap})
}

func (h *Handler) startProgramItem(c *gin.Context) {
	if h.programItemManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "program item manager not configured"})
		return
	}

	itemID := c.Param("itemId")
	item, err := h.programItemManager.GetSnapshot(itemID)
	if err != nil {
		h.writeProgramItemErr(c, err)
		return
	}
	if !h.authorizeControl(c, item.SessionID) {
		return
	}
	if !h.ensureProgramItemRuntimeAllowed(c, item.SessionID) {
		return
	}

	snap, err := h.programItemManager.Start(itemID)
	if err != nil {
		h.writeProgramItemErr(c, err)
		return
	}

	_, nextSnap, nextErr := h.programItemManager.CurrentAndNextSnapshots(item.SessionID, time.Now().UTC())
	if nextErr != nil {
		h.writeProgramItemErr(c, nextErr)
		return
	}

	h.broadcastProgramItemEvent(item.SessionID, programitem.Event{
		Type:            programitem.EventStarted,
		SessionID:       item.SessionID,
		ProgramItem:     &snap,
		NextProgramItem: nextSnap,
	}, c)

	c.JSON(http.StatusOK, gin.H{"programItem": snap, "nextProgramItem": nextSnap})
}

func (h *Handler) endProgramItem(c *gin.Context) {
	if h.programItemManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "program item manager not configured"})
		return
	}

	itemID := c.Param("itemId")
	item, err := h.programItemManager.GetSnapshot(itemID)
	if err != nil {
		h.writeProgramItemErr(c, err)
		return
	}
	if !h.authorizeControl(c, item.SessionID) {
		return
	}
	if !h.ensureProgramItemRuntimeAllowed(c, item.SessionID) {
		return
	}

	snap, err := h.programItemManager.End(itemID)
	if err != nil {
		h.writeProgramItemErr(c, err)
		return
	}

	_, nextSnap, nextErr := h.programItemManager.CurrentAndNextSnapshots(item.SessionID, time.Now().UTC())
	if nextErr != nil {
		h.writeProgramItemErr(c, nextErr)
		return
	}

	h.broadcastProgramItemEvent(item.SessionID, programitem.Event{
		Type:            programitem.EventEnded,
		SessionID:       item.SessionID,
		ProgramItem:     &snap,
		NextProgramItem: nextSnap,
	}, c)

	c.JSON(http.StatusOK, gin.H{"programItem": snap, "nextProgramItem": nextSnap})
}

type reorderProgramItemsBody struct {
	Items []programitem.ReorderItem `json:"items"`
}

func (h *Handler) reorderProgramItems(c *gin.Context) {
	if h.programItemManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "program item manager not configured"})
		return
	}

	sessionID := c.Param("id")
	if !h.authorizeControl(c, sessionID) {
		return
	}

	var body reorderProgramItemsBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	items, err := h.programItemManager.Reorder(sessionID, body.Items)
	if err != nil {
		h.writeProgramItemErr(c, err)
		return
	}

	h.broadcastProgramItemEvent(sessionID, programitem.Event{
		Type:         programitem.EventReordered,
		SessionID:    sessionID,
		ProgramItems: items,
	}, c)

	c.JSON(http.StatusOK, gin.H{"programItems": items})
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
	h.hub.BroadcastWithRequestID(id, event, RequestIDFromContext(c))
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
	h.hub.BroadcastWithRequestID(id, event, RequestIDFromContext(c))
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
	h.hub.BroadcastWithRequestID(id, event, RequestIDFromContext(c))
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
	h.hub.BroadcastWithRequestID(id, event, RequestIDFromContext(c))
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
	h.hub.BroadcastWithRequestID(id, event, RequestIDFromContext(c))
}

func (h *Handler) sessionSocket(c *gin.Context) {
	sessionID := c.Param("id")
	if !h.manager.SessionExists(sessionID) {
		c.JSON(http.StatusNotFound, gin.H{"error": session.ErrNotFound.Error()})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("ws_upgrade_failed", "error", err, "session_id", sessionID, "request_id", RequestIDFromContext(c))
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

func (h *Handler) requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.authService == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "auth service not configured"})
			c.Abort()
			return
		}

		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": auth.ErrUnauthorized.Error()})
			c.Abort()
			return
		}

		rawToken := strings.TrimSpace(authHeader[7:])
		claims, err := h.authService.ValidateToken(rawToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": auth.ErrUnauthorized.Error()})
			c.Abort()
			return
		}

		c.Set("authUserID", claims.Subject)
		c.Set("authUserType", claims.UserType)
		c.Next()
	}
}

func (h *Handler) ensureProgramItemRuntimeAllowed(c *gin.Context, sessionID string) bool {
	snap, err := h.manager.GetSnapshot(sessionID)
	if err != nil {
		h.writeDomainErr(c, err)
		return false
	}

	if snap.Status != session.StatusLive && snap.Status != session.StatusPaused {
		c.JSON(http.StatusConflict, gin.H{"error": "program item runtime transitions require session status LIVE or PAUSED"})
		return false
	}

	return true
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

func (h *Handler) writeProgramItemErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, programitem.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, programitem.ErrSessionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, programitem.ErrInvalidType):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, programitem.ErrOverlap), errors.Is(err, programitem.ErrDuplicatePosition), errors.Is(err, programitem.ErrInvalidRange), errors.Is(err, programitem.ErrInvalidStatus), errors.Is(err, programitem.ErrInvalidStatusTransition), errors.Is(err, programitem.ErrInProgressExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func (h *Handler) broadcastProgramItemEvent(sessionID string, event programitem.Event, c *gin.Context) {
	if h.hub == nil {
		return
	}
	if sessionID == "" {
		return
	}
	h.hub.BroadcastWithRequestID(sessionID, event, RequestIDFromContext(c))
}

func CORSMiddleware() gin.HandlerFunc {
	allowed := os.Getenv("CORS_ALLOW_ORIGIN")
	if allowed == "" {
		allowed = "*"
	}

	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowed)
		c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-Control-Token, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
