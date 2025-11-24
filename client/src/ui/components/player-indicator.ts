/**
 * Player indicator component for displaying player ID and name.
 * 
 * Labels: scope:integration loop:g7-ui layer:ui dep:state,net b:player-indicator r:medium
 */

import { Container, Text } from 'pixi.js'

export interface PlayerIndicatorConfig {
  x?: number
  y?: number
  textStyle?: {
    fontFamily?: string
    fontSize?: number
    fill?: number
  }
}

/**
 * Player indicator component that displays player ID and optional name.
 * Styled as small text indicator positioned in HUD area.
 */
export class PlayerIndicator {
  private container: Container
  private textLabel: Text
  private config: Required<PlayerIndicatorConfig>
  private currentPlayerId: number | null = null
  private currentPlayerName: string | null = null

  constructor(parent: Container, config: PlayerIndicatorConfig = {}) {
    this.container = new Container()
    this.container.label = 'player-indicator'
    parent.addChild(this.container)

    // Default configuration
    this.config = {
      x: config.x ?? 20,
      y: config.y ?? 80,
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
      text: '',
      style: this.config.textStyle
    })
    this.textLabel.label = 'player-indicator-text'
    this.container.addChild(this.textLabel)
  }

  /**
   * Updates the displayed player ID and optional name.
   * 
   * @param playerId Player ID to display
   * @param playerName Optional player name to display
   */
  update(playerId: number, playerName?: string): void {
    this.currentPlayerId = playerId
    this.currentPlayerName = playerName ?? null

    // Update text display
    if (playerName) {
      this.textLabel.text = `${playerName} (ID: ${playerId})`
    } else {
      this.textLabel.text = `Player ${playerId}`
    }
  }

  /**
   * Destroys the player indicator and removes it from parent.
   */
  destroy(): void {
    if (this.container.parent) {
      this.container.parent.removeChild(this.container)
    }
    this.container.destroy({ children: true })
  }
}

