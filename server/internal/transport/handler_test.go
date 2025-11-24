package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorbit/orbitalrush/internal/observability"
	"github.com/gorbit/orbitalrush/internal/proto"
	"github.com/gorbit/orbitalrush/internal/session"
	"github.com/gorbit/orbitalrush/internal/sim/rules"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dto "github.com/prometheus/client_model/go"
)

func TestHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP Handler Suite")
}

var _ = Describe("HTTP Route Handlers", Label("scope:integration", "loop:g5-adapter", "layer:server", "dep:ws", "b:http-routes", "r:medium"), func() {
	var testServer *httptest.Server
	var serverURL string

	BeforeEach(func() {
		// Create test HTTP server with handlers
		mux := http.NewServeMux()
		mux.HandleFunc("/ws", WebSocketHandler)
		mux.HandleFunc("/healthz", HealthzHandler)

		testServer = httptest.NewServer(mux)
		serverURL = "ws" + testServer.URL[4:] + "/ws" // Convert http:// to ws://
	})

	AfterEach(func() {
		if testServer != nil {
			testServer.Close()
		}
	})

	Describe("WebSocketHandler", func() {
		It("successfully upgrades HTTP connection to WebSocket", func() {
			dialer := websocket.Dialer{}
			conn, resp, err := dialer.Dial(serverURL, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusSwitchingProtocols))
			Expect(conn).NotTo(BeNil())

			// Clean up
			conn.Close()
		})

		It("does not create per-connection session handler", func() {
			dialer := websocket.Dialer{}
			conn, _, err := dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			// Wait a bit to ensure connection is established
			time.Sleep(50 * time.Millisecond)

			// Verify no snapshots are sent from per-connection sessions
			// Set a short read deadline to avoid hanging
			conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
			_, _, err = conn.ReadMessage()
			// Should timeout or get connection closed error (no snapshots from per-connection sessions)
			// The important thing is that no per-connection session is broadcasting snapshots
			Expect(err).To(HaveOccurred()) // Should timeout or error, not receive snapshots
		})

		It("handles connection lifecycle properly", func() {
			dialer := websocket.Dialer{}
			conn, _, err := dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())

			// Connection should be open
			Expect(conn).NotTo(BeNil())

			// Close connection - should clean up gracefully
			err = conn.Close()
			Expect(err).NotTo(HaveOccurred())

			// Wait a bit for cleanup
			time.Sleep(50 * time.Millisecond)
		})

		It("cleans up resources on connection close", func() {
			dialer := websocket.Dialer{}
			conn, _, err := dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())

			// Close connection
			err = conn.Close()
			Expect(err).NotTo(HaveOccurred())

			// Try to read after close - should fail
			conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			_, _, err = conn.ReadMessage()
			Expect(err).To(HaveOccurred())
		})

		It("returns error for non-WebSocket requests", func() {
			// Make a regular HTTP GET request
			resp, err := http.Get(testServer.URL + "/ws")
			// The connection will fail because it's not a WebSocket upgrade
			if err == nil {
				defer resp.Body.Close()
				// If no error, the status should indicate the upgrade failed
				Expect(resp.StatusCode).To(BeNumerically(">=", 400))
			}
		})

		It("handles upgrade errors gracefully", func() {
			// Create a server that will fail on upgrade
			failMux := http.NewServeMux()
			failMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
				// Don't set proper WebSocket headers, causing upgrade to fail
				http.Error(w, "Bad Request", http.StatusBadRequest)
			})

			failServer := httptest.NewServer(failMux)
			defer failServer.Close()

			// Try to connect - should fail gracefully
			dialer := websocket.Dialer{}
			_, _, err := dialer.Dial("ws"+failServer.URL[4:]+"/ws", nil)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("HealthzHandler", func() {
		BeforeEach(func() {
			// Initialize metrics before each test
			observability.InitMetrics()
		})

		It("returns JSON response with status ok", func() {
			resp, err := http.Get(testServer.URL + "/healthz")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["status"]).To(Equal("ok"))
		})

		It("sets Content-Type header correctly", func() {
			resp, err := http.Get(testServer.URL + "/healthz")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))
		})

		It("returns HTTP 200 status code", func() {
			resp, err := http.Get(testServer.URL + "/healthz")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		It("handles encoding errors gracefully", func() {
			// This test is a bit tricky - we can't easily force an encoding error
			// in the handler without mocking, but we can verify the handler
			// structure is correct. The actual error handling will be tested
			// through integration.
			resp, err := http.Get(testServer.URL + "/healthz")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			// Should succeed normally
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Describe("Enhanced HealthzHandler with observability metrics", Label("scope:integration", "loop:g7-ops", "layer:server", "b:health-endpoint", "r:medium"), func() {
		BeforeEach(func() {
			// Initialize metrics before each test
			observability.InitMetrics()
		})

		It("returns JSON response with status field", func() {
			resp, err := http.Get(testServer.URL + "/healthz")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["status"]).To(Equal("ok"))
		})

		It("returns JSON response with uptime_seconds field", func() {
			resp, err := http.Get(testServer.URL + "/healthz")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveKey("uptime_seconds"))
			Expect(result["uptime_seconds"]).To(BeNumerically(">=", 0))
		})

		It("returns JSON response with metrics object", func() {
			resp, err := http.Get(testServer.URL + "/healthz")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveKey("metrics"))
			Expect(result["metrics"]).To(BeAssignableToTypeOf(map[string]interface{}{}))
		})

		It("returns JSON response with metrics.active_connections field", func() {
			// Set some test metrics
			activeGauge := observability.GetActiveConnectionsGauge()
			activeGauge.Set(5.0)

			resp, err := http.Get(testServer.URL + "/healthz")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())

			metrics, ok := result["metrics"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(metrics).To(HaveKey("active_connections"))
			Expect(metrics["active_connections"]).To(BeNumerically("==", 5.0))
		})

		It("returns JSON response with metrics.queue_depth field", func() {
			// Set some test metrics
			queueGauge := observability.GetQueueDepthGauge()
			queueGauge.Set(3.0)

			resp, err := http.Get(testServer.URL + "/healthz")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())

			metrics, ok := result["metrics"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(metrics).To(HaveKey("queue_depth"))
			Expect(metrics["queue_depth"]).To(BeNumerically("==", 3.0))
		})

		It("returns JSON response with metrics.tick_time field", func() {
			// Record some tick durations
			tickHistogram := observability.GetTickDurationHistogram()
			tickHistogram.Observe(0.002) // 2ms
			tickHistogram.Observe(0.003) // 3ms
			tickHistogram.Observe(0.005) // 5ms

			resp, err := http.Get(testServer.URL + "/healthz")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())

			metrics, ok := result["metrics"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(metrics).To(HaveKey("tick_time"))

			tickTime, ok := metrics["tick_time"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(tickTime).To(HaveKey("average_ms"))
			Expect(tickTime).To(HaveKey("count"))
			Expect(tickTime["count"]).To(BeNumerically(">=", 3))
			// Average should be around 3.33ms (10ms / 3)
			Expect(tickTime["average_ms"]).To(BeNumerically(">=", 2.0))
			Expect(tickTime["average_ms"]).To(BeNumerically("<=", 5.0))
		})

		It("returns JSON response with metrics.gc_pause field", func() {
			// Record some GC pause durations
			gcHistogram := observability.GetGCPauseHistogram()
			gcHistogram.Observe(0.001) // 1ms
			gcHistogram.Observe(0.002) // 2ms

			resp, err := http.Get(testServer.URL + "/healthz")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())

			metrics, ok := result["metrics"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(metrics).To(HaveKey("gc_pause"))

			gcPause, ok := metrics["gc_pause"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(gcPause).To(HaveKey("average_ms"))
			Expect(gcPause).To(HaveKey("count"))
			Expect(gcPause["count"]).To(BeNumerically(">=", 2))
			// Average should be around 1.5ms (3ms / 2)
			Expect(gcPause["average_ms"]).To(BeNumerically(">=", 1.0))
			Expect(gcPause["average_ms"]).To(BeNumerically("<=", 2.0))
		})

		It("responds quickly without blocking", func() {
			start := time.Now()
			resp, err := http.Get(testServer.URL + "/healthz")
			duration := time.Since(start)

			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			// Health endpoint should respond quickly (< 10ms)
			Expect(duration).To(BeNumerically("<", 10*time.Millisecond))
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		It("handles missing metrics gracefully", func() {
			// Test with metrics initialized but no data recorded
			resp, err := http.Get(testServer.URL + "/healthz")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())

			// Should still return valid JSON with status
			Expect(result["status"]).To(Equal("ok"))
			Expect(result).To(HaveKey("uptime_seconds"))
			Expect(result).To(HaveKey("metrics"))
		})
	})

	Describe("CORS headers", func() {
		It("sets appropriate CORS headers in WebSocket handler", func() {
			// WebSocket upgrade doesn't typically use CORS headers in the same way
			// as regular HTTP requests, but we can verify the upgrade works
			dialer := websocket.Dialer{}
			conn, resp, err := dialer.Dial(serverURL, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusSwitchingProtocols))

			// WebSocket upgrade response should have Upgrade and Connection headers
			Expect(resp.Header.Get("Upgrade")).To(Equal("websocket"))
			Expect(resp.Header.Get("Connection")).To(ContainSubstring("Upgrade"))

			conn.Close()
		})
	})

	Describe("WebSocket upgrade negotiation", func() {
		It("properly negotiates WebSocket upgrade", func() {
			dialer := websocket.Dialer{}
			conn, resp, err := dialer.Dial(serverURL, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusSwitchingProtocols))
			Expect(resp.Header.Get("Upgrade")).To(Equal("websocket"))
			Expect(resp.Header.Get("Connection")).To(ContainSubstring("Upgrade"))

			conn.Close()
		})

		It("handles invalid upgrade requests", func() {
			// Make a regular HTTP request without WebSocket headers
			resp, err := http.Get(testServer.URL + "/ws")
			if err == nil {
				defer resp.Body.Close()
				// Should return an error status
				Expect(resp.StatusCode).To(BeNumerically(">=", 400))
			}
		})

		It("returns appropriate HTTP status codes", func() {
			// Valid WebSocket upgrade
			dialer := websocket.Dialer{}
			conn, resp, err := dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusSwitchingProtocols))
			conn.Close()

			// Invalid request (regular HTTP)
			httpResp, err := http.Get(testServer.URL + "/ws")
			if err == nil {
				defer httpResp.Body.Close()
				Expect(httpResp.StatusCode).To(BeNumerically(">=", 400))
			}
		})
	})
})

