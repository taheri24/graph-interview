package routers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsEndpoint(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	router := SetupMetricsRouter()

	// Create a test request to /metrics endpoint
	req, err := http.NewRequest("GET", "/metrics", nil)
	require.NoError(t, err)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response contains Prometheus metrics format
	body := w.Body.String()
	assert.NotEmpty(t, body)

	// Basic validation that it looks like Prometheus metrics
	// Prometheus metrics typically contain lines like "# HELP", "# TYPE", or metric names
	lines := strings.Split(strings.TrimSpace(body), "\n")
	assert.Greater(t, len(lines), 0)

	// Check for typical Prometheus format indicators
	hasPrometheusContent := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# HELP") ||
			strings.HasPrefix(line, "# TYPE") ||
			(len(line) > 0 && !strings.HasPrefix(line, "#")) {
			hasPrometheusContent = true
			break
		}
	}
	assert.True(t, hasPrometheusContent, "Response should contain Prometheus metrics format")
}
