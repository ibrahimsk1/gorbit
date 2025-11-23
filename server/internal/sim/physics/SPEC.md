# Orbital Rush – Physics Subsystem Specification

This document describes the physics calculations and mechanics for Orbital Rush. It defines the formulas, algorithms, and invariants for all physics operations.

---

## Scope & Location

**Scope**: Physics calculations for Orbital Rush (gravity, integration, collisions).

**Code location**: `server/internal/sim/physics`

**Design Goals**:
- Pure, deterministic physics calculations (no IO, no side effects)
- All physics operates on canonical entity types from `entities` package
- Formulas are mathematically well-defined and testable
- Physics is independent of game rules (rules layer composes physics)

---

## Core Physics Operations

### Gravity

**File**: `server/internal/sim/physics/gravity.go`

**Concept**: Inverse-square law gravity from multiple planets with maximum acceleration clamping. Gravity from all planets is summed (superposition principle).

**Single Planet Gravity Formula**:
- Direction: `direction = planetPos - shipPos`
- Distance squared: `distanceSq = |direction|²`
- Acceleration magnitude: `|a| = G * M / distanceSq` (inverse-square law)
- Clamped magnitude: `|a|_clamped = min(|a|, aMax)`
- Final acceleration: `acc = normalize(direction) * |a|_clamped`

**Multiple Planets Gravity (Summation)**:
- Initialize: `totalAcc = Vec2{0, 0}`
- For each planet in planets array:
  - Calculate planet acceleration: `planetAcc = GravityAcceleration(shipPos, planet.Pos, planet.Mass, G, aMax)`
  - Sum: `totalAcc = totalAcc.Add(planetAcc)`
- Return: `totalAcc` (sum of all planet gravities)

**Function Signature**: `CalculateTotalGravity(shipPos Vec2, planets []Planet, G, aMax float64) Vec2`

**Semantics**:
- Multiple gravity sources (3–5 planets per match)
- Gravity from each planet calculated independently using inverse-square law
- Total gravity is vector sum of all planet gravities (superposition principle)
- Each planet's gravity is clamped individually at `aMax` before summation
- Zero mass or zero distance returns zero acceleration (no division by zero)

**Parameters**:
- `G` (float64): Gravitational constant (game-scale, 1.0 in v1)
- `aMax` (float64): Maximum acceleration magnitude (m/s², 100.0 in v1)

**Invariants**:
- Each planet's acceleration magnitude never exceeds `aMax` (clamped before summation)
- Acceleration vector from each planet points toward that planet (attractive force)
- Total acceleration is vector sum of all planet accelerations
- Zero mass or zero distance produces zero acceleration
- All calculations use finite float64 values

**Note**: Gravity summation applies to each ship independently. In multiplayer, each ship's gravity is calculated from all planets.

---

### Integration

**File**: `server/internal/sim/physics/integrator.go`

**Concept**: Semi-implicit Euler (symplectic Euler) integration for position and velocity.

**Algorithm**:
1. Update velocity: `v_new = v_old + a * dt`
2. Update position: `p_new = p_old + v_new * dt`

**Semantics**:
- Symplectic method that better conserves energy than explicit Euler
- Suitable for physics simulations with constant or slowly-varying acceleration
- Velocity is updated first, then position uses the new velocity
- Time step `dt` is constant (1/30 seconds for 30 Hz tick rate)

**Parameters**:
- `pos` (Vec2): Current position
- `vel` (Vec2): Current velocity
- `acc` (Vec2): Acceleration (constant for this step)
- `dt` (float64): Time step in seconds

**Invariants**:
- Inputs are finite Vec2 values
- Outputs are finite Vec2 values
- Integration is deterministic (same inputs produce same outputs)

---

### Collisions

**File**: `server/internal/sim/physics/collision.go`

**Concept**: Distance-based collision detection for ship-planet and ship-pallet interactions.

#### Ship-Planet Collision

**Semantics**:
- Collision occurs when distance from ship center to planet center ≤ planet radius
- Uses squared distance comparison to avoid square root
- Formula: `distanceSq ≤ radius²`
- In multiplayer, check each ship against all planets

**Parameters**:
- `shipPos` (Vec2): Ship position
- `planetPos` (Vec2): Planet center position
- `planetRadius` (float32): Planet radius

**Returns**: `true` if colliding, `false` otherwise

**Function Signature**: `ShipPlanetCollision(shipPos Vec2, planetPos Vec2, planetRadius float32) bool`

