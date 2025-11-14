/**
 * Planet sprite factory for creating and updating planet sprites.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { Graphics } from 'pixi.js'
import type { PlanetSnapshot } from '../../net/protocol'

/**
 * Factory for creating and managing Planet sprites.
 */
export class PlanetSpriteFactory {
  /**
   * Creates a new planet sprite from planet state.
   * 
   * @param planetState Planet state snapshot
   * @returns Pixi Graphics sprite representing the planet
   */
  static create(planetState: PlanetSnapshot): Graphics {
    const sprite = new Graphics()
    
    // Draw circle for planet
    sprite.circle(0, 0, planetState.radius)
    sprite.fill(0xffaa00) // Orange/yellow color
    
    // Set position
    sprite.x = planetState.pos.x
    sprite.y = planetState.pos.y
    
    return sprite
  }

  /**
   * Updates an existing planet sprite from planet state.
   * 
   * @param sprite Existing planet sprite
   * @param planetState Planet state snapshot
   */
  static update(sprite: Graphics, planetState: PlanetSnapshot): void {
    sprite.x = planetState.pos.x
    sprite.y = planetState.pos.y
    // Note: Radius changes would require redrawing, but planets typically don't change size
    // For now, we only update position
  }

  /**
   * Destroys a planet sprite and removes it from its parent.
   * 
   * @param sprite Planet sprite to destroy
   */
  static destroy(sprite: Graphics): void {
    if (sprite.parent) {
      sprite.parent.removeChild(sprite)
    }
    sprite.destroy()
  }
}

