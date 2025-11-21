# Orbital Rush v0 – Simulation Package API

This document provides detailed API reference for the simulation package.

---

## Scope & Location

**Scope**: Client-side simulation system (state management, prediction, reconciliation, interpolation, local simulation).

**Code location**: `client/src/sim`

**Design Goals**:
- Manage authoritative, predicted, and interpolated game states
- Provide client-side prediction for responsive gameplay
- Reconcile predicted state with server snapshots
- Interpolate between snapshots for smooth rendering
- Run local simulation matching server physics/rules

---

## StateManager

**File**: `state-manager.ts`

**Concept**: Coordinates authoritative, predicted, and interpolated states for client-side prediction and smooth rendering.

### Interface

```typescript
export interface GameState {
  tick: number
  ship: ShipSnapshot
  planets: PlanetSnapshot[]
  pallets: PalletSnapshot[]
  done: boolean
  win: boolean
}

export class StateManager {
  constructor()
  updateAuthoritative(snapshot: SnapshotMessage): void
  updatePredicted(state: GameState): void
  clearPredicted(): void
  updateInterpolated(state: GameState): void
  getAuthoritative(): GameState | null
  getPredicted(): GameState | null
  getInterpolated(): GameState | null
  getRenderState(): GameState
  reset(): void
  hasAuthoritative(): boolean
  hasPredicted(): boolean
  hasInterpolated(): boolean
}
```

### Methods

#### `constructor()`

Creates a new StateManager instance.

**Returns:** StateManager instance

---

#### `updateAuthoritative(snapshot: SnapshotMessage): void`

Updates authoritative state from server snapshot.

**Parameters:**
- `snapshot: SnapshotMessage` - Server snapshot

---

#### `updatePredicted(state: GameState): void`

Updates predicted state from local simulation.

**Parameters:**
- `state: GameState` - Predicted game state

---

#### `clearPredicted(): void`

Clears predicted state (used during reconciliation rollback).

---

#### `updateInterpolated(state: GameState): void`

Updates interpolated state for smooth rendering.

**Parameters:**
- `state: GameState` - Interpolated game state

---

#### `getAuthoritative(): GameState | null`

Gets current authoritative state.

**Returns:** Authoritative state or null

---

#### `getPredicted(): GameState | null`

Gets current predicted state.

**Returns:** Predicted state or null

---

#### `getInterpolated(): GameState | null`

Gets current interpolated state.

**Returns:** Interpolated state or null

---

#### `getRenderState(): GameState`

Gets state for rendering. Prefers interpolated, falls back to authoritative.

**Returns:** GameState for rendering (never null, returns empty state if needed)

---

#### `reset(): void`

Resets all states to empty.

---

#### `hasAuthoritative(): boolean`

Checks if authoritative state exists.

---

#### `hasPredicted(): boolean`

Checks if predicted state exists.

---

#### `hasInterpolated(): boolean`

Checks if interpolated state exists.

### Semantics

- StateManager maintains three separate state layers:
  - **Authoritative**: Ground truth from server (never modified locally)
  - **Predicted**: Local simulation result (may be rolled back)
  - **Interpolated**: Smoothed state for rendering (derived from authoritative)
- States are deep cloned when stored
- Render state prefers interpolated, falls back to authoritative
- Predicted state can be cleared during reconciliation rollback

### Lifecycle

1. **Creation**: `new StateManager()` – creates empty state manager
2. **Authoritative Updates**: `updateAuthoritative()` – receives server snapshots
3. **Prediction**: `updatePredicted()` – stores local simulation results
4. **Interpolation**: `updateInterpolated()` – stores smoothed state
5. **Rendering**: `getRenderState()` – provides state for rendering
6. **Reset**: `reset()` – clears all states

### Invariants

- Authoritative state is never modified locally (only updated from server)
- Predicted state is independent of authoritative state
- Interpolated state is derived from authoritative state
- States are always deep cloned (no shared references)
- Render state never returns null (returns empty state if needed)

---

## LocalSimulator

**File**: `local-simulator.ts`

**Concept**: Runs deterministic physics/rules matching server behavior for client-side prediction.

### Interface

```typescript
export class LocalSimulator {
  constructor(dt?: number, G?: number, aMax?: number, pickupRadius?: number)
  step(state: GameState, input: { thrust: number, turn: number }): GameState
}
```

### Methods

#### `constructor(dt?: number, G?: number, aMax?: number, pickupRadius?: number)`

