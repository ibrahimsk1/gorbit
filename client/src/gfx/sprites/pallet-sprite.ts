/**
 * Pallet sprite factory for creating and updating pallet sprites.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { Graphics } from 'pixi.js'
import type { PalletSnapshot } from '../../net/protocol'

/**
 * Factory for creating and managing Pallet sprites.
 */
export class PalletSpriteFactory {
  /**
   * Creates a new pallet sprite from pallet state.
   * 
   * @param palletState Pallet state snapshot
   * @returns Pixi Graphics sprite representing the pallet
   */
  static create(palletState: PalletSnapshot): Graphics {
    const sprite = new Graphics()
    
    // Draw small circle/square for pallet
    const size = 5 // Pallet size
    sprite.circle(0, 0, size)
    sprite.fill(0x00ffff) // Cyan color
    
    // Set position and visibility based on active state
    sprite.x = palletState.pos.x
    sprite.y = palletState.pos.y
    sprite.visible = palletState.active
    
    return sprite
  }

  /**
   * Updates an existing pallet sprite from pallet state.
   * 
   * @param sprite Existing pallet sprite
   * @param palletState Pallet state snapshot
   */
  static update(sprite: Graphics, palletState: PalletSnapshot): void {
    sprite.x = palletState.pos.x
    sprite.y = palletState.pos.y
    sprite.visible = palletState.active
  }

  /**
   * Destroys a pallet sprite and removes it from its parent.
   * 
   * @param sprite Pallet sprite to destroy
   */
  static destroy(sprite: Graphics): void {
    if (sprite.parent) {
      sprite.parent.removeChild(sprite)
    }
    sprite.destroy()
  }
}

