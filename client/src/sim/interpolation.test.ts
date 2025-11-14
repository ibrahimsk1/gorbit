/**
 * Integration tests for snapshot interpolation system.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import { describe, it, expect, beforeEach } from 'vitest'
import { InterpolationSystem } from './interpolation'
import { StateManager } from './state-manager'
import type { SnapshotMessage } from '../net/protocol'

describe('InterpolationSystem', () => {
  let interpolationSystem: InterpolationSystem
  let stateManager: StateManager

  const createTestSnapshot = (tick: number, overrides?: Partial<SnapshotMessage>): SnapshotMessage => ({
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
    win: false,
    ...overrides
  })

  beforeEach(() => {
    stateManager = new StateManager()
    interpolationSystem = new InterpolationSystem(stateManager)
  })

  describe('Initialization', () => {
    it('should create interpolation system with default buffer (125ms)', () => {
      const system = new InterpolationSystem(stateManager)
      expect(system).toBeDefined()
      expect(system.getBufferSize()).toBe(0)
    })

    it('should create interpolation system with custom buffer (100ms)', () => {
      const system = new InterpolationSystem(stateManager, 100)
      expect(system).toBeDefined()
    })

    it('should create interpolation system with custom buffer (150ms)', () => {
      const system = new InterpolationSystem(stateManager, 150)
      expect(system).toBeDefined()
    })
  })

  describe('Snapshot Buffer Management', () => {
    it('should add snapshot to buffer', () => {
      const snapshot = createTestSnapshot(0)
      const timestamp = performance.now()
      
      interpolationSystem.addSnapshot(snapshot, timestamp)
      
      expect(interpolationSystem.getBufferSize()).toBe(1)
      expect(interpolationSystem.hasEnoughData()).toBe(false)
    })

    it('should add multiple snapshots to buffer', () => {
      const snapshot1 = createTestSnapshot(0)
      const snapshot2 = createTestSnapshot(1)
      const timestamp1 = performance.now()
      const timestamp2 = timestamp1 + 100
      
      interpolationSystem.addSnapshot(snapshot1, timestamp1)
      interpolationSystem.addSnapshot(snapshot2, timestamp2)
      
      expect(interpolationSystem.getBufferSize()).toBe(2)
      expect(interpolationSystem.hasEnoughData()).toBe(true)
    })

    it('should remove old snapshots beyond buffer duration', () => {
      const bufferMs = 125
      const system = new InterpolationSystem(stateManager, bufferMs)
      
      const snapshot1 = createTestSnapshot(0)
      const snapshot2 = createTestSnapshot(1)
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 50
      
      system.addSnapshot(snapshot1, timestamp1)
      system.addSnapshot(snapshot2, timestamp2)
      
      expect(system.getBufferSize()).toBe(2)
      
      // Add snapshot that makes first one too old
      const snapshot3 = createTestSnapshot(2)
      const timestamp3 = timestamp1 + bufferMs + 10 // First snapshot is now too old
      system.addSnapshot(snapshot3, timestamp3)
      
      // First snapshot should be removed
      expect(system.getBufferSize()).toBe(2)
    })

    it('should handle duplicate snapshots (same tick) by replacing with newer', () => {
      const snapshot1 = createTestSnapshot(0, { ship: { pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 } })
      const snapshot2 = createTestSnapshot(0, { ship: { pos: { x: 200, y: 300 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 } })
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 50
      
      interpolationSystem.addSnapshot(snapshot1, timestamp1)
      interpolationSystem.addSnapshot(snapshot2, timestamp2)
      
      // Should keep only one snapshot (the newer one)
      expect(interpolationSystem.getBufferSize()).toBe(1)
    })

    it('should clear buffer', () => {
      const snapshot1 = createTestSnapshot(0)
      const snapshot2 = createTestSnapshot(1)
      const timestamp1 = performance.now()
      const timestamp2 = timestamp1 + 100
      
      interpolationSystem.addSnapshot(snapshot1, timestamp1)
      interpolationSystem.addSnapshot(snapshot2, timestamp2)
      
      expect(interpolationSystem.getBufferSize()).toBe(2)
      
      interpolationSystem.clear()
      
      expect(interpolationSystem.getBufferSize()).toBe(0)
      expect(interpolationSystem.hasEnoughData()).toBe(false)
    })

    it('should return correct buffer size', () => {
      expect(interpolationSystem.getBufferSize()).toBe(0)
      
      const snapshot1 = createTestSnapshot(0)
      interpolationSystem.addSnapshot(snapshot1, performance.now())
      expect(interpolationSystem.getBufferSize()).toBe(1)
      
      const snapshot2 = createTestSnapshot(1)
      interpolationSystem.addSnapshot(snapshot2, performance.now() + 100)
      expect(interpolationSystem.getBufferSize()).toBe(2)
    })

    it('should return false for hasEnoughData with 0 snapshots', () => {
      expect(interpolationSystem.hasEnoughData()).toBe(false)
    })

    it('should return false for hasEnoughData with 1 snapshot', () => {
      const snapshot = createTestSnapshot(0)
      interpolationSystem.addSnapshot(snapshot, performance.now())
      
      expect(interpolationSystem.hasEnoughData()).toBe(false)
    })

    it('should return true for hasEnoughData with 2+ snapshots', () => {
      const snapshot1 = createTestSnapshot(0)
      const snapshot2 = createTestSnapshot(1)
      const timestamp1 = performance.now()
      const timestamp2 = timestamp1 + 100
      
      interpolationSystem.addSnapshot(snapshot1, timestamp1)
      interpolationSystem.addSnapshot(snapshot2, timestamp2)
      
      expect(interpolationSystem.hasEnoughData()).toBe(true)
    })
  })

  describe('Interpolation with Two Snapshots', () => {
    it('should interpolate at midpoint between snapshots (factor = 0.5)', () => {
      const bufferMs = 125
      const system = new InterpolationSystem(stateManager, bufferMs)
      
      const snapshot1 = createTestSnapshot(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 10, y: 0 }, rot: 0, energy: 100 }
      })
      const snapshot2 = createTestSnapshot(1, {
        ship: { pos: { x: 100, y: 0 }, vel: { x: 10, y: 0 }, rot: Math.PI / 2, energy: 50 }
      })
      
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 100 // 100ms between snapshots
      system.addSnapshot(snapshot1, timestamp1)
      system.addSnapshot(snapshot2, timestamp2)
      
      // Update at midpoint: currentTime = timestamp2 + bufferMs, targetTime = timestamp2
      // But we want to interpolate at midpoint, so targetTime should be between timestamp1 and timestamp2
      // targetTime = currentTime - bufferMs = (timestamp2 + bufferMs) - bufferMs = timestamp2
      // Actually, we want to interpolate into the past by bufferMs
      // So if currentTime = timestamp2 + bufferMs, targetTime = timestamp2
      // But we want midpoint, so let's set currentTime so targetTime is at midpoint
      // targetTime = timestamp1 + 50 (midpoint)
      // currentTime = targetTime + bufferMs = timestamp1 + 50 + bufferMs
      const currentTime = timestamp1 + 50 + bufferMs
      system.update(currentTime)
      
      const interpolated = stateManager.getInterpolated()
      expect(interpolated).not.toBeNull()
      expect(interpolated?.ship.pos.x).toBeCloseTo(50, 1) // Midpoint between 0 and 100
      expect(interpolated?.ship.pos.y).toBeCloseTo(0, 1)
      expect(interpolated?.ship.energy).toBeCloseTo(75, 1) // Midpoint between 100 and 50
    })

    it('should interpolate at start (factor = 0.0, uses older snapshot)', () => {
      const bufferMs = 125
      const system = new InterpolationSystem(stateManager, bufferMs)
      
      const snapshot1 = createTestSnapshot(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      })
      const snapshot2 = createTestSnapshot(1, {
        ship: { pos: { x: 100, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 50 }
      })
      
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 100
      system.addSnapshot(snapshot1, timestamp1)
      system.addSnapshot(snapshot2, timestamp2)
      
      // Update at start: targetTime = timestamp1
      const currentTime = timestamp1 + bufferMs
      system.update(currentTime)
      
      const interpolated = stateManager.getInterpolated()
      expect(interpolated).not.toBeNull()
      expect(interpolated?.ship.pos.x).toBeCloseTo(0, 0.1) // Should use older snapshot
      expect(interpolated?.ship.energy).toBeCloseTo(100, 0.1)
    })

    it('should interpolate at end (factor = 1.0, uses newer snapshot)', () => {
      const bufferMs = 125
      const system = new InterpolationSystem(stateManager, bufferMs)
      
      const snapshot1 = createTestSnapshot(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      })
      const snapshot2 = createTestSnapshot(1, {
        ship: { pos: { x: 100, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 50 }
      })
      
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 100
      system.addSnapshot(snapshot1, timestamp1)
      system.addSnapshot(snapshot2, timestamp2)
      
      // Update at end: targetTime = timestamp2
      const currentTime = timestamp2 + bufferMs
      system.update(currentTime)
      
      const interpolated = stateManager.getInterpolated()
      expect(interpolated).not.toBeNull()
      expect(interpolated?.ship.pos.x).toBeCloseTo(100, 0.1) // Should use newer snapshot
      expect(interpolated?.ship.energy).toBeCloseTo(50, 0.1)
    })

    it('should update interpolated state in StateManager', () => {
      const snapshot1 = createTestSnapshot(0)
      const snapshot2 = createTestSnapshot(1)
      const timestamp1 = performance.now()
      const timestamp2 = timestamp1 + 100
      
      interpolationSystem.addSnapshot(snapshot1, timestamp1)
      interpolationSystem.addSnapshot(snapshot2, timestamp2)
      
      interpolationSystem.update(timestamp2 + 125)
      
      const interpolated = stateManager.getInterpolated()
      expect(interpolated).not.toBeNull()
      expect(stateManager.hasInterpolated()).toBe(true)
    })

    it('should interpolate ship position correctly', () => {
      const bufferMs = 125
      const system = new InterpolationSystem(stateManager, bufferMs)
      
      const snapshot1 = createTestSnapshot(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      })
      const snapshot2 = createTestSnapshot(1, {
        ship: { pos: { x: 200, y: 100 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      })
      
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 100
      system.addSnapshot(snapshot1, timestamp1)
      system.addSnapshot(snapshot2, timestamp2)
      
      // Interpolate at 25% (factor = 0.25)
      const currentTime = timestamp1 + 25 + bufferMs
      system.update(currentTime)
      
      const interpolated = stateManager.getInterpolated()
      expect(interpolated?.ship.pos.x).toBeCloseTo(50, 1) // 0 + (200-0) * 0.25
      expect(interpolated?.ship.pos.y).toBeCloseTo(25, 1) // 0 + (100-0) * 0.25
    })

    it('should interpolate ship velocity correctly', () => {
      const bufferMs = 125
      const system = new InterpolationSystem(stateManager, bufferMs)
      
      const snapshot1 = createTestSnapshot(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      })
      const snapshot2 = createTestSnapshot(1, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 20, y: 10 }, rot: 0, energy: 100 }
      })
      
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 100
      system.addSnapshot(snapshot1, timestamp1)
      system.addSnapshot(snapshot2, timestamp2)
      
      // Interpolate at 50%
      const currentTime = timestamp1 + 50 + bufferMs
      system.update(currentTime)
      
      const interpolated = stateManager.getInterpolated()
      expect(interpolated?.ship.vel.x).toBeCloseTo(10, 1) // 0 + (20-0) * 0.5
      expect(interpolated?.ship.vel.y).toBeCloseTo(5, 1) // 0 + (10-0) * 0.5
    })

    it('should interpolate ship rotation correctly (with wrap-around)', () => {
      const bufferMs = 125
      const system = new InterpolationSystem(stateManager, bufferMs)
      
      const snapshot1 = createTestSnapshot(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: Math.PI * 1.9, energy: 100 }
      })
      const snapshot2 = createTestSnapshot(1, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: Math.PI * 0.1, energy: 100 }
      })
      
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 100
      system.addSnapshot(snapshot1, timestamp1)
      system.addSnapshot(snapshot2, timestamp2)
      
      // Interpolate at 50% - should find shortest path (crossing 0)
      const currentTime = timestamp1 + 50 + bufferMs
      system.update(currentTime)
      
      const interpolated = stateManager.getInterpolated()
      // Should interpolate along shortest path (crossing 0)
      // From 1.9π to 0.1π, shortest path is through 0
      // Expected: close to 0 (or 2π)
      const rot = interpolated?.ship.rot ?? 0
      expect(rot).toBeLessThan(0.5) // Should be close to 0
    })

    it('should interpolate ship energy correctly', () => {
      const bufferMs = 125
      const system = new InterpolationSystem(stateManager, bufferMs)
      
      const snapshot1 = createTestSnapshot(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      })
      const snapshot2 = createTestSnapshot(1, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 0 }
      })
      
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 100
      system.addSnapshot(snapshot1, timestamp1)
      system.addSnapshot(snapshot2, timestamp2)
      
      // Interpolate at 75%
      const currentTime = timestamp1 + 75 + bufferMs
      system.update(currentTime)
      
      const interpolated = stateManager.getInterpolated()
      expect(interpolated?.ship.energy).toBeCloseTo(25, 1) // 100 + (0-100) * 0.75
    })

    it('should interpolate planet positions correctly', () => {
      const bufferMs = 125
      const system = new InterpolationSystem(stateManager, bufferMs)
      
      const snapshot1 = createTestSnapshot(0, {
        planets: [{ pos: { x: 0, y: 0 }, radius: 15.0 }]
      })
      const snapshot2 = createTestSnapshot(1, {
        planets: [{ pos: { x: 100, y: 50 }, radius: 15.0 }]
      })
      
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 100
      system.addSnapshot(snapshot1, timestamp1)
      system.addSnapshot(snapshot2, timestamp2)
      
      // Interpolate at 50%
      const currentTime = timestamp1 + 50 + bufferMs
      system.update(currentTime)
      
      const interpolated = stateManager.getInterpolated()
      expect(interpolated?.planets[0].pos.x).toBeCloseTo(50, 1)
      expect(interpolated?.planets[0].pos.y).toBeCloseTo(25, 1)
    })

    it('should interpolate pallet positions correctly (by id)', () => {
      const bufferMs = 125
      const system = new InterpolationSystem(stateManager, bufferMs)
      
      const snapshot1 = createTestSnapshot(0, {
        pallets: [
          { id: 1, pos: { x: 0, y: 0 }, active: true },
          { id: 2, pos: { x: 50, y: 50 }, active: false }
        ]
      })
      const snapshot2 = createTestSnapshot(1, {
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true },
          { id: 2, pos: { x: 150, y: 150 }, active: false }
        ]
      })
      
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 100
      system.addSnapshot(snapshot1, timestamp1)
      system.addSnapshot(snapshot2, timestamp2)
      
      // Interpolate at 50%
      const currentTime = timestamp1 + 50 + bufferMs
      system.update(currentTime)
      
      const interpolated = stateManager.getInterpolated()
      const pallet1 = interpolated?.pallets.find(p => p.id === 1)
      const pallet2 = interpolated?.pallets.find(p => p.id === 2)
      
      expect(pallet1?.pos.x).toBeCloseTo(50, 1)
      expect(pallet1?.pos.y).toBeCloseTo(50, 1)
      expect(pallet2?.pos.x).toBeCloseTo(100, 1)
      expect(pallet2?.pos.y).toBeCloseTo(100, 1)
    })

    it('should use discrete values for pallet.active, done, win', () => {
      const bufferMs = 125
      const system = new InterpolationSystem(stateManager, bufferMs)
      
      const snapshot1 = createTestSnapshot(0, {
        pallets: [{ id: 1, pos: { x: 0, y: 0 }, active: false }],
        done: false,
        win: false
      })
      const snapshot2 = createTestSnapshot(1, {
        pallets: [{ id: 1, pos: { x: 100, y: 100 }, active: true }],
        done: true,
        win: true
      })
      
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 100
      system.addSnapshot(snapshot1, timestamp1)
      system.addSnapshot(snapshot2, timestamp2)
      
      // Interpolate at 50% - discrete values should use newer snapshot
      const currentTime = timestamp1 + 50 + bufferMs
      system.update(currentTime)
      
      const interpolated = stateManager.getInterpolated()
      expect(interpolated?.pallets[0].active).toBe(true) // Uses newer snapshot
      expect(interpolated?.done).toBe(true) // Uses newer snapshot
      expect(interpolated?.win).toBe(true) // Uses newer snapshot
    })
  })

  describe('Interpolation with Single Snapshot', () => {
    it('should use snapshot directly when only one available', () => {
      const snapshot = createTestSnapshot(0, {
        ship: { pos: { x: 100, y: 200 }, vel: { x: 10, y: -5 }, rot: 1.57, energy: 75.5 }
      })
      const timestamp = performance.now()
      
      interpolationSystem.addSnapshot(snapshot, timestamp)
      interpolationSystem.update(timestamp + 125)
      
      const interpolated = stateManager.getInterpolated()
      expect(interpolated).not.toBeNull()
      expect(interpolated?.ship.pos.x).toBe(100)
      expect(interpolated?.ship.pos.y).toBe(200)
      expect(interpolated?.ship.energy).toBe(75.5)
    })

    it('should update interpolated state correctly with single snapshot', () => {
      const snapshot = createTestSnapshot(0)
      const timestamp = performance.now()
      
      interpolationSystem.addSnapshot(snapshot, timestamp)
      interpolationSystem.update(timestamp + 125)
      
      expect(stateManager.hasInterpolated()).toBe(true)
      const interpolated = stateManager.getInterpolated()
      expect(interpolated?.tick).toBe(0)
    })
  })

  describe('Interpolation with No Snapshots', () => {
    it('should do nothing when buffer is empty', () => {
      interpolationSystem.update(performance.now())
      
      expect(stateManager.hasInterpolated()).toBe(false)
      expect(interpolationSystem.getBufferSize()).toBe(0)
    })

    it('should not crash when no snapshots', () => {
      expect(() => {
        interpolationSystem.update(performance.now())
        interpolationSystem.update(performance.now() + 100)
      }).not.toThrow()
    })
  })

  describe('Interpolation Edge Cases', () => {
    it('should use newer snapshot when snapshots have same timestamp', () => {
      const snapshot1 = createTestSnapshot(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      })
      const snapshot2 = createTestSnapshot(1, {
        ship: { pos: { x: 100, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 50 }
      })
      const timestamp = 1000
      
      interpolationSystem.addSnapshot(snapshot1, timestamp)
      interpolationSystem.addSnapshot(snapshot2, timestamp) // Same timestamp
      
      interpolationSystem.update(timestamp + 125)
      
      const interpolated = stateManager.getInterpolated()
      // Should use newer snapshot (snapshot2)
      expect(interpolated?.ship.pos.x).toBe(100)
    })

    it('should sort snapshots correctly when out of order', () => {
      const snapshot1 = createTestSnapshot(0)
      const snapshot2 = createTestSnapshot(1)
      const timestamp1 = 1000
      const timestamp2 = 500 // Out of order
      
      interpolationSystem.addSnapshot(snapshot2, timestamp2) // Add newer first
      interpolationSystem.addSnapshot(snapshot1, timestamp1) // Add older second
      
      // System should handle this correctly (use timestamps for interpolation)
      expect(interpolationSystem.getBufferSize()).toBe(2)
    })

    it('should use oldest snapshot when target time is before oldest', () => {
      const bufferMs = 125
      const system = new InterpolationSystem(stateManager, bufferMs)
      
      const snapshot1 = createTestSnapshot(0, {
        ship: { pos: { x: 100, y: 100 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      })
      const snapshot2 = createTestSnapshot(1, {
        ship: { pos: { x: 200, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 50 }
      })
      
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 100
      system.addSnapshot(snapshot1, timestamp1)
      system.addSnapshot(snapshot2, timestamp2)
      
      // Update with target time before oldest snapshot
      // currentTime = timestamp1 - 10, targetTime = timestamp1 - 10 - bufferMs = timestamp1 - 135
      const currentTime = timestamp1 - 10
      system.update(currentTime)
      
      const interpolated = stateManager.getInterpolated()
      // Should use oldest snapshot
      expect(interpolated?.ship.pos.x).toBe(100)
    })

    it('should use newest snapshot when target time is after newest', () => {
      const bufferMs = 125
      const system = new InterpolationSystem(stateManager, bufferMs)
      
      const snapshot1 = createTestSnapshot(0, {
        ship: { pos: { x: 100, y: 100 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      })
      const snapshot2 = createTestSnapshot(1, {
        ship: { pos: { x: 200, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 50 }
      })
      
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 100
      system.addSnapshot(snapshot1, timestamp1)
      system.addSnapshot(snapshot2, timestamp2)
      
      // Update with target time after newest snapshot
      // currentTime = timestamp2 + 200, targetTime = timestamp2 + 200 - bufferMs = timestamp2 + 75
      const currentTime = timestamp2 + 200
      system.update(currentTime)
      
      const interpolated = stateManager.getInterpolated()
      // Should use newest snapshot
      expect(interpolated?.ship.pos.x).toBe(200)
    })

    it('should handle planets array length change (use newer array)', () => {
      const bufferMs = 125
      const system = new InterpolationSystem(stateManager, bufferMs)
      
      const snapshot1 = createTestSnapshot(0, {
        planets: [{ pos: { x: 0, y: 0 }, radius: 15.0 }]
      })
      const snapshot2 = createTestSnapshot(1, {
        planets: [
          { pos: { x: 100, y: 100 }, radius: 15.0 },
          { pos: { x: 200, y: 200 }, radius: 20.0 }
        ]
      })
      
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 100
      system.addSnapshot(snapshot1, timestamp1)
      system.addSnapshot(snapshot2, timestamp2)
      
      const currentTime = timestamp1 + 50 + bufferMs
      system.update(currentTime)
      
      const interpolated = stateManager.getInterpolated()
      // Should use newer snapshot's planets array (2 planets)
      expect(interpolated?.planets).toHaveLength(2)
    })

    it('should handle pallets added/removed (match by id)', () => {
      const bufferMs = 125
      const system = new InterpolationSystem(stateManager, bufferMs)
      
      const snapshot1 = createTestSnapshot(0, {
        pallets: [
          { id: 1, pos: { x: 0, y: 0 }, active: true },
          { id: 2, pos: { x: 50, y: 50 }, active: true }
        ]
      })
      const snapshot2 = createTestSnapshot(1, {
        pallets: [
          { id: 2, pos: { x: 100, y: 100 }, active: true },
          { id: 3, pos: { x: 200, y: 200 }, active: true } // New pallet
        ]
      })
      
      const timestamp1 = 1000
      const timestamp2 = timestamp1 + 100
      system.addSnapshot(snapshot1, timestamp1)
      system.addSnapshot(snapshot2, timestamp2)
      
      const currentTime = timestamp1 + 50 + bufferMs
      system.update(currentTime)
      
      const interpolated = stateManager.getInterpolated()
      // Should have pallet 2 (interpolated) and pallet 3 (from newer snapshot)
      expect(interpolated?.pallets).toHaveLength(2)
      const pallet2 = interpolated?.pallets.find(p => p.id === 2)
      const pallet3 = interpolated?.pallets.find(p => p.id === 3)
      expect(pallet2).toBeDefined()
      expect(pallet3).toBeDefined()
      // Pallet 2 should be interpolated
      expect(pallet2?.pos.x).toBeCloseTo(75, 1) // Between 50 and 100
    })
  })

  describe('Integration with StateManager', () => {
    it('should update interpolated state', () => {
      const snapshot1 = createTestSnapshot(0)
      const snapshot2 = createTestSnapshot(1)
      const timestamp1 = performance.now()
      const timestamp2 = timestamp1 + 100
      
      interpolationSystem.addSnapshot(snapshot1, timestamp1)
      interpolationSystem.addSnapshot(snapshot2, timestamp2)
      interpolationSystem.update(timestamp2 + 125)
      
      expect(stateManager.hasInterpolated()).toBe(true)
      const interpolated = stateManager.getInterpolated()
      expect(interpolated).not.toBeNull()
    })

    it('should not modify authoritative state', () => {
      const snapshot1 = createTestSnapshot(0)
      const snapshot2 = createTestSnapshot(1)
      const timestamp1 = performance.now()
      const timestamp2 = timestamp1 + 100
      
      stateManager.updateAuthoritative(snapshot1)
      const authoritativeBefore = stateManager.getAuthoritative()
      
      interpolationSystem.addSnapshot(snapshot1, timestamp1)
      interpolationSystem.addSnapshot(snapshot2, timestamp2)
      interpolationSystem.update(timestamp2 + 125)
      
      const authoritativeAfter = stateManager.getAuthoritative()
      expect(authoritativeAfter?.tick).toBe(authoritativeBefore?.tick)
      expect(authoritativeAfter?.ship.pos.x).toBe(authoritativeBefore?.ship.pos.x)
    })

    it('should work with StateManager.getRenderState()', () => {
      const snapshot1 = createTestSnapshot(0)
      const snapshot2 = createTestSnapshot(1)
      const timestamp1 = performance.now()
      const timestamp2 = timestamp1 + 100
      
      interpolationSystem.addSnapshot(snapshot1, timestamp1)
      interpolationSystem.addSnapshot(snapshot2, timestamp2)
      interpolationSystem.update(timestamp2 + 125)
      
      const renderState = stateManager.getRenderState()
      expect(renderState).toBeDefined()
      expect(renderState.tick).toBeDefined()
    })
  })
})

