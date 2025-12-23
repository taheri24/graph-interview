package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

func TestMetricsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		statusCode     int
		expectedLabels []string
	}{
		{
			name:           "GET request success",
			method:         "GET",
			path:           "/test",
			statusCode:     200,
			expectedLabels: []string{"GET", "/test", "200"},
		},
		{
			name:           "POST request created",
			method:         "POST",
			path:           "/tasks",
			statusCode:     201,
			expectedLabels: []string{"POST", "/tasks", "201"},
		},
		{
			name:           "PUT request success",
			method:         "PUT",
			path:           "/tasks/123",
			statusCode:     200,
			expectedLabels: []string{"PUT", "/tasks/:id", "200"},
		},
		{
			name:           "DELETE request no content",
			method:         "DELETE",
			path:           "/tasks/123",
			statusCode:     204,
			expectedLabels: []string{"DELETE", "/tasks/:id", "204"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset metrics before test
			requestsTotal.Reset()
			requestLatencyHistogram.Reset()

			router := gin.New()
			router.Use(MetricsMiddleware())

			router.Handle(tt.method, tt.path, func(c *gin.Context) {
				c.Status(tt.statusCode)
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check that metrics were recorded
			registry := prometheus.NewRegistry()
			registry.MustRegister(requestsTotal, requestLatencyHistogram)

			metricFamilies, err := registry.Gather()
			if err != nil {
				t.Fatalf("Error gathering metrics: %v", err)
			}

			// Verify requests_total metric
			var found bool
			for _, mf := range metricFamilies {
				if mf.GetName() == "requests_total" {
					found = true
					if len(mf.GetMetric()) == 0 {
						t.Error("Expected at least one metric for requests_total")
					}
					break
				}
			}
			if !found {
				t.Error("requests_total metric not found")
			}
		})
	}
}

func TestMetricsMiddleware_SkipHealthAndSwagger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Reset metrics
	requestsTotal.Reset()

	router := gin.New()
	router.Use(MetricsMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.Status(200)
	})
	router.GET("/swagger", func(c *gin.Context) {
		c.Status(200)
	})
	router.GET("/tasks", func(c *gin.Context) {
		c.Status(200)
	})

	// Test health endpoint (should be skipped)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Test swagger endpoint (should be skipped)
	req = httptest.NewRequest("GET", "/swagger", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Test tasks endpoint (should be recorded)
	req = httptest.NewRequest("GET", "/tasks", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that only one metric was recorded (for /tasks)
	registry := prometheus.NewRegistry()
	registry.MustRegister(requestsTotal)

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("Error gathering metrics: %v", err)
	}

	var metricCount int
	for _, mf := range metricFamilies {
		if mf.GetName() == "requests_total" {
			metricCount = len(mf.GetMetric())
		}
	}

	if metricCount != 1 {
		t.Errorf("Expected 1 metric recorded, got %d", metricCount)
	}
}

func TestUpdateTasksCount(t *testing.T) {
	// Reset the gauge
	tasksCount.Set(0)

	// Test updating the gauge
	UpdateTasksCount(42.0)

	// Check that the gauge was updated
	registry := prometheus.NewRegistry()
	registry.MustRegister(tasksCount)

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("Error gathering metrics: %v", err)
	}

	var found bool
	var value float64
	for _, mf := range metricFamilies {
		if mf.GetName() == "tasks_count" {
			found = true
			if len(mf.GetMetric()) > 0 {
				value = mf.GetMetric()[0].GetGauge().GetValue()
			}
			break
		}
	}

	if !found {
		t.Error("tasks_count metric not found")
	}

	if value != 42.0 {
		t.Errorf("Expected tasks_count to be 42.0, got %f", value)
	}
}
