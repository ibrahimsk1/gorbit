# Orbital Rush – Protocol Subsystem Specification

This document describes the network protocol contract for Orbital Rush. It defines message schemas, validation rules, and versioning policy for client-server communication.

---

## Scope & Location

**Scope**: Network protocol contract for Orbital Rush (message types, validation, versioning).

**Code location**: `server/internal/proto`

**Design Goals**:
- Frozen protocol contract that must be honored by all adapters
- JSON messages over WebSocket for real-time communication
- Server is authoritative; clients send input, server broadcasts state
- Protocol types are separate from entity types (used for serialization only)

---

## Protocol Version

**Current Version**: `v1` (see `ProtocolVersionV1` constant)

**Version Format**: `"v"` followed by a positive integer (e.g., `"v1"`, `"v2"`)

**Compatibility Rules**:
- Same major version: Compatible (e.g., `v1` ↔ `v1`)
- Different major versions: Incompatible (e.g., `v1` ↔ `v2`)

**Compatibility Function**: `IsCompatible(clientVersion, serverVersion ProtocolVersion) bool`

---

## Message Schemas

### Client → Server Messages

#### Room Management Messages

##### CreateRoomMessage

**Purpose**: Client request to create a new room.

**JSON Schema**:
```json
{
  "t": "createRoom"
}
```

**Fields**:
- `t` (string, required): Message type, must be `"createRoom"`

**Semantics**:
- Server creates new room, generates unique room code
- Returns `RoomCreatedMessage` with room code

**Validation Rules**:
- `Type` must equal `"createRoom"`

**Validation Function**: `ValidateCreateRoomMessage(msg *CreateRoomMessage) error`

---

##### JoinRoomMessage

**Purpose**: Client request to join an existing room by code.

**JSON Schema**:
```json
{
  "t": "joinRoom",
  "roomCode": <string>
}
```

**Fields**:
- `t` (string, required): Message type, must be `"joinRoom"`
- `roomCode` (string, required): 6-character alphanumeric room code

**Semantics**:
- Server adds player to room if code is valid and room has capacity
- Returns `RoomStateMessage` or error message

**Validation Rules**:
- `Type` must equal `"joinRoom"`
- `RoomCode` must be exactly 6 characters, alphanumeric (A-Z, 0-9)

**Validation Function**: `ValidateJoinRoomMessage(msg *JoinRoomMessage) error`

---

##### LeaveRoomMessage

**Purpose**: Client request to leave current room.

**JSON Schema**:
```json
{
  "t": "leaveRoom"
}
```

**Fields**:
- `t` (string, required): Message type, must be `"leaveRoom"`

**Semantics**:
- Server removes player from room
- Room is cleaned up if empty

**Validation Rules**:
- `Type` must equal `"leaveRoom"`

**Validation Function**: `ValidateLeaveRoomMessage(msg *LeaveRoomMessage) error`

---

##### StartMatchMessage

**Purpose**: Host request to start the match (host only).

**JSON Schema**:
```json
{
  "t": "startMatch"
}
```

**Fields**:
- `t` (string, required): Message type, must be `"startMatch"`

**Semantics**:
- Only room host can start match
- Requires minimum 2 players in room
- Server creates session and starts match

**Validation Rules**:
- `Type` must equal `"startMatch"`

**Validation Function**: `ValidateStartMatchMessage(msg *StartMatchMessage) error`

---

#### InputMessage

**Purpose**: Client input command for controlling the ship (in-game).

**JSON Schema**:
```json
{
  "t": "input",
  "seq": <uint32>,
  "thrust": <float32>,
  "turn": <float32>
}
```

**Fields**:
- `t` (string, required): Message type, must be `"input"`
- `seq` (uint32, required): Sequence number, must be > 0
- `thrust` (float32, required): Thrust input, range [0.0, 1.0]
- `turn` (float32, required): Turn input, range [-1.0, 1.0]

**Semantics**:
- Sequence numbers are used for deduplication and ordering
- Thrust and turn values are clamped to valid ranges by validation
- Input is processed by rules layer to update ship state (ship identified by player ID from connection)

**Validation Rules**:
- `Type` must equal `"input"`
- `Seq` must be > 0
- `Thrust` must be in range [0.0, 1.0]
- `Turn` must be in range [-1.0, 1.0]

**Validation Function**: `ValidateInputMessage(msg *InputMessage) error`

---

### Server → Client Messages

#### Room Management Messages

##### RoomCreatedMessage

**Purpose**: Server response to createRoom request.

**JSON Schema**:
```json
{
  "t": "roomCreated",
  "roomCode": <string>
}
```