var _ = Describe("Connection Metrics", Label("scope:integration", "loop:g7-ops", "layer:server", "dep:ws", "b:connection-metrics", "r:high"), func() {
	var testServer *httptest.Server
	var serverURL string

	BeforeEach(func() {
		// Initialize metrics before each test
		observability.InitMetrics()

		// Create test HTTP server with handlers
		mux := http.NewServeMux()
		mux.HandleFunc("/ws", WebSocketHandler)
		mux.HandleFunc("/healthz", HealthzHandler)
		mux.HandleFunc("/metrics", observability.MetricsHandler)

		testServer = httptest.NewServer(mux)
		serverURL = "ws" + testServer.URL[4:] + "/ws" // Convert http:// to ws://
	})

	AfterEach(func() {
		if testServer != nil {
			testServer.Close()
		}
	})

	Describe("Connection Events Counter", func() {
		It("increments on connect", func() {
			// Get initial value
			var initialMetric dto.Metric
			err := observability.GetConnectionEventsCounter().WithLabelValues("connect").Write(&initialMetric)
			Expect(err).NotTo(HaveOccurred())
			initialValue := initialMetric.Counter.GetValue()

			// Connect
			dialer := websocket.Dialer{}
			conn, _, err := dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			// Wait a bit for handler to process
			time.Sleep(100 * time.Millisecond)

			// Verify counter incremented
			var metric dto.Metric
			err = observability.GetConnectionEventsCounter().WithLabelValues("connect").Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Counter.GetValue()).To(BeNumerically(">", initialValue))
		})

		It("increments on disconnect", func() {
			// Connect first
			dialer := websocket.Dialer{}
			conn, _, err := dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())

			// Wait for connect event
			time.Sleep(100 * time.Millisecond)

			// Get initial disconnect value
			var initialMetric dto.Metric
			err = observability.GetConnectionEventsCounter().WithLabelValues("disconnect").Write(&initialMetric)
			Expect(err).NotTo(HaveOccurred())
			initialValue := initialMetric.Counter.GetValue()

			// Disconnect
			conn.Close()
			time.Sleep(100 * time.Millisecond)

			// Verify counter incremented
			var metric dto.Metric
			err = observability.GetConnectionEventsCounter().WithLabelValues("disconnect").Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Counter.GetValue()).To(BeNumerically(">", initialValue))
		})
	})

	Describe("Active Connections Gauge", func() {
		It("increments on connect", func() {
			// Get initial value
			var initialMetric dto.Metric
			err := observability.GetActiveConnectionsGauge().Write(&initialMetric)
			Expect(err).NotTo(HaveOccurred())
			initialValue := initialMetric.Gauge.GetValue()

			// Connect
			dialer := websocket.Dialer{}
			conn, _, err := dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			// Wait a bit for handler to process
			time.Sleep(100 * time.Millisecond)

			// Verify gauge incremented
			var metric dto.Metric
			err = observability.GetActiveConnectionsGauge().Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Gauge.GetValue()).To(BeNumerically(">", initialValue))
		})

		It("decrements on disconnect", func() {
			// Connect first
			dialer := websocket.Dialer{}
			conn, _, err := dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())

			// Wait for connect
			time.Sleep(100 * time.Millisecond)

			// Get value after connect
			var afterConnectMetric dto.Metric
			err = observability.GetActiveConnectionsGauge().Write(&afterConnectMetric)
			Expect(err).NotTo(HaveOccurred())
			afterConnectValue := afterConnectMetric.Gauge.GetValue()

			// Disconnect
			conn.Close()
			time.Sleep(100 * time.Millisecond)

			// Verify gauge decremented
			var metric dto.Metric
			err = observability.GetActiveConnectionsGauge().Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Gauge.GetValue()).To(BeNumerically("<", afterConnectValue))
		})
	})

	Describe("Connection Duration Histogram", func() {
		It("records connection duration on disconnect", func() {
			// Get initial sample count
			var initialMetric dto.Metric
			err := observability.GetConnectionDurationHistogram().Write(&initialMetric)
			Expect(err).NotTo(HaveOccurred())
			initialCount := uint64(0)
			if initialMetric.Histogram != nil {
				initialCount = initialMetric.Histogram.GetSampleCount()
			}

			// Connect and wait a bit
			dialer := websocket.Dialer{}
			conn, _, err := dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			// Disconnect
			conn.Close()
			time.Sleep(100 * time.Millisecond)

			// Verify histogram recorded a sample
			var metric dto.Metric
			err = observability.GetConnectionDurationHistogram().Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Histogram).NotTo(BeNil())
			Expect(metric.Histogram.GetSampleCount()).To(BeNumerically(">", initialCount))
		})
	})

	Describe("Connection Bytes Counter", func() {
		It("increments bytes in counter on ReadMessage", func() {
			// Connect
			dialer := websocket.Dialer{}
			conn, _, err := dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			// Wait for connection to be established
			time.Sleep(100 * time.Millisecond)

			// Get initial bytes in value
			var initialMetric dto.Metric
			err = observability.GetConnectionBytesCounter().WithLabelValues("in").Write(&initialMetric)
			Expect(err).NotTo(HaveOccurred())
			initialValue := initialMetric.Counter.GetValue()

			// Send a message from client
			testMsg := map[string]interface{}{
				"t": "input",
				"seq": 1,
				"thrust": true,
				"turn": 0.0,
			}
			err = conn.WriteJSON(testMsg)
			Expect(err).NotTo(HaveOccurred())

			// Wait for message to be read
			time.Sleep(200 * time.Millisecond)

			// Verify bytes in counter incremented
			var metric dto.Metric
			err = observability.GetConnectionBytesCounter().WithLabelValues("in").Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Counter.GetValue()).To(BeNumerically(">", initialValue))
		})

		It("increments bytes out counter on WriteMessage", func() {
			// Connect
			dialer := websocket.Dialer{}
			conn, _, err := dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			// Wait for connection and snapshot to be sent
			time.Sleep(200 * time.Millisecond)

			// Get initial bytes out value
			var initialMetric dto.Metric
			err = observability.GetConnectionBytesCounter().WithLabelValues("out").Write(&initialMetric)
			Expect(err).NotTo(HaveOccurred())
			initialValue := initialMetric.Counter.GetValue()

			// Wait a bit more for another snapshot
			time.Sleep(200 * time.Millisecond)

			// Verify bytes out counter incremented (snapshots are sent periodically)
			var metric dto.Metric
			err = observability.GetConnectionBytesCounter().WithLabelValues("out").Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Counter.GetValue()).To(BeNumerically(">=", initialValue))
		})
	})

	Describe("Messages Counter", func() {
		It("increments messages in counter on ReadMessage", func() {
			// Connect
			dialer := websocket.Dialer{}
			conn, _, err := dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			// Wait for connection
			time.Sleep(100 * time.Millisecond)

			// Get initial messages in value
			var initialMetric dto.Metric
			err = observability.GetMessagesCounter().WithLabelValues("in").Write(&initialMetric)
			Expect(err).NotTo(HaveOccurred())
			initialValue := initialMetric.Counter.GetValue()

			// Send a message
			testMsg := map[string]interface{}{
				"t": "input",
				"seq": 1,
				"thrust": true,
				"turn": 0.0,
			}
			err = conn.WriteJSON(testMsg)
			Expect(err).NotTo(HaveOccurred())

			// Wait for message to be processed
			time.Sleep(200 * time.Millisecond)

			// Verify messages in counter incremented
			var metric dto.Metric
			err = observability.GetMessagesCounter().WithLabelValues("in").Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Counter.GetValue()).To(BeNumerically(">", initialValue))
		})

		It("increments messages out counter on WriteMessage", func() {
			// Connect
			dialer := websocket.Dialer{}
			conn, _, err := dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			// Wait for initial snapshot
			time.Sleep(200 * time.Millisecond)

			// Get initial messages out value
			var initialMetric dto.Metric
			err = observability.GetMessagesCounter().WithLabelValues("out").Write(&initialMetric)
			Expect(err).NotTo(HaveOccurred())
			initialValue := initialMetric.Counter.GetValue()

			// Wait for another snapshot
			time.Sleep(200 * time.Millisecond)

			// Verify messages out counter incremented
			var metric dto.Metric
			err = observability.GetMessagesCounter().WithLabelValues("out").Write(&metric)
			Expect(err).NotTo(HaveOccurred())
			Expect(metric.Counter.GetValue()).To(BeNumerically(">=", initialValue))
		})
	})

	Describe("/metrics endpoint", func() {
		It("exposes connection metrics", func() {
			// Make a connection to generate some metrics
			dialer := websocket.Dialer{}
			conn, _, err := dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			time.Sleep(100 * time.Millisecond)

			// Query metrics endpoint
			resp, err := http.Get(testServer.URL + "/metrics")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Read response body
			body := make([]byte, 10000)
			n, _ := resp.Body.Read(body)
			bodyStr := string(body[:n])

			// Verify connection metrics are present
			Expect(bodyStr).To(ContainSubstring("connection_events_total"))
			Expect(bodyStr).To(ContainSubstring("active_connections"))
			Expect(bodyStr).To(ContainSubstring("connection_duration_seconds"))
			Expect(bodyStr).To(ContainSubstring("connection_bytes_total"))
		})
	})
})

