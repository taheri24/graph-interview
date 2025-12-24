package middleware

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDKey is the context key for request ID
type RequestIDKey string

const requestIDKey RequestIDKey = "requestID"

// LoggerKey is the context key for logger
type LoggerKey string

const loggerKey LoggerKey = "logger"

// RequestIDMiddleware adds a unique request ID to each request for tracing
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate or get existing request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in context and response header
		c.Set(string(requestIDKey), requestID)
		c.Header("X-Request-ID", requestID)

		// Create a logger with request ID attribute
		logger := slog.With(slog.String("requestID", requestID))

		// Also add to context for downstream use
		ctx := context.WithValue(c.Request.Context(), requestIDKey, requestID)
		ctx = context.WithValue(ctx, loggerKey, logger)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetRequestID retrieves the request ID from the Gin context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(string(requestIDKey)); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// GetRequestIDFromContext retrieves the request ID from the context
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID := ctx.Value(requestIDKey); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// GetLoggerFromContext retrieves the logger from the context
func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	if logger := ctx.Value(loggerKey); logger != nil {
		if l, ok := logger.(*slog.Logger); ok {
			return l
		}
	}
	return slog.Default()
}
