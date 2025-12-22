package routers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHealthChecker is a mock implementation of the HealthChecker interface
type MockHealthChecker struct {
	mock.Mock
}

func (m *MockHealthChecker) Health() error {
	args := m.Called()
	return args.Error(0)
}

func TestSetupHealthRouter_Healthy(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create mock health checker
	mockHealthChecker := new(MockHealthChecker)
	mockHealthChecker.On("Health").Return(nil)

	// Create gin router
	router := gin.New()

	// Setup health router
	SetupHealthRouter(router, mockHealthChecker)

	// Create request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
	assert.Contains(t, w.Body.String(), "connected")
	mockHealthChecker.AssertExpectations(t)
}

func TestSetupHealthRouter_Unhealthy(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create mock health checker
	mockHealthChecker := new(MockHealthChecker)
	mockHealthChecker.On("Health").Return(errors.New("database connection failed"))

	// Create gin router
	router := gin.New()

	// Setup health router
	SetupHealthRouter(router, mockHealthChecker)

	// Create request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "unhealthy")
	assert.Contains(t, w.Body.String(), "database connection failed")
	mockHealthChecker.AssertExpectations(t)
}

func TestSetupHealthRouter_RouteRegistration(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create mock health checker
	mockHealthChecker := new(MockHealthChecker)

	// Create gin router
	router := gin.New()

	// Setup health router
	SetupHealthRouter(router, mockHealthChecker)

	// Get routes info
	routes := router.Routes()

	// Verify health route is registered
	var healthRouteFound bool
	for _, route := range routes {
		if route.Path == "/health" && route.Method == "GET" {
			healthRouteFound = true
			break
		}
	}

	assert.True(t, healthRouteFound, "Health route should be registered")
}
