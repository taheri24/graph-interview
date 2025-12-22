package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthChecker interface defines the health check functionality
type HealthChecker interface {
	Health() error
}

// SetupHealthRouter configures the health check endpoint
func SetupHealthRouter(router *gin.Engine, healthChecker HealthChecker) {
	router.GET("/health", func(c *gin.Context) {
		if err := healthChecker.Health(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"database": "connected",
		})
	})
}
