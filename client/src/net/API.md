# Orbital Rush v0 – Network Package API

This document provides detailed API reference for the network package.

---

## Scope & Location

**Scope**: Network communication layer (WebSocket, protocol types, command history).

**Code location**: `client/src/net`

**Design Goals**:
- Provide WebSocket connection management
- Define protocol message types matching server contracts
- Track command history for prediction/reconciliation
- Abstract network communication details
- Handle message validation and type safety

---

## NetworkClient

**File**: `client.ts`

**Concept**: High-level network client that manages WebSocket connection and message routing.

### Interface

```typescript
export class NetworkClient {
  constructor()
  async connect(url: string): Promise<void>
  disconnect(): void
  sendInput(seq: number, thrust: number, turn: number): void
  sendRestart(): void
  onSnapshot(callback: (snapshot: SnapshotMessage) => void): void
  onConnect(callback: () => void): void
  onDisconnect(callback: () => void): void
  onError(callback: (error: Error) => void): void
  isConnected(): boolean
}
```

### Methods

#### `constructor()`

Creates a new NetworkClient instance.

**Returns:** NetworkClient instance

---

#### `async connect(url: string): Promise<void>`

Connects to WebSocket server at the specified URL.

**Parameters:**
- `url: string` - WebSocket server URL (e.g., 'ws://localhost:8080/ws')

**Returns:** Promise that resolves when connected

**Throws:** Error if connection fails

**Example:**
```typescript
await client.connect('ws://localhost:8080/ws')
```

---

#### `disconnect(): void`

Disconnects from the server.

**Example:**
```typescript
client.disconnect()
```

---

#### `sendInput(seq: number, thrust: number, turn: number): void`

Sends an input command to the server.

**Parameters:**
- `seq: number` - Sequence number (must be > 0)
- `thrust: number` - Thrust value [0.0, 1.0]
- `turn: number` - Turn value [-1.0, 1.0]

**Throws:** Error if not connected

**Example:**
```typescript
client.sendInput(1, 0.5, 0.0)
```

---

#### `sendRestart(): void`

Sends a restart command to the server.

**Throws:** Error if not connected

**Example:**
```typescript
client.sendRestart()
```

---

#### `onSnapshot(callback: (snapshot: SnapshotMessage) => void): void`

Registers callback for snapshot messages from server.

**Parameters:**
- `callback: (snapshot: SnapshotMessage) => void` - Function to call when snapshot received

**Example:**
```typescript
client.onSnapshot((snapshot) => {
  console.log('Received snapshot:', snapshot.tick)
})
```

---

#### `onConnect(callback: () => void): void`

Registers callback for connection events.

**Parameters:**
- `callback: () => void` - Function to call when connected

---

#### `onDisconnect(callback: () => void): void`

Registers callback for disconnection events.

**Parameters:**
- `callback: () => void` - Function to call when disconnected

---

#### `onError(callback: (error: Error) => void): void`

Registers callback for error events.

**Parameters:**
- `callback: (error: Error) => void` - Function to call when error occurs

---

#### `isConnected(): boolean`

Returns true if connected to server.

**Returns:** true if connected, false otherwise

### Semantics

- NetworkClient wraps WebSocketClient for higher-level API
- Handles message routing (only processes snapshot messages)
- Converts server 'sun' field to client 'planets' array for extensibility
- Event handlers are arrays (multiple handlers supported)
- Connection state managed internally

### Lifecycle

1. **Creation**: `new NetworkClient()` – creates client instance
2. **Registration**: Register event handlers with `on*()` methods
3. **Connection**: `await client.connect(url)` – connects to server
4. **Usage**: Send messages with `sendInput()` or `sendRestart()`
5. **Disconnection**: `client.disconnect()` – disconnects from server

### Invariants

- Must be connected before sending messages (throws error if not)
- Event handlers persist until disconnect
- Snapshot messages are validated before calling handlers
- Connection state is consistent with underlying WebSocket

---

## WebSocketClient

**File**: `ws.ts`

**Concept**: Low-level WebSocket wrapper with connection lifecycle management.

### Interface

```typescript
export class WebSocketClient {
  constructor()
  async connect(url: string): Promise<void>
  disconnect(): void
  send(message: object): void
  onMessage(callback: (data: any) => void): void
  onOpen(callback: () => void): void
  onClose(callback: () => void): void
  onError(callback: (error: Error) => void): void
  isConnected(): boolean
  getReadyState(): number
}
```

