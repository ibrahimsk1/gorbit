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

**Concept**: Inverse-square law gravity from a single sun with maximum acceleration clamping.

**Formula**:
- Direction: `direction = sunPos - shipPos`
- Distance squared: `distanceSq = |direction|²`
- Acceleration magnitude: `|a| = G * M / distanceSq` (inverse-square law)
- Clamped magnitude: `|a|_clamped = min(|a|, aMax)`
- Final acceleration: `acc = normalize(direction) * |a|_clamped`

**Semantics**:
- Single gravity source (the Sun)
- Gravity computed from sun position to ship position
- Zero mass or zero distance returns zero acceleration (no division by zero)

**Parameters**:
- `G` (float64): Gravitational constant (game-scale, typically 1.0)
- `aMax` (float64): Maximum acceleration magnitude (m/s², typically 100.0)

**Invariants**:
- Acceleration magnitude never exceeds `aMax`
- Acceleration vector points toward the sun (attractive force)
- Zero mass or zero distance produces zero acceleration
- All calculations use finite float64 values

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

**Concept**: Distance-based collision detection for ship-sun and ship-pallet interactions.

#### Ship-Sun Collision

**Semantics**:
- Collision occurs when distance from ship center to sun center ≤ sun radius
- Uses squared distance comparison to avoid square root
- Formula: `distanceSq ≤ radius²`

**Parameters**:
- `shipPos` (Vec2): Ship position
- `sunPos` (Vec2): Sun center position
- `sunRadius` (float32): Sun radius

**Returns**: `true` if colliding, `false` otherwise

**Invariants**:
- Collision is symmetric (ship collides with sun = sun collides with ship)
- Zero-radius sun never collides (unless ship exactly at sun center)
- All positions are finite Vec2 values

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

---

## Constants

**Standardized Physics Constants** (from TDD):
- `G = 1.0` – Gravitational constant (game-scale)
- `A_MAX = 100.0` – Maximum acceleration (m/s²)
- `DRAG_K = 0.12` – Linear drag coefficient (s⁻¹) – *Note: may not be implemented*
- `SHIP_RADIUS = 1.0` – Ship collision radius (m)
- `PICKUP_RADIUS = 15.0` – Pallet pickup radius (m)
- `WORLD_WIDTH = <value>` – World width (m) – *May be defined as constant, not a type*
- `WORLD_HEIGHT = <value>` – World height (m) – *May be defined as constant, not a type*
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

This spec describes the current physics implementation. Key characteristics:
- Single Sun gravity (not multiple planets)
- No wraparound logic (ships may exit world bounds or be handled differently)
- Semi-implicit Euler integration
- Ship-sun and ship-pallet collision detection

