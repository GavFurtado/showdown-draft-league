package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	loggerKey    = "slogger"
	requestIDKey = "request_id"
)

func requestIDFromRequest(c *gin.Context) string {
	if id := c.GetHeader("X-Request-ID"); id != "" {
		return id
	}
	return uuid.New().String()
}

func RequestContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := requestIDFromRequest(c)
		path := c.Request.URL.Path
		method := c.Request.Method

		logger := slog.Default().With(
			slog.String("request_id", reqID),
			slog.String("method", method),
			slog.String("path", path),
		)

		c.Set(loggerKey, logger)
		c.Set(requestIDKey, reqID)
		c.Header("X-Request-ID", reqID)

		c.Next()
	}
}

func GetLogger(c *gin.Context) *slog.Logger {
	if l, exists := c.Get(loggerKey); exists {
		if logger, ok := l.(*slog.Logger); ok {
			return logger
		}
	}
	return slog.Default()
}

func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(requestIDKey); exists {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}

func SetUserOnLogger(c *gin.Context, userID string, role string) {
	logger := GetLogger(c)
	logger = logger.With(
		slog.String("user_id", userID),
		slog.String("role", role),
	)
	c.Set(loggerKey, logger)
}
