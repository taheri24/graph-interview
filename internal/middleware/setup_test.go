package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetupGlobalMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Setup global middleware
	SetupGlobalMiddleware(router)

	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "middleware test"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check that request ID header is set (from RequestIDMiddleware)
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)
}

func TestSetupMetricsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Setup metrics endpoint
	SetupMetricsEndpoint(router)

	req, _ := http.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Metrics endpoint should return 200 with Prometheus metrics format
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
}

func TestSetupMetricsEndpoint_WithGlobalMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Setup both global middleware and metrics endpoint
	SetupGlobalMiddleware(router)
	SetupMetricsEndpoint(router)

	req, _ := http.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still work with global middleware
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
}
