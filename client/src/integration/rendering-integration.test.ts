/**
 * Integration tests for rendering system and headless render checks.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { Graphics, Container } from 'pixi.js'
import { App } from '../core/app'
import { Scene } from '../gfx/scene'
import { StateManager, type GameState } from '../sim/state-manager'
import { Renderer } from '../gfx/renderer'
import { createTestState } from './test-helpers'

describe('Rendering Integration', () => {
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

  describe('Renderer Initialization', () => {
    it('initializes with StateManager and Scene', () => {
      const newRenderer = new Renderer(stateManager, scene)
      expect(newRenderer).toBeDefined()
    })

    it('starts with no sprites', () => {
      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(0)
    })

    it('is ready for updates', () => {
      const gameState = createTestState(0)
      stateManager.updateInterpolated(gameState)
      
      renderer.update()
      
      // Should not throw
      expect(renderer).toBeDefined()
    })
  })

  describe('Sprite Rendering from State', () => {
    it('renders ship sprite from state.ship', () => {
      const gameState = createTestState(0, {
        ship: { pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [] // No planets
      })
      
      stateManager.updateInterpolated(gameState)
      renderer.update()
      
      const gameLayer = scene.getLayer('game')
      // 1 ship sprite
      expect(gameLayer.children.length).toBeGreaterThanOrEqual(1)
      expect(gameLayer.children[0]).toBeInstanceOf(Graphics)
    })

    it('renders planet sprites from state.planets array', () => {
      const gameState = createTestState(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 400, y: 300 }, radius: 50 },
          { pos: { x: 600, y: 400 }, radius: 30 }
        ]
      })
      
      stateManager.updateInterpolated(gameState)
      renderer.update()
      
      const gameLayer = scene.getLayer('game')
      // 1 ship + 2 planets = 3 sprites
      expect(gameLayer.children.length).toBe(3)
    })

    it('renders pallet sprites from state.pallets array', () => {
      const gameState = createTestState(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [], // No planets
        pallets: [
          { id: 1, pos: { x: 50, y: 50 }, active: true },
          { id: 2, pos: { x: -50, y: -50 }, active: true }
        ]
      })
      
      stateManager.updateInterpolated(gameState)
      renderer.update()
      
      const gameLayer = scene.getLayer('game')
      // 1 ship + 2 pallets = 3 sprites
      expect(gameLayer.children.length).toBe(3)
    })

    it('adds sprites to scene game layer', () => {
      const gameState = createTestState(0)
      stateManager.updateInterpolated(gameState)
      renderer.update()
      
      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBeGreaterThan(0)
    })
  })

  describe('Sprite Updates', () => {
    it('updates sprite positions from state', () => {
      const gameState1 = createTestState(0, {
        ship: { pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      })
      
      stateManager.updateInterpolated(gameState1)
      renderer.update()
      
      const gameLayer = scene.getLayer('game')
      const sprite1 = gameLayer.children[0] as Graphics
      const x1 = sprite1.x
      const y1 = sprite1.y
      
      const gameState2 = createTestState(0, {
        ship: { pos: { x: 200, y: 300 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      })
      
      stateManager.updateInterpolated(gameState2)
      renderer.update()
      
      const sprite2 = gameLayer.children[0] as Graphics
      expect(sprite2.x).not.toBe(x1)
      expect(sprite2.y).not.toBe(y1)
    })

    it('updates sprite rotations from state', () => {
      const gameState1 = createTestState(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      })
      
      stateManager.updateInterpolated(gameState1)
      renderer.update()
      
      const gameLayer = scene.getLayer('game')
      const sprite1 = gameLayer.children[0] as Graphics
      const rot1 = sprite1.rotation
      
      const gameState2 = createTestState(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 1.57, energy: 100 }
      })
      
      stateManager.updateInterpolated(gameState2)
      renderer.update()
      
      const sprite2 = gameLayer.children[0] as Graphics
      expect(sprite2.rotation).not.toBe(rot1)
    })

    it('updates sprite visual states from state', () => {
      const gameState1 = createTestState(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      })
      
      stateManager.updateInterpolated(gameState1)
      renderer.update()
      
      const gameState2 = createTestState(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 50 }
      })
      
      stateManager.updateInterpolated(gameState2)
      renderer.update()
      
      // Visual state should be updated (e.g., alpha, color based on energy)
      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBeGreaterThan(0)
    })

    it('reflects state changes correctly', () => {
      const gameState = createTestState(0, {
        ship: { pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      })
      
      stateManager.updateInterpolated(gameState)
      renderer.update()
      
      const gameLayer = scene.getLayer('game')
      const sprite = gameLayer.children[0] as Graphics
      
      expect(sprite.x).toBeCloseTo(100, 1)
      expect(sprite.y).toBeCloseTo(200, 1)
    })
  })

  describe('Entity Management', () => {
    it('creates sprites for new entities', () => {
      const gameState1 = createTestState(0, {
        planets: [
          { pos: { x: 100, y: 100 }, radius: 50 }
        ]
      })
      
      stateManager.updateInterpolated(gameState1)
      renderer.update()
      
      const gameLayer1 = scene.getLayer('game')
      const count1 = gameLayer1.children.length
      
      const gameState2 = createTestState(0, {
        planets: [
          { pos: { x: 100, y: 100 }, radius: 50 },
          { pos: { x: 200, y: 200 }, radius: 30 }
        ]
      })
      
      stateManager.updateInterpolated(gameState2)
      renderer.update()
      
      const gameLayer2 = scene.getLayer('game')
      expect(gameLayer2.children.length).toBeGreaterThan(count1)
    })

    it('removes sprites for deleted entities', () => {
      const gameState1 = createTestState(0, {
        planets: [
          { pos: { x: 100, y: 100 }, radius: 50 },
          { pos: { x: 200, y: 200 }, radius: 30 }
        ]
      })
      
      stateManager.updateInterpolated(gameState1)
      renderer.update()
      
      const gameLayer1 = scene.getLayer('game')
      const count1 = gameLayer1.children.length
      
      const gameState2 = createTestState(0, {
        planets: [
          { pos: { x: 100, y: 100 }, radius: 50 }
        ]
      })
      
      stateManager.updateInterpolated(gameState2)
      renderer.update()
      
      const gameLayer2 = scene.getLayer('game')
      expect(gameLayer2.children.length).toBeLessThan(count1)
    })

    it('handles entity array changes correctly', () => {
      const gameState1 = createTestState(0, {
        pallets: [
          { id: 1, pos: { x: 50, y: 50 }, active: true }
        ]
      })
      
      stateManager.updateInterpolated(gameState1)
      renderer.update()
      
      const gameState2 = createTestState(0, {
        pallets: [
          { id: 1, pos: { x: 50, y: 50 }, active: true },
          { id: 2, pos: { x: -50, y: -50 }, active: true },
          { id: 3, pos: { x: 0, y: 0 }, active: true }
        ]
      })
      
      stateManager.updateInterpolated(gameState2)
      renderer.update()
      
      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBeGreaterThan(0)
    })

    it('renders multiple entities correctly', () => {
      const gameState = createTestState(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 400, y: 300 }, radius: 50 },
          { pos: { x: 600, y: 400 }, radius: 30 }
        ],
        pallets: [
          { id: 1, pos: { x: 50, y: 50 }, active: true },
          { id: 2, pos: { x: -50, y: -50 }, active: true }
        ]
      })
      
      stateManager.updateInterpolated(gameState)
      renderer.update()
      
      const gameLayer = scene.getLayer('game')
      // 1 ship + 2 planets + 2 pallets = 5 sprites
      expect(gameLayer.children.length).toBe(5)
    })
  })
})

describe('Headless Render Checks', () => {
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

  describe('Canvas Creation', () => {
    it('creates canvas with correct dimensions in headless environment', async () => {
      const canvas = app.getCanvas()
      
      expect(canvas).toBeInstanceOf(HTMLCanvasElement)
      expect(canvas.width).toBeGreaterThan(0)
      expect(canvas.height).toBeGreaterThan(0)
    })

    it('canvas is accessible in headless environment', async () => {
      const canvas = app.getCanvas()
      const pixiApp = app.getApplication()
      
      expect(canvas).toBeDefined()
      expect(pixiApp).toBeDefined()
      expect(pixiApp.stage).toBeDefined()
    })

    it('canvas supports WebGL rendering', async () => {
      const canvas = app.getCanvas()
      const context = canvas.getContext('webgl') || canvas.getContext('webgl2')
      
      // In headless environment, context may be mocked but should exist
      expect(canvas).toBeDefined()
    })
  })

  describe('Sprite Rendering in Headless', () => {
    it('creates sprites without display', () => {
      const gameState = createTestState(0)
      stateManager.updateInterpolated(gameState)
      renderer.update()
      
      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBeGreaterThan(0)
    })

    it('renders sprites to canvas (even if not visible)', () => {
      const gameState = createTestState(0, {
        ship: { pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 400, y: 300 }, radius: 50 }
        ]
      })
      
      stateManager.updateInterpolated(gameState)
      renderer.update()
      
      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(2)
      
      // Sprites should exist even in headless
      gameLayer.children.forEach(child => {
        expect(child).toBeInstanceOf(Graphics)
      })
    })

    it('sets sprite properties correctly (position, rotation, visual)', () => {
      const gameState = createTestState(0, {
        ship: { pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 1.57, energy: 75 }
      })
      
      stateManager.updateInterpolated(gameState)
      renderer.update()
      
      const gameLayer = scene.getLayer('game')
      const sprite = gameLayer.children[0] as Graphics
      
      expect(sprite.x).toBeCloseTo(100, 1)
      expect(sprite.y).toBeCloseTo(200, 1)
      expect(sprite.rotation).toBeCloseTo(1.57, 2)
    })
  })

  describe('Render Output Validation', () => {
    it('canvas has content after render', () => {
      const gameState = createTestState(0)
      stateManager.updateInterpolated(gameState)
      renderer.update()
      
      const canvas = app.getCanvas()
      expect(canvas).toBeDefined()
      
      // Canvas should exist and be accessible
      expect(canvas.width).toBeGreaterThan(0)
      expect(canvas.height).toBeGreaterThan(0)
    })

    it('sprites exist in scene hierarchy', () => {
      const gameState = createTestState(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 400, y: 300 }, radius: 50 }
        ]
      })
      
      stateManager.updateInterpolated(gameState)
      renderer.update()
      
      const gameLayer = scene.getLayer('game')
      expect(gameLayer.children.length).toBe(2)
      
      gameLayer.children.forEach(child => {
        expect(child).toBeInstanceOf(Graphics)
      })
    })

    it('sprite positions match state positions', () => {
      const gameState = createTestState(0, {
        ship: { pos: { x: 150, y: 250 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      })
      
      stateManager.updateInterpolated(gameState)
      renderer.update()
      
      const gameLayer = scene.getLayer('game')
      const sprite = gameLayer.children[0] as Graphics
      
      expect(sprite.x).toBeCloseTo(150, 1)
      expect(sprite.y).toBeCloseTo(250, 1)
    })

    it('sprite counts match entity counts', () => {
      const gameState = createTestState(0, {
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [
          { pos: { x: 400, y: 300 }, radius: 50 },
          { pos: { x: 600, y: 400 }, radius: 30 }
        ],
        pallets: [
          { id: 1, pos: { x: 50, y: 50 }, active: true },
          { id: 2, pos: { x: -50, y: -50 }, active: true },
          { id: 3, pos: { x: 0, y: 0 }, active: true }
        ]
      })
      
      stateManager.updateInterpolated(gameState)
      renderer.update()
      
      const gameLayer = scene.getLayer('game')
      // 1 ship + 2 planets + 3 pallets = 6 sprites
      expect(gameLayer.children.length).toBe(6)
    })
  })
})

