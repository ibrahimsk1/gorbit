/**
 * Game banner component for displaying win/lose messages.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { Container, Graphics, Text } from 'pixi.js'

export interface GameBannerConfig {
  winMessage?: string
  loseMessage?: string
  backgroundColor?: number
  textStyle?: {
    fontFamily?: string
    fontSize?: number
    fill?: number
  }
}

/**
 * Game banner component that displays win/lose messages when game is finished.
 */
export class GameBanner {
  private container: Container
  private backgroundPanel: Graphics
  private textLabel: Text
  private config: Required<GameBannerConfig>

  constructor(parent: Container, config: GameBannerConfig = {}) {
    this.container = new Container()
    this.container.label = 'game-banner'
    parent.addChild(this.container)

    // Default configuration
    this.config = {
      winMessage: config.winMessage ?? 'You Win!',
      loseMessage: config.loseMessage ?? 'Game Over!',
      backgroundColor: config.backgroundColor ?? 0x000000,
      textStyle: {
        fontFamily: config.textStyle?.fontFamily ?? 'Arial',
        fontSize: config.textStyle?.fontSize ?? 48,
        fill: config.textStyle?.fill ?? 0xffffff
      }
    }

    // Initially hidden
    this.container.visible = false

    // Create background panel (full screen overlay)
    this.backgroundPanel = new Graphics()
    this.backgroundPanel.label = 'game-banner-background'
    this.container.addChild(this.backgroundPanel)

    // Create text label
    this.textLabel = new Text({
      text: '',
      style: this.config.textStyle
    })
    this.textLabel.label = 'game-banner-text'
    this.textLabel.anchor.set(0.5) // Center anchor
    this.container.addChild(this.textLabel)
  }

  /**
   * Updates banner size and position based on screen dimensions.
   * Should be called when screen size changes.
   * 
   * @param width Screen width
   * @param height Screen height
   */
  updateSize(width: number, height: number): void {
    // Update background panel to cover full screen
    this.backgroundPanel.clear()
    this.backgroundPanel.rect(0, 0, width, height)
    this.backgroundPanel.fill({ color: this.config.backgroundColor, alpha: 0.7 })

    // Center text in screen
    this.textLabel.x = width / 2
    this.textLabel.y = height / 2
  }

  /**
   * Shows the win message banner.
   */
  showWin(): void {
    this.textLabel.text = this.config.winMessage
    this.textLabel.style.fill = 0x00ff00 // Green for win
    this.container.visible = true
    
    // Update size if parent has dimensions
    const parent = this.container.parent
    if (parent && 'width' in parent && 'height' in parent) {
      this.updateSize(parent.width as number, parent.height as number)
    }
  }

  /**
   * Shows the lose message banner.
   */
  showLose(): void {
    this.textLabel.text = this.config.loseMessage
    this.textLabel.style.fill = 0xff0000 // Red for lose
    this.container.visible = true
    
    // Update size if parent has dimensions
    const parent = this.container.parent
    if (parent && 'width' in parent && 'height' in parent) {
      this.updateSize(parent.width as number, parent.height as number)
    }
  }

  /**
   * Hides the banner.
   */
  hide(): void {
    this.container.visible = false
  }

  /**
   * Destroys the game banner and removes it from parent.
   */
  destroy(): void {
    if (this.container.parent) {
      this.container.parent.removeChild(this.container)
    }
    this.container.destroy({ children: true })
  }
}

