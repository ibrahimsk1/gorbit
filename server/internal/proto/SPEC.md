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

#### InputMessage

**Purpose**: Client input command for controlling the ship.

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
- Input is processed by rules layer to update ship state

**Validation Rules**:
- `Type` must equal `"input"`
- `Seq` must be > 0
- `Thrust` must be in range [0.0, 1.0]
- `Turn` must be in range [-1.0, 1.0]

**Validation Function**: `ValidateInputMessage(msg *InputMessage) error`

---

#### RestartMessage

**Purpose**: Client request to restart the game.

**JSON Schema**:
```json
{
  "t": "restart"
}
```

**Fields**:
- `t` (string, required): Message type, must be `"restart"`

**Semantics**:
- Resets the game world to initial state
- Typically handled by session/orchestration layer

**Validation Rules**:
- `Type` must equal `"restart"`

**Validation Function**: `ValidateRestartMessage(msg *RestartMessage) error`

---

### Server → Client Messages

#### SnapshotMessage

**Purpose**: Server state snapshot containing the complete game state.

**JSON Schema**:
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

**Fields**:
- `t` (string, required): Message type, must be `"snapshot"`
- `tick` (uint32, required): Current simulation tick
- `ship` (ShipSnapshot, required): Ship state
- `sun` (SunSnapshot, required): Sun state
- `pallets` (array of PalletSnapshot, required): List of energy pallets
- `done` (bool, required): Whether the game is finished
- `win` (bool, required): Whether the player won (only valid if Done is true)

**Semantics**:
- Snapshot contains complete authoritative game state
- Broadcast to all players at regular intervals (typically 10-15 Hz)
- Clients use snapshots for rendering and state synchronization
- Tick number increments monotonically

**Validation Rules**:
- `Type` must equal `"snapshot"`
- All `Ship` fields must be valid (see ShipSnapshot validation)
- All `Sun` fields must be valid (see SunSnapshot validation)
- All `Pallets` must be valid (see PalletSnapshot validation)

**Validation Function**: `ValidateSnapshotMessage(msg *SnapshotMessage) error`

---

### Snapshot Sub-Types

#### ShipSnapshot

**JSON Schema**:
```json
{
  "pos": <Vec2Snapshot>,
  "vel": <Vec2Snapshot>,
  "rot": <float64>,
  "energy": <float32>
}
```

**Fields**:
- `pos` (Vec2Snapshot, required): Position
- `vel` (Vec2Snapshot, required): Velocity
- `rot` (float64, required): Rotation angle in radians
- `energy` (float32, required): Current energy level

**Validation Rules**:
- `Pos` must be valid (see Vec2Snapshot validation)
- `Vel` must be valid (see Vec2Snapshot validation)
- `Energy` must be >= 0.0

**Validation Function**: `ValidateShipSnapshot(ship *ShipSnapshot) error`

---

#### SunSnapshot

**JSON Schema**:
```json
{
  "pos": <Vec2Snapshot>,
  "radius": <float32>
}
```

**Fields**:
- `pos` (Vec2Snapshot, required): Position
- `radius` (float32, required): Radius

**Validation Rules**:
- `Pos` must be valid (see Vec2Snapshot validation)
- `Radius` must be > 0.0

**Note**: Mass is not included in snapshot as it is only used for simulation calculations.

**Validation Function**: `ValidateSunSnapshot(sun *SunSnapshot) error`

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

- `ValidateInputMessage(msg *InputMessage) error`
- `ValidateRestartMessage(msg *RestartMessage) error`
- `ValidateSnapshotMessage(msg *SnapshotMessage) error`
- `ValidateShipSnapshot(ship *ShipSnapshot) error`
- `ValidateSunSnapshot(sun *SunSnapshot) error`
- `ValidatePalletSnapshot(pallet *PalletSnapshot) error`
- `ValidateVec2Snapshot(vec *Vec2Snapshot) error`

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

This spec describes the current protocol. Key features:
- JSON messages over WebSocket
- Input messages with sequence numbers
- Snapshot messages with complete game state (ship, sun, pallets)
- Protocol versioning with compatibility checking

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

