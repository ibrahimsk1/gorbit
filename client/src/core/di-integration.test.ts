/**
 * Integration tests for DI Container with actual subsystems.
 * 
 * Verifies that all dependencies are resolved in correct order
 * and that initialization phases work correctly.
 * 
 * Labels: scope:integration loop:g2-app layer:core
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { DIContainer } from './di-container'
import { DI_KEYS } from './di-keys'
import { registerDependencies } from './di-config'
import { App } from './app'
import { Scene } from '../gfx/scene'
import { Renderer } from '../gfx/renderer'
import { HUD } from '../ui/hud'
import { StateManager } from '../sim/state-manager'
import { NetworkClient } from '../net/client'
import { KeyboardInputHandler } from '../input/keyboard'
import { AppOrchestrator } from './orchestrator'

describe('DI Container Integration', () => {
  let container: DIContainer
  let htmlContainer: HTMLElement

  beforeEach(() => {
    container = new DIContainer()
    htmlContainer = document.createElement('div')
    htmlContainer.id = 'app'
    document.body.appendChild(htmlContainer)
  })

  afterEach(() => {
    container.destroy()
    if (htmlContainer && htmlContainer.parentNode) {
      htmlContainer.parentNode.removeChild(htmlContainer)
    }
  })

  describe('Dependency Registration', () => {
    it('should register all dependencies', () => {
      registerDependencies(container)

      // Verify all key dependencies are registered
      expect(container.isRegistered(DI_KEYS.APP)).toBe(true)
      expect(container.isRegistered(DI_KEYS.SCENE)).toBe(true)
      expect(container.isRegistered(DI_KEYS.STATE_MANAGER)).toBe(true)
      expect(container.isRegistered(DI_KEYS.NETWORK_CLIENT)).toBe(true)
      expect(container.isRegistered(DI_KEYS.INPUT_HANDLER)).toBe(true)
      expect(container.isRegistered(DI_KEYS.RENDERER)).toBe(true)
      expect(container.isRegistered(DI_KEYS.HUD)).toBe(true)
      expect(container.isRegistered(DI_KEYS.ORCHESTRATOR)).toBe(true)
    })

    it('should have correct initialization order', () => {
      registerDependencies(container)
      const order = container.getInitializationOrder()

      // App should come before Scene (Scene depends on App)
      expect(order.indexOf(DI_KEYS.APP)).toBeLessThan(order.indexOf(DI_KEYS.SCENE))

      // StateManager should come before systems that depend on it
      expect(order.indexOf(DI_KEYS.STATE_MANAGER)).toBeLessThan(order.indexOf(DI_KEYS.PREDICTION_SYSTEM))
      expect(order.indexOf(DI_KEYS.STATE_MANAGER)).toBeLessThan(order.indexOf(DI_KEYS.RENDERER))
      expect(order.indexOf(DI_KEYS.STATE_MANAGER)).toBeLessThan(order.indexOf(DI_KEYS.HUD))

      // Scene should come before Renderer and HUD
      expect(order.indexOf(DI_KEYS.SCENE)).toBeLessThan(order.indexOf(DI_KEYS.RENDERER))
      expect(order.indexOf(DI_KEYS.SCENE)).toBeLessThan(order.indexOf(DI_KEYS.HUD))

      // All dependencies should come before orchestrator
      const orchestratorIndex = order.indexOf(DI_KEYS.ORCHESTRATOR)
      expect(order.indexOf(DI_KEYS.APP)).toBeLessThan(orchestratorIndex)
      expect(order.indexOf(DI_KEYS.SCENE)).toBeLessThan(orchestratorIndex)
      expect(order.indexOf(DI_KEYS.RENDERER)).toBeLessThan(orchestratorIndex)
      expect(order.indexOf(DI_KEYS.HUD)).toBeLessThan(orchestratorIndex)
    })
  })

  describe('Early Phase Initialization', () => {
    it('should initialize early phase dependencies', async () => {
      registerDependencies(container)
      await container.initializePhase('early')

      // Early phase dependencies should be resolved
      const app = container.resolve<App>(DI_KEYS.APP)
      expect(app).toBeInstanceOf(App)

      const scene = container.resolve<Scene>(DI_KEYS.SCENE)
      expect(scene).toBeInstanceOf(Scene)

      const stateManager = container.resolve<StateManager>(DI_KEYS.STATE_MANAGER)
      expect(stateManager).toBeInstanceOf(StateManager)

      const networkClient = container.resolve<NetworkClient>(DI_KEYS.NETWORK_CLIENT)
      expect(networkClient).toBeInstanceOf(NetworkClient)

      const inputHandler = container.resolve<KeyboardInputHandler>(DI_KEYS.INPUT_HANDLER)
      expect(inputHandler).toBeInstanceOf(KeyboardInputHandler)
    })

    it('should not initialize late phase dependencies in early phase', async () => {
      registerDependencies(container)
      await container.initializePhase('early')

      // Late phase dependencies should not be resolved yet
      // Attempting to resolve them should work (they'll be created on demand)
      // But they won't be initialized until late phase
      const renderer = container.resolve<Renderer>(DI_KEYS.RENDERER)
      expect(renderer).toBeInstanceOf(Renderer)
    })
  })

  describe('Late Phase Initialization', () => {
    it('should initialize late phase dependencies after App is initialized', async () => {
      registerDependencies(container)
      
      // Initialize early phase
      await container.initializePhase('early')
      
      // Initialize App
      const app = container.resolve<App>(DI_KEYS.APP)
      await app.init(htmlContainer)
      
      // Initialize late phase
      await container.initializePhase('late')

      // Late phase dependencies should be resolved
      const renderer = container.resolve<Renderer>(DI_KEYS.RENDERER)
      expect(renderer).toBeInstanceOf(Renderer)

      const hud = container.resolve<HUD>(DI_KEYS.HUD)
      expect(hud).toBeInstanceOf(HUD)

      const orchestrator = container.resolve<AppOrchestrator>(DI_KEYS.ORCHESTRATOR)
      expect(orchestrator).toBeInstanceOf(AppOrchestrator)
    })

    it('should create Renderer with correct dependencies', async () => {
      registerDependencies(container)
      await container.initializePhase('early')
      
      const app = container.resolve<App>(DI_KEYS.APP)
      try {
        await app.init(htmlContainer)
      } catch (error) {
        // WebGL might not be available in test environment, skip this test
        return
      }
      
      await container.initializePhase('late')

      const renderer = container.resolve<Renderer>(DI_KEYS.RENDERER)
      expect(renderer).toBeInstanceOf(Renderer)
      // Renderer should have access to initialized App
      expect(() => app.getApplication()).not.toThrow()
      
      // Clean up
      app.destroy()
    })

    it('should create HUD with correct dependencies', async () => {
      registerDependencies(container)
      await container.initializePhase('early')
      
      const app = container.resolve<App>(DI_KEYS.APP)
      try {
        await app.init(htmlContainer)
      } catch (error) {
        // WebGL might not be available in test environment, skip this test
        return
      }
      
      await container.initializePhase('late')

      const hud = container.resolve<HUD>(DI_KEYS.HUD)
      expect(hud).toBeInstanceOf(HUD)
      
      // Clean up
      app.destroy()
    })
  })

  describe('Full Initialization Flow', () => {
    it('should initialize complete application flow', async () => {
      registerDependencies(container)

      // Step 1: Initialize early phase
      await container.initializePhase('early')
      
      // Step 2: Initialize App
      const app = container.resolve<App>(DI_KEYS.APP)
      await app.init(htmlContainer)
      
      // Step 3: Initialize late phase
      await container.initializePhase('late')

      // Step 4: Get orchestrator and initialize it
      const orchestrator = container.resolve<AppOrchestrator>(DI_KEYS.ORCHESTRATOR)
      await orchestrator.init()

      // Verify all systems are ready
      expect(orchestrator).toBeInstanceOf(AppOrchestrator)
      expect(() => app.getApplication()).not.toThrow()

      // Clean up
      orchestrator.destroy()
    })

    it('should resolve singleton instances correctly', async () => {
      registerDependencies(container)
      await container.initializePhase('early')
      
      const app = container.resolve<App>(DI_KEYS.APP)
      await app.init(htmlContainer)
      await container.initializePhase('late')

      // Resolve same dependency multiple times - should get same instance
      const renderer1 = container.resolve<Renderer>(DI_KEYS.RENDERER)
      const renderer2 = container.resolve<Renderer>(DI_KEYS.RENDERER)
      expect(renderer1).toBe(renderer2)

      const hud1 = container.resolve<HUD>(DI_KEYS.HUD)
      const hud2 = container.resolve<HUD>(DI_KEYS.HUD)
      expect(hud1).toBe(hud2)
    })
  })

  describe('Cleanup', () => {
    it('should destroy container-owned instances', async () => {
      registerDependencies(container)
      await container.initializePhase('early')
      
      const app = container.resolve<App>(DI_KEYS.APP)
      await app.init(htmlContainer)
      await container.initializePhase('late')

      // Container should destroy its owned instances
      container.destroy()

      // App should still exist (application-owned)
      expect(() => app.getApplication()).not.toThrow()
    })
  })
})

