package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorbit/orbitalrush/internal/observability"
	"github.com/gorbit/orbitalrush/internal/session"
)

// WebSocketHandler handles WebSocket upgrade requests at the /ws endpoint.
// It upgrades the HTTP connection to WebSocket, creates a session handler,
// and manages the connection lifecycle.
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	logger := observability.NewLogger().WithValues("component", "transport", "handler", "websocket")
	
	// Generate a simple connection ID from remote address and timestamp
	connectionID := fmt.Sprintf("%s-%d", r.RemoteAddr, time.Now().UnixNano())
	connLogger := logger.WithValues("connection_id", connectionID)

	// Upgrade HTTP connection to WebSocket
	conn, err := UpgradeConnection(w, r)
	if err != nil {
		// UpgradeConnection may have already written error headers
		// Log the error and record error event
		connLogger.Error(err, "WebSocket upgrade failed", "message_type", "upgrade_error")
		if eventsCounter := observability.GetConnectionEventsCounter(); eventsCounter != nil {
			eventsCounter.WithLabelValues("error").Inc()
		}
		return
	}

	// Create Connection wrapper
	wsConn := NewConnection(conn)
	connectionStartTime := wsConn.GetStartTime()
	
	defer func() {
		// Calculate connection duration
		duration := time.Since(connectionStartTime).Seconds()
		
		// Record disconnect event and duration
		if eventsCounter := observability.GetConnectionEventsCounter(); eventsCounter != nil {
			eventsCounter.WithLabelValues("disconnect").Inc()
		}
		if activeGauge := observability.GetActiveConnectionsGauge(); activeGauge != nil {
			activeGauge.Dec()
		}
		if durationHist := observability.GetConnectionDurationHistogram(); durationHist != nil {
			durationHist.Observe(duration)
		}
		
		// Log disconnect with duration
		connLogger.Info("WebSocket connection closed", "message_type", "disconnect", "duration_seconds", duration)
		
		if err := wsConn.Close(); err != nil {
			connLogger.Error(err, "Error closing WebSocket connection", "message_type", "close_error")
		}
	}()

	// Record connect event and increment active connections
	if eventsCounter := observability.GetConnectionEventsCounter(); eventsCounter != nil {
		eventsCounter.WithLabelValues("connect").Inc()
	}
	if activeGauge := observability.GetActiveConnectionsGauge(); activeGauge != nil {
		activeGauge.Inc()
	}
	
	// Create session handler with real clock and initial world
	clock := session.NewRealClock()
	initialWorld := NewInitialWorld()
	// Create session logger with connection context
	sessionLogger := connLogger.WithValues("component", "session")
	sessionHandler := NewSessionHandler(wsConn, clock, initialWorld, sessionLogger)

	connLogger.Info("WebSocket connection established", "message_type", "connect", "remote_addr", r.RemoteAddr)

	// Start session handler (runs session loop and snapshot broadcasting)
	sessionHandler.Start()
	defer sessionHandler.Stop()

	// Handle incoming messages in a loop
	for {
		// Read message from WebSocket
		data, err := wsConn.ReadMessage()
		if err != nil {
			// Connection closed or error reading
			// This is normal when client disconnects
			// Note: defer will handle disconnect metrics and logging
			break
		}

		// Route message to session handler
		err = RouteMessage(data, sessionHandler, sessionHandler)
		if err != nil {
			// Record error event
			if eventsCounter := observability.GetConnectionEventsCounter(); eventsCounter != nil {
				eventsCounter.WithLabelValues("error").Inc()
			}
			// Send error response to client
			connLogger.Error(err, "Failed to route message", "message_type", "route_error")
			errorMsg := NewErrorMessage(err)
			if writeErr := wsConn.WriteMessage(errorMsg); writeErr != nil {
				// Failed to write error, connection likely closed
				if eventsCounter := observability.GetConnectionEventsCounter(); eventsCounter != nil {
					eventsCounter.WithLabelValues("error").Inc()
				}
				connLogger.Error(writeErr, "Failed to write error message", "message_type", "write_error")
				break
			}
		}
	}
}

// HealthzHandler handles health check requests at the /healthz endpoint.
// Returns a JSON response with status and observability metrics summary.
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	logger := observability.NewLogger().WithValues("component", "transport", "handler", "healthz")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Get health metrics summary
	healthMetrics := observability.GetHealthMetrics()

	// Build response with metrics
	response := map[string]interface{}{
		"status":        "ok",
		"uptime_seconds": healthMetrics.UptimeSeconds,
		"metrics": map[string]interface{}{
			"active_connections": healthMetrics.ActiveConnections,
			"queue_depth":        healthMetrics.QueueDepth,
			"tick_time": map[string]interface{}{
				"average_ms": healthMetrics.TickTime.AverageMs,
				"count":      healthMetrics.TickTime.Count,
			},
			"gc_pause": map[string]interface{}{
				"average_ms": healthMetrics.GCPause.AverageMs,
				"count":      healthMetrics.GCPause.Count,
			},
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(err, "Error encoding healthz response", "message_type", "encode_error")
	}
}

