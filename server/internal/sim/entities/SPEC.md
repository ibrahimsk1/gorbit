# Orbital Rush – Simulation Model Specification

This document describes the canonical simulation model (entities) for Orbital Rush. It defines the semantic meaning, invariants, and ownership rules for all simulation state types.

---

## Scope & Location

**Scope**: Canonical simulation state for Orbital Rush (single ship, single sun, pallets).

**Code location**: `server/internal/sim/entities`

**Design Goals**:
- Single-player world model
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

### Sun

**Concept**: Single central gravity source and collision obstacle.

**Key Fields**:
- `Pos Vec2` – Position in world coordinates (meters), typically at origin (0, 0)
- `Radius float32` – Collision radius (meters)
- `Mass float64` – Mass for gravity calculations (game units)

**Semantics**:
- Single sun per match (static, position does not change)
- Used by gravity system (single gravity source)
- Used by collision detection (ship-sun collisions)
- Typically centered at world origin

**Invariants**:
- Radius > 0
- Mass > 0
- Pos is finite Vec2

**Ownership**: Only `server/internal/sim/entities` defines Sun. No parallel sun/gravity-source types elsewhere.

---

### Ship

**File**: `server/internal/sim/entities/ship.go`

**Concept**: Single player-controlled ship in the simulation.

**Key Fields**:
- `Pos Vec2` – Position in world coordinates (meters)
- `Vel Vec2` – Velocity vector (m/s)
- `Rot float64` – Rotation angle in radians
- `Energy float32` – Current energy level (0-100)

**Semantics**:
- Single ship per match (single-player)
- No player ID or ship ID needed (only one ship exists)

**Invariants**:
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

**Concept**: Complete simulation state for a match.

**Key Fields**:
- `Ship Ship` – Single ship in the match (not an array)
- `Sun Sun` – Single sun/gravity source (not an array)
- `Pallets []Pallet` – All pallets in the match
- `Tick uint32` – Current simulation tick
- `Done bool` – Whether match has ended
- `Win bool` – Whether match ended in victory (only valid if Done is true)

**Semantics**:
- All sim state for a match is inside World
- World is the root container passed to physics and rules systems
- Single ship and single sun (single-player, single gravity source)
- World bounds may be defined as constants (not a WorldBounds type)

**Invariants**:
- Pallet IDs are unique within Pallets array
- All entity positions are finite Vec2 values
- Tick increments monotonically during simulation

**Ownership**: Only `server/internal/sim/entities` defines World. No parallel world/state structs in session, proto, or client packages.

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

This spec describes the current entity model. Key characteristics:
- Single Ship (not an array)
- Single Sun (not Planet, not an array)
- World bounds may be defined as constants (not a WorldBounds type)
- World contains single instances, not arrays (except Pallets)

