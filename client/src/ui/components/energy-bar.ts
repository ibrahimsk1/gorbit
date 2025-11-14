/**
 * Energy bar component for displaying ship energy level.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { Container, Graphics, Text } from 'pixi.js'

export interface EnergyBarConfig {
  x?: number
  y?: number
  width?: number
  height?: number
  backgroundColor?: number
  foregroundColor?: number
  showText?: boolean
}

/**
 * Energy bar component that displays ship energy as a visual bar.
 */
export class EnergyBar {
  private container: Container
  private backgroundBar: Graphics
  private foregroundBar: Graphics
  private textLabel: Text | null = null
  private config: Required<EnergyBarConfig>

  constructor(parent: Container, config: EnergyBarConfig = {}) {
    this.container = new Container()
    this.container.label = 'energy-bar'
    parent.addChild(this.container)

    // Default configuration
    this.config = {
      x: config.x ?? 20,
      y: config.y ?? 20,
      width: config.width ?? 200,
      height: config.height ?? 20,
      backgroundColor: config.backgroundColor ?? 0x333333,
      foregroundColor: config.foregroundColor ?? 0x00ff00,
      showText: config.showText ?? true
    }

    // Set container position
    this.container.x = this.config.x
    this.container.y = this.config.y

    // Create background bar (full width)
    this.backgroundBar = new Graphics()
    this.backgroundBar.label = 'energy-background'
    this.backgroundBar.rect(0, 0, this.config.width, this.config.height)
    this.backgroundBar.fill(this.config.backgroundColor)
    this.container.addChild(this.backgroundBar)

    // Create foreground bar (width based on energy)
    this.foregroundBar = new Graphics()
    this.foregroundBar.label = 'energy-foreground'
    this.foregroundBar.rect(0, 0, 0, this.config.height)
    this.foregroundBar.fill(this.config.foregroundColor)
    this.container.addChild(this.foregroundBar)

    // Create text label if enabled
    if (this.config.showText) {
      this.textLabel = new Text({
        text: 'Energy: 100',
        style: {
          fontFamily: 'Arial',
          fontSize: 12,
          fill: 0xffffff
        }
      })
      this.textLabel.label = 'energy-text'
      this.textLabel.x = this.config.width + 10
      this.textLabel.y = (this.config.height - this.textLabel.height) / 2
      this.container.addChild(this.textLabel)
    }
  }

  /**
   * Updates the energy bar based on current energy level.
   * 
   * @param energy Current energy value
   * @param maxEnergy Maximum energy value (default: 100.0)
   */
  update(energy: number, maxEnergy: number = 100.0): void {
    // Clamp energy to valid range
    const clampedEnergy = Math.max(0, Math.min(energy, maxEnergy))
    
    // Calculate energy ratio
    const ratio = maxEnergy > 0 ? clampedEnergy / maxEnergy : 0
    
    // Update foreground bar width
    const newWidth = this.config.width * ratio
    
    // Clear and redraw foreground bar
    this.foregroundBar.clear()
    this.foregroundBar.rect(0, 0, newWidth, this.config.height)
    
    // Update color based on energy level (green -> yellow -> red)
    let color = this.config.foregroundColor
    if (ratio < 0.3) {
      color = 0xff0000 // Red when low
    } else if (ratio < 0.6) {
      color = 0xffff00 // Yellow when medium
    } else {
      color = 0x00ff00 // Green when high
    }
    this.foregroundBar.fill(color)
    
    // Update text label if enabled
    if (this.textLabel) {
      this.textLabel.text = `Energy: ${Math.round(clampedEnergy)}/${Math.round(maxEnergy)}`
    }
  }

  /**
   * Destroys the energy bar and removes it from parent.
   */
  destroy(): void {
    if (this.container.parent) {
      this.container.parent.removeChild(this.container)
    }
    this.container.destroy({ children: true })
  }
}

