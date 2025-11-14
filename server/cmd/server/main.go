package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorbit/orbitalrush/internal/transport"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP mux and register handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", transport.WebSocketHandler)
	mux.HandleFunc("/healthz", transport.HealthzHandler)

	// Create HTTP server
	addr := fmt.Sprintf(":%s", port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		fmt.Printf("Orbital Rush Server starting on %s\n", addr)
		fmt.Println("WebSocket endpoint available at /ws")
		fmt.Println("Health check available at /healthz")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
