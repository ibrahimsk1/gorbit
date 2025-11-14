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
    
    // Draw triangle shape for ship (pointing up by default)
    // Triangle: top point, bottom left, bottom right
    const size = 10 // Ship size
    sprite.moveTo(0, -size) // Top point
    sprite.lineTo(-size, size) // Bottom left
    sprite.lineTo(size, size) // Bottom right
    sprite.lineTo(0, -size) // Close triangle
    sprite.fill(0x00ff00) // Green color
    
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

