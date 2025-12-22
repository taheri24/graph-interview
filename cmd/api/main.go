package main

import (
	"log"

	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/handlers"
	"taheri24.ir/graph1/internal/routers"
	"taheri24.ir/graph1/pkg/config"

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

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Set up Gin router
	router := gin.Default()

	// Setup routers
	routers.SetupHealthRouter(router, db)

	// Initialize task handler
	taskHandler := handlers.NewTaskHandler(db)
	routers.SetupTaskRouter(router, taskHandler)

	routers.SetupSwaggerRouter(router)

	// Start server
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
