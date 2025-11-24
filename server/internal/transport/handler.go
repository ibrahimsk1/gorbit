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
	"github.com/gorbit/orbitalrush/internal/sim/entities"
	"github.com/gorbit/orbitalrush/internal/sim/rules"
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
	CreateRoomFunc         func() (string, error)
	JoinRoomFunc          func(roomCode string, conn *Connection) (RoomData, uint32, error)
	LeaveRoomFunc         func(roomCode string, playerID uint32) error
	GetRoomFunc           func(roomCode string) (RoomData, error)
	StartMatchFunc        func(roomCode string, hostPlayerID uint32, clock session.Clock) error // May be nil if not implemented
	EnqueueCommandToRoomFunc func(roomCode string, playerID uint32, seq uint32, cmd rules.InputCommand) error
	GetWorldFromRoomFunc  func(roomCode string) (entities.World, error)
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
	broadcaster *SnapshotBroadcaster // Optional: will be wired in step 8
}

// NewRoomHandler creates a new RoomHandler.
func NewRoomHandler(registry *ConnectionRegistry, ops RoomOperations, conn *Connection, clock session.Clock) *RoomHandler {
	return &RoomHandler{
		registry: registry,
		ops:      ops,
		conn:     conn,
		clock:    clock,
		broadcaster: nil, // Will be set in step 8
	}
}

// SetBroadcaster sets the snapshot broadcaster for the room handler.
// This will be called in step 8 when wiring up the room-based flow.
func (h *RoomHandler) SetBroadcaster(broadcaster *SnapshotBroadcaster) {
	h.broadcaster = broadcaster
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

	// Start snapshot broadcasting for this room
	if h.broadcaster != nil {
		h.broadcaster.StartBroadcasting(roomCode)
	}

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

// RoomInputHandler handles InputMessage messages by routing them to the room's session.
type RoomInputHandler struct {
	registry *ConnectionRegistry
	ops      RoomOperations
	conn     *Connection
}

// NewRoomInputHandler creates a new RoomInputHandler.
func NewRoomInputHandler(registry *ConnectionRegistry, ops RoomOperations, conn *Connection) *RoomInputHandler {
	return &RoomInputHandler{
		registry: registry,
		ops:      ops,
		conn:     conn,
	}
}

// HandleInput handles InputMessage by routing it to the room's session.
func (h *RoomInputHandler) HandleInput(msg *proto.InputMessage) error {
	// Look up connection's room code and player ID from registry
	roomCode, playerID, err := h.registry.GetRoomInfo(h.conn)
	if err != nil {
		return fmt.Errorf("connection not associated with any room: %w", err)
	}

	// Convert InputMessage to InputCommand
	cmd := rules.InputCommand{
		Thrust: msg.Thrust,
		Turn:   msg.Turn,
	}

	// Call RoomOperations.EnqueueCommandToRoomFunc
	if h.ops.EnqueueCommandToRoomFunc == nil {
		return fmt.Errorf("EnqueueCommandToRoomFunc not provided")
	}

	err = h.ops.EnqueueCommandToRoomFunc(roomCode, playerID, msg.Seq, cmd)
	if err != nil {
		return fmt.Errorf("failed to enqueue command to room: %w", err)
	}

	return nil
}

// SnapshotBroadcaster manages per-room snapshot broadcasting.
// It creates a goroutine per room that polls session world state and broadcasts to all players.
type SnapshotBroadcaster struct {
	ops      RoomOperations
	mu       sync.RWMutex
	rooms    map[string]*broadcastRoom // roomCode -> broadcastRoom
	stopChan chan string                // Channel to signal room to stop broadcasting
}

type broadcastRoom struct {
	roomCode string
	done     chan struct{}
}

// NewSnapshotBroadcaster creates a new SnapshotBroadcaster.
func NewSnapshotBroadcaster(ops RoomOperations) *SnapshotBroadcaster {
	return &SnapshotBroadcaster{
		ops:      ops,
		rooms:    make(map[string]*broadcastRoom),
		stopChan: make(chan string, 10), // Buffered channel for stop signals
	}
}

// StartBroadcasting starts snapshot broadcasting for a room.
// Creates a goroutine that polls session world state at 10 Hz (100ms interval)
// and broadcasts snapshots to all players in the room.
// Only broadcasts when room is in "playing" state.
func (sb *SnapshotBroadcaster) StartBroadcasting(roomCode string) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	// Check if already broadcasting for this room
	if _, exists := sb.rooms[roomCode]; exists {
		return // Already broadcasting
	}

	// Create broadcast room
	br := &broadcastRoom{
		roomCode: roomCode,
		done:     make(chan struct{}),
	}
	sb.rooms[roomCode] = br

	// Start broadcasting goroutine
	go sb.broadcastLoop(roomCode, br.done)
}

// StopBroadcasting stops snapshot broadcasting for a room.
func (sb *SnapshotBroadcaster) StopBroadcasting(roomCode string) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	br, exists := sb.rooms[roomCode]
	if !exists {
		return // Not broadcasting
	}

	// Signal stop
	close(br.done)
	delete(sb.rooms, roomCode)
}

