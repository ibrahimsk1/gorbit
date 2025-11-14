package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorbit/orbitalrush/internal/observability"
	"github.com/gorbit/orbitalrush/internal/transport"
)

func main() {
	logger := observability.NewLogger().WithValues("component", "server")
	
	// Initialize Prometheus metrics
	observability.InitMetrics()
	logger.Info("Metrics initialized", "metrics_endpoint", "/metrics")
	
	// Start GC monitor with 5 second interval
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gcMonitorInterval := 5 * time.Second
	stopGCMonitor := observability.StartGCMonitor(ctx, gcMonitorInterval, logger)
	logger.Info("GC monitor started", "interval_seconds", gcMonitorInterval.Seconds())
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP mux and register handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", transport.WebSocketHandler)
	mux.HandleFunc("/healthz", transport.HealthzHandler)
	mux.HandleFunc("/metrics", observability.MetricsHandler)

	// Create HTTP server
	addr := fmt.Sprintf(":%s", port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server starting", "address", addr, "ws_endpoint", "/ws", "health_endpoint", "/healthz", "metrics_endpoint", "/metrics")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err, "Server failed to start", "address", addr)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Stop GC monitor
	close(stopGCMonitor)
	cancel() // Cancel context to stop GC monitor goroutine
	logger.Info("GC monitor stopped")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error(err, "Server forced to shutdown")
		os.Exit(1)
	}

	logger.Info("Server exited")
}
