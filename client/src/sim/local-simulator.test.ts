/**
 * Integration tests for local simulation engine matching server physics/rules.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import { describe, it, expect, beforeEach } from 'vitest'
import { LocalSimulator } from './local-simulator'
import type { GameState } from './state-manager'

describe('LocalSimulator', () => {
  const epsilon = 1e-6
  const dt = 1.0 / 30.0 // 30Hz tick rate
  const G = 1.0 // Gravitational constant
  const aMax = 100.0 // Maximum acceleration
  const pickupRadius = 1.2 // Pallet pickup radius

  // Default planet mass (since PlanetSnapshot doesn't include mass)
  const DEFAULT_PLANET_MASS = 1000.0

  let simulator: LocalSimulator

  const createTestState = (overrides?: Partial<GameState>): GameState => ({
    tick: 0,
    ship: {
      pos: { x: 10.0, y: 0.0 },
      vel: { x: 0.0, y: 0.0 },
      rot: 0.0,
      energy: 100.0
    },
    planets: [
      { pos: { x: 0.0, y: 0.0 }, radius: 50.0 }
    ],
    pallets: [],
    done: false,
    win: false,
    ...overrides
  })

  beforeEach(() => {
    simulator = new LocalSimulator()
  })

  describe('Step Function - Basic Behavior', () => {
    it('should update physics with no input (gravity only)', () => {
      const state = createTestState()
      const initialPos = { ...state.ship.pos }
      const initialTick = state.tick

      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })

      // Position should change (gravity pulls ship toward planet)
      // At least one coordinate should change (ship at (10,0) and planet at (0,0) means x changes)
      const posChanged = result.ship.pos.x !== initialPos.x || result.ship.pos.y !== initialPos.y
      expect(posChanged).toBe(true)
      // Tick should increment
      expect(result.tick).toBe(initialTick + 1)
    })

    it('should apply thrust input correctly', () => {
      const state = createTestState({
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0, // Facing right
          energy: 100.0
        }
      })
      const initialVel = { ...state.ship.vel }
      const initialEnergy = state.ship.energy

      const result = simulator.step(state, { thrust: 1.0, turn: 0.0 })

      // Velocity should increase (thrust applied)
      const initialVelLength = Math.sqrt(initialVel.x ** 2 + initialVel.y ** 2)
      const finalVelLength = Math.sqrt(result.ship.vel.x ** 2 + result.ship.vel.y ** 2)
      expect(finalVelLength).toBeGreaterThan(initialVelLength)
      // Energy should decrease (drained by thrust)
      expect(result.ship.energy).toBeLessThan(initialEnergy)
    })

    it('should apply turn input correctly', () => {
      const state = createTestState({
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      const initialRot = state.ship.rot

      const result = simulator.step(state, { thrust: 0.0, turn: 1.0 })

      // Rotation should change
      expect(result.ship.rot).toBeGreaterThan(initialRot)
    })

    it('should update physics correctly (gravity + integrator)', () => {
      const state = createTestState({
        ship: {
          pos: { x: 10.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      const initialPos = { ...state.ship.pos }
      const initialVel = { ...state.ship.vel }

      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })

      // Position should change (gravity pulls toward planet)
      // Ship at (10,0) and planet at (0,0) means x changes, y stays 0
      const posChanged = result.ship.pos.x !== initialPos.x || result.ship.pos.y !== initialPos.y
      expect(posChanged).toBe(true)
      // Velocity should change (gravity accelerates)
      const velChanged = result.ship.vel.x !== initialVel.x || result.ship.vel.y !== initialVel.y
      expect(velChanged).toBe(true)
    })

    it('should process pallet pickup correctly (deactivate pallet, restore energy)', () => {
      const state = createTestState({
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 50.0
        },
        pallets: [
          { id: 1, pos: { x: 0.0, y: 0.0 }, active: true } // Ship at pallet position
        ]
      })
      const initialEnergy = state.ship.energy

      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })

      // Pallet should be deactivated
      expect(result.pallets[0].active).toBe(false)
      // Energy should be restored
      expect(result.ship.energy).toBeCloseTo(initialEnergy + 25.0, epsilon) // PalletRestoreAmount = 25.0
    })

    it('should process multiple pallet pickups in one step', () => {
      const state = createTestState({
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 50.0
        },
        pallets: [
          { id: 1, pos: { x: 0.5, y: 0.0 }, active: true }, // Within pickup radius
          { id: 2, pos: { x: -0.5, y: 0.0 }, active: true } // Within pickup radius
        ]
      })
      const initialEnergy = state.ship.energy

      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })

      // Both pallets should be deactivated
      expect(result.pallets[0].active).toBe(false)
      expect(result.pallets[1].active).toBe(false)
      // Energy should be restored twice (clamped to MaxEnergy = 100.0)
      const expectedEnergy = Math.min(initialEnergy + 2.0 * 25.0, 100.0)
      expect(result.ship.energy).toBeCloseTo(expectedEnergy, epsilon)
    })

    it('should evaluate win condition correctly (all pallets collected)', () => {
      const state = createTestState({
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        },
        pallets: [
          { id: 1, pos: { x: 0.0, y: 0.0 }, active: true } // Ship at pallet position
        ]
      })

      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })

      // Win condition should be met
      expect(result.done).toBe(true)
      expect(result.win).toBe(true)
    })

    it('should evaluate lose condition correctly (planet collision)', () => {
      const state = createTestState({
        ship: {
          pos: { x: 50.0, y: 0.0 }, // At planet radius
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        },
        planets: [
          { pos: { x: 0.0, y: 0.0 }, radius: 50.0 }
        ],
        pallets: [
          { id: 1, pos: { x: 100.0, y: 0.0 }, active: true } // Active pallet far away
        ]
      })

      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })

      // Lose condition should be met
      expect(result.done).toBe(true)
      expect(result.win).toBe(false)
    })

    it('should prioritize win over lose condition', () => {
      const state = createTestState({
        ship: {
          pos: { x: 50.0, y: 0.0 }, // At planet radius
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        },
        planets: [
          { pos: { x: 0.0, y: 0.0 }, radius: 50.0 }
        ],
        pallets: [
          { id: 1, pos: { x: 0.0, y: 0.0 }, active: false } // All pallets collected (win condition)
        ]
      })

      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })

      // Win should take precedence
      expect(result.done).toBe(true)
      expect(result.win).toBe(true)
    })

    it('should increment tick counter correctly', () => {
      const state = createTestState({ tick: 42 })

      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })

      // Tick should increment
      expect(result.tick).toBe(43)
    })

    it('should skip processing when game is already done (only increments tick)', () => {
      const state = createTestState({
        done: true,
        win: true,
        tick: 10,
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      const initialPos = { ...state.ship.pos }
      const initialVel = { ...state.ship.vel }
      const initialRot = state.ship.rot
      const initialEnergy = state.ship.energy

      const result = simulator.step(state, { thrust: 1.0, turn: 1.0 })

      // State should be unchanged (except tick)
      expect(result.ship.pos.x).toBe(initialPos.x)
      expect(result.ship.pos.y).toBe(initialPos.y)
      expect(result.ship.vel.x).toBe(initialVel.x)
      expect(result.ship.vel.y).toBe(initialVel.y)
      expect(result.ship.rot).toBe(initialRot)
      expect(result.ship.energy).toBe(initialEnergy)
      expect(result.done).toBe(true)
      expect(result.win).toBe(true)
      // Tick should increment
      expect(result.tick).toBe(11)
    })
  })

  describe('Input Processing', () => {
    it('should clamp thrust input to [0.0, 1.0]', () => {
      const state = createTestState()
      
      // Test negative thrust
      const result1 = simulator.step(state, { thrust: -1.0, turn: 0.0 })
      expect(result1.ship.vel.x).toBeCloseTo(0.0, epsilon)
      
      // Test thrust > 1.0
      const state2 = createTestState()
      const result2 = simulator.step(state2, { thrust: 2.0, turn: 0.0 })
      // Should behave same as thrust = 1.0
      const state3 = createTestState()
      const result3 = simulator.step(state3, { thrust: 1.0, turn: 0.0 })
      expect(Math.abs(result2.ship.vel.x - result3.ship.vel.x)).toBeLessThan(epsilon)
    })

    it('should clamp turn input to [-1.0, 1.0]', () => {
      const state = createTestState()
      
      // Test turn < -1.0
      const result1 = simulator.step(state, { thrust: 0.0, turn: -2.0 })
      const result2 = simulator.step(createTestState(), { thrust: 0.0, turn: -1.0 })
      expect(Math.abs(result1.ship.rot - result2.ship.rot)).toBeLessThan(epsilon)
      
      // Test turn > 1.0
      const result3 = simulator.step(createTestState(), { thrust: 0.0, turn: 2.0 })
      const result4 = simulator.step(createTestState(), { thrust: 0.0, turn: 1.0 })
      expect(Math.abs(result3.ship.rot - result4.ship.rot)).toBeLessThan(epsilon)
    })

    it('should drain energy when thrusting', () => {
      const state = createTestState({ ship: { ...createTestState().ship, energy: 100.0 } })
      
      const result = simulator.step(state, { thrust: 1.0, turn: 0.0 })
      
      // Energy should decrease by ThrustDrainRate (0.5)
      expect(result.ship.energy).toBeCloseTo(100.0 - 0.5, epsilon)
    })

    it('should not drain energy when not thrusting', () => {
      const state = createTestState({ ship: { ...createTestState().ship, energy: 100.0 } })
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Energy should remain unchanged
      expect(result.ship.energy).toBe(100.0)
    })

    it('should not apply thrust when energy is zero', () => {
      const state = createTestState({
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 0.0 // No energy
        }
      })
      const initialVel = { ...state.ship.vel }
      
      const result = simulator.step(state, { thrust: 1.0, turn: 0.0 })
      
      // Velocity should not change (no thrust without energy)
      expect(Math.abs(result.ship.vel.x - initialVel.x)).toBeLessThan(epsilon)
      expect(Math.abs(result.ship.vel.y - initialVel.y)).toBeLessThan(epsilon)
    })

    it('should update rotation even when energy is zero', () => {
      const state = createTestState({
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 0.0
        }
      })
      const initialRot = state.ship.rot
      
      const result = simulator.step(state, { thrust: 0.0, turn: 1.0 })
      
      // Rotation should still change (rotation doesn't require energy)
      expect(result.ship.rot).toBeGreaterThan(initialRot)
    })

    it('should normalize rotation to [0, 2π)', () => {
      const state = createTestState({
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 6.0 * Math.PI, // Multiple rotations
          energy: 100.0
        }
      })
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Rotation should be normalized
      expect(result.ship.rot).toBeGreaterThanOrEqual(0.0)
      expect(result.ship.rot).toBeLessThan(2.0 * Math.PI)
    })
  })

  describe('Physics - Gravity and Integration', () => {
    it('should calculate gravity acceleration correctly', () => {
      const state = createTestState({
        ship: {
          pos: { x: 10.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        },
        planets: [
          { pos: { x: 0.0, y: 0.0 }, radius: 50.0 }
        ]
      })
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Ship should move toward planet (negative x direction)
      expect(result.ship.pos.x).toBeLessThan(10.0)
      // Velocity should point toward planet
      expect(result.ship.vel.x).toBeLessThan(0.0)
    })

    it('should clamp gravity acceleration to aMax', () => {
      const state = createTestState({
        ship: {
          pos: { x: 0.1, y: 0.0 }, // Very close to planet
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        },
        planets: [
          { pos: { x: 0.0, y: 0.0 }, radius: 50.0 }
        ]
      })
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Acceleration should be clamped (velocity change should be reasonable)
      const velChange = Math.sqrt(
        (result.ship.vel.x - state.ship.vel.x) ** 2 +
        (result.ship.vel.y - state.ship.vel.y) ** 2
      )
      const maxVelChange = aMax * dt + epsilon
      expect(velChange).toBeLessThanOrEqual(maxVelChange)
    })

    it('should handle zero distance (ship at planet center)', () => {
      const state = createTestState({
        ship: {
          pos: { x: 0.0, y: 0.0 }, // At planet center
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        },
        planets: [
          { pos: { x: 0.0, y: 0.0 }, radius: 50.0 }
        ]
      })
      const initialVel = { ...state.ship.vel }
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Velocity should remain zero (no acceleration at zero distance)
      expect(Math.abs(result.ship.vel.x - initialVel.x)).toBeLessThan(epsilon)
      expect(Math.abs(result.ship.vel.y - initialVel.y)).toBeLessThan(epsilon)
    })

    it('should use semi-implicit Euler integration', () => {
      const state = createTestState({
        ship: {
          pos: { x: 10.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Position should change based on new velocity (semi-implicit Euler)
      // v_new = v_old + a * dt
      // p_new = p_old + v_new * dt
      expect(result.ship.pos.x).not.toBe(10.0)
      expect(result.ship.vel.x).not.toBe(0.0)
    })
  })

  describe('Collision Detection', () => {
    it('should detect ship-planet collision at boundary', () => {
      const state = createTestState({
        ship: {
          pos: { x: 50.0, y: 0.0 }, // Exactly at planet radius
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        },
        planets: [
          { pos: { x: 0.0, y: 0.0 }, radius: 50.0 }
        ]
      })
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Should detect collision (lose condition)
      expect(result.done).toBe(true)
      expect(result.win).toBe(false)
    })

    it('should detect ship-pallet collision within pickup radius', () => {
      const state = createTestState({
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 50.0
        },
        pallets: [
          { id: 1, pos: { x: 1.0, y: 0.0 }, active: true } // Within pickup radius (1.2)
        ]
      })
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Pallet should be picked up
      expect(result.pallets[0].active).toBe(false)
      expect(result.ship.energy).toBeGreaterThan(50.0)
    })

    it('should not detect false collisions when ship is far from objects', () => {
      const state = createTestState({
        ship: {
          pos: { x: 1000.0, y: 1000.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        },
        planets: [
          { pos: { x: 0.0, y: 0.0 }, radius: 50.0 }
        ],
        pallets: [
          { id: 1, pos: { x: 0.0, y: 0.0 }, active: true }
        ]
      })
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Should not detect collisions
      expect(result.done).toBe(false)
      expect(result.pallets[0].active).toBe(true)
    })
  })

  describe('Energy Economy', () => {
    it('should clamp energy to [0, MaxEnergy]', () => {
      const state = createTestState({
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 150.0 // Above max
        },
        pallets: [
          { id: 1, pos: { x: 0.0, y: 0.0 }, active: true }
        ]
      })
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Energy should be clamped to MaxEnergy (100.0)
      expect(result.ship.energy).toBeLessThanOrEqual(100.0)
    })

    it('should restore energy on pallet pickup', () => {
      const state = createTestState({
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 50.0
        },
        pallets: [
          { id: 1, pos: { x: 0.0, y: 0.0 }, active: true }
        ]
      })
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Energy should be restored by PalletRestoreAmount (25.0)
      expect(result.ship.energy).toBeCloseTo(75.0, epsilon)
    })

    it('should clamp restored energy to MaxEnergy', () => {
      const state = createTestState({
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 90.0 // Close to max
        },
        pallets: [
          { id: 1, pos: { x: 0.0, y: 0.0 }, active: true }
        ]
      })
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Energy should be clamped to MaxEnergy (100.0)
      expect(result.ship.energy).toBeLessThanOrEqual(100.0)
    })
  })

  describe('Multiple Planets Support', () => {
    it('should calculate gravity from multiple planets', () => {
      const state = createTestState({
        ship: {
          pos: { x: 5.0, y: 0.0 }, // Closer to first planet
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        },
        planets: [
          { pos: { x: 0.0, y: 0.0 }, radius: 50.0 },
          { pos: { x: 20.0, y: 0.0 }, radius: 50.0 } // Planet on the right
        ]
      })
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Ship should be affected by gravity from both planets
      // Since ship is closer to first planet, net effect should pull left
      expect(result.ship.pos.x).toBeLessThan(5.0)
      expect(result.ship.vel.x).toBeLessThan(0.0)
    })

    it('should detect collision with any planet', () => {
      const state = createTestState({
        ship: {
          pos: { x: 20.0, y: 0.0 }, // At second planet radius
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        },
        planets: [
          { pos: { x: 0.0, y: 0.0 }, radius: 50.0 },
          { pos: { x: 20.0, y: 0.0 }, radius: 50.0 } // Ship at this planet
        ]
      })
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Should detect collision with second planet
      expect(result.done).toBe(true)
      expect(result.win).toBe(false)
    })
  })

  describe('Determinism', () => {
    it('should produce identical results for identical inputs', () => {
      const state1 = createTestState()
      const state2 = createTestState()
      
      const result1 = simulator.step(state1, { thrust: 1.0, turn: 0.5 })
      const result2 = simulator.step(state2, { thrust: 1.0, turn: 0.5 })
      
      // States should be identical
      expect(result1.ship.pos.x).toBeCloseTo(result2.ship.pos.x, epsilon)
      expect(result1.ship.pos.y).toBeCloseTo(result2.ship.pos.y, epsilon)
      expect(result1.ship.vel.x).toBeCloseTo(result2.ship.vel.x, epsilon)
      expect(result1.ship.vel.y).toBeCloseTo(result2.ship.vel.y, epsilon)
      expect(result1.ship.rot).toBeCloseTo(result2.ship.rot, epsilon)
      expect(result1.ship.energy).toBeCloseTo(result2.ship.energy, epsilon)
      expect(result1.tick).toBe(result2.tick)
    })

    it('should produce deterministic results across multiple steps', () => {
      const state1 = createTestState()
      const state2 = createTestState()
      const input = { thrust: 1.0, turn: 0.5 }
      
      // Apply same inputs multiple times
      let result1 = state1
      let result2 = state2
      for (let i = 0; i < 10; i++) {
        result1 = simulator.step(result1, input)
        result2 = simulator.step(result2, input)
      }
      
      // Final states should be identical
      expect(result1.ship.pos.x).toBeCloseTo(result2.ship.pos.x, epsilon)
      expect(result1.ship.pos.y).toBeCloseTo(result2.ship.pos.y, epsilon)
      expect(result1.ship.vel.x).toBeCloseTo(result2.ship.vel.x, epsilon)
      expect(result1.ship.vel.y).toBeCloseTo(result2.ship.vel.y, epsilon)
      expect(result1.ship.rot).toBeCloseTo(result2.ship.rot, epsilon)
      expect(result1.ship.energy).toBeCloseTo(result2.ship.energy, epsilon)
      expect(result1.tick).toBe(result2.tick)
    })
  })

  describe('Edge Cases', () => {
    it('should handle empty pallet list', () => {
      const state = createTestState({
        ship: {
          pos: { x: 100.0, y: 0.0 }, // Far from planet to avoid collision
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        },
        pallets: []
      })
      
      const result = simulator.step(state, { thrust: 1.0, turn: 0.0 })
      
      // Should complete without errors
      expect(result.tick).toBe(1)
      expect(result.done).toBe(false) // No win condition (no pallets), no collision (far from planet)
    })

    it('should handle empty planet list', () => {
      const state = createTestState({ planets: [] })
      
      const result = simulator.step(state, { thrust: 1.0, turn: 0.0 })
      
      // Should complete without errors (no gravity)
      expect(result.tick).toBe(1)
      // Ship should move due to thrust only
      expect(result.ship.vel.x).toBeGreaterThan(0.0)
    })

    it('should handle zero input', () => {
      const state = createTestState()
      const initialEnergy = state.ship.energy
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Energy should not change (no thrust)
      expect(result.ship.energy).toBe(initialEnergy)
    })

    it('should handle energy at maximum (clamping)', () => {
      const state = createTestState({
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0 // At max
        },
        pallets: [
          { id: 1, pos: { x: 0.0, y: 0.0 }, active: true }
        ]
      })
      
      const result = simulator.step(state, { thrust: 0.0, turn: 0.0 })
      
      // Energy should be clamped to MaxEnergy (100.0)
      expect(result.ship.energy).toBeLessThanOrEqual(100.0)
    })

    it('should handle many consecutive steps', () => {
      const state = createTestState({ tick: 0 })
      const input = { thrust: 1.0, turn: 0.0 }
      
      let result = state
      for (let i = 0; i < 100; i++) {
        result = simulator.step(result, input)
      }
      
      // Tick should increment correctly
      expect(result.tick).toBe(100)
      // Energy should be drained
      expect(result.ship.energy).toBeLessThan(100.0)
      // Energy should not go negative
      expect(result.ship.energy).toBeGreaterThanOrEqual(0.0)
    })
  })
})