**Fields**:
- `t` (string, required): Message type, must be `"roomCreated"`
- `roomCode` (string, required): 6-character alphanumeric room code

**Validation Function**: `ValidateRoomCreatedMessage(msg *RoomCreatedMessage) error`

---

##### RoomStateMessage

**Purpose**: Server room state update (sent when room state changes or player joins/leaves).

**JSON Schema**:
```json
{
  "t": "roomState",
  "roomCode": <string>,
  "players": [<PlayerInfo>],
  "state": <string>,
  "hostId": <uint32>
}
```

**Fields**:
- `t` (string, required): Message type, must be `"roomState"`
- `roomCode` (string, required): Room code
- `players` (array of PlayerInfo, required): List of players in room
- `state` (string, required): Room state ("lobby", "playing", "ended")
- `hostId` (uint32, required): Host player ID

**PlayerInfo Schema**:
```json
{
  "id": <uint32>,
  "name": <string>
}
```

**Validation Function**: `ValidateRoomStateMessage(msg *RoomStateMessage) error`

---

##### PlayerJoinedMessage

**Purpose**: Server notification when a player joins the room.

**JSON Schema**:
```json
{
  "t": "playerJoined",
  "player": <PlayerInfo>
}
```

**Validation Function**: `ValidatePlayerJoinedMessage(msg *PlayerJoinedMessage) error`

---

##### PlayerLeftMessage

**Purpose**: Server notification when a player leaves the room.

**JSON Schema**:
```json
{
  "t": "playerLeft",
  "playerId": <uint32>
}
```

**Validation Function**: `ValidatePlayerLeftMessage(msg *PlayerLeftMessage) error`

---

##### MatchStartedMessage

**Purpose**: Server notification when match starts.

**JSON Schema**:
```json
{
  "t": "matchStarted"
}
```

**Validation Function**: `ValidateMatchStartedMessage(msg *MatchStartedMessage) error`

---

##### MatchEndedMessage

**Purpose**: Server notification when match ends (optional).

**JSON Schema**:
```json
{
  "t": "matchEnded",
  "winnerId": <uint32>
}
```

**Fields**:
- `winnerId` (uint32, optional): Winner player ID (if applicable)

**Validation Function**: `ValidateMatchEndedMessage(msg *MatchEndedMessage) error`

---

#### SnapshotMessage

**Purpose**: Server state snapshot containing the complete game state (in-game).

**JSON Schema**:
```json
{
  "t": "snapshot",
  "tick": <uint32>,
  "ships": [<ShipSnapshot>],
  "planets": [<PlanetSnapshot>],
  "pallets": [<PalletSnapshot>],
  "worldBounds": <WorldBounds>,
  "myShipId": <uint32>,
  "done": <bool>,
  "win": <bool>
}
```

**Fields**:
- `t` (string, required): Message type, must be `"snapshot"`
- `tick` (uint32, required): Current simulation tick
- `ships` (array of ShipSnapshot, required): All ships in match (2–8 ships)
- `planets` (array of PlanetSnapshot, required): All planets in match (3–5 planets)
- `pallets` (array of PalletSnapshot, required): List of energy pallets
- `worldBounds` (WorldBounds, required): World bounds (width, height)
- `myShipId` (uint32, required): Player's ship ID (identifies which ship belongs to this player)
- `done` (bool, required): Whether the game is finished
- `win` (bool, required): Whether the player won (only valid if Done is true)

**Semantics**:
- Snapshot contains complete authoritative game state
- Broadcast to all players in room at regular intervals (typically 10-15 Hz)
- Clients use snapshots for rendering and state synchronization
- Tick number increments monotonically
- `myShipId` tells client which ship in `ships` array belongs to them

**Validation Rules**:
- `Type` must equal `"snapshot"`
- All `Ships` must be valid (see ShipSnapshot validation)
- All `Planets` must be valid (see PlanetSnapshot validation)
- All `Pallets` must be valid (see PalletSnapshot validation)
- `WorldBounds` must be valid (see WorldBounds validation)
- `MyShipId` must be > 0

**Validation Function**: `ValidateSnapshotMessage(msg *SnapshotMessage) error`

---

### Snapshot Sub-Types

#### ShipSnapshot

**JSON Schema**:
```json
{
  "id": <uint32>,
  "pos": <Vec2Snapshot>,
  "vel": <Vec2Snapshot>,
  "rot": <float64>,
  "energy": <float32>
}
```

**Fields**:
- `id` (uint32, required): Ship identifier (player ID)
- `pos` (Vec2Snapshot, required): Position
- `vel` (Vec2Snapshot, required): Velocity
- `rot` (float64, required): Rotation angle in radians
- `energy` (float32, required): Current energy level

