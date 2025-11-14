/**
 * Integration tests for Renderer system.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { Graphics, Container } from 'pixi.js'
import { App } from '../core/app'
import { Scene } from './scene'
import { StateManager, type GameState } from '../sim/state-manager'
import { Renderer } from './renderer'
import type { ShipSnapshot, PlanetSnapshot, PalletSnapshot } from '../net/protocol'

describe('Renderer', () => {
  let app: App
  let scene: Scene
  let stateManager: StateManager
  let renderer: Renderer
  let container: HTMLElement

  beforeEach(async () => {
    container = document.createElement('div')
    container.id = 'app'
    document.body.appendChild(container)

    app = new App()
    await app.init(container)
    scene = new Scene(app)
    stateManager = new StateManager()
    renderer = new Renderer(stateManager, scene)
  })

  afterEach(() => {
    if (renderer) {
      renderer.destroy()
    }
    if (scene) {
      scene.destroy()
    }
    if (app) {
      app.destroy()
    }
    if (container && container.parentNode) {
      container.parentNode.removeChild(container)
    }
  })

  describe('Initialization', () => {
    it('creates renderer with StateManager and Scene', () => {
      const newRenderer = new Renderer(stateManager, scene)

      expect(newRenderer).toBeDefined()
    })

    it('starts with no sprites', () => {
      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(0)
    })
  })

  describe('Sprite Creation', () => {
    it('creates ship sprite from state', () => {
      const gameState: GameState = {
        tick: 0,
        ship: { pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(1)
      expect(gameLayer.children[0]).toBeInstanceOf(Graphics)
    })

    it('creates planet sprites from state.planets array', () => {
      const gameState: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 400, y: 300 }, radius: 50 },
          { pos: { x: 600, y: 400 }, radius: 30 }
        ],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      // 1 ship + 2 planets = 3 sprites
      expect(gameLayer.children.length).toBe(3)
    })

    it('creates pallet sprites from state.pallets array', () => {
      const gameState: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true },
          { id: 2, pos: { x: 200, y: 200 }, active: true }
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      // 1 ship + 2 pallets = 3 sprites
      expect(gameLayer.children.length).toBe(3)
    })

    it('creates sprites for ship, planets, and pallets together', () => {
      const gameState: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 400, y: 300 }, radius: 50 }
        ],
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true }
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      // 1 ship + 1 planet + 1 pallet = 3 sprites
      expect(gameLayer.children.length).toBe(3)
    })

    it('adds sprites to scene game layer', () => {
      const gameState: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBeGreaterThan(0)
    })
  })

  describe('Sprite Updates', () => {
    it('updates ship sprite position from state', () => {
      const gameState1: GameState = {
        tick: 0,
        ship: { pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState1)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      const shipSprite = gameLayer.children[0] as Graphics

      const gameState2: GameState = {
        tick: 1,
        ship: { pos: { x: 300, y: 400 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState2)
      renderer.update()

      expect(shipSprite.x).toBe(300)
      expect(shipSprite.y).toBe(400)
    })

    it('updates ship sprite rotation from state', () => {
      const gameState1: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState1)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      const shipSprite = gameLayer.children[0] as Graphics

      const gameState2: GameState = {
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: Math.PI / 2, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState2)
      renderer.update()

      expect(shipSprite.rotation).toBeCloseTo(Math.PI / 2)
    })

    it('updates planet sprites from state.planets array', () => {
      const gameState1: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 100, y: 100 }, radius: 50 }
        ],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState1)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      const planetSprite = gameLayer.children[1] as Graphics // Index 1 (after ship)

      const gameState2: GameState = {
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 500, y: 600 }, radius: 50 }
        ],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState2)
      renderer.update()

      expect(planetSprite.x).toBe(500)
      expect(planetSprite.y).toBe(600)
    })

    it('updates pallet sprites from state.pallets array', () => {
      const gameState1: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true }
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState1)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      const palletSprite = gameLayer.children[1] as Graphics // Index 1 (after ship)

      const gameState2: GameState = {
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [
          { id: 1, pos: { x: 500, y: 600 }, active: true }
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState2)
      renderer.update()

      expect(palletSprite.x).toBe(500)
      expect(palletSprite.y).toBe(600)
    })

    it('updates pallet sprite visibility from active state', () => {
      const gameState1: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true }
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState1)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      const palletSprite = gameLayer.children[1] as Graphics

      const gameState2: GameState = {
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: false }
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState2)
      renderer.update()

      expect(palletSprite.visible).toBe(false)
    })
  })

  describe('Entity Management', () => {
    it('creates sprites for new planets when array grows', () => {
      const gameState1: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 400, y: 300 }, radius: 50 }
        ],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState1)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(2) // 1 ship + 1 planet

      const gameState2: GameState = {
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 400, y: 300 }, radius: 50 },
          { pos: { x: 600, y: 400 }, radius: 30 }
        ],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState2)
      renderer.update()

      expect(gameLayer.children.length).toBe(3) // 1 ship + 2 planets
    })

    it('creates sprites for new pallets when array grows', () => {
      const gameState1: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true }
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState1)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(2) // 1 ship + 1 pallet

      const gameState2: GameState = {
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true },
          { id: 2, pos: { x: 200, y: 200 }, active: true }
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState2)
      renderer.update()

      expect(gameLayer.children.length).toBe(3) // 1 ship + 2 pallets
    })

    it('removes sprites for deleted planets when array shrinks', () => {
      const gameState1: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 400, y: 300 }, radius: 50 },
          { pos: { x: 600, y: 400 }, radius: 30 }
        ],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState1)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(3) // 1 ship + 2 planets

      const gameState2: GameState = {
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 400, y: 300 }, radius: 50 }
        ],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState2)
      renderer.update()

      expect(gameLayer.children.length).toBe(2) // 1 ship + 1 planet
    })

    it('removes sprites for deleted pallets when array shrinks', () => {
      const gameState1: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true },
          { id: 2, pos: { x: 200, y: 200 }, active: true }
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState1)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(3) // 1 ship + 2 pallets

      const gameState2: GameState = {
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true }
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState2)
      renderer.update()

      expect(gameLayer.children.length).toBe(2) // 1 ship + 1 pallet
    })

    it('handles empty arrays (no planets/pallets)', () => {
      const gameState: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(1) // Only ship
    })
  })

  describe('Generic Array Pattern', () => {
    it('supports multiple planets (array-based iteration)', () => {
      const gameState: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 400, y: 300 }, radius: 50 },
          { pos: { x: 600, y: 400 }, radius: 30 },
          { pos: { x: 200, y: 500 }, radius: 40 }
        ],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(4) // 1 ship + 3 planets
    })

    it('supports multiple pallets (array-based iteration)', () => {
      const gameState: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true },
          { id: 2, pos: { x: 200, y: 200 }, active: true },
          { id: 3, pos: { x: 300, y: 300 }, active: true }
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(4) // 1 ship + 3 pallets
    })

    it('matches planets by index', () => {
      const gameState1: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 100, y: 100 }, radius: 50 },
          { pos: { x: 200, y: 200 }, radius: 30 }
        ],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState1)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      const planetSprite1 = gameLayer.children[1] as Graphics
      const planetSprite2 = gameLayer.children[2] as Graphics

      const gameState2: GameState = {
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 500, y: 500 }, radius: 50 }, // Same index 0, different position
          { pos: { x: 600, y: 600 }, radius: 30 }  // Same index 1, different position
        ],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState2)
      renderer.update()

      // Same sprites should be updated (matched by index)
      expect(gameLayer.children[1]).toBe(planetSprite1)
      expect(gameLayer.children[2]).toBe(planetSprite2)
      expect(planetSprite1.x).toBe(500)
      expect(planetSprite2.x).toBe(600)
    })

    it('matches pallets by id', () => {
      const gameState1: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true },
          { id: 2, pos: { x: 200, y: 200 }, active: true }
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState1)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      const palletSprite1 = gameLayer.children[1] as Graphics
      const palletSprite2 = gameLayer.children[2] as Graphics

      const gameState2: GameState = {
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [
          { id: 2, pos: { x: 600, y: 600 }, active: true }, // id 2, different position
          { id: 1, pos: { x: 500, y: 500 }, active: true }  // id 1, different position (order changed)
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState2)
      renderer.update()

      // Same sprites should be updated (matched by id, not index)
      expect(palletSprite1.x).toBe(500) // id 1
      expect(palletSprite2.x).toBe(600) // id 2
    })
  })

  describe('State Integration', () => {
    it('uses StateManager.getRenderState() to get interpolated state', () => {
      const gameState: GameState = {
        tick: 0,
        ship: { pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      const shipSprite = gameLayer.children[0] as Graphics
      expect(shipSprite.x).toBe(100)
      expect(shipSprite.y).toBe(200)
    })

    it('handles null/empty states gracefully', () => {
      // No state set, should not crash
      expect(() => renderer.update()).not.toThrow()
    })

    it('handles state with no interpolated state (uses authoritative fallback)', () => {
      const gameState: GameState = {
        tick: 0,
        ship: { pos: { x: 50, y: 50 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      }

      // Set authoritative but not interpolated
      stateManager.updateAuthoritative({
        t: 'snapshot',
        tick: 0,
        ship: gameState.ship,
        planets: gameState.planets,
        pallets: gameState.pallets,
        done: false,
        win: false
      })

      renderer.update()

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBeGreaterThan(0)
    })
  })

  describe('Clear and Destroy', () => {
    it('clear() removes all sprites from scene', () => {
      const gameState: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 400, y: 300 }, radius: 50 }
        ],
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true }
        ],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBeGreaterThan(0)

      renderer.clear()

      expect(gameLayer.children.length).toBe(0)
    })

    it('destroy() cleans up all resources', () => {
      const gameState: GameState = {
        tick: 0,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      }

      stateManager.updateInterpolated(gameState)
      renderer.update()

      expect(() => renderer.destroy()).not.toThrow()
    })
  })
})

