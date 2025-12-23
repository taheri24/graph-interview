package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupGlobalMiddleware configures all global middleware for the router
func SetupGlobalMiddleware(router *gin.Engine) {
	// Recovery middleware to handle panics
	router.Use(Recovery())

	// Request ID middleware for tracing
	router.Use(RequestIDMiddleware())

	// Prometheus metrics middleware
	router.Use(MetricsMiddleware())
}

// SetupMetricsEndpoint adds the /metrics endpoint to the router
func SetupMetricsEndpoint(router *gin.Engine) {
	// Add metrics endpoint with promhttp handler
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
