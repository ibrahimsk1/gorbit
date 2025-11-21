# Orbital Rush v0 – Graphics Package API

This document provides detailed API reference for the graphics package.

---

## Scope & Location

**Scope**: Graphics and rendering system (scene management, sprite rendering, coordinate transformation).

**Code location**: `client/src/gfx`

**Design Goals**:
- Provide PixiJS scene hierarchy management
- Transform game state into visual representation
- Handle world-to-screen coordinate transformation
- Manage sprite lifecycle (create, update, destroy)
- Support extensible sprite factories

---

## Scene

**File**: `scene.ts`

**Concept**: Manages PixiJS scene hierarchy with layers for organizing game objects.

### Interface

```typescript
export class Scene {
  constructor(app: App)
  getRoot(): Container
  getLayer(name: string): Container
  addChild(child: Container, layerName?: string): void
  removeChild(child: Container, layerName?: string): void
  destroy(): void
}
```

### Methods

#### `constructor(app: App)`

Creates a new scene with default layers (background, game, ui).

**Parameters:**
- `app: App` - App instance for PixiJS access

**Returns:** Scene instance

---

#### `getRoot(): Container`

Returns the root container.

**Returns:** PixiJS Container (root)

---

#### `getLayer(name: string): Container`

Gets or creates a layer with the specified name.

**Parameters:**
- `name: string` - Layer name

**Returns:** PixiJS Container for the layer

**Example:**
```typescript
const gameLayer = scene.getLayer('game')
```

---

#### `addChild(child: Container, layerName?: string): void`

Adds a child to the scene.

**Parameters:**
- `child: Container` - PixiJS container to add
- `layerName?: string` - Optional layer name (adds to root if omitted)

**Example:**
```typescript
scene.addChild(sprite, 'game')
```

---

#### `removeChild(child: Container, layerName?: string): void`

Removes a child from the scene.

**Parameters:**
- `child: Container` - PixiJS container to remove
- `layerName?: string` - Optional layer name

---

#### `destroy(): void`

Destroys the scene and all children.

**Example:**
```typescript
scene.destroy()
```

### Semantics

- Scene manages a root container added to PixiJS stage
- Layers are created on-demand when accessed
- Default layers: background, game, ui
- Root container is always attached to PixiJS stage
- Destroy removes root from stage and cleans up all children

### Lifecycle

1. **Creation**: `new Scene(app)` – creates scene with default layers
2. **Usage**: Access layers with `getLayer()`, add/remove children
3. **Cleanup**: `scene.destroy()` – destroys scene and all children

### Invariants

- Root container is always attached to PixiJS stage
- Layers are created lazily (on first access)
- Destroy is idempotent (safe to call multiple times)
- All children are destroyed when scene is destroyed

---

## Renderer

**File**: `renderer.ts`

**Concept**: Updates PixiJS sprites from game state, handling coordinate transformation and sprite lifecycle.

### Interface

```typescript
export class Renderer {
  constructor(stateManager: StateManager, scene: Scene, app: App)
  update(): void
  clear(): void
  destroy(): void
}
```

### Methods

#### `constructor(stateManager: StateManager, scene: Scene, app: App)`

Creates a new renderer.

**Parameters:**
- `stateManager: StateManager` - State manager for game state
- `scene: Scene` - Scene for sprite hierarchy
- `app: App` - App for coordinate transformation

**Returns:** Renderer instance

---

#### `update(): void`

Updates all sprites from current game state. Should be called each frame.

**Example:**
```typescript
renderer.update()
```

---

#### `clear(): void`

Clears all sprites from scene.

**Example:**
```typescript
renderer.clear()
```

---

#### `destroy(): void`

Destroys renderer and cleans up all resources.

**Example:**
```typescript
renderer.destroy()
```

### Semantics

- Renderer transforms world coordinates to screen coordinates
- World (0,0) maps to screen center
- Y-axis is flipped (world Y-up, screen Y-down)
- Sprites are created on first update, updated on subsequent calls
- Sprites are removed when entities are removed from state
- Uses sprite factories for create/update/destroy operations

