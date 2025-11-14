/**
 * Pixi Application wrapper for initializing and managing the game renderer.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { Application, ApplicationOptions, WebGLRenderer } from 'pixi.js'

export interface AppConfig {
  width?: number
  height?: number
  backgroundColor?: number
  resolution?: number
  autoResize?: boolean
}

export class App {
  private pixiApp: Application | null = null
  private config: Required<AppConfig>
  private resizeHandler: (() => void) | null = null

  constructor(config: AppConfig = {}) {
    this.config = {
      width: config.width ?? 1280,
      height: config.height ?? 720,
      backgroundColor: config.backgroundColor ?? 0x0a0a0a,
      resolution: config.resolution ?? 1,
      autoResize: config.autoResize ?? true,
    }
  }

  async init(container?: HTMLElement): Promise<void> {
    if (this.pixiApp) {
      // Already initialized, clean up first
      this.destroy()
    }

    // Pixi.js v8: Create Application without options, then init with options
    this.pixiApp = new Application()

    const targetContainer = container ?? document.getElementById('app')
    if (!targetContainer) {
      throw new Error('Container element not found. Provide container or ensure #app exists in DOM.')
    }

    const options: Partial<ApplicationOptions> = {
      width: this.config.width,
      height: this.config.height,
      backgroundColor: this.config.backgroundColor,
      resolution: this.config.resolution,
      autoDensity: this.config.autoResize,
      preference: 'webgl', // Prefer WebGL renderer
    }

    await this.pixiApp.init(options)
    targetContainer.appendChild(this.pixiApp.canvas)

    // Setup resize handler
    if (this.config.autoResize) {
      this.resizeHandler = () => {
        if (this.pixiApp && targetContainer) {
          const width = targetContainer.clientWidth || this.config.width
          const height = targetContainer.clientHeight || this.config.height
          this.pixiApp.renderer.resize(width, height)
        }
      }
      window.addEventListener('resize', this.resizeHandler)
      // Initial resize
      this.resizeHandler()
    }
  }

  getApplication(): Application {
    if (!this.pixiApp) {
      throw new Error('Application not initialized. Call init() first.')
    }
    return this.pixiApp
  }

  getCanvas(): HTMLCanvasElement {
    if (!this.pixiApp) {
      throw new Error('Application not initialized. Call init() first.')
    }
    return this.pixiApp.canvas
  }

  destroy(): void {
    if (this.resizeHandler) {
      window.removeEventListener('resize', this.resizeHandler)
      this.resizeHandler = null
    }

    if (this.pixiApp) {
      this.pixiApp.destroy(true, { children: true, texture: true })
      this.pixiApp = null
    }
  }
}

