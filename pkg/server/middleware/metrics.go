package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "osmedeus_http_requests_total",
		Help: "Total HTTP requests",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "osmedeus_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	httpRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "osmedeus_http_requests_in_flight",
		Help: "Current number of HTTP requests being processed",
	})
)

// PrometheusMetrics returns a Fiber middleware that records HTTP metrics
func PrometheusMetrics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		httpRequestsInFlight.Inc()
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()
		httpRequestsInFlight.Dec()

		// Normalize path to avoid high cardinality (replace IDs with :id)
		path := normalizePath(c.Route().Path)

		httpRequestDuration.WithLabelValues(c.Method(), path).Observe(duration)
		httpRequestsTotal.WithLabelValues(c.Method(), path, strconv.Itoa(c.Response().StatusCode())).Inc()

		return err
	}
}

// normalizePath normalizes the path to avoid high cardinality metrics
func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}