### Coordinate Transformation

**World Coordinates:**
- Origin at (0, 0)
- Y increases upward

**Screen Coordinates:**
- Origin at top-left
- Y increases downward
- World (0, 0) maps to screen center

**Transformation:**
```typescript
screenX = worldX + screenWidth / 2
screenY = -worldY + screenHeight / 2  // Flip Y-axis
```

### Lifecycle

1. **Creation**: `new Renderer(stateManager, scene, app)` – creates renderer
2. **Usage**: Call `update()` each frame to sync sprites with state
3. **Cleanup**: `renderer.destroy()` – destroys all sprites and cleans up

### Invariants

- Sprites are always in sync with game state after `update()`
- Sprites are created lazily (on first update)
- Sprites are removed when entities removed from state
- Coordinate transformation is consistent (world center = screen center)
- Destroy clears all sprites

---

## Sprite Factories

**File**: `sprites/*.ts`

**Concept**: Factory pattern for creating, updating, and destroying sprites.

### ShipSpriteFactory

**File**: `sprites/ship-sprite.ts`

- `create(ship: ShipSnapshot): Graphics` – Creates ship sprite
- `update(sprite: Graphics, ship: ShipSnapshot): void` – Updates ship sprite
- `destroy(sprite: Graphics): void` – Destroys ship sprite

### PlanetSpriteFactory

**File**: `sprites/planet-sprite.ts`

- `create(planet: PlanetSnapshot): Graphics` – Creates planet sprite
- `update(sprite: Graphics, planet: PlanetSnapshot): void` – Updates planet sprite
- `destroy(sprite: Graphics): void` – Destroys planet sprite

### PalletSpriteFactory

**File**: `sprites/pallet-sprite.ts`

- `create(pallet: PalletSnapshot): Graphics` – Creates pallet sprite
- `update(sprite: Graphics, pallet: PalletSnapshot): void` – Updates pallet sprite
- `destroy(sprite: Graphics): void` – Destroys pallet sprite

### Semantics

- Factories use Graphics objects for rendering
- Create initializes sprite with entity state
- Update modifies existing sprite to match entity state
- Destroy removes sprite from scene and cleans up resources
- Factories are stateless (pure functions)

### Invariants

- Create always produces valid sprite
- Update only modifies sprite properties (doesn't recreate)
- Destroy removes sprite from parent and cleans up
- Factories are idempotent (safe to call multiple times)

---

## Ownership & Dependencies

### Graphics Package Ownership

- **Only `client/src/gfx` may define graphics and rendering logic**
- Graphics package handles sprite rendering but does not implement game logic
- Graphics package transforms state to visuals but does not manage state
- Sprite factories must live in `/gfx/sprites`

### Dependencies

- **Imports**:
  - `pixi.js` – PixiJS Graphics and Container types (external dependency)
  - `@core` – App for PixiJS access
  - `@sim` – StateManager for game state
- **No dependencies on**: net, input, ui packages

### No Duplication Rules

- **No sprite rendering elsewhere**: Sprite management must live in `/gfx`
- **No scene management elsewhere**: Scene hierarchy must live in `/gfx`
- **No coordinate transformation elsewhere**: World-to-screen transform must live in `/gfx`
- **Graphics does not implement**: Simulation, networking, or input handling

---

## Error Handling

### Scene Errors

- Scene operations are safe (no errors thrown)
- Destroy is idempotent

### Renderer Errors

- Renderer operations are safe (no errors thrown)
- Update handles missing state gracefully
- Destroy is idempotent

---

## Dependencies

- `pixi.js` - PixiJS Graphics and Container types
- `@core` - App for PixiJS access
- `@sim` - StateManager for game state

---

## Version Notes

This API describes v0 graphics layer. Key features:
- PixiJS scene hierarchy
- Sprite factories for entities
- World-to-screen coordinate transformation
- Automatic sprite lifecycle management

Future extensions may include:
- Sprite caching
- Batch rendering
- Particle effects
- Animation system
- Texture atlases

