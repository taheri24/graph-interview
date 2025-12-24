package middleware

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		existingID     string
		expectedHeader bool
	}{
		{
			name:           "generates new request ID",
			existingID:     "",
			expectedHeader: true,
		},
		{
			name:           "uses existing request ID",
			existingID:     uuid.New().String(),
			expectedHeader: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RequestIDMiddleware())

			router.GET("/test", func(c *gin.Context) {
				// Check that logger is set
				logger := GetLoggerFromContext(c.Request.Context())
				if logger == nil {
					t.Error("Expected logger to be set in context")
				}
				c.JSON(200, gin.H{"message": "test"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingID != "" {
				req.Header.Set("X-Request-ID", tt.existingID)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check response header
			responseID := w.Header().Get("X-Request-ID")
			if !tt.expectedHeader && responseID != "" {
				t.Errorf("Expected no X-Request-ID header, got %s", responseID)
			}
			if tt.expectedHeader && responseID == "" {
				t.Error("Expected X-Request-ID header to be set")
			}

			// If existing ID was provided, it should be preserved
			if tt.existingID != "" && responseID != tt.existingID {
				t.Errorf("Expected request ID to be %s, got %s", tt.existingID, responseID)
			}

			// If no existing ID, a new UUID should be generated
			if tt.existingID == "" && responseID != "" {
				if _, err := uuid.Parse(responseID); err != nil {
					t.Errorf("Expected valid UUID, got %s", responseID)
				}
			}
		})
	}
}

func TestGetRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		requestID   string
		expectError bool
	}{
		{
			name:        "retrieves existing request ID",
			requestID:   uuid.New().String(),
			expectError: false,
		},
		{
			name:        "returns empty string when no request ID",
			requestID:   "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			if tt.requestID != "" {
				c.Set(string(requestIDKey), tt.requestID)
			}

			retrievedID := GetRequestID(c)

			if tt.expectError && retrievedID != "" {
				t.Errorf("Expected empty string, got %s", retrievedID)
			}
			if !tt.expectError && retrievedID != tt.requestID {
				t.Errorf("Expected request ID to be %s, got %s", tt.requestID, retrievedID)
			}
		})
	}
}

func TestGetRequestIDFromContext(t *testing.T) {
	tests := []struct {
		name        string
		requestID   string
		expectError bool
	}{
		{
			name:        "retrieves existing request ID from context",
			requestID:   uuid.New().String(),
			expectError: false,
		},
		{
			name:        "returns empty string when no request ID in context",
			requestID:   "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context

			if tt.requestID != "" {
				ctx = context.WithValue(context.Background(), requestIDKey, tt.requestID)
			} else {
				ctx = context.Background()
			}

			retrievedID := GetRequestIDFromContext(ctx)

			if tt.expectError && retrievedID != "" {
				t.Errorf("Expected empty string, got %s", retrievedID)
			}
			if !tt.expectError && retrievedID != tt.requestID {
				t.Errorf("Expected request ID to be %s, got %s", tt.requestID, retrievedID)
			}
		})
	}
}

func TestGetLoggerFromContext(t *testing.T) {
	tests := []struct {
		name      string
		requestID string
		hasLogger bool
	}{
		{
			name:      "retrieves logger with request ID from context",
			requestID: uuid.New().String(),
			hasLogger: true,
		},
		{
			name:      "returns default logger when no logger in context",
			requestID: "",
			hasLogger: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context

			if tt.hasLogger {
				logger := slog.With(slog.String("requestID", tt.requestID))
				ctx = context.WithValue(context.Background(), loggerKey, logger)
			} else {
				ctx = context.Background()
			}

			retrievedLogger := GetLoggerFromContext(ctx)

			if tt.hasLogger {
				if retrievedLogger == nil {
					t.Error("Expected logger, got nil")
				}
				// Note: Testing the exact attributes might require a custom handler or buffer
				// For now, just check that we get a logger
			} else {
				if retrievedLogger != slog.Default() {
					t.Error("Expected default logger when no logger in context")
				}
			}
		})
	}
}
