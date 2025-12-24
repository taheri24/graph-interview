package utils

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDumpRouter_EmptyEngine(t *testing.T) {
	r := gin.New()
	result := DumpRouter(r)
	assert.Equal(t, map[string][]string{}, result)
}

func TestDumpRouter_SingleRoute(t *testing.T) {
	r := gin.New()
	r.GET("/api/users", func(c *gin.Context) {})

	result := DumpRouter(r)
	expected := map[string][]string{
		"/api/users": {"GET"},
	}
	assert.Equal(t, expected, result)
}

func TestDumpRouter_MultipleMethodsSamePath(t *testing.T) {
	r := gin.New()
	r.GET("/api/users", func(c *gin.Context) {})
	r.POST("/api/users", func(c *gin.Context) {})
	r.PUT("/api/users", func(c *gin.Context) {})

	result := DumpRouter(r)
	expected := map[string][]string{
		"/api/users": {"GET", "POST", "PUT"},
	}
	assert.Equal(t, expected, result)
}

func TestDumpRouter_MultiplePaths(t *testing.T) {
	r := gin.New()
	r.GET("/api/users", func(c *gin.Context) {})
	r.POST("/api/users", func(c *gin.Context) {})
	r.GET("/api/posts", func(c *gin.Context) {})
	r.DELETE("/api/comments", func(c *gin.Context) {})

	result := DumpRouter(r)
	expected := map[string][]string{
		"/api/users":    {"GET", "POST"},
		"/api/posts":    {"GET"},
		"/api/comments": {"DELETE"},
	}
	assert.Equal(t, expected, result)
}

func TestDumpRouter_MethodSorting(t *testing.T) {
	r := gin.New()
	// Add methods in non-alphabetical order
	r.PUT("/api/test", func(c *gin.Context) {})
	r.GET("/api/test", func(c *gin.Context) {})
	r.POST("/api/test", func(c *gin.Context) {})
	r.DELETE("/api/test", func(c *gin.Context) {})

	result := DumpRouter(r)
	expected := map[string][]string{
		"/api/test": {"DELETE", "GET", "POST", "PUT"},
	}
	assert.Equal(t, expected, result)
}
