/**
 * Protocol types matching G4 contracts.
 * 
 * Labels: scope:contract loop:g4-proto layer:contract
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
 * v1 format includes id field for multiplayer support.
 */
export interface ShipSnapshot {
  id: number
  pos: Vec2Snapshot
  vel: Vec2Snapshot
  rot: number
  energy: number
}

/**
 * Planet state snapshot (replaces SunSnapshot for extensibility).
 * Supports multiple planets in the game world.
 * v1 format includes id field for multiplayer support.
 */
export interface PlanetSnapshot {
  id: number
  pos: Vec2Snapshot
  radius: number
}

/**
 * World bounds defining the playable area.
 */
export interface WorldBounds {
  width: number
  height: number
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
 * Player information for room state messages.
 */
export interface PlayerInfo {
  id: number
  name: string
}

/**
 * Create room message sent from client to server.
 */
export interface CreateRoomMessage {
  t: 'createRoom'
  version?: number
}

/**
 * Join room message sent from client to server.
 */
export interface JoinRoomMessage {
  t: 'joinRoom'
  roomCode: string
  version?: number
}

/**
 * Leave room message sent from client to server.
 */
export interface LeaveRoomMessage {
  t: 'leaveRoom'
  version?: number
}

/**
 * Start match message sent from client to server (host only).
 */
export interface StartMatchMessage {
  t: 'startMatch'
  version?: number
}

/**
 * Room created message sent from server to client.
 */
export interface RoomCreatedMessage {
  t: 'roomCreated'
  roomCode: string
  version?: number
}

/**
 * Room state message sent from server to client.
 */
export interface RoomStateMessage {
  t: 'roomState'
  roomCode: string
  players: PlayerInfo[]
  state: 'lobby' | 'playing' | 'ended'
  hostId: number
  version?: number
}

/**
 * Player joined message sent from server to client.
 */
export interface PlayerJoinedMessage {
  t: 'playerJoined'
  player: PlayerInfo
  version?: number
}

/**
 * Player left message sent from server to client.
 */
export interface PlayerLeftMessage {
  t: 'playerLeft'
  playerId: number
  version?: number
}

/**
 * Match started message sent from server to client.
 */
export interface MatchStartedMessage {
  t: 'matchStarted'
  version?: number
}

/**
 * Match ended message sent from server to client.
 */
export interface MatchEndedMessage {
  t: 'matchEnded'
  winnerId?: number
  version?: number
}

/**
 * Game state snapshot message sent from server to client.
 * v1 multiplayer format: uses ships array, includes worldBounds and myShipId.
 * 
 * Backward compatibility note: v0 format used single `ship` field instead of `ships[]`.
 * Clients should handle both formats during transition period.
 */
export interface SnapshotMessage {
  t: 'snapshot'
  tick: number
  ships: ShipSnapshot[]
  planets: PlanetSnapshot[]
  pallets: PalletSnapshot[]
  worldBounds: WorldBounds
  myShipId: number
  done: boolean
  win: boolean
  version?: number
}

/**
 * Union type of all possible protocol messages.
 */
export type Message =
  | InputMessage
  | RestartMessage
  | CreateRoomMessage
  | JoinRoomMessage
  | LeaveRoomMessage
  | StartMatchMessage
  | RoomCreatedMessage
  | RoomStateMessage
  | PlayerJoinedMessage
  | PlayerLeftMessage
  | MatchStartedMessage
  | MatchEndedMessage
  | SnapshotMessage

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
 * Supports v1 multiplayer format: ships array, worldBounds, myShipId.
 */
export function isSnapshotMessage(msg: unknown): msg is SnapshotMessage {
  if (!msg || typeof msg !== 'object') return false
  const m = msg as Record<string, unknown>
  
  if (m.t !== 'snapshot') return false
  if (typeof m.tick !== 'number') return false
  if (typeof m.done !== 'boolean') return false
  if (typeof m.win !== 'boolean') return false
  
  // v1 format: ships array
  if (!Array.isArray(m.ships)) return false
  for (const ship of m.ships) {
    if (!isValidShipSnapshot(ship)) return false
  }
  
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
  
  // v1 format: worldBounds and myShipId
  if (!isValidWorldBounds(m.worldBounds)) return false
  if (typeof m.myShipId !== 'number') return false
  
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
 * Type guard for PlayerInfo.
 */
export function isPlayerInfo(p: unknown): p is PlayerInfo {
  if (!p || typeof p !== 'object') return false
  const player = p as Record<string, unknown>
  return (
    typeof player.id === 'number' &&
    typeof player.name === 'string'
  )
}

/**
 * Type guard for CreateRoomMessage.
 */
export function isCreateRoomMessage(msg: unknown): msg is CreateRoomMessage {
  if (!msg || typeof msg !== 'object') return false
  const m = msg as Record<string, unknown>
  return m.t === 'createRoom'
}

/**
 * Type guard for JoinRoomMessage.
 */
export function isJoinRoomMessage(msg: unknown): msg is JoinRoomMessage {
  if (!msg || typeof msg !== 'object') return false
  const m = msg as Record<string, unknown>
  return (
    m.t === 'joinRoom' &&
    typeof m.roomCode === 'string'
  )
}

/**
 * Type guard for LeaveRoomMessage.
 */
export function isLeaveRoomMessage(msg: unknown): msg is LeaveRoomMessage {
  if (!msg || typeof msg !== 'object') return false
  const m = msg as Record<string, unknown>
  return m.t === 'leaveRoom'
}

/**
 * Type guard for StartMatchMessage.
 */
export function isStartMatchMessage(msg: unknown): msg is StartMatchMessage {
  if (!msg || typeof msg !== 'object') return false
  const m = msg as Record<string, unknown>
  return m.t === 'startMatch'
}

/**
 * Type guard for RoomCreatedMessage.
 */
export function isRoomCreatedMessage(msg: unknown): msg is RoomCreatedMessage {
  if (!msg || typeof msg !== 'object') return false
  const m = msg as Record<string, unknown>
  return (
    m.t === 'roomCreated' &&
    typeof m.roomCode === 'string'
  )
}

/**
 * Type guard for RoomStateMessage.
 */
export function isRoomStateMessage(msg: unknown): msg is RoomStateMessage {
  if (!msg || typeof msg !== 'object') return false
  const m = msg as Record<string, unknown>
  
  if (m.t !== 'roomState') return false
  if (typeof m.roomCode !== 'string') return false
  if (typeof m.state !== 'string') return false
  if (typeof m.hostId !== 'number') return false
  
  if (!Array.isArray(m.players)) return false
  for (const player of m.players) {
    if (!isPlayerInfo(player)) return false
  }
  
  return true
}

/**
 * Type guard for PlayerJoinedMessage.
 */
export function isPlayerJoinedMessage(msg: unknown): msg is PlayerJoinedMessage {
  if (!msg || typeof msg !== 'object') return false
  const m = msg as Record<string, unknown>
  return (
    m.t === 'playerJoined' &&
    isPlayerInfo(m.player)
  )
}

/**
 * Type guard for PlayerLeftMessage.
 */
export function isPlayerLeftMessage(msg: unknown): msg is PlayerLeftMessage {
  if (!msg || typeof msg !== 'object') return false
  const m = msg as Record<string, unknown>
  return (
    m.t === 'playerLeft' &&
    typeof m.playerId === 'number'
  )
}

/**
 * Type guard for MatchStartedMessage.
 */
export function isMatchStartedMessage(msg: unknown): msg is MatchStartedMessage {
  if (!msg || typeof msg !== 'object') return false
  const m = msg as Record<string, unknown>
  return m.t === 'matchStarted'
}

/**
 * Type guard for MatchEndedMessage.
 */
export function isMatchEndedMessage(msg: unknown): msg is MatchEndedMessage {
  if (!msg || typeof msg !== 'object') return false
  const m = msg as Record<string, unknown>
  if (m.t !== 'matchEnded') return false
  // winnerId is optional
  if (m.winnerId !== undefined && typeof m.winnerId !== 'number') return false
  return true
}

/**
 * Type guard for Message union type.
 */
export function isMessage(msg: unknown): msg is Message {
  return (
    isInputMessage(msg) ||
    isRestartMessage(msg) ||
    isSnapshotMessage(msg) ||
    isCreateRoomMessage(msg) ||
    isJoinRoomMessage(msg) ||
    isLeaveRoomMessage(msg) ||
    isStartMatchMessage(msg) ||
    isRoomCreatedMessage(msg) ||
    isRoomStateMessage(msg) ||
    isPlayerJoinedMessage(msg) ||
    isPlayerLeftMessage(msg) ||
    isMatchStartedMessage(msg) ||
    isMatchEndedMessage(msg)
  )
}

/**
 * Internal type guard for ShipSnapshot.
 * v1 format includes id field.
 */
function isValidShipSnapshot(s: unknown): s is ShipSnapshot {
  if (!s || typeof s !== 'object') return false
  const ship = s as Record<string, unknown>
  return (
    typeof ship.id === 'number' &&
    isValidVec2Snapshot(ship.pos) &&
    isValidVec2Snapshot(ship.vel) &&
    typeof ship.rot === 'number' &&
    typeof ship.energy === 'number'
  )
}

/**
 * Internal type guard for PlanetSnapshot.
 * v1 format includes id field.
 */
function isValidPlanetSnapshot(p: unknown): p is PlanetSnapshot {
  if (!p || typeof p !== 'object') return false
  const planet = p as Record<string, unknown>
  return (
    typeof planet.id === 'number' &&
    isValidVec2Snapshot(planet.pos) &&
    typeof planet.radius === 'number'
  )
}

/**
 * Internal type guard for WorldBounds.
 */
function isValidWorldBounds(b: unknown): b is WorldBounds {
  if (!b || typeof b !== 'object') return false
  const bounds = b as Record<string, unknown>
  return (
    typeof bounds.width === 'number' &&
    typeof bounds.height === 'number'
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
 * Validates and returns a WorldBounds, or null if invalid.
 */
export function validateWorldBounds(b: unknown): WorldBounds | null {
  if (isValidWorldBounds(b)) {
    return b
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

/**
 * Creates a CreateRoomMessage.
 */
export function createCreateRoomMessage(): CreateRoomMessage {
  return {
    t: 'createRoom'
  }
}

/**
 * Creates a JoinRoomMessage with the given room code.
 */
export function createJoinRoomMessage(roomCode: string): JoinRoomMessage {
  return {
    t: 'joinRoom',
    roomCode
  }
}

/**
 * Creates a LeaveRoomMessage.
 */
export function createLeaveRoomMessage(): LeaveRoomMessage {
  return {
    t: 'leaveRoom'
  }
}

/**
 * Creates a StartMatchMessage.
 */
export function createStartMatchMessage(): StartMatchMessage {
  return {
    t: 'startMatch'
  }
}
