/**
 * Radar system showing full world view with all entities.
 * 
 * Labels: scope:integration loop:g6-rendering layer:gfx dep:state b:radar-minimap r:high
 */

import { Graphics, Container } from 'pixi.js'
import type { SnapshotMessage, ShipSnapshot, PlanetSnapshot, PalletSnapshot } from '../net/protocol'

/**
 * Radar class that displays a mini-map showing all entities in the world.
 * Shows full world view (2000 m × 2000 m) in fixed 200×200 px area.
 */
export class Radar {
  private worldWidth: number
  private worldHeight: number
  private radarWidth: number
  private radarHeight: number
  private container: Container
  private graphics: Graphics | null = null

  /**
   * Creates a new Radar instance.
   * 
   * @param worldWidth - World width in meters (e.g., 2000.0)
   * @param worldHeight - World height in meters (e.g., 2000.0)
   * @param radarWidth - Radar width in pixels (e.g., 200)
   * @param radarHeight - Radar height in pixels (e.g., 200)
   * @param container - PixiJS container for radar graphics
   */
  constructor(
    worldWidth: number,
    worldHeight: number,
    radarWidth: number,
    radarHeight: number,
    container: Container
  ) {
    this.worldWidth = worldWidth
    this.worldHeight = worldHeight
    this.radarWidth = radarWidth
    this.radarHeight = radarHeight
    this.container = container
  }

  /**
   * Maps world coordinates to radar pixel coordinates.
   * World bounds [-worldWidth/2, worldWidth/2] × [-worldHeight/2, worldHeight/2] 
   * map to [0, radarWidth] × [0, radarHeight] pixels.
   * 
   * @param worldX - World X coordinate
   * @param worldY - World Y coordinate
   * @returns Radar pixel coordinates
   */
  private worldToRadar(worldX: number, worldY: number): { x: number, y: number } {
    // Map world X from [-worldWidth/2, worldWidth/2] to [0, radarWidth]
    const radarX = ((worldX - (-this.worldWidth / 2)) / this.worldWidth) * this.radarWidth
    
    // Map world Y from [-worldHeight/2, worldHeight/2] to [0, radarHeight]
    // Note: Y-axis is flipped (world Y-up, radar Y-down)
    const radarY = ((worldY - (-this.worldHeight / 2)) / this.worldHeight) * this.radarHeight
    // Flip Y: radarY = radarHeight - radarY
    const flippedY = this.radarHeight - radarY
    
    return { x: radarX, y: flippedY }
  }

  /**
   * Updates radar with latest snapshot data.
   * Clears and redraws all graphics each frame.
   * 
   * @param snapshot - Server snapshot with game state
   * @param myShipId - Player's ship ID for highlighting
   */
  update(snapshot: SnapshotMessage, myShipId: number): void {
    // Clear previous graphics
    this.clear()
    
    // Create new graphics object
    this.graphics = new Graphics()
    this.container.addChild(this.graphics)
    
    // Render world bounds rectangle outline
    this.renderWorldBounds()
    
    // Render all planets as circles (size proportional to radius, color-coded)
    this.renderPlanets(snapshot.planets)
    
    // Render all ships as dots (color-coded by player ID, own ship highlighted)
    this.renderShips(snapshot.ships, myShipId)
    
    // Render active pallets as small dots (distinct color)
    this.renderPallets(snapshot.pallets)
  }

  /**
   * Renders world bounds as rectangle outline with gray background.
   */
  private renderWorldBounds(): void {
    if (!this.graphics) return
    
    // Draw gray background covering entire radar area
    this.graphics.rect(0, 0, this.radarWidth, this.radarHeight)
    this.graphics.fill(0x808080) // Gray background
    
    // Draw rectangle outline on top
    this.graphics.rect(0, 0, this.radarWidth, this.radarHeight)
    this.graphics.stroke({ width: 2, color: 0xffffff }) // White outline for contrast
  }

