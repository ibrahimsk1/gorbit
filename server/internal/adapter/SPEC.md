# Orbital Rush – Adapter Subsystem Specification

This document describes the adapter subsystem for Orbital Rush. It defines how the RoomManager is wired to the Transport layer, following clean architecture principles.

---

## Scope & Location

**Scope**: Adapter layer that wires RoomManager to Transport layer, avoiding circular dependencies.

**Code location**: `server/internal/adapter`

**Design Goals**:
- Wire RoomManager to Transport layer without circular dependencies
- Follow clean architecture principles (adapter pattern)
- Keep main.go minimal and focused on orchestration
- Provide thread-safe conversion between room and transport data types

---

## Core Components

### WireRoomToTransport

**File**: `server/internal/adapter/adapter.go`

**Concept**: Wires RoomManager to Transport layer by setting up the RoomOperations adapter.

**Function Signature**: `WireRoomToTransport(roomManager *room.RoomManager)`

**Semantics**:
- Sets up adapter functions that connect RoomManager methods to Transport's RoomOperations interface
- Should be called during server initialization, before any WebSocket connections are established
- Uses the adapter pattern to avoid circular dependencies (room package doesn't import transport)

**Architecture**:
- **Transport** depends on `RoomOperations` interface (abstraction)
- **Room** package doesn't depend on Transport (avoids circular dependency)
- **Adapter** package imports both and connects them (adapter pattern)

**Invariants**:
- Must be called before any WebSocket connections are established
- RoomManager must be initialized before wiring
- Adapter functions are thread-safe (they call thread-safe RoomManager methods)

---

### convertRoomToRoomData

**File**: `server/internal/adapter/adapter.go`

**Concept**: Helper function that converts `room.Room` to `transport.RoomData`.

**Function Signature**: `convertRoomToRoomData(rm *room.Room) transport.RoomData`

**Semantics**:
- Extracts thread-safe data from room using room's getter methods
- Converts `room.PlayerConnection` to `transport.PlayerData`
- Converts `room.RoomState` to string representation

**Data Conversion**:
- `RoomCode`: Direct copy
- `Players`: Converted from `[]*PlayerConnection` to `[]PlayerData`
- `State`: Converted from `RoomState` enum to string ("lobby", "playing", "ended")
- `HostPlayerID`: Direct copy

**Thread Safety**:
- Uses room's thread-safe getter methods (`GetPlayers()`, `GetState()`, `GetHostPlayerID()`)
- Returns copies of data to prevent external modification

**Invariants**:
- All room data access is thread-safe
- Returned data is a snapshot (not a reference to internal room state)
- State string conversion matches RoomState.String() implementation

---

## Adapter Functions

The adapter implements the following `RoomOperations` functions:

### CreateRoomFunc
- **Implementation**: `roomManager.CreateRoom()`
- **Returns**: Room code (string) and error

### JoinRoomFunc
- **Implementation**: `roomManager.JoinRoom(roomCode, conn)`
- **Returns**: RoomData (converted), playerID, and error
- **Conversion**: Uses `convertRoomToRoomData()` to convert room to transport format

### LeaveRoomFunc
- **Implementation**: `roomManager.LeaveRoom(roomCode, playerID)`
- **Returns**: Error

### GetRoomFunc
- **Implementation**: `roomManager.GetRoom(roomCode)`
- **Returns**: RoomData (converted) and error
- **Conversion**: Uses `convertRoomToRoomData()` to convert room to transport format

### StartMatchFunc
- **Implementation**: Currently returns error (not yet implemented in RoomManager)
- **TODO**: Implement when `RoomManager.StartMatch()` is available

### EnqueueCommandToRoomFunc
- **Implementation**: `roomManager.EnqueueCommandToRoom(roomCode, playerID, seq, cmd)`
- **Returns**: Error

### GetWorldFromRoomFunc
- **Implementation**: `roomManager.GetWorldFromRoom(roomCode)`
- **Returns**: World and error

---

## Usage

### Initialization Flow

1. **Initialize RoomManager**: `roomManager := room.NewRoomManager()`
2. **Wire to Transport**: `adapter.WireRoomToTransport(roomManager)`
3. **Start Server**: Transport layer can now use RoomManager through the adapter

### Example

```go
// In main.go
roomManager := room.NewRoomManager()
adapter.WireRoomToTransport(roomManager)
// Now transport.WebSocketHandler can use room operations
```

---

## Dependencies

### Imports
- `room` package (for RoomManager and Room types)
- `transport` package (for RoomOperations interface and types)
- `session` package (for Clock interface)
- `sim/entities` package (for World type)
- `sim/rules` package (for InputCommand type)

### No Circular Dependencies
- Room package doesn't import transport
- Transport package doesn't import room
- Adapter package imports both (adapter pattern)

---

## Testing

**Test File**: `adapter_test.go`

**Test Coverage**:
- Wiring function sets up adapter correctly
- Room data conversion works correctly
- Empty rooms are handled correctly
- Room state conversion is correct

**Labels**: `scope:integration loop:g9-scenario layer:adapter`

---

## Notes

This adapter layer follows clean architecture principles:
- **Separation of Concerns**: Adapter logic is isolated in its own package
- **Dependency Inversion**: Transport depends on RoomOperations abstraction, not RoomManager directly
- **Single Responsibility**: Adapter only handles wiring, not business logic
- **Testability**: Adapter can be tested independently

The adapter pattern allows:
- Room package to remain independent of transport layer
- Transport layer to use room operations without importing room package
- Main.go to stay minimal (just orchestration, no wiring logic)

---

## Future Extensions

- Implement `StartMatchFunc` when `RoomManager.StartMatch()` is available
- Add more adapter functions as needed
- Consider adding adapter for other subsystems if needed

