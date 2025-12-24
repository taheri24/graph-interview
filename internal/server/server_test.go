package server

import (
	"testing"

	"github.com/gin-gonic/gin"
	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/handlers"
	"taheri24.ir/graph1/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestSetupAppServer_Snapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handlers with nil repo (for testing route setup only)
	mockDB := &database.Database{}
	taskHandler := &handlers.TaskHandler{}
	alertHandler := handlers.NewAlertHandler()

	// Setup the server
	router := SetupAppServer(mockDB, taskHandler, alertHandler)

	// Dump the routes
	routesDump := utils.DumpRouter(router)

	// Hardcoded snapshot of expected routes
	expected := map[string][]string{
		"/metrics":          {"GET"},
		"/api/v1/health":    {"GET"},
		"/api/v1/tasks":     {"GET", "POST"},
		"/api/v1/tasks/:id": {"DELETE", "GET", "PUT"},
		"/api/v1/alerts":    {"GET"},
	}
	for key := range expected {
		t.Run(key, func(t *testing.T) {

			assert.Equal(t, expected[key], routesDump[key])
		})
	}
}
