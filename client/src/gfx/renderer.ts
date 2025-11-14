/**
 * Renderer system for updating Pixi sprites from game state.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { Graphics } from 'pixi.js'
import { StateManager, type GameState } from '../sim/state-manager'
import { Scene } from './scene'
import { App } from '../core/app'
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
  private app: App
  private shipSprite: Graphics | null = null
  private planetSprites: Map<number, Graphics> = new Map()
  private palletSprites: Map<number, Graphics> = new Map()

  constructor(stateManager: StateManager, scene: Scene, app: App) {
    this.stateManager = stateManager
    this.scene = scene
    this.app = app
  }

  /**
   * Transforms world coordinates to screen coordinates.
   * World (0,0) maps to screen center.
   * Y-axis is flipped because screen Y increases downward, while world Y increases upward.
   */
  private worldToScreen(worldX: number, worldY: number): { x: number, y: number } {
    const pixiApp = this.app.getApplication()
    const screenWidth = pixiApp.screen.width
    const screenHeight = pixiApp.screen.height
    
    return {
      x: worldX + screenWidth / 2,
      y: -worldY + screenHeight / 2  // Flip Y-axis
    }
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
    const screenPos = this.worldToScreen(ship.pos.x, ship.pos.y)
    const transformedShip = {
      ...ship,
      pos: { x: screenPos.x, y: screenPos.y }
    }
    
    if (!this.shipSprite) {
      this.shipSprite = ShipSpriteFactory.create(transformedShip)
      gameLayer.addChild(this.shipSprite)
    } else {
      ShipSpriteFactory.update(this.shipSprite, transformedShip)
    }
  }

  /**
   * Updates planet sprites from planets array (generic iteration, match by index).
   */
  private updatePlanetSprites(planets: GameState['planets'], gameLayer: typeof gameLayer): void {
    // Create/update sprites for planets in array
    planets.forEach((planet, index) => {
      const screenPos = this.worldToScreen(planet.pos.x, planet.pos.y)
      const transformedPlanet = {
        ...planet,
        pos: { x: screenPos.x, y: screenPos.y }
      }
      
      let sprite = this.planetSprites.get(index)
      if (!sprite) {
        sprite = PlanetSpriteFactory.create(transformedPlanet)
        this.planetSprites.set(index, sprite)
        gameLayer.addChild(sprite)
      } else {
        PlanetSpriteFactory.update(sprite, transformedPlanet)
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
      const screenPos = this.worldToScreen(pallet.pos.x, pallet.pos.y)
      const transformedPallet = {
        ...pallet,
        pos: { x: screenPos.x, y: screenPos.y }
      }
      
      let sprite = this.palletSprites.get(pallet.id)
      if (!sprite) {
        sprite = PalletSpriteFactory.create(transformedPallet)
        this.palletSprites.set(pallet.id, sprite)
        gameLayer.addChild(sprite)
      } else {
        PalletSpriteFactory.update(sprite, transformedPallet)
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