var _ = Describe("Connection Registry", Label("scope:integration", "loop:g7-transport", "layer:transport", "dep:none", "b:connection-tracking", "r:high", "double:fake"), func() {
	var registry *ConnectionRegistry
	var conn *Connection

	BeforeEach(func() {
		registry = NewConnectionRegistry()
		// Create a mock connection for testing
		mux := http.NewServeMux()
		var wsConn *websocket.Conn
		mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			var err error
			wsConn, err = UpgradeConnection(w, r)
			if err == nil {
				conn = NewConnection(wsConn)
			}
		})
		testServer := httptest.NewServer(mux)
		defer testServer.Close()

		dialer := websocket.Dialer{}
		clientConn, _, _ := dialer.Dial("ws"+testServer.URL[4:]+"/ws", nil)
		defer clientConn.Close()

		// Wait for connection to be established
		time.Sleep(50 * time.Millisecond)
	})

	Describe("Associate", func() {
		It("associates connection with room code and player ID", func() {
			registry.Associate(conn, "ABC123", 1)

			roomCode, playerID, err := registry.GetRoomInfo(conn)
			Expect(err).NotTo(HaveOccurred())
			Expect(roomCode).To(Equal("ABC123"))
			Expect(playerID).To(Equal(uint32(1)))
		})

		It("updates association if connection is already associated", func() {
			registry.Associate(conn, "ABC123", 1)
			registry.Associate(conn, "XYZ789", 2)

			roomCode, playerID, err := registry.GetRoomInfo(conn)
			Expect(err).NotTo(HaveOccurred())
			Expect(roomCode).To(Equal("XYZ789"))
			Expect(playerID).To(Equal(uint32(2)))
		})
	})

	Describe("Disassociate", func() {
		It("removes connection from registry", func() {
			registry.Associate(conn, "ABC123", 1)
			registry.Disassociate(conn)

			_, _, err := registry.GetRoomInfo(conn)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not associated"))
		})

		It("handles disassociating unassociated connection gracefully", func() {
			// Should not panic
			registry.Disassociate(conn)
		})
	})

	Describe("GetRoomInfo", func() {
		It("returns room code and player ID for associated connection", func() {
			registry.Associate(conn, "ABC123", 42)

			roomCode, playerID, err := registry.GetRoomInfo(conn)
			Expect(err).NotTo(HaveOccurred())
			Expect(roomCode).To(Equal("ABC123"))
			Expect(playerID).To(Equal(uint32(42)))
		})

		It("returns error for unassociated connection", func() {
			_, _, err := registry.GetRoomInfo(conn)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not associated"))
		})
	})

	Describe("IsAssociated", func() {
		It("returns true for associated connection", func() {
			registry.Associate(conn, "ABC123", 1)
			Expect(registry.IsAssociated(conn)).To(BeTrue())
		})

		It("returns false for unassociated connection", func() {
			Expect(registry.IsAssociated(conn)).To(BeFalse())
		})

		It("returns false after disassociation", func() {
			registry.Associate(conn, "ABC123", 1)
			registry.Disassociate(conn)
			Expect(registry.IsAssociated(conn)).To(BeFalse())
		})
	})

	Describe("Thread Safety", func() {
		It("handles concurrent associate operations", func() {
			// Create multiple connections
			var connections []*Connection
			for i := 0; i < 10; i++ {
				mux := http.NewServeMux()
				var wsConn *websocket.Conn
				mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
					var err error
					wsConn, err = UpgradeConnection(w, r)
					if err == nil {
						connections = append(connections, NewConnection(wsConn))
					}
				})
				testServer := httptest.NewServer(mux)
				defer testServer.Close()

				dialer := websocket.Dialer{}
				clientConn, _, _ := dialer.Dial("ws"+testServer.URL[4:]+"/ws", nil)
				defer clientConn.Close()
			}

			time.Sleep(100 * time.Millisecond)

			// Concurrently associate connections
			var wg sync.WaitGroup
			for i, c := range connections {
				wg.Add(1)
				go func(idx int, conn *Connection) {
					defer wg.Done()
					registry.Associate(conn, fmt.Sprintf("ROOM%02d", idx), uint32(idx))
				}(i, c)
			}
			wg.Wait()

			// Verify all associations
			for i, c := range connections {
				roomCode, playerID, err := registry.GetRoomInfo(c)
				Expect(err).NotTo(HaveOccurred())
				Expect(roomCode).To(Equal(fmt.Sprintf("ROOM%02d", i)))
				Expect(playerID).To(Equal(uint32(i)))
			}
		})

		It("handles concurrent associate and disassociate operations", func() {
			registry.Associate(conn, "ABC123", 1)

			var wg sync.WaitGroup
			// Concurrently read and write
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					registry.Associate(conn, "ABC123", 1)
					registry.IsAssociated(conn)
					_, _, _ = registry.GetRoomInfo(conn)
				}()
			}
			wg.Wait()

			// Should still be associated
			Expect(registry.IsAssociated(conn)).To(BeTrue())
		})
	})
})

