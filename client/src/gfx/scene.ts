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

  constructor(app: App) {
    this.app = app
    this.root = new Container()
    
    // Add root to application stage
    const pixiApp = this.app.getApplication()
    pixiApp.stage.addChild(this.root)

    // Create default layers
    this.getLayer('background')
    this.getLayer('game')
    this.getLayer('ui')
  }

  getRoot(): Container {
    return this.root
  }

  getLayer(name: string): Container {
    if (!this.layers.has(name)) {
      const layer = new Container()
      this.layers.set(name, layer)
      this.root.addChild(layer)
    }
    return this.layers.get(name)!
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
    // Remove root from stage
    const pixiApp = this.app.getApplication()
    if (pixiApp.stage.children.includes(this.root)) {
      pixiApp.stage.removeChild(this.root)
    }

    // Clean up
    this.root.destroy({ children: true })
    this.layers.clear()
  }
}


