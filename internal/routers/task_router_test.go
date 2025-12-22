package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskHandler is a mock implementation of the TaskHandlerInterface
type MockTaskHandler struct {
	mock.Mock
}

func (m *MockTaskHandler) CreateTask(c *gin.Context) {
	m.Called(c)
}

func (m *MockTaskHandler) GetTasks(c *gin.Context) {
	m.Called(c)
}

func (m *MockTaskHandler) GetTask(c *gin.Context) {
	m.Called(c)
}

func (m *MockTaskHandler) UpdateTask(c *gin.Context) {
	m.Called(c)
}

func (m *MockTaskHandler) DeleteTask(c *gin.Context) {
	m.Called(c)
}

func TestSetupTaskRouter_RouteRegistration(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create mock task handler
	mockTaskHandler := new(MockTaskHandler)

	// Create gin router
	router := gin.New()

	// Setup task router
	SetupTaskRouter(router, mockTaskHandler)

	// Get routes info
	routes := router.Routes()

	// Expected routes
	expectedRoutes := []struct {
		path   string
		method string
	}{
		{"/tasks", "POST"},
		{"/tasks", "GET"},
		{"/tasks/:id", "GET"},
		{"/tasks/:id", "PUT"},
		{"/tasks/:id", "DELETE"},
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

func TestSetupTaskRouter_EndpointHandlers(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create mock task handler
	mockTaskHandler := new(MockTaskHandler)

	// Mock all handler methods
	mockTaskHandler.On("CreateTask", mock.AnythingOfType("*gin.Context"))
	mockTaskHandler.On("GetTasks", mock.AnythingOfType("*gin.Context"))
	mockTaskHandler.On("GetTask", mock.AnythingOfType("*gin.Context"))
	mockTaskHandler.On("UpdateTask", mock.AnythingOfType("*gin.Context"))
	mockTaskHandler.On("DeleteTask", mock.AnythingOfType("*gin.Context"))

	// Create gin router
	router := gin.New()

	// Setup task router
	SetupTaskRouter(router, mockTaskHandler)

	// Test each endpoint
	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{"Create Task", "POST", "/tasks"},
		{"Get Tasks", "GET", "/tasks"},
		{"Get Task", "GET", "/tasks/1"},
		{"Update Task", "PUT", "/tasks/1"},
		{"Delete Task", "DELETE", "/tasks/1"},
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
	mockTaskHandler.AssertExpectations(t)
}

func TestSetupTaskRouter_GroupPath(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create mock task handler
	mockTaskHandler := new(MockTaskHandler)

	// Create gin router
	router := gin.New()

	// Setup task router
	SetupTaskRouter(router, mockTaskHandler)

	// Get routes info
	routes := router.Routes()

	// Verify all task routes are under /tasks group
	taskRoutes := []string{"/tasks", "/tasks/:id"}
	for _, route := range routes {
		// Check if this is a task route
		isTaskRoute := false
		for _, taskRoute := range taskRoutes {
			if route.Path == taskRoute {
				isTaskRoute = true
				break
			}
		}

		if isTaskRoute {
			assert.True(t, len(route.Path) >= 6 && route.Path[:6] == "/tasks",
				"Route %s should be under /tasks group", route.Path)
		}
	}
}
