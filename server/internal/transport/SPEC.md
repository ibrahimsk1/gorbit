# Orbital Rush – Transport Subsystem Specification

This document describes the transport layer for Orbital Rush. It defines WebSocket connection management, HTTP route handlers, message routing, and entity-to-protocol conversion.

---

## Scope & Location

**Scope**: Transport layer for Orbital Rush (WebSocket, HTTP handlers, message routing, conversion).

**Code location**: `server/internal/transport`

**Design Goals**:
- Manages WebSocket connection lifecycle
- Handles HTTP route endpoints
- Routes messages between protocol and session layers
- Converts between entity types and protocol types
- Provides connection management (ping/pong, deadlines, cleanup)

---

## Core Components

### Connection

**File**: `server/internal/transport/ws.go`

**Concept**: Manages WebSocket connection lifecycle with read/write pumps and graceful closure.

**Key Operations**:
- `NewConnection(conn)` – Create connection wrapper
- `ReadMessage()` – Read JSON text message from WebSocket
- `WriteMessage(data)` – Enqueue message for writing
- `Close()` – Gracefully close connection
- `GetStartTime()` – Get connection start time

**Connection Lifecycle**:
1. **Upgrade**: HTTP connection upgraded to WebSocket
2. **Setup**: Read deadline and pong handler configured
3. **Write Pump**: Goroutine handles all writes (messages + pings)
4. **Read Loop**: Main goroutine reads messages
5. **Cleanup**: On disconnect, close write pump and underlying connection

**Write Pump**:
- Handles all writes to WebSocket (single writer goroutine)
- Sends periodic ping messages (every PingPeriod)
- Processes messages from writeChan
- Batches pending messages for efficiency
- Exits when connection closed

**Read Semantics**:
- Sets read deadline (PongWait duration)
- Only accepts text messages (JSON)
- Records metrics (bytes in, message count)
- Returns error on connection close or invalid message type

**Write Semantics**:
- Messages enqueued to writeChan
- Write pump serializes all writes
- Ping messages sent periodically
- Returns error if connection closed

**Invariants**:
- Only one goroutine writes to WebSocket (write pump)
- Read deadline refreshed on pong
- Connection closed gracefully (signals done, closes channels)
- Metrics recorded for all messages

---

### HTTP Handlers

**File**: `server/internal/transport/handler.go`

#### WebSocketHandler

**Endpoint**: `GET /ws`

**Concept**: Handles WebSocket upgrade requests and manages connection lifecycle.

**Flow**:
1. Upgrade HTTP connection to WebSocket
2. Create Connection wrapper
3. Create SessionHandler with initial world
4. Start session handler (tick loop + snapshot broadcasting)
5. Read messages in loop, route to session handler
6. On disconnect, stop session handler and close connection

**Error Handling**:
- Upgrade failures: Log error, record metrics, return
- Message routing errors: Send error message to client, log error
- Connection errors: Break read loop, defer handles cleanup

**Metrics**:
- Connection events (connect, disconnect, error)
- Active connections gauge
- Connection duration histogram

**Invariants**:
- One WebSocket connection per handler call
- Session handler created per connection
- Connection cleaned up on disconnect (defer)
- Metrics recorded for all events

#### HealthzHandler

**Endpoint**: `GET /healthz`

**Concept**: Health check endpoint that returns server status and metrics summary.

**Response**:
```json
{
  "status": "ok",
  "uptime_seconds": <float64>,
  "metrics": {
    "active_connections": <int>,
    "queue_depth": <int>,
    "tick_time": {
      "average_ms": <float64>,
      "count": <int>
    },
    "gc_pause": {
      "average_ms": <float64>,
      "count": <int>
    }
  }
}
```

**Semantics**:
- Returns 200 OK with JSON response
- Includes observability metrics summary
- Used for health checks and monitoring

**Invariants**:
- Always returns 200 OK (if server is running)
- Response is valid JSON
- Metrics are current at time of request

---

### Message Routing

**File**: `server/internal/transport/ws.go`

**Concept**: Parses JSON messages and routes them to appropriate handlers.

**Function**: `RouteMessage(data, inputHandler, restartHandler)`

**Algorithm**:
1. Parse JSON to determine message type (check "t" field)
2. Validate message type is string
3. Route to handler:
   - `"input"` → InputMessageHandler.HandleInput()
   - `"restart"` → RestartMessageHandler.HandleRestart()
   - Unknown type → return error

**Semantics**:
- Messages must be valid JSON
- Message type must be present ("t" field)
- Unknown message types are rejected
- Validation happens in protocol layer

**Invariants**:
- All messages are validated before routing
- Unknown message types return error
- Handlers receive validated protocol messages

---

### Entity-to-Protocol Conversion

**File**: `server/internal/transport/convert.go`

**Concept**: Converts between entity types (simulation) and protocol types (wire format).

**Conversion Functions**:
- `Vec2ToSnapshot(v entities.Vec2) proto.Vec2Snapshot`
- `ShipToSnapshot(s entities.Ship) proto.ShipSnapshot`
- `SunToSnapshot(s entities.Sun) proto.SunSnapshot`
- `PalletToSnapshot(p entities.Pallet) proto.PalletSnapshot`
- `WorldToSnapshot(w entities.World) proto.SnapshotMessage`

**Semantics**:
- One-way conversion: entities → protocol (for snapshots)
- Protocol types mirror entities but optimized for JSON
- Conversion is pure (no side effects)
- Empty slices converted to empty arrays (not nil)
- Sun.Mass is not included in SunSnapshot (only used for simulation)

