package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP requests counter
	HttpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, labeled by endpoint and status",
		},
		[]string{"endpoint", "status"},
	)

	// Request latency
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
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

	ContinuationsAdded = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "continuations_added_total",
			Help: "Total number of continuations submitted",
		},
	)

	ContinuationsAccepted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "continuations_accepted_total",
			Help: "Total number of continuations accepted",
		},
	)
)