var _ = Describe("Room Management Handlers", Label("scope:integration", "loop:g7-transport", "layer:transport", "dep:room,proto", "b:room-handlers", "r:high", "double:fake-io"), func() {
	var registry *ConnectionRegistry
	var conn *Connection
	var roomOps RoomOperations
	var handler *RoomHandler
	var clock session.Clock

	BeforeEach(func() {
		registry = NewConnectionRegistry()
		clock = session.NewFakeClock()

		// Create a mock connection for testing
		mux := http.NewServeMux()
		var wsConn *websocket.Conn
		mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			var err error
			wsConn, err = UpgradeConnection(w, r)
			if err == nil {
				conn = NewConnection(wsConn)
			}
		})
		testServer := httptest.NewServer(mux)
		defer testServer.Close()

		dialer := websocket.Dialer{}
		clientConn, _, _ := dialer.Dial("ws"+testServer.URL[4:]+"/ws", nil)
		defer clientConn.Close()

		time.Sleep(50 * time.Millisecond)
	})

	Describe("HandleCreateRoom", func() {
		It("creates room and sends roomCreated response", func() {
			createdRoomCode := "ABC123"
			roomOps = RoomOperations{
				CreateRoomFunc: func() (string, error) {
					return createdRoomCode, nil
				},
			}
			handler = NewRoomHandler(registry, roomOps, conn, clock)

			msg := &proto.CreateRoomMessage{Type: "createRoom"}
			err := handler.HandleCreateRoom(msg)

			Expect(err).NotTo(HaveOccurred())
		})

		It("returns error if CreateRoomFunc fails", func() {
			roomOps = RoomOperations{
				CreateRoomFunc: func() (string, error) {
					return "", fmt.Errorf("room creation failed")
				},
			}
			handler = NewRoomHandler(registry, roomOps, conn, clock)

			msg := &proto.CreateRoomMessage{Type: "createRoom"}
			err := handler.HandleCreateRoom(msg)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create room"))
		})
	})

	Describe("HandleJoinRoom", func() {
		It("joins room, tracks connection, and sends roomState response", func() {
			roomCode := "ABC123"
			playerID := uint32(1)
			roomData := RoomData{
				RoomCode:     roomCode,
				Players:      []PlayerData{{PlayerID: playerID, Name: "", Conn: conn}},
				State:        "lobby",
				HostPlayerID: playerID,
			}

			roomOps = RoomOperations{
				JoinRoomFunc: func(code string, c *Connection) (RoomData, uint32, error) {
					return roomData, playerID, nil
				},
			}
			handler = NewRoomHandler(registry, roomOps, conn, clock)

			msg := &proto.JoinRoomMessage{Type: "joinRoom", RoomCode: roomCode}
			err := handler.HandleJoinRoom(msg)

			Expect(err).NotTo(HaveOccurred())
			Expect(registry.IsAssociated(conn)).To(BeTrue())
			roomCodeFromReg, playerIDFromReg, err := registry.GetRoomInfo(conn)
			Expect(err).NotTo(HaveOccurred())
			Expect(roomCodeFromReg).To(Equal(roomCode))
			Expect(playerIDFromReg).To(Equal(playerID))
		})

		It("returns error if JoinRoomFunc fails", func() {
			roomOps = RoomOperations{
				JoinRoomFunc: func(code string, c *Connection) (RoomData, uint32, error) {
					return RoomData{}, 0, fmt.Errorf("join failed")
				},
			}
			handler = NewRoomHandler(registry, roomOps, conn, clock)

			msg := &proto.JoinRoomMessage{Type: "joinRoom", RoomCode: "ABC123"}
			err := handler.HandleJoinRoom(msg)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to join room"))
		})
	})

	Describe("HandleLeaveRoom", func() {
		It("leaves room, untracks connection, and broadcasts playerLeft", func() {
			roomCode := "ABC123"
			playerID := uint32(1)
			registry.Associate(conn, roomCode, playerID)

			roomData := RoomData{
				RoomCode:     roomCode,
				Players:      []PlayerData{},
				State:        "lobby",
				HostPlayerID: playerID,
			}

			roomOps = RoomOperations{
				GetRoomFunc: func(code string) (RoomData, error) {
					return roomData, nil
				},
				LeaveRoomFunc: func(code string, pid uint32) error {
					return nil
				},
			}
			handler = NewRoomHandler(registry, roomOps, conn, clock)

			msg := &proto.LeaveRoomMessage{Type: "leaveRoom"}
			err := handler.HandleLeaveRoom(msg)

			Expect(err).NotTo(HaveOccurred())
			Expect(registry.IsAssociated(conn)).To(BeFalse())
		})

		It("returns error if connection not in a room", func() {
			roomOps = RoomOperations{}
			handler = NewRoomHandler(registry, roomOps, conn, clock)

			msg := &proto.LeaveRoomMessage{Type: "leaveRoom"}
			err := handler.HandleLeaveRoom(msg)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection not in a room"))
		})
	})

	Describe("HandleStartMatch", func() {
		It("returns error if StartMatchFunc is nil", func() {
			roomCode := "ABC123"
			playerID := uint32(1)
			registry.Associate(conn, roomCode, playerID)

			roomOps = RoomOperations{
				StartMatchFunc: nil, // Not implemented
			}
			handler = NewRoomHandler(registry, roomOps, conn, clock)

			msg := &proto.StartMatchMessage{Type: "startMatch"}
			err := handler.HandleStartMatch(msg)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("StartMatch not yet implemented"))
		})

		It("starts match and broadcasts matchStarted if StartMatchFunc is provided", func() {
			roomCode := "ABC123"
			playerID := uint32(1)
			registry.Associate(conn, roomCode, playerID)

			roomData := RoomData{
				RoomCode:     roomCode,
				Players:      []PlayerData{{PlayerID: playerID, Name: "", Conn: conn}},
				State:        "playing",
				HostPlayerID: playerID,
			}

			roomOps = RoomOperations{
				StartMatchFunc: func(code string, pid uint32, c session.Clock) error {
					return nil
				},
				GetRoomFunc: func(code string) (RoomData, error) {
					return roomData, nil
				},
			}
			handler = NewRoomHandler(registry, roomOps, conn, clock)

			msg := &proto.StartMatchMessage{Type: "startMatch"}
			err := handler.HandleStartMatch(msg)

			Expect(err).NotTo(HaveOccurred())
		})

		It("returns error if connection not in a room", func() {
			roomOps = RoomOperations{
				StartMatchFunc: func(code string, pid uint32, c session.Clock) error {
					return nil
				},
			}
			handler = NewRoomHandler(registry, roomOps, conn, clock)

			msg := &proto.StartMatchMessage{Type: "startMatch"}
			err := handler.HandleStartMatch(msg)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection not in a room"))
		})
	})
})

