package observability

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
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

	// connectionDurationHistogram tracks connection duration
	connectionDurationHistogram prometheus.Histogram

	// connectionBytesCounter tracks bytes in/out
	connectionBytesCounter *prometheus.CounterVec

	// metricsInitialized tracks whether metrics have been initialized
	metricsInitialized bool

	// serverStartTime tracks when the server started (when InitMetrics was called)
	serverStartTime time.Time
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
		if connectionDurationHistogram != nil {
			prometheus.Unregister(connectionDurationHistogram)
		}
		if connectionBytesCounter != nil {
			prometheus.Unregister(connectionBytesCounter)
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

	// Connection duration histogram with buckets for connection lifetimes
	// Buckets: 1s, 10s, 1m, 5m, 15m, 1h
	connectionDurationHistogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "connection_duration_seconds",
			Help:    "Connection duration in seconds",
			Buckets: []float64{1, 10, 60, 300, 900, 3600}, // 1s, 10s, 1m, 5m, 15m, 1h
		},
	)

	// Connection bytes counter
	connectionBytesCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "connection_bytes_total",
			Help: "Total bytes transferred over connections",
		},
		[]string{"direction"}, // direction: in, out
	)

	// Register all metrics
	prometheus.MustRegister(connectionEventsCounter)
	prometheus.MustRegister(messagesCounter)
	prometheus.MustRegister(activeConnectionsGauge)
	prometheus.MustRegister(queueDepthGauge)
	prometheus.MustRegister(tickDurationHistogram)
	prometheus.MustRegister(gcPauseHistogram)
	prometheus.MustRegister(connectionDurationHistogram)
	prometheus.MustRegister(connectionBytesCounter)

	// Record server start time
	serverStartTime = time.Now()

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

// GetConnectionDurationHistogram returns the connection duration histogram metric.
func GetConnectionDurationHistogram() prometheus.Histogram {
	return connectionDurationHistogram
}

// GetConnectionBytesCounter returns the connection bytes counter metric.
func GetConnectionBytesCounter() *prometheus.CounterVec {
	return connectionBytesCounter
}

// UpdateQueueDepth updates the queue depth gauge metric with the current queue size.
func UpdateQueueDepth(size int) {
	if queueDepthGauge != nil {
		queueDepthGauge.Set(float64(size))
	}
}

// MetricsHandler handles HTTP requests to the /metrics endpoint.
// It returns Prometheus-formatted metrics.
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

// HealthMetrics contains summary statistics for health endpoint
type HealthMetrics struct {
	ActiveConnections float64
	QueueDepth        float64
	TickTime          TickTimeStats
	GCPause           GCPauseStats
	UptimeSeconds     float64
}

// TickTimeStats contains tick time statistics
type TickTimeStats struct {
	AverageMs float64
	Count     uint64
}

// GCPauseStats contains GC pause statistics
type GCPauseStats struct {
	AverageMs float64
	Count     uint64
}

// GetHealthMetrics extracts summary statistics from Prometheus metrics.
// Returns a HealthMetrics struct with current metric values.
// If metrics are not initialized, returns zero values.
func GetHealthMetrics() HealthMetrics {
	metrics := HealthMetrics{}

	// Calculate uptime
	if !serverStartTime.IsZero() {
		metrics.UptimeSeconds = time.Since(serverStartTime).Seconds()
	}

	// Extract active connections gauge
	if activeConnectionsGauge != nil {
		var metric dto.Metric
		if err := activeConnectionsGauge.Write(&metric); err == nil && metric.Gauge != nil {
			metrics.ActiveConnections = metric.Gauge.GetValue()
		}
	}

	// Extract queue depth gauge
	if queueDepthGauge != nil {
		var metric dto.Metric
		if err := queueDepthGauge.Write(&metric); err == nil && metric.Gauge != nil {
			metrics.QueueDepth = metric.Gauge.GetValue()
		}
	}

	// Extract tick time histogram statistics
	if tickDurationHistogram != nil {
		var metric dto.Metric
		if err := tickDurationHistogram.Write(&metric); err == nil && metric.Histogram != nil {
			count := metric.Histogram.GetSampleCount()
			sum := metric.Histogram.GetSampleSum()
			metrics.TickTime.Count = count
			if count > 0 {
				metrics.TickTime.AverageMs = (sum / float64(count)) * 1000.0 // Convert to milliseconds
			}
		}
	}

	// Extract GC pause histogram statistics
	if gcPauseHistogram != nil {
		var metric dto.Metric
		if err := gcPauseHistogram.Write(&metric); err == nil && metric.Histogram != nil {
			count := metric.Histogram.GetSampleCount()
			sum := metric.Histogram.GetSampleSum()
			metrics.GCPause.Count = count
			if count > 0 {
				metrics.GCPause.AverageMs = (sum / float64(count)) * 1000.0 // Convert to milliseconds
			}
		}
	}

	return metrics
}

