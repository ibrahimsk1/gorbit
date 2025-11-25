/**
 * AppOrchestrator coordinates client subsystems (net, gfx, ui, input, sim).
 * 
 * Labels: scope:unit loop:g2-app layer:core double:fake-io
 */

import { App } from './app'
import { RenderLoop } from './render-loop'
import { Scene } from '../gfx/scene'
import { MainMenu } from '../ui/main-menu'
import { RoomLobby } from '../ui/room-lobby'
import { Container } from 'pixi.js'
import { DIContainer } from './di-container'
import { DI_KEYS } from './di-keys'
import { Renderer } from '../gfx/renderer'
import { HUD } from '../ui/hud'
import { StateManager } from '../sim/state-manager'
import { CommandHistory } from '../net/command-history'
import { PredictionSystem } from '../sim/prediction'
import { InterpolationSystem } from '../sim/interpolation'
import { ReconciliationSystem } from '../sim/reconciliation'

// UI state type
export type UIState = 'main-menu' | 'lobby' | 'in-game'

// Room management types
export interface PlayerInfo {
  id: number
  name: string
}

export interface RoomState {
  roomCode: string
  players: PlayerInfo[]
  state: 'lobby' | 'playing' | 'ended'
  hostId: number
}

// Subsystem interfaces for dependency injection
export interface INetworkClient {
  onSnapshot(callback: (snapshot: unknown) => void): void
  onConnect(callback: () => void): void
  onDisconnect(callback: () => void): void
  onError(callback: (error: Error) => void): void
  onRoomState(callback: (state: RoomState) => void): void
  onPlayerJoined(callback: (player: PlayerInfo) => void): void
  onPlayerLeft(callback: (playerId: number) => void): void
  onMatchStarted(callback: () => void): void
  onMatchEnded(callback: (winnerId?: number) => void): void
  connect(url: string): Promise<void>
  disconnect(): void
  createRoom(): Promise<void>
  joinRoom(roomCode: string): Promise<void>
  leaveRoom(): void
  startMatch(): void
  sendInput(seq: number, thrust: number, turn: number): void
  isConnected(): boolean
}

export interface IRenderer {
  update(): void
  destroy(): void
}

export interface IMainMenu {
  show(): void
  hide(): void
  destroy(): void
}

export interface IRoomLobby {
  show(): void
  hide(): void
  update(roomState: RoomState): void
  destroy(): void
}

export interface IHUD {
  show(): void
  hide(): void
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
  updateAuthoritative(snapshot: unknown): void
}

export interface IScene {
  getRoot(): Container
  getLayer(name: string): Container
  destroy(): void
}

export interface ICommandHistory {
  addCommand(seq: number, thrust: number, turn: number): void
  getNextSequence(): number
}

export interface IPredictionSystem {
  predict(input: { thrust: number, turn: number }): void
}

export interface IInterpolationSystem {
  addSnapshot(snapshot: unknown, timestamp: number): void
  update(now: number): void
}

export interface IReconciliationSystem {
  reconcile(snapshot: unknown): void
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
  private commandHistory: ICommandHistory | null = null
  private predictionSystem: IPredictionSystem | null = null
  private interpolationSystem: IInterpolationSystem | null = null
  private reconciliationSystem: IReconciliationSystem | null = null
  private initialized: boolean = false
  private initError: Error | null = null
  private gameLoopId: number | null = null
  private isGameLoopRunning: boolean = false
  private inputLoopId: number | null = null
  private isInputLoopRunning: boolean = false
  private interpolationLoopId: number | null = null
  private isInterpolationLoopRunning: boolean = false
  private beforeunloadHandler: (() => void) | null = null
  private uiState: UIState = 'main-menu'
  private mainMenu: IMainMenu | null = null
  private roomLobby: IRoomLobby | null = null
  private inputSendIntervalMs: number = 1000 / 30 // 30Hz input rate
  private lastInputSendTime: number = 0
  private commandSequence: number = 0
  private container: DIContainer | null = null
  private gameSystemsInitialized: boolean = false
  
