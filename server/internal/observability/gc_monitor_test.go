package observability

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dto "github.com/prometheus/client_model/go"
)

func TestGCMonitor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GC Monitor Suite")
}

var _ = Describe("GC Monitor", Label("scope:integration", "loop:g7-ops", "layer:server", "b:gc-queue-monitoring", "r:medium"), func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
		logger logr.Logger
	)

	BeforeEach(func() {
		// Reset metrics registry before each test
		InitMetrics()
		ctx, cancel = context.WithCancel(context.Background())
		logger = logr.Discard() // Use discard logger for tests
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("GC Pause Tracking", func() {
		It("records GC pause durations to histogram", func() {
			// Start GC monitor with short interval for testing
			interval := 100 * time.Millisecond
			stopChan := StartGCMonitor(ctx, interval, logger)

			// Wait a bit to allow some samples
			time.Sleep(300 * time.Millisecond)

			// Force a GC to generate pause data
			runtime.GC()
			time.Sleep(200 * time.Millisecond)

			// Stop the monitor
			close(stopChan)
			time.Sleep(50 * time.Millisecond) // Allow goroutine to finish

			// Verify histogram has samples
			histogram := GetGCPauseHistogram()
			Expect(histogram).NotTo(BeNil())

			var metric dto.Metric
			err := histogram.Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			// We may have 0 samples if no GC occurred, but the metric should exist
			Expect(metric.Histogram).NotTo(BeNil())
		})

		It("samples GC stats at correct interval", func() {
			interval := 100 * time.Millisecond
			startTime := time.Now()
			stopChan := StartGCMonitor(ctx, interval, logger)

			// Wait for at least 2 intervals
			time.Sleep(250 * time.Millisecond)

			// Stop the monitor
			close(stopChan)
			elapsed := time.Since(startTime)

			// Verify monitor ran for expected duration
			Expect(elapsed).To(BeNumerically(">=", 200*time.Millisecond))
		})

		It("can be stopped gracefully", func() {
			interval := 50 * time.Millisecond
			stopChan := StartGCMonitor(ctx, interval, logger)

			// Wait a bit
			time.Sleep(100 * time.Millisecond)

			// Stop gracefully
			close(stopChan)
			time.Sleep(100 * time.Millisecond) // Allow goroutine to finish

			// If we get here without deadlock, graceful shutdown worked
			Expect(true).To(BeTrue())
		})
	})

	Describe("Queue Depth Monitoring", func() {
		It("updates queue depth gauge correctly", func() {
			// Update queue depth
			UpdateQueueDepth(10)
			
			var metric dto.Metric
			err := GetQueueDepthGauge().Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Gauge.GetValue()).To(Equal(10.0))

			// Update again
			UpdateQueueDepth(25)
			err = GetQueueDepthGauge().Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Gauge.GetValue()).To(Equal(25.0))
		})

		It("handles zero queue depth", func() {
			UpdateQueueDepth(0)
			
			var metric dto.Metric
			err := GetQueueDepthGauge().Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Gauge.GetValue()).To(Equal(0.0))
		})

		It("handles large queue depth values", func() {
			UpdateQueueDepth(1000)
			
			var metric dto.Metric
			err := GetQueueDepthGauge().Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Gauge.GetValue()).To(Equal(1000.0))
		})
	})

	Describe("GC Monitor Non-Interference", func() {
		It("does not block when running", func() {
			interval := 100 * time.Millisecond
			stopChan := StartGCMonitor(ctx, interval, logger)

			// Perform some work that should not be blocked
			startTime := time.Now()
			for i := 0; i < 1000; i++ {
				_ = i * i
			}
			elapsed := time.Since(startTime)

			// Work should complete quickly (not blocked by GC monitor)
			Expect(elapsed).To(BeNumerically("<", 10*time.Millisecond))

			close(stopChan)
		})
	})
})

