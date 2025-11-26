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
- Dynamic layers: radar, hud (created by Renderer)
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

**Concept**: Updates PixiJS sprites from game state, handling coordinate transformation and sprite lifecycle. Integrates camera and radar systems.

### Interface

```typescript
export class Renderer {
  constructor(stateManager: StateManager, scene: Scene, app: App, camera?: Camera, radar?: Radar)
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

- Renderer transforms world coordinates to screen coordinates using camera
- Renderer positions 'radar' layer at top-right corner (screen.width - 210, 10)
- Renderer positions 'hud' layer at top-left corner (0, 0) for absolute screen coordinates
- Camera follows player's ship smoothly (lerp factor ~0.1–0.15)
- World (0,0) maps to screen center (adjusted by camera position)
- Y-axis is flipped (world Y-up, screen Y-down)
- Sprites are created on first update, updated on subsequent calls
- Sprites are removed when entities are removed from state
- Uses sprite factories for create/update/destroy operations
- Radar is updated each frame with latest snapshot data

### Coordinate Transformation

**World Coordinates:**
- Origin at (0, 0)
- Y increases upward
- World bounds: [-1000, 1000] × [-1000, 1000] (2000 m × 2000 m)

**Screen Coordinates:**
- Origin at top-left
- Y increases downward
- World (0, 0) maps to screen center (adjusted by camera)

**Transformation (with camera):**
```typescript
screenX = worldX - cameraX + screenWidth / 2
screenY = -(worldY - cameraY) + screenHeight / 2  // Flip Y-axis
```

**Camera:**
- Camera position lerps toward player ship position
- Camera stays within world bounds (no wraparound)
- Camera position updated before sprite rendering

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

## Camera

**File**: `camera.ts`

**Concept**: Camera system that smoothly follows player's ship with lerp smoothing.

### Interface

```typescript
export class Camera {
  constructor(worldWidth: number, worldHeight: number, viewportWidth: number, viewportHeight: number, lerpFactor?: number)
  update(targetPos: Vec2): void
  getPosition(): Vec2
  getViewport(): { width: number, height: number }
  setLerpFactor(factor: number): void
}
```

### Methods

#### `constructor(worldWidth, worldHeight, viewportWidth, viewportHeight, lerpFactor?)`

Creates a new Camera instance.

**Parameters:**
- `worldWidth: number` - World width in meters (2000.0)
- `worldHeight: number` - World height in meters (2000.0)
- `viewportWidth: number` - Viewport width in pixels
- `viewportHeight: number` - Viewport height in pixels
- `lerpFactor?: number` - Lerp factor (default: 0.1–0.15)

#### `update(targetPos: Vec2): void`

Updates camera position to smoothly follow target.

**Parameters:**
- `targetPos: Vec2` - Target position (player ship position)

#### `getPosition(): Vec2`

Gets current camera position.

**Returns:** Camera position in world coordinates

### Semantics

- Camera lerps position toward target: `cameraPos = cameraPos + (targetPos - cameraPos) * lerpFactor`
- Camera position clamped to world bounds (stays within world, no wraparound)
- Lerp factor typically 0.1–0.15 per frame for smooth following
- Camera does not wrap around (unlike ships)

### Invariants

- Camera position always within world bounds
- Lerp factor is between 0 and 1
- Camera position updates before sprite rendering

---

## Radar

**File**: `radar.ts`

**Concept**: Mini-map showing full world view with all entities (ships, planets, pallets).

### Interface

```typescript
export class Radar {
  constructor(
    worldWidth: number,
    worldHeight: number,
    radarWidth: number,
    radarHeight: number,
    container: Container
  )
  update(snapshot: SnapshotMessage, myShipId: number): void
  destroy(): void
}
```

### Methods

#### `constructor(worldWidth, worldHeight, radarWidth, radarHeight, container)`

Creates a new Radar instance.

**Parameters:**
- `worldWidth: number` - World width in meters (2000.0)
- `worldHeight: number` - World height in meters (2000.0)
- `radarWidth: number` - Radar width in pixels (200)
- `radarHeight: number` - Radar height in pixels (200)
- `container: Container` - PixiJS container for radar graphics

#### `update(snapshot: SnapshotMessage, myShipId: number): void`

Updates radar with latest snapshot data.

**Parameters:**
- `snapshot: SnapshotMessage` - Server snapshot
- `myShipId: number` - Player's ship ID

### Semantics

- Radar shows full world view (2000 m × 2000 m) in fixed 200×200 px area
- World coordinates mapped to radar pixels: `radarX = ((worldX - (-WORLD_WIDTH/2)) / WORLD_WIDTH) * RADAR_WIDTH`
- Shows: world bounds (rectangle outline), planets (circles, size by radius, color by mass), ships (dots, color by player ID), own ship (highlighted), pallets (small dots)
- Updates every frame with latest snapshot data

### Coordinate Mapping

**World to Radar:**
```typescript
radarX = ((worldX - (-WORLD_WIDTH/2)) / WORLD_WIDTH) * RADAR_WIDTH
radarY = ((worldY - (-WORLD_HEIGHT/2)) / WORLD_HEIGHT) * RADAR_HEIGHT
```

**World bounds [-1000, 1000] × [-1000, 1000] map to [0, 200] × [0, 200] pixels**

### Invariants

- Radar size is fixed (200×200 px)
- World coordinates always map to valid radar pixel coordinates
- Own ship is always highlighted differently
- Radar updates every frame

---

## Version Notes

This API describes v1 graphics layer. Key features:
- PixiJS scene hierarchy
- Sprite factories for entities (ships, planets, pallets)
- World-to-screen coordinate transformation with camera
- Camera system following player's ship smoothly
- Radar system showing full world view
- Automatic sprite lifecycle management

**Changes from v0**:
- Added Camera class for smooth ship following
- Added Radar class for mini-map visualization
- Renderer integrates camera and radar updates

Future extensions may include:
- Sprite caching
- Batch rendering
- Particle effects
- Animation system
- Texture atlases

