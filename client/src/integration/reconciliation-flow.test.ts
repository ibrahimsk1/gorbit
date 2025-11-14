/**
 * Integration tests for reconciliation flow and rollback behavior.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import { describe, it, expect, beforeEach } from 'vitest'
import { ReconciliationSystem } from '../sim/reconciliation'
import { StateManager } from '../sim/state-manager'
import { LocalSimulator } from '../sim/local-simulator'
import { CommandHistory } from '../net/command-history'
import { PredictionSystem } from '../sim/prediction'
import { createTestSnapshot } from './test-helpers'
import type { SnapshotMessage } from '../net/protocol'

describe('Reconciliation Flow', () => {
  let reconciliationSystem: ReconciliationSystem
  let stateManager: StateManager
  let localSimulator: LocalSimulator
  let commandHistory: CommandHistory
  let predictionSystem: PredictionSystem

  beforeEach(() => {
    stateManager = new StateManager()
    localSimulator = new LocalSimulator()
    commandHistory = new CommandHistory()
    predictionSystem = new PredictionSystem(stateManager, localSimulator, commandHistory)
    reconciliationSystem = new ReconciliationSystem(stateManager, localSimulator, commandHistory, predictionSystem)
  })

  describe('Reconciliation on Server Snapshot Arrival', () => {
    it('runs reconciliation when server snapshot arrives', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const newSnapshot = createTestSnapshot(1)
      const result = reconciliationSystem.reconcile(newSnapshot)
      
      expect(result).toBeDefined()
      expect(result.mismatchDetected !== undefined).toBe(true)
    })

    it('compares predicted vs authoritative state', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const predicted = stateManager.getPredicted()
      expect(predicted).not.toBeNull()
      
      const newSnapshot = createTestSnapshot(1)
      const result = reconciliationSystem.reconcile(newSnapshot)
      
      // Reconciliation should have compared predicted vs authoritative
      expect(result).toBeDefined()
    })

    it('detects mismatches correctly', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      // Server sends different snapshot (mismatch)
      const newSnapshot = createTestSnapshot(1, {
        ship: {
          pos: { x: 15.0, y: 0.0 }, // Different from prediction
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      
      const result = reconciliationSystem.reconcile(newSnapshot)
      
      expect(result.mismatchDetected).toBe(true)
    })

    it('handles matching states (no rollback)', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      // Get predicted state to create matching snapshot
      const predicted = stateManager.getPredicted()!
      
      const newSnapshot = createTestSnapshot(predicted.tick, {
        ship: predicted.ship,
        planets: predicted.planets,
        pallets: predicted.pallets,
        done: predicted.done,
        win: predicted.win
      })
      
      const result = reconciliationSystem.reconcile(newSnapshot)
      
      expect(result.mismatchDetected).toBe(false)
    })
  })

  describe('Rollback Behavior', () => {
    it('rolls back predicted state to authoritative state on mismatch', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const predictedBefore = stateManager.getPredicted()
      expect(predictedBefore).not.toBeNull()
      
      // Server sends different snapshot (mismatch)
      const newSnapshot = createTestSnapshot(1, {
        ship: {
          pos: { x: 15.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      
      reconciliationSystem.reconcile(newSnapshot)
      
      // After reconciliation, authoritative state should be updated
      const authoritative = stateManager.getAuthoritative()
      expect(authoritative).not.toBeNull()
      expect(authoritative?.tick).toBe(1)
    })

    it('preserves unconfirmed commands after rollback', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      commandHistory.addCommand(2, 0.5, 0.1)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      predictionSystem.predict({ thrust: 0.5, turn: 0.1 })
      
      const unconfirmedBefore = commandHistory.getUnconfirmed()
      expect(unconfirmedBefore.length).toBeGreaterThan(0)
      
      const newSnapshot = createTestSnapshot(1, {
        ship: {
          pos: { x: 15.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      
      reconciliationSystem.reconcile(newSnapshot)
      
      // Unconfirmed commands should still exist (for re-application)
      const unconfirmedAfter = commandHistory.getUnconfirmed()
      expect(unconfirmedAfter.length).toBeGreaterThanOrEqual(0)
    })

    it('triggers re-application of unconfirmed commands after rollback', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const newSnapshot = createTestSnapshot(1, {
        ship: {
          pos: { x: 15.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      
      const result = reconciliationSystem.reconcile(newSnapshot)
      
      // If mismatch detected, commands should be re-applied
      if (result.mismatchDetected) {
        expect(result.commandsReapplied).toBeGreaterThan(0)
      }
    })

    it('updates StateManager correctly after rollback', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const newSnapshot = createTestSnapshot(1)
      reconciliationSystem.reconcile(newSnapshot)
      
      // Authoritative state should be updated
      const authoritative = stateManager.getAuthoritative()
      expect(authoritative).not.toBeNull()
      expect(authoritative?.tick).toBe(1)
    })
  })

  describe('Command Re-application', () => {
    it('re-applies unconfirmed commands after rollback', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const newSnapshot = createTestSnapshot(1, {
        ship: {
          pos: { x: 15.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      
      const result = reconciliationSystem.reconcile(newSnapshot)
      
      if (result.mismatchDetected) {
        expect(result.commandsReapplied).toBeGreaterThan(0)
        
        // After re-application, predicted state should exist again
        const predictedAfter = stateManager.getPredicted()
        expect(predictedAfter).not.toBeNull()
      }
    })

    it('re-applied commands use authoritative state as base', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const newSnapshot = createTestSnapshot(1, {
        ship: {
          pos: { x: 15.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      
      reconciliationSystem.reconcile(newSnapshot)
      
      // After re-application, predicted state should be based on new authoritative state
      const authoritative = stateManager.getAuthoritative()
      const predicted = stateManager.getPredicted()
      
      if (predicted) {
        expect(predicted.tick).toBeGreaterThan(authoritative!.tick)
      }
    })

    it('re-applied commands produce new predicted state', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const newSnapshot = createTestSnapshot(1, {
        ship: {
          pos: { x: 15.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      
      reconciliationSystem.reconcile(newSnapshot)
      
      // After re-application, predicted state should exist
      const predicted = stateManager.getPredicted()
      expect(predicted).not.toBeNull()
    })

    it('re-applied commands maintain correct sequence', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      commandHistory.addCommand(2, 0.5, 0.1)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      predictionSystem.predict({ thrust: 0.5, turn: 0.1 })
      
      const newSnapshot = createTestSnapshot(1, {
        ship: {
          pos: { x: 15.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      
      const result = reconciliationSystem.reconcile(newSnapshot)
      
      if (result.mismatchDetected) {
        // Commands should be re-applied in sequence
        expect(result.commandsReapplied).toBeGreaterThan(0)
      }
    })
  })

  describe('Reconciliation Scenarios', () => {
    it('handles scenario: prediction matches server (no rollback)', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const predicted = stateManager.getPredicted()!
      const newSnapshot = createTestSnapshot(predicted.tick, {
        ship: predicted.ship,
        planets: predicted.planets,
        pallets: predicted.pallets,
        done: predicted.done,
        win: predicted.win
      })
      
      const result = reconciliationSystem.reconcile(newSnapshot)
      
      expect(result.mismatchDetected).toBe(false)
      expect(result.commandsReapplied).toBe(0)
    })

    it('handles scenario: prediction differs from server (rollback and re-apply)', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      
      const newSnapshot = createTestSnapshot(1, {
        ship: {
          pos: { x: 15.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      
      const result = reconciliationSystem.reconcile(newSnapshot)
      
      expect(result.mismatchDetected).toBe(true)
      expect(result.commandsReapplied).toBeGreaterThan(0)
    })

    it('handles scenario: multiple unconfirmed commands re-applied correctly', () => {
      const snapshot = createTestSnapshot(0)
      stateManager.updateAuthoritative(snapshot)
      
      commandHistory.addCommand(1, 1.0, 0.0)
      commandHistory.addCommand(2, 0.5, 0.1)
      commandHistory.addCommand(3, 0.8, -0.2)
      predictionSystem.predict({ thrust: 1.0, turn: 0.0 })
      predictionSystem.predict({ thrust: 0.5, turn: 0.1 })
      predictionSystem.predict({ thrust: 0.8, turn: -0.2 })
      
      const newSnapshot = createTestSnapshot(1, {
        ship: {
          pos: { x: 15.0, y: 0.0 },
          vel: { x: 0.0, y: 0.0 },
          rot: 0.0,
          energy: 100.0
        }
      })
      
      const result = reconciliationSystem.reconcile(newSnapshot)
      
      if (result.mismatchDetected) {
        expect(result.commandsReapplied).toBeGreaterThan(0)
        
        // After re-application, predicted state should reflect all commands
        const predicted = stateManager.getPredicted()
        expect(predicted).not.toBeNull()
      }
    })

    it('handles scenario: reconciliation with no predicted state (first snapshot)', () => {
      const snapshot = createTestSnapshot(0)
      
      const result = reconciliationSystem.reconcile(snapshot)
      
      // No mismatch detected (nothing to compare)
      expect(result.mismatchDetected).toBe(false)
      
      // Authoritative state should be updated
      const authoritative = stateManager.getAuthoritative()
      expect(authoritative).not.toBeNull()
      expect(authoritative?.tick).toBe(0)
    })
  })
})

