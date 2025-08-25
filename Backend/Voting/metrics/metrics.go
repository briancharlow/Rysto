package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"time"
	"net/http"
)

// --- Business Metrics ---
var (
	// Count of total votes cast
	VotesCast = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "votes_cast_total",
			Help: "Total number of votes cast",
		},
	)

	// Count of votes deleted (withdrawn)
	VotesDeleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "votes_deleted_total",
			Help: "Total number of votes deleted",
		},
	)

	// Active votes gauge (increment on vote, decrement on delete)
	ActiveVotes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_votes",
			Help: "Number of currently active votes",
		},
	)

	// Tracks HTTP requests
	HttpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests labeled by endpoint and status",
		},
		[]string{"endpoint", "status"},
	)

	// Latency histogram
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Request latency distribution",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)
)

// --- Middleware ---
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := http.StatusText(c.Writer.Status())

		HttpRequests.WithLabelValues(c.FullPath(), status).Inc()
		HttpRequestDuration.WithLabelValues(c.FullPath()).Observe(duration)
	}
}

// --- Expose metrics endpoint ---
func RegisterMetricsEndpoint(r *gin.Engine) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
