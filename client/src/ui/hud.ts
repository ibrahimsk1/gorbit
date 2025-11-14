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

/**
 * HUD coordinator that manages all UI components and updates them from game state.
 */
export class HUD {
  private scene: Scene
  private stateManager: StateManager
  private energyBar: EnergyBar
  private palletCounter: PalletCounter
  private gameBanner: GameBanner
  private maxEnergy: number = 100.0

  constructor(scene: Scene, stateManager: StateManager) {
    this.scene = scene
    this.stateManager = stateManager

    // Get UI layer from scene
    const uiLayer = this.scene.getLayer('ui')

    // Initialize components
    this.energyBar = new EnergyBar(uiLayer, {
      x: 20,
      y: 20,
      width: 200,
      height: 20
    })

    this.palletCounter = new PalletCounter(uiLayer, {
      x: 20,
      y: 50
    })

    this.gameBanner = new GameBanner(uiLayer)
  }

  /**
   * Updates all HUD components from current game state.
   * Should be called each frame or when state changes.
   */
  update(): void {
    const state = this.stateManager.getRenderState()

    // Update energy bar
    const energy = state.ship.energy
    this.energyBar.update(energy, this.maxEnergy)

    // Update pallet counter (count active pallets)
    const activePallets = state.pallets.filter(p => p.active).length
    const totalPallets = state.pallets.length
    this.palletCounter.update(activePallets, totalPallets)

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
      const screen = (app as any).screen
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
    this.gameBanner.destroy()
  }
}

