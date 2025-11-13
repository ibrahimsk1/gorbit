# Protocol Contract Documentation

This package defines the frozen protocol contract for client-server communication in Orbital Rush. This contract must be honored by all G5+ adapters.

## Overview

The protocol uses JSON messages over WebSocket for real-time communication between the client and server. The server is authoritative and runs the simulation at 30 Hz, broadcasting snapshots to clients. Clients send input commands to control their ship.

**Protocol Version**: `v1` (see `ProtocolVersionV1` constant)

**Transport**: WebSocket (JSON messages)

## Message Schemas

### Client → Server Messages

#### InputMessage

Client input command for controlling the ship.

**JSON Schema:**
```json
{
  "t": "input",
  "seq": <uint32>,
  "thrust": <float32>,
  "turn": <float32>
}
```

**Fields:**
- `t` (string, required): Message type, must be `"input"`
- `seq` (uint32, required): Sequence number, must be > 0
- `thrust` (float32, required): Thrust input, range [0.0, 1.0]
- `turn` (float32, required): Turn input, range [-1.0, 1.0]

**Example:**
```json
{
  "t": "input",
  "seq": 42,
  "thrust": 0.75,
  "turn": -0.5
}
```

#### RestartMessage

Client request to restart the game.

**JSON Schema:**
```json
{
  "t": "restart"
}
```

**Fields:**
- `t` (string, required): Message type, must be `"restart"`

**Example:**
```json
{
  "t": "restart"
}
```

### Server → Client Messages

#### SnapshotMessage

Server state snapshot containing the complete game state.

**JSON Schema:**
```json
{
  "t": "snapshot",
  "tick": <uint32>,
  "ship": <ShipSnapshot>,
  "sun": <SunSnapshot>,
  "pallets": [<PalletSnapshot>],
  "done": <bool>,
  "win": <bool>
}
```

**Fields:**
- `t` (string, required): Message type, must be `"snapshot"`
- `tick` (uint32, required): Current simulation tick
- `ship` (ShipSnapshot, required): Ship state (see below)
- `sun` (SunSnapshot, required): Sun state (see below)
- `pallets` (array of PalletSnapshot, required): List of energy pallets
- `done` (bool, required): Whether the game is finished
- `win` (bool, required): Whether the player won (only valid if `done` is `true`)

**Example:**
```json
{
  "t": "snapshot",
  "tick": 100,
  "ship": {
    "pos": {"x": 10.5, "y": 20.3},
    "vel": {"x": 1.0, "y": -2.0},
    "rot": 1.57,
    "energy": 75.5
  },
  "sun": {
    "pos": {"x": 0.0, "y": 0.0},
    "radius": 5.0
  },
  "pallets": [
    {"id": 1, "pos": {"x": 15.0, "y": 15.0}, "active": true},
    {"id": 2, "pos": {"x": -10.0, "y": 10.0}, "active": false}
  ],
  "done": false,
  "win": false
}
```

### Nested Structures

#### ShipSnapshot

Ship state in a snapshot.

**JSON Schema:**
```json
{
  "pos": <Vec2Snapshot>,
  "vel": <Vec2Snapshot>,
  "rot": <float64>,
  "energy": <float32>
}
```

**Fields:**
- `pos` (Vec2Snapshot, required): Position vector
- `vel` (Vec2Snapshot, required): Velocity vector
- `rot` (float64, required): Rotation angle in radians
- `energy` (float32, required): Current energy level, must be >= 0.0

#### SunSnapshot

Sun state in a snapshot.

**JSON Schema:**
```json
{
  "pos": <Vec2Snapshot>,
  "radius": <float32>
}
```

**Fields:**
- `pos` (Vec2Snapshot, required): Position vector
- `radius` (float32, required): Radius, must be > 0.0

#### PalletSnapshot

Energy pallet state in a snapshot.

**JSON Schema:**
```json
{
  "id": <uint32>,
  "pos": <Vec2Snapshot>,
  "active": <bool>
}
```

**Fields:**
- `id` (uint32, required): Unique identifier, must be > 0
- `pos` (Vec2Snapshot, required): Position vector
- `active` (bool, required): Whether the pallet is active/collectible

#### Vec2Snapshot

2D vector in a snapshot.

**JSON Schema:**
```json
{
  "x": <float64>,
  "y": <float64>
}
```

**Fields:**
- `x` (float64, required): X coordinate, must be finite (not NaN, not Inf)
- `y` (float64, required): Y coordinate, must be finite (not NaN, not Inf)

## Validation Rules

All messages must be validated before processing. The package provides validation functions for each message type.

### InputMessage Validation

- `Type` must equal `"input"`
- `Seq` must be > 0
- `Thrust` must be in range [0.0, 1.0]
- `Turn` must be in range [-1.0, 1.0]

**Validation Function:** `ValidateInputMessage(msg *InputMessage) error`

### RestartMessage Validation

- `Type` must equal `"restart"`

**Validation Function:** `ValidateRestartMessage(msg *RestartMessage) error`

### SnapshotMessage Validation

- `Type` must equal `"snapshot"`
- `Ship` must be valid (see ShipSnapshot validation)
- `Sun` must be valid (see SunSnapshot validation)
- All `Pallets` must be valid (see PalletSnapshot validation)

**Validation Function:** `ValidateSnapshotMessage(msg *SnapshotMessage) error`

### ShipSnapshot Validation

- `Pos` must be valid (see Vec2Snapshot validation)
- `Vel` must be valid (see Vec2Snapshot validation)
- `Energy` must be >= 0.0

**Validation Function:** `ValidateShipSnapshot(ship *ShipSnapshot) error`

### SunSnapshot Validation

- `Pos` must be valid (see Vec2Snapshot validation)
- `Radius` must be > 0.0

**Validation Function:** `ValidateSunSnapshot(sun *SunSnapshot) error`

### PalletSnapshot Validation

- `ID` must be > 0
- `Pos` must be valid (see Vec2Snapshot validation)

**Validation Function:** `ValidatePalletSnapshot(pallet *PalletSnapshot) error`

### Vec2Snapshot Validation

- `X` must be finite (not NaN, not Inf)
- `Y` must be finite (not NaN, not Inf)

**Validation Function:** `ValidateVec2Snapshot(vec *Vec2Snapshot) error`

## Versioning Policy

### Version Format

Protocol versions use a simple major version format: `"v1"`, `"v2"`, etc.

- **Current Version**: `v1` (see `ProtocolVersionV1` constant)
- **Format**: `"v"` followed by a positive integer

### Compatibility Rules

- **Same major version**: Compatible (e.g., `v1` ↔ `v1`)
- **Different major versions**: Incompatible (e.g., `v1` ↔ `v2`)

**Compatibility Function:** `IsCompatible(clientVersion, serverVersion ProtocolVersion) bool`

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

## Contract Stability

This protocol contract is **frozen** for version `v1`. All G5+ adapters must honor this contract exactly as documented.

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

## Testing

The package includes comprehensive contract tests that verify:
- Message serialization/deserialization round-trip integrity
- Validation rules for all message types
- Schema compatibility (forward/backward)
- Breaking change detection
- Edge cases (large numbers, empty arrays, etc.)

All tests are labeled with `scope:contract loop:g4-proto layer:contract`.

## References

- TDD Specification: `docs/tdd_orbitalrush_v0.md`
- Implementation: `server/internal/proto/`
- Tests: `server/internal/proto/proto_test.go`

