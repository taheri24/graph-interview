package main

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	_ "taheri24.ir/graph1/docs"
	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/server"
	"taheri24.ir/graph1/pkg/config"
	"taheri24.ir/graph1/pkg/utils"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	cfg := config.Load()
	db := utils.Must(database.NewDatabase(cfg))
	defer db.Close()
	if err := db.Health(); err != nil {
		slog.Error("Database connection health failed", "err", err)
		return
	}
	if err := database.Migrate(db.DB); err != nil {
		slog.Error("Database Migrate failed ", "err", err)
		return
	} else {
		slog.Info("Migrate Passed")
	}

	// Set up rootRouter
	rootRouter := server.SetupAppServer(db, cfg)
	if rootRouter == nil {
		slog.Error("Failed to setup server")
		return
	}

	slog.Info("Server starting on port ", "port", cfg.Server.Port)
	if err := rootRouter.Run(":" + cfg.Server.Port); err != nil {
		slog.Error("Failed to start server: %v", "err", err)
	}
}
