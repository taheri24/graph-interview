package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)

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

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "test panic")

	// Test panic with string
	req, _ = http.NewRequest("GET", "/panic-string", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request")

	// Test normal route
	req, _ = http.NewRequest("GET", "/normal", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}
