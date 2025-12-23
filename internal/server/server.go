package server

import (
	"github.com/gin-gonic/gin"
	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/handlers"
	"taheri24.ir/graph1/internal/middleware"
	"taheri24.ir/graph1/internal/routers"
)

// @title Task Management API
// @version 1.0
// @description A RESTful API for managing tasks built with Go, Gin, and PostgreSQL.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /

// @externalDocs.description OpenAPI
// @externalDocs.url https://swagger.io/resources/open-api/

func SetupAppServer(db *database.Database, taskHandler *handlers.TaskHandler, alertHandler *handlers.AlertHandler) *gin.Engine {
	rootRouter := gin.Default()
	apiRouter := rootRouter.Group("/api")
	// Setup global middleware
	middleware.SetupGlobalMiddleware(rootRouter)

	// Setup routes
	routers.SetupHealthRouter(apiRouter, db)
	routers.SetupTaskRouter(apiRouter, taskHandler)
	routers.SetupAlertRouter(apiRouter, alertHandler)
	routers.SetupSwaggerRouter(rootRouter)

	// Setup metrics endpoint
	middleware.SetupMetricsEndpoint(rootRouter)

	return rootRouter
}