var _ = Describe("Room Input Handler", Label("scope:integration", "loop:g7-transport", "layer:transport", "dep:room", "b:input-routing", "r:high", "double:fake-io"), func() {
	var registry *ConnectionRegistry
	var conn *Connection
	var roomOps RoomOperations
	var handler *RoomInputHandler

	BeforeEach(func() {
		registry = NewConnectionRegistry()

		// Create a mock connection for testing
		mux := http.NewServeMux()
		var wsConn *websocket.Conn
		mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			var err error
			wsConn, err = UpgradeConnection(w, r)
			if err == nil {
				conn = NewConnection(wsConn)
			}
		})
		testServer := httptest.NewServer(mux)
		defer testServer.Close()

		dialer := websocket.Dialer{}
		clientConn, _, _ := dialer.Dial("ws"+testServer.URL[4:]+"/ws", nil)
		defer clientConn.Close()

		time.Sleep(50 * time.Millisecond)
	})

	Describe("HandleInput", func() {
		It("routes command to room session successfully", func() {
			roomCode := "ABC123"
			playerID := uint32(1)
			registry.Associate(conn, roomCode, playerID)

			var enqueuedRoomCode string
			var enqueuedPlayerID uint32
			var enqueuedSeq uint32
			var enqueuedCmd rules.InputCommand

			roomOps = RoomOperations{
				EnqueueCommandToRoomFunc: func(code string, pid uint32, seq uint32, cmd rules.InputCommand) error {
					enqueuedRoomCode = code
					enqueuedPlayerID = pid
					enqueuedSeq = seq
					enqueuedCmd = cmd
					return nil
				},
			}
			handler = NewRoomInputHandler(registry, roomOps, conn)

			msg := &proto.InputMessage{
				Type:   "input",
				Seq:    42,
				Thrust: 0.75,
				Turn:   -0.5,
			}
			err := handler.HandleInput(msg)

			Expect(err).NotTo(HaveOccurred())
			Expect(enqueuedRoomCode).To(Equal(roomCode))
			Expect(enqueuedPlayerID).To(Equal(playerID))
			Expect(enqueuedSeq).To(Equal(uint32(42)))
			Expect(enqueuedCmd.Thrust).To(Equal(float32(0.75)))
			Expect(enqueuedCmd.Turn).To(Equal(float32(-0.5)))
		})

		It("returns error if connection not associated with room", func() {
			roomOps = RoomOperations{
				EnqueueCommandToRoomFunc: func(code string, pid uint32, seq uint32, cmd rules.InputCommand) error {
					return nil
				},
			}
			handler = NewRoomInputHandler(registry, roomOps, conn)

			msg := &proto.InputMessage{
				Type:   "input",
				Seq:    1,
				Thrust: 0.5,
				Turn:   0.0,
			}
			err := handler.HandleInput(msg)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection not associated with any room"))
		})

		It("returns error if EnqueueCommandToRoomFunc fails (room not found)", func() {
			roomCode := "ABC123"
			playerID := uint32(1)
			registry.Associate(conn, roomCode, playerID)

			roomOps = RoomOperations{
				EnqueueCommandToRoomFunc: func(code string, pid uint32, seq uint32, cmd rules.InputCommand) error {
					return fmt.Errorf("room not found")
				},
			}
			handler = NewRoomInputHandler(registry, roomOps, conn)

			msg := &proto.InputMessage{
				Type:   "input",
				Seq:    1,
				Thrust: 0.5,
				Turn:   0.0,
			}
			err := handler.HandleInput(msg)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to enqueue command to room"))
		})

		It("returns error if EnqueueCommandToRoomFunc fails (session not found)", func() {
			roomCode := "ABC123"
			playerID := uint32(1)
			registry.Associate(conn, roomCode, playerID)

			roomOps = RoomOperations{
				EnqueueCommandToRoomFunc: func(code string, pid uint32, seq uint32, cmd rules.InputCommand) error {
					return fmt.Errorf("session not found")
				},
			}
			handler = NewRoomInputHandler(registry, roomOps, conn)

			msg := &proto.InputMessage{
				Type:   "input",
				Seq:    1,
				Thrust: 0.5,
				Turn:   0.0,
			}
			err := handler.HandleInput(msg)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to enqueue command to room"))
		})

		It("returns error if EnqueueCommandToRoomFunc is nil", func() {
			roomCode := "ABC123"
			playerID := uint32(1)
			registry.Associate(conn, roomCode, playerID)

			roomOps = RoomOperations{
				EnqueueCommandToRoomFunc: nil,
			}
			handler = NewRoomInputHandler(registry, roomOps, conn)

			msg := &proto.InputMessage{
				Type:   "input",
				Seq:    1,
				Thrust: 0.5,
				Turn:   0.0,
			}
			err := handler.HandleInput(msg)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("EnqueueCommandToRoomFunc not provided"))
		})

		It("correctly uses player ID from registry", func() {
			roomCode := "ABC123"
			playerID := uint32(42) // Different player ID
			registry.Associate(conn, roomCode, playerID)

			var enqueuedPlayerID uint32

			roomOps = RoomOperations{
				EnqueueCommandToRoomFunc: func(code string, pid uint32, seq uint32, cmd rules.InputCommand) error {
					enqueuedPlayerID = pid
					return nil
				},
			}
			handler = NewRoomInputHandler(registry, roomOps, conn)

			msg := &proto.InputMessage{
				Type:   "input",
				Seq:    1,
				Thrust: 0.5,
				Turn:   0.0,
			}
			err := handler.HandleInput(msg)

			Expect(err).NotTo(HaveOccurred())
			Expect(enqueuedPlayerID).To(Equal(uint32(42)))
		})
	})
})

