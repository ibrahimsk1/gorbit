/**
 * Snapshot interpolation system for smooth rendering between server snapshots.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import { StateManager } from './state-manager'
import type { GameState } from './state-manager'
import type { SnapshotMessage, Vec2Snapshot } from '../net/protocol'

/**
 * Snapshot entry in the buffer with timestamp.
 */
interface SnapshotEntry {
  snapshot: SnapshotMessage
  timestamp: number
}

/**
 * Interpolation system that smooths between server snapshots using a buffer.
 * 
 * The system maintains a buffer of snapshots (100-150ms worth) and interpolates
 * between them to provide smooth rendering. It interpolates into the past by
 * the buffer amount to account for network jitter.
 */
export class InterpolationSystem {
  private stateManager: StateManager
  private bufferMs: number
  private snapshots: SnapshotEntry[] = []
  private maxBufferSize: number = 10 // Maximum snapshots to keep in buffer

  /**
   * Creates a new interpolation system.
   * 
   * @param stateManager State manager for accessing authoritative state and updating interpolated state
   * @param bufferMs Buffer duration in milliseconds (default: 125ms, range: 100-150ms)
   */
  constructor(stateManager: StateManager, bufferMs: number = 125) {
    this.stateManager = stateManager
    this.bufferMs = Math.max(100, Math.min(150, bufferMs)) // Clamp to 100-150ms
  }

  /**
   * Adds a snapshot to the buffer with timestamp.
   * Removes old snapshots that are beyond buffer duration relative to the new snapshot.
   * 
   * @param snapshot Server snapshot to add
   * @param timestamp Timestamp in milliseconds (use performance.now())
   */
  addSnapshot(snapshot: SnapshotMessage, timestamp: number): void {
    // Check if snapshot with same tick exists (replace with newer)
    const existingIndex = this.snapshots.findIndex(entry => entry.snapshot.tick === snapshot.tick)
    if (existingIndex >= 0) {
      // Replace existing snapshot with same tick
      this.snapshots[existingIndex] = { snapshot, timestamp }
    } else {
      // Add new snapshot
      this.snapshots.push({ snapshot, timestamp })
    }

    // Sort by timestamp (oldest first), then by tick (older tick first when timestamps equal)
    this.snapshots.sort((a, b) => {
      // If timestamps are equal, prefer older tick first (so newer tick comes after in array)
      // This way when we iterate, we find older first, then newer
      if (a.timestamp === b.timestamp) {
        return a.snapshot.tick - b.snapshot.tick
      }
      return a.timestamp - b.timestamp
    })

    // Remove snapshots older than buffer duration (only if we have more than 2 snapshots)
    // This prevents removing snapshots needed for interpolation when we have few snapshots
    if (this.snapshots.length > 2) {
      const cutoffTime = timestamp - this.bufferMs
      this.snapshots = this.snapshots.filter(entry => entry.timestamp >= cutoffTime)
    }

    // Limit buffer size (keep newest snapshots)
    if (this.snapshots.length > this.maxBufferSize) {
      this.snapshots = this.snapshots.slice(-this.maxBufferSize)
    }
  }

  /**
   * Updates interpolated state based on current time.
   * Interpolates between snapshots in the buffer.
   * Also removes old snapshots beyond buffer duration.
   * 
   * @param currentTime Current time in milliseconds (use performance.now())
   */
  update(currentTime: number): void {
    if (this.snapshots.length === 0) {
      // No snapshots, nothing to interpolate
      return
    }

    // Calculate target time (interpolate into the past by buffer amount)
    const targetTime = currentTime - this.bufferMs

    // Remove snapshots that are too old (older than target time by more than buffer duration)
    // Keep at least 2 snapshots if available for interpolation
    const minKeepTime = targetTime - this.bufferMs
    if (this.snapshots.length > 2) {
      this.snapshots = this.snapshots.filter(entry => entry.timestamp >= minKeepTime)
    } else {
      // Keep all snapshots if we have 2 or fewer (need them for interpolation)
      this.snapshots = this.snapshots.filter(entry => entry.timestamp >= minKeepTime)
    }

    if (this.snapshots.length === 0) {
      // All snapshots were too old
      return
    }

    if (this.snapshots.length === 1) {
      // Only one snapshot, use it directly
      const entry = this.snapshots[0]
      const interpolatedState = this.snapshotToGameState(entry.snapshot)
      this.stateManager.updateInterpolated(interpolatedState)
      return
    }

    // Find two snapshots that bracket the target time
    let olderIndex = -1
    let newerIndex = -1

    for (let i = 0; i < this.snapshots.length - 1; i++) {
      const older = this.snapshots[i]
      const newer = this.snapshots[i + 1]

      // Handle case where timestamps are equal - use newer snapshot (higher tick)
      if (older.timestamp === newer.timestamp) {
        // If target time matches the timestamp, use the newer snapshot (higher tick)
        if (targetTime === older.timestamp) {
          const interpolatedState = this.snapshotToGameState(newer.snapshot)
          this.stateManager.updateInterpolated(interpolatedState)
          return
        }
        // Continue searching if target time doesn't match
        continue
      }

      if (targetTime >= older.timestamp && targetTime <= newer.timestamp) {
        olderIndex = i
        newerIndex = i + 1
        break
      }
    }

    // Handle extrapolation guards
    if (olderIndex === -1) {
      // Target time is before oldest snapshot or after newest snapshot
      if (targetTime < this.snapshots[0].timestamp) {
        // Use oldest snapshot
        const entry = this.snapshots[0]
        const interpolatedState = this.snapshotToGameState(entry.snapshot)
        this.stateManager.updateInterpolated(interpolatedState)
        return
      } else {
        // Use newest snapshot
        const entry = this.snapshots[this.snapshots.length - 1]
        const interpolatedState = this.snapshotToGameState(entry.snapshot)
        this.stateManager.updateInterpolated(interpolatedState)
        return
      }
    }

    // Interpolate between two snapshots
    const olderEntry = this.snapshots[olderIndex]
    const newerEntry = this.snapshots[newerIndex]

    // Calculate interpolation factor
    const timeDiff = newerEntry.timestamp - olderEntry.timestamp
    let factor: number
    
    if (timeDiff <= 0) {
      // Same timestamp or out of order - use newer snapshot
      factor = 1.0
    } else {
      factor = Math.max(0, Math.min(1, (targetTime - olderEntry.timestamp) / timeDiff))
    }

    // Interpolate states
    const interpolatedState = this.interpolateStates(
      olderEntry.snapshot,
      newerEntry.snapshot,
      factor
    )

    this.stateManager.updateInterpolated(interpolatedState)
  }

