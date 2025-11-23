# Orbital Rush – Room Management Subsystem Specification

This document describes the room management subsystem for Orbital Rush. It defines how rooms are created, how players join/leave rooms, and how room lifecycle is managed.

---

## Scope & Location

**Scope**: Room lifecycle, player management, session coordination for multiplayer matches.

**Code location**: `server/internal/room`

**Design Goals**:
- Manage room lifecycle (create, join, leave, start match)
- Generate unique room codes (6-character alphanumeric)
- Coordinate player connections with simulation sessions
- Enforce room limits (2–8 players per room)
- In-memory room state (no persistence in v1)

---

## Core Components

### Room

**File**: `server/internal/room/room.go`

**Concept**: Represents a game room with players, state, and associated session.

**Key Fields**:
- `RoomCode string` – 6-character alphanumeric room code (unique identifier)
- `Players []*PlayerConnection` – List of connected players
- `State RoomState` – Current room state (lobby, playing, ended)
- `HostPlayerID uint32` – Player ID of room host (first player to join)
- `Session *session.Session` – Simulation session (created when match starts, nil in lobby)
- `mu sync.RWMutex` – Mutex for concurrent access

**Semantics**:
- Room manages one Session instance shared by all players
- Room state transitions: lobby → playing → ended
- Host player can start match (requires min 2 players)
- Room is cleaned up when all players leave

**Invariants**:
- RoomCode is unique (no collisions)
- Players array length is between 0 and 8
- HostPlayerID must be in Players array (if room has players)
- Session is nil in lobby state, non-nil in playing state
- Room state transitions are valid (lobby → playing → ended)

---

### PlayerConnection

**File**: `server/internal/room/room.go`

**Concept**: Represents a player's connection to a room.

**Key Fields**:
- `Conn *transport.Connection` – WebSocket connection
- `PlayerID uint32` – Unique player identifier
- `Name string` – Player name (optional, may be empty in v1)

**Semantics**:
- Each WebSocket connection maps to one PlayerConnection
- PlayerID is unique per room (assigned when player joins)
- Connection is closed when player leaves room

**Invariants**:
- PlayerID is unique within a room
- Connection is valid (not nil) while player is in room

---

### RoomManager

**File**: `server/internal/room/manager.go`

**Concept**: Manages all rooms, provides room operations (create, join, leave, start match).

**Key Operations**:
- `CreateRoom() (string, error)` – Create new room, return room code
- `JoinRoom(roomCode string, conn *transport.Connection) (*Room, uint32, error)` – Join room, return room and player ID
- `LeaveRoom(roomCode string, playerID uint32) error` – Remove player from room
- `StartMatch(roomCode string, hostPlayerID uint32) error` – Start match (host only, requires min 2 players)
- `GetRoom(roomCode string) (*Room, error)` – Get room by code
- `CleanupEmptyRooms()` – Remove rooms with no players

**Data Structure**:
- `rooms map[string]*Room` – Map from room code to room instance
- `mu sync.RWMutex` – Mutex for concurrent access

**Semantics**:
- RoomManager is singleton (one instance per server)
- All room operations are thread-safe (use mutex)
- Rooms are created on-demand, cleaned up when empty
- Room codes are generated with collision detection

**Invariants**:
- Room codes are unique (no collisions in rooms map)
- Room operations are thread-safe (mutex-protected)
- Empty rooms are eventually cleaned up

---

## Room Code Generation

**File**: `server/internal/room/manager.go`

**Algorithm**:
1. Character set: `"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"` (36 characters: A-Z, 0-9)
2. Generate 6 random characters:
   - Use `crypto/rand.Reader` for cryptographically secure randomness
   - For each character: generate random byte, map to character set using modulo 36
3. Check collision: if room code exists in rooms map, retry (max 10 retries)
4. Return unique room code

**Function Signature**: `GenerateRoomCode(rooms map[string]*Room) (string, error)`

**Semantics**:
- Room codes are 6-character alphanumeric (uppercase letters and digits)
- Collision detection ensures uniqueness
- Generation is cryptographically secure (not predictable)
- Max retries prevent infinite loops

**Invariants**:
- Room code is exactly 6 characters
- Room code contains only uppercase letters (A-Z) and digits (0-9)
- Room code is unique (no collisions)
- Generation is deterministic (same random seed → same codes, but uses crypto/rand)

---

## Room Lifecycle

### 1. Room Creation

**Trigger**: Player sends `createRoom` message

**Process**:
1. RoomManager generates unique room code
2. Create Room struct with:
   - RoomCode: generated code
   - Players: empty array
   - State: lobby
   - HostPlayerID: 0 (will be set when first player joins)
   - Session: nil
3. Add room to rooms map
4. Return room code to client

**Result**: Room exists in lobby state, waiting for players

---

### 2. Player Join

**Trigger**: Player sends `joinRoom(roomCode)` message

