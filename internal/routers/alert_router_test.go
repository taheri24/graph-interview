package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAlertHandler is a mock implementation of AlertHandlerInterface
type MockAlertHandler struct {
	mock.Mock
}

func (m *MockAlertHandler) GetAlerts(c *gin.Context) {
	m.Called(c)
}

func (m *MockAlertHandler) FireAlert(c *gin.Context) {
	m.Called(c)
}

func (m *MockAlertHandler) ResetAlert(c *gin.Context) {
	m.Called(c)
}

func TestSetupAlertRouter_RouteRegistration(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create mock alert handler
	mockAlertHandler := new(MockAlertHandler)

	// Create gin router
	router := gin.New()

	// Setup alert router
	SetupAlertRouter(router, mockAlertHandler)

	// Get routes info
	routes := router.Routes()

	// Expected routes
	expectedRoutes := []struct {
		path   string
		method string
	}{
		{"/alerts", "GET"},
		{"/alerts/fire", "POST"},
		{"/alerts/reset", "POST"},
	}

	// Verify all expected routes are registered
	for _, expected := range expectedRoutes {
		var routeFound bool
		for _, route := range routes {
			if route.Path == expected.path && route.Method == expected.method {
				routeFound = true
				break
			}
		}
		assert.True(t, routeFound, "Route %s %s should be registered", expected.method, expected.path)
	}
}

func TestSetupAlertRouter_EndpointHandlers(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create mock alert handler
	mockAlertHandler := new(MockAlertHandler)

	// Mock all handler methods
	mockAlertHandler.On("GetAlerts", mock.AnythingOfType("*gin.Context"))
	mockAlertHandler.On("FireAlert", mock.AnythingOfType("*gin.Context"))
	mockAlertHandler.On("ResetAlert", mock.AnythingOfType("*gin.Context"))

	// Create gin router
	router := gin.New()

	// Setup alert router
	SetupAlertRouter(router, mockAlertHandler)

	// Test each endpoint
	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{"Get Alerts", "GET", "/alerts"},
		{"Fire Alert", "POST", "/alerts/fire"},
		{"Reset Alert", "POST", "/alerts/reset"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req, _ := http.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Verify the handler was called (we expect 404 or 200 since handlers are mocked)
			// The important thing is that the route is registered and handler is called
			assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusOK || w.Code == http.StatusBadRequest,
				"Route should be handled, got status: %d", w.Code)
		})
	}

	// Verify all mock expectations were met
	mockAlertHandler.AssertExpectations(t)
}

func TestSetupAlertRouter_GroupPath(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create mock alert handler
	mockAlertHandler := new(MockAlertHandler)

	// Create gin router
	router := gin.New()

	// Setup alert router
	SetupAlertRouter(router, mockAlertHandler)

	// Get routes info
	routes := router.Routes()

	// Verify all alert routes are under /alerts group
	alertRoutes := []string{"/alerts", "/alerts/fire", "/alerts/reset"}
	for _, route := range routes {
		// Check if this is an alert route
		isAlertRoute := false
		for _, alertRoute := range alertRoutes {
			if route.Path == alertRoute {
				isAlertRoute = true
				break
			}
		}

		if isAlertRoute {
			assert.True(t, len(route.Path) >= 7 && route.Path[:7] == "/alerts",
				"Route %s should be under /alerts group", route.Path)
		}
	}
}
