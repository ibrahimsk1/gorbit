# Orbital Rush v0 – Input Package API

This document provides detailed API reference for the input package.

---

## Scope & Location

**Scope**: Input handling system (keyboard input, command creation).

**Code location**: `client/src/input`

**Design Goals**:
- Handle keyboard input for game controls
- Convert key presses to game input actions
- Provide command creation utilities
- Abstract input device details

---

## KeyboardInputHandler

**File**: `keyboard.ts`

**Concept**: Tracks keyboard state and converts key presses to game input actions (thrust and turn).

### Interface

```typescript
export class KeyboardInputHandler {
  constructor()
  getThrust(): number
  getTurn(): number
  onKeyDown(key: string): void
  onKeyUp(key: string): void
  attach(): void
  detach(): void
  reset(): void
}
```

### Methods

#### `constructor()`

Creates a new KeyboardInputHandler instance.

**Returns:** KeyboardInputHandler instance

---

#### `getThrust(): number`

Gets the current thrust value [0.0, 1.0].

**Returns:** 1.0 if thrust key is pressed, 0.0 otherwise

**Example:**
```typescript
const thrust = keyboard.getThrust()
```

---

#### `getTurn(): number`

Gets the current turn value [-1.0, 1.0].

**Returns:** -1.0 if turn left, 1.0 if turn right, 0.0 if neither or both

**Example:**
```typescript
const turn = keyboard.getTurn()
```

---

#### `onKeyDown(key: string): void`

Handles a key press event.

**Parameters:**
- `key: string` - Key name (e.g., 'ArrowUp', 'w', 'ArrowLeft', 'a', 'ArrowRight', 'd')

**Example:**
```typescript
keyboard.onKeyDown('ArrowUp')
```

---

#### `onKeyUp(key: string): void`

Handles a key release event.

**Parameters:**
- `key: string` - Key name

**Example:**
```typescript
keyboard.onKeyUp('ArrowUp')
```

---

#### `attach(): void`

Attaches event listeners to the window for keyboard events.

**Example:**
```typescript
keyboard.attach()
```

---

#### `detach(): void`

Removes event listeners from the window.

**Example:**
```typescript
keyboard.detach()
```

---

#### `reset(): void`

Resets all input state to neutral (thrust=0, turn=0).

**Example:**
```typescript
keyboard.reset()
```

### Key Mappings

- **Thrust**: ArrowUp or W key
- **Turn left**: ArrowLeft or A key
- **Turn right**: ArrowRight or D key

### Semantics

- KeyboardInputHandler tracks keyboard state internally
- Key names are case-insensitive (normalized to lowercase)
- Thrust is binary (0.0 or 1.0)
- Turn is ternary (-1.0, 0.0, or 1.0)
- Both turn keys pressed cancels out (returns 0.0)
- Event listeners are attached to window when `attach()` is called

### Lifecycle

1. **Creation**: `new KeyboardInputHandler()` – creates handler
2. **Attachment**: `attach()` – starts listening for keyboard events
3. **Usage**: `getThrust()` and `getTurn()` – query current input state
4. **Detachment**: `detach()` – stops listening for keyboard events
5. **Reset**: `reset()` – clears input state

### Invariants

- Thrust is always 0.0 or 1.0
- Turn is always -1.0, 0.0, or 1.0
- Attach/detach are idempotent (safe to call multiple times)
- Reset clears all input state
- Event listeners are only attached when `attach()` is called

---

## Command

**File**: `command.ts`

**Concept**: Command creation utilities for input commands.

### Functions

#### `createCommand(thrust: number, turn: number): InputCommand`

Creates an input command with thrust and turn values.

**Parameters:**
- `thrust: number` - Thrust value [0.0, 1.0]
- `turn: number` - Turn value [-1.0, 1.0]

**Returns:** InputCommand object

### Semantics

- Command creation is a pure function
- Values are not clamped (caller responsible for valid ranges)
- Command format matches server InputCommand type

### Invariants

- Command creation is pure (no side effects)
- Command structure matches server contract

---

## Ownership & Dependencies

### Input Package Ownership

- **Only `client/src/input` may define input handling logic**
- Input package handles keyboard but does not implement game logic
- Input package converts keys to actions but does not send commands
- Keyboard input handling must live in `/input`

### Dependencies

- **Imports**:
  - Native DOM API (KeyboardEvent, window)
- **No dependencies on**: core, gfx, sim, net, ui packages
- Input is independent layer

### No Duplication Rules

- **No keyboard handling elsewhere**: Keyboard input must live in `/input`
- **No command creation elsewhere**: Command utilities must live in `/input`
- **Input does not implement**: Simulation, rendering, networking, or UI

---

## Error Handling

### KeyboardInputHandler Errors

- Input operations are safe (no errors thrown)
- Invalid keys are ignored (no effect)

---

## Dependencies

- Native DOM API (KeyboardEvent, window)

---

## Version Notes

This API describes v0 input layer. Key features:
- Keyboard input handling
- Key-to-action mapping
- Event listener management
- Command creation utilities

Future extensions may include:
- Gamepad support
- Mouse input
- Touch input
- Input remapping
- Input buffering

