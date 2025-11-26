/**
 * Test utilities and helpers for integration tests.
 * 
 * Provides common test data generators and setup helpers.
 */

import type { SnapshotMessage, ShipSnapshot, PlanetSnapshot, PalletSnapshot } from '../net/protocol'
import type { GameState } from '../sim/state-manager'

/**
 * Creates a test SnapshotMessage with default values.
 * v1 multiplayer format: uses ships array, includes myShipId and worldBounds.
 */
export function createTestSnapshot(tick: number, overrides?: Partial<SnapshotMessage>): SnapshotMessage {
  return {
    t: 'snapshot',
    tick,
    ships: [
      { id: 1, pos: { x: 10.0, y: 0.0 }, vel: { x: 0.0, y: 0.0 }, rot: 0.0, energy: 100.0 }
    ],
    planets: [
      { id: 1, pos: { x: 0.0, y: 0.0 }, radius: 50.0 }
    ],
    pallets: [],
    worldBounds: { width: 2000, height: 2000 },
    myShipId: 1,
    done: false,
    win: false,
    ...overrides
  }
}

/**
 * Creates a test GameState with default values.
 * v1 multiplayer format: uses ships array, includes myShipId and worldBounds.
 */
export function createTestState(tick: number, overrides?: Partial<GameState>): GameState {
  return {
    tick,
    ships: [
      { id: 1, pos: { x: 10.0, y: 0.0 }, vel: { x: 0.0, y: 0.0 }, rot: 0.0, energy: 100.0 }
    ],
    planets: [
      { id: 1, pos: { x: 0.0, y: 0.0 }, radius: 50.0 }
    ],
    pallets: [],
    worldBounds: { width: 2000, height: 2000 },
    myShipId: 1,
    done: false,
    win: false,
    ...overrides
  }
}

/**
 * Creates a test ShipSnapshot.
 * v1 format includes id field.
 */
export function createTestShip(overrides?: Partial<ShipSnapshot>): ShipSnapshot {
  return {
    id: 1,
    pos: { x: 10.0, y: 0.0 },
    vel: { x: 0.0, y: 0.0 },
    rot: 0.0,
    energy: 100.0,
    ...overrides
  }
}

/**
 * Creates a test PlanetSnapshot.
 * v1 format includes id field.
 */
export function createTestPlanet(overrides?: Partial<PlanetSnapshot>): PlanetSnapshot {
  return {
    id: 1,
    pos: { x: 0.0, y: 0.0 },
    radius: 50.0,
    ...overrides
  }
}

/**
 * Creates a test PalletSnapshot.
 */
export function createTestPallet(id: number, overrides?: Partial<PalletSnapshot>): PalletSnapshot {
  return {
    id,
    pos: { x: 0.0, y: 0.0 },
    active: true,
    ...overrides
  }
}

/**
 * Calculates the distance between two positions.
 */
export function distance(pos1: { x: number; y: number }, pos2: { x: number; y: number }): number {
  const dx = pos2.x - pos1.x
  const dy = pos2.y - pos1.y
  return Math.sqrt(dx * dx + dy * dy)
}

/**
 * Calculates the difference between two rotations (in radians).
 * Returns the smallest angle difference.
 */
export function rotationDifference(rot1: number, rot2: number): number {
  let diff = rot2 - rot1
  // Normalize to [-π, π]
  while (diff > Math.PI) diff -= 2 * Math.PI
  while (diff < -Math.PI) diff += 2 * Math.PI
  return Math.abs(diff)
}

/**
 * Checks if two states are approximately equal (within threshold).
 * v1 multiplayer format: compares player's ship by myShipId.
 */
export function statesApproximatelyEqual(
  state1: GameState,
  state2: GameState,
  positionThreshold: number = 5.0,
  rotationThreshold: number = 0.1,
  energyThreshold: number = 1.0
): { equal: boolean; errors: { position?: number; rotation?: number; energy?: number } } {
  const errors: { position?: number; rotation?: number; energy?: number } = {}
  
  // Find player's ship in both states
  const ship1 = state1.ships.find(s => s.id === state1.myShipId)
  const ship2 = state2.ships.find(s => s.id === state2.myShipId)
  
  if (!ship1 || !ship2) {
    // Ships not found, return not equal
    return { equal: false, errors: { position: Infinity } }
  }
  
  const posError = distance(ship1.pos, ship2.pos)
  if (posError > positionThreshold) {
    errors.position = posError
  }
  
  const rotError = rotationDifference(ship1.rot, ship2.rot)
  if (rotError > rotationThreshold) {
    errors.rotation = rotError
  }
  
  const energyError = Math.abs(ship1.energy - ship2.energy)
  if (energyError > energyThreshold) {
    errors.energy = energyError
  }
  
  return {
    equal: Object.keys(errors).length === 0,
    errors
  }
}

/**
 * Mock WebSocket class for testing.
 */
export class MockWebSocket {
  url: string
  readyState: number = WebSocket.CONNECTING
  onopen: ((event: Event) => void) | null = null
  onclose: ((event: CloseEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null
  onmessage: ((event: MessageEvent) => void) | null = null
  sentMessages: string[] = []

  constructor(url: string) {
    this.url = url
    setTimeout(() => {
      this.readyState = WebSocket.OPEN
      if (this.onopen) {
        this.onopen(new Event('open'))
      }
    }, 10)
  }

  send(data: string): void {
    this.sentMessages.push(data)
  }

  close(): void {
    this.readyState = WebSocket.CLOSED
    if (this.onclose) {
      this.onclose(new CloseEvent('close'))
    }
  }

  simulateMessage(data: string): void {
    if (this.onmessage) {
      this.onmessage(new MessageEvent('message', { data }))
    }
  }

  simulateError(): void {
    if (this.onerror) {
      this.onerror(new Event('error'))
    }
  }
}