Creates a new LocalSimulator with physics constants.

**Parameters:**
- `dt?: number` - Time step (default: 1/30)
- `G?: number` - Gravitational constant (default: 1.0)
- `aMax?: number` - Maximum acceleration (default: 100.0)
- `pickupRadius?: number` - Pallet pickup radius (default: 15.0)

---

#### `step(state: GameState, input: { thrust: number, turn: number }): GameState`

Performs one simulation step with the given input.

**Parameters:**
- `state: GameState` - Current game state
- `input: { thrust: number, turn: number }` - Input command

**Returns:** New game state after simulation step

### Semantics

- LocalSimulator matches server physics/rules exactly
- Simulation is deterministic (same input → same output)
- Processes: input application, gravity, physics integration, collisions, rules
- Constants match server values (30Hz tick rate, same physics parameters)

### Invariants

- Simulation is deterministic (pure function)
- Output state is always valid (all fields present)
- Physics constants match server values
- Step increments tick by 1

---

## PredictionSystem

**File**: `prediction.ts`

**Concept**: Runs local simulation immediately when input is sent for responsive gameplay.

### Interface

```typescript
export class PredictionSystem {
  constructor(
    stateManager: StateManager,
    localSimulator: LocalSimulator,
    commandHistory: CommandHistory
  )
  predict(input: { thrust: number, turn: number }): void
  getPredictedState(): GameState | null
  hasPredictedState(): boolean
  reset(): void
}
```

### Methods

#### `constructor(stateManager, localSimulator, commandHistory)`

Creates a new PredictionSystem.

**Parameters:**
- `stateManager: StateManager` - State manager
- `localSimulator: LocalSimulator` - Local simulator
- `commandHistory: CommandHistory` - Command history

---

#### `predict(input: { thrust: number, turn: number }): void`

Runs prediction for a new input command.

**Parameters:**
- `input: { thrust: number, turn: number }` - Input command

---

#### `getPredictedState(): GameState | null`

Gets the current predicted state.

**Returns:** Predicted state or null

---

#### `hasPredictedState(): boolean`

Checks if predicted state exists.

---

#### `reset(): void`

Resets prediction state by clearing predicted state.

### Semantics

- Prediction runs immediately when input is sent (before server confirmation)
- Uses predicted state as base for chaining (allows multiple predictions)
- Falls back to authoritative state if no predicted state exists
- Updates predicted state in StateManager

### Lifecycle

1. **Creation**: `new PredictionSystem(...)` – creates prediction system
2. **Prediction**: `predict(input)` – runs local simulation immediately
3. **Reconciliation**: `reset()` – clears predicted state during rollback
4. **Re-application**: Prediction runs again after reconciliation

### Invariants

