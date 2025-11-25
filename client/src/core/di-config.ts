/**
 * Dependency Injection Configuration
 * 
 * Registers all application dependencies in the DI container.
 * This centralizes all dependency wiring in one place.
 * 
 * Labels: scope:integration loop:g2-app layer:core
 */

import { DIContainer } from './di-container'
import { DI_KEYS } from './di-keys'
import { App } from './app'
import { Scene } from '../gfx/scene'
import { Renderer } from '../gfx/renderer'
import { StateManager } from '../sim/state-manager'
import { LocalSimulator } from '../sim/local-simulator'
import { CommandHistory } from '../net/command-history'
import { PredictionSystem } from '../sim/prediction'
import { ReconciliationSystem } from '../sim/reconciliation'
import { InterpolationSystem } from '../sim/interpolation'
import { NetworkClient } from '../net/client'
import { KeyboardInputHandler } from '../input/keyboard'
import { HUD } from '../ui/hud'
import { MainMenu } from '../ui/main-menu'
import { RoomLobby } from '../ui/room-lobby'
import { AppOrchestrator } from './orchestrator'

/**
 * Registers all dependencies in the DI container.
 * 
 * @param container DI container to register dependencies in
 */
export function registerDependencies(container: DIContainer): void {
  // ============================================
  // BOOTSTRAP PHASE: Core infrastructure
  // ============================================
  
  // Register App (bootstrap phase, application-owned)
  container.register(DI_KEYS.APP, () => new App(), {
    phase: 'bootstrap',
    scope: 'singleton',
    owner: 'application'
  })

  // Register network client (bootstrap phase, container-owned)
  container.register(DI_KEYS.NETWORK_CLIENT, () => new NetworkClient(), {
    phase: 'bootstrap',
    scope: 'singleton',
    owner: 'container'
  })

  // Register input handler (bootstrap phase, container-owned)
  container.register(DI_KEYS.INPUT_HANDLER, () => new KeyboardInputHandler(), {
    phase: 'bootstrap',
    scope: 'singleton',
    owner: 'container'
  })

  // ============================================
  // MENU PHASE: Systems needed for menu UI
  // ============================================
  
  // Register Scene (menu phase, orchestrator-owned, depends on App)
  // Scene initialization happens automatically via initializeAfter hook
  container.register(DI_KEYS.SCENE, () => {
    const app = container.resolve<App>(DI_KEYS.APP)
    return new Scene(app)
  }, {
    phase: 'menu',
    scope: 'singleton',
    owner: 'orchestrator',
    dependencies: [DI_KEYS.APP],
    // Scene initialization hook - called automatically after App.init()
    initializeAfter: async (scene: Scene) => {
      // Ensure App is initialized first (should be done in main.ts before menu phase)
      const app = container.resolve<App>(DI_KEYS.APP)
      try {
        app.getApplication() // Throws if not initialized
        scene.initialize()
      } catch (error) {
        throw new Error(
          `Scene initialization failed: App must be initialized before Scene. ` +
          `Ensure app.init() is called before initializePhase('menu').`
        )
      }
    }
  })

  // Register orchestrator (menu phase, depends on bootstrap systems and Scene)
  // Game systems (renderer, hud, stateManager) are optional and resolved lazily
  container.register(DI_KEYS.ORCHESTRATOR, () => {
    const app = container.resolve<App>(DI_KEYS.APP)
    const networkClient = container.resolve<NetworkClient>(DI_KEYS.NETWORK_CLIENT)
    const inputHandler = container.resolve<KeyboardInputHandler>(DI_KEYS.INPUT_HANDLER)
    const scene = container.resolve<Scene>(DI_KEYS.SCENE)
    
    // Game systems are optional - will be resolved lazily when game phase initializes
    // Try to resolve them, but they may not exist yet (will be null if not registered)
    let renderer: Renderer | null = null
    let hud: HUD | null = null
    let stateManager: StateManager | null = null
    let commandHistory: CommandHistory | null = null
    let predictionSystem: PredictionSystem | null = null
    let interpolationSystem: InterpolationSystem | null = null
    let reconciliationSystem: ReconciliationSystem | null = null
    
    try {
      if (container.isRegistered(DI_KEYS.RENDERER)) {
        renderer = container.resolve<Renderer>(DI_KEYS.RENDERER)
      }
    } catch {}
    
    try {
      if (container.isRegistered(DI_KEYS.HUD)) {
        hud = container.resolve<HUD>(DI_KEYS.HUD)
      }
    } catch {}
    
    try {
      if (container.isRegistered(DI_KEYS.STATE_MANAGER)) {
        stateManager = container.resolve<StateManager>(DI_KEYS.STATE_MANAGER)
      }
    } catch {}
    
    try {
      if (container.isRegistered(DI_KEYS.COMMAND_HISTORY)) {
        commandHistory = container.resolve<CommandHistory>(DI_KEYS.COMMAND_HISTORY)
      }
    } catch {}
    
    try {
      if (container.isRegistered(DI_KEYS.PREDICTION_SYSTEM)) {
        predictionSystem = container.resolve<PredictionSystem>(DI_KEYS.PREDICTION_SYSTEM)
      }
    } catch {}
    
    try {
      if (container.isRegistered(DI_KEYS.INTERPOLATION_SYSTEM)) {
        interpolationSystem = container.resolve<InterpolationSystem>(DI_KEYS.INTERPOLATION_SYSTEM)
      }
    } catch {}
    
    try {
      if (container.isRegistered(DI_KEYS.RECONCILIATION_SYSTEM)) {
        reconciliationSystem = container.resolve<ReconciliationSystem>(DI_KEYS.RECONCILIATION_SYSTEM)
      }
    } catch {}
    
    return new AppOrchestrator({
      app,
      networkClient,
      renderer: renderer ?? undefined,
      hud: hud ?? undefined,
      inputHandler,
      stateManager: stateManager ?? undefined,
      scene,
      commandHistory: commandHistory ?? undefined,
      predictionSystem: predictionSystem ?? undefined,
      interpolationSystem: interpolationSystem ?? undefined,
      reconciliationSystem: reconciliationSystem ?? undefined,
      // Pass container for lazy resolution
      container
    })
  }, {
    phase: 'menu',
    scope: 'singleton',
    owner: 'application',
    dependencies: [
      DI_KEYS.APP,
      DI_KEYS.NETWORK_CLIENT,
      DI_KEYS.INPUT_HANDLER,
      DI_KEYS.SCENE
    ]
  })

  // ============================================
  // GAME PHASE: Systems only needed during gameplay (lazy loaded)
  // ============================================
  
  // Register simulation systems (game phase, container-owned)
  container.register(DI_KEYS.STATE_MANAGER, () => new StateManager(), {
    phase: 'game',
    scope: 'singleton',
    owner: 'container'
  })

  container.register(DI_KEYS.LOCAL_SIMULATOR, () => new LocalSimulator(), {
    phase: 'game',
    scope: 'singleton',
    owner: 'container'
  })

  container.register(DI_KEYS.COMMAND_HISTORY, () => new CommandHistory(), {
    phase: 'game',
    scope: 'singleton',
    owner: 'container'
  })

  container.register(DI_KEYS.PREDICTION_SYSTEM, () => {
    const stateManager = container.resolve<StateManager>(DI_KEYS.STATE_MANAGER)
    const localSimulator = container.resolve<LocalSimulator>(DI_KEYS.LOCAL_SIMULATOR)
    const commandHistory = container.resolve<CommandHistory>(DI_KEYS.COMMAND_HISTORY)
    return new PredictionSystem(stateManager, localSimulator, commandHistory)
  }, {
    phase: 'game',
    scope: 'singleton',
    owner: 'container',
    dependencies: [DI_KEYS.STATE_MANAGER, DI_KEYS.LOCAL_SIMULATOR, DI_KEYS.COMMAND_HISTORY]
  })

  container.register(DI_KEYS.RECONCILIATION_SYSTEM, () => {
    const stateManager = container.resolve<StateManager>(DI_KEYS.STATE_MANAGER)
    const localSimulator = container.resolve<LocalSimulator>(DI_KEYS.LOCAL_SIMULATOR)
    const commandHistory = container.resolve<CommandHistory>(DI_KEYS.COMMAND_HISTORY)
    const predictionSystem = container.resolve<PredictionSystem>(DI_KEYS.PREDICTION_SYSTEM)
    return new ReconciliationSystem(stateManager, localSimulator, commandHistory, predictionSystem)
  }, {
    phase: 'game',
    scope: 'singleton',
    owner: 'container',
    dependencies: [DI_KEYS.STATE_MANAGER, DI_KEYS.LOCAL_SIMULATOR, DI_KEYS.COMMAND_HISTORY, DI_KEYS.PREDICTION_SYSTEM]
  })

  container.register(DI_KEYS.INTERPOLATION_SYSTEM, () => {
    const stateManager = container.resolve<StateManager>(DI_KEYS.STATE_MANAGER)
    return new InterpolationSystem(stateManager)
  }, {
    phase: 'game',
    scope: 'singleton',
    owner: 'container',
    dependencies: [DI_KEYS.STATE_MANAGER]
  })

  // Register Renderer (game phase, needs Scene and StateManager)
  container.register(DI_KEYS.RENDERER, () => {
    const stateManager = container.resolve<StateManager>(DI_KEYS.STATE_MANAGER)
    const scene = container.resolve<Scene>(DI_KEYS.SCENE)
    const app = container.resolve<App>(DI_KEYS.APP)
    return new Renderer(stateManager, scene, app)
  }, {
    phase: 'game',
    scope: 'singleton',
    owner: 'orchestrator',
    dependencies: [DI_KEYS.STATE_MANAGER, DI_KEYS.SCENE, DI_KEYS.APP]
  })

  // Register HUD (game phase, needs Scene and StateManager)
  container.register(DI_KEYS.HUD, () => {
    const scene = container.resolve<Scene>(DI_KEYS.SCENE)
    const stateManager = container.resolve<StateManager>(DI_KEYS.STATE_MANAGER)
    return new HUD(scene, stateManager)
  }, {
    phase: 'game',
    scope: 'singleton',
    owner: 'orchestrator',
    dependencies: [DI_KEYS.SCENE, DI_KEYS.STATE_MANAGER]
  })

  // MainMenu and RoomLobby are created by orchestrator during init()
  // They need callbacks that orchestrator provides, so we can't register them here
  // They will be created in orchestrator.init() when needed
}
