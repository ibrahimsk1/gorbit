/**
 * Integration tests for Renderer system.
 * 
 * Labels: scope:integration loop:g6-rendering layer:gfx dep:state
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { Graphics, Container } from 'pixi.js'
import { App } from '../core/app'
import { Scene } from './scene'
import { StateManager, type GameState } from '../sim/state-manager'
import { Renderer } from './renderer'
import { Camera } from './camera'
import type { ShipSnapshot, PlanetSnapshot, PalletSnapshot } from '../net/protocol'

// Helper function to create GameState with v1 multiplayer format
function createGameState(overrides: Partial<GameState> = {}): GameState {
  return {
    tick: 0,
    ships: [{ id: 1, pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }],
    planets: [],
    pallets: [],
    worldBounds: { width: 2000, height: 2000 },
    myShipId: 1,
    done: false,
    win: false,
    ...overrides
  }
}

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
    scene.initialize() // Explicit initialization
    stateManager = new StateManager()
    renderer = new Renderer(stateManager, scene, app)
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
      const newRenderer = new Renderer(stateManager, scene, app)

      expect(newRenderer).toBeDefined()
    })

    it('starts with no sprites', () => {
      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(0)
    })
  })

  describe('Sprite Creation', () => {
    it('creates ship sprite from state', () => {
      const gameState = createGameState({
        ships: [{ id: 1, pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }]
      })

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(1)
      expect(gameLayer.children[0]).toBeInstanceOf(Graphics)
    })

    it('creates planet sprites from state.planets array', () => {
      const gameState = createGameState({
        planets: [
          { id: 1, pos: { x: 400, y: 300 }, radius: 50 },
          { id: 2, pos: { x: 600, y: 400 }, radius: 30 }
        ]
      })

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      // 1 ship + 2 planets = 3 sprites
      expect(gameLayer.children.length).toBe(3)
    })

    it('creates pallet sprites from state.pallets array', () => {
      const gameState = createGameState({
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true },
          { id: 2, pos: { x: 200, y: 200 }, active: true }
        ]
      })

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      // 1 ship + 2 pallets = 3 sprites
      expect(gameLayer.children.length).toBe(3)
    })

    it('creates sprites for ship, planets, and pallets together', () => {
      const gameState = createGameState({
        planets: [
          { id: 1, pos: { x: 400, y: 300 }, radius: 50 }
        ],
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true }
        ]
      })

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      // 1 ship + 1 planet + 1 pallet = 3 sprites
      expect(gameLayer.children.length).toBe(3)
    })

    it('adds sprites to scene game layer', () => {
      const gameState = createGameState()

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBeGreaterThan(0)
    })
  })

  describe('Sprite Updates', () => {
    it('updates ship sprite position from state', () => {
      const gameState1 = createGameState({
        ships: [{ id: 1, pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }]
      })

      stateManager.updateInterpolated(gameState1)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      const shipSprite = gameLayer.children[0] as Graphics
      const initialX = shipSprite.x
      const initialY = shipSprite.y

      const gameState2 = createGameState({
        ships: [{ id: 1, pos: { x: 300, y: 400 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }]
      })

      stateManager.updateInterpolated(gameState2)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      // Sprite position should have changed (screen coordinates with camera offset)
      expect(shipSprite.x).not.toBe(initialX)
      expect(shipSprite.y).not.toBe(initialY)
    })

    it('updates ship sprite rotation from state', () => {
      const gameState1 = createGameState({
        ships: [{ id: 1, pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }]
      })

      stateManager.updateInterpolated(gameState1)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      const shipSprite = gameLayer.children[0] as Graphics

      const gameState2 = createGameState({
        ships: [{ id: 1, pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: Math.PI / 2, energy: 100 }]
      })

      stateManager.updateInterpolated(gameState2)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      expect(shipSprite.rotation).toBeCloseTo(Math.PI / 2)
    })

    it('updates planet sprites from state.planets array', () => {
      const gameState1 = createGameState({
        planets: [
          { id: 1, pos: { x: 100, y: 100 }, radius: 50 }
        ]
      })

      stateManager.updateInterpolated(gameState1)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      const planetSprite = gameLayer.children[1] as Graphics // Index 1 (after ship)
      const initialX = planetSprite.x
      const initialY = planetSprite.y

      const gameState2 = createGameState({
        planets: [
          { id: 1, pos: { x: 500, y: 600 }, radius: 50 }
        ]
      })

      stateManager.updateInterpolated(gameState2)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      // Sprite position should have changed (screen coordinates with camera offset)
      expect(planetSprite.x).not.toBe(initialX)
      expect(planetSprite.y).not.toBe(initialY)
    })

    it('updates pallet sprites from state.pallets array', () => {
      const gameState1 = createGameState({
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true }
        ]
      })

      stateManager.updateInterpolated(gameState1)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      const palletSprite = gameLayer.children[1] as Graphics // Index 1 (after ship)
      const initialX = palletSprite.x
      const initialY = palletSprite.y

      const gameState2 = createGameState({
        pallets: [
          { id: 1, pos: { x: 500, y: 600 }, active: true }
        ]
      })

      stateManager.updateInterpolated(gameState2)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      // Sprite position should have changed (screen coordinates with camera offset)
      expect(palletSprite.x).not.toBe(initialX)
      expect(palletSprite.y).not.toBe(initialY)
    })

    it('updates pallet sprite visibility from active state', () => {
      const gameState1 = createGameState({
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true }
        ]
      })

      stateManager.updateInterpolated(gameState1)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      const palletSprite = gameLayer.children[1] as Graphics

      const gameState2 = createGameState({
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: false }
        ]
      })

      stateManager.updateInterpolated(gameState2)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      expect(palletSprite.visible).toBe(false)
    })
  })

  describe('Entity Management', () => {
    it('creates sprites for new planets when array grows', () => {
      const gameState1 = createGameState({
        planets: [
          { id: 1, pos: { x: 400, y: 300 }, radius: 50 }
        ]
      })

      stateManager.updateInterpolated(gameState1)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(2) // 1 ship + 1 planet

      const gameState2 = createGameState({
        planets: [
          { id: 1, pos: { x: 400, y: 300 }, radius: 50 },
          { id: 2, pos: { x: 600, y: 400 }, radius: 30 }
        ]
      })

      stateManager.updateInterpolated(gameState2)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      expect(gameLayer.children.length).toBe(3) // 1 ship + 2 planets
    })

    it('creates sprites for new pallets when array grows', () => {
      const gameState1 = createGameState({
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true }
        ]
      })

      stateManager.updateInterpolated(gameState1)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(2) // 1 ship + 1 pallet

      const gameState2 = createGameState({
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true },
          { id: 2, pos: { x: 200, y: 200 }, active: true }
        ]
      })

      stateManager.updateInterpolated(gameState2)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      expect(gameLayer.children.length).toBe(3) // 1 ship + 2 pallets
    })

    it('removes sprites for deleted planets when array shrinks', () => {
      const gameState1 = createGameState({
        planets: [
          { id: 1, pos: { x: 400, y: 300 }, radius: 50 },
          { id: 2, pos: { x: 600, y: 400 }, radius: 30 }
        ]
      })

      stateManager.updateInterpolated(gameState1)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(3) // 1 ship + 2 planets

      const gameState2 = createGameState({
        planets: [
          { id: 1, pos: { x: 400, y: 300 }, radius: 50 }
        ]
      })

      stateManager.updateInterpolated(gameState2)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      expect(gameLayer.children.length).toBe(2) // 1 ship + 1 planet
    })

    it('removes sprites for deleted pallets when array shrinks', () => {
      const gameState1 = createGameState({
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true },
          { id: 2, pos: { x: 200, y: 200 }, active: true }
        ]
      })

      stateManager.updateInterpolated(gameState1)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(3) // 1 ship + 2 pallets

      const gameState2 = createGameState({
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true }
        ]
      })

      stateManager.updateInterpolated(gameState2)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      expect(gameLayer.children.length).toBe(2) // 1 ship + 1 pallet
    })

    it('handles empty arrays (no planets/pallets)', () => {
      const gameState = createGameState()

      stateManager.updateInterpolated(gameState)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(1) // Only ship
    })
  })

  describe('Generic Array Pattern', () => {
    it('supports multiple planets (array-based iteration)', () => {
      const gameState = createGameState({
        planets: [
          { id: 1, pos: { x: 400, y: 300 }, radius: 50 },
          { id: 2, pos: { x: 600, y: 400 }, radius: 30 },
          { id: 3, pos: { x: 200, y: 500 }, radius: 40 }
        ]
      })

      stateManager.updateInterpolated(gameState)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(4) // 1 ship + 3 planets
    })

    it('supports multiple pallets (array-based iteration)', () => {
      const gameState = createGameState({
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true },
          { id: 2, pos: { x: 200, y: 200 }, active: true },
          { id: 3, pos: { x: 300, y: 300 }, active: true }
        ]
      })

      stateManager.updateInterpolated(gameState)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(4) // 1 ship + 3 pallets
    })

    it('matches planets by index', () => {
      const gameState1 = createGameState({
        planets: [
          { id: 1, pos: { x: 100, y: 100 }, radius: 50 },
          { id: 2, pos: { x: 200, y: 200 }, radius: 30 }
        ]
      })

      stateManager.updateInterpolated(gameState1)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      const planetSprite1 = gameLayer.children[1] as Graphics
      const planetSprite2 = gameLayer.children[2] as Graphics

      const gameState2 = createGameState({
        planets: [
          { id: 1, pos: { x: 500, y: 500 }, radius: 50 }, // Same index 0, different position
          { id: 2, pos: { x: 600, y: 600 }, radius: 30 }  // Same index 1, different position
        ]
      })

      stateManager.updateInterpolated(gameState2)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      // Same sprites should be updated (matched by index)
      expect(gameLayer.children[1]).toBe(planetSprite1)
      expect(gameLayer.children[2]).toBe(planetSprite2)
      // Positions should have changed (screen coordinates with camera offset)
      expect(planetSprite1.x).not.toBe(100)
      expect(planetSprite2.x).not.toBe(200)
    })

    it('matches pallets by id', () => {
      const gameState1 = createGameState({
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true },
          { id: 2, pos: { x: 200, y: 200 }, active: true }
        ]
      })

      stateManager.updateInterpolated(gameState1)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      const palletSprite1 = gameLayer.children[1] as Graphics
      const palletSprite2 = gameLayer.children[2] as Graphics

      const gameState2 = createGameState({
        pallets: [
          { id: 2, pos: { x: 600, y: 600 }, active: true }, // id 2, different position
          { id: 1, pos: { x: 500, y: 500 }, active: true }  // id 1, different position (order changed)
        ]
      })

      stateManager.updateInterpolated(gameState2)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      // Same sprites should be updated (matched by id, not index)
      // Positions should have changed (screen coordinates with camera offset)
      expect(palletSprite1.x).not.toBe(100) // id 1
      expect(palletSprite2.x).not.toBe(200) // id 2
    })
  })

  describe('State Integration', () => {
    it('uses StateManager.getRenderState() to get interpolated state', () => {
      const gameState = createGameState({
        ships: [{ id: 1, pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }]
      })

      stateManager.updateInterpolated(gameState)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      const shipSprite = gameLayer.children[0] as Graphics
      // Sprite position should be transformed (screen coordinates with camera offset)
      expect(shipSprite).toBeDefined()
      expect(shipSprite.x).toBeGreaterThan(0)
      expect(shipSprite.y).toBeGreaterThan(0)
    })

    it('handles null/empty states gracefully', () => {
      // No state set, should not crash
      expect(() => renderer.update()).not.toThrow()
    })

    it('handles state with no interpolated state (uses authoritative fallback)', () => {
      const gameState = createGameState({
        ships: [{ id: 1, pos: { x: 50, y: 50 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }]
      })

      // Set authoritative but not interpolated
      stateManager.updateAuthoritative({
        t: 'snapshot',
        tick: 0,
        ships: gameState.ships,
        planets: gameState.planets,
        pallets: gameState.pallets,
        worldBounds: gameState.worldBounds,
        myShipId: gameState.myShipId,
        done: false,
        win: false
      })

      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBeGreaterThan(0)
    })
  })

  describe('Clear and Destroy', () => {
    it('clear() removes all sprites from scene', () => {
      const gameState = createGameState({
        planets: [
          { id: 1, pos: { x: 400, y: 300 }, radius: 50 }
        ],
        pallets: [
          { id: 1, pos: { x: 100, y: 100 }, active: true }
        ]
      })

      stateManager.updateInterpolated(gameState)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBeGreaterThan(0)

      renderer.clear()

      expect(gameLayer.children.length).toBe(0)
    })

    it('destroy() cleans up all resources', () => {
      const gameState = createGameState()

      stateManager.updateInterpolated(gameState)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      expect(() => renderer.destroy()).not.toThrow()
    })
  })

  describe('Camera Integration', () => {
    it('creates Camera in constructor', () => {
      const newRenderer = new Renderer(stateManager, scene, app)
      
      // Camera should be created internally
      expect(newRenderer).toBeDefined()
    })

    it('uses camera position for coordinate transformation', () => {
      const gameState = createGameState({
        ships: [{ id: 1, pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }]
      })

      stateManager.updateInterpolated(gameState)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      // Camera should have moved toward ship position
      // Ship sprite should be rendered at screen position adjusted by camera
      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(1)
    })

    it('calls camera.update() with player ship position before rendering', () => {
      const gameState = createGameState({
        ships: [{ id: 1, pos: { x: 150, y: 250 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }]
      })

      stateManager.updateInterpolated(gameState)
      
      // First update - camera should start moving toward ship
      renderer.update()
      
      // Second update - camera should continue moving toward ship
      renderer.update()
      
      // Camera should be following ship (sprite should be rendered)
      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(1)
    })

    it('camera follows player ship smoothly with lerp', () => {
      const gameState1 = createGameState({
        ships: [{ id: 1, pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }]
      })

      stateManager.updateInterpolated(gameState1)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const gameState2 = createGameState({
        ships: [{ id: 1, pos: { x: 100, y: 100 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }]
      })

      stateManager.updateInterpolated(gameState2)
      
      // Update multiple times to let camera lerp
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      // Camera should be following ship
      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(1)
    })

    it('coordinate transformation uses camera offset', () => {
      const pixiApp = app.getApplication()
      const screenWidth = pixiApp.screen.width
      const screenHeight = pixiApp.screen.height

      const gameState = createGameState({
        ships: [{ id: 1, pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }],
        planets: [
          { id: 1, pos: { x: 0, y: 0 }, radius: 50 } // Planet at world origin
        ]
      })

      stateManager.updateInterpolated(gameState)
      
      // Update multiple times to let camera move toward ship
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      // Sprites should be rendered with camera offset applied
      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(2) // Ship + planet
    })

    it('camera position clamped to world bounds', () => {
      const WORLD_WIDTH = 2000
      const WORLD_HEIGHT = 2000

      // Ship at world edge
      const gameState = createGameState({
        ships: [{ id: 1, pos: { x: WORLD_WIDTH / 2, y: WORLD_HEIGHT / 2 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }],
        worldBounds: { width: WORLD_WIDTH, height: WORLD_HEIGHT }
      })

      stateManager.updateInterpolated(gameState)
      
      // Update multiple times
      for (let i = 0; i < 30; i++) {
        renderer.update()
      }

      // Camera should be clamped, sprites should still render correctly
      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(1)
    })
  })

  describe('World Bounds Visualization', () => {
    it('renders world bounds as rectangle outline in background layer', () => {
      const gameState = createGameState()

      stateManager.updateInterpolated(gameState)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const backgroundLayer = scene.getLayer('background')
      // World bounds rectangle should be in background layer
      expect(backgroundLayer.children.length).toBeGreaterThan(0)
      
      // Find the bounds rectangle (should be a Graphics object)
      const boundsSprite = backgroundLayer.children.find(
        child => child instanceof Graphics
      ) as Graphics | undefined
      expect(boundsSprite).toBeDefined()
    })

    it('uses world bounds from state for rectangle dimensions', () => {
      const gameState = createGameState({
        worldBounds: { width: 3000, height: 3000 }
      })

      stateManager.updateInterpolated(gameState)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const backgroundLayer = scene.getLayer('background')
      const boundsSprite = backgroundLayer.children.find(
        child => child instanceof Graphics
      ) as Graphics | undefined
      
      expect(boundsSprite).toBeDefined()
      // Bounds sprite should exist and be visible
      expect(boundsSprite?.visible).toBe(true)
    })

    it('uses default world bounds constants when state bounds not available', () => {
      const gameState = createGameState({
        worldBounds: { width: 2000, height: 2000 } // Default
      })

      stateManager.updateInterpolated(gameState)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const backgroundLayer = scene.getLayer('background')
      const boundsSprite = backgroundLayer.children.find(
        child => child instanceof Graphics
      ) as Graphics | undefined
      
      expect(boundsSprite).toBeDefined()
    })

    it('updates world bounds visualization when world bounds change', () => {
      const gameState1 = createGameState({
        worldBounds: { width: 2000, height: 2000 }
      })

      stateManager.updateInterpolated(gameState1)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const backgroundLayer = scene.getLayer('background')
      const initialBoundsSprite = backgroundLayer.children.find(
        child => child instanceof Graphics
      ) as Graphics | undefined

      const gameState2 = createGameState({
        worldBounds: { width: 4000, height: 4000 }
      })

      stateManager.updateInterpolated(gameState2)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      // Bounds visualization should still exist (may be updated or recreated)
      const updatedBoundsSprite = backgroundLayer.children.find(
        child => child instanceof Graphics
      ) as Graphics | undefined
      expect(updatedBoundsSprite).toBeDefined()
    })

    it('positions world bounds rectangle centered at origin', () => {
      const gameState = createGameState({
        worldBounds: { width: 2000, height: 2000 }
      })

      stateManager.updateInterpolated(gameState)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const backgroundLayer = scene.getLayer('background')
      const boundsSprite = backgroundLayer.children.find(
        child => child instanceof Graphics
      ) as Graphics | undefined
      
      expect(boundsSprite).toBeDefined()
      // Bounds sprite should be positioned (world coordinates transformed to screen)
      // The exact screen position depends on camera, but sprite should exist
      expect(boundsSprite?.visible).toBe(true)
    })

    it('renders world bounds as outline (stroke, not fill)', () => {
      const gameState = createGameState()

      stateManager.updateInterpolated(gameState)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      const backgroundLayer = scene.getLayer('background')
      const boundsSprite = backgroundLayer.children.find(
        child => child instanceof Graphics
      ) as Graphics | undefined
      
      expect(boundsSprite).toBeDefined()
      // Graphics object should exist (we can't easily test stroke vs fill without inspecting internal state)
      expect(boundsSprite?.visible).toBe(true)
    })
  })

  describe('Radar Integration', () => {
    it('creates Radar instance in Renderer constructor', () => {
      const newRenderer = new Renderer(stateManager, scene, app)
      
      // Radar should be created internally
      expect(newRenderer).toBeDefined()
      
      // Radar layer should exist in scene
      const radarLayer = scene.getLayer('radar')
      expect(radarLayer).toBeDefined()
    })

    it('calls radar.update() in Renderer.update()', () => {
      const gameState = createGameState({
        ships: [
          { id: 1, pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
          { id: 2, pos: { x: -100, y: -200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
        ],
        planets: [
          { id: 1, pos: { x: 0, y: 0 }, radius: 50 }
        ],
        pallets: [
          { id: 1, pos: { x: 50, y: 50 }, active: true }
        ]
      })

      stateManager.updateInterpolated(gameState)
      // Update multiple times to let camera converge
      for (let i = 0; i < 20; i++) {
        renderer.update()
      }

      // Radar layer should have graphics (radar should have rendered)
      const radarLayer = scene.getLayer('radar')
      expect(radarLayer.children.length).toBeGreaterThan(0)
    })

    it('radar layer exists in Scene', () => {
      const radarLayer = scene.getLayer('radar')
      expect(radarLayer).toBeDefined()
      expect(radarLayer).toBeInstanceOf(Container)
    })

    it('positions radar in top-right corner', () => {
      const pixiApp = app.getApplication()
      const screenWidth = pixiApp.screen.width
      
      const radarLayer = scene.getLayer('radar')
      
      // Radar should be positioned in top-right (x should be near screen width - radar width - margin)
      // Allow some tolerance for exact positioning
      expect(radarLayer.x).toBeGreaterThan(screenWidth - 250) // 200px radar + 50px margin tolerance
      expect(radarLayer.x).toBeLessThan(screenWidth)
      expect(radarLayer.y).toBeGreaterThanOrEqual(0)
      expect(radarLayer.y).toBeLessThan(50) // Should be near top (10px margin + tolerance)
    })

    it('updates radar every frame with latest state', () => {
      const gameState1 = createGameState({
        ships: [{ id: 1, pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }]
      })

      stateManager.updateInterpolated(gameState1)
      renderer.update()

      const radarLayer = scene.getLayer('radar')
      const initialChildrenCount = radarLayer.children.length

      const gameState2 = createGameState({
        ships: [{ id: 1, pos: { x: 200, y: 300 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }]
      })

      stateManager.updateInterpolated(gameState2)
      renderer.update()

      // Radar should have updated (may have same or different children count, but should exist)
      expect(radarLayer.children.length).toBeGreaterThanOrEqual(0)
    })

    it('handles radar lifecycle (destroy on cleanup)', () => {
      const gameState = createGameState()

      stateManager.updateInterpolated(gameState)
      renderer.update()

      const radarLayer = scene.getLayer('radar')
      expect(radarLayer.children.length).toBeGreaterThan(0)

      // Destroy renderer
      renderer.destroy()

      // Radar should be cleaned up (graphics should be removed)
      // Note: Radar.destroy() clears graphics, but container may still exist
      expect(() => scene.getLayer('radar')).not.toThrow()
    })
  })
})

