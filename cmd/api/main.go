package main

import (
	"log/slog"

	_ "taheri24.ir/graph1/docs"
	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/handlers"
	"taheri24.ir/graph1/internal/middleware"
	"taheri24.ir/graph1/internal/routers"
	"taheri24.ir/graph1/pkg/config"
	"taheri24.ir/graph1/pkg/utils"

	"github.com/gin-gonic/gin"
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

func setupAppServer(db *database.Database, taskHandler *handlers.TaskHandler, alertHandler *handlers.AlertHandler) *gin.Engine {
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

func main() {

	// Load configuration
	cfg := config.Load()
	// Initialize database
	db := utils.Must(database.NewDatabase(cfg))
	defer db.Close()
	if err := db.Health(); err != nil {
		slog.Error(err.Error())
		return
	}
	if err := database.Migrate(db.DB); err != nil {
		slog.Error("Database Migrate failed ", "err", err)
		return
	} else {
		slog.Info("Migrate Passed")
	}
	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(db)
	alertHandler := handlers.NewAlertHandler()

	// Set up rootRouter
	rootRouter := setupAppServer(db, taskHandler, alertHandler)

	slog.Info("Server starting on port ", "port", cfg.Server.Port)
	if err := rootRouter.Run(":" + cfg.Server.Port); err != nil {
		slog.Error("Failed to start server: %v", "err", err)
	}
}