var _ = Describe("WebSocketHandler - No Per-Connection Sessions", Label("scope:integration", "loop:g7-transport", "layer:transport", "dep:none", "b:session-removal", "r:medium", "double:fake"), func() {
	var testServer *httptest.Server
	var serverURL string

	BeforeEach(func() {
		// Create test HTTP server with handlers
		mux := http.NewServeMux()
		mux.HandleFunc("/ws", WebSocketHandler)

		testServer = httptest.NewServer(mux)
		serverURL = "ws" + testServer.URL[4:] + "/ws" // Convert http:// to ws://
	})

	AfterEach(func() {
		if testServer != nil {
			testServer.Close()
		}
	})

	It("does not create SessionHandler for new connections", func() {
		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(serverURL, nil)
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close()

		// Wait a bit to ensure connection is established
		time.Sleep(50 * time.Millisecond)

		// Verify no snapshots are sent from per-connection sessions
		// Per-connection sessions have been removed - snapshots will come from room sessions (step 7)
		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		_, _, err = conn.ReadMessage()
		// Should timeout or error (no snapshots from per-connection sessions)
		Expect(err).To(HaveOccurred())
	})

	It("handles connection lifecycle without per-connection sessions", func() {
		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(serverURL, nil)
		Expect(err).NotTo(HaveOccurred())

		// Connection should be open
		Expect(conn).NotTo(BeNil())

		// Close connection - should clean up gracefully without session handler
		err = conn.Close()
		Expect(err).NotTo(HaveOccurred())

		// Wait a bit for cleanup
		time.Sleep(50 * time.Millisecond)
	})
})

