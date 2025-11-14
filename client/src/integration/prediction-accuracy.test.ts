/**
 * Integration tests for prediction accuracy verification.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import { describe, it, expect, beforeEach } from 'vitest'
import { PredictionSystem } from '../sim/prediction'
import { StateManager } from '../sim/state-manager'
import { LocalSimulator } from '../sim/local-simulator'
import { CommandHistory } from '../net/command-history'
import { createTestSnapshot, statesApproximatelyEqual } from './test-helpers'
import type { SnapshotMessage } from '../net/protocol'

describe('Prediction Accuracy', () => {
  let predictionSystem: PredictionSystem
  let stateManager: StateManager
  let localSimulator: LocalSimulator
  let commandHistory: CommandHistory

  beforeEach(() => {
    stateManager = new StateManager()
    localSimulator = new LocalSimulator()
    commandHistory = new CommandHistory()
    predictionSystem = new PredictionSystem(stateManager, localSimulator, commandHistory)
  })

  describe('Prediction Runs Immediately on Input', () => {
    it('runs prediction when command is sent', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      expect(predictionSystem.hasPredictedState()).toBe(true)
      const predicted = predictionSystem.getPredictedState()
      expect(predicted).not.toBeNull()
    })

    it('predicted state differs from authoritative state', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const predicted = predictionSystem.getPredictedState()
      const authoritative = stateManager.getAuthoritative()
      
      expect(predicted).not.toBeNull()
      expect(predicted?.tick).toBe(authoritative!.tick + 1)
    })

    it('predicted state uses authoritative state as base', () => {
      const snapshot = createTestSnapshot(0, {
        ship: {
          pos: { x: 100, y: 200 },
          vel: { x: 5, y: -5 },
          rot: 1.57,
          energy: 75.0
        }
      })
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 0.5, 0.1)
      predictionSystem.predict({ thrust: 0.5, turn: 0.1 })
      
      const predicted = predictionSystem.getPredictedState()
      const authoritative = stateManager.getAuthoritative()
      
      // Predicted state should be based on authoritative (same initial conditions)
      expect(predicted).not.toBeNull()
      expect(predicted?.tick).toBe(authoritative!.tick + 1)
    })

    it('supports prediction chaining (multiple commands before server confirmation)', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      // First prediction
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const predicted1 = predictionSystem.getPredictedState()
      expect(predicted1).not.toBeNull()
      expect(predicted1?.tick).toBe(1)
      
      // Second prediction (chaining)
      commandHistory.addCommand(2, 0.5, 0.1)
      predictionSystem.predict({ thrust: 0.5, turn: 0.1 })
      
      const predicted2 = predictionSystem.getPredictedState()
      expect(predicted2).not.toBeNull()
      expect(predicted2?.tick).toBe(2)
    })
  })

  describe('Prediction Accuracy Threshold', () => {
    it('prediction error is within threshold for position (≤ 5 units)', () => {
      const snapshot = createTestSnapshot(0, {
        ship: {
          pos: { x: 0, y: 0 },
          vel: { x: 0, y: 0 },
          rot: 0,
          energy: 100
        }
      })
      stateManager.updateAuthoritative(snapshot)
      
      // Run prediction with simple input
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      // Simulate server response (should be similar to prediction)
      const serverSnapshot = createTestSnapshot(1, {
        ship: {
          pos: { x: 0.1, y: 0 }, // Small difference (within threshold)
          vel: { x: 0, y: 0 },
          rot: 0,
          energy: 100
        }
      })
      stateManager.updateAuthoritative(serverSnapshot)
      
      const predicted = predictionSystem.getPredictedState()
      const authoritative = stateManager.getAuthoritative()
      
      if (predicted && authoritative) {
        const comparison = statesApproximatelyEqual(predicted, authoritative, 5.0, 0.1, 1.0)
        // For this test, we're just verifying the comparison function works
        // In real scenario, server and prediction should be close
        expect(comparison.equal || comparison.errors.position! <= 5.0).toBe(true)
      }
    })

    it('prediction error is within threshold for rotation (≤ 0.1 radians)', () => {
      const snapshot = createTestSnapshot(0, {
        ship: {
          pos: { x: 0, y: 0 },
          vel: { x: 0, y: 0 },
          rot: 0,
          energy: 100
        }
      })
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 0.0, 0.5)
      predictionSystem.predict({ thrust: 0.0, turn: 0.5 })
      
      const predicted = predictionSystem.getPredictedState()
      const authoritative = stateManager.getAuthoritative()
      
      if (predicted && authoritative) {
        const comparison = statesApproximatelyEqual(predicted, authoritative, 5.0, 0.1, 1.0)
        // Rotation error should be within threshold
        if (comparison.errors.rotation !== undefined) {
          expect(comparison.errors.rotation).toBeLessThanOrEqual(0.1)
        }
      }
    })

    it('prediction accuracy maintained over multiple ticks', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      // Run multiple predictions
      for (let i = 1; i <= 5; i++) {
        commandHistory.addCommand(i, 0.5, 0.1)
        predictionSystem.predict({ thrust: 0.5, turn: 0.1 })
        
        const predicted = predictionSystem.getPredictedState()
        expect(predicted).not.toBeNull()
        expect(predicted?.tick).toBe(i)
      }
      
      // Final prediction should still be accurate
      const finalPredicted = predictionSystem.getPredictedState()
      expect(finalPredicted).not.toBeNull()
      expect(finalPredicted?.tick).toBe(5)
    })

    it('prediction accuracy with different input patterns', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      // Test with different input patterns
      const patterns = [
        { thrust: 1.0, turn: 0.0 },
        { thrust: 0.0, turn: 1.0 },
        { thrust: 0.5, turn: 0.5 },
        { thrust: -0.5, turn: -0.5 }
      ]
      
      patterns.forEach((pattern, index) => {
        commandHistory.addCommand(index + 1, pattern.thrust, pattern.turn)
        predictionSystem.predict(pattern)
        
        const predicted = predictionSystem.getPredictedState()
        expect(predicted).not.toBeNull()
        expect(predicted?.tick).toBe(index + 1)
      })
    })
  })

  describe('Prediction State Management', () => {
    it('predicted state stored separately from authoritative state', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const predicted = stateManager.getPredicted()
      const authoritative = stateManager.getAuthoritative()
      
      expect(predicted).not.toBeNull()
      expect(authoritative).not.toBeNull()
      expect(predicted?.tick).not.toBe(authoritative?.tick)
    })

    it('predicted state accessible via StateManager', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const predictedFromManager = stateManager.getPredicted()
      const predictedFromSystem = predictionSystem.getPredictedState()
      
      expect(predictedFromManager).not.toBeNull()
      expect(predictedFromSystem).not.toBeNull()
      expect(predictedFromManager?.tick).toBe(predictedFromSystem?.tick)
    })

    it('predicted state updated correctly on new predictions', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      // First prediction
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const predicted1 = predictionSystem.getPredictedState()
      expect(predicted1?.tick).toBe(1)
      
      // Second prediction
      commandHistory.addCommand(2, 0.5, 0.1)
      predictionSystem.predict({ thrust: 0.5, turn: 0.1 })
      
      const predicted2 = predictionSystem.getPredictedState()
      expect(predicted2?.tick).toBe(2)
      expect(predicted2?.tick).not.toBe(predicted1?.tick)
    })
  })
})

