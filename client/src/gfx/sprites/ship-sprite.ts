/**
 * Ship sprite factory for creating and updating ship sprites.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { Graphics } from 'pixi.js'
import type { ShipSnapshot } from '../../net/protocol'

/**
 * Factory for creating and managing Ship sprites.
 */
export class ShipSpriteFactory {
  /**
   * Creates a new ship sprite from ship state.
   * 
   * @param shipState Ship state snapshot
   * @returns Pixi Graphics sprite representing the ship
   */
  static create(shipState: ShipSnapshot): Graphics {
    const sprite = new Graphics()
    
    // Draw a more detailed ship shape with clear front indication
    // Ship points to the right (0 rotation = pointing right)
    const length = 18  // Ship length (front to back) - 1.5x original
    const width = 12   // Ship width - 1.5x original
    
    // Main body (pointed front, wider back)
    sprite.moveTo(length / 2, 0)           // Front point (nose)
    sprite.lineTo(-length / 4, -width / 2) // Back left
    sprite.lineTo(-length / 2, 0)          // Back center
    sprite.lineTo(-length / 4, width / 2)   // Back right
    sprite.lineTo(length / 2, 0)           // Close to front
    sprite.fill(0x00ff00) // Green body
    
    // Add outline for better visibility
    sprite.stroke({ width: 1, color: 0x00aa00 })
    
    // Add a small indicator at the front to make direction obvious
    sprite.circle(length / 2, 0, 3) // Slightly larger indicator for bigger ship
    sprite.fill(0xffffff) // White front indicator
    
    // Set position and rotation
    sprite.x = shipState.pos.x
    sprite.y = shipState.pos.y
    sprite.rotation = shipState.rot
    
    return sprite
  }

  /**
   * Updates an existing ship sprite from ship state.
   * 
   * @param sprite Existing ship sprite
   * @param shipState Ship state snapshot
   */
  static update(sprite: Graphics, shipState: ShipSnapshot): void {
    sprite.x = shipState.pos.x
    sprite.y = shipState.pos.y
    sprite.rotation = shipState.rot
  }

  /**
   * Destroys a ship sprite and removes it from its parent.
   * 
   * @param sprite Ship sprite to destroy
   */
  static destroy(sprite: Graphics): void {
    if (sprite.parent) {
      sprite.parent.removeChild(sprite)
    }
    sprite.destroy()
  }
}