**Validation Rules**:
- `ID` must be > 0
- `Pos` must be valid (see Vec2Snapshot validation)
- `Vel` must be valid (see Vec2Snapshot validation)
- `Energy` must be >= 0.0

**Validation Function**: `ValidateShipSnapshot(ship *ShipSnapshot) error`

---

#### PlanetSnapshot

**JSON Schema**:
```json
{
  "id": <uint32>,
  "pos": <Vec2Snapshot>,
  "radius": <float32>
}
```

**Fields**:
- `id` (uint32, required): Planet identifier
- `pos` (Vec2Snapshot, required): Position
- `radius` (float32, required): Radius

**Validation Rules**:
- `ID` must be > 0
- `Pos` must be valid (see Vec2Snapshot validation)
- `Radius` must be > 0.0

**Note**: Mass is not included in snapshot as it is only used for simulation calculations.

**Validation Function**: `ValidatePlanetSnapshot(planet *PlanetSnapshot) error`

---

#### WorldBounds

**JSON Schema**:
```json
{
  "width": <float64>,
  "height": <float64>
}
```

**Fields**:
- `width` (float64, required): World width in meters (2000.0 m)
- `height` (float64, required): World height in meters (2000.0 m)

**Validation Rules**:
- `Width` must be > 0.0
- `Height` must be > 0.0

**Validation Function**: `ValidateWorldBounds(bounds *WorldBounds) error`

---

#### PalletSnapshot

**JSON Schema**:
```json
{
  "id": <uint32>,
  "pos": <Vec2Snapshot>,
  "active": <bool>
}
```

**Fields**:
- `id` (uint32, required): Pallet identifier
- `pos` (Vec2Snapshot, required): Position
- `active` (bool, required): Whether the pallet is active/collectible

**Validation Rules**:
- `ID` must be > 0
- `Pos` must be valid (see Vec2Snapshot validation)

**Validation Function**: `ValidatePalletSnapshot(pallet *PalletSnapshot) error`

---

#### Vec2Snapshot

**JSON Schema**:
```json
{
  "x": <float64>,
  "y": <float64>
}
```

**Fields**:
- `x` (float64, required): X coordinate
- `y` (float64, required): Y coordinate

**Validation Rules**:
- `X` must be finite (not NaN, not Inf)
- `Y` must be finite (not NaN, not Inf)

**Validation Function**: `ValidateVec2Snapshot(vec *Vec2Snapshot) error`

---

## Validation

### Validation Principles

- **All messages must be validated before processing**
- Validation functions return errors describing validation failures
- Invalid messages are rejected (not processed)
- Validation ensures protocol contract is honored

### Validation Functions

**Room Management Messages**:
- `ValidateCreateRoomMessage(msg *CreateRoomMessage) error`
- `ValidateJoinRoomMessage(msg *JoinRoomMessage) error`
- `ValidateLeaveRoomMessage(msg *LeaveRoomMessage) error`
- `ValidateStartMatchMessage(msg *StartMatchMessage) error`
- `ValidateRoomCreatedMessage(msg *RoomCreatedMessage) error`
- `ValidateRoomStateMessage(msg *RoomStateMessage) error`
- `ValidatePlayerJoinedMessage(msg *PlayerJoinedMessage) error`
- `ValidatePlayerLeftMessage(msg *PlayerLeftMessage) error`
- `ValidateMatchStartedMessage(msg *MatchStartedMessage) error`
- `ValidateMatchEndedMessage(msg *MatchEndedMessage) error`

**Game Messages**:
- `ValidateInputMessage(msg *InputMessage) error`
- `ValidateSnapshotMessage(msg *SnapshotMessage) error`
- `ValidateShipSnapshot(ship *ShipSnapshot) error`
- `ValidatePlanetSnapshot(planet *PlanetSnapshot) error`
- `ValidatePalletSnapshot(pallet *PalletSnapshot) error`
- `ValidateVec2Snapshot(vec *Vec2Snapshot) error`
- `ValidateWorldBounds(bounds *WorldBounds) error`

### Common Validation Rules

- Message type fields must match expected values
- Numeric ranges must be within specified bounds
- Required fields must be present
- Float values must be finite (not NaN, not Inf)
- Array fields must contain valid elements

---

## Ownership & Dependencies

### Protocol Package Ownership

- **Only `server/internal/proto` may define protocol message types and validation**
- Protocol types are separate from entity types (used for serialization/deserialization)
- Protocol types mirror entity types but are optimized for JSON transport
- Validation logic lives in protocol package

### Dependencies

- **No dependencies on**: entities, physics, rules, session, transport packages
- Protocol is a contract layer that defines wire format
- Transport layer uses protocol types for WebSocket communication

