package routers

import (
	"github.com/gin-gonic/gin"
)

// TaskHandlerInterface defines the task handler methods needed by the router
type TaskHandlerInterface interface {
	CreateTask(c *gin.Context)
	GetTasks(c *gin.Context)
	GetTask(c *gin.Context)
	UpdateTask(c *gin.Context)
	DeleteTask(c *gin.Context)
}

// SetupTaskRouter configures the task-related endpoints
func SetupTaskRouter(router *gin.Engine, taskHandler TaskHandlerInterface) {
	api := router.Group("/tasks")
	{
		api.POST("", taskHandler.CreateTask)
		api.GET("", taskHandler.GetTasks)
		api.GET("/:id", taskHandler.GetTask)
		api.PUT("/:id", taskHandler.UpdateTask)
		api.DELETE("/:id", taskHandler.DeleteTask)
	}
}
