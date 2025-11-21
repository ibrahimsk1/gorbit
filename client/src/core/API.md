# Orbital Rush v0 – Core Package API

This document provides detailed API reference for the core package.

---

## Scope & Location

**Scope**: Core application infrastructure (PixiJS app wrapper, render loop management).

**Code location**: `client/src/core`

**Design Goals**:
- Encapsulate PixiJS application lifecycle
- Provide render loop with FPS tracking
- Handle window resizing and canvas management
- Abstract PixiJS initialization details
- Foundation layer for rendering, simulation, and UI systems

---

## App

**File**: `app.ts`

**Concept**: PixiJS Application wrapper that manages initialization, canvas creation, and window resizing.

### Interface

```typescript
export interface AppConfig {
  width?: number        // Default: 1280
  height?: number       // Default: 720
  backgroundColor?: number  // Default: 0x0a0a0a
  resolution?: number   // Default: 1
  autoResize?: boolean  // Default: true
}

export class App {
  constructor(config?: AppConfig)
  async init(container?: HTMLElement): Promise<void>
  getApplication(): Application
  getCanvas(): HTMLCanvasElement
  destroy(): void
}
```

### Methods

#### `constructor(config?: AppConfig)`

Creates a new App instance with optional configuration.

**Parameters:**
- `config?: AppConfig` - Optional configuration object

**Returns:** App instance

---

#### `async init(container?: HTMLElement): Promise<void>`

Initializes the PixiJS application and attaches canvas to the DOM.

**Parameters:**
- `container?: HTMLElement` - Container element (default: `#app`)

**Returns:** Promise that resolves when initialization is complete

**Throws:** Error if container not found

**Example:**
```typescript
const app = new App()
await app.init()
```

---

#### `getApplication(): Application`

Returns the underlying PixiJS Application instance.

**Returns:** PixiJS Application

**Throws:** Error if app not initialized

**Example:**
```typescript
const pixiApp = app.getApplication()
```

---

#### `getCanvas(): HTMLCanvasElement`

Returns the canvas element.

**Returns:** HTMLCanvasElement

**Throws:** Error if app not initialized

**Example:**
```typescript
const canvas = app.getCanvas()
```

---

#### `destroy(): void`

Destroys the application and cleans up all resources. Idempotent.

**Example:**
```typescript
app.destroy()
```

### Semantics

- App manages one PixiJS Application instance
- Canvas is appended to container (or `#app` element by default)
- Auto-resize updates canvas size when window resizes
- Destroy cleans up all resources (textures, children, etc.)
- Initialization is async (PixiJS v8 requirement)

### Lifecycle

1. **Creation**: `new App(config)` – creates app wrapper with configuration
2. **Initialization**: `await app.init(container)` – initializes PixiJS and attaches canvas
3. **Usage**: `app.getApplication()` – access PixiJS Application for rendering
4. **Cleanup**: `app.destroy()` – destroys application and cleans up resources

### Invariants

- App must be initialized before use (throws error if not)
- Only one Application instance per App instance
- Canvas is always attached to DOM when initialized
- Destroy is idempotent (safe to call multiple times)
- Re-initialization destroys previous instance first

---

## RenderLoop

**File**: `render-loop.ts`

**Concept**: Manages the render loop using `requestAnimationFrame` and tracks FPS.

### Interface

```typescript
export class RenderLoop {
  constructor(app: App)
  start(): void
  stop(): void
  getFPS(): number
}
```

### Methods

#### `constructor(app: App)`

Creates a render loop for the given app.

**Parameters:**
- `app: App` - App instance to render

**Returns:** RenderLoop instance

---

#### `start(): void`

Starts the render loop. Idempotent.

**Example:**
```typescript
renderLoop.start()
```

---

#### `stop(): void`

Stops the render loop. Idempotent.

**Example:**
```typescript
renderLoop.stop()
```

---

#### `getFPS(): number`

Returns the current FPS (calculated over 1-second intervals).

**Returns:** Current FPS (0 if not started)

**Example:**
```typescript
const fps = renderLoop.getFPS()
console.log(`FPS: ${fps}`)
```

### Semantics

- Uses `requestAnimationFrame` for browser-optimized rendering
- FPS calculated over 1-second intervals
- Calls `app.render()` each frame
- Stops gracefully if app is destroyed
- FPS updates every second (1000ms interval)

### Lifecycle

1. **Creation**: `new RenderLoop(app)` – creates render loop for app
2. **Start**: `renderLoop.start()` – begins render loop
3. **Running**: Loop calls `app.render()` each frame, tracks FPS
4. **Stop**: `renderLoop.stop()` – stops render loop

### Invariants

- Start/stop are idempotent (safe to call multiple times)
- FPS updates every second
- Render loop stops if app is destroyed (error handling)
- Only one render loop can run at a time per instance

---

## Types

### AppConfig

Configuration options for App initialization.

```typescript
interface AppConfig {
  width?: number        // Canvas width in pixels (default: 1280)
  height?: number       // Canvas height in pixels (default: 720)
  backgroundColor?: number  // Background color as hex number (default: 0x0a0a0a)
  resolution?: number   // Resolution multiplier (default: 1)
  autoResize?: boolean  // Enable automatic window resize handling (default: true)
}
```

---

## Constants

### App Defaults

- `DEFAULT_WIDTH = 1280` - Default canvas width (pixels)
- `DEFAULT_HEIGHT = 720` - Default canvas height (pixels)
- `DEFAULT_BACKGROUND_COLOR = 0x0a0a0a` - Default background color (dark gray)
- `DEFAULT_RESOLUTION = 1` - Default resolution (1:1 pixel ratio)
- `DEFAULT_AUTO_RESIZE = true` - Default auto-resize enabled

### RenderLoop Constants

- `FPS_UPDATE_INTERVAL = 1000` - FPS calculation interval (milliseconds)

---

## Ownership & Dependencies

### Core Package Ownership

- **Only `client/src/core` may define core application infrastructure**
- Core package manages PixiJS lifecycle but does not implement game logic
- Core package provides render loop but does not implement rendering logic
- Core is foundation layer that other packages depend on

### Dependencies

- **Imports**:
  - `pixi.js` – PixiJS Application and renderer types (external dependency)
- **No dependencies on**: gfx, sim, net, input, ui packages
- Core is the lowest-level package in the client architecture

### No Duplication Rules

- **No PixiJS app management elsewhere**: Application lifecycle must live in `/core`
- **No render loop elsewhere**: Render loop coordination must live in `/core`
- **Core does not implement**: Rendering, simulation, networking, or input handling
- Other packages depend on core but core does not depend on them

---

## Error Handling

### App Errors

- **Container not found**: Thrown when `init()` is called and no container element is found
- **App not initialized**: Thrown when `getApplication()` or `getCanvas()` is called before `init()`

### RenderLoop Errors

- RenderLoop handles app destruction gracefully and stops automatically if the app is destroyed

---

## Dependencies

- `pixi.js` - PixiJS Application and renderer types

---

## Version Notes

This API describes v0 core infrastructure. Key features:
- PixiJS v8 application wrapper
- Render loop with FPS tracking
- Auto-resize support
- Resource cleanup
- Async initialization

Future extensions may include:
- Multiple render targets
- Render quality settings
- Performance monitoring
- Frame rate limiting
- Render statistics API
