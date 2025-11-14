/**
 * Render loop manager for controlling frame updates and FPS tracking.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { App } from './app'

export class RenderLoop {
  private app: App
  private animationFrameId: number | null = null
  private isRunning: boolean = false
  private lastFrameTime: number = 0
  private frameCount: number = 0
  private fps: number = 0
  private fpsUpdateInterval: number = 1000 // Update FPS every second
  private lastFpsUpdate: number = 0

  constructor(app: App) {
    this.app = app
  }

  start(): void {
    if (this.isRunning) {
      return
    }

    this.isRunning = true
    this.lastFrameTime = performance.now()
    this.lastFpsUpdate = this.lastFrameTime
    this.frameCount = 0
    this.fps = 0

    const loop = (currentTime: number) => {
      if (!this.isRunning) {
        return
      }

      this.lastFrameTime = currentTime

      // Update FPS calculation
      this.frameCount++
      if (currentTime - this.lastFpsUpdate >= this.fpsUpdateInterval) {
        this.fps = (this.frameCount * 1000) / (currentTime - this.lastFpsUpdate)
        this.frameCount = 0
        this.lastFpsUpdate = currentTime
      }

      // Render frame
      try {
        const pixiApp = this.app.getApplication()
        pixiApp.render()
      } catch (error) {
        // App may have been destroyed
        this.stop()
        return
      }

      this.animationFrameId = requestAnimationFrame(loop)
    }

    this.animationFrameId = requestAnimationFrame(loop)
  }

  stop(): void {
    this.isRunning = false
    if (this.animationFrameId !== null) {
      cancelAnimationFrame(this.animationFrameId)
      this.animationFrameId = null
    }
  }

  getFPS(): number {
    return this.fps
  }
}