### No Duplication Rules

- **No protocol types elsewhere**: Message schemas and validation must live in `/proto`
- **Protocol vs Entities**: Protocol types mirror entities but are separate (e.g., `ShipSnapshot` vs `Ship`)
- **Validation centralization**: All message validation logic lives in protocol package

### Conversion

- Transport layer converts between entity types and protocol types
- Conversion functions (e.g., `ShipToSnapshot`, `WorldToSnapshot`) live in transport package
- Conversion is one-way: entities → protocol (for snapshots), protocol → entities (for input)

---

## Versioning Policy

### Version Format

Protocol versions use a simple major version format: `"v1"`, `"v2"`, etc.

- **Current Version**: `v1` (see `ProtocolVersionV1` constant)
- **Format**: `"v"` followed by a positive integer

### Compatibility Rules

- **Same major version**: Compatible (e.g., `v1` ↔ `v1`)
- **Different major versions**: Incompatible (e.g., `v1` ↔ `v2`)

**Compatibility Function**: `IsCompatible(clientVersion, serverVersion ProtocolVersion) bool`

### Breaking vs Non-Breaking Changes

#### Breaking Changes (require major version increment)

- Removing fields from messages
- Changing field types
- Changing required fields to optional (or vice versa)
- Changing message structure significantly
- Changing JSON field names

#### Non-Breaking Changes (same major version)

- Adding new optional fields
- Adding new message types
- Extending validation rules (making them stricter)
- Documentation updates

### Version Functions

- `ParseVersion(versionStr string) (ProtocolVersion, error)`: Parse a version string
- `IsCompatible(clientVersion, serverVersion ProtocolVersion) bool`: Check compatibility
- `CompareVersion(v1, v2 ProtocolVersion) int`: Compare versions (-1, 0, or 1)

---

## Contract Stability

This protocol contract is **frozen** for the current version. All adapters must honor this contract exactly as documented.

### Schema Stability

- JSON field names are fixed and cannot change
- Field types are fixed and cannot change
- Required fields cannot be removed
- Message structure is fixed

### Forward Compatibility

- Messages with extra JSON fields are handled gracefully (extra fields are ignored)
- This allows future versions to add optional fields without breaking existing clients

### Backward Compatibility

- All current fields are required
- Future versions may add optional fields that older clients can ignore

---

## Usage Examples

### Validating an InputMessage

```go
import "github.com/gorbit/orbitalrush/internal/proto"

msg := &proto.InputMessage{
    Type:   "input",
    Seq:    1,
    Thrust: 0.5,
    Turn:   0.3,
}

if err := proto.ValidateInputMessage(msg); err != nil {
    // Handle validation error
    log.Printf("Invalid input message: %v", err)
}
```

### Checking Protocol Compatibility

```go
clientVersion := proto.ProtocolVersionV1
serverVersion := proto.ProtocolVersionV1

if proto.IsCompatible(clientVersion, serverVersion) {
    // Versions are compatible
} else {
    // Versions are incompatible, handle upgrade
}
```

### Parsing a Version String

```go
version, err := proto.ParseVersion("v1")
if err != nil {
    // Handle parse error
    log.Printf("Invalid version: %v", err)
}
```

---

## Notes

This spec describes the v1 protocol. Key features:
- JSON messages over WebSocket
- Room management messages (createRoom, joinRoom, leaveRoom, startMatch)
- Room state updates (roomState, playerJoined, playerLeft, matchStarted, matchEnded)
- Input messages with sequence numbers (in-game)
- Snapshot messages with complete game state (ships array, planets array, pallets array, worldBounds, myShipId)
- Protocol versioning with compatibility checking

**Changes from v0**:
- Added room management messages (createRoom, joinRoom, leaveRoom, startMatch)
- Added room state messages (roomState, playerJoined, playerLeft, matchStarted, matchEnded)
- Snapshot format changed: single `ship` → `ships[]` array, `sun` → `planets[]` array
- Added `worldBounds` and `myShipId` fields to snapshot
- Removed `restart` message (not in v1 scope)
- ShipSnapshot and PlanetSnapshot now include `id` field

Future extensions may include:
- Delta snapshots (only changed entities)
- Binary message encoding for bandwidth optimization
- Compression
- Message batching
- Protocol version negotiation

---

## Testing

The package includes comprehensive contract tests that verify:
- Message serialization/deserialization round-trip integrity
- Validation rules for all message types
- Schema compatibility (forward/backward)
- Breaking change detection
- Edge cases (large numbers, empty arrays, etc.)

All tests are labeled with `scope:contract loop:g4-proto layer:contract`.

