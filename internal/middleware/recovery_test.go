package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock FullErrorCapture for testing
var captureCalled bool
var capturedError error

func init() {
	// Override FullErrorCapture for testing
	FullErrorCapture = func(err error) {
		captureCalled = true
		capturedError = err
	}
}

func TestRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Reset mock variables
	captureCalled = false
	capturedError = nil

	// Create a router with recovery middleware
	router := gin.New()
	router.Use(Recovery())

	// Add a route that panics
	router.GET("/panic", func(c *gin.Context) {
		panic(errors.New("test panic"))
	})

	// Add a route that panics with string
	router.GET("/panic-string", func(c *gin.Context) {
		panic("test panic string")
	})

	// Add a normal route
	router.GET("/normal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Test panic with error
	req, _ := http.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "test panic")
	assert.True(t, captureCalled, "FullErrorCapture should be called")
	assert.Equal(t, "test panic", capturedError.Error())

	// Reset mock variables for next test
	captureCalled = false
	capturedError = nil

	// Test panic with string
	req, _ = http.NewRequest("GET", "/panic-string", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request")
	assert.False(t, captureCalled, "FullErrorCapture should not be called for string panics")

	// Test normal route
	req, _ = http.NewRequest("GET", "/normal", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestErrorResponse(t *testing.T) {
	// Test ErrorResponse struct
	errResp := ErrorResponse{
		Error:   "test error",
		Message: "test message",
	}

	assert.Equal(t, "test error", errResp.Error)
	assert.Equal(t, "test message", errResp.Message)

	// Test ErrorResponse with only error
	errRespOnly := ErrorResponse{
		Error: "test error only",
	}

	assert.Equal(t, "test error only", errRespOnly.Error)
	assert.Equal(t, "", errRespOnly.Message)
}
