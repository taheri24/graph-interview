package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// requestsTotal counts total HTTP requests with method and path labels
	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	// requestLatencyHistogram tracks request duration
	requestLatencyHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "request_latency_histogram_seconds",
		Help:    "Request latency in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	// tasksCount tracks current number of tasks
	tasksCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tasks_count",
		Help: "Current number of tasks in the database",
	})

	// alertTrigger is a gauge that can be set to trigger alerts manually
	alertTrigger = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alert_trigger",
		Help: "Manually triggered alert gauge",
	}, []string{"alert_name"})
)

// MetricsMiddleware tracks Prometheus metrics for HTTP requests
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		method := c.Request.Method

		// Process request
		c.Next()

		// Record metrics after request is processed
		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		// Skip metrics for health and swagger endpoints
		if path != "/health" && path != "/swagger" && path != "/swagger/" && path != "/swagger/*" {
			requestsTotal.WithLabelValues(method, path, status).Inc()
			requestLatencyHistogram.WithLabelValues(method, path).Observe(duration)
		}
	}
}

// UpdateTasksCount updates the tasks count gauge
func UpdateTasksCount(count float64) {
	tasksCount.Set(count)
}

// TriggerAlert sets the alert trigger gauge to 1 for the given alert name
func TriggerAlert(alertName string) {
	alertTrigger.WithLabelValues(alertName).Set(1)
}

// ResetAlert sets the alert trigger gauge to 0 for the given alert name
func ResetAlert(alertName string) {
	alertTrigger.WithLabelValues(alertName).Set(0)
}
