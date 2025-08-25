package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HttpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "story_http_requests_total",
			Help: "Total number of HTTP requests processed by StoryService, labeled by endpoint and status",
		},
		[]string{"endpoint", "status"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "story_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds for StoryService",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	// Business metrics
	StoriesCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "stories_created_total",
			Help: "Total number of stories created",
		},
	)

	ContinuationsSubmitted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "continuations_submitted_total",
			Help: "Total number of continuations submitted",
		},
	)

	ContinuationsAccepted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "continuations_accepted_total",
			Help: "Total number of continuations accepted",
		},
	)

	StoriesDeleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "stories_deleted_total",
			Help: "Total number of stories deleted",
		},
	)

	ContinuationsDeleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "continuations_deleted_total",
			Help: "Total number of continuations deleted",
		},
	)
)

// 🔹 Middleware for Prometheus
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}

		status := strconv.Itoa(c.Writer.Status())
		HttpRequests.WithLabelValues(endpoint, status).Inc()

		duration := time.Since(start).Seconds()
		HttpRequestDuration.WithLabelValues(endpoint).Observe(duration)
	}
}
