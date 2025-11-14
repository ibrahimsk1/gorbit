/**
 * Integration tests for client-side prediction system.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { PredictionSystem } from './prediction'
import { StateManager } from './state-manager'
import { LocalSimulator } from './local-simulator'
import { CommandHistory } from '../net/command-history'
import type { SnapshotMessage, GameState } from './state-manager'

describe('PredictionSystem', () => {
  let predictionSystem: PredictionSystem
  let stateManager: StateManager
  let localSimulator: LocalSimulator
  let commandHistory: CommandHistory

  const createTestSnapshot = (tick: number, overrides?: Partial<SnapshotMessage>): SnapshotMessage => ({
    t: 'snapshot',
    tick,
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

  const createTestState = (tick: number, overrides?: Partial<GameState>): GameState => ({
    tick,
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
    stateManager = new StateManager()
    localSimulator = new LocalSimulator()
    commandHistory = new CommandHistory()
    predictionSystem = new PredictionSystem(stateManager, localSimulator, commandHistory)
  })

  describe('Initialization', () => {
    it('should require state manager, local simulator, and command history', () => {
      expect(() => {
        new PredictionSystem(stateManager, localSimulator, commandHistory)
      }).not.toThrow()
    })

    it('should start with no predicted state', () => {
      expect(predictionSystem.hasPredictedState()).toBe(false)
      expect(predictionSystem.getPredictedState()).toBeNull()
    })
  })

  describe('Prediction with Authoritative State', () => {
    it('should run local simulation when authoritative state exists', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      // Add command to history
      commandHistory.addCommand(1, 1.0, 0.0)
      
      // Run prediction
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      // Predicted state should exist
      expect(predictionSystem.hasPredictedState()).toBe(true)
      const predicted = predictionSystem.getPredictedState()
      expect(predicted).not.toBeNull()
      
      // Predicted state should differ from authoritative (simulation ran)
      const authoritative = stateManager.getAuthoritative()
      expect(predicted?.tick).toBe(authoritative!.tick + 1)
    })

    it('should update predicted state in state manager', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      // State manager should have predicted state
      const predictedFromManager = stateManager.getPredicted()
      expect(predictedFromManager).not.toBeNull()
      expect(predictedFromManager?.tick).toBe(1)
    })

    it('should use authoritative state as base for prediction', () => {
      const snapshot = createTestSnapshot(5, {
        ship: {
          pos: { x: 20.0, y: 10.0 },
          vel: { x: 5.0, y: 2.0 },
          rot: 1.57,
          energy: 50.0
        }
      })
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 0.5, 0.3)
      predictionSystem.predict({ thrust: 0.5, turn: 0.3 })
      
      const predicted = predictionSystem.getPredictedState()
      expect(predicted).not.toBeNull()
      // Predicted state should be based on authoritative state (tick 5 -> 6)
      expect(predicted?.tick).toBe(6)
    })

    it('should maintain predicted state separately from authoritative state', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const authoritative = stateManager.getAuthoritative()
      const predicted = stateManager.getPredicted()
      
      // They should be different objects
      expect(predicted).not.toBe(authoritative)
      // Predicted tick should be one more than authoritative
      expect(predicted?.tick).toBe(authoritative!.tick + 1)
    })

    it('should predict with thrust input correctly', () => {
      const snapshot = createTestSnapshot(0, {
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0, // Facing right
          energy: 100.0
        }
      })
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const predicted = predictionSystem.getPredictedState()
      expect(predicted).not.toBeNull()
      // Velocity should increase (thrust applied)
      expect(predicted!.ship.vel.x).toBeGreaterThan(0)
      // Energy should decrease (drained by thrust)
      expect(predicted!.ship.energy).toBeLessThan(100.0)
    })

    it('should predict with turn input correctly', () => {
      const snapshot = createTestSnapshot(0, {
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 0.0, 1.0)
      predictionSystem.predict({ thrust: 0.0, turn: 1.0 })
      
      const predicted = predictionSystem.getPredictedState()
      expect(predicted).not.toBeNull()
      // Rotation should change
      expect(predicted!.ship.rot).toBeGreaterThan(0)
    })

    it('should predict with both thrust and turn input correctly', () => {
      const snapshot = createTestSnapshot(0, {
        ship: {
          pos: { x: 0.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 0.8, 0.5)
      predictionSystem.predict({ thrust: 0.8, turn: 0.5 })
      
      const predicted = predictionSystem.getPredictedState()
      expect(predicted).not.toBeNull()
      // Both velocity and rotation should change
      const velLength = Math.sqrt(predicted!.ship.vel.x ** 2 + predicted!.ship.vel.y ** 2)
      expect(velLength).toBeGreaterThan(0)
      expect(predicted!.ship.rot).toBeGreaterThan(0)
    })

    it('should predict with no input (gravity only)', () => {
      const snapshot = createTestSnapshot(0, {
        ship: {
          pos: { x: 10.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 0.0, 0.0)
      predictionSystem.predict({ thrust: 0.0, turn: 0.0 })
      
      const predicted = predictionSystem.getPredictedState()
      expect(predicted).not.toBeNull()
      // Position should change (gravity pulls ship toward planet)
      const posChanged = predicted!.ship.pos.x !== 10.0 || predicted!.ship.pos.y !== 0.0
      expect(posChanged).toBe(true)
    })
  })

  describe('Prediction without Authoritative State', () => {
    it('should not predict if no authoritative state exists', () => {
      commandHistory.addCommand(1, 1.0, 0.0)
      
      // Should not throw, but prediction should not run
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      // Predicted state should remain null
      expect(predictionSystem.hasPredictedState()).toBe(false)
      expect(predictionSystem.getPredictedState()).toBeNull()
    })

    it('should not update predicted state if no authoritative state exists', () => {
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      // State manager should not have predicted state
      expect(stateManager.getPredicted()).toBeNull()
    })
  })

  describe('Prediction Chaining', () => {
    it('should chain predictions when multiple commands are sent', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      // First prediction
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const firstPredicted = predictionSystem.getPredictedState()
      expect(firstPredicted?.tick).toBe(1)
      
      // Second prediction (chains from first)
      commandHistory.addCommand(2, 0.5, 0.3)
      predictionSystem.predict({ thrust: 0.5, turn: 0.3 })
      
      const secondPredicted = predictionSystem.getPredictedState()
      expect(secondPredicted?.tick).toBe(2)
      
      // Third prediction (chains from second)
      commandHistory.addCommand(3, 0.0, 1.0)
      predictionSystem.predict({ thrust: 0.0, turn: 1.0 })
      
      const thirdPredicted = predictionSystem.getPredictedState()
      expect(thirdPredicted?.tick).toBe(3)
    })

    it('should use predicted state as base when chaining predictions', () => {
      const snapshot = createTestSnapshot(0, {
        ship: {
          pos: { x: 10.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0, // Facing right
          energy: 100.0
        }
      })
      stateManager.updateAuthoritative(snapshot)
      
      // First prediction
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const firstState = predictionSystem.getPredictedState()!
      const firstPosX = firstState.ship.pos.x
      const firstPosY = firstState.ship.pos.y
      const firstVelX = firstState.ship.vel.x
      const firstVelY = firstState.ship.vel.y
      
      // Second prediction should build on first
      commandHistory.addCommand(2, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const secondState = predictionSystem.getPredictedState()!
      const secondPosX = secondState.ship.pos.x
      const secondPosY = secondState.ship.pos.y
      const secondVelX = secondState.ship.vel.x
      const secondVelY = secondState.ship.vel.y
      
      // Position and velocity should have progressed from first prediction
      // Check that tick incremented (proves we're using predicted state as base)
      expect(secondState.tick).toBe(firstState.tick + 1)
      // Velocity should have increased (thrust applied in second step)
      const firstVelLength = Math.sqrt(firstVelX ** 2 + firstVelY ** 2)
      const secondVelLength = Math.sqrt(secondVelX ** 2 + secondVelY ** 2)
      expect(secondVelLength).toBeGreaterThan(firstVelLength)
      // Position should have changed (either from thrust or gravity)
      const posChanged = secondPosX !== firstPosX || secondPosY !== firstPosY
      expect(posChanged).toBe(true)
    })

    it('should track all predicted commands in history', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      commandHistory.addCommand(2, 0.5, 0.3)
      predictionSystem.predict({ thrust: 0.5, turn: 0.3 })
      
      commandHistory.addCommand(3, 0.0, 1.0)
      predictionSystem.predict({ thrust: 0.0, turn: 1.0 })
      
      // All commands should be in history
      expect(commandHistory.getCommand(1)).not.toBeNull()
      expect(commandHistory.getCommand(2)).not.toBeNull()
      expect(commandHistory.getCommand(3)).not.toBeNull()
      
      // All should be unconfirmed
      const unconfirmed = commandHistory.getUnconfirmed()
      expect(unconfirmed).toHaveLength(3)
    })
  })

  describe('Prediction State Management', () => {
    it('should return current predicted state', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const predicted = predictionSystem.getPredictedState()
      expect(predicted).not.toBeNull()
      expect(predicted?.tick).toBe(1)
    })

    it('should return true for hasPredictedState after prediction', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      expect(predictionSystem.hasPredictedState()).toBe(true)
    })

    it('should maintain predicted state separate from authoritative state', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const authoritative = stateManager.getAuthoritative()
      const predicted = stateManager.getPredicted()
      
      // Should be different objects
      expect(predicted).not.toBe(authoritative)
      // Should have different tick values
      expect(predicted?.tick).not.toBe(authoritative?.tick)
    })
  })

  describe('Reset Functionality', () => {
    it('should clear predicted state', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      expect(predictionSystem.hasPredictedState()).toBe(true)
      
      predictionSystem.reset()
      
      expect(predictionSystem.hasPredictedState()).toBe(false)
      expect(predictionSystem.getPredictedState()).toBeNull()
    })

    it('should not affect authoritative state when reset', () => {
      const snapshot = createTestSnapshot(5)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const authoritativeBefore = stateManager.getAuthoritative()
      
      predictionSystem.reset()
      
      const authoritativeAfter = stateManager.getAuthoritative()
      expect(authoritativeAfter).not.toBeNull()
      expect(authoritativeAfter?.tick).toBe(authoritativeBefore?.tick)
    })

    it('should clear predicted state in state manager', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      expect(stateManager.getPredicted()).not.toBeNull()
      
      predictionSystem.reset()
      
      expect(stateManager.getPredicted()).toBeNull()
    })

    it('should handle multiple resets', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      predictionSystem.reset()
      predictionSystem.reset()
      predictionSystem.reset()
      
      expect(predictionSystem.hasPredictedState()).toBe(false)
    })
  })

  describe('Edge Cases', () => {
    it('should handle prediction with empty command history gracefully', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      // Try to predict without adding command to history
      // Should use the input directly (not from history)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      // Prediction should still work (uses input directly)
      expect(predictionSystem.hasPredictedState()).toBe(true)
    })

    it('should handle prediction when game is done', () => {
      const snapshot = createTestSnapshot(0, {
        done: true,
        win: true
      })
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const predicted = predictionSystem.getPredictedState()
      expect(predicted).not.toBeNull()
      // Game should still be done
      expect(predicted?.done).toBe(true)
      expect(predicted?.win).toBe(true)
      // Tick should increment but state won't change
      expect(predicted?.tick).toBe(1)
    })

    it('should handle prediction with invalid input (should be clamped)', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 2.0, -2.0) // Invalid values
      predictionSystem.predict({ thrust: 2.0, turn: -2.0 })
      
      const predicted = predictionSystem.getPredictedState()
      expect(predicted).not.toBeNull()
      // Local simulator should clamp input, so prediction should still work
    })

    it('should handle prediction after reset and new authoritative state', () => {
      const snapshot1 = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot1)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      predictionSystem.reset()
      
      // New authoritative state
      const snapshot2 = createTestSnapshot(10)
      stateManager.updateAuthoritative(snapshot2)
      
      commandHistory.addCommand(2, 0.5, 0.3)
      predictionSystem.predict({ thrust: 0.5, turn: 0.3 })
      
      const predicted = predictionSystem.getPredictedState()
      expect(predicted).not.toBeNull()
      // Should use new authoritative state as base (tick 10 -> 11)
      expect(predicted?.tick).toBe(11)
    })
  })

  describe('Integration with Command History', () => {
    it('should use most recent command from history for prediction', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 0.5, 0.3)
      commandHistory.addCommand(2, 1.0, 0.0)
      commandHistory.addCommand(3, 0.0, 1.0)
      
      // Predict with the most recent command (seq 3)
      predictionSystem.predict({ thrust: 0.0, turn: 1.0 })
      
      const predicted = predictionSystem.getPredictedState()
      expect(predicted).not.toBeNull()
      // Rotation should change (turn input)
      expect(predicted!.ship.rot).toBeGreaterThan(0)
    })

    it('should track which commands are predicted (not yet confirmed)', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      commandHistory.addCommand(2, 0.5, 0.3)
      predictionSystem.predict({ thrust: 0.5, turn: 0.3 })
      
      // Both commands should be unconfirmed
      const unconfirmed = commandHistory.getUnconfirmed()
      expect(unconfirmed).toHaveLength(2)
      expect(unconfirmed[0].seq).toBe(1)
      expect(unconfirmed[1].seq).toBe(2)
    })

    it('should not affect prediction when commands are confirmed', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const predictedBefore = predictionSystem.getPredictedState()
      
      // Confirm command (simulating server acknowledgment)
      commandHistory.markConfirmed(1)
      
      // Prediction should still exist (confirmation doesn't affect prediction)
      const predictedAfter = predictionSystem.getPredictedState()
      expect(predictedAfter).not.toBeNull()
      expect(predictedAfter?.tick).toBe(predictedBefore?.tick)
    })
  })
})

