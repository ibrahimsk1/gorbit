/**
 * Pallet counter component for displaying active pallet count.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { Container, Text } from 'pixi.js'

export interface PalletCounterConfig {
  x?: number
  y?: number
  textStyle?: {
    fontFamily?: string
    fontSize?: number
    fill?: number
  }
}

/**
 * Pallet counter component that displays the count of active (collectible) pallets.
 */
export class PalletCounter {
  private container: Container
  private textLabel: Text
  private config: Required<PalletCounterConfig>

  constructor(parent: Container, config: PalletCounterConfig = {}) {
    this.container = new Container()
    this.container.label = 'pallet-counter'
    parent.addChild(this.container)

    // Default configuration
    this.config = {
      x: config.x ?? 20,
      y: config.y ?? 50,
      textStyle: {
        fontFamily: config.textStyle?.fontFamily ?? 'Arial',
        fontSize: config.textStyle?.fontSize ?? 14,
        fill: config.textStyle?.fill ?? 0xffffff
      }
    }

    // Set container position
    this.container.x = this.config.x
    this.container.y = this.config.y

    // Create text label
    this.textLabel = new Text({
      text: 'Pallets: 0',
      style: this.config.textStyle
    })
    this.textLabel.label = 'pallet-counter-text'
    this.container.addChild(this.textLabel)
  }

  /**
   * Updates the pallet counter based on active pallet count.
   * 
   * @param activeCount Number of active (collectible) pallets
   * @param totalCount Total number of pallets (optional, for display format)
   */
  update(activeCount: number, totalCount: number = 0): void {
    // Ensure counts are non-negative
    const active = Math.max(0, activeCount)
    const total = Math.max(0, totalCount)

    // Update text display
    if (total > 0) {
      this.textLabel.text = `Pallets: ${active}/${total}`
    } else {
      this.textLabel.text = `Pallets: ${active}`
    }
  }

  /**
   * Destroys the pallet counter and removes it from parent.
   */
  destroy(): void {
    if (this.container.parent) {
      this.container.parent.removeChild(this.container)
    }
    this.container.destroy({ children: true })
  }
}

