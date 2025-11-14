/**
 * Server reconciliation system that compares server snapshots to predicted state.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import type { GameState } from './state-manager'
import { StateManager } from './state-manager'
import { LocalSimulator } from './local-simulator'
import { CommandHistory } from '../net/command-history'
import { PredictionSystem } from './prediction'
import type { SnapshotMessage } from '../net/protocol'

/**
 * Result of a reconciliation operation.
 */
export interface ReconciliationResult {
  mismatchDetected: boolean
  commandsReapplied: number
  tick: number
  predictedTick: number | null
}

/**
 * Tolerance for floating-point comparisons in state mismatch detection.
 */
const FLOAT_TOLERANCE = 0.001

/**
 * Server reconciliation system that compares server snapshots to predicted state.
 * 
 * When a mismatch is detected, the system:
 * 1. Rolls back predicted state to authoritative state
 * 2. Re-applies unconfirmed commands to authoritative state
 * 3. Updates predicted state with re-applied result
 * 
 * This ensures the client stays in sync with the server while maintaining
 * responsive feel through client-side prediction.
 */
export class ReconciliationSystem {
  private stateManager: StateManager
  private localSimulator: LocalSimulator
  private commandHistory: CommandHistory
  private predictionSystem: PredictionSystem

  constructor(
    stateManager: StateManager,
    localSimulator: LocalSimulator,
    commandHistory: CommandHistory,
    predictionSystem: PredictionSystem
  ) {
    this.stateManager = stateManager
    this.localSimulator = localSimulator
    this.commandHistory = commandHistory
    this.predictionSystem = predictionSystem
  }

  /**
   * Reconciles server snapshot with predicted state.
   * 
   * Flow:
   * 1. Update authoritative state from snapshot
   * 2. Compare predicted state with authoritative state
   * 3. If mismatch detected:
   *    - Rollback predicted state
   *    - Re-apply unconfirmed commands
   * 4. If no mismatch:
   *    - Mark commands as confirmed
   * 
   * @param snapshot Server snapshot to reconcile
   * @returns Reconciliation result
   */
  reconcile(snapshot: SnapshotMessage): ReconciliationResult {
    // Update authoritative state from snapshot
    this.stateManager.updateAuthoritative(snapshot)
    
    const authoritative = this.stateManager.getAuthoritative()
    if (!authoritative) {
      // Should not happen, but handle gracefully
      return {
        mismatchDetected: false,
        commandsReapplied: 0,
        tick: snapshot.tick,
        predictedTick: null
      }
    }

    const predicted = this.stateManager.getPredicted()
    
    // If no predicted state exists, just update authoritative (no reconciliation needed)
    if (!predicted) {
      return {
        mismatchDetected: false,
        commandsReapplied: 0,
        tick: snapshot.tick,
        predictedTick: null
      }
    }

    // Compare predicted state with authoritative state
    // Note: We compare at the same tick (use authoritative tick as reference)
    // If predicted is ahead, we need to compare what predicted would be at authoritative tick
    // For now, we compare directly - if predicted is ahead, we'll detect a mismatch and re-apply
    // This is acceptable because if predicted is ahead, we want to re-sync anyway
    
    // Compare states - if ticks differ, we need to handle it
    // If predicted tick > authoritative tick, predicted is ahead (expected)
    // We compare the states directly - if they match at authoritative tick, no mismatch
    let hasMismatch = false
    
    if (predicted.tick === authoritative.tick) {
      // Same tick: compare directly
      hasMismatch = this.hasMismatch(predicted, authoritative)
    } else if (predicted.tick > authoritative.tick) {
      // Predicted is ahead: we can't directly compare
      // For now, we'll assume mismatch if ticks differ (conservative approach)
      // In practice, we'd need to "roll back" predicted to authoritative tick to compare
      // But for simplicity, if ticks differ significantly, we re-sync
      hasMismatch = true
    } else {
      // Predicted is behind authoritative (shouldn't happen, but handle)
      hasMismatch = true
    }

    if (hasMismatch) {
      // Mismatch detected: rollback and re-apply
      this.rollback()
      const commandsReapplied = this.reapplyCommands()
      
      const newPredicted = this.stateManager.getPredicted()
      
      return {
        mismatchDetected: true,
        commandsReapplied,
        tick: snapshot.tick,
        predictedTick: newPredicted?.tick ?? null
      }
    } else {
      // No mismatch: mark commands as confirmed
      // Mark all unconfirmed commands as confirmed (they matched server state)
      const unconfirmed = this.commandHistory.getUnconfirmed()
      for (const cmd of unconfirmed) {
        this.commandHistory.markConfirmed(cmd.seq)
      }
      
      return {
        mismatchDetected: false,
        commandsReapplied: 0,
        tick: snapshot.tick,
        predictedTick: predicted.tick
      }
    }
  }

