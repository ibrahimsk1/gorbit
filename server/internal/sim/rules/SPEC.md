# Orbital Rush – Rules Subsystem Specification

This document describes the game rules, state transitions, and gameplay mechanics for Orbital Rush. It defines how player input is processed, how game state evolves, and when win/lose conditions are triggered.

---

## Scope & Location

**Scope**: Game rules and state transitions for Orbital Rush (input processing, energy economy, win/lose conditions).

**Code location**: `server/internal/sim/rules`

**Design Goals**:
- Pure, deterministic rule logic (no IO, no side effects)
- Rules compose physics operations but do not re-implement physics
- All rules operate on canonical entity types from `entities` package
- Rules define gameplay behavior, not physics behavior

---

## Core Rules Operations

### Input Processing

**File**: `server/internal/sim/rules/input.go`

**Concept**: Convert player input commands into ship state changes (rotation, velocity, energy).

#### Input Command

**Type**: `InputCommand`
- `Thrust float32` – Thrust input, clamped to [0.0, 1.0]
- `Turn float32` – Turn input, clamped to [-1.0, 1.0]
  - Positive values turn right (counter-clockwise)
  - Negative values turn left (clockwise)

#### Input Processing Steps

1. **Clamp Input**: Ensure Thrust ∈ [0.0, 1.0] and Turn ∈ [-1.0, 1.0]
2. **Update Rotation**: `newRot = currentRot + TurnRate * turnInput * dt`
   - Rotation normalized to [0, 2π) range
   - `TurnRate = 3.0` rad/s per unit turn input
3. **Calculate Thrust Acceleration**: `thrustAcc = ThrustAcceleration * thrustInput * direction(rotation)`
   - `ThrustAcceleration = 20.0` m/s² per unit thrust
   - Direction derived from rotation angle (cos/sin)
   - Y component negated to match screen coordinate system
4. **Apply Thrust to Velocity**: `newVel = oldVel + thrustAcc * dt`
   - Only applied if `energy > MinEnergyForThrust` (typically 0.0)
   - Thrust requires energy to be available
5. **Drain Energy**: If thrusting, drain energy by `ThrustDrainRate` per tick

**Semantics**:
- Rotation always works (no energy required)
- Thrust only works when energy > 0
- Input processing does not update position (position updated by physics integration)
- Input is applied before physics step in the game loop

**Invariants**:
- Rotation is always in [0, 2π) range after normalization
- Thrust acceleration magnitude is ≤ ThrustAcceleration
- Energy cannot go below 0
- Velocity changes are finite Vec2 values

---

### Energy Economy

**File**: `server/internal/sim/rules/energy.go`

**Concept**: Energy management system that drains on thrust and restores on pallet pickup.

#### Energy Constants

- `MaxEnergy = 100.0` – Maximum energy value (energy bar cap)
- `ThrustDrainRate = 0.5` – Energy drained per tick when thrusting
- `PalletRestoreAmount = 25.0` – Energy restored per pallet pickup

#### Energy Operations

**Drain Energy on Thrust**:
- If `isThrusting == true`: `newEnergy = currentEnergy - ThrustDrainRate`
- Energy clamped to [0, MaxEnergy] after draining
- If `isThrusting == false`: energy unchanged

**Restore Energy on Pickup**:
- When pallet is collected: `newEnergy = currentEnergy + PalletRestoreAmount`
- Energy clamped to [0, MaxEnergy] after restoring
- Pallet is deactivated (Active = false) when collected

**Clamp Energy**:
- Energy is always clamped to [0, MaxEnergy]
- Negative energy becomes 0
- Energy above MaxEnergy becomes MaxEnergy

**Semantics**:
- Energy is a finite resource that limits thrust capability
- Players must collect pallets to maintain energy
- Energy economy creates strategic gameplay (when to thrust, when to conserve)

**Invariants**:
- Energy is always in [0, MaxEnergy] range
- Energy changes are deterministic (same inputs → same outputs)
- Thrust requires energy > 0 (enforced by input processing)

---

### Game State Evaluation

**File**: `server/internal/sim/rules/game.go`

**Concept**: Win/lose condition checking and game state transitions.

#### Win Condition

**Check**: All pallets are collected (all Pallets have `Active == false`)

**Semantics**:
- If there are no pallets configured, win condition cannot trigger
- Win condition is checked first (takes precedence over lose condition)
- When win condition is met: `Done = true`, `Win = true`