**Process**:
1. RoomManager finds room by code
2. Check room state: must be in lobby state (cannot join playing/ended room)
3. Check room capacity: must have < 8 players
4. Assign player ID (incrementing counter or UUID)
5. Create PlayerConnection with connection and player ID
6. Add player to room.Players array
7. If first player (room empty), set HostPlayerID
8. Broadcast roomState update to all players in room

**Result**: Player added to room, roomState broadcast to all players

---

### 3. Match Start

**Trigger**: Host sends `startMatch` message

**Process**:
1. RoomManager finds room by code
2. Verify host: playerID must match room.HostPlayerID
3. Check minimum players: room must have ≥ 2 players
4. Check room state: must be in lobby state
5. Create initial World state:
   - Generate planets (3–5 planets, see entities/SPEC.md)
   - Create ships for all players (one ship per player, with player ID)
   - Generate pallets (8–12 pallets, scales with player count)
6. Create Session with initial World
7. Start session.Run() in goroutine
8. Update room state: lobby → playing
9. Broadcast matchStarted message to all players

**Result**: Room in playing state, session running, match started

---

### 4. Match End

**Trigger**: Session reports world.Done == true

**Process**:
1. Session stops (world.Done == true)
2. Update room state: playing → ended
3. Broadcast matchEnded message to all players (optional, with winnerId)
4. Room remains in ended state (players can leave, room cleaned up when empty)

**Result**: Room in ended state, session stopped

---

### 5. Player Leave

**Trigger**: Player sends `leaveRoom` message or connection closes

**Process**:
1. RoomManager finds room by code
2. Remove player from room.Players array
3. Close player's WebSocket connection
4. If room is playing and player leaves:
   - Option 1: End match (all players must leave)
   - Option 2: Continue match (player's ship remains, but no input)
5. If room is empty (no players):
   - Stop session if running
   - Remove room from rooms map (cleanup)
6. Broadcast playerLeft message to remaining players

**Result**: Player removed from room, room cleaned up if empty

---

## Session Coordination

**Pattern**: One Session per Room (not per Connection)

**Semantics**:
- Room manages one Session instance shared by all players
- Session.Run() runs in dedicated goroutine (one per room)
- All player inputs are routed to room's session
- Session processes inputs: each input message includes playerID, apply to corresponding ship
- Snapshot broadcasts: session.GetWorld() → convert to protocol → broadcast to all room.Players[]

**Input Processing**:
- Each input message includes playerID (identifies which ship to update)
- Session command queue processes inputs in sequence order
- Input applied to ship with matching ID in World.Ships[]

**Snapshot Broadcasting**:
- Session generates snapshots at 10–15 Hz
- Snapshot converted to protocol format (World → SnapshotMessage)
- Snapshot broadcast to all players in room (via their WebSocket connections)

**Invariants**:
- One session per room (not per connection)
- Session tick loop runs once per room (processes all player inputs in single World state)
- All players in room receive same snapshots (shared game state)

---

## Concurrency Model

**Mutex Usage**:
- RoomManager uses sync.RWMutex for rooms map access
  - Read lock: GetRoom, room queries
  - Write lock: CreateRoom, JoinRoom, LeaveRoom, StartMatch
- Room uses sync.RWMutex for player list and state access
  - Read lock: Get players, get state
  - Write lock: Add/remove players, change state

**Goroutines**:
- Each room's session runs in dedicated goroutine (session.Run())
- Snapshot broadcasting runs in session handler goroutine
- Room cleanup runs in background (periodic or on-demand)

**Invariants**:
- All room operations are thread-safe (mutex-protected)
- Session goroutines are isolated (one per room)
- No race conditions in room state access

---

## Constants

**Room Constants**:
- `MIN_PLAYERS = 2` – Minimum players required to start match
- `MAX_PLAYERS = 8` – Maximum players per room
- `ROOM_CODE_LENGTH = 6` – Room code length (characters)
- `ROOM_CODE_CHARSET = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"` – Room code character set (36 chars)
- `MAX_CODE_GENERATION_RETRIES = 10` – Maximum retries for room code generation

---

## Ownership & Dependencies

### Room Package Ownership

- **Only `server/internal/room` may define room management logic**
- Room manages room lifecycle and player coordination
- Room creates and manages sessions (but does not implement session logic)

### Dependencies

- **Imports**:
  - `session` package (for Session type)
  - `transport` package (for Connection type)
  - `entities` package (for World type, planet generation)
  - `sim` package (for initial world creation)
- **No dependencies on**: proto, physics, rules packages (room is orchestration layer)

### No Duplication Rules

- **No room management elsewhere**: Room lifecycle must live in `/room`
- **No room code generation elsewhere**: Room code generation must live in `/room`
- **Room does not implement**: Physics, rules, or session orchestration (room coordinates sessions)

---

## Notes

This spec describes the v1 room management implementation. Key features:
- Room-based matchmaking (create/join by code)
- One session per room (shared by all players)
- In-memory room state (no persistence)
- Room code generation with collision detection
- Thread-safe room operations

Future extensions may include:
- Room persistence (database storage)
- Room discovery (list public rooms)
- Room settings (private/public, password protection)
- Spectator mode
- Room analytics

