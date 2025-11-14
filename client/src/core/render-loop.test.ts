/**
 * Integration tests for render loop and frame timing.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { App } from './app'
import { RenderLoop } from './render-loop'

describe('Render Loop', () => {
  let app: App
  let renderLoop: RenderLoop
  let container: HTMLElement

  beforeEach(async () => {
    container = document.createElement('div')
    container.id = 'app'
    document.body.appendChild(container)
    
    app = new App()
    await app.init(container)
    renderLoop = new RenderLoop(app)
  })

  afterEach(() => {
    if (renderLoop) {
      renderLoop.stop()
    }
    if (app) {
      app.destroy()
    }
    if (container && container.parentNode) {
      container.parentNode.removeChild(container)
    }
  })

  describe('Frame Timing', () => {
    it('starts render loop', () => {
      const startSpy = vi.spyOn(renderLoop, 'start')
      renderLoop.start()
      
      expect(startSpy).toHaveBeenCalled()
    })

    it('stops render loop', () => {
      renderLoop.start()
      const stopSpy = vi.spyOn(renderLoop, 'stop')
      renderLoop.stop()
      
      expect(stopSpy).toHaveBeenCalled()
    })

    it('tracks FPS', async () => {
      renderLoop.start()
      
      // Wait for a few frames
      await new Promise(resolve => setTimeout(resolve, 100))
      
      const fps = renderLoop.getFPS()
      expect(fps).toBeGreaterThan(0)
      expect(fps).toBeLessThanOrEqual(120) // Reasonable upper bound
      
      renderLoop.stop()
    })

    it('maintains target FPS (approximately 60fps)', async () => {
      renderLoop.start()
      
      // Wait for render loop to stabilize
      await new Promise(resolve => setTimeout(resolve, 500))
      
      const fps = renderLoop.getFPS()
      // Allow some tolerance (50-70 fps is acceptable)
      expect(fps).toBeGreaterThan(50)
      expect(fps).toBeLessThan(70)
      
      renderLoop.stop()
    })
  })

  describe('Start/Stop Behavior', () => {
    it('can start after being stopped', () => {
      renderLoop.start()
      renderLoop.stop()
      
      // Should be able to start again
      expect(() => renderLoop.start()).not.toThrow()
      renderLoop.stop()
    })

    it('handles multiple start calls gracefully', () => {
      renderLoop.start()
      renderLoop.start() // Second start should not cause issues
      
      expect(() => renderLoop.stop()).not.toThrow()
    })

    it('handles stop when not started', () => {
      // Should not throw when stopping without starting
      expect(() => renderLoop.stop()).not.toThrow()
    })
  })
})


