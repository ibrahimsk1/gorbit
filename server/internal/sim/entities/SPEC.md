# Orbital Rush – Simulation Model Specification

This document describes the canonical simulation model (entities) for Orbital Rush. It defines the semantic meaning, invariants, and ownership rules for all simulation state types.

---

## Scope & Location

**Scope**: Canonical simulation state for Orbital Rush (multiple ships, multiple planets, pallets).

**Code location**: `server/internal/sim/entities`

**Design Goals**:
- Multiplayer world model (2–8 players per room)
- All simulation state lives under `entities`; no ad-hoc world/state structs in other packages
- Entities are mostly data; behavior lives in physics/rules

---

## Core Types

### Vec2

**File**: `server/internal/sim/entities/vec2.go`

**Concept**: 2D vector with X, Y float64 components for positions and velocities.

**Semantics**:
- All simulation positions and velocities use this type
- No other plain x/y structs allowed in sim code
- Provides standard vector operations: Add, Sub, Scale, Dot, Length, Normalize

**Invariants**:
- X and Y are finite float64 values (no NaN, no Inf in valid state)

---

### Planet

**Concept**: Gravity source and collision obstacle (replaces Sun in v1).

**Key Fields**:
- `ID uint32` – Unique planet identifier
- `Pos Vec2` – Position in world coordinates (meters)
- `Radius float32` – Collision radius (meters, typically 30–80 m)
- `Mass float64` – Mass for gravity calculations (game units, typically 500–2000)

**Semantics**:
- Multiple planets per match (3–5 planets, static positions)
- Each planet provides gravity (inverse-square law, summed for all planets)
- Used by collision detection (ship-planet collisions)
- Planets distributed across world with minimum spacing (200 m apart)

**Invariants**:
- Radius > 0 (typically 30–80 m)
- Mass > 0 (typically 500–2000 game units)
- Pos is finite Vec2
- ID uniqueness within a World
- Minimum distance between planets: 200 m (enforced during generation)

**Ownership**: Only `server/internal/sim/entities` defines Planet. No parallel planet/gravity-source types elsewhere.

---

### Ship

**File**: `server/internal/sim/entities/ship.go`

**Concept**: Player-controlled ship in the simulation (multiple ships per match in multiplayer).

**Key Fields**:
- `ID uint32` – Unique ship identifier (player ID)
- `Pos Vec2` – Position in world coordinates (meters)
- `Vel Vec2` – Velocity vector (m/s)
- `Rot float64` – Rotation angle in radians
- `Energy float32` – Current energy level (0-100)

**Semantics**:
- Multiple ships per match (2–8 ships, one per player)
- Ship ID corresponds to player ID (unique per room)
- Each ship has independent state (position, velocity, rotation, energy)

**Invariants**:
- ID uniqueness within a World
- Energy >= 0 (typically clamped to [0, MAX_ENERGY])
- Pos and Vel are finite Vec2 values
- Rot is in radians (typically normalized to [0, 2π) or [-π, π])

**Ownership**: Only `server/internal/sim/entities` defines Ship. Other packages import and operate on Ship instances but may not define parallel ship types for sim state.

---

### Pallet

**File**: `server/internal/sim/entities/world.go` (or separate `pallet.go`)

**Concept**: Collectible energy pallet that restores ship energy.

**Key Fields**:
- `ID uint32` – Unique pallet identifier
- `Pos Vec2` – Position in world coordinates (meters)
- `Active bool` – Whether pallet is collectible (false after collection)

**Semantics**:
- Collect-once: when the ship picks up a pallet, Active becomes false
- Position remains in world for rendering/debugging, but inactive pallets are ignored by collision detection
- Typically 8-12 pallets per match

**Invariants**:
- ID uniqueness within a World
- Pos is finite Vec2
- Active pallets must have valid positions within or near world bounds

**Ownership**: Only `server/internal/sim/entities` defines Pallet.

---

### World

**File**: `server/internal/sim/entities/world.go`

**Concept**: Complete simulation state for a match (multiplayer).

**Key Fields**:
- `Ships []Ship` – All ships in the match (2–8 ships, one per player)
- `Planets []Planet` – All planets in the match (3–5 planets)
- `Pallets []Pallet` – All pallets in the match (8–12 pallets)
- `Tick uint32` – Current simulation tick
- `Done bool` – Whether match has ended (per-player or global)
- `Win bool` – Whether match ended in victory (per-player or global, only valid if Done is true)

