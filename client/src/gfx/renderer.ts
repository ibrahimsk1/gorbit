/**
 * Renderer system for updating Pixi sprites from game state.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { Graphics } from 'pixi.js'
import { StateManager, type GameState } from '../sim/state-manager'
import { Scene } from './scene'
import { ShipSpriteFactory } from './sprites/ship-sprite'
import { PlanetSpriteFactory } from './sprites/planet-sprite'
import { PalletSpriteFactory } from './sprites/pallet-sprite'

/**
 * Renderer that updates Pixi sprites from game state.
 * Uses generic array-based iteration pattern for extensibility.
 */
export class Renderer {
  private stateManager: StateManager
  private scene: Scene
  private shipSprite: Graphics | null = null
  private planetSprites: Map<number, Graphics> = new Map()
  private palletSprites: Map<number, Graphics> = new Map()

  constructor(stateManager: StateManager, scene: Scene) {
    this.stateManager = stateManager
    this.scene = scene
  }

  /**
   * Updates all sprites from current game state.
   */
  update(): void {
    const state = this.stateManager.getRenderState()
    const gameLayer = this.scene.getLayer('game')

    // Update ship sprite
    this.updateShipSprite(state.ship, gameLayer)

    // Update planet sprites (generic array iteration, match by index)
    this.updatePlanetSprites(state.planets, gameLayer)

    // Update pallet sprites (generic array iteration, match by id)
    this.updatePalletSprites(state.pallets, gameLayer)
  }

  /**
   * Updates ship sprite from ship state.
   */
  private updateShipSprite(ship: GameState['ship'], gameLayer: typeof gameLayer): void {
    if (!this.shipSprite) {
      this.shipSprite = ShipSpriteFactory.create(ship)
      gameLayer.addChild(this.shipSprite)
    } else {
      ShipSpriteFactory.update(this.shipSprite, ship)
    }
  }

  /**
   * Updates planet sprites from planets array (generic iteration, match by index).
   */
  private updatePlanetSprites(planets: GameState['planets'], gameLayer: typeof gameLayer): void {
    // Create/update sprites for planets in array
    planets.forEach((planet, index) => {
      let sprite = this.planetSprites.get(index)
      if (!sprite) {
        sprite = PlanetSpriteFactory.create(planet)
        this.planetSprites.set(index, sprite)
        gameLayer.addChild(sprite)
      } else {
        PlanetSpriteFactory.update(sprite, planet)
      }
    })

    // Remove sprites for planets no longer in array
    const currentIndices = new Set(planets.map((_, index) => index))
    for (const [index, sprite] of this.planetSprites.entries()) {
      if (!currentIndices.has(index)) {
        PlanetSpriteFactory.destroy(sprite)
        this.planetSprites.delete(index)
      }
    }
  }

  /**
   * Updates pallet sprites from pallets array (generic iteration, match by id).
   */
  private updatePalletSprites(pallets: GameState['pallets'], gameLayer: typeof gameLayer): void {
    // Create/update sprites for pallets in array
    pallets.forEach((pallet) => {
      let sprite = this.palletSprites.get(pallet.id)
      if (!sprite) {
        sprite = PalletSpriteFactory.create(pallet)
        this.palletSprites.set(pallet.id, sprite)
        gameLayer.addChild(sprite)
      } else {
        PalletSpriteFactory.update(sprite, pallet)
      }
    })

    // Remove sprites for pallets no longer in array
    const currentIds = new Set(pallets.map(p => p.id))
    for (const [id, sprite] of this.palletSprites.entries()) {
      if (!currentIds.has(id)) {
        PalletSpriteFactory.destroy(sprite)
        this.palletSprites.delete(id)
      }
    }
  }

  /**
   * Clears all sprites from scene.
   */
  clear(): void {
    const gameLayer = this.scene.getLayer('game')

    // Destroy ship sprite
    if (this.shipSprite) {
      ShipSpriteFactory.destroy(this.shipSprite)
      this.shipSprite = null
    }

    // Destroy planet sprites
    for (const sprite of this.planetSprites.values()) {
      PlanetSpriteFactory.destroy(sprite)
    }
    this.planetSprites.clear()

    // Destroy pallet sprites
    for (const sprite of this.palletSprites.values()) {
      PalletSpriteFactory.destroy(sprite)
    }
    this.palletSprites.clear()
  }

  /**
   * Destroys renderer and cleans up all resources.
   */
  destroy(): void {
    this.clear()
  }
}

