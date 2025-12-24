package server

import (
	"fmt"
	"log/slog"
	"net/http/pprof"
	"os"

	"github.com/gin-gonic/gin"
	"taheri24.ir/graph1/internal/cache"
	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/handlers/alert"
	"taheri24.ir/graph1/internal/handlers/task"
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
		taskCache = cache.NewRedisCacheImpl[models.Task]("tasks", redisCache)
		slog.Info("Cache enabled")
	} else {
		taskCache = cache.NewNoOpCacheImpl[models.Task]()
		slog.Info("Cache disabled")
	}

	// Initialize handlers
	taskHandler := task.NewTaskHandler(db, taskCache)
	alertHandler := alert.NewAlertHandler()

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

	// Setup pprof endpoints if enabled
	if os.Getenv("PPROF_ENABLED") == "true" {
		setupPprofEndpoints(rootRouter)
	}

	return rootRouter
}

// setupPprofEndpoints adds pprof debugging endpoints to the router
func setupPprofEndpoints(router *gin.Engine) {
	pprofGroup := router.Group("/debug/pprof")
	{
		pprofGroup.GET("/", gin.WrapF(pprof.Index))
		pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
		pprofGroup.POST("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
		pprofGroup.GET("/heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
		pprofGroup.GET("/goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
		pprofGroup.GET("/threadcreate", gin.WrapF(pprof.Handler("threadcreate").ServeHTTP))
		pprofGroup.GET("/block", gin.WrapF(pprof.Handler("block").ServeHTTP))
	}

	slog.Info("Pprof endpoints enabled at /debug/pprof")
}
