package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorbit/orbitalrush/internal/observability"
	"github.com/gorbit/orbitalrush/internal/proto"
	"github.com/gorbit/orbitalrush/internal/session"
	"github.com/gorbit/orbitalrush/internal/sim/entities"
	"github.com/gorbit/orbitalrush/internal/sim/rules"
	"github.com/gorilla/websocket"
)

const (
	// ReadDeadline is the read deadline for WebSocket connections (60 seconds)
	ReadDeadline = 60 * time.Second
	// WriteDeadline is the write deadline for WebSocket connections (10 seconds)
	WriteDeadline = 10 * time.Second
	// PongWait is the time to wait for pong response (must be less than ReadDeadline)
	PongWait = 60 * time.Second
	// PingPeriod is how often to send ping messages (must be less than PongWait)
	PingPeriod = (PongWait * 9) / 10
)

var (
	// upgrader is the WebSocket upgrader used for HTTP to WebSocket upgrades
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// For now, allow all origins. In production, this should validate
			// the origin against a whitelist.
			return true
		},
	}
)

// Connection manages a WebSocket connection lifecycle.
// It provides methods for reading and writing messages, and graceful closure.
type Connection struct {
	conn      *websocket.Conn
	done      chan struct{}
	writeChan chan []byte
	startTime time.Time
}

