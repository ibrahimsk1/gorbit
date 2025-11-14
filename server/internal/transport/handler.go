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
		// Log the error but don't write additional response
		connLogger.Error(err, "WebSocket upgrade failed", "message_type", "upgrade_error")
		return
	}

	// Create Connection wrapper
	wsConn := NewConnection(conn)
	defer func() {
		if err := wsConn.Close(); err != nil {
			connLogger.Error(err, "Error closing WebSocket connection", "message_type", "close_error")
		}
	}()

	// Create session handler with real clock and initial world
	clock := session.NewRealClock()
	initialWorld := NewInitialWorld()
	// Create session logger with connection context
	sessionLogger := connLogger.WithValues("component", "session")
	sessionHandler := NewSessionHandler(wsConn, clock, initialWorld, sessionLogger)

	connLogger.Info("WebSocket connection established", "message_type", "connect")

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
			connLogger.Info("WebSocket connection closed", "message_type", "disconnect")
			break
		}

		// Route message to session handler
		err = RouteMessage(data, sessionHandler, sessionHandler)
		if err != nil {
			// Send error response to client
			connLogger.Error(err, "Failed to route message", "message_type", "route_error")
			errorMsg := NewErrorMessage(err)
			if writeErr := wsConn.WriteMessage(errorMsg); writeErr != nil {
				// Failed to write error, connection likely closed
				connLogger.Error(writeErr, "Failed to write error message", "message_type", "write_error")
				break
			}
		}
	}
}

// HealthzHandler handles health check requests at the /healthz endpoint.
// Returns a JSON response with status "ok".
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	logger := observability.NewLogger().WithValues("component", "transport", "handler", "healthz")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"status": "ok",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(err, "Error encoding healthz response", "message_type", "encode_error")
	}
}

