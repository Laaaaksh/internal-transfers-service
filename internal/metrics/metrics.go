// Package metrics provides Prometheus metrics for the application.
package metrics

import (
	"github.com/internal-transfers-service/internal/constants"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTP metrics
var (
	// HTTPRequestDuration tracks HTTP request latency
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    constants.MetricRequestDuration,
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{constants.LabelMethod, constants.LabelPath, constants.LabelStatusCode},
	)

	// HTTPRequestsTotal tracks total HTTP requests
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: constants.MetricRequestTotal,
			Help: "Total number of HTTP requests",
		},
		[]string{constants.LabelMethod, constants.LabelPath, constants.LabelStatusCode},
	)
)

// Business metrics
var (
	// TransfersTotal tracks total transfer attempts
	TransfersTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: constants.MetricTransferTotal,
			Help: "Total number of transfer attempts",
		},
	)

	// TransfersSuccess tracks successful transfers
	TransfersSuccess = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: constants.MetricTransferSuccess,
			Help: "Total number of successful transfers",
		},
	)

	// TransfersFailed tracks failed transfers
	TransfersFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: constants.MetricTransferFailed,
			Help: "Total number of failed transfers",
		},
		[]string{constants.LabelReason},
	)
)

// Database metrics
var (
	// DBConnectionsOpen tracks open database connections
	DBConnectionsOpen = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: constants.MetricDBConnectionsOpen,
			Help: "Number of open database connections",
		},
	)

	// DBConnectionsIdle tracks idle database connections
	DBConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: constants.MetricDBConnectionsIdle,
			Help: "Number of idle database connections",
		},
	)
)

// RecordHTTPRequest records an HTTP request metric
func RecordHTTPRequest(method, path string, statusCode int, durationSeconds float64) {
	statusCodeStr := statusCodeToLabel(statusCode)
	HTTPRequestDuration.WithLabelValues(method, path, statusCodeStr).Observe(durationSeconds)
	HTTPRequestsTotal.WithLabelValues(method, path, statusCodeStr).Inc()
}

// RecordTransferAttempt records a transfer attempt
func RecordTransferAttempt() {
	TransfersTotal.Inc()
}

// RecordTransferSuccess records a successful transfer
func RecordTransferSuccess() {
	TransfersSuccess.Inc()
}

// RecordTransferFailure records a failed transfer with reason
func RecordTransferFailure(reason string) {
	TransfersFailed.WithLabelValues(reason).Inc()
}

// UpdateDBConnectionMetrics updates database connection metrics
func UpdateDBConnectionMetrics(open, idle int64) {
	DBConnectionsOpen.Set(float64(open))
	DBConnectionsIdle.Set(float64(idle))
}

// statusCodeToLabel converts HTTP status code to a label
func statusCodeToLabel(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}
