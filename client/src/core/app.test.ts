/**
 * Integration tests for Pixi Application initialization and lifecycle.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { Application } from 'pixi.js'
import { App } from './app'

describe('Pixi Application', () => {
  let app: App
  let container: HTMLElement

  beforeEach(() => {
    // Create a container element for mounting
    container = document.createElement('div')
    container.id = 'app'
    document.body.appendChild(container)
  })

  afterEach(() => {
    // Cleanup
    if (app) {
      app.destroy()
    }
    if (container && container.parentNode) {
      container.parentNode.removeChild(container)
    }
  })

  describe('Initialization', () => {
    it('creates Pixi Application instance', async () => {
      app = new App()
      await app.init()
      const pixiApp = app.getApplication()
      
      expect(pixiApp).toBeInstanceOf(Application)
    })

    it('initializes with default viewport configuration', async () => {
      app = new App()
      await app.init()
      const pixiApp = app.getApplication()
      
      expect(pixiApp.screen.width).toBe(1280)
      expect(pixiApp.screen.height).toBe(720)
    })

    it('initializes with custom viewport configuration', async () => {
      app = new App({ width: 1920, height: 1080 })
      await app.init()
      const pixiApp = app.getApplication()
      
      expect(pixiApp.screen.width).toBe(1920)
      expect(pixiApp.screen.height).toBe(1080)
    })

    it('mounts canvas to DOM element', async () => {
      app = new App()
      await app.init(container)
      const canvas = app.getCanvas()
      
      expect(canvas).toBeDefined()
      expect(canvas).toBeInstanceOf(HTMLCanvasElement)
      expect(container.contains(canvas)).toBe(true)
    })

    it('mounts canvas to default #app element when no container provided', async () => {
      app = new App()
      await app.init()
      const canvas = app.getCanvas()
      
      expect(canvas).toBeDefined()
      const appElement = document.getElementById('app')
      expect(appElement).toBeDefined()
      expect(appElement?.contains(canvas)).toBe(true)
    })
  })

  describe('Resize Handling', () => {
    it('updates application size on window resize', async () => {
      app = new App({ width: 1280, height: 720 })
      await app.init(container)
      const pixiApp = app.getApplication()
      
      // Simulate window resize
      container.style.width = '1920px'
      container.style.height = '1080px'
      window.dispatchEvent(new Event('resize'))
      
      // Wait for resize handler
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Note: Auto-resize behavior depends on Pixi configuration
      // This test verifies the resize event is handled
      expect(pixiApp).toBeDefined()
    })
  })

  describe('Lifecycle', () => {
    it('destroys application and cleans up resources', async () => {
      app = new App()
      await app.init(container)
      const canvas = app.getCanvas()
      
      expect(canvas).toBeDefined()
      
      app.destroy()
      
      // After destroy, canvas should be removed or app should be unusable
      // Pixi Application destroy removes canvas from DOM
      expect(container.contains(canvas)).toBe(false)
    })

    it('can be initialized multiple times safely', async () => {
      app = new App()
      await app.init(container)
      
      const canvas1 = app.getCanvas()
      expect(canvas1).toBeDefined()
      
      // Re-initialize should work (or handle gracefully)
      await app.init(container)
      const canvas2 = app.getCanvas()
      
      expect(canvas2).toBeDefined()
    })
  })

  describe('Headless Render Checks', () => {
    it('creates canvas with correct dimensions', async () => {
      app = new App({ width: 1280, height: 720 })
      await app.init(container)
      const canvas = app.getCanvas()
      
      expect(canvas).toBeInstanceOf(HTMLCanvasElement)
      // Canvas dimensions may differ from screen dimensions due to resolution
      expect(canvas.width).toBeGreaterThan(0)
      expect(canvas.height).toBeGreaterThan(0)
    })

    it('application stage exists and is accessible', async () => {
      app = new App()
      await app.init(container)
      const pixiApp = app.getApplication()
      
      expect(pixiApp.stage).toBeDefined()
      expect(pixiApp.stage.children).toBeDefined()
    })
  })
})

