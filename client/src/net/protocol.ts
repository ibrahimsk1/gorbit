/**
 * Protocol types matching G4 contracts.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

/**
 * Protocol version constant.
 * Used for protocol evolution and backward compatibility.
 */
export const PROTOCOL_VERSION = 1

/**
 * 2D vector snapshot for position and velocity data.
 */
export interface Vec2Snapshot {
  x: number
  y: number
}

/**
 * Ship state snapshot containing position, velocity, rotation, and energy.
 */
export interface ShipSnapshot {
  pos: Vec2Snapshot
  vel: Vec2Snapshot
  rot: number
  energy: number
}

/**
 * Planet state snapshot (replaces SunSnapshot for extensibility).
 * Supports multiple planets in the game world.
 */
export interface PlanetSnapshot {
  pos: Vec2Snapshot
  radius: number
}

/**
 * Legacy SunSnapshot type (kept for backward compatibility during transition).
 * @deprecated Use PlanetSnapshot instead
 */
export interface SunSnapshot {
  pos: Vec2Snapshot
  radius: number
}

/**
 * Energy pallet snapshot containing position and active state.
 */
export interface PalletSnapshot {
  id: number
  pos: Vec2Snapshot
  active: boolean
}

/**
 * Input command message sent from client to server.
 */
export interface InputMessage {
  t: 'input'
  seq: number
  thrust: number
  turn: number
  version?: number
}

/**
 * Restart command message sent from client to server.
 */
export interface RestartMessage {
  t: 'restart'
  version?: number
}

/**
 * Game state snapshot message sent from server to client.
 * Uses array-based entities (planets, pallets) for extensibility.
 */
export interface SnapshotMessage {
  t: 'snapshot'
  tick: number
  ship: ShipSnapshot
  planets: PlanetSnapshot[]
  pallets: PalletSnapshot[]
  done: boolean
  win: boolean
  version?: number
}

/**
 * Union type of all possible protocol messages.
 */
export type Message = InputMessage | RestartMessage | SnapshotMessage

/**
 * Type guard for Vec2Snapshot.
 */
function isValidVec2Snapshot(v: unknown): v is Vec2Snapshot {
  if (!v || typeof v !== 'object') return false
  const vec = v as Record<string, unknown>
  return (
    typeof vec.x === 'number' &&
    typeof vec.y === 'number'
  )
}

/**
 * Type guard for InputMessage.
 */
export function isInputMessage(msg: unknown): msg is InputMessage {
  if (!msg || typeof msg !== 'object') return false
  const m = msg as Record<string, unknown>
  return (
    m.t === 'input' &&
    typeof m.seq === 'number' &&
    typeof m.thrust === 'number' &&
    typeof m.turn === 'number'
  )
}

/**
 * Type guard for SnapshotMessage.
 */
export function isSnapshotMessage(msg: unknown): msg is SnapshotMessage {
  if (!msg || typeof msg !== 'object') return false
  const m = msg as Record<string, unknown>
  
  if (m.t !== 'snapshot') return false
  if (typeof m.tick !== 'number') return false
  if (typeof m.done !== 'boolean') return false
  if (typeof m.win !== 'boolean') return false
  
  if (!isValidShipSnapshot(m.ship)) return false
  if (!Array.isArray(m.planets)) return false
  if (!Array.isArray(m.pallets)) return false
  
  // Validate planets array
  for (const planet of m.planets) {
    if (!isValidPlanetSnapshot(planet)) return false
  }
  
  // Validate pallets array
  for (const pallet of m.pallets) {
    if (!isValidPalletSnapshot(pallet)) return false
  }
  
  return true
}

/**
 * Type guard for RestartMessage.
 */
export function isRestartMessage(msg: unknown): msg is RestartMessage {
  if (!msg || typeof msg !== 'object') return false
  const m = msg as Record<string, unknown>
  return m.t === 'restart'
}

/**
 * Type guard for Message union type.
 */
export function isMessage(msg: unknown): msg is Message {
  return isInputMessage(msg) || isRestartMessage(msg) || isSnapshotMessage(msg)
}

/**
 * Internal type guard for ShipSnapshot.
 */
function isValidShipSnapshot(s: unknown): s is ShipSnapshot {
  if (!s || typeof s !== 'object') return false
  const ship = s as Record<string, unknown>
  return (
    isValidVec2Snapshot(ship.pos) &&
    isValidVec2Snapshot(ship.vel) &&
    typeof ship.rot === 'number' &&
    typeof ship.energy === 'number'
  )
}

/**
 * Internal type guard for PlanetSnapshot.
 */
function isValidPlanetSnapshot(p: unknown): p is PlanetSnapshot {
  if (!p || typeof p !== 'object') return false
  const planet = p as Record<string, unknown>
  return (
    isValidVec2Snapshot(planet.pos) &&
    typeof planet.radius === 'number'
  )
}

/**
 * Internal type guard for PalletSnapshot.
 */
function isValidPalletSnapshot(p: unknown): p is PalletSnapshot {
  if (!p || typeof p !== 'object') return false
  const pallet = p as Record<string, unknown>
  return (
    typeof pallet.id === 'number' &&
    isValidVec2Snapshot(pallet.pos) &&
    typeof pallet.active === 'boolean'
  )
}

/**
 * Validates and returns a Vec2Snapshot, or null if invalid.
 */
export function validateVec2Snapshot(v: unknown): Vec2Snapshot | null {
  if (isValidVec2Snapshot(v)) {
    return v
  }
  return null
}

/**
 * Validates and returns a ShipSnapshot, or null if invalid.
 */
export function validateShipSnapshot(s: unknown): ShipSnapshot | null {
  if (isValidShipSnapshot(s)) {
    return s
  }
  return null
}

/**
 * Validates and returns a PlanetSnapshot, or null if invalid.
 */
export function validatePlanetSnapshot(p: unknown): PlanetSnapshot | null {
  if (isValidPlanetSnapshot(p)) {
    return p
  }
  return null
}

/**
 * Validates and returns a PalletSnapshot, or null if invalid.
 */
export function validatePalletSnapshot(p: unknown): PalletSnapshot | null {
  if (isValidPalletSnapshot(p)) {
    return p
  }
  return null
}

/**
 * Creates an InputMessage with the given parameters.
 */
export function createInputMessage(seq: number, thrust: number, turn: number): InputMessage {
  return {
    t: 'input',
    seq,
    thrust,
    turn
  }
}

/**
 * Creates a RestartMessage.
 */
export function createRestartMessage(): RestartMessage {
  return {
    t: 'restart'
  }
}