  constructor(
    dependencies: {
      app: App
      networkClient: INetworkClient
      renderer?: IRenderer  // Optional - lazy loaded from game phase
      hud?: IHUD  // Optional - lazy loaded from game phase
      inputHandler: IInputHandler
      stateManager?: IStateManager  // Optional - lazy loaded from game phase
      scene: IScene
      commandHistory?: ICommandHistory  // Optional - lazy loaded from game phase
      predictionSystem?: IPredictionSystem  // Optional - lazy loaded from game phase
      interpolationSystem?: IInterpolationSystem  // Optional - lazy loaded from game phase
      reconciliationSystem?: IReconciliationSystem  // Optional - lazy loaded from game phase
      mainMenu?: IMainMenu  // Optional - created by orchestrator if not provided
      roomLobby?: IRoomLobby  // Optional - created by orchestrator if not provided
      container?: DIContainer  // Optional - for lazy resolution of game systems
    }
  ) {
    // All dependencies are injected and resolved by DI container
    // No creation logic here - just store references
    this.app = dependencies.app
    this.networkClient = dependencies.networkClient
    this.renderer = dependencies.renderer ?? null
    this.hud = dependencies.hud ?? null
    this.inputHandler = dependencies.inputHandler
    this.stateManager = dependencies.stateManager ?? null
    this.scene = dependencies.scene
    this.commandHistory = dependencies.commandHistory ?? null
    this.predictionSystem = dependencies.predictionSystem ?? null
    this.interpolationSystem = dependencies.interpolationSystem ?? null
    this.reconciliationSystem = dependencies.reconciliationSystem ?? null
    this.mainMenu = dependencies.mainMenu ?? null
    this.roomLobby = dependencies.roomLobby ?? null
    this.container = dependencies.container ?? null
  }

