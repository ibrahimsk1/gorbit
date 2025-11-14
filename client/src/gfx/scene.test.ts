/**
 * Integration tests for Scene hierarchy and container management.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { Container } from 'pixi.js'
import { App } from '../core/app'
import { Scene } from './scene'

describe('Scene Hierarchy', () => {
  let app: App
  let scene: Scene
  let container: HTMLElement

  beforeEach(async () => {
    container = document.createElement('div')
    container.id = 'app'
    document.body.appendChild(container)
    
    app = new App()
    await app.init(container)
    scene = new Scene(app)
  })

  afterEach(() => {
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

  describe('Container Creation', () => {
    it('creates root container', () => {
      const root = scene.getRoot()
      
      expect(root).toBeInstanceOf(Container)
    })

    it('root container is added to application stage', () => {
      const root = scene.getRoot()
      const pixiApp = app.getApplication()
      
      expect(pixiApp.stage.children).toContain(root)
    })
  })

  describe('Layer Management', () => {
    it('creates named layers', () => {
      const backgroundLayer = scene.getLayer('background')
      const gameLayer = scene.getLayer('game')
      const uiLayer = scene.getLayer('ui')
      
      expect(backgroundLayer).toBeInstanceOf(Container)
      expect(gameLayer).toBeInstanceOf(Container)
      expect(uiLayer).toBeInstanceOf(Container)
    })

    it('returns same layer instance for same name', () => {
      const layer1 = scene.getLayer('background')
      const layer2 = scene.getLayer('background')
      
      expect(layer1).toBe(layer2)
    })

    it('layers are children of root container', () => {
      const root = scene.getRoot()
      const backgroundLayer = scene.getLayer('background')
      const gameLayer = scene.getLayer('game')
      
      expect(root.children).toContain(backgroundLayer)
      expect(root.children).toContain(gameLayer)
    })
  })

  describe('Child Management', () => {
    it('adds child to root container', () => {
      const root = scene.getRoot()
      const child = new Container()
      
      scene.addChild(child)
      
      expect(root.children).toContain(child)
    })

    it('adds child to specific layer', () => {
      const gameLayer = scene.getLayer('game')
      const child = new Container()
      
      scene.addChild(child, 'game')
      
      expect(gameLayer.children).toContain(child)
    })

    it('removes child from container', () => {
      const root = scene.getRoot()
      const child = new Container()
      
      scene.addChild(child)
      expect(root.children).toContain(child)
      
      scene.removeChild(child)
      expect(root.children).not.toContain(child)
    })

    it('removes child from specific layer', () => {
      const gameLayer = scene.getLayer('game')
      const child = new Container()
      
      scene.addChild(child, 'game')
      expect(gameLayer.children).toContain(child)
      
      scene.removeChild(child, 'game')
      expect(gameLayer.children).not.toContain(child)
    })
  })
})


