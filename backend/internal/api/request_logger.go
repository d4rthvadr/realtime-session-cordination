package api

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"

	"realtime-session-coordination/backend/internal/logging"

	"github.com/gin-gonic/gin"
)

const requestIDContextKey = "requestID"
const requestIDHeader = "X-Request-ID"

func RequestIDFromContext(c *gin.Context) string {
	if c == nil {
		return ""
	}

	if value, ok := c.Get(requestIDContextKey); ok {
		if id, ok := value.(string); ok {
			return id
		}
	}

	return ""
}

func getOrCreateRequestID(c *gin.Context) string {
	requestID := strings.TrimSpace(c.GetHeader(requestIDHeader))
	if requestID != "" {
		return requestID
	}

	return newRequestID()
}

func newRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "req_fallback_" + time.Now().UTC().Format("20060102150405.000000000")
	}

	return "req_" + hex.EncodeToString(buf)
}

func RequestLoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = logging.Default()
	}
	logger = logger.With("component", "http_access")

	return func(c *gin.Context) {
		requestID := getOrCreateRequestID(c)
		c.Set(requestIDContextKey, requestID)
		c.Header(requestIDHeader, requestID)

		start := time.Now()
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []any{
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"response_bytes", c.Writer.Size(),
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}

		switch {
		case status >= 500:
			logger.Error("http_request", attrs...)
		case status >= 400:
			logger.Warn("http_request", attrs...)
		default:
			logger.Info("http_request", attrs...)
		}
	}
}