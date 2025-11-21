# Orbital Rush v0 – UI Package API

This document provides detailed API reference for the UI package.

---

## Scope & Location

**Scope**: User interface system (HUD, UI components).

**Code location**: `client/src/ui`

**Design Goals**:
- Provide HUD (Heads-Up Display) coordination
- Manage UI components (energy bar, pallet counter, game banner)
- Update UI from game state
- Abstract UI component details

---

## HUD

**File**: `hud.ts`

**Concept**: Coordinator that manages all UI components and updates them from game state.

### Interface

```typescript
export class HUD {
  constructor(scene: Scene, stateManager: StateManager)
  update(): void
  destroy(): void
}
```

### Methods

#### `constructor(scene: Scene, stateManager: StateManager)`

Creates a new HUD coordinator.

**Parameters:**
- `scene: Scene` - Scene for UI layer access
- `stateManager: StateManager` - State manager for game state

**Returns:** HUD instance

---

#### `update(): void`

Updates all HUD components from current game state. Should be called each frame or when state changes.

**Example:**
```typescript
hud.update()
```

---

#### `destroy(): void`

Destroys the HUD and cleans up all components.

**Example:**
```typescript
hud.destroy()
```

### Semantics

- HUD manages three UI components: EnergyBar, PalletCounter, GameBanner
- Components are created in UI layer during construction
- Update reads game state and updates all components
- Components are positioned at fixed screen coordinates
- Game banner shows win/lose messages based on game state

### Lifecycle

1. **Creation**: `new HUD(scene, stateManager)` – creates HUD with components
2. **Usage**: `update()` – updates components from game state
3. **Cleanup**: `destroy()` – destroys all components

### Invariants

- HUD always has three components (energy bar, pallet counter, game banner)
- Components are always in UI layer
- Update reads from render state (interpolated or authoritative)
- Destroy cleans up all components

---

## UI Components

**File**: `components/*.ts`

**Concept**: Individual UI components for displaying game information.

### EnergyBar

**File**: `components/energy-bar.ts`

Displays ship energy level as a progress bar.

**Methods:**
- `constructor(container: Container, config: EnergyBarConfig)`
- `update(energy: number, maxEnergy: number): void`
- `destroy(): void`

### PalletCounter

**File**: `components/pallet-counter.ts`

Displays count of active/total pallets.

**Methods:**
- `constructor(container: Container, config: PalletCounterConfig)`
- `update(active: number, total: number): void`
- `destroy(): void`

### GameBanner

**File**: `components/game-banner.ts`

Displays win/lose messages.

**Methods:**
- `constructor(container: Container)`
- `showWin(): void`
- `showLose(): void`
- `hide(): void`
- `updateSize(width: number, height: number): void`
- `destroy(): void`

### Semantics

- Components use PixiJS Graphics for rendering
- Components are added to UI layer container
- Update methods modify component appearance
- Destroy removes components from scene

### Invariants

- Components are always in UI layer
- Update methods are idempotent
- Destroy removes components from parent

---

## Ownership & Dependencies

### UI Package Ownership

- **Only `client/src/ui` may define UI component logic**
- UI package handles HUD but does not implement game logic
- UI package displays state but does not manage state
- UI components must live in `/ui/components`

### Dependencies

- **Imports**:
  - `pixi.js` – PixiJS Graphics and Container types (external dependency)
  - `@gfx` – Scene for UI layer access
  - `@sim` – StateManager for game state
- **No dependencies on**: core, net, input packages

### No Duplication Rules

- **No HUD coordination elsewhere**: HUD management must live in `/ui`
- **No UI components elsewhere**: UI components must live in `/ui/components`
- **UI does not implement**: Simulation, rendering, networking, or input handling

---

## Error Handling

### HUD Errors

- HUD operations are safe (no errors thrown)
- Update handles missing state gracefully
- Destroy is idempotent

---

## Dependencies

- `pixi.js` - PixiJS Graphics and Container types
- `@gfx` - Scene for UI layer access
- `@sim` - StateManager for game state

---

## Version Notes

This API describes v0 UI layer. Key features:
- HUD coordinator for UI management
- Energy bar, pallet counter, game banner components
- State-driven UI updates
- PixiJS-based rendering

Future extensions may include:
- More UI components
- UI animations
- Responsive layout
- UI themes
- Accessibility features