// NewConnection creates a new Connection wrapper around a WebSocket connection.
func NewConnection(conn *websocket.Conn) *Connection {
	c := &Connection{
		conn:      conn,
		done:      make(chan struct{}),
		writeChan: make(chan []byte, 256),
		startTime: time.Now(),
	}

	// Set read deadline and pong handler
	conn.SetReadDeadline(time.Now().Add(PongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	// Start write pump (handles all writes including pings)
	go c.writePump()

	return c
}

// GetStartTime returns the connection start time.
func (c *Connection) GetStartTime() time.Time {
	return c.startTime
}

// UpgradeConnection upgrades an HTTP connection to a WebSocket connection.
// Returns the WebSocket connection or an error if the upgrade fails.
func UpgradeConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// ReadMessage reads a JSON text message from the WebSocket connection.
// Returns the message bytes or an error if the read fails.
func (c *Connection) ReadMessage() ([]byte, error) {
	messageType, data, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	// Only accept text messages (JSON)
	if messageType != websocket.TextMessage {
		return nil, websocket.ErrBadHandshake
	}

	// Record bytes in and message count
	if len(data) > 0 {
		if bytesCounter := observability.GetConnectionBytesCounter(); bytesCounter != nil {
			bytesCounter.WithLabelValues("in").Add(float64(len(data)))
		}
		if msgCounter := observability.GetMessagesCounter(); msgCounter != nil {
			msgCounter.WithLabelValues("in").Inc()
		}
	}

	return data, nil
}

// WriteMessage enqueues a JSON text message to be written to the WebSocket connection.
// Returns an error if the connection is closed or the message cannot be enqueued.
func (c *Connection) WriteMessage(data []byte) error {
	select {
	case <-c.done:
		return fmt.Errorf("connection closed")
	case c.writeChan <- data:
		return nil
	}
}

// Close gracefully closes the WebSocket connection.
// It can be called multiple times safely.
// Closing c.done signals writePump to exit, then the underlying connection is closed.
func (c *Connection) Close() error {
	select {
	case <-c.done:
		// Already closed
		return nil
	default:
		close(c.done)
		// Close writeChan to signal writePump to exit
		// This is safe because writePump will see c.done is closed and exit,
		// and WriteMessage checks c.done before sending, so no new sends will occur.
		close(c.writeChan)
		return c.conn.Close()
	}
}

// writePump handles all writes to the WebSocket connection.
// It processes messages from writeChan and sends periodic ping messages.
// This ensures only one goroutine writes to the connection, preventing concurrent write panics.
// Messages are prioritized over pings, and pending messages are batched for efficiency.
func (c *Connection) writePump() {
	pingTicker := time.NewTicker(PingPeriod)
	defer pingTicker.Stop()

	for {
		select {
		case <-c.done:
			return

		case data, ok := <-c.writeChan:
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.writeMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-pingTicker.C:
			// Before sending a ping, check if there is a message ready.
			select {
			case data, ok := <-c.writeChan:
				if !ok {
					_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				if err := c.writeMessage(websocket.TextMessage, data); err != nil {
					return
				}
			default:
				// Truly idle: safe to ping
				if err := c.writeMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}

		// Drain pending messages after any write for efficiency
	drain:
		for {
			select {
			case <-c.done:
				return
			case data, ok := <-c.writeChan:
				if !ok {
					_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				if err := c.writeMessage(websocket.TextMessage, data); err != nil {
					return
				}
			default:
				break drain
			}
		}
	}
}

// writeMessage writes a message to the WebSocket connection and records metrics.
func (c *Connection) writeMessage(messageType int, data []byte) error {
	c.conn.SetWriteDeadline(time.Now().Add(WriteDeadline))
	if err := c.conn.WriteMessage(messageType, data); err != nil {
		return err
	}

	if messageType == websocket.TextMessage && len(data) > 0 {
		c.recordMetrics(data)
	}

	return nil
}

// recordMetrics records bytes and message count metrics for outgoing messages.
func (c *Connection) recordMetrics(data []byte) {
	if len(data) > 0 {
		if bytesCounter := observability.GetConnectionBytesCounter(); bytesCounter != nil {
			bytesCounter.WithLabelValues("out").Add(float64(len(data)))
		}
		if msgCounter := observability.GetMessagesCounter(); msgCounter != nil {
			msgCounter.WithLabelValues("out").Inc()
		}
	}
}

// InputMessageHandler handles InputMessage messages.
type InputMessageHandler interface {
	HandleInput(msg *proto.InputMessage) error
}

// RestartMessageHandler handles RestartMessage messages.
type RestartMessageHandler interface {
	HandleRestart(msg *proto.RestartMessage) error
}

// ParseMessage parses a JSON message and returns a typed message (InputMessage or RestartMessage).
// Returns an error if the message is malformed, invalid, or of unknown type.
func ParseMessage(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty message")
	}

	// First, parse into a generic map to determine message type
	var msgType map[string]interface{}
	if err := json.Unmarshal(data, &msgType); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check if "t" field exists
	typeField, ok := msgType["t"]
	if !ok {
		return nil, fmt.Errorf("missing message type field 't'")
	}

	typeStr, ok := typeField.(string)
	if !ok {
		return nil, fmt.Errorf("message type field 't' must be a string")
	}

	// Route to appropriate message type
	switch typeStr {
	case "input":
		var msg proto.InputMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse InputMessage: %w", err)
		}
		if err := proto.ValidateInputMessage(&msg); err != nil {
			return nil, fmt.Errorf("invalid InputMessage: %w", err)
		}
		return &msg, nil

	case "restart":
		var msg proto.RestartMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse RestartMessage: %w", err)
		}
		if err := proto.ValidateRestartMessage(&msg); err != nil {
			return nil, fmt.Errorf("invalid RestartMessage: %w", err)
		}
		return &msg, nil

	default:
		return nil, fmt.Errorf("unknown message type: %s", typeStr)
	}
}

// RouteMessage parses a JSON message, validates it, and routes it to the appropriate handler.
// Returns an error if parsing, validation, or handler execution fails.
func RouteMessage(data []byte, inputHandler InputMessageHandler, restartHandler RestartMessageHandler) error {
	msg, err := ParseMessage(data)
	if err != nil {
		return err
	}

	// Route to appropriate handler
	switch m := msg.(type) {
	case *proto.InputMessage:
		if inputHandler == nil {
			return fmt.Errorf("InputMessageHandler is nil")
		}
		return inputHandler.HandleInput(m)

	case *proto.RestartMessage:
		if restartHandler == nil {
			return fmt.Errorf("RestartMessageHandler is nil")
		}
		return restartHandler.HandleRestart(m)

	default:
		return fmt.Errorf("unexpected message type: %T", msg)
	}
}

// ErrorMessage represents an error response message.
type ErrorMessage struct {
	Type    string `json:"t"`
	Message string `json:"message"`
}

// NewErrorMessage creates a JSON error response message.
func NewErrorMessage(err error) []byte {
	errorMsg := ErrorMessage{
		Type:    "error",
		Message: err.Error(),
	}
	data, _ := json.Marshal(errorMsg)
	return data
}

// NewInitialWorld creates a default initial world state for new sessions.
// Ship at position (10, 0) with zero velocity, 100 energy.
// Sun at origin (0, 0) with radius 50, mass 1000.
// Empty pallets array.
func NewInitialWorld() entities.World {
	ship := entities.NewShip(
		entities.NewVec2(10.0, 0.0),
		entities.NewVec2(0.0, 0.0),
		0.0,
		100.0,
	)
	sun := entities.NewSun(
		entities.NewVec2(0.0, 0.0),
		50.0,
		1000.0,
	)
	return entities.NewWorld(ship, sun, nil)
}

// SessionHandler manages a session for a WebSocket connection.
// It implements InputMessageHandler and RestartMessageHandler interfaces.
type SessionHandler struct {
	session        *session.Session
	conn           *Connection
	clock          session.Clock
	initialWorld   entities.World
	done           chan struct{}
	snapshotTicker *time.Ticker
}

// NewSessionHandler creates a new SessionHandler with a new session.
// The logger parameter is optional. If provided and enabled, it will be injected into the session for tick time logging.
func NewSessionHandler(conn *Connection, clock session.Clock, initialWorld entities.World, logger logr.Logger) *SessionHandler {
	sess := session.NewSession(clock, initialWorld, 100) // maxQueueSize = 100
	// Set logger if it's enabled (zero logger will return false)
	if logger.Enabled() {
		sess.SetLogger(logger)
	}
	snapshotInterval := 100 * time.Millisecond // 10 Hz (100ms = 10 snapshots per second)

	return &SessionHandler{
		session:        sess,
		conn:           conn,
		clock:          clock,
		initialWorld:   initialWorld,
		done:           make(chan struct{}),
		snapshotTicker: time.NewTicker(snapshotInterval),
	}
}

// HandleInput enqueues an input command to the session.
func (h *SessionHandler) HandleInput(msg *proto.InputMessage) error {
	cmd := rules.InputCommand{
		Thrust: msg.Thrust,
		Turn:   msg.Turn,
	}

	success := h.session.EnqueueCommand(msg.Seq, cmd)
	if !success {
		return fmt.Errorf("failed to enqueue command with seq %d", msg.Seq)
	}

	return nil
}

// HandleRestart resets the session to the initial world state.
func (h *SessionHandler) HandleRestart(msg *proto.RestartMessage) error {
	// Stop current session
	h.session.Stop()

	// Create new session with initial world state
	h.session = session.NewSession(h.clock, h.initialWorld, 100)

	return nil
}

// Start starts the session run loop and snapshot broadcasting.
func (h *SessionHandler) Start() {
	// Start session run loop (30Hz = ~33ms per tick)
	sessionTicker := time.NewTicker(33 * time.Millisecond)
	go func() {
		defer sessionTicker.Stop()
		for {
			select {
			case <-h.done:
				return
			case <-sessionTicker.C:
				// Run session to process ticks (limit to 10 ticks per call to prevent lag)
				h.session.Run(10)
			}
		}
	}()

	// Start snapshot broadcasting loop (~10 Hz = 100ms per snapshot)
	go func() {
		for {
			select {
			case <-h.done:
				return
			case <-h.snapshotTicker.C:
				// Get world state and broadcast snapshot
				world := h.session.GetWorld()
				snapshot := WorldToSnapshot(world)

				// Serialize and send snapshot
				data, err := json.Marshal(snapshot)
				if err != nil {
					// Log error but continue
					continue
				}

				// Write snapshot (ignore errors - connection may be closed)
				_ = h.conn.WriteMessage(data)
			}
		}
	}()
}

// Stop stops the session handler and cleans up resources.
func (h *SessionHandler) Stop() {
	close(h.done)
	h.snapshotTicker.Stop()
	h.session.Stop()
}
