package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestMetrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metrics Suite")
}

var _ = Describe("Metrics", Label("scope:integration", "loop:g7-ops", "layer:server", "dep:prometheus", "b:metrics-foundation", "r:high"), func() {
	BeforeEach(func() {
		// Reset metrics registry before each test
		InitMetrics()
	})

	Describe("Metrics Initialization", func() {
		It("initializes all metrics successfully", func() {
			Expect(GetConnectionEventsCounter()).NotTo(BeNil())
			Expect(GetMessagesCounter()).NotTo(BeNil())
			Expect(GetActiveConnectionsGauge()).NotTo(BeNil())
			Expect(GetQueueDepthGauge()).NotTo(BeNil())
			Expect(GetTickDurationHistogram()).NotTo(BeNil())
			Expect(GetGCPauseHistogram()).NotTo(BeNil())
		})

		It("registers metrics with Prometheus registry", func() {
			// Try to register the same metric again - should fail because already registered
			err := prometheus.DefaultRegisterer.Register(GetConnectionEventsCounter())
			Expect(err).To(HaveOccurred()) // Should fail because already registered
			// Error message should indicate duplicate registration
			Expect(err.Error()).To(Or(ContainSubstring("duplicate"), ContainSubstring("register"), ContainSubstring("registration")))
		})
	})

	Describe("Connection Events Counter", func() {
		It("can increment connection events", func() {
			counter := GetConnectionEventsCounter()
			counter.WithLabelValues("connect").Inc()
			counter.WithLabelValues("disconnect").Inc()
			counter.WithLabelValues("error").Inc()

			// Verify metric values
			var metric dto.Metric
			err := counter.WithLabelValues("connect").Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Counter.GetValue()).To(Equal(1.0))
		})
	})

	Describe("Messages Counter", func() {
		It("can increment message counts", func() {
			counter := GetMessagesCounter()
			counter.WithLabelValues("in").Inc()
			counter.WithLabelValues("out").Add(2.0)

			var metric dto.Metric
			err := counter.WithLabelValues("in").Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Counter.GetValue()).To(Equal(1.0))

			err = counter.WithLabelValues("out").Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Counter.GetValue()).To(Equal(2.0))
		})
	})

	Describe("Active Connections Gauge", func() {
		It("can set and get active connections", func() {
			gauge := GetActiveConnectionsGauge()
			gauge.Set(5.0)
			gauge.Inc()
			gauge.Dec()

			var metric dto.Metric
			err := gauge.Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Gauge.GetValue()).To(Equal(5.0))
		})
	})

	Describe("Queue Depth Gauge", func() {
		It("can set queue depth", func() {
			gauge := GetQueueDepthGauge()
			gauge.Set(10.0)

			var metric dto.Metric
			err := gauge.Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Gauge.GetValue()).To(Equal(10.0))
		})
	})

	Describe("Tick Duration Histogram", func() {
		It("can record tick durations", func() {
			histogram := GetTickDurationHistogram()
			histogram.Observe(0.005) // 5ms
			histogram.Observe(0.01)  // 10ms
			histogram.Observe(0.05)  // 50ms

			var metric dto.Metric
			err := histogram.Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Histogram.GetSampleCount()).To(Equal(uint64(3)))
		})
	})

	Describe("GC Pause Histogram", func() {
		It("can record GC pause durations", func() {
			histogram := GetGCPauseHistogram()
			histogram.Observe(0.001) // 1ms
			histogram.Observe(0.002) // 2ms

			var metric dto.Metric
			err := histogram.Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Histogram.GetSampleCount()).To(Equal(uint64(2)))
		})
	})

	Describe("/metrics endpoint", func() {
		It("returns valid Prometheus format", func() {
			// Set some metric values to ensure they appear in output
			GetConnectionEventsCounter().WithLabelValues("connect").Inc()
			GetMessagesCounter().WithLabelValues("in").Inc()
			GetActiveConnectionsGauge().Set(2.0)
			GetQueueDepthGauge().Set(5.0)
			GetTickDurationHistogram().Observe(0.01)

			// Create request
			req := httptest.NewRequest("GET", "/metrics", nil)
			w := httptest.NewRecorder()

			// Call metrics handler
			MetricsHandler(w, req)

			// Verify response
			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Header().Get("Content-Type")).To(ContainSubstring("text/plain"))

			body := w.Body.String()
			Expect(body).NotTo(BeEmpty())

			// Verify Prometheus format - should contain metric names
			Expect(body).To(ContainSubstring("connection_events_total"))
			Expect(body).To(ContainSubstring("messages_total"))
			Expect(body).To(ContainSubstring("active_connections"))
			Expect(body).To(ContainSubstring("queue_depth"))
			Expect(body).To(ContainSubstring("tick_duration_seconds"))

			// Verify metric types
			Expect(body).To(ContainSubstring("# TYPE connection_events_total counter"))
			Expect(body).To(ContainSubstring("# TYPE messages_total counter"))
			Expect(body).To(ContainSubstring("# TYPE active_connections gauge"))
			Expect(body).To(ContainSubstring("# TYPE queue_depth gauge"))
			Expect(body).To(ContainSubstring("# TYPE tick_duration_seconds histogram"))
		})

		It("includes HELP comments for metrics", func() {
			// Set some metric values to ensure they appear in output
			GetConnectionEventsCounter().WithLabelValues("connect").Inc()
			GetMessagesCounter().WithLabelValues("in").Inc()

			req := httptest.NewRequest("GET", "/metrics", nil)
			w := httptest.NewRecorder()

			MetricsHandler(w, req)

			body := w.Body.String()
			Expect(body).To(ContainSubstring("# HELP connection_events_total"))
			Expect(body).To(ContainSubstring("# HELP messages_total"))
			Expect(body).To(ContainSubstring("# HELP active_connections"))
			Expect(body).To(ContainSubstring("# HELP queue_depth"))
			Expect(body).To(ContainSubstring("# HELP tick_duration_seconds"))
		})

		It("returns all expected metric names", func() {
			// Set some metric values to ensure they appear in output
			GetConnectionEventsCounter().WithLabelValues("connect").Inc()
			GetMessagesCounter().WithLabelValues("in").Inc()
			GetActiveConnectionsGauge().Set(1.0)
			GetQueueDepthGauge().Set(1.0)
			GetTickDurationHistogram().Observe(0.001)
			GetGCPauseHistogram().Observe(0.0001)

			req := httptest.NewRequest("GET", "/metrics", nil)
			w := httptest.NewRecorder()

			MetricsHandler(w, req)

			body := w.Body.String()
			expectedMetrics := []string{
				"connection_events_total",
				"messages_total",
				"active_connections",
				"queue_depth",
				"tick_duration_seconds",
				"gc_pause_seconds",
			}

			for _, metricName := range expectedMetrics {
				Expect(body).To(ContainSubstring(metricName), "should contain metric: %s", metricName)
			}
		})
	})
})

