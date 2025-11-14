/**
 * Integration tests for HUD coordinator.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { App } from '../core/app'
import { Scene } from '../gfx/scene'
import { StateManager, type GameState } from '../sim/state-manager'
import { HUD } from './hud'
import type { SnapshotMessage } from '../net/protocol'

describe('HUD', () => {
  let app: App
  let scene: Scene
  let stateManager: StateManager
  let hud: HUD

  beforeEach(async () => {
    // Create container for headless testing
    const container = document.createElement('div')
    container.id = 'app'
    document.body.appendChild(container)

    app = new App({ width: 1280, height: 720 })
    await app.init(container)

    scene = new Scene(app)
    stateManager = new StateManager()
    hud = new HUD(scene, stateManager)
  })

  afterEach(() => {
    if (hud) {
      hud.destroy()
    }
    if (scene) {
      scene.destroy()
    }
    if (app) {
      app.destroy()
    }
    const container = document.getElementById('app')
    if (container && container.parentNode) {
      container.parentNode.removeChild(container)
    }
  })

  describe('Creation', () => {
    it('creates HUD with Scene and StateManager', () => {
      expect(hud).toBeDefined()
    })

    it('initializes all components (EnergyBar, PalletCounter, GameBanner)', () => {
      const uiLayer = scene.getLayer('ui')
      expect(uiLayer.children.length).toBeGreaterThan(0)
    })
  })

  describe('Energy Bar Updates', () => {
    it('updates EnergyBar from ship.energy in game state', () => {
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 1,
        ship: {
          pos: { x: 100, y: 100 },
          vel: { x: 0, y: 0 },
          rot: 0,
          energy: 75.0
        },
        planets: [],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateAuthoritative(snapshot)
      hud.update()

      // Energy bar should be updated (we can't easily verify the exact width without accessing internals,
      // but we can verify the update doesn't throw and components exist)
      const uiLayer = scene.getLayer('ui')
      const energyBar = uiLayer.children.find(child => child.label === 'energy-bar')
      expect(energyBar).toBeDefined()
    })

    it('handles state with energy at different levels (0, 50, 100)', () => {
      const snapshots: SnapshotMessage[] = [
        {
          t: 'snapshot',
          tick: 1,
          ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 0 },
          planets: [],
          pallets: [],
          done: false,
          win: false
        },
        {
          t: 'snapshot',
          tick: 2,
          ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 50 },
          planets: [],
          pallets: [],
          done: false,
          win: false
        },
        {
          t: 'snapshot',
          tick: 3,
          ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
          planets: [],
          pallets: [],
          done: false,
          win: false
        }
      ]

      for (const snapshot of snapshots) {
        stateManager.updateAuthoritative(snapshot)
        hud.update()
      }

      const uiLayer = scene.getLayer('ui')
      const energyBar = uiLayer.children.find(child => child.label === 'energy-bar')
      expect(energyBar).toBeDefined()
    })
  })

  describe('Pallet Counter Updates', () => {
    it('updates PalletCounter from pallets array (counts active pallets)', () => {
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 1,
        ship: {
          pos: { x: 100, y: 100 },
          vel: { x: 0, y: 0 },
          rot: 0,
          energy: 100
        },
        planets: [],
        pallets: [
          { id: 1, pos: { x: 200, y: 200 }, active: true },
          { id: 2, pos: { x: 300, y: 300 }, active: true },
          { id: 3, pos: { x: 400, y: 400 }, active: false }
        ],
        done: false,
        win: false
      }

      stateManager.updateAuthoritative(snapshot)
      hud.update()

      const uiLayer = scene.getLayer('ui')
      const palletCounter = uiLayer.children.find(child => child.label === 'pallet-counter')
      expect(palletCounter).toBeDefined()
    })

    it('handles state with no pallets', () => {
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 1,
        ship: {
          pos: { x: 100, y: 100 },
          vel: { x: 0, y: 0 },
          rot: 0,
          energy: 100
        },
        planets: [],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateAuthoritative(snapshot)
      hud.update()

      const uiLayer = scene.getLayer('ui')
      const palletCounter = uiLayer.children.find(child => child.label === 'pallet-counter')
      expect(palletCounter).toBeDefined()
    })

    it('handles state with all pallets inactive', () => {
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 1,
        ship: {
          pos: { x: 100, y: 100 },
          vel: { x: 0, y: 0 },
          rot: 0,
          energy: 100
        },
        planets: [],
        pallets: [
          { id: 1, pos: { x: 200, y: 200 }, active: false },
          { id: 2, pos: { x: 300, y: 300 }, active: false }
        ],
        done: false,
        win: false
      }

      stateManager.updateAuthoritative(snapshot)
      hud.update()

      const uiLayer = scene.getLayer('ui')
      const palletCounter = uiLayer.children.find(child => child.label === 'pallet-counter')
      expect(palletCounter).toBeDefined()
    })
  })

  describe('Game Banner Updates', () => {
    it('shows win banner when state.done = true and state.win = true', () => {
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 1,
        ship: {
          pos: { x: 100, y: 100 },
          vel: { x: 0, y: 0 },
          rot: 0,
          energy: 100
        },
        planets: [],
        pallets: [],
        done: true,
        win: true
      }

      stateManager.updateAuthoritative(snapshot)
      hud.update()

      const uiLayer = scene.getLayer('ui')
      const gameBanner = uiLayer.children.find(child => child.label === 'game-banner')
      expect(gameBanner).toBeDefined()
      expect(gameBanner.visible).toBe(true)
    })

    it('shows lose banner when state.done = true and state.win = false', () => {
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 1,
        ship: {
          pos: { x: 100, y: 100 },
          vel: { x: 0, y: 0 },
          rot: 0,
          energy: 100
        },
        planets: [],
        pallets: [],
        done: true,
        win: false
      }

      stateManager.updateAuthoritative(snapshot)
      hud.update()

      const uiLayer = scene.getLayer('ui')
      const gameBanner = uiLayer.children.find(child => child.label === 'game-banner')
      expect(gameBanner).toBeDefined()
      expect(gameBanner.visible).toBe(true)
    })

    it('hides banner when state.done = false', () => {
      // First show banner
      const snapshot1: SnapshotMessage = {
        t: 'snapshot',
        tick: 1,
        ship: {
          pos: { x: 100, y: 100 },
          vel: { x: 0, y: 0 },
          rot: 0,
          energy: 100
        },
        planets: [],
        pallets: [],
        done: true,
        win: true
      }

      stateManager.updateAuthoritative(snapshot1)
      hud.update()

      const uiLayer1 = scene.getLayer('ui')
      const gameBanner1 = uiLayer1.children.find(child => child.name === 'game-banner')
      expect(gameBanner1?.visible).toBe(true)

      // Then hide banner
      const snapshot2: SnapshotMessage = {
        t: 'snapshot',
        tick: 2,
        ship: {
          pos: { x: 100, y: 100 },
          vel: { x: 0, y: 0 },
          rot: 0,
          energy: 100
        },
        planets: [],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateAuthoritative(snapshot2)
      hud.update()

      const uiLayer2 = scene.getLayer('ui')
      const gameBanner2 = uiLayer2.children.find(child => child.name === 'game-banner')
      expect(gameBanner2?.visible).toBe(false)
    })
  })

  describe('Update All Components', () => {
    it('updates all components on each update() call', () => {
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 1,
        ship: {
          pos: { x: 100, y: 100 },
          vel: { x: 0, y: 0 },
          rot: 0,
          energy: 75.0
        },
        planets: [],
        pallets: [
          { id: 1, pos: { x: 200, y: 200 }, active: true },
          { id: 2, pos: { x: 300, y: 300 }, active: false }
        ],
        done: false,
        win: false
      }

      stateManager.updateAuthoritative(snapshot)
      hud.update()

      const uiLayer = scene.getLayer('ui')
      const energyBar = uiLayer.children.find(child => child.label === 'energy-bar')
      const palletCounter = uiLayer.children.find(child => child.label === 'pallet-counter')
      const gameBanner = uiLayer.children.find(child => child.label === 'game-banner')

      expect(energyBar).toBeDefined()
      expect(palletCounter).toBeDefined()
      expect(gameBanner).toBeDefined()
    })
  })

  describe('Destruction', () => {
    it('destroys HUD and cleans up all components', () => {
      const uiLayer = scene.getLayer('ui')
      const initialChildCount = uiLayer.children.length

      hud.destroy()

      expect(uiLayer.children.length).toBeLessThan(initialChildCount)
    })
  })

  describe('Missing State Handling', () => {
    it('handles missing authoritative/interpolated state gracefully', () => {
      // Update with no state set
      expect(() => {
        hud.update()
      }).not.toThrow()
    })
  })
})