  /**
   * Checks if two game states have a mismatch.
   * Uses tolerance for floating-point comparisons.
   * 
   * @param predicted Predicted state
   * @param authoritative Authoritative state
   * @returns True if mismatch detected, false otherwise
   */
  hasMismatch(predicted: GameState, authoritative: GameState): boolean {
    // Compare ship position
    if (!this.vec2Equals(predicted.ship.pos, authoritative.ship.pos)) {
      return true
    }

    // Compare ship velocity
    if (!this.vec2Equals(predicted.ship.vel, authoritative.ship.vel)) {
      return true
    }

    // Compare ship rotation (with tolerance)
    if (Math.abs(predicted.ship.rot - authoritative.ship.rot) > FLOAT_TOLERANCE) {
      // Handle rotation wrap-around (0 and 2π are the same)
      const rotDiff = Math.abs(predicted.ship.rot - authoritative.ship.rot)
      const rotDiffWrapped = Math.abs(rotDiff - 2 * Math.PI)
      if (Math.min(rotDiff, rotDiffWrapped) > FLOAT_TOLERANCE) {
        return true
      }
    }

    // Compare ship energy (exact comparison for discrete value)
    if (predicted.ship.energy !== authoritative.ship.energy) {
      return true
    }

    // Compare planets
    if (predicted.planets.length !== authoritative.planets.length) {
      return true
    }
    for (let i = 0; i < predicted.planets.length; i++) {
      const predPlanet = predicted.planets[i]
      const authPlanet = authoritative.planets[i]
      if (!this.vec2Equals(predPlanet.pos, authPlanet.pos)) {
        return true
      }
      if (Math.abs(predPlanet.radius - authPlanet.radius) > FLOAT_TOLERANCE) {
        return true
      }
    }

    // Compare pallets
    if (predicted.pallets.length !== authoritative.pallets.length) {
      return true
    }
    for (let i = 0; i < predicted.pallets.length; i++) {
      const predPallet = predicted.pallets[i]
      const authPallet = authoritative.pallets[i]
      if (predPallet.id !== authPallet.id) {
        return true
      }
      if (!this.vec2Equals(predPallet.pos, authPallet.pos)) {
        return true
      }
      // Exact comparison for active state (discrete value)
      if (predPallet.active !== authPallet.active) {
        return true
      }
    }

    // Compare game state flags (exact comparison)
    if (predicted.done !== authoritative.done) {
      return true
    }
    if (predicted.win !== authoritative.win) {
      return true
    }

    // No mismatch detected
    return false
  }

  /**
   * Rolls back predicted state to authoritative state.
   * Clears predicted state without affecting authoritative state.
   */
  rollback(): void {
    this.predictionSystem.reset()
  }

  /**
   * Re-applies unconfirmed commands to authoritative state.
   * 
   * Flow:
   * 1. Get authoritative state as base
   * 2. Get unconfirmed commands in sequence order
   * 3. For each command, run local simulation step
   * 4. Chain results (each command builds on previous)
   * 5. Update predicted state with final result
   * 
   * @returns Number of commands re-applied
   */
  reapplyCommands(): number {
    const authoritative = this.stateManager.getAuthoritative()
    if (!authoritative) {
      // Cannot re-apply without authoritative state
      return 0
    }

    const unconfirmed = this.commandHistory.getUnconfirmed()
    if (unconfirmed.length === 0) {
      // No commands to re-apply
      return 0
    }

    // Start with authoritative state
    let currentState = authoritative

    // Re-apply each unconfirmed command in sequence order
    for (const cmd of unconfirmed) {
      currentState = this.localSimulator.step(currentState, {
        thrust: cmd.thrust,
        turn: cmd.turn
      })
    }

    // Update predicted state with re-applied result
    this.stateManager.updatePredicted(currentState)

    return unconfirmed.length
  }

  /**
   * Compares two Vec2 values with tolerance.
   * 
   * @param a First vector
   * @param b Second vector
   * @returns True if vectors are equal within tolerance
   */
  private vec2Equals(a: { x: number, y: number }, b: { x: number, y: number }): boolean {
    const dx = Math.abs(a.x - b.x)
    const dy = Math.abs(a.y - b.y)
    return dx <= FLOAT_TOLERANCE && dy <= FLOAT_TOLERANCE
  }
}

