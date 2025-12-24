package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupAppServerBasic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a real database for integration testing
	cfg := config.NewTestConfig()
	db, err := database.NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	testCfg := &config.Config{
		Database:     cfg.Database,
		Redis:        cfg.Redis,
		CacheEnabled: false, // Disable to avoid Redis connection issues
		Server:       cfg.Server,
	}

	router := SetupAppServer(db, testCfg)

	// Verify router was created
	assert.NotNil(t, router)
	assert.IsType(t, &gin.Engine{}, router)
}

func TestSetupAppServerHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.NewTestConfig()
	db, err := database.NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	testCfg := &config.Config{
		Database:     cfg.Database,
		Redis:        cfg.Redis,
		CacheEnabled: false,
		Server:       cfg.Server,
	}

	router := SetupAppServer(db, testCfg)

	req, err := http.NewRequest("GET", "/api/v1/health", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestSetupAppServerInvalidRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.NewTestConfig()
	db, err := database.NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	testCfg := &config.Config{
		Database:     cfg.Database,
		Redis:        cfg.Redis,
		CacheEnabled: false,
		Server:       cfg.Server,
	}

	router := SetupAppServer(db, testCfg)

	invalidRoutes := []string{
		"/invalid",
		"/api/v1/nonexistent",
		"/api/v2/tasks",
		"/swagger/nonexistent",
	}

	for _, route := range invalidRoutes {
		t.Run("Invalid route: "+route, func(t *testing.T) {
			req, err := http.NewRequest("GET", route, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	}
}

func TestSetupAppServerCacheDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.NewTestConfig()
	db, err := database.NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	testCfg := &config.Config{
		Database:     cfg.Database,
		Redis:        cfg.Redis,
		CacheEnabled: false,
		Server:       cfg.Server,
	}

	router := SetupAppServer(db, testCfg)

	// Server should still work with cache disabled
	req, err := http.NewRequest("GET", "/api/v1/health", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetupAppServerRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.NewTestConfig()
	db, err := database.NewDatabase(cfg)
	require.NoError(t, err)
	defer db.Close()

	testCfg := &config.Config{
		Database:     cfg.Database,
		Redis:        cfg.Redis,
		CacheEnabled: false,
		Server:       cfg.Server,
	}

	router := SetupAppServer(db, testCfg)

	// Test that recovery middleware handles panics
	req, err := http.NewRequest("GET", "/api/v1/tasks", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should not return 500 Internal Server Error (recovery middleware should handle any panics)
	// Even if the handler fails due to no cache, it should return a proper error response
	assert.NotEqual(t, http.StatusInternalServerError, w.Code,
		"Recovery middleware should prevent 500 errors")
}
