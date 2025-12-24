package server

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"taheri24.ir/graph1/internal/cache"
	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/handlers"
	"taheri24.ir/graph1/internal/middleware"
	"taheri24.ir/graph1/internal/models"
	"taheri24.ir/graph1/internal/routers"
	"taheri24.ir/graph1/pkg/config"
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
// @BasePath /api/v1

// @externalDocs.description OpenAPI
// @externalDocs.url https://swagger.io/resources/open-api/

func SetupAppServer(db *database.Database, cfg *config.Config) *gin.Engine {
	// Initialize cache
	var taskCache cache.CacheInterface[models.Task]
	if cfg.CacheEnabled {
		redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
		redisCache, err := cache.NewRedisCache(redisAddr, cfg.Redis.Password, cfg.Redis.DB)
		if err != nil {
			slog.Error("Failed to initialize Redis cache", "err", err)
			return nil
		}
		defer redisCache.Close()
		taskCache = cache.NewRedisCacheImpl[models.Task]("tasks", redisCache)
		slog.Info("Cache enabled")
	} else {
		taskCache = cache.NewNoOpCacheImpl[models.Task]()
		slog.Info("Cache disabled")
	}

	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(db, taskCache)
	alertHandler := handlers.NewAlertHandler()

	rootRouter := gin.Default()
	apiRouter := rootRouter.Group("/api/v1")
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
