package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Tracks total HTTP requests with labels for endpoint + status
	HttpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, labeled by endpoint and status",
		},
		[]string{"endpoint", "status"},
	)

	// Tracks request duration (latency) for each endpoint
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	// Count of successful logins
	SuccessfulLogins = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "successful_logins_total",
			Help: "Total number of successful logins",
		},
	)

	// Count of failed logins
	FailedLogins = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "failed_logins_total",
			Help: "Total number of failed login attempts",
		},
	)

	// Gauge for active sessions (increased on login, decreased on logout)
	ActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_sessions",
			Help: "Number of currently active sessions",
		},
	)
)

// 🔹 Prometheus Middleware for Gin
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Endpoint path (route)
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}

		// Status code
		status := strconv.Itoa(c.Writer.Status())

		// Record request count
		HttpRequests.WithLabelValues(endpoint, status).Inc()

		// Record request duration
		duration := time.Since(start).Seconds()
		HttpRequestDuration.WithLabelValues(endpoint).Observe(duration)
	}
}
