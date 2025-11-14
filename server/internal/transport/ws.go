package transport

import (
	"net/http"
	"time"

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
	conn     *websocket.Conn
	done     chan struct{}
	writeChan chan []byte
}

// NewConnection creates a new Connection wrapper around a WebSocket connection.
func NewConnection(conn *websocket.Conn) *Connection {
	c := &Connection{
		conn:      conn,
		done:      make(chan struct{}),
		writeChan: make(chan []byte, 256),
	}

	// Set read deadline and pong handler
	conn.SetReadDeadline(time.Now().Add(PongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	// Start ping ticker
	go c.pingTicker()

	return c
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

	return data, nil
}

// WriteMessage writes a JSON text message to the WebSocket connection.
// Returns an error if the write fails.
func (c *Connection) WriteMessage(data []byte) error {
	c.conn.SetWriteDeadline(time.Now().Add(WriteDeadline))
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// Close gracefully closes the WebSocket connection.
// It can be called multiple times safely.
func (c *Connection) Close() error {
	select {
	case <-c.done:
		// Already closed
		return nil
	default:
		close(c.done)
		return c.conn.Close()
	}
}

// pingTicker sends ping messages periodically to keep the connection alive.
func (c *Connection) pingTicker() {
	ticker := time.NewTicker(PingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(WriteDeadline))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}

