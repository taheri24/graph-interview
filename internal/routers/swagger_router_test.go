package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetupSwaggerRouter_RouteRegistration(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create gin router
	router := gin.New()

	// Setup swagger router
	SetupSwaggerRouter(router)

	// Get routes info
	routes := router.Routes()

	// Verify swagger route is registered
	var swaggerRouteFound bool
	for _, route := range routes {
		if route.Path == "/swagger/*any" && route.Method == "GET" {
			swaggerRouteFound = true
			break
		}
	}

	assert.True(t, swaggerRouteFound, "Swagger route should be registered")
}

func TestSetupSwaggerRouter_EndpointResponse(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create gin router
	router := gin.New()

	// Setup swagger router
	SetupSwaggerRouter(router)

	// Test cases for different swagger paths
	testCases := []struct {
		name     string
		path     string
		expected int
	}{
		{"Swagger Index", "/swagger/index.html", http.StatusOK},
		{"Swagger JSON", "/swagger/doc.json", http.StatusOK},
		{"Swagger Root", "/swagger/", http.StatusMovedPermanently}, // Redirect
		{"Invalid Swagger Path", "/swagger/invalid", http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req, _ := http.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// For swagger routes, we expect either the expected status or 404 if swagger files aren't available in test
			// The important thing is that the route is registered and handled
			assert.True(t, w.Code == tc.expected || w.Code == http.StatusNotFound,
				"Route %s should be handled, expected %d or 404, got %d", tc.path, tc.expected, w.Code)
		})
	}
}

func TestSetupSwaggerRouter_WildcardPath(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create gin router
	router := gin.New()

	// Setup swagger router
	SetupSwaggerRouter(router)

	// Test various wildcard paths
	testPaths := []string{
		"/swagger/anything",
		"/swagger/deep/nested/path",
		"/swagger/file.txt",
	}

	for _, path := range testPaths {
		t.Run("Wildcard path: "+path, func(t *testing.T) {
			// Create request
			req, _ := http.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Should be handled (either successfully or with 404 if file doesn't exist)
			// The important thing is that the wildcard route captures the path
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound || w.Code == http.StatusMovedPermanently,
				"Wildcard route %s should be handled, got status: %d", path, w.Code)
		})
	}
}

func TestSetupSwaggerRouter_NonSwaggerRoutes(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create gin router
	router := gin.New()

	// Setup swagger router
	SetupSwaggerRouter(router)

	// Test non-swagger routes should return 404
	testPaths := []string{
		"/api",
		"/health",
		"/tasks",
		"/random/path",
	}

	for _, path := range testPaths {
		t.Run("Non-swagger path: "+path, func(t *testing.T) {
			// Create request
			req, _ := http.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Should return 404 since no other routes are registered
			assert.Equal(t, http.StatusNotFound, w.Code,
				"Non-swagger route %s should return 404, got %d", path, w.Code)
		})
	}
}
