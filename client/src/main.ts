/**
 * Orbital Rush Client - Main entry point
 * Initializes all systems and wires them together for full game functionality.
 * 
 * Labels: scope:integration loop:g2-app layer:core
 */

import { DIContainer } from './core/di-container'
import { DI_KEYS } from './core/di-keys'
import { registerDependencies } from './core/di-config'
import { App } from './core/app'
import { AppOrchestrator } from './core/orchestrator'
import { NetworkClient } from './net/client'

// Configuration
const WS_URL = 'ws://localhost:8080/ws'

async function init() {
  // Create DI container
  const container = new DIContainer()

  // Register all dependencies
  registerDependencies(container)

  // BOOTSTRAP PHASE: Initialize core infrastructure
  await container.initializePhase('bootstrap')

  // Initialize App (must be done before menu phase)
  const app = container.resolve<App>(DI_KEYS.APP)
  await app.init()

  // MENU PHASE: Initialize menu systems (Scene auto-initializes via hook)
  await container.initializePhase('menu')

  // Get orchestrator (menu systems ready, game systems will be lazy loaded)
  const orchestrator = container.resolve<AppOrchestrator>(DI_KEYS.ORCHESTRATOR)
  await orchestrator.init()

  // Get network client for connection
  const networkClient = container.resolve<NetworkClient>(DI_KEYS.NETWORK_CLIENT)

  // Connect to server
  try {
    await networkClient.connect(WS_URL)
  } catch (error) {
    console.error('Failed to connect to server:', error)
    // Continue anyway - might be testing without server
  }

  // Start orchestrator (starts render loop for UI, game loops start when match begins)
  orchestrator.start()

  // GAME PHASE: Systems initialized lazily when match starts
  // (handled in orchestrator.transitionToInGame())

  // Cleanup on exit
  window.addEventListener('beforeunload', () => {
    orchestrator.destroy()
    
    // Destroy App (application-owned, not container-owned)
    // App must be destroyed after orchestrator (which uses it) but before container
    const app = container.resolve<App>(DI_KEYS.APP)
    app.destroy()
    
    container.destroy()
  })
}

init().catch(console.error)