**Invariants**:
- All entity fields mapped to protocol fields (except simulation-only fields)
- Conversion is deterministic (same input → same output)
- Protocol types match schema in protocol spec
- No data loss in conversion (all relevant fields included)

**Note**: Protocol-to-entity conversion (for input) happens in session handler when converting InputMessage to InputCommand.

---

### Session Handler

**File**: `server/internal/transport/ws.go`

**Concept**: Bridges transport layer and session layer, manages session lifecycle and snapshot broadcasting.

**Key Operations**:
- `NewSessionHandler(conn, clock, initialWorld, logger)` – Create handler
- `HandleInput(msg)` – Enqueue input command to session
- `HandleRestart(msg)` – Reset session to initial world
- `Start()` – Start session run loop and snapshot broadcasting
- `Stop()` – Stop session and snapshot broadcasting

**Session Lifecycle**:
1. **Creation**: Handler created with initial world state
2. **Start**: Session.Run() called in goroutine, snapshot ticker started
3. **Running**: Session processes ticks, snapshots broadcast at 10 Hz
4. **Stop**: Session stopped, snapshot ticker stopped, goroutines cleaned up

**Snapshot Broadcasting**:
- Snapshots sent at 10 Hz (100ms interval)
- World state converted to SnapshotMessage
- Sent via Connection.WriteMessage()
- Continues until session stopped or connection closed

**Invariants**:
- One session per connection
- Session started after connection established
- Session stopped before connection closed
- Snapshots sent at fixed rate (10 Hz)

---

## Constants

**Transport Constants**:
- `ReadDeadline = 60s` – Read deadline for WebSocket connections
- `WriteDeadline = 10s` – Write deadline for WebSocket connections
- `PongWait = 60s` – Time to wait for pong response
- `PingPeriod = 54s` – How often to send ping (90% of PongWait)
- `SnapshotInterval = 100ms` – Snapshot broadcast interval (10 Hz)
- `WriteBufferSize = 1024` – WebSocket write buffer size
- `ReadBufferSize = 1024` – WebSocket read buffer size
- `WriteChanSize = 256` – Write channel buffer size

**HTTP Endpoints**:
- `/ws` – WebSocket upgrade endpoint
- `/healthz` – Health check endpoint
- `/metrics` – Prometheus metrics endpoint (in observability package)

---

## Ownership & Dependencies

### Transport Package Ownership

- **Only `server/internal/transport` may define transport layer logic**
- Transport handles all network IO (WebSocket, HTTP)
- Transport converts between protocol and entity types
- Transport manages connection lifecycle

### Dependencies

- **Imports**:
  - `proto` package (for message types and validation)
  - `session` package (for Session and Clock)
  - `entities` package (for World and entity types)
  - `rules` package (for InputCommand)
  - `observability` package (for metrics and logging)
  - `websocket` package (gorilla/websocket)
  - `http` package (standard library)
- **No dependencies on**: physics packages
- Transport is adapter layer that bridges network and session

### No Duplication Rules

- **No WebSocket handling elsewhere**: Connection management must live in `/transport`
- **No HTTP handlers elsewhere**: Route handlers must live in `/transport`
- **No entity conversion elsewhere**: Entity-to-protocol conversion must live in `/transport`
- **Transport does not implement**: Physics, rules, or session orchestration

---

## Connection Management

### WebSocket Upgrade

**Process**:
1. HTTP request arrives at `/ws`
2. Upgrader checks origin (currently allows all)
3. HTTP connection upgraded to WebSocket
4. Connection wrapper created
5. Read deadline and pong handler configured

**Upgrader Configuration**:
- ReadBufferSize: 1024 bytes
- WriteBufferSize: 1024 bytes
- CheckOrigin: Currently allows all (should be whitelist in production)

### Ping/Pong Keepalive

**Semantics**:
- Server sends ping every PingPeriod (54s)
- Client must respond with pong within PongWait (60s)
- Read deadline refreshed on pong
- Connection closed if pong not received

**Invariants**:
- Ping sent periodically by write pump
- Pong handler refreshes read deadline
- Connection closed if deadline exceeded

### Graceful Closure

**Process**:
1. Close signal sent (close done channel)
2. Write pump exits (sees done closed)
3. Write channel closed
4. Underlying WebSocket connection closed
5. Metrics updated (disconnect event, active connections)

**Invariants**:
- Close is idempotent (safe to call multiple times)
- All goroutines exit before connection closed
- Metrics updated on disconnect

---

## Error Handling

### Connection Errors

**Types**:
- Upgrade failures (invalid handshake)
- Read errors (connection closed, timeout)
- Write errors (connection closed, timeout)
- Message parsing errors (invalid JSON)

**Handling**:
- Log errors with context
- Record error metrics
- Send error messages to client (when possible)
- Clean up resources (defer)

### Message Errors

**Types**:
- Invalid JSON
- Unknown message type
- Validation failures (handled in protocol layer)
- Routing failures

**Handling**:
- Parse errors: Log and send error message
- Unknown type: Log and send error message
- Validation errors: Handled by protocol layer, logged
- Routing errors: Log and send error message

**Invariants**:
- All errors are logged
- Error metrics recorded
- Client notified of errors (when possible)
- Resources cleaned up on error

---

## Notes

This spec describes the current transport implementation. Key features:
- WebSocket connection management
- HTTP route handlers (/ws, /healthz)
- Message routing and parsing
- Entity-to-protocol conversion
- Session handler integration
- Ping/pong keepalive
- Observability integration

Future extensions may include:
- Origin whitelist for WebSocket upgrade
- Connection rate limiting
- Message compression
- Binary message encoding
- Connection pooling

