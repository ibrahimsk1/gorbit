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
import { Radar } from './radar'
import type { SnapshotMessage } from '../net/protocol'
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
  private boundsSprite: Graphics | null = null
  private radar: Radar

  // World bounds constants (2000 m × 2000 m)
  private readonly WORLD_WIDTH = 2000
  private readonly WORLD_HEIGHT = 2000
  private readonly RADAR_WIDTH = 200
  private readonly RADAR_HEIGHT = 200

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

    // Create Radar with world bounds, 200×200px size, radar container from scene
    const radarLayer = scene.getLayer('radar')
    this.radar = new Radar(
      this.WORLD_WIDTH,
      this.WORLD_HEIGHT,
      this.RADAR_WIDTH,
      this.RADAR_HEIGHT,
      radarLayer
    )

    // Position radar in top-right corner
    const pixiAppScreen = pixiApp.screen
    radarLayer.x = pixiAppScreen.width - this.RADAR_WIDTH - 10 // 10px margin from edge
    radarLayer.y = 10 // 10px margin from top

    // Position HUD layer at top-left (0, 0) for absolute screen coordinates
    const hudLayer = scene.getLayer('hud')
    hudLayer.x = 0
    hudLayer.y = 0
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

    // Update world bounds visualization
    this.updateWorldBounds(state, gameLayer)

    // Update radar with latest snapshot and myShipId
    const snapshot = this.gameStateToSnapshot(state)
    this.radar.update(snapshot, state.myShipId)
  }

  /**
   * Converts GameState to SnapshotMessage format for Radar.
   * Radar needs SnapshotMessage, but Renderer only has GameState.
   */
  private gameStateToSnapshot(state: GameState): SnapshotMessage {
    return {
      t: 'snapshot',
      tick: state.tick,
      ships: state.ships,
      planets: state.planets,
      pallets: state.pallets,
      worldBounds: state.worldBounds,
      myShipId: state.myShipId,
      done: state.done,
      win: state.win
    }
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
   * Updates world bounds visualization (rectangle outline).
   * Uses world bounds from state or falls back to constants.
   */
  private updateWorldBounds(state: GameState, gameLayer: typeof gameLayer): void {
    const backgroundLayer = this.scene.getLayer('background')
    
    // Use world bounds from state, or fall back to constants
    const worldWidth = state.worldBounds?.width ?? this.WORLD_WIDTH
    const worldHeight = state.worldBounds?.height ?? this.WORLD_HEIGHT

    // World bounds rectangle is centered at origin
    // Top-left corner: (-width/2, height/2) in world coordinates
    // Bottom-right corner: (width/2, -height/2) in world coordinates
    const topLeft = this.worldToScreen(-worldWidth / 2, worldHeight / 2)
    const topRight = this.worldToScreen(worldWidth / 2, worldHeight / 2)
    const bottomRight = this.worldToScreen(worldWidth / 2, -worldHeight / 2)
    const bottomLeft = this.worldToScreen(-worldWidth / 2, -worldHeight / 2)

    if (!this.boundsSprite) {
      // Create new bounds sprite
      this.boundsSprite = new Graphics()
      backgroundLayer.addChild(this.boundsSprite)
    }

    // Clear and redraw bounds rectangle outline
    this.boundsSprite.clear()
    this.boundsSprite.moveTo(topLeft.x, topLeft.y)
    this.boundsSprite.lineTo(topRight.x, topRight.y)
    this.boundsSprite.lineTo(bottomRight.x, bottomRight.y)
    this.boundsSprite.lineTo(bottomLeft.x, bottomLeft.y)
    this.boundsSprite.lineTo(topLeft.x, topLeft.y)
    
    // Stroke outline (not fill)
    this.boundsSprite.stroke({ width: 2, color: 0xffffff, alpha: 0.5 })
  }

  /**
   * Clears all sprites from scene.
   */
  clear(): void {
    const gameLayer = this.scene.getLayer('game')
    const backgroundLayer = this.scene.getLayer('background')

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

    // Destroy world bounds sprite
    if (this.boundsSprite) {
      if (this.boundsSprite.parent) {
        this.boundsSprite.parent.removeChild(this.boundsSprite)
      }
      this.boundsSprite.destroy()
      this.boundsSprite = null
    }

    // Destroy radar
    if (this.radar) {
      this.radar.destroy()
    }
  }

  /**
   * Destroys renderer and cleans up all resources.
   */
  destroy(): void {
    this.clear()
  }
}