### Methods

#### `constructor()`

Creates a new WebSocketClient instance.

**Returns:** WebSocketClient instance

---

#### `async connect(url: string): Promise<void>`

Connects to WebSocket server.

**Parameters:**
- `url: string` - WebSocket server URL

**Returns:** Promise that resolves on connection, rejects on error

**Throws:** Error if connection fails

---

#### `disconnect(): void`

Disconnects from server.

---

#### `send(message: object): void`

Sends JSON message over WebSocket.

**Parameters:**
- `message: object` - Message object to send (will be JSON stringified)

**Throws:** Error if not connected

---

#### `onMessage(callback: (data: any) => void): void`

Registers callback for incoming messages.

**Parameters:**
- `callback: (data: any) => void` - Function to call when message received

---

#### `onOpen(callback: () => void): void`

Registers callback for connection open events.

---

#### `onClose(callback: () => void): void`

Registers callback for connection close events.

---

#### `onError(callback: (error: Error) => void): void`

Registers callback for error events.

---

#### `isConnected(): boolean`

Returns true if WebSocket is connected.

**Returns:** true if connected, false otherwise

---

#### `getReadyState(): number`

Returns WebSocket ready state.

**Returns:** WebSocket ready state (CONNECTING, OPEN, CLOSING, CLOSED)

### Semantics

- Wraps native WebSocket API
- Automatically parses JSON messages
- Handles connection errors and malformed JSON
- Event handlers are arrays (multiple handlers supported)
- Promise-based connection (resolves on open, rejects on error)

### Lifecycle

1. **Creation**: `new WebSocketClient()` – creates client instance
2. **Registration**: Register event handlers
3. **Connection**: `await ws.connect(url)` – connects to server
4. **Usage**: Send/receive messages
5. **Disconnection**: `ws.disconnect()` – closes connection

### Invariants

- Must be connected before sending (throws error if not)
- Messages are automatically JSON stringified
- Incoming messages are parsed as JSON (errors handled)
- Event handlers persist until disconnect
- Ready state matches underlying WebSocket state

---

## CommandHistory

**File**: `command-history.ts`

**Concept**: Tracks sent commands with sequence numbers for client-side prediction and reconciliation.

### Interface

```typescript
export interface CommandEntry {
  seq: number
  thrust: number
  turn: number
  timestamp: number
  confirmed: boolean
}

export class CommandHistory {
  constructor()
  getNextSequence(): number
  addCommand(seq: number, thrust: number, turn: number): void
  markConfirmed(seq: number): void
  markConfirmedUpTo(seq: number): void
  getUnconfirmed(): CommandEntry[]
  getCommand(seq: number): CommandEntry | null
  clear(): void
}
```

### Methods

#### `constructor()`

Creates a new CommandHistory instance.

**Returns:** CommandHistory instance

---

#### `getNextSequence(): number`

Returns the next sequence number for a new command.

**Returns:** Next sequence number (starts at 1, increments)

**Example:**
```typescript
const seq = history.getNextSequence()
```

---

#### `addCommand(seq: number, thrust: number, turn: number): void`

Adds a command to history.

**Parameters:**
- `seq: number` - Sequence number
- `thrust: number` - Thrust value [0.0, 1.0]
- `turn: number` - Turn value [-1.0, 1.0]

**Example:**
```typescript
history.addCommand(1, 0.5, 0.0)
```

---

#### `markConfirmed(seq: number): void`

Marks a command as confirmed by server.

**Parameters:**
- `seq: number` - Sequence number of command to confirm

---

#### `markConfirmedUpTo(seq: number): void`

Marks all commands up to and including seq as confirmed.

**Parameters:**
- `seq: number` - Sequence number up to which commands should be confirmed

---

#### `getUnconfirmed(): CommandEntry[]`

Returns all unconfirmed commands in sequence order.

**Returns:** Array of unconfirmed commands sorted by sequence number

**Example:**
```typescript
const unconfirmed = history.getUnconfirmed()
```

---

#### `getCommand(seq: number): CommandEntry | null`

Gets command by sequence number.

**Parameters:**
- `seq: number` - Sequence number

**Returns:** Command entry or null if not found

---

#### `clear(): void`

Clears all commands and resets sequence to 1.

**Example:**
```typescript
history.clear()
```

### Semantics

- Sequence numbers start at 1 and increment
- Commands stored in Map keyed by sequence number
- Confirmed commands are marked but not removed
- Unconfirmed commands returned in sequence order (sorted)
- Timestamp recorded when command added