- Prediction always uses valid base state (authoritative or predicted)
- Predicted state is updated in StateManager after prediction
- Reset only clears predicted state (doesn't affect authoritative)

---

## ReconciliationSystem

**File**: `reconciliation.ts`

**Concept**: Compares server snapshots to predicted state and handles mismatches.

### Interface

```typescript
export interface ReconciliationResult {
  mismatchDetected: boolean
  commandsReapplied: number
  tick: number
  predictedTick: number | null
}

export class ReconciliationSystem {
  constructor(
    stateManager: StateManager,
    localSimulator: LocalSimulator,
    commandHistory: CommandHistory,
    predictionSystem: PredictionSystem
  )
  reconcile(snapshot: SnapshotMessage): ReconciliationResult
  hasMismatch(predicted: GameState, authoritative: GameState): boolean
  rollback(): void
  reapplyCommands(): number
}
```

### Methods

#### `reconcile(snapshot: SnapshotMessage): ReconciliationResult`

Reconciles server snapshot with predicted state.

**Parameters:**
- `snapshot: SnapshotMessage` - Server snapshot

**Returns:** Reconciliation result

---

#### `hasMismatch(predicted: GameState, authoritative: GameState): boolean`

Checks if two game states have a mismatch.

**Parameters:**
- `predicted: GameState` - Predicted state
- `authoritative: GameState` - Authoritative state

**Returns:** True if mismatch detected

---

#### `rollback(): void`

Rolls back predicted state to authoritative state.

---

#### `reapplyCommands(): number`

Re-applies unconfirmed commands to authoritative state.

**Returns:** Number of commands re-applied

### Semantics

- Reconciliation compares predicted state with authoritative state
- If mismatch detected: rollback, then re-apply unconfirmed commands
- If no mismatch: mark commands as confirmed
- Uses tolerance for floating-point comparisons
- Re-applies commands in sequence order

### Lifecycle

1. **Snapshot Received**: `reconcile(snapshot)` – receives server snapshot
2. **Comparison**: Compares predicted vs authoritative state
3. **Mismatch Handling**: Rollback and re-apply if mismatch, confirm if match
4. **Result**: Returns reconciliation result

### Invariants

- Reconciliation always updates authoritative state first
- Rollback only clears predicted state (doesn't modify authoritative)
- Re-applied commands are processed in sequence order
- Commands are only confirmed if states match

---

## InterpolationSystem

**File**: `interpolation.ts`

**Concept**: Smooths between server snapshots using a buffer for jitter-free rendering.

### Interface

```typescript
export class InterpolationSystem {
  constructor(stateManager: StateManager, bufferMs?: number)
  addSnapshot(snapshot: SnapshotMessage, timestamp: number): void
  update(currentTime: number): void
  clear(): void
  getBufferSize(): number
  hasEnoughData(): boolean
}
```

### Methods

#### `constructor(stateManager: StateManager, bufferMs?: number)`

Creates a new InterpolationSystem.

**Parameters:**
- `stateManager: StateManager` - State manager
- `bufferMs?: number` - Buffer duration in milliseconds (default: 125, clamped to 100-150)

---

#### `addSnapshot(snapshot: SnapshotMessage, timestamp: number): void`

Adds a snapshot to the buffer with timestamp.

**Parameters:**
- `snapshot: SnapshotMessage` - Server snapshot
- `timestamp: number` - Timestamp in milliseconds (use performance.now())

---

#### `update(currentTime: number): void`

Updates interpolated state based on current time.

**Parameters:**
- `currentTime: number` - Current time in milliseconds (use performance.now())

---

#### `clear(): void`

Clears the snapshot buffer.

---

#### `getBufferSize(): number`

Gets the number of snapshots in the buffer.

**Returns:** Number of snapshots

---

#### `hasEnoughData(): boolean`

Checks if buffer has enough data for interpolation (at least 2 snapshots).

**Returns:** True if enough data

### Semantics

- Interpolation buffers snapshots (100-150ms worth)
- Interpolates into the past by buffer amount (accounts for network jitter)
- Linear interpolation for positions/velocities
- Angle interpolation handles wrap-around
- Discrete values (active, done, win) use newer snapshot

### Lifecycle

1. **Creation**: `new InterpolationSystem(stateManager, bufferMs)` – creates system
2. **Buffering**: `addSnapshot()` – adds snapshots to buffer
3. **Interpolation**: `update()` – interpolates and updates state
4. **Cleanup**: `clear()` – clears buffer

### Invariants

- Buffer size is limited (max 10 snapshots)
- Old snapshots are removed beyond buffer duration
- Interpolation requires at least 2 snapshots
- Interpolated state is always valid (never null)

---

## Ownership & Dependencies

### Simulation Package Ownership

- **Only `client/src/sim` may define simulation logic**
- Simulation package manages state but does not implement rendering
- Simulation package runs local simulation but does not handle networking
- State management, prediction, reconciliation, and interpolation must live in `/sim`

### Dependencies

- **Imports**:
  - `@net` – CommandHistory, protocol types
- **No dependencies on**: core, gfx, input, ui packages
- Simulation is independent layer (used by gfx for state)

### No Duplication Rules

- **No state management elsewhere**: State coordination must live in `/sim`
- **No local simulation elsewhere**: Client-side simulation must live in `/sim`
- **No prediction/reconciliation elsewhere**: Prediction logic must live in `/sim`
- **Simulation does not implement**: Rendering, networking, or input handling

---

## Error Handling

### StateManager Errors

- StateManager operations are safe (no errors thrown)
- Returns null for missing states (use has* methods to check)

### LocalSimulator Errors

- Simulation is deterministic (no errors thrown)
- Invalid input is clamped to valid ranges

---

## Dependencies

- `@net` - CommandHistory, protocol types

---

## Version Notes

This API describes v0 simulation layer. Key features:
- Three-layer state management (authoritative, predicted, interpolated)
- Client-side prediction for responsive gameplay
- Server reconciliation with rollback
- Snapshot interpolation for smooth rendering
- Deterministic local simulation matching server

Future extensions may include:
- Entity interpolation
- Lag compensation
- Rollback snapshots
- Prediction accuracy metrics
- Adaptive interpolation buffer

