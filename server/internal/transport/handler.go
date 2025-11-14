package transport

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorbit/orbitalrush/internal/session"
)

// WebSocketHandler handles WebSocket upgrade requests at the /ws endpoint.
// It upgrades the HTTP connection to WebSocket, creates a session handler,
// and manages the connection lifecycle.
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := UpgradeConnection(w, r)
	if err != nil {
		// UpgradeConnection may have already written error headers
		// Log the error but don't write additional response
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create Connection wrapper
	wsConn := NewConnection(conn)
	defer func() {
		if err := wsConn.Close(); err != nil {
			log.Printf("Error closing WebSocket connection: %v", err)
		}
	}()

	// Create session handler with real clock and initial world
	clock := session.NewRealClock()
	initialWorld := NewInitialWorld()
	sessionHandler := NewSessionHandler(wsConn, clock, initialWorld)

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
			break
		}

		// Route message to session handler
		err = RouteMessage(data, sessionHandler, sessionHandler)
		if err != nil {
			// Send error response to client
			errorMsg := NewErrorMessage(err)
			if writeErr := wsConn.WriteMessage(errorMsg); writeErr != nil {
				// Failed to write error, connection likely closed
				break
			}
		}
	}
}

// HealthzHandler handles health check requests at the /healthz endpoint.
// Returns a JSON response with status "ok".
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"status": "ok",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding healthz response: %v", err)
	}
}