  /**
   * Initializes the app and all subsystems.
   * Initializes PixiJS App, creates Scene, and sets up subsystem event handlers.
   * 
   * @param container Optional container element for App initialization
   * @throws Error if initialization fails
   */
  async init(_container?: HTMLElement): Promise<void> {
    if (this.initialized) {
      // Already initialized, skip
      return
    }

    try {
      // App is already initialized in main.ts before orchestrator is created
      // Do NOT call app.init() here as it would destroy the existing app instance
      // and all its children (including Scene root)
      
      // Verify app is initialized (will throw if not)
      try {
        this.app.getApplication()
      } catch {
        throw new Error('App must be initialized before orchestrator.init()')
      }

      // Scene is always provided by DI container
      // Scene is initialized automatically via DI container's initializeAfter hook
      // No manual initialization needed here

      // Set up NetworkClient event handlers if provided
      if (this.networkClient && this.stateManager) {
        this.setupNetworkEventHandlers()
      }

      // Set up snapshot handler for game systems (if available)
      if (this.networkClient && this.stateManager && this.interpolationSystem && this.reconciliationSystem) {
        this.setupSnapshotHandler()
      }

      // Set up room management event handlers if provided
      if (this.networkClient) {
        this.setupRoomManagementHandlers()
      }

      // Create MainMenu and RoomLobby instances if not provided and we have a real Scene
      if (!this.mainMenu && this.scene && this.scene instanceof Scene) {
        const uiLayer = this.scene.getLayer('ui')
        this.mainMenu = new MainMenu(
          uiLayer,
          () => this.handleCreateRoom(),
          (code: string) => this.handleJoinRoom(code)
        )
      }

      if (!this.roomLobby && this.scene && this.scene instanceof Scene) {
        const uiLayer = this.scene.getLayer('ui')
        const initialRoomState: RoomState = {
          roomCode: '',
          players: [],
          state: 'lobby',
          hostId: 0
        }
        // Determine if current player is host (will be updated when room state is received)
        const isHost = false // Will be determined from room state
        this.roomLobby = new RoomLobby(
          uiLayer,
          initialRoomState,
          isHost,
          () => this.handleStartMatch(),
          () => this.handleLeaveRoom()
        )
      }

      // Show main menu on start and hide HUD
      if (this.mainMenu) {
        this.mainMenu.show()
      }
      if (this.hud) {
        this.hud.hide()
      }

      // Center UI layer (UI-specific logic moved from Scene)
      if (this.scene && this.scene instanceof Scene) {
        this.centerUILayer()
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
   * Sets up snapshot handler for game systems.
   * Handles server snapshots: updates authoritative state, adds to interpolation buffer, and reconciles.
   * Private helper method called during initialization.
   */
  private setupSnapshotHandler(): void {
    if (!this.networkClient || !this.stateManager || !this.interpolationSystem || !this.reconciliationSystem) {
      return
    }

    this.networkClient.onSnapshot((snapshot) => {
      // Update authoritative state from server
      this.stateManager!.updateAuthoritative(snapshot)
      
      // Add snapshot to interpolation buffer
      this.interpolationSystem!.addSnapshot(snapshot, performance.now())
      
      // Reconcile predicted state with authoritative
      this.reconciliationSystem!.reconcile(snapshot)
    })
  }

  /**
   * Handles create room action.
   * Creates a room via network client. The creator is automatically added as host.
   * Room state will be received via onRoomState callback, which will trigger transition to lobby.
   */
  private async handleCreateRoom(): Promise<void> {
    if (!this.networkClient) {
      console.error('Cannot create room: network client not available')
      return
    }

    try {
      await this.networkClient.createRoom()
      // Room state will be received via onRoomState callback, which will trigger transition
      // No need to join separately - server automatically adds creator to room
    } catch (error) {
      console.error('Failed to create room:', error)
    }
  }

  /**
   * Handles join room action.
   * Joins a room via network client and transitions to lobby.
   */
  private async handleJoinRoom(roomCode: string): Promise<void> {
    if (!this.networkClient) {
      console.error('Cannot join room: network client not available')
      return
    }

    try {
      await this.networkClient.joinRoom(roomCode)
      console.log('Joined room:', roomCode)
      // Room state will be received via onRoomState callback, which will trigger transition
    } catch (error) {
      console.error('Failed to join room:', error)
    }
  }

  /**
   * Handles start match action.
   * Starts the match via network client.
   */
  private handleStartMatch(): void {
    if (!this.networkClient) {
      console.error('Cannot start match: network client not available')
      return
    }

    try {
      this.networkClient.startMatch()
      console.log('Match start requested')
    } catch (error) {
      console.error('Failed to start match:', error)
    }
  }

  /**
   * Handles leave room action.
   * Leaves the room via network client and transitions to main menu.
   */
  private handleLeaveRoom(): void {
    if (!this.networkClient) {
      console.error('Cannot leave room: network client not available')
      return
    }

    try {
      this.networkClient.leaveRoom()
      console.log('Left room')
      this.transitionToMainMenu()
    } catch (error) {
      console.error('Failed to leave room:', error)
    }
  }

  /**
   * Sets up room management event handlers for NetworkClient.
   * Private helper method called during initialization.
   */
  private setupRoomManagementHandlers(): void {
    if (!this.networkClient) {
      return
    }

    this.networkClient.onRoomState((roomState) => {
      // Update room lobby with new state
      if (this.roomLobby) {
        // Update room lobby with new state
        this.roomLobby.update(roomState)
      }

      if (roomState.state === 'lobby') {
        this.transitionToLobby(roomState)
      }
    })

    this.networkClient.onPlayerJoined((player) => {
      // Player joined - lobby will be updated via onRoomState
      if (this.uiState === 'lobby') {
        console.log('Player joined:', player.name)
      }
    })

    this.networkClient.onPlayerLeft((playerId) => {
      // Player left - lobby will be updated via onRoomState
      if (this.uiState === 'lobby') {
        console.log('Player left:', playerId)
      }
    })

    this.networkClient.onMatchStarted(async () => {
      await this.transitionToInGame()
    })

    this.networkClient.onMatchEnded((winnerId) => {
      // Transition back to lobby after match ends
      this.transitionToLobby()
      if (winnerId !== undefined) {
        console.log('Match ended, winner:', winnerId)
      }
    })
  }

  /**
   * Gets the current UI state.
   * @returns Current UI state
   */
  getUIState(): UIState {
    return this.uiState
  }

  /**
   * Transitions to main menu state.
   * Shows main menu, hides lobby/HUD, disables input, stops game loop (keeps render loop running for UI).
   */
  transitionToMainMenu(): void {
    this.uiState = 'main-menu'

    if (this.mainMenu) {
      this.mainMenu.show()
    }
    if (this.roomLobby) {
      this.roomLobby.hide()
    }
    if (this.hud) {
      this.hud.hide()
    }

    // Disable input
    if (this.inputHandler) {
      this.inputHandler.detach()
    }

    // Stop all game systems (but keep render loop running for UI visibility)
    this.stopGameSystems()
  }

  /**
   * Transitions to lobby state.
   * Shows lobby, hides main menu/HUD, disables input, stops game loop (keeps render loop running for UI).
   * 
   * @param roomState Optional room state to update lobby with
   */
  transitionToLobby(roomState?: RoomState): void {
    this.uiState = 'lobby'

    if (this.mainMenu) {
      this.mainMenu.hide()
    }
    if (this.roomLobby) {
      this.roomLobby.show()
      if (roomState) {
        this.roomLobby.update(roomState)
      }
    }
    if (this.hud) {
      this.hud.hide()
    }

    // Disable input
    if (this.inputHandler) {
      this.inputHandler.detach()
    }

    // Stop all game systems (but keep render loop running for UI visibility)
    this.stopGameSystems()
  }

  /**
   * Lazy initialization of game systems.
   * Called when transitioning to in-game state.
   * Initializes game phase dependencies if not already initialized.
   */
  private async initializeGameSystems(): Promise<void> {
    if (this.gameSystemsInitialized || !this.container) {
      return
    }

    try {
      // Initialize game phase (StateManager, Renderer, HUD, etc.)
      await this.container.initializePhase('game')
      
      // Re-resolve dependencies that were just created
      if (this.container.isRegistered(DI_KEYS.STATE_MANAGER)) {
        this.stateManager = this.container.resolve<StateManager>(DI_KEYS.STATE_MANAGER)
      }
      if (this.container.isRegistered(DI_KEYS.RENDERER)) {
        this.renderer = this.container.resolve<Renderer>(DI_KEYS.RENDERER)
      }
      if (this.container.isRegistered(DI_KEYS.HUD)) {
        this.hud = this.container.resolve<HUD>(DI_KEYS.HUD)
      }
      if (this.container.isRegistered(DI_KEYS.COMMAND_HISTORY)) {
        this.commandHistory = this.container.resolve<CommandHistory>(DI_KEYS.COMMAND_HISTORY)
      }
      if (this.container.isRegistered(DI_KEYS.PREDICTION_SYSTEM)) {
        this.predictionSystem = this.container.resolve<PredictionSystem>(DI_KEYS.PREDICTION_SYSTEM)
      }
      if (this.container.isRegistered(DI_KEYS.INTERPOLATION_SYSTEM)) {
        this.interpolationSystem = this.container.resolve<InterpolationSystem>(DI_KEYS.INTERPOLATION_SYSTEM)
      }
      if (this.container.isRegistered(DI_KEYS.RECONCILIATION_SYSTEM)) {
        this.reconciliationSystem = this.container.resolve<ReconciliationSystem>(DI_KEYS.RECONCILIATION_SYSTEM)
      }

      this.gameSystemsInitialized = true
    } catch (error) {
      console.error('Failed to initialize game systems:', error)
      throw error
    }
  }

  /**
   * Transitions to in-game state.
   * Lazy-loads game systems if not already initialized.
   * Shows HUD, hides main menu/lobby, enables input, starts rendering and game loops.
   */
  async transitionToInGame(): Promise<void> {
    // Lazy initialize game systems
    await this.initializeGameSystems()

    this.uiState = 'in-game'

    if (this.mainMenu) {
      this.mainMenu.hide()
    }
    if (this.roomLobby) {
      this.roomLobby.hide()
    }
    if (this.hud) {
      this.hud.show()
    }

    // Enable input
    if (this.inputHandler) {
      this.inputHandler.attach()
    }

    // Start all game systems (render loop, game loop, input loop, interpolation loop)
    if (this.initialized) {
      this.startGameSystems()
    }
  }

  /**
   * Starts the render loop (always needed for UI visibility).
   * This is called on initial startup to show menus.
   * 
   * @throws Error if not initialized
   */
  start(): void {
    if (!this.initialized) {
      throw new Error('AppOrchestrator must be initialized before starting')
    }

    // Always start render loop (needed for UI visibility in all states)
    // RenderLoop handles its own running state internally
    if (!this.renderLoop) {
      this.renderLoop = new RenderLoop(this.app)
    }
    this.renderLoop.start()

    // Set up beforeunload handler
    this.setupBeforeUnloadHandler()
  }

  /**
   * Starts all game systems: game loop, input loop, and interpolation loop.
   * Only runs when in 'in-game' state.
   * Called automatically when transitioning to in-game state.
   */
  startGameSystems(): void {
    if (!this.initialized) {
      throw new Error('AppOrchestrator must be initialized before starting game systems')
    }

    // Start render loop if not already started
    if (!this.renderLoop) {
      this.renderLoop = new RenderLoop(this.app)
      this.renderLoop.start()
    }

    // Only start game systems if in 'in-game' state
    if (this.uiState !== 'in-game') {
      return
    }

    // Start game loop (updates renderer and HUD)
    this.startGameLoop()

    // Start input loop (sends input commands to server)
    this.startInputLoop()

    // Start interpolation loop (smooths between snapshots)
    this.startInterpolationLoop()
  }

  /**
   * Starts the game loop that updates renderer and HUD.
   * Only runs when in 'in-game' state.
   */
  private startGameLoop(): void {
    if (this.isGameLoopRunning) {
      return // Already running
    }

    this.isGameLoopRunning = true
    const gameLoop = () => {
      if (!this.isGameLoopRunning || this.uiState !== 'in-game') {
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
   * Starts the input loop that sends input commands to the server at 30Hz.
   * Only processes input when in 'in-game' state.
   */
  private startInputLoop(): void {
    if (this.isInputLoopRunning) {
      return // Already running
    }

    if (!this.networkClient || !this.inputHandler || !this.commandHistory || !this.predictionSystem) {
      console.warn('Cannot start input loop: required dependencies not available')
      return
    }

    this.isInputLoopRunning = true
    this.lastInputSendTime = 0
    this.commandSequence = 0

    const inputLoop = () => {
      if (!this.isInputLoopRunning) {
        return
      }

      const now = performance.now()

      // Only process input when in-game
      if (this.uiState === 'in-game') {
        // Send input commands at regular intervals
        if (now - this.lastInputSendTime >= this.inputSendIntervalMs) {
          const thrust = this.inputHandler!.getThrust()
          const turn = this.inputHandler!.getTurn()

          // Only send if there's actual input
          if (thrust > 0 || turn !== 0) {
            this.commandSequence++
            this.commandHistory!.addCommand(this.commandSequence, thrust, turn)
            this.networkClient!.sendInput(this.commandSequence, thrust, turn)
            
            // Immediately predict locally for responsive feel
            this.predictionSystem!.predict({ thrust, turn })
          }

          this.lastInputSendTime = now
        }
      }

      this.inputLoopId = requestAnimationFrame(inputLoop)
    }

    this.inputLoopId = requestAnimationFrame(inputLoop)
  }

  /**
   * Starts the interpolation loop that smooths between server snapshots.
   * Only interpolates when in 'in-game' state.
   */
  private startInterpolationLoop(): void {
    if (this.isInterpolationLoopRunning) {
      return // Already running
    }

    if (!this.interpolationSystem) {
      console.warn('Cannot start interpolation loop: interpolation system not available')
      return
    }

    this.isInterpolationLoopRunning = true

    const interpolationLoop = () => {
      if (!this.isInterpolationLoopRunning) {
        return
      }

      // Only interpolate when in-game
      if (this.uiState === 'in-game') {
        this.interpolationSystem!.update(performance.now())
      }

      this.interpolationLoopId = requestAnimationFrame(interpolationLoop)
    }

    this.interpolationLoopId = requestAnimationFrame(interpolationLoop)
  }

  /**
   * Centers the UI layer on screen (UI-specific logic moved from Scene).
   * Private helper method called during initialization.
   * Scene must be initialized before calling this.
   */
  private centerUILayer(): void {
    if (!this.scene || !(this.scene instanceof Scene)) {
      return
    }

    // Scene should be initialized at this point (called after scene.initialize())
    if (!this.scene.isInitialized()) {
      console.warn('Scene not initialized, skipping UI layer centering')
      return
    }

    try {
      const pixiApp = this.app.getApplication()
      
      // Verify scene state before getting layer
      if (!this.scene.isInitialized()) {
        throw new Error('Scene not initialized when trying to center UI layer')
      }
      
      // Get UI layer - this should never return null (throws if missing)
      let uiLayer: Container | null = null
      try {
        uiLayer = this.scene.getLayer('ui')
      } catch (error) {
        throw new Error(
          `Failed to get UI layer: ${error instanceof Error ? error.message : String(error)}. ` +
          `Scene initialized: ${this.scene.isInitialized()}`
        )
      }
      
      // Defensive check: ensure uiLayer is valid (should never be null due to getLayer() contract)
      if (!uiLayer) {
        throw new Error(
          'UI layer is null after getLayer() call. ' +
          `Scene initialized: ${this.scene.isInitialized()}`
        )
      }
      
      // Verify layer is not destroyed
      if (uiLayer.destroyed) {
        throw new Error('UI layer is already destroyed')
      }
      
      const centerUI = () => {
        // Re-check uiLayer in case it was destroyed (defensive)
        if (!uiLayer || uiLayer.destroyed) {
          console.warn('UI layer destroyed, skipping centering')
          return
        }
        try {
          const screen = pixiApp.screen
          uiLayer.x = screen.width / 2
          uiLayer.y = screen.height / 2
        } catch (error) {
          console.error('Error centering UI layer:', error)
        }
      }
      
      // Center initially
      centerUI()
      
      // Re-center on resize (only if layer is valid)
      if (uiLayer && !uiLayer.destroyed) {
        pixiApp.renderer.on('resize', centerUI)
      }
    } catch (error) {
      // This should not happen if initialization order is correct
      console.error('Failed to center UI layer:', error)
      throw error // Re-throw to surface the problem
    }
  }

  /**
   * Sets up beforeunload event handler.
   * Private helper method called during start().
   */
  private setupBeforeUnloadHandler(): void {
    if (this.beforeunloadHandler) {
      return // Already set up
    }

    this.beforeunloadHandler = () => {
      this.destroy()
    }

    window.addEventListener('beforeunload', this.beforeunloadHandler)
  }

  /**
   * Stops all game systems: game loop, input loop, and interpolation loop.
   * Keeps render loop running for UI visibility.
   * Used when transitioning to main-menu or lobby states.
   */
  private stopGameSystems(): void {
    this.stopGameLoop()
    this.stopInputLoop()
    this.stopInterpolationLoop()
  }

  /**
   * Stops only the game loop (keeps render loop running for UI visibility).
   */
  private stopGameLoop(): void {
    this.isGameLoopRunning = false
    if (this.gameLoopId !== null) {
      cancelAnimationFrame(this.gameLoopId)
      this.gameLoopId = null
    }
  }

  /**
   * Stops the input loop.
   */
  private stopInputLoop(): void {
    this.isInputLoopRunning = false
    if (this.inputLoopId !== null) {
      cancelAnimationFrame(this.inputLoopId)
      this.inputLoopId = null
    }
  }

  /**
   * Stops the interpolation loop.
   */
  private stopInterpolationLoop(): void {
    this.isInterpolationLoopRunning = false
    if (this.interpolationLoopId !== null) {
      cancelAnimationFrame(this.interpolationLoopId)
      this.interpolationLoopId = null
    }
  }

  /**
   * Stops the render loop and all game systems.
   * Also detaches input handler and disconnects network client.
   */
  stop(): void {
    // Stop all game systems
    this.stopGameSystems()

    // Stop RenderLoop
    if (this.renderLoop) {
      this.renderLoop.stop()
    }

    // Detach input handler
    if (this.inputHandler) {
      this.inputHandler.detach()
    }

    // Disconnect network client
    if (this.networkClient) {
      this.networkClient.disconnect()
    }
  }

  /**
   * Destroys all subsystems and cleans up resources.
   * Destroys subsystems in reverse dependency order.
   */
  destroy(): void {
    // Stop loops and detach/disconnect
    this.stop()

    // Destroy subsystems in reverse dependency order
    // Input handler (already detached in stop())
    if (this.inputHandler) {
      this.inputHandler.reset()
      this.inputHandler = null
    }

    // HUD
    if (this.hud) {
      this.hud.destroy()
      this.hud = null
    }

    // Renderer
    if (this.renderer) {
      this.renderer.destroy()
      this.renderer = null
    }

    // Network client (already disconnected in stop())
    this.networkClient = null

    // State manager (no destroy method, just clear reference)
    this.stateManager = null

    // Main menu
    if (this.mainMenu) {
      this.mainMenu.destroy()
      this.mainMenu = null
    }

    // Room lobby
    if (this.roomLobby) {
      this.roomLobby.destroy()
      this.roomLobby = null
    }

    // Scene is owned by orchestrator (from DI container)
    // Destroy is handled by orchestrator's lifecycle
    // IMPORTANT: Destroy Scene BEFORE App, as Scene needs App to remove itself from stage
    if (this.scene) {
      try {
        this.scene.destroy()
      } catch {
        // Ignore errors - scene may already be destroyed or app not initialized
      }
      this.scene = null
    }

    // RenderLoop - already stopped by this.stop() above, just clear reference
    if (this.renderLoop) {
      this.renderLoop = null
    }

    // App is owned by 'application' (not orchestrator), so do NOT destroy it here
    // App should be destroyed in main.ts by the application owner
    // Removing this line to fix ownership violation:
    // this.app.destroy()

    // Remove beforeunload listener
    if (this.beforeunloadHandler) {
      window.removeEventListener('beforeunload', this.beforeunloadHandler)
      this.beforeunloadHandler = null
    }

    // Reset state
    this.initialized = false
    this.initError = null
    this.uiState = 'main-menu'
  }
}

