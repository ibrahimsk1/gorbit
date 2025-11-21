# Orbital Rush – Session Subsystem Specification

This document describes the session orchestration layer for Orbital Rush. It defines how the game loop is orchestrated, how input commands are queued and processed, and how time is managed.

---

## Scope & Location

**Scope**: Session orchestration for Orbital Rush (tick loop, command queue, time management).

**Code location**: `server/internal/session`

**Design Goals**:
- Orchestrates the game loop at fixed rate (30 Hz)
- Manages input command queue with sequence-based deduplication
- Provides time abstraction for deterministic testing
- Bridges between transport layer (WebSocket) and simulation layer (rules/physics)
- No direct IO; all IO happens in transport layer

---

## Core Components

### Session

**File**: `server/internal/session/session.go`

**Concept**: Orchestrates the game loop by combining ticker, command queue, and game rules.

**Key Fields**:
- `world entities.World` – Current world state
- `queue *CommandQueue` – Input command queue
- `ticker *Ticker` – Fixed-rate ticker (30 Hz)
- `clock Clock` – Time abstraction interface
- `dt float64` – Time step (1/30 seconds)
- `G, aMax, pickupRadius float64` – Physics constants
- `running bool` – Whether session is active
- `logger logr.Logger` – Optional logger for observability

**Semantics**:
- Session manages one game instance (one World)
- Tick loop processes commands and advances simulation
- Session is started/stopped by transport layer
- Session does not handle network IO (that's transport layer)

**Lifecycle**:
1. **Creation**: `NewSession(clock, world, maxQueueSize)` – creates session with initial world
2. **Start**: `Run(maxTicks)` – starts tick loop (called by transport layer)
3. **Stop**: `Stop()` – stops tick loop gracefully
4. **Query**: `GetWorld()` – returns current world state

**Invariants**:
- Session is single-threaded (one Run() call at a time)
- Commands are processed in sequence order
- Tick rate is fixed at 30 Hz (33ms intervals)
- World state is only modified through rules.Step()

---

### Command Queue

**File**: `server/internal/session/queue.go`

**Concept**: Queue that stores input commands with sequence numbers, maintains ordering, and deduplicates.

**Key Operations**:
- `Enqueue(seq, cmd)` – Add command with sequence number
- `Dequeue()` – Remove and return next command in sequence order
- `Peek()` – View next command without removing
- `Size()` – Get current queue size
- `Clear()` – Remove all commands

**Enqueue Semantics**:
- **Rejects duplicates**: If sequence number already exists, returns false
- **Rejects old sequences**: If sequence < nextSequence (already processed), returns false
- **Rejects when full**: If queue size >= maxSize, returns false
- **Maintains order**: Commands stored in sorted sequence order

**Dequeue Semantics**:
- **Returns lowest sequence**: Always dequeues command with lowest sequence number
- **Updates nextSequence**: Sets nextSequence = dequeuedSequence + 1
- **Returns false if empty**: If no commands, returns false

**Invariants**:
- Commands are always processed in sequence order (lowest first)
- No duplicate sequence numbers in queue
- Queue size never exceeds maxSize
- nextSequence tracks what has been processed

**Queue Size**:
- Typical maxSize: 100 commands
- Queue depth monitored for observability
- Threshold logging at 50% of max size

---

### Ticker

**File**: `server/internal/session/ticker.go`

**Concept**: Generates ticks at a fixed rate using a clock abstraction.

**Key Operations**:
- `ShouldTick(now)` – Check if enough time has passed
- `Tick(now)` – Advance ticker if interval elapsed
- `Reset()` – Reset ticker to current time

**Semantics**:
- Fixed-rate ticker at 30 Hz (33ms intervals)
- Uses clock interface for deterministic testing
- Tracks lastTick time to maintain interval
- Can process multiple ticks if time advanced significantly

**Invariants**:
- Interval is constant (33ms for 30 Hz)
- Ticks occur at fixed rate when time advances normally
- Can catch up if time jumps forward (processes multiple ticks)

---

### Clock Abstraction

**File**: `server/internal/session/ticker.go`

**Concept**: Time abstraction interface for deterministic testing.

**Interface**:
```go
type Clock interface {
    Now() time.Time
}
```

**Implementations**:
- **RealClock**: Wraps `time.Now()` for production
- **FakeClock**: Deterministic clock for testing
  - `Advance(duration)` – Move time forward
  - `SetTime(time)` – Set to specific time

**Semantics**:
- Production uses RealClock (real time)
- Tests use FakeClock (deterministic time)
- Clock abstraction enables deterministic session tests

**Invariants**:
- Clock time only moves forward (or stays same)
- FakeClock allows precise time control for tests

---

## Tick Loop Algorithm

**File**: `server/internal/session/session.go` – `Run(maxTicks)`

**Algorithm**:
1. **Calculate ticks needed**: Based on elapsed time since lastTick
2. **Limit to maxTicks**: Don't process more than requested
3. **For each tick**:
   - Advance ticker (update lastTick)
   - Dequeue next command (or use zero command if empty)
   - Call `rules.Step(world, input, dt, G, aMax, pickupRadius)`
   - Update world state
   - Record tick duration metrics
   - Log slow ticks (>10ms threshold)
   - Break if world.Done == true

**Semantics**:
- Processes all ticks that should have occurred based on elapsed time
- Handles time jumps (processes multiple ticks if needed)
- Commands processed in sequence order
- Zero command used when queue is empty (no input)
- Tick loop stops when world.Done == true

**Invariants**:
- Ticks processed at fixed rate (30 Hz)
- Commands processed in sequence order
- World state only modified through rules.Step()
- Tick duration monitored and logged

---

## Constants

**Session Constants**:
- `TICK_RATE = 30.0` – Tick rate (Hz)
- `DT = 1.0 / 30.0` – Time step (seconds, ~0.0333)
- `G = 1.0` – Gravitational constant (passed to rules)
- `A_MAX = 100.0` – Maximum acceleration (passed to rules)
- `PICKUP_RADIUS = 15.0` – Pallet pickup radius (passed to rules)
- `MAX_QUEUE_SIZE = 100` – Typical maximum command queue size
- `QUEUE_THRESHOLD_PERCENT = 0.5` – Queue depth threshold for logging (50%)
- `TICK_DURATION_THRESHOLD = 10ms` – Slow tick threshold for logging

---

## Ownership & Dependencies

### Session Package Ownership

- **Only `server/internal/session` may define session orchestration logic**
- Session orchestrates but does not implement physics or rules
- Session manages time and command queue, but does not handle network IO

### Dependencies

- **Imports**:
  - `entities` package (for World type)
  - `rules` package (for InputCommand and Step function)
  - `observability` package (for metrics and logging)
- **No dependencies on**: transport, proto packages
- Session is orchestration layer that composes rules

### No Duplication Rules

- **No tick loop elsewhere**: Game loop orchestration must live in `/session`
- **No command queue elsewhere**: Command queuing logic must live in `/session`
- **No time abstraction elsewhere**: Clock interface must live in `/session`
- **Session does not implement**: Physics formulas, game rules, or network IO

---

## Integration with Transport Layer

**Session Handler Pattern**:
- Transport layer creates SessionHandler that wraps Session
- SessionHandler manages:
  - Session lifecycle (start/stop)
  - Snapshot broadcasting (at 10 Hz)
  - Message routing (input/restart messages)
- SessionHandler calls Session.Run() in a loop
- SessionHandler reads world state and converts to protocol messages

**Semantics**:
- Session is pure orchestration (no IO)
- Transport layer handles all network communication
- SessionHandler bridges session and transport

---

## Notes

This spec describes the current session implementation. Key features:
- Fixed-rate tick loop at 30 Hz
- Sequence-based command queue with deduplication
- Clock abstraction for deterministic testing
- Observability integration (metrics, logging)
- Single-threaded session (one Run() call at a time)

Future extensions may include:
- Multi-threaded session support
- Rollback/snapshot management (if not already present)
- Command prediction and reconciliation
- Lag compensation

