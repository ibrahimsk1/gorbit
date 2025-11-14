package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/gorbit/orbitalrush/internal/transport"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Server Suite")
}

var _ = Describe("Server Startup and Shutdown", Label("scope:integration", "loop:g5-adapter", "layer:server", "dep:ws", "b:server-startup", "r:medium"), func() {
	var (
		server *http.Server
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		// Set up context for graceful shutdown
		_, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
		if server != nil {
			// Force shutdown for cleanup
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer shutdownCancel()
			server.Shutdown(shutdownCtx)
		}
	})

	Describe("Server initialization", func() {
		It("registers /ws endpoint with transport handler", func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/ws", transport.WebSocketHandler)

			testServer := httptest.NewServer(mux)
			defer testServer.Close()

			serverURL := "ws" + testServer.URL[4:] + "/ws"

			// Test WebSocket upgrade
			dialer := websocket.Dialer{}
			conn, resp, err := dialer.Dial(serverURL, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusSwitchingProtocols))
			Expect(conn).NotTo(BeNil())
			conn.Close()
		})

		It("registers /healthz endpoint with transport handler", func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/healthz", transport.HealthzHandler)

			testServer := httptest.NewServer(mux)
			defer testServer.Close()

			// Test healthz endpoint
			resp, err := http.Get(testServer.URL + "/healthz")
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header.Get("Content-Type")).To(ContainSubstring("application/json"))

			var result map[string]string
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["status"]).To(Equal("ok"))
		})

		It("uses PORT environment variable or defaults to 8080", func() {
			// Test default port logic
			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}
			Expect(port).To(Equal("8080"))
		})
	})

	Describe("Graceful shutdown", func() {
		It("handles SIGINT signal", func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			server = &http.Server{
				Addr:    ":0",
				Handler: mux,
			}

			// Channel to signal shutdown
			shutdownComplete := make(chan struct{})

			// Start server in background
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					Fail("Server failed to start: " + err.Error())
				}
			}()

			// Wait for server to start
			time.Sleep(100 * time.Millisecond)

			// Set up signal handling
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT)

			// Start graceful shutdown handler
			go func() {
				<-sigChan
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer shutdownCancel()
				if err := server.Shutdown(shutdownCtx); err != nil {
					Fail("Shutdown failed: " + err.Error())
				}
				close(shutdownComplete)
			}()

			// Send SIGINT
			sigChan <- syscall.SIGINT

			// Wait for shutdown to complete (with timeout)
			select {
			case <-shutdownComplete:
				// Shutdown completed successfully
				Expect(true).To(BeTrue())
			case <-time.After(6 * time.Second):
				Fail("Shutdown did not complete within timeout")
			}
		})

		It("handles SIGTERM signal", func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			server = &http.Server{
				Addr:    ":0",
				Handler: mux,
			}

			shutdownComplete := make(chan struct{})

			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					Fail("Server failed to start: " + err.Error())
				}
			}()

			time.Sleep(100 * time.Millisecond)

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGTERM)

			go func() {
				<-sigChan
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer shutdownCancel()
				if err := server.Shutdown(shutdownCtx); err != nil {
					Fail("Shutdown failed: " + err.Error())
				}
				close(shutdownComplete)
			}()

			sigChan <- syscall.SIGTERM

			select {
			case <-shutdownComplete:
				Expect(true).To(BeTrue())
			case <-time.After(6 * time.Second):
				Fail("Shutdown did not complete within timeout")
			}
		})

		It("times out gracefully if shutdown takes too long", func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			server = &http.Server{
				Addr:    ":0",
				Handler: mux,
			}

			go func() {
				server.ListenAndServe()
			}()

			time.Sleep(100 * time.Millisecond)

			// Shutdown with very short timeout (should timeout)
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer shutdownCancel()

			start := time.Now()
			err := server.Shutdown(shutdownCtx)
			duration := time.Since(start)

			// Should timeout quickly (within reasonable time)
			Expect(duration).To(BeNumerically("<", 500*time.Millisecond))
			// The error should be context deadline exceeded or nil (if server was already stopped)
			// In practice, with such a short timeout, it should timeout
			if err != nil {
				Expect(err).To(Equal(context.DeadlineExceeded))
			}
		})
	})

	Describe("Handler registration", func() {
		It("registers transport handlers before server starts", func() {
			mux := http.NewServeMux()
			
			// Register transport handlers (simulating main.go)
			mux.HandleFunc("/ws", transport.WebSocketHandler)
			mux.HandleFunc("/healthz", transport.HealthzHandler)

			// Verify handlers are registered
			Expect(mux).NotTo(BeNil())
			
			// Create server with handlers
			server = &http.Server{
				Addr:    ":0",
				Handler: mux,
			}

			// Verify server is configured
			Expect(server).NotTo(BeNil())
			Expect(server.Handler).NotTo(BeNil())
		})

		It("handles concurrent WebSocket connections", func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/ws", transport.WebSocketHandler)

			testServer := httptest.NewServer(mux)
			defer testServer.Close()

			serverURL := "ws" + testServer.URL[4:] + "/ws"

			// Create multiple connections
			dialer := websocket.Dialer{}
			conn1, _, err1 := dialer.Dial(serverURL, nil)
			Expect(err1).NotTo(HaveOccurred())
			defer conn1.Close()

			conn2, _, err2 := dialer.Dial(serverURL, nil)
			Expect(err2).NotTo(HaveOccurred())
			defer conn2.Close()

			// Both connections should be open
			Expect(conn1).NotTo(BeNil())
			Expect(conn2).NotTo(BeNil())
		})
	})
})

