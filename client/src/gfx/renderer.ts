/**
 * Renderer system for updating Pixi sprites from game state.
 * 
 * Labels: scope:integration loop:g6-rendering layer:gfx dep:state
 */

import { Graphics } from 'pixi.js'
import { StateManager, type GameState } from '../sim/state-manager'
import { Scene } from './scene'
import { App } from '../core/app'
import { Camera } from './camera'
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
  private camera: Camera
  private shipSprites: Map<number, Graphics> = new Map()
  private planetSprites: Map<number, Graphics> = new Map()
  private palletSprites: Map<number, Graphics> = new Map()

  // World bounds constants (2000 m × 2000 m)
  private readonly WORLD_WIDTH = 2000
  private readonly WORLD_HEIGHT = 2000

  constructor(stateManager: StateManager, scene: Scene, app: App) {
    this.stateManager = stateManager
    this.scene = scene
    this.app = app
    
    // Create Camera with world bounds and viewport from app
    const pixiApp = app.getApplication()
    this.camera = new Camera(
      this.WORLD_WIDTH,
      this.WORLD_HEIGHT,
      pixiApp.screen.width,
      pixiApp.screen.height,
      0.1 // Default lerp factor
    )
  }

  /**
   * Transforms world coordinates to screen coordinates using camera position.
   * World (0,0) maps to screen center adjusted by camera position.
   * Y-axis is flipped because screen Y increases downward, while world Y increases upward.
   * Formula: screenX = worldX - cameraX + screenWidth/2, screenY = -(worldY - cameraY) + screenHeight/2
   */
  private worldToScreen(worldX: number, worldY: number): { x: number, y: number } {
    const pixiApp = this.app.getApplication()
    const screenWidth = pixiApp.screen.width
    const screenHeight = pixiApp.screen.height
    const cameraPos = this.camera.getPosition()
    
    return {
      x: worldX - cameraPos.x + screenWidth / 2,
      y: -(worldY - cameraPos.y) + screenHeight / 2  // Flip Y-axis with camera offset
    }
  }

  /**
   * Updates all sprites from current game state.
   * Camera is updated first to follow player's ship, then sprites are rendered.
   * v1 multiplayer format: handles multiple ships from ships array.
   */
  update(): void {
    const state = this.stateManager.getRenderState()
    const gameLayer = this.scene.getLayer('game')

    // Find player's ship for camera following
    const playerShip = state.ships.find(ship => ship.id === state.myShipId)
    if (playerShip) {
      // Update camera to follow player's ship (call before rendering)
      this.camera.update(playerShip.pos)
    }

    // Update ship sprites (v1 multiplayer format: ships array, match by ID)
    this.updateShipSprites(state.ships, gameLayer)

    // Update planet sprites (generic array iteration, match by index)
    this.updatePlanetSprites(state.planets, gameLayer)

    // Update pallet sprites (generic array iteration, match by id)
    this.updatePalletSprites(state.pallets, gameLayer)
  }

  /**
   * Updates ship sprites from ships array (v1 multiplayer format, match by ID).
   */
  private updateShipSprites(ships: GameState['ships'], gameLayer: typeof gameLayer): void {
    // Create/update sprites for ships in array
    ships.forEach((ship) => {
      const screenPos = this.worldToScreen(ship.pos.x, ship.pos.y)
      const transformedShip = {
        ...ship,
        pos: { x: screenPos.x, y: screenPos.y }
      }
      
      let sprite = this.shipSprites.get(ship.id)
      if (!sprite) {
        sprite = ShipSpriteFactory.create(transformedShip)
        this.shipSprites.set(ship.id, sprite)
        gameLayer.addChild(sprite)
      } else {
        ShipSpriteFactory.update(sprite, transformedShip)
      }
    })

    // Remove sprites for ships no longer in array
    const currentIds = new Set(ships.map(s => s.id))
    for (const [id, sprite] of this.shipSprites.entries()) {
      if (!currentIds.has(id)) {
        ShipSpriteFactory.destroy(sprite)
        this.shipSprites.delete(id)
      }
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

    // Destroy ship sprites (v1 multiplayer format: multiple ships)
    for (const sprite of this.shipSprites.values()) {
      ShipSpriteFactory.destroy(sprite)
    }
    this.shipSprites.clear()

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