### Lifecycle

1. **Creation**: `new CommandHistory()` – creates empty history
2. **Usage**: Add commands with `addCommand()`, mark confirmed with `markConfirmed()`
3. **Query**: Get unconfirmed commands with `getUnconfirmed()`
4. **Reset**: `clear()` – clears all commands and resets sequence

### Invariants

- Sequence numbers are unique and monotonically increasing
- Commands are never removed (only marked confirmed)
- Unconfirmed commands are always sorted by sequence
- Next sequence is always > highest sequence in history
- Clear resets sequence to 1

---

## Protocol Types

**File**: `protocol.ts`

**Concept**: Protocol message types, validation functions, and type guards matching server contracts.

### Interfaces

```typescript
export const PROTOCOL_VERSION = 1

export interface Vec2Snapshot {
  x: number
  y: number
}

export interface ShipSnapshot {
  pos: Vec2Snapshot
  vel: Vec2Snapshot
  rot: number
  energy: number
}

export interface PlanetSnapshot {
  pos: Vec2Snapshot
  radius: number
}

export interface PalletSnapshot {
  id: number
  pos: Vec2Snapshot
  active: boolean
}

export interface InputMessage {
  t: 'input'
  seq: number
  thrust: number
  turn: number
  version?: number
}

export interface RestartMessage {
  t: 'restart'
  version?: number
}

export interface SnapshotMessage {
  t: 'snapshot'
  tick: number
  ship: ShipSnapshot
  planets: PlanetSnapshot[]
  pallets: PalletSnapshot[]
  done: boolean
  win: boolean
  version?: number
}

export type Message = InputMessage | RestartMessage | SnapshotMessage
```

### Type Guards

- `isInputMessage(msg: unknown): msg is InputMessage`
- `isRestartMessage(msg: unknown): msg is RestartMessage`
- `isSnapshotMessage(msg: unknown): msg is SnapshotMessage`
- `isMessage(msg: unknown): msg is Message`

### Validation Functions

- `validateVec2Snapshot(v: unknown): Vec2Snapshot | null`
- `validateShipSnapshot(s: unknown): ShipSnapshot | null`
- `validatePlanetSnapshot(p: unknown): PlanetSnapshot | null`
- `validatePalletSnapshot(p: unknown): PalletSnapshot | null`

### Factory Functions

- `createInputMessage(seq: number, thrust: number, turn: number): InputMessage`
- `createRestartMessage(): RestartMessage`

### Semantics

- Protocol types match server G4 contracts exactly
- Type guards perform runtime validation
- Validation functions return null on invalid input
- Factory functions create valid messages
- Message types use discriminated union ('t' field)

### Invariants

- All protocol types match server schema
- Type guards are exhaustive (cover all cases)
- Validation functions are pure (no side effects)
- Factory functions always produce valid messages
- Message type field ('t') is required and literal

---

## Ownership & Dependencies

### Network Package Ownership

- **Only `client/src/net` may define network communication logic**
- Network package handles WebSocket but does not implement game logic
- Network package defines protocol types but does not implement simulation
- Protocol types must match server G4 contracts

### Dependencies

- **Imports**:
  - Native WebSocket API (browser)
- **No dependencies on**: core, gfx, sim, input, ui packages
- Network is independent layer (can be used by sim for command history)

### No Duplication Rules

- **No WebSocket handling elsewhere**: Connection management must live in `/net`
- **No protocol types elsewhere**: Message types must live in `/net/protocol`
- **No command history elsewhere**: Command tracking must live in `/net/command-history`
- **Network does not implement**: Simulation, rendering, or input handling

---

## Error Handling

### NetworkClient Errors

- **Not connected**: Thrown when `sendInput()` or `sendRestart()` called before connection
- **Connection failed**: Thrown when `connect()` fails

### WebSocketClient Errors

- **Not connected**: Thrown when `send()` called before connection
- **Connection failed**: Thrown when `connect()` fails
- **Malformed JSON**: Handled internally, error callback invoked

---

## Dependencies

- Native WebSocket API (browser)

---

## Version Notes

This API describes v0 network layer. Key features:
- WebSocket client wrapper
- Protocol types matching server v0
- Command history for prediction/reconciliation
- Type guards and validation
- Event-based message handling

Future extensions may include:
- Binary message encoding
- Message compression
- Connection retry logic
- Protocol version negotiation
- Delta snapshots

