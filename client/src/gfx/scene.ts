/**
 * Scene hierarchy manager for organizing game objects into layers.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { Container } from 'pixi.js'
import { App } from '../core/app'

export class Scene {
  private app: App
  private root: Container
  private layers: Map<string, Container> = new Map()
  private attached: boolean = false

  constructor(app: App) {
    this.app = app
    this.root = new Container()
    
    // Create default layers (but don't attach to stage yet)
    // Layers are created without requiring App to be initialized
    this.createDefaultLayers()
  }

  /**
   * Creates default layers without attaching to stage.
   * Private helper method called during construction.
   */
  private createDefaultLayers(): void {
    // Create layers directly without calling getLayer (which requires attachment)
    const background = new Container()
    this.layers.set('background', background)
    this.root.addChild(background)

    const game = new Container()
    this.layers.set('game', game)
    this.root.addChild(game)

    const ui = new Container()
    this.layers.set('ui', ui)
    this.root.addChild(ui)
  }

  /**
   * Initializes the scene by attaching to the stage.
   * MUST be called after App is initialized.
   * 
   * @throws Error if App is not initialized
   */
  initialize(): void {
    if (this.attached) {
      return // Idempotent - safe to call multiple times
    }

    // Fail-fast: throw if App not ready (no silent failures)
    const pixiApp = this.app.getApplication()
    
    if (!pixiApp.stage.children.includes(this.root)) {
      pixiApp.stage.addChild(this.root)
      this.attached = true
    }
  }

  /**
   * Checks if scene is initialized.
   * 
   * @returns True if scene is attached to stage
   */
  isInitialized(): boolean {
    return this.attached
  }

  getRoot(): Container {
    return this.root
  }

  /**
   * Gets a layer. Scene must be initialized first.
   * 
   * @param name Layer name
   * @returns Container for the layer
   * @throws Error if Scene is not initialized
   */
  getLayer(name: string): Container {
    if (!this.attached) {
      throw new Error(
        `Scene not initialized. Call scene.initialize() after App.init()`
      )
    }
    
    if (!this.layers.has(name)) {
      const layer = new Container()
      this.layers.set(name, layer)
      this.root.addChild(layer)
    }
    
    const layer = this.layers.get(name)
    if (!layer) {
      // This should never happen, but if it does, provide detailed error
      const availableLayers = Array.from(this.layers.keys()).join(', ')
      throw new Error(
        `Internal error: layer '${name}' not found after creation. ` +
        `Available layers: ${availableLayers || '(none)'}. ` +
        `Scene attached: ${this.attached}, layers map size: ${this.layers.size}`
      )
    }
    
    return layer
  }

  addChild(child: Container, layerName?: string): void {
    if (layerName) {
      const layer = this.getLayer(layerName)
      layer.addChild(child)
    } else {
      this.root.addChild(child)
    }
  }

  removeChild(child: Container, layerName?: string): void {
    if (layerName) {
      const layer = this.layers.get(layerName)
      if (layer && layer.children.includes(child)) {
        layer.removeChild(child)
      }
    } else {
      if (this.root.children.includes(child)) {
        this.root.removeChild(child)
      }
    }
  }

  destroy(): void {
    // Remove root from stage (if attached)
    if (this.attached) {
      try {
        const pixiApp = this.app.getApplication()
        if (pixiApp.stage.children.includes(this.root)) {
          pixiApp.stage.removeChild(this.root)
        }
      } catch (error) {
        // App may have been destroyed already
      }
      this.attached = false
    }

    // Clean up
    this.root.destroy({ children: true })
    this.layers.clear()
  }
}