// broadcastLoop is the main broadcasting loop for a room.
// Polls session world state at 10 Hz (100ms interval) and broadcasts to all players.
func (sb *SnapshotBroadcaster) broadcastLoop(roomCode string, done chan struct{}) {
	ticker := time.NewTicker(100 * time.Millisecond) // 10 Hz
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// Get room data to check state and get players
			roomData, err := sb.ops.GetRoomFunc(roomCode)
			if err != nil {
				// Room not found, stop broadcasting
				sb.StopBroadcasting(roomCode)
				return
			}

			// Only broadcast when room is in "playing" state
			if roomData.State != "playing" {
				// If room ended or in lobby, stop broadcasting
				if roomData.State == "ended" {
					sb.StopBroadcasting(roomCode)
					return
				}
				// If in lobby, skip this tick but continue
				continue
			}

			// Get world state from room's session
			if sb.ops.GetWorldFromRoomFunc == nil {
				continue // Skip if not wired yet
			}

			world, err := sb.ops.GetWorldFromRoomFunc(roomCode)
			if err != nil {
				// Session not found or room not found, skip this tick
				continue
			}

			// Broadcast to all players in room
			// Each player gets a snapshot with their own myShipId
			for _, player := range roomData.Players {
				if player.Conn != nil {
					// Convert world to snapshot with this player's ID as myShipId
					snapshot := WorldToSnapshot(world, player.PlayerID)

					// Serialize snapshot
					data, err := json.Marshal(snapshot)
					if err != nil {
						// Log error but continue
						continue
					}

					// Send snapshot (ignore errors - connection might be closed)
					_ = player.Conn.WriteMessage(data)
				}
			}
		}
	}
}

// Global connection registry and snapshot broadcaster (initialized on first use)
var (
	globalRegistry   *ConnectionRegistry
	globalBroadcaster *SnapshotBroadcaster
	initOnce         sync.Once
)

// getGlobalRegistry returns the global connection registry (initialized on first call).
func getGlobalRegistry() *ConnectionRegistry {
	initOnce.Do(func() {
		globalRegistry = NewConnectionRegistry()
	})
	return globalRegistry
}

// getGlobalBroadcaster returns the global snapshot broadcaster (initialized on first call).
func getGlobalBroadcaster() *SnapshotBroadcaster {
	initOnce.Do(func() {
		// Create RoomOperations with adapter functions
		// NOTE: SetRoomOperationsAdapter must be called from main.go to wire RoomManager
		roomOps := createRoomOperations()
		globalBroadcaster = NewSnapshotBroadcaster(roomOps)
	})
	return globalBroadcaster
}

// createRoomOperations creates RoomOperations with adapter functions.
// NOTE: The actual adapter functions that call RoomManager will be set by SetRoomOperationsAdapter
// to avoid circular dependency (room package imports transport).
var roomOperationsAdapter func() RoomOperations

// SetRoomOperationsAdapter sets the function that creates RoomOperations with adapter functions.
// This should be called from main.go or an adapter package to wire RoomManager to RoomOperations.
func SetRoomOperationsAdapter(adapter func() RoomOperations) {
	roomOperationsAdapter = adapter
}

// createRoomOperations creates RoomOperations using the adapter function if set.
func createRoomOperations() RoomOperations {
	if roomOperationsAdapter != nil {
		return roomOperationsAdapter()
	}
	// Return empty operations if adapter not set (for testing)
	return RoomOperations{}
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
	
	// Get global connection registry and create room operations
	registry := getGlobalRegistry()
	roomOps := createRoomOperations()
	broadcaster := getGlobalBroadcaster()

	// Create room handler and input handler with wired operations
	clock := session.NewRealClock()
	roomHandler := NewRoomHandler(registry, roomOps, wsConn, clock)
	roomHandler.SetBroadcaster(broadcaster)
	roomInputHandler := NewRoomInputHandler(registry, roomOps, wsConn)

	connLogger.Info("WebSocket connection established", "message_type", "connect", "remote_addr", r.RemoteAddr)

	// Handle incoming messages in a loop
	for {
		// Read message from WebSocket
		data, err := wsConn.ReadMessage()
		if err != nil {
			// Connection closed or error reading
			// This is normal when client disconnects
			// Handle disconnection by calling LeaveRoom if in room
			if registry.IsAssociated(wsConn) {
				roomCode, playerID, err := registry.GetRoomInfo(wsConn)
				if err == nil && roomOps.LeaveRoomFunc != nil {
					// Call LeaveRoom to handle cleanup
					// This will remove player from room, close connection, and clean up if room becomes empty
					_ = roomOps.LeaveRoomFunc(roomCode, playerID)
				}
			}
			// Cleanup connection tracking
			registry.Disassociate(wsConn)
			// Note: defer will handle disconnect metrics and logging
			break
		}

		// Route message: input messages go to room input handler, room management messages go to room handler
		err = RouteMessage(data, roomInputHandler, roomHandler, roomHandler, roomHandler, roomHandler)
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

