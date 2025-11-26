/**
 * HUD (Heads-Up Display) coordinator for managing all UI components.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { Scene } from '../gfx/scene'
import { StateManager } from '../sim/state-manager'
import { EnergyBar } from './components/energy-bar'
import { PalletCounter } from './components/pallet-counter'
import { GameBanner } from './components/game-banner'
import { PlayerIndicator } from './components/player-indicator'

/**
 * HUD coordinator that manages all UI components and updates them from game state.
 */
export class HUD {
  private scene: Scene
  private stateManager: StateManager
  private energyBar: EnergyBar
  private palletCounter: PalletCounter
  private gameBanner: GameBanner
  private playerIndicator: PlayerIndicator
  private maxEnergy: number = 100.0

  constructor(scene: Scene, stateManager: StateManager) {
    this.scene = scene
    this.stateManager = stateManager

    // Get HUD layer from scene (separate from UI layer for menu components)
    const hudLayer = this.scene.getLayer('hud')

    // Initialize components - positioned at top-left
    this.energyBar = new EnergyBar(hudLayer, {
      x: 20,
      y: 10,
      width: 200,
      height: 20
    })

    this.palletCounter = new PalletCounter(hudLayer, {
      x: 20,
      y: 40
    })

    this.playerIndicator = new PlayerIndicator(hudLayer, {
      x: 20,
      y: 70
    })

    this.gameBanner = new GameBanner(hudLayer)
  }

  /**
   * Shows the HUD (makes all components visible).
   */
  show(): void {
    this.energyBar.getContainer().visible = true
    this.palletCounter.getContainer().visible = true
    this.playerIndicator.getContainer().visible = true
    this.gameBanner.getContainer().visible = true
  }

  /**
   * Hides the HUD (makes all components invisible).
   */
  hide(): void {
    this.energyBar.getContainer().visible = false
    this.palletCounter.getContainer().visible = false
    this.playerIndicator.getContainer().visible = false
    this.gameBanner.getContainer().visible = false
  }

  /**
   * Updates all HUD components from current game state.
   * Should be called each frame or when state changes.
   * v1 multiplayer format: uses ships array with myShipId.
   */
  update(): void {
    const state = this.stateManager.getRenderState()

    // Get player's ship from ships array (v1 multiplayer format)
    const playerShip = state.ships.find(ship => ship.id === state.myShipId)
    const energy = playerShip?.energy ?? 0

    // Update energy bar
    this.energyBar.update(energy, this.maxEnergy)

    // Update pallet counter (count active pallets)
    const activePallets = state.pallets.filter(p => p.active).length
    const totalPallets = state.pallets.length
    this.palletCounter.update(activePallets, totalPallets)

    // Update player indicator with player's ship ID
    this.playerIndicator.update(state.myShipId)

    // Update game banner based on done/win flags
    if (state.done) {
      if (state.win) {
        this.gameBanner.showWin()
      } else {
        this.gameBanner.showLose()
      }
    } else {
      this.gameBanner.hide()
    }

    // Update banner size if needed (for responsive layout)
    const app = this.scene.getRoot().parent
    if (app && 'screen' in app) {
      const screen = (app as { screen?: { width?: number; height?: number } }).screen
      if (screen && screen.width && screen.height) {
        this.gameBanner.updateSize(screen.width, screen.height)
      }
    }
  }

  /**
   * Destroys the HUD and cleans up all components.
   */
  destroy(): void {
    this.energyBar.destroy()
    this.palletCounter.destroy()
    this.playerIndicator.destroy()
    this.gameBanner.destroy()
  }
}

