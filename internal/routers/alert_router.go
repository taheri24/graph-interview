package routers

import (
	"github.com/gin-gonic/gin"
)

// AlertHandlerInterface defines the alert handler methods needed by the router
type AlertHandlerInterface interface {
	GetAlerts(c *gin.Context)
	FireAlert(c *gin.Context)
	ResetAlert(c *gin.Context)
}

// SetupAlertRouter configures the alert-related endpoints
func SetupAlertRouter(router gin.IRouter, alertHandler AlertHandlerInterface) {
	api := router.Group("/alerts")
	{
		api.GET("", alertHandler.GetAlerts)
		api.POST("/fire", alertHandler.FireAlert)
		api.POST("/reset", alertHandler.ResetAlert)
	}
}