**Semantics**:
- All sim state for a match is inside World
- World is the root container passed to physics and rules systems
- Multiple ships (one per player) and multiple planets (3–5 per match)
- World bounds defined as constants: `WORLD_WIDTH = 2000.0 m`, `WORLD_HEIGHT = 2000.0 m`
- World center at origin (0, 0), bounds from [-1000, 1000] × [-1000, 1000]

**Invariants**:
- Ship IDs are unique within Ships array
- Planet IDs are unique within Planets array
- Pallet IDs are unique within Pallets array
- All entity positions are finite Vec2 values
- Tick increments monotonically during simulation
- Minimum distance between planets: 200 m (enforced during generation)

**Ownership**: Only `server/internal/sim/entities` defines World. No parallel world/state structs in session, proto, or client packages.

### Planet Generation

**File**: `server/internal/sim/entities/world.go` (or `planet.go`)

**Concept**: Generate multiple planets with varied sizes, masses, and positions for a match.

**Algorithm**:
1. **Function**: `GeneratePlanets(count int, worldWidth, worldHeight float64) []Planet`
2. **For each planet** (count = 3–5):
   - Random radius: `radius = 30.0 + rand.Float64() * 50.0` (range [30, 80] m)
   - Random mass: `mass = 500.0 + rand.Float64() * 1500.0` (range [500, 2000] game units)
   - Random position: 
     - `posX = (rand.Float64() - 0.5) * worldWidth` (bounds [-WORLD_WIDTH/2, WORLD_WIDTH/2])
     - `posY = (rand.Float64() - 0.5) * worldHeight` (bounds [-WORLD_HEIGHT/2, WORLD_HEIGHT/2])
   - Check minimum distance: for each existing planet, if `distance(newPos, existingPos) < 200.0`, retry position (max 50 retries per planet)
   - Check overlap: ensure `distance >= newRadius + existingRadius` (no overlap)
3. **Assign IDs**: Assign unique ID to each planet (incrementing counter or UUID)

**Semantics**:
- Planets are generated once per match (during room match start)
- Planet positions are distributed across world with minimum spacing
- Planet sizes and masses are varied for gameplay interest
- Generation uses crypto/rand for secure randomness (or math/rand for deterministic testing)

**Parameters**:
- `count` (int): Number of planets to generate (3–5 per match)
- `worldWidth` (float64): World width (2000.0 m)
- `worldHeight` (float64): World height (2000.0 m)

**Invariants**:
- Planet count is between 3 and 5
- All planets have radius in [30, 80] m range
- All planets have mass in [500, 2000] game units range
- Minimum distance between planets: 200 m
- No overlapping planets (distance >= sum of radii)
- All planet positions are within world bounds

**Note**: Planet generation is called when room match starts, before creating initial World state.

---

## Ownership & Dependencies

### Entity Package Ownership

- **Only `server/internal/sim/entities` may define these types**
- Other packages can import and operate on them, but may not define parallel types for sim state
- Protocol layer (`/proto`) defines snapshot/transport types that mirror entities, but these are separate and used only for serialization
- Client may have TypeScript types that mirror entities, but these are for rendering/state management, not authoritative sim state

### Dependencies

- Entities package has no dependencies on other sim packages (physics, rules)
- Physics and rules packages import entities and operate on them
- Session/orchestration packages import entities to manage World state
- Protocol packages convert entities to/from wire formats

### No Duplication Rules

- **No parallel World structs**: Session or transport layers must use `entities.World`, not define their own world/state types
- **No parallel entity types**: Ship, Sun, Pallet, Vec2 are defined once in entities package
- **No ad-hoc state**: All game state that affects simulation must live in entities

---

## Notes

This spec describes the v1 entity model. Key characteristics:
- Multiple Ships (array, 2–8 ships per match)
- Multiple Planets (array, 3–5 planets per match, replaces single Sun)
- World bounds defined as constants: `WORLD_WIDTH = 2000.0 m`, `WORLD_HEIGHT = 2000.0 m`
- World contains arrays for Ships, Planets, and Pallets
- Ship.ID and Planet.ID added for multiplayer support

