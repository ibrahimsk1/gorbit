/**
 * Integration tests for state manager coordinating authoritative, predicted, and interpolated states.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import { describe, it, expect, beforeEach } from 'vitest'
import { StateManager } from './state-manager'
import type { SnapshotMessage } from '../net/protocol'

describe('StateManager', () => {
  let stateManager: StateManager

  const createTestSnapshot = (tick: number): SnapshotMessage => ({
    t: 'snapshot',
    tick,
    ship: {
      pos: { x: 100, y: 200 },
      vel: { x: 10, y: -5 },
      rot: 1.57,
      energy: 75.5
    },
    planets: [
      { pos: { x: 0, y: 0 }, radius: 15.0 }
    ],
    pallets: [
      { id: 1, pos: { x: 50, y: 50 }, active: true },
      { id: 2, pos: { x: -50, y: -50 }, active: false }
    ],
    done: false,
    win: false
  })

  beforeEach(() => {
    stateManager = new StateManager()
  })

  describe('Authoritative State Management', () => {
    it('should store authoritative state from SnapshotMessage', () => {
      const snapshot = createTestSnapshot(42)
      stateManager.updateAuthoritative(snapshot)

      const authoritative = stateManager.getAuthoritative()
      expect(authoritative).not.toBeNull()
      expect(authoritative?.tick).toBe(42)
      expect(authoritative?.ship.pos.x).toBe(100)
      expect(authoritative?.ship.pos.y).toBe(200)
      expect(authoritative?.planets).toHaveLength(1)
      expect(authoritative?.pallets).toHaveLength(2)
    })

    it('should return null for authoritative state before any update', () => {
      expect(stateManager.getAuthoritative()).toBeNull()
    })

    it('should support multiple planets in authoritative state', () => {
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 0, y: 0 }, radius: 5.0 },
          { pos: { x: 100, y: 100 }, radius: 3.0 }
        ],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateAuthoritative(snapshot)
      const authoritative = stateManager.getAuthoritative()
      expect(authoritative?.planets).toHaveLength(2)
      expect(authoritative?.planets[0].radius).toBe(5.0)
      expect(authoritative?.planets[1].radius).toBe(3.0)
    })

    it('should support empty planets array in authoritative state', () => {
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateAuthoritative(snapshot)
      const authoritative = stateManager.getAuthoritative()
      expect(authoritative?.planets).toHaveLength(0)
    })

    it('should support multiple pallets in authoritative state', () => {
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [
          { id: 1, pos: { x: 10, y: 10 }, active: true },
          { id: 2, pos: { x: 20, y: 20 }, active: true },
          { id: 3, pos: { x: 30, y: 30 }, active: false }
        ],
        done: false,
        win: false
      }

      stateManager.updateAuthoritative(snapshot)
      const authoritative = stateManager.getAuthoritative()
      expect(authoritative?.pallets).toHaveLength(3)
      expect(authoritative?.pallets[0].id).toBe(1)
      expect(authoritative?.pallets[2].id).toBe(3)
    })

    it('should update authoritative state with new snapshot', () => {
      const snapshot1 = createTestSnapshot(10)
      stateManager.updateAuthoritative(snapshot1)
      expect(stateManager.getAuthoritative()?.tick).toBe(10)

      const snapshot2 = createTestSnapshot(20)
      stateManager.updateAuthoritative(snapshot2)
      expect(stateManager.getAuthoritative()?.tick).toBe(20)
    })
  })

  describe('Predicted State Management', () => {
    it('should store predicted state separately from authoritative', () => {
      const authoritativeSnapshot = createTestSnapshot(42)
      stateManager.updateAuthoritative(authoritativeSnapshot)

      const predictedState = {
        tick: 43,
        ship: {
          pos: { x: 110, y: 195 },
          vel: { x: 10, y: -5 },
          rot: 1.6,
          energy: 74.5
        },
        planets: [
          { pos: { x: 0, y: 0 }, radius: 15.0 }
        ],
        pallets: [
          { id: 1, pos: { x: 50, y: 50 }, active: false }
        ],
        done: false,
        win: false
      }

      stateManager.updatePredicted(predictedState)

      const authoritative = stateManager.getAuthoritative()
      const predicted = stateManager.getPredicted()

      expect(authoritative?.tick).toBe(42)
      expect(predicted?.tick).toBe(43)
      expect(authoritative?.ship.pos.x).toBe(100)
      expect(predicted?.ship.pos.x).toBe(110)
      expect(authoritative?.pallets).toHaveLength(2)
      expect(predicted?.pallets).toHaveLength(1)
    })

    it('should return null for predicted state before any update', () => {
      expect(stateManager.getPredicted()).toBeNull()
    })

    it('should support array-based entities in predicted state', () => {
      const predictedState = {
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 0, y: 0 }, radius: 5.0 },
          { pos: { x: 100, y: 100 }, radius: 3.0 }
        ],
        pallets: [
          { id: 1, pos: { x: 10, y: 10 }, active: true }
        ],
        done: false,
        win: false
      }

      stateManager.updatePredicted(predictedState)
      const predicted = stateManager.getPredicted()
      expect(predicted?.planets).toHaveLength(2)
      expect(predicted?.pallets).toHaveLength(1)
    })
  })

  describe('Interpolated State Management', () => {
    it('should store interpolated state separately', () => {
      const interpolatedState = {
        tick: 42.5,
        ship: {
          pos: { x: 105, y: 197.5 },
          vel: { x: 10, y: -5 },
          rot: 1.585,
          energy: 75.0
        },
        planets: [
          { pos: { x: 0, y: 0 }, radius: 15.0 }
        ],
        pallets: [
          { id: 1, pos: { x: 50, y: 50 }, active: true }
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(interpolatedState)

      const interpolated = stateManager.getInterpolated()
      expect(interpolated).not.toBeNull()
      expect(interpolated?.tick).toBe(42.5)
      expect(interpolated?.ship.pos.x).toBe(105)
    })

    it('should return null for interpolated state before any update', () => {
      expect(stateManager.getInterpolated()).toBeNull()
    })

    it('should support array-based entities in interpolated state', () => {
      const interpolatedState = {
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 0, y: 0 }, radius: 5.0 }
        ],
        pallets: [
          { id: 1, pos: { x: 10, y: 10 }, active: true },
          { id: 2, pos: { x: 20, y: 20 }, active: false }
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(interpolatedState)
      const interpolated = stateManager.getInterpolated()
      expect(interpolated?.planets).toHaveLength(1)
      expect(interpolated?.pallets).toHaveLength(2)
    })
  })

  describe('State Manager Coordination', () => {
    it('should allow all three state layers to exist independently', () => {
      const authoritativeSnapshot = createTestSnapshot(40)
      stateManager.updateAuthoritative(authoritativeSnapshot)

      const predictedState = {
        tick: 41,
        ship: { pos: { x: 110, y: 195 }, vel: { x: 10, y: -5 }, rot: 1.6, energy: 74.5 },
        planets: [{ pos: { x: 0, y: 0 }, radius: 15.0 }],
        pallets: [{ id: 1, pos: { x: 50, y: 50 }, active: false }],
        done: false,
        win: false
      }
      stateManager.updatePredicted(predictedState)

      const interpolatedState = {
        tick: 40.5,
        ship: { pos: { x: 105, y: 197.5 }, vel: { x: 10, y: -5 }, rot: 1.585, energy: 75.0 },
        planets: [{ pos: { x: 0, y: 0 }, radius: 15.0 }],
        pallets: [{ id: 1, pos: { x: 50, y: 50 }, active: true }],
        done: false,
        win: false
      }
      stateManager.updateInterpolated(interpolatedState)

      expect(stateManager.getAuthoritative()?.tick).toBe(40)
      expect(stateManager.getPredicted()?.tick).toBe(41)
      expect(stateManager.getInterpolated()?.tick).toBe(40.5)
    })

    it('should return interpolated state in getRenderState() when available', () => {
      const authoritativeSnapshot = createTestSnapshot(40)
      stateManager.updateAuthoritative(authoritativeSnapshot)

      const interpolatedState = {
        tick: 40.5,
        ship: { pos: { x: 105, y: 197.5 }, vel: { x: 10, y: -5 }, rot: 1.585, energy: 75.0 },
        planets: [{ pos: { x: 0, y: 0 }, radius: 15.0 }],
        pallets: [{ id: 1, pos: { x: 50, y: 50 }, active: true }],
        done: false,
        win: false
      }
      stateManager.updateInterpolated(interpolatedState)

      const renderState = stateManager.getRenderState()
      expect(renderState.tick).toBe(40.5)
      expect(renderState.ship.pos.x).toBe(105)
    })

    it('should fall back to authoritative state in getRenderState() when interpolated is not available', () => {
      const authoritativeSnapshot = createTestSnapshot(40)
      stateManager.updateAuthoritative(authoritativeSnapshot)

      const renderState = stateManager.getRenderState()
      expect(renderState.tick).toBe(40)
      expect(renderState.ship.pos.x).toBe(100)
    })

    it('should reset all states', () => {
      const snapshot = createTestSnapshot(42)
      stateManager.updateAuthoritative(snapshot)
      stateManager.updatePredicted({
        tick: 43,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      })
      stateManager.updateInterpolated({
        tick: 42.5,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      })

      stateManager.reset()

      expect(stateManager.getAuthoritative()).toBeNull()
      expect(stateManager.getPredicted()).toBeNull()
      expect(stateManager.getInterpolated()).toBeNull()
    })

    it('should check if states exist', () => {
      expect(stateManager.hasAuthoritative()).toBe(false)
      expect(stateManager.hasPredicted()).toBe(false)
      expect(stateManager.hasInterpolated()).toBe(false)

      const snapshot = createTestSnapshot(42)
      stateManager.updateAuthoritative(snapshot)
      expect(stateManager.hasAuthoritative()).toBe(true)

      stateManager.updatePredicted({
        tick: 43,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      })
      expect(stateManager.hasPredicted()).toBe(true)

      stateManager.updateInterpolated({
        tick: 42.5,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      })
      expect(stateManager.hasInterpolated()).toBe(true)
    })
  })

  describe('State Conversion', () => {
    it('should convert SnapshotMessage to GameState correctly', () => {
      const snapshot = createTestSnapshot(42)
      stateManager.updateAuthoritative(snapshot)

      const authoritative = stateManager.getAuthoritative()
      expect(authoritative).not.toBeNull()
      expect(authoritative?.tick).toBe(snapshot.tick)
      expect(authoritative?.ship).toEqual(snapshot.ship)
      expect(authoritative?.planets).toEqual(snapshot.planets)
      expect(authoritative?.pallets).toEqual(snapshot.pallets)
      expect(authoritative?.done).toBe(snapshot.done)
      expect(authoritative?.win).toBe(snapshot.win)
    })

    it('should preserve array structure through state updates', () => {
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 0, y: 0 }, radius: 5.0 },
          { pos: { x: 100, y: 100 }, radius: 3.0 }
        ],
        pallets: [
          { id: 1, pos: { x: 10, y: 10 }, active: true },
          { id: 2, pos: { x: 20, y: 20 }, active: false }
        ],
        done: false,
        win: false
      }

      stateManager.updateAuthoritative(snapshot)
      const authoritative = stateManager.getAuthoritative()

      expect(authoritative?.planets).toHaveLength(2)
      expect(authoritative?.pallets).toHaveLength(2)
      expect(authoritative?.planets[0].radius).toBe(5.0)
      expect(authoritative?.planets[1].radius).toBe(3.0)
      expect(authoritative?.pallets[0].id).toBe(1)
      expect(authoritative?.pallets[1].id).toBe(2)
    })
  })
})

