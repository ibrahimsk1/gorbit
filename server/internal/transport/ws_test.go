package transport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWebSocket(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WebSocket Connection Suite")
}

var _ = Describe("WebSocket Connection Handler", Label("scope:integration", "loop:g5-adapter", "layer:server", "dep:ws", "b:ws-connection", "r:high"), func() {
	var testServer *httptest.Server
	var serverURL string

	BeforeEach(func() {
		// Create test HTTP server with WebSocket endpoint
		mux := http.NewServeMux()
		mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			conn, err := UpgradeConnection(w, r)
			if err != nil {
				// UpgradeConnection may have already written headers on error
				// Just return without calling http.Error to avoid superfluous WriteHeader
				return
			}
			defer conn.Close()
		})

		testServer = httptest.NewServer(mux)
		serverURL = "ws" + testServer.URL[4:] + "/ws" // Convert http:// to ws://
	})

	AfterEach(func() {
		if testServer != nil {
			testServer.Close()
		}
	})

	Describe("UpgradeConnection", func() {
		It("successfully upgrades HTTP connection to WebSocket", func() {
			// Create a WebSocket client
			dialer := websocket.Dialer{}
			conn, resp, err := dialer.Dial(serverURL, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusSwitchingProtocols))
			Expect(conn).NotTo(BeNil())

			// Clean up
			conn.Close()
		})

		It("returns error for non-WebSocket requests", func() {
			// Make a regular HTTP GET request
			resp, err := http.Get(testServer.URL + "/ws")
			// The connection will fail because it's not a WebSocket upgrade
			// UpgradeConnection will return an error, but we don't write an HTTP error response
			// to avoid superfluous WriteHeader calls. The connection will just close.
			if err == nil {
				resp.Body.Close()
				// If no error, the status should indicate the upgrade failed
				Expect(resp.StatusCode).To(BeNumerically(">=", 400))
			}
			// Either way, the request should not succeed as a WebSocket upgrade
		})

		It("sets appropriate headers", func() {
			dialer := websocket.Dialer{}
			conn, resp, err := dialer.Dial(serverURL, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Header.Get("Upgrade")).To(Equal("websocket"))
			Expect(resp.Header.Get("Connection")).To(ContainSubstring("Upgrade"))

			conn.Close()
		})
	})

	Describe("Connection ReadMessage", func() {
		var conn *websocket.Conn
		var clientConn *websocket.Conn

		BeforeEach(func() {
			// Create a test server that accepts connections
			mux := http.NewServeMux()
			mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
				var err error
				conn, err = UpgradeConnection(w, r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			})

			testServer = httptest.NewServer(mux)
			serverURL = "ws" + testServer.URL[4:] + "/ws"

			// Connect client
			dialer := websocket.Dialer{}
			var err error
			clientConn, _, err = dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if conn != nil {
				conn.Close()
			}
			if clientConn != nil {
				clientConn.Close()
			}
		})

		It("reads JSON text messages correctly", func() {
			// Wait for connection to be established
			Eventually(func() bool {
				return conn != nil
			}).Should(BeTrue())

			// Send a JSON message from client
			testMessage := map[string]interface{}{
				"t": "test",
				"data": "hello",
			}
			err := clientConn.WriteJSON(testMessage)
			Expect(err).NotTo(HaveOccurred())

			// Read message using Connection wrapper
			connection := NewConnection(conn)
			data, err := connection.ReadMessage()
			Expect(err).NotTo(HaveOccurred())

			// Verify message content
			var received map[string]interface{}
			err = json.Unmarshal(data, &received)
			Expect(err).NotTo(HaveOccurred())
			Expect(received["t"]).To(Equal("test"))
			Expect(received["data"]).To(Equal("hello"))
		})

		It("handles connection close gracefully", func() {
			Eventually(func() bool {
				return conn != nil
			}).Should(BeTrue())

			connection := NewConnection(conn)

			// Close client connection
			clientConn.Close()

			// Try to read - should detect close
			_, err := connection.ReadMessage()
			Expect(err).To(HaveOccurred())
			Expect(websocket.IsCloseError(err, websocket.CloseNormalClosure) || 
				websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway)).To(BeTrue())
		})

		It("returns error for binary messages (should only accept text)", func() {
			Eventually(func() bool {
				return conn != nil
			}).Should(BeTrue())

			// Send binary message
			err := clientConn.WriteMessage(websocket.BinaryMessage, []byte("binary data"))
			Expect(err).NotTo(HaveOccurred())

			connection := NewConnection(conn)
			_, err = connection.ReadMessage()
			// Should handle binary gracefully (either accept or reject)
			// The implementation may accept binary but we expect text
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Connection WriteMessage", func() {
		var conn *websocket.Conn
		var clientConn *websocket.Conn

		BeforeEach(func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
				var err error
				conn, err = UpgradeConnection(w, r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			})

			testServer = httptest.NewServer(mux)
			serverURL = "ws" + testServer.URL[4:] + "/ws"

			dialer := websocket.Dialer{}
			var err error
			clientConn, _, err = dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if conn != nil {
				conn.Close()
			}
			if clientConn != nil {
				clientConn.Close()
			}
		})

		It("writes JSON text messages correctly", func() {
			Eventually(func() bool {
				return conn != nil
			}).Should(BeTrue())

			connection := NewConnection(conn)

			// Write JSON message
			testMessage := map[string]interface{}{
				"t": "snapshot",
				"tick": 42,
			}
			messageBytes, err := json.Marshal(testMessage)
			Expect(err).NotTo(HaveOccurred())

			err = connection.WriteMessage(messageBytes)
			Expect(err).NotTo(HaveOccurred())

			// Read from client
			var received map[string]interface{}
			err = clientConn.ReadJSON(&received)
			Expect(err).NotTo(HaveOccurred())
			Expect(received["t"]).To(Equal("snapshot"))
			Expect(received["tick"]).To(Equal(float64(42))) // JSON numbers are float64
		})
	})

	Describe("Connection Close", func() {
		var conn *websocket.Conn
		var clientConn *websocket.Conn

		BeforeEach(func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
				var err error
				conn, err = UpgradeConnection(w, r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			})

			testServer = httptest.NewServer(mux)
			serverURL = "ws" + testServer.URL[4:] + "/ws"

			dialer := websocket.Dialer{}
			var err error
			clientConn, _, err = dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if conn != nil {
				conn.Close()
			}
			if clientConn != nil {
				clientConn.Close()
			}
		})

		It("closes connection gracefully", func() {
			Eventually(func() bool {
				return conn != nil
			}).Should(BeTrue())

			connection := NewConnection(conn)

			// Close connection
			err := connection.Close()
			Expect(err).NotTo(HaveOccurred())

			// Verify connection is closed
			_, _, err = conn.ReadMessage()
			Expect(err).To(HaveOccurred())
		})

		It("can be called multiple times safely", func() {
			Eventually(func() bool {
				return conn != nil
			}).Should(BeTrue())

			connection := NewConnection(conn)

			// Close multiple times
			err := connection.Close()
			Expect(err).NotTo(HaveOccurred())

			err = connection.Close()
			Expect(err).NotTo(HaveOccurred()) // Should not error on second close
		})
	})

	Describe("Connection Lifecycle", func() {
		var conn *websocket.Conn
		var clientConn *websocket.Conn

		BeforeEach(func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
				var err error
				conn, err = UpgradeConnection(w, r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			})

			testServer = httptest.NewServer(mux)
			serverURL = "ws" + testServer.URL[4:] + "/ws"

			dialer := websocket.Dialer{}
			var err error
			clientConn, _, err = dialer.Dial(serverURL, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if conn != nil {
				conn.Close()
			}
			if clientConn != nil {
				clientConn.Close()
			}
		})

		It("handles concurrent read/write operations", func() {
			Eventually(func() bool {
				return conn != nil
			}).Should(BeTrue())

			connection := NewConnection(conn)

			// Start concurrent read
			readDone := make(chan error, 1)
			go func() {
				_, err := connection.ReadMessage()
				readDone <- err
			}()

			// Write from client
			testMessage := map[string]interface{}{"t": "test"}
			err := clientConn.WriteJSON(testMessage)
			Expect(err).NotTo(HaveOccurred())

			// Wait for read to complete
			select {
			case err := <-readDone:
				Expect(err).NotTo(HaveOccurred())
			case <-time.After(1 * time.Second):
				Fail("Read operation timed out")
			}
		})
	})
})

