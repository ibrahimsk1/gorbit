/**
 * Client-side prediction system that runs local simulation when input is sent.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import type { GameState } from './state-manager'
import { StateManager } from './state-manager'
import { LocalSimulator } from './local-simulator'
import { CommandHistory } from '../net/command-history'

/**
 * Client-side prediction system that immediately runs local simulation
 * when input commands are sent to the server.
 * 
 * The prediction system:
 * - Maintains predicted state separately from authoritative state
 * - Runs local simulation immediately when input is sent (before server confirmation)
 * - Supports prediction chaining (multiple commands before server confirmation)
 * - Integrates with state manager, local simulator, and command history
 * 
 * Prediction flow:
 * 1. Input command is sent and added to command history
 * 2. Prediction runs immediately using authoritative state as base (or predicted if chaining)
 * 3. Predicted state is updated in state manager
 * 4. When server snapshot arrives, reconciliation compares predicted vs authoritative
 */
export class PredictionSystem {
  private stateManager: StateManager
  private localSimulator: LocalSimulator
  private commandHistory: CommandHistory

  constructor(
    stateManager: StateManager,
    localSimulator: LocalSimulator,
    commandHistory: CommandHistory
  ) {
    this.stateManager = stateManager
    this.localSimulator = localSimulator
    this.commandHistory = commandHistory
  }

  /**
   * Runs prediction for a new input command.
   * Immediately runs local simulation and updates predicted state.
   * 
   * @param input Input command (thrust, turn) to predict
   */
  predict(input: { thrust: number, turn: number }): void {
    // Get base state for prediction
    // Prefer authoritative state (most accurate), fallback to predicted (for chaining)
    const baseState = this.getBaseState()
    
    // Cannot predict without base state
    if (!baseState) {
      return
    }

    // Run local simulation step with input
    const predictedState = this.localSimulator.step(baseState, input)

    // Update predicted state in state manager
    this.stateManager.updatePredicted(predictedState)
  }

  /**
   * Gets the base state for prediction.
   * Prefers predicted state for chaining, falls back to authoritative state.
   * 
   * @returns Base state for prediction, or null if no state exists
   */
  private getBaseState(): GameState | null {
    // Prefer predicted state (for chaining predictions)
    // This allows multiple predictions to chain before server confirmation
    const predicted = this.stateManager.getPredicted()
    if (predicted) {
      return predicted
    }

    // Fallback to authoritative state (most accurate base for first prediction)
    const authoritative = this.stateManager.getAuthoritative()
    if (authoritative) {
      return authoritative
    }

    // No base state available
    return null
  }

  /**
   * Gets the current predicted state.
   * 
   * @returns Current predicted state, or null if no prediction exists
   */
  getPredictedState(): GameState | null {
    return this.stateManager.getPredicted()
  }

  /**
   * Checks if predicted state exists.
   * 
   * @returns True if predicted state exists, false otherwise
   */
  hasPredictedState(): boolean {
    return this.stateManager.hasPredicted()
  }

  /**
   * Resets prediction state by clearing predicted state.
   * Used during reconciliation rollback.
   * Does not affect authoritative state.
   */
  reset(): void {
    this.stateManager.clearPredicted()
  }
}

