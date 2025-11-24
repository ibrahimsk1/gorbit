package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorbit/orbitalrush/internal/observability"
	"github.com/gorbit/orbitalrush/internal/proto"
	"github.com/gorbit/orbitalrush/internal/session"
)

// connectionInfo stores room association information for a connection.
type connectionInfo struct {
	RoomCode string
	PlayerID uint32
}

// ConnectionRegistry tracks connections and their room associations.
// It provides thread-safe operations to associate/disassociate connections with rooms.
type ConnectionRegistry struct {
	connections map[*Connection]*connectionInfo
	mu          sync.RWMutex
}

// NewConnectionRegistry creates a new ConnectionRegistry.
func NewConnectionRegistry() *ConnectionRegistry {
	return &ConnectionRegistry{
		connections: make(map[*Connection]*connectionInfo),
	}
}

// Associate registers a connection with a room code and player ID.
// If the connection is already associated, it updates the association.
func (cr *ConnectionRegistry) Associate(conn *Connection, roomCode string, playerID uint32) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.connections[conn] = &connectionInfo{
		RoomCode: roomCode,
		PlayerID: playerID,
	}
}

// Disassociate removes a connection from the registry.
func (cr *ConnectionRegistry) Disassociate(conn *Connection) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	delete(cr.connections, conn)
}

// GetRoomInfo returns the room code and player ID for a connection.
// Returns an error if the connection is not associated with any room.
func (cr *ConnectionRegistry) GetRoomInfo(conn *Connection) (string, uint32, error) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	info, exists := cr.connections[conn]
	if !exists {
		return "", 0, fmt.Errorf("connection not associated with any room")
	}

	return info.RoomCode, info.PlayerID, nil
}

// IsAssociated returns true if the connection is associated with a room.
func (cr *ConnectionRegistry) IsAssociated(conn *Connection) bool {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	_, exists := cr.connections[conn]
	return exists
}

// RoomOperations defines callback functions for room operations.
// This avoids circular dependencies by using function callbacks instead of importing room package.
type RoomOperations struct {
	CreateRoomFunc func() (string, error)
	JoinRoomFunc   func(roomCode string, conn *Connection) (RoomData, uint32, error)
	LeaveRoomFunc  func(roomCode string, playerID uint32) error
	GetRoomFunc    func(roomCode string) (RoomData, error)
	StartMatchFunc func(roomCode string, hostPlayerID uint32, clock session.Clock) error // May be nil if not implemented
}

// RoomData represents room data needed by transport layer.
type RoomData struct {
	RoomCode     string
	Players      []PlayerData
	State        string // "lobby", "playing", or "ended"
	HostPlayerID uint32
}

// PlayerData represents player data needed by transport layer.
type PlayerData struct {
	PlayerID uint32
	Name     string
	Conn     *Connection
}

// RoomHandler handles room management messages and implements all room management handler interfaces.
type RoomHandler struct {
	registry  *ConnectionRegistry
	ops       RoomOperations
	conn      *Connection
	clock     session.Clock
}

// NewRoomHandler creates a new RoomHandler.
func NewRoomHandler(registry *ConnectionRegistry, ops RoomOperations, conn *Connection, clock session.Clock) *RoomHandler {
	return &RoomHandler{
		registry: registry,
		ops:      ops,
		conn:     conn,
		clock:    clock,
	}
}

// HandleCreateRoom handles CreateRoomMessage by creating a room and sending roomCreated response.
func (h *RoomHandler) HandleCreateRoom(msg *proto.CreateRoomMessage) error {
	if h.ops.CreateRoomFunc == nil {
		return fmt.Errorf("CreateRoomFunc not provided")
	}

	// Create room
	roomCode, err := h.ops.CreateRoomFunc()
	if err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}

	// Send roomCreated response
	response := proto.RoomCreatedMessage{
		Type:     "roomCreated",
		RoomCode: roomCode,
	}
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal roomCreated response: %w", err)
	}

	return h.conn.WriteMessage(data)
}

// HandleJoinRoom handles JoinRoomMessage by joining a room, tracking connection, and sending responses.
func (h *RoomHandler) HandleJoinRoom(msg *proto.JoinRoomMessage) error {
	if h.ops.JoinRoomFunc == nil {
		return fmt.Errorf("JoinRoomFunc not provided")
	}

	// Join room
	roomData, playerID, err := h.ops.JoinRoomFunc(msg.RoomCode, h.conn)
	if err != nil {
		return fmt.Errorf("failed to join room: %w", err)
	}

	// Track connection
	h.registry.Associate(h.conn, msg.RoomCode, playerID)

	// Convert room to RoomStateMessage
	roomState := roomToRoomStateMessage(roomData)

	// Send roomState response to joining player
	data, err := json.Marshal(roomState)
	if err != nil {
		return fmt.Errorf("failed to marshal roomState response: %w", err)
	}
	if err := h.conn.WriteMessage(data); err != nil {
		return fmt.Errorf("failed to send roomState response: %w", err)
	}

	// Broadcast playerJoined to all other players in room
	playerJoined := proto.PlayerJoinedMessage{
		Type: "playerJoined",
		Player: proto.PlayerInfo{
			ID:   playerID,
			Name: "", // Empty name in v1
		},
	}
	broadcastData, err := json.Marshal(playerJoined)
	if err != nil {
		return fmt.Errorf("failed to marshal playerJoined message: %w", err)
	}
	broadcastToRoom(roomData, h.conn, broadcastData) // Exclude joining player

	return nil
}

