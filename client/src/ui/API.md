# Orbital Rush v0 – UI Package API

This document provides detailed API reference for the UI package.

---

## Scope & Location

**Scope**: User interface system (main menu, room lobby, HUD, UI components).

**Code location**: `client/src/ui`

**Design Goals**:
- Provide main menu UI (create room, join room)
- Provide room lobby UI (room code, player list, start match)
- Provide HUD (Heads-Up Display) coordination
- Manage UI components (energy bar, pallet counter, game banner, player indicator)
- Update UI from game state
- Abstract UI component details

---

## MainMenu

**File**: `main-menu.ts`

**Concept**: Main menu UI with "Create Room" and "Join Room" options.

### Interface

```typescript
export class MainMenu {
  constructor(container: Container, onCreateRoom: () => void, onJoinRoom: (code: string) => void)
  show(): void
  hide(): void
  destroy(): void
}
```

### Methods

#### `constructor(container, onCreateRoom, onJoinRoom)`

Creates a new MainMenu instance.

**Parameters:**
- `container: Container` - PixiJS container for menu UI
- `onCreateRoom: () => void` - Callback when "Create Room" clicked
- `onJoinRoom: (code: string) => void` - Callback when "Join Room" clicked (with room code)

#### `show(): void`

Shows the main menu.

#### `hide(): void`

Hides the main menu.

#### `destroy(): void`

Destroys the main menu and cleans up.

### Semantics

- Main menu is entry point (shown on app start)
- "Create Room" button triggers onCreateRoom callback
- "Join Room" button shows input field for room code, triggers onJoinRoom callback
- Menu is hidden when transitioning to lobby or game

---

## RoomLobby

**File**: `room-lobby.ts`

**Concept**: Room lobby UI showing room code, player list, and start match button.

### Interface

```typescript
export interface PlayerInfo {
  id: number
  name: string
}

export class RoomLobby {
  constructor(
    container: Container,
    roomCode: string,
    players: PlayerInfo[],
    isHost: boolean,
    onStartMatch: () => void,
    onLeaveRoom: () => void
  )
  updatePlayers(players: PlayerInfo[]): void
  updateRoomCode(roomCode: string): void
  show(): void
  hide(): void
  destroy(): void
}
```

### Methods

#### `constructor(container, roomCode, players, isHost, onStartMatch, onLeaveRoom)`

Creates a new RoomLobby instance.

**Parameters:**
- `container: Container` - PixiJS container for lobby UI
- `roomCode: string` - Room code (6-character alphanumeric)
- `players: PlayerInfo[]` - List of players in room
- `isHost: boolean` - Whether current player is host
- `onStartMatch: () => void` - Callback when "Start Match" clicked (host only)
- `onLeaveRoom: () => void` - Callback when "Leave Room" clicked

#### `updatePlayers(players: PlayerInfo[]): void`

Updates player list.

**Parameters:**
- `players: PlayerInfo[]` - Updated player list

#### `updateRoomCode(roomCode: string): void`

Updates room code display.

**Parameters:**
- `roomCode: string` - Room code

#### `show(): void`

Shows the lobby.

#### `hide(): void`

Hides the lobby.

#### `destroy(): void`

Destroys the lobby and cleans up.

### Semantics

- Lobby shows room code prominently (copyable)
- Player list shows all connected players
- "Start Match" button only visible to host, disabled if < 2 players
- "Leave Room" button always visible
- Waiting message shown if not host
- Lobby is hidden when match starts

---

## HUD

**File**: `hud.ts`

**Concept**: Coordinator that manages all in-game UI components and updates them from game state.

### Interface

```typescript
export class HUD {
  constructor(scene: Scene, stateManager: StateManager, myShipId: number)
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

- HUD manages four UI components: EnergyBar, PalletCounter, GameBanner, PlayerIndicator
- Components are created in UI layer during construction
- Update reads game state and updates all components
- Components are positioned at fixed screen coordinates
- Game banner shows win/lose messages based on game state
- Player indicator shows player name/ID (from myShipId)

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

### PlayerIndicator

**File**: `components/player-indicator.ts`

Displays player name/ID indicator.

**Methods:**
- `constructor(container: Container, config: PlayerIndicatorConfig)`
- `update(playerId: number, playerName?: string): void`
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

This API describes v1 UI layer. Key features:
- Main menu UI (create room, join room)
- Room lobby UI (room code, player list, start match)
- HUD coordinator for in-game UI management
- Energy bar, pallet counter, game banner, player indicator components
- State-driven UI updates
- PixiJS-based rendering

**Changes from v0**:
- Added MainMenu class for entry point UI
- Added RoomLobby class for room management UI
- Added PlayerIndicator component to HUD
- HUD now takes myShipId parameter for player identification

Future extensions may include:
- More UI components
- UI animations
- Responsive layout
- UI themes
- Accessibility features