  /**
   * Renders all planets as circles.
   * Size proportional to radius, color-coded.
   * 
   * @param planets - Array of planet snapshots
   */
  private renderPlanets(planets: PlanetSnapshot[]): void {
    if (!this.graphics) return
    
    planets.forEach((planet) => {
      const radarPos = this.worldToRadar(planet.pos.x, planet.pos.y)
      
      // Scale radius to radar size (proportional to world size)
      // Max planet radius in world is ~100, radar is 200px, so scale factor ~2
      const scaleFactor = this.radarWidth / this.worldWidth
      const radarRadius = Math.max(2, planet.radius * scaleFactor) // Minimum 2px
      
      // Color-code by planet ID (use different hues)
      const hue = (planet.id * 60) % 360 // Spread colors
      const color = this.hslToHex(hue, 70, 50)
      
      this.graphics!.circle(radarPos.x, radarPos.y, radarRadius)
      this.graphics!.fill(color)
    })
  }

  /**
   * Renders all ships as dots.
   * Color-coded by player ID, own ship highlighted differently.
   * 
   * @param ships - Array of ship snapshots
   * @param myShipId - Player's ship ID
   */
  private renderShips(ships: ShipSnapshot[], myShipId: number): void {
    if (!this.graphics) return
    
    ships.forEach((ship) => {
      const radarPos = this.worldToRadar(ship.pos.x, ship.pos.y)
      
      if (ship.id === myShipId) {
        // Own ship: larger, brighter, with outline
        this.graphics!.circle(radarPos.x, radarPos.y, 4)
        this.graphics!.fill(0x00ff00) // Bright green
        this.graphics!.circle(radarPos.x, radarPos.y, 4)
        this.graphics!.stroke({ width: 1, color: 0xffffff }) // White outline
      } else {
        // Other ships: smaller, color-coded by player ID
        const hue = (ship.id * 60) % 360 // Spread colors
        const color = this.hslToHex(hue, 70, 50)
        
        this.graphics!.circle(radarPos.x, radarPos.y, 2)
        this.graphics!.fill(color)
      }
    })
  }

  /**
   * Renders active pallets as small dots.
   * Distinct color from ships and planets.
   * 
   * @param pallets - Array of pallet snapshots
   */
  private renderPallets(pallets: PalletSnapshot[]): void {
    if (!this.graphics) return
    
    pallets.forEach((pallet) => {
      if (!pallet.active) return // Only render active pallets
      
      const radarPos = this.worldToRadar(pallet.pos.x, pallet.pos.y)
      
      // Small yellow dot for pallets
      this.graphics!.circle(radarPos.x, radarPos.y, 1.5)
      this.graphics!.fill(0xffff00) // Yellow
    })
  }

  /**
   * Converts HSL color to hex.
   * 
   * @param h - Hue (0-360)
   * @param s - Saturation (0-100)
   * @param l - Lightness (0-100)
   * @returns Hex color value
   */
  private hslToHex(h: number, s: number, l: number): number {
    h /= 360
    s /= 100
    l /= 100
    
    const c = (1 - Math.abs(2 * l - 1)) * s
    const x = c * (1 - Math.abs((h * 6) % 2 - 1))
    const m = l - c / 2
    
    let r = 0, g = 0, b = 0
    
    if (h < 1/6) {
      r = c; g = x; b = 0
    } else if (h < 2/6) {
      r = x; g = c; b = 0
    } else if (h < 3/6) {
      r = 0; g = c; b = x
    } else if (h < 4/6) {
      r = 0; g = x; b = c
    } else if (h < 5/6) {
      r = x; g = 0; b = c
    } else {
      r = c; g = 0; b = x
    }
    
    r = Math.round((r + m) * 255)
    g = Math.round((g + m) * 255)
    b = Math.round((b + m) * 255)
    
    return (r << 16) | (g << 8) | b
  }

  /**
   * Clears radar graphics.
   */
  private clear(): void {
    if (this.graphics) {
      this.graphics.clear()
      if (this.graphics.parent) {
        this.graphics.parent.removeChild(this.graphics)
      }
      this.graphics.destroy()
      this.graphics = null
    }
  }

  /**
   * Destroys radar and cleans up all resources.
   */
  destroy(): void {
    this.clear()
  }
}

