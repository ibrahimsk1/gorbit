/**
 * AppOrchestrator coordinates client subsystems (net, gfx, ui, input, sim).
 * 
 * Labels: scope:unit loop:g2-app layer:core double:fake-io
 */

import { App } from './app'
import { RenderLoop } from './render-loop'
import { Scene } from '../gfx/scene'
import type { Container } from 'pixi.js'

// Subsystem interfaces for dependency injection
export interface INetworkClient {
  onSnapshot(callback: (snapshot: any) => void): void
  onConnect(callback: () => void): void
  onDisconnect(callback: () => void): void
  onError(callback: (error: Error) => void): void
  connect(url: string): Promise<void>
  isConnected(): boolean
}

export interface IRenderer {
  update(): void
  destroy(): void
}

export interface IHUD {
  update(): void
  destroy(): void
}

export interface IInputHandler {
  getThrust(): number
  getTurn(): number
  attach(): void
  detach(): void
  reset(): void
}

export interface IStateManager {
  updateAuthoritative(snapshot: any): void
}

export interface IScene {
  getRoot(): Container
  getLayer(name: string): Container
  destroy(): void
}

/**
 * AppOrchestrator coordinates all client subsystems.
 * Handles initialization, render loop, lifecycle, and UI state transitions.
 */
export class AppOrchestrator {
  private app: App
  private renderLoop: RenderLoop | null = null
  private networkClient: INetworkClient | null = null
  private renderer: IRenderer | null = null
  private hud: IHUD | null = null
  private inputHandler: IInputHandler | null = null
  private stateManager: IStateManager | null = null
  private scene: IScene | null = null
  private initialized: boolean = false
  private initError: Error | null = null
  private gameLoopId: number | null = null
  private isGameLoopRunning: boolean = false

  constructor(
    app: App,
    subsystems: {
      networkClient?: INetworkClient | (() => INetworkClient)
      renderer?: IRenderer | (() => IRenderer)
      hud?: IHUD | (() => IHUD)
      inputHandler?: IInputHandler | (() => IInputHandler)
      stateManager?: IStateManager | (() => IStateManager)
      scene?: IScene | (() => IScene)
    } = {}
  ) {
    this.app = app

    // Resolve subsystems (handle both instances and factory functions)
    if (subsystems.networkClient) {
      this.networkClient = typeof subsystems.networkClient === 'function'
        ? subsystems.networkClient()
        : subsystems.networkClient
    }

    if (subsystems.renderer) {
      this.renderer = typeof subsystems.renderer === 'function'
        ? subsystems.renderer()
        : subsystems.renderer
    }

    if (subsystems.hud) {
      this.hud = typeof subsystems.hud === 'function'
        ? subsystems.hud()
        : subsystems.hud
    }

    if (subsystems.inputHandler) {
      this.inputHandler = typeof subsystems.inputHandler === 'function'
        ? subsystems.inputHandler()
        : subsystems.inputHandler
    }

    if (subsystems.stateManager) {
      this.stateManager = typeof subsystems.stateManager === 'function'
        ? subsystems.stateManager()
        : subsystems.stateManager
    }

    if (subsystems.scene) {
      this.scene = typeof subsystems.scene === 'function'
        ? subsystems.scene()
        : subsystems.scene
    }
  }

  /**
   * Initializes the app and all subsystems.
   * Initializes PixiJS App, creates Scene, and sets up subsystem event handlers.
   * 
   * @param container Optional container element for App initialization
   * @throws Error if initialization fails
   */
  async init(container?: HTMLElement): Promise<void> {
    if (this.initialized) {
      // Already initialized, skip
      return
    }

    try {
      // Initialize PixiJS App
      await this.app.init(container)

      // Create Scene if not provided
      if (!this.scene) {
        this.scene = new Scene(this.app)
      }

      // Set up NetworkClient event handlers if provided
      if (this.networkClient && this.stateManager) {
        this.setupNetworkEventHandlers()
      }

      this.initialized = true
      this.initError = null
    } catch (error) {
      this.initError = error instanceof Error ? error : new Error(String(error))
      // Log error and rethrow to allow caller to handle
      console.error('AppOrchestrator initialization error:', this.initError)
      throw this.initError
    }
  }

  /**
   * Sets up event handlers for NetworkClient.
   * Private helper method called during initialization.
   */
  private setupNetworkEventHandlers(): void {
    if (!this.networkClient || !this.stateManager) {
      return
    }

    // Set up snapshot handler
    this.networkClient.onSnapshot((snapshot) => {
      this.stateManager!.updateAuthoritative(snapshot)
    })

    // Set up connection handlers
    this.networkClient.onConnect(() => {
      console.log('Connected to game server')
    })

    this.networkClient.onDisconnect(() => {
      console.log('Disconnected from game server')
    })

    this.networkClient.onError((error) => {
      console.error('Network error:', error)
    })
  }

  /**
   * Starts the render loop and game loop.
   * Creates and starts RenderLoop, then starts game loop that updates renderer and HUD.
   * 
   * @throws Error if not initialized
   */
  start(): void {
    if (!this.initialized) {
      throw new Error('AppOrchestrator must be initialized before starting')
    }

    if (this.isGameLoopRunning) {
      // Already running, skip
      return
    }

    // Create and start RenderLoop
    if (!this.renderLoop) {
      this.renderLoop = new RenderLoop(this.app)
    }
    this.renderLoop.start()

    // Start game loop
    this.isGameLoopRunning = true
    const gameLoop = () => {
      if (!this.isGameLoopRunning) {
        return
      }

      // Update renderer if provided
      if (this.renderer) {
        this.renderer.update()
      }

      // Update HUD if provided
      if (this.hud) {
        this.hud.update()
      }

      // Continue game loop
      this.gameLoopId = requestAnimationFrame(gameLoop)
    }

    this.gameLoopId = requestAnimationFrame(gameLoop)
  }

  /**
   * Stops the render loop and game loop.
   * Basic implementation - full lifecycle management will be in CU cu/subsystem-lifecycle.
   */
  stop(): void {
    // Stop game loop
    this.isGameLoopRunning = false
    if (this.gameLoopId !== null) {
      cancelAnimationFrame(this.gameLoopId)
      this.gameLoopId = null
    }

    // Stop RenderLoop
    if (this.renderLoop) {
      this.renderLoop.stop()
    }
  }

  /**
   * Destroys all subsystems and cleans up resources.
   * Will be implemented in CU cu/subsystem-lifecycle.
   */
  destroy(): void {
    // Placeholder - implementation in CU cu/subsystem-lifecycle
  }
}

