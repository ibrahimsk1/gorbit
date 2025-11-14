/**
 * State manager coordinating authoritative, predicted, and interpolated states.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import type { SnapshotMessage, ShipSnapshot, PlanetSnapshot, PalletSnapshot } from '../net/protocol'

/**
 * Game state structure matching SnapshotMessage (without the 't' discriminator).
 * Used internally for state management layers.
 */
export interface GameState {
  tick: number
  ship: ShipSnapshot
  planets: PlanetSnapshot[]
  pallets: PalletSnapshot[]
  done: boolean
  win: boolean
}

/**
 * Converts a SnapshotMessage to GameState format.
 */
function snapshotToGameState(snapshot: SnapshotMessage): GameState {
  return {
    tick: snapshot.tick,
    ship: snapshot.ship,
    planets: snapshot.planets,
    pallets: snapshot.pallets,
    done: snapshot.done,
    win: snapshot.win
  }
}

/**
 * Creates a deep clone of a GameState.
 */
function cloneGameState(state: GameState): GameState {
  return {
    tick: state.tick,
    ship: {
      pos: { ...state.ship.pos },
      vel: { ...state.ship.vel },
      rot: state.ship.rot,
      energy: state.ship.energy
    },
    planets: state.planets.map(planet => ({
      pos: { ...planet.pos },
      radius: planet.radius
    })),
    pallets: state.pallets.map(pallet => ({
      id: pallet.id,
      pos: { ...pallet.pos },
      active: pallet.active
    })),
    done: state.done,
    win: state.win
  }
}

/**
 * State manager coordinating authoritative, predicted, and interpolated states.
 * 
 * - Authoritative State: Ground truth from server snapshots
 * - Predicted State: Local simulation result (may be rolled back)
 * - Interpolated State: Smoothed state for rendering (derived from authoritative)
 */
export class StateManager {
  private authoritativeState: GameState | null = null
  private predictedState: GameState | null = null
  private interpolatedState: GameState | null = null

  /**
   * Updates authoritative state from server snapshot.
   */
  updateAuthoritative(snapshot: SnapshotMessage): void {
    this.authoritativeState = snapshotToGameState(snapshot)
  }

  /**
   * Updates predicted state from local simulation.
   */
  updatePredicted(state: GameState): void {
    this.predictedState = cloneGameState(state)
  }

  /**
   * Clears predicted state (used during reconciliation rollback).
   */
  clearPredicted(): void {
    this.predictedState = null
  }

  /**
   * Updates interpolated state for smooth rendering.
   */
  updateInterpolated(state: GameState): void {
    this.interpolatedState = cloneGameState(state)
  }

  /**
   * Gets current authoritative state.
   */
  getAuthoritative(): GameState | null {
    return this.authoritativeState
  }

  /**
   * Gets current predicted state.
   */
  getPredicted(): GameState | null {
    return this.predictedState
  }

  /**
   * Gets current interpolated state.
   */
  getInterpolated(): GameState | null {
    return this.interpolatedState
  }

  /**
   * Gets state for rendering. Prefers interpolated, falls back to authoritative.
   */
  getRenderState(): GameState {
    if (this.interpolatedState) {
      return this.interpolatedState
    }
    if (this.authoritativeState) {
      return this.authoritativeState
    }
    // Return empty state if nothing is available (shouldn't happen in practice)
    return {
      tick: 0,
      ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 0 },
      planets: [],
      pallets: [],
      done: false,
      win: false
    }
  }

  /**
   * Resets all states to empty.
   */
  reset(): void {
    this.authoritativeState = null
    this.predictedState = null
    this.interpolatedState = null
  }

  /**
   * Checks if authoritative state exists.
   */
  hasAuthoritative(): boolean {
    return this.authoritativeState !== null
  }

  /**
   * Checks if predicted state exists.
   */
  hasPredicted(): boolean {
    return this.predictedState !== null
  }

  /**
   * Checks if interpolated state exists.
   */
  hasInterpolated(): boolean {
    return this.interpolatedState !== null
  }
}

