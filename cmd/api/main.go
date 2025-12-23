package main

import (
	"log/slog"

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

func setupAppServer(db *database.Database, taskHandler *handlers.TaskHandler) *gin.Engine {
	rootRouter := gin.Default()

	// Setup global middleware
	middleware.SetupGlobalMiddleware(rootRouter)

	// Setup routes
	routers.SetupHealthRouter(rootRouter, db)
	routers.SetupTaskRouter(rootRouter, taskHandler)
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

	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(db)

	// Set up rootRouter
	rootRouter := setupAppServer(db, taskHandler)

	slog.Info("Server starting on port ", "port", cfg.Server.Port)
	if err := rootRouter.Run(":" + cfg.Server.Port); err != nil {
		slog.Error("Failed to start server: %v", "err", err)
	}
}
