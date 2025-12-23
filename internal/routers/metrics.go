package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupMetricsRouter creates a router for Prometheus metrics
func SetupMetricsRouter() *gin.Engine {
	router := gin.New()

	// Metrics endpoint for Prometheus scraping
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return router
}