  /**
   * Clears the snapshot buffer.
   */
  clear(): void {
    this.snapshots = []
  }

  /**
   * Gets the number of snapshots in the buffer.
   * 
   * @returns Number of snapshots in buffer
   */
  getBufferSize(): number {
    return this.snapshots.length
  }

  /**
   * Checks if buffer has enough data for interpolation (at least 2 snapshots).
   * 
   * @returns True if buffer has enough data, false otherwise
   */
  hasEnoughData(): boolean {
    return this.snapshots.length >= 2
  }

  /**
   * Converts a SnapshotMessage to GameState format.
   */
  private snapshotToGameState(snapshot: SnapshotMessage): GameState {
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
   * Interpolates between two snapshots.
   * 
   * @param older Older snapshot
   * @param newer Newer snapshot
   * @param factor Interpolation factor (0.0 = older, 1.0 = newer)
   * @returns Interpolated game state
   */
  private interpolateStates(
    older: SnapshotMessage,
    newer: SnapshotMessage,
    factor: number
  ): GameState {
    // Interpolate ship
    const ship = {
      pos: this.lerpVec2(older.ship.pos, newer.ship.pos, factor),
      vel: this.lerpVec2(older.ship.vel, newer.ship.vel, factor),
      rot: this.lerpAngle(older.ship.rot, newer.ship.rot, factor),
      energy: this.lerp(older.ship.energy, newer.ship.energy, factor)
    }

    // Interpolate planets (match by index)
    const planets = newer.planets.map((newPlanet, index) => {
      const oldPlanet = older.planets[index]
      if (oldPlanet) {
        return {
          pos: this.lerpVec2(oldPlanet.pos, newPlanet.pos, factor),
          radius: newPlanet.radius // Use newer radius (discrete)
        }
      }
      // New planet, use newer snapshot
      return newPlanet
    })

    // Interpolate pallets (match by id)
    const pallets: typeof newer.pallets = []
    
    // Process pallets from newer snapshot
    for (const newPallet of newer.pallets) {
      const oldPallet = older.pallets.find(p => p.id === newPallet.id)
      if (oldPallet) {
        // Interpolate position, use newer active state (discrete)
        pallets.push({
          id: newPallet.id,
          pos: this.lerpVec2(oldPallet.pos, newPallet.pos, factor),
          active: newPallet.active // Use newer active state (discrete)
        })
      } else {
        // New pallet, use newer snapshot
        pallets.push(newPallet)
      }
    }

    // Use discrete values for game state flags
    return {
      tick: newer.tick, // Use newer tick
      ship,
      planets,
      pallets,
      done: newer.done, // Use newer done state (discrete)
      win: newer.win // Use newer win state (discrete)
    }
  }

  /**
   * Linear interpolation between two numbers.
   * 
   * @param a Start value
   * @param b End value
   * @param t Interpolation factor (0.0 = a, 1.0 = b)
   * @returns Interpolated value
   */
  private lerp(a: number, b: number, t: number): number {
    const clampedT = Math.max(0, Math.min(1, t))
    return a + (b - a) * clampedT
  }

  /**
   * Linear interpolation between two Vec2 values.
   * 
   * @param a Start vector
   * @param b End vector
   * @param t Interpolation factor (0.0 = a, 1.0 = b)
   * @returns Interpolated vector
   */
  private lerpVec2(a: Vec2Snapshot, b: Vec2Snapshot, t: number): Vec2Snapshot {
    return {
      x: this.lerp(a.x, b.x, t),
      y: this.lerp(a.y, b.y, t)
    }
  }

  /**
   * Linear interpolation between two angles, handling wrap-around.
   * Finds the shortest path between angles.
   * 
   * @param a Start angle in radians
   * @param b End angle in radians
   * @param t Interpolation factor (0.0 = a, 1.0 = b)
   * @returns Interpolated angle in [0, 2π)
   */
  private lerpAngle(a: number, b: number, t: number): number {
    // Normalize angles to [0, 2π)
    const normalize = (angle: number): number => {
      angle = angle % (2 * Math.PI)
      if (angle < 0) {
        angle += 2 * Math.PI
      }
      return angle
    }

    const aNorm = normalize(a)
    const bNorm = normalize(b)

    // Find shortest path (considering wrap-around)
    let diff = bNorm - aNorm
    if (diff > Math.PI) {
      diff -= 2 * Math.PI
    } else if (diff < -Math.PI) {
      diff += 2 * Math.PI
    }

    // Interpolate along shortest path
    const result = aNorm + diff * Math.max(0, Math.min(1, t))
    return normalize(result)
  }
}

