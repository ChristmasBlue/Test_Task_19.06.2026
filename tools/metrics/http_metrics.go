package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var httpRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "path", "status"},
)

var httpRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "path", "status"},
)

var httpErrorsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_errors_total",
		Help: "Total number of HTTP errors (status >= 400)",
	},
	[]string{"method", "path", "status"},
)

func HTTPMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		method := c.Request.Method
		path := c.FullPath()
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()

		if c.Writer.Status() >= 400 {
			httpErrorsTotal.WithLabelValues(method, path, status).Inc()
		}

		httpRequestDuration.WithLabelValues(method, path, status).Observe(duration)

	}
}