// HandleLeaveRoom handles LeaveRoomMessage by leaving a room, untracking connection, and broadcasting.
func (h *RoomHandler) HandleLeaveRoom(msg *proto.LeaveRoomMessage) error {
	if h.ops.LeaveRoomFunc == nil {
		return fmt.Errorf("LeaveRoomFunc not provided")
	}

	// Get room info from registry
	roomCode, playerID, err := h.registry.GetRoomInfo(h.conn)
	if err != nil {
		return fmt.Errorf("connection not in a room: %w", err)
	}

	// Get room before leaving (to broadcast to other players)
	roomData, err := h.ops.GetRoomFunc(roomCode)
	if err != nil {
		// Room might not exist, but we still need to untrack
		h.registry.Disassociate(h.conn)
		return fmt.Errorf("room not found: %w", err)
	}

	// Leave room
	err = h.ops.LeaveRoomFunc(roomCode, playerID)
	if err != nil {
		return fmt.Errorf("failed to leave room: %w", err)
	}

	// Untrack connection
	h.registry.Disassociate(h.conn)

	// Broadcast playerLeft to all other players in room (if room still exists)
	// Note: Room might be deleted if it became empty, so we check again
	roomData, err = h.ops.GetRoomFunc(roomCode)
	if err == nil {
		playerLeft := proto.PlayerLeftMessage{
			Type:     "playerLeft",
			PlayerID: playerID,
		}
		broadcastData, err := json.Marshal(playerLeft)
		if err == nil {
			broadcastToRoom(roomData, h.conn, broadcastData) // Exclude leaving player
		}
	}

	return nil
}

// HandleStartMatch handles StartMatchMessage by starting a match and broadcasting matchStarted.
// NOTE: RoomManager.StartMatch needs to be implemented (currently missing from manager.go)
func (h *RoomHandler) HandleStartMatch(msg *proto.StartMatchMessage) error {
	if h.ops.StartMatchFunc == nil {
		return fmt.Errorf("StartMatch not yet implemented in RoomManager - prerequisite missing")
	}

	// Get room info from registry
	roomCode, playerID, err := h.registry.GetRoomInfo(h.conn)
	if err != nil {
		return fmt.Errorf("connection not in a room: %w", err)
	}

	// Start match
	err = h.ops.StartMatchFunc(roomCode, playerID, h.clock)
	if err != nil {
		return fmt.Errorf("failed to start match: %w", err)
	}

	// Broadcast matchStarted to all players in room
	matchStarted := proto.MatchStartedMessage{
		Type: "matchStarted",
	}
	broadcastData, err := json.Marshal(matchStarted)
	if err != nil {
		return fmt.Errorf("failed to marshal matchStarted message: %w", err)
	}
	roomData, err := h.ops.GetRoomFunc(roomCode)
	if err != nil {
		return fmt.Errorf("room not found after starting match: %w", err)
	}
	broadcastToRoom(roomData, nil, broadcastData) // Broadcast to all players

	return nil
}

// roomToRoomStateMessage converts RoomData to a RoomStateMessage.
func roomToRoomStateMessage(r RoomData) proto.RoomStateMessage {
	playerInfos := make([]proto.PlayerInfo, len(r.Players))
	for i, player := range r.Players {
		playerInfos[i] = proto.PlayerInfo{
			ID:   player.PlayerID,
			Name: player.Name,
		}
	}

	return proto.RoomStateMessage{
		Type:     "roomState",
		RoomCode: r.RoomCode,
		Players:  playerInfos,
		State:    r.State,
		HostID:   r.HostPlayerID,
	}
}

// broadcastToRoom sends a message to all players in a room except the excluded connection.
func broadcastToRoom(r RoomData, excludeConn *Connection, data []byte) {
	for _, player := range r.Players {
		// Skip excluded connection
		if excludeConn != nil && player.Conn == excludeConn {
			continue
		}
		// Send message (ignore errors - connection might be closed)
		_ = player.Conn.WriteMessage(data)
	}
}

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

		// Route message to session handler (room handlers will be added in step 4)
		err = RouteMessage(data, sessionHandler, nil, nil, nil, nil)
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

