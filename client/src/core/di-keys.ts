/**
 * Dependency injection keys for all registered dependencies.
 * 
 * These keys are used to register and resolve dependencies in the DI container.
 * Using constants prevents typos and provides autocomplete support.
 * 
 * Labels: scope:unit loop:g2-app layer:core
 */

export const DI_KEYS = {
  // Core
  APP: 'app',
  SCENE: 'scene',
  
  // Simulation
  STATE_MANAGER: 'stateManager',
  LOCAL_SIMULATOR: 'localSimulator',
  COMMAND_HISTORY: 'commandHistory',
  PREDICTION_SYSTEM: 'predictionSystem',
  RECONCILIATION_SYSTEM: 'reconciliationSystem',
  INTERPOLATION_SYSTEM: 'interpolationSystem',
  
  // Network
  NETWORK_CLIENT: 'networkClient',
  
  // Input
  INPUT_HANDLER: 'inputHandler',
  
  // Graphics
  RENDERER: 'renderer',
  
  // UI
  HUD: 'hud',
  MAIN_MENU: 'mainMenu',
  ROOM_LOBBY: 'roomLobby',
  
  // Orchestration
  ORCHESTRATOR: 'orchestrator',
} as const

/**
 * Type for dependency keys (for type-safe key usage).
 */
export type DependencyKey = typeof DI_KEYS[keyof typeof DI_KEYS]

