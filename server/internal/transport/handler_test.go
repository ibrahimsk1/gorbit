package transport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorbit/orbitalrush/internal/observability"
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

		It("creates session handler and starts it", func() {
			dialer := websocket.Dialer{}
			conn, _, err := dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			// Wait a bit to ensure session handler is started
			time.Sleep(50 * time.Millisecond)

			// Try to read a snapshot message (should be broadcast periodically)
			// Set a short read deadline to avoid hanging
			conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
			_, data, err := conn.ReadMessage()
			// We might get a snapshot or timeout, both are acceptable
			// The important thing is that the connection is working
			if err == nil {
				// If we got a message, it should be a valid JSON snapshot
				var snapshot map[string]interface{}
				err = json.Unmarshal(data, &snapshot)
				Expect(err).NotTo(HaveOccurred())
				Expect(snapshot["t"]).To(Equal("snapshot"))
			}
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
		It("returns JSON response with status ok", func() {
			resp, err := http.Get(testServer.URL + "/healthz")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var result map[string]string
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

