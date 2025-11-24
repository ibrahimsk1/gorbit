/**
 * Camera system that smoothly follows player's ship with lerp smoothing.
 * 
 * Labels: scope:integration loop:g6-rendering layer:gfx dep:state b:camera-follow r:high
 */

import type { Vec2Snapshot } from '../net/protocol'

/**
 * Camera class that smoothly follows a target position with lerp smoothing.
 * Camera position is clamped to world bounds (no wraparound).
 */
export class Camera {
  private worldWidth: number
  private worldHeight: number
  private viewportWidth: number
  private viewportHeight: number
  private lerpFactor: number
  private position: Vec2Snapshot

  /**
   * Creates a new Camera instance.
   * 
   * @param worldWidth - World width in meters (e.g., 2000.0)
   * @param worldHeight - World height in meters (e.g., 2000.0)
   * @param viewportWidth - Viewport width in pixels
   * @param viewportHeight - Viewport height in pixels
   * @param lerpFactor - Lerp factor for smooth following (default: 0.1, range: 0.0-1.0)
   */
  constructor(
    worldWidth: number,
    worldHeight: number,
    viewportWidth: number,
    viewportHeight: number,
    lerpFactor: number = 0.1
  ) {
    this.worldWidth = worldWidth
    this.worldHeight = worldHeight
    this.viewportWidth = viewportWidth
    this.viewportHeight = viewportHeight
    this.lerpFactor = lerpFactor
    this.position = { x: 0, y: 0 }
  }

  /**
   * Updates camera position to smoothly follow target.
   * Camera position lerps toward target: cameraPos = cameraPos + (targetPos - cameraPos) * lerpFactor
   * Camera position is clamped to world bounds (stays within [-worldWidth/2, worldWidth/2] × [-worldHeight/2, worldHeight/2]).
   * 
   * @param targetPos - Target position (player ship position)
   */
  update(targetPos: Vec2Snapshot): void {
    // Lerp camera position toward target
    this.position.x = this.position.x + (targetPos.x - this.position.x) * this.lerpFactor
    this.position.y = this.position.y + (targetPos.y - this.position.y) * this.lerpFactor

    // Clamp camera position to world bounds (no wraparound)
    const halfWorldWidth = this.worldWidth / 2
    const halfWorldHeight = this.worldHeight / 2
    
    this.position.x = Math.max(-halfWorldWidth, Math.min(halfWorldWidth, this.position.x))
    this.position.y = Math.max(-halfWorldHeight, Math.min(halfWorldHeight, this.position.y))
  }

  /**
   * Gets current camera position.
   * 
   * @returns Camera position in world coordinates
   */
  getPosition(): Vec2Snapshot {
    return { x: this.position.x, y: this.position.y }
  }

  /**
   * Sets lerp factor for runtime adjustment.
   * 
   * @param factor - Lerp factor (range: 0.0-1.0)
   */
  setLerpFactor(factor: number): void {
    this.lerpFactor = factor
  }

  /**
   * Gets viewport dimensions.
   * 
   * @returns Viewport width and height
   */
  getViewport(): { width: number, height: number } {
    return {
      width: this.viewportWidth,
      height: this.viewportHeight
    }
  }
}

