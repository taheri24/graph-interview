package main

import (
	"log"

	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/handlers"
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

func main() {
	// Load configuration
	cfg := config.Load()
	// Initialize database
	db := utils.Must(database.NewDatabase(cfg))
	defer db.Close()
	if err := db.Health(); err != nil {
		log.Fatal(err)
	}
	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(db)

	// Set up rootRouter
	rootRouter := gin.Default()
	{
		routers.SetupHealthRouter(rootRouter, db)
		routers.SetupTaskRouter(rootRouter, taskHandler)
		routers.SetupSwaggerRouter(rootRouter)
	}
	log.Printf("Server starting on port %q", cfg.Server.Port)
	if err := rootRouter.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
