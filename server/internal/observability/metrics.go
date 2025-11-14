package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// connectionEventsCounter tracks connection lifecycle events
	connectionEventsCounter *prometheus.CounterVec

	// messagesCounter tracks message counts (in/out)
	messagesCounter *prometheus.CounterVec

	// activeConnectionsGauge tracks current number of active connections
	activeConnectionsGauge prometheus.Gauge

	// queueDepthGauge tracks current command queue depth
	queueDepthGauge prometheus.Gauge

	// tickDurationHistogram tracks tick execution duration
	tickDurationHistogram prometheus.Histogram

	// gcPauseHistogram tracks GC pause durations
	gcPauseHistogram prometheus.Histogram

	// metricsInitialized tracks whether metrics have been initialized
	metricsInitialized bool
)

// InitMetrics initializes and registers all Prometheus metrics.
// This should be called once during server startup.
func InitMetrics() {
	if metricsInitialized {
		// Reset metrics by unregistering and re-registering
		// Unregister will return false if metric is not registered, which is fine
		if connectionEventsCounter != nil {
			prometheus.Unregister(connectionEventsCounter)
		}
		if messagesCounter != nil {
			prometheus.Unregister(messagesCounter)
		}
		if activeConnectionsGauge != nil {
			prometheus.Unregister(activeConnectionsGauge)
		}
		if queueDepthGauge != nil {
			prometheus.Unregister(queueDepthGauge)
		}
		if tickDurationHistogram != nil {
			prometheus.Unregister(tickDurationHistogram)
		}
		if gcPauseHistogram != nil {
			prometheus.Unregister(gcPauseHistogram)
		}
	}

	// Connection events counter
	connectionEventsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "connection_events_total",
			Help: "Total number of connection events",
		},
		[]string{"event"}, // event: connect, disconnect, error
	)

	// Messages counter
	messagesCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "messages_total",
			Help: "Total number of messages processed",
		},
		[]string{"direction"}, // direction: in, out
	)

	// Active connections gauge
	activeConnectionsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Current number of active connections",
		},
	)

	// Queue depth gauge
	queueDepthGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "queue_depth",
			Help: "Current command queue depth",
		},
	)

	// Tick duration histogram with buckets for p50/p95/p99 percentiles
	// Buckets: 1ms, 5ms, 10ms, 50ms, 100ms, +Inf
	tickDurationHistogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "tick_duration_seconds",
			Help:    "Tick execution duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1}, // 1ms, 5ms, 10ms, 50ms, 100ms
		},
	)

	// GC pause histogram
	gcPauseHistogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "gc_pause_seconds",
			Help:    "GC pause duration in seconds",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.002, 0.005}, // 0.1ms, 0.5ms, 1ms, 2ms, 5ms
		},
	)

	// Register all metrics
	prometheus.MustRegister(connectionEventsCounter)
	prometheus.MustRegister(messagesCounter)
	prometheus.MustRegister(activeConnectionsGauge)
	prometheus.MustRegister(queueDepthGauge)
	prometheus.MustRegister(tickDurationHistogram)
	prometheus.MustRegister(gcPauseHistogram)

	metricsInitialized = true
}

// GetConnectionEventsCounter returns the connection events counter metric.
func GetConnectionEventsCounter() *prometheus.CounterVec {
	return connectionEventsCounter
}

// GetMessagesCounter returns the messages counter metric.
func GetMessagesCounter() *prometheus.CounterVec {
	return messagesCounter
}

// GetActiveConnectionsGauge returns the active connections gauge metric.
func GetActiveConnectionsGauge() prometheus.Gauge {
	return activeConnectionsGauge
}

// GetQueueDepthGauge returns the queue depth gauge metric.
func GetQueueDepthGauge() prometheus.Gauge {
	return queueDepthGauge
}

// GetTickDurationHistogram returns the tick duration histogram metric.
func GetTickDurationHistogram() prometheus.Histogram {
	return tickDurationHistogram
}

// GetGCPauseHistogram returns the GC pause histogram metric.
func GetGCPauseHistogram() prometheus.Histogram {
	return gcPauseHistogram
}

// MetricsHandler handles HTTP requests to the /metrics endpoint.
// It returns Prometheus-formatted metrics.
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