**Invariants**:
- Collision is symmetric (ship collides with planet = planet collides with ship)
- Zero-radius planet never collides (unless ship exactly at planet center)
- All positions are finite Vec2 values

**Note**: In multiplayer, collision detection loops over all ships and all planets.

#### Ship-Pallet Collision

**Semantics**:
- Collision occurs when distance from ship center to pallet center ≤ pickup radius
- Uses squared distance comparison
- Formula: `distanceSq ≤ pickupRadius²`
- Only active pallets are checked (inactive pallets are ignored)

**Parameters**:
- `shipPos` (Vec2): Ship position
- `palletPos` (Vec2): Pallet center position
- `pickupRadius` (float64): Pickup radius (typically 15.0 m)

**Returns**: `true` if colliding, `false` otherwise

**Invariants**:
- Pickup radius is typically larger than ship radius for gameplay feel
- Zero pickup radius never collides (unless ship exactly at pallet center)
- All positions are finite Vec2 values

### World Wraparound

**File**: `server/internal/sim/physics/wraparound.go` (or in integrator.go)

**Concept**: Wraparound logic that teleports ships exiting world bounds to opposite side.

**Formula**:
- For X coordinate: `pos.x = mod(pos.x + WORLD_WIDTH/2, WORLD_WIDTH) - WORLD_WIDTH/2`
- For Y coordinate: `pos.y = mod(pos.y + WORLD_HEIGHT/2, WORLD_HEIGHT) - WORLD_HEIGHT/2`
- Uses `math.Mod` for modulo operation

**Semantics**:
- Applied after SemiImplicitEuler integration (after position update)
- Applied before collision detection
- In multiplayer, applied to all ships in World.Ships array
- Wraparound is deterministic and applied consistently every tick
- World bounds: [-WORLD_WIDTH/2, WORLD_WIDTH/2] × [-WORLD_HEIGHT/2, WORLD_HEIGHT/2]
- World constants: `WORLD_WIDTH = 2000.0 m`, `WORLD_HEIGHT = 2000.0 m`

**Function Signature**: `ApplyWraparound(pos Vec2, worldWidth, worldHeight float64) Vec2`

**Invariants**:
- Wraparound is deterministic (same input → same output)
- Output position is always within world bounds
- Applied to all ships every tick (after integration, before collision)
- Velocity is not modified by wraparound (only position)

**Note**: Wraparound applies to ship positions only. Camera and rendering may handle wraparound differently (camera stays within bounds, no wraparound).

---

## Constants

**Standardized Physics Constants** (from TDD v1):
- `G = 1.0` – Gravitational constant (game-scale)
- `A_MAX = 100.0` – Maximum acceleration (m/s²)
- `DRAG_K = 0.12` – Linear drag coefficient (s⁻¹)
- `SHIP_RADIUS = 1.0` – Ship collision radius (m)
- `PICKUP_RADIUS = 15.0` – Pallet pickup radius (m)
- `WORLD_WIDTH = 2000.0` – World width (m)
- `WORLD_HEIGHT = 2000.0` – World height (m)
- `TICK_RATE = 30.0` – Simulation tick rate (Hz)
- `DT = 1.0 / TICK_RATE` – Time step (seconds, ~0.0333)

---

## Ownership & Dependencies

### Physics Package Ownership

- **Only `server/internal/sim/physics` may define physics formulas and calculations**
- Physics functions are pure (no side effects, deterministic)
- Physics operates on entity types from `entities` package (does not define its own state types)
- Higher layers (rules, session) call physics functions but do not re-implement physics formulas

### Dependencies

- **Imports**: `entities` package (for Vec2, Ship, Sun, Pallet)
- **No dependencies on**: rules, session, proto, transport packages
- Physics is the lowest-level simulation layer (G1)

### No Duplication Rules

- **No physics formulas elsewhere**: Gravity, integration, collision calculations must live in `/sim/physics`
- **No parallel physics types**: Physics uses entity types, does not define its own position/velocity types
- **Determinism requirement**: All physics functions must be deterministic (same inputs → same outputs)

---

## Notes

This spec describes the v1 physics implementation. Key characteristics:
- Multiple Planets gravity (3–5 planets, gravity summation via superposition)
- World wraparound (ships exiting bounds re-enter from opposite side)
- Semi-implicit Euler integration
- Ship-planet and ship-pallet collision detection
- Gravity and wraparound applied to all ships in multiplayer