**Invariants**:
- Win condition is idempotent (checking multiple times produces same result)
- Empty pallet list means win condition cannot trigger

#### Lose Condition

**Check**: Ship collides with sun (using `ShipSunCollision` from physics)

**Semantics**:
- Collision detection uses physics collision function
- Lose condition is checked after win condition
- When lose condition is met: `Done = true`, `Win = false`

**Invariants**:
- Lose condition is idempotent
- Collision detection is deterministic (uses physics layer)

#### State Evaluation

**Function**: `EvaluateGameState(world)`

**Algorithm**:
1. If `world.Done == true`: return unchanged (idempotent)
2. Check win condition: if true, set `Done = true`, `Win = true`, return
3. Check lose condition: if true, set `Done = true`, `Win = false`, return
4. Otherwise: return unchanged (game continues)

**Semantics**:
- Win condition takes precedence (if both true, win is set)
- Once `Done == true`, state should not change (idempotent evaluation)
- Other world fields (Ship, Sun, Pallets, Tick) are not modified by evaluation
- Evaluation is called after physics and collision processing

**Invariants**:
- Evaluation is idempotent (multiple calls with same state produce same result)
- `Win` is only meaningful when `Done == true`
- State evaluation does not modify entity positions or velocities

---

### Game Loop Step

**File**: `server/internal/sim/rules/step.go`

**Concept**: Complete game loop step that orchestrates input, physics, collisions, and rules.

**Algorithm** (if `world.Done == false`):
1. **Apply Input**: Process player input (thrust, turn) → updates rotation, velocity, energy
2. **Update Physics**: Calculate gravity acceleration, integrate position and velocity
3. **Process Collisions**: Check for pallet pickups → deactivate pallet, restore energy
4. **Evaluate Rules**: Check win/lose conditions → update Done/Win flags
5. **Update State**: Increment tick counter

**If `world.Done == true`**:
- Skip all processing, only increment tick counter

**Semantics**:
- Step function orchestrates the complete game loop
- Order matters: input → physics → collisions → rules → state
- Physics operations are called from rules layer (rules composes physics)
- Step is deterministic (same inputs → same outputs)

**Parameters**:
- `world` (World): Current world state
- `input` (InputCommand): Player input command
- `dt` (float64): Time step in seconds
- `G` (float64): Gravitational constant
- `aMax` (float64): Maximum acceleration
- `pickupRadius` (float64): Pallet pickup radius

**Invariants**:
- Step is deterministic
- If `world.Done == true`, only tick is incremented
- All physics operations use canonical physics functions
- State transitions are well-defined and testable

---

## Constants

**Standardized Rules Constants** (from code):
- `ThrustAcceleration = 20.0` – Acceleration magnitude per unit thrust (m/s²)
- `TurnRate = 3.0` – Angular velocity per unit turn input (rad/s)
- `MinEnergyForThrust = 0.0` – Minimum energy required to thrust
- `MaxEnergy = 100.0` – Maximum energy value
- `ThrustDrainRate = 0.5` – Energy drained per tick when thrusting
- `PalletRestoreAmount = 25.0` – Energy restored per pallet pickup

---

## Ownership & Dependencies

### Rules Package Ownership

- **Only `server/internal/sim/rules` may define game rule logic and state transitions**
- Rules functions are pure (no side effects, deterministic)
- Rules operate on entity types from `entities` package
- Rules compose physics operations but do not re-implement physics formulas

### Dependencies

- **Imports**: 
  - `entities` package (for World, Ship, Sun, Pallet, Vec2)
  - `physics` package (for gravity, integration, collision functions)
- **No dependencies on**: session, proto, transport packages
- Rules is the game logic layer that composes physics

### No Duplication Rules

- **No rule logic elsewhere**: Win/lose conditions, energy economy, input processing must live in `/sim/rules`
- **No parallel rule types**: Rules uses entity types, does not define its own game state types
- **Determinism requirement**: All rules functions must be deterministic
- **Physics composition**: Rules calls physics functions, does not re-implement physics

---

## Notes

This spec describes the current rules implementation. Key features:
- Input processing with rotation and thrust
- Energy economy (drain on thrust, restore on pickup)
- Win condition: collect all pallets
- Lose condition: collide with sun
- Deterministic game loop step

Future extensions may include:
- Multiple players (each with their own ship and input)
- Abilities or special actions
- More complex win/lose conditions
- Power-ups or temporary effects

