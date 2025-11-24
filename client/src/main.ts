/**
 * Orbital Rush Client - Main entry point
 * Initializes all systems and wires them together for full game functionality.
 * 
 * Labels: scope:integration loop:g2-app layer:core
 */

import { App } from './core/app'
import { AppOrchestrator } from './core/orchestrator'
import { Scene } from './gfx/scene'
import { Renderer } from './gfx/renderer'
import { StateManager } from './sim/state-manager'
import { LocalSimulator } from './sim/local-simulator'
import { CommandHistory } from './net/command-history'
import { PredictionSystem } from './sim/prediction'
import { ReconciliationSystem } from './sim/reconciliation'
import { InterpolationSystem } from './sim/interpolation'
import { NetworkClient } from './net/client'
import { KeyboardInputHandler } from './input/keyboard'
import { HUD } from './ui/hud'
import type { SnapshotMessage } from './net/protocol'

// Configuration
const WS_URL = 'ws://localhost:8080/ws'
const INPUT_SEND_INTERVAL_MS = 1000 / 30 // 30Hz input rate (matches server tick rate)

async function init() {
  // Initialize Pixi Application
  const app = new App()

  // Initialize simulation systems (not part of orchestrator's interface)
  const stateManager = new StateManager()
  const localSimulator = new LocalSimulator()
  const commandHistory = new CommandHistory()
  const predictionSystem = new PredictionSystem(stateManager, localSimulator, commandHistory)
  const reconciliationSystem = new ReconciliationSystem(stateManager, localSimulator, commandHistory, predictionSystem)
  const interpolationSystem = new InterpolationSystem(stateManager)

  // Create subsystems
  const networkClient = new NetworkClient()
  const keyboardInput = new KeyboardInputHandler()
  
  // Scene needs to be created before Renderer (Renderer depends on Scene)
  const scene = new Scene(app)
  const renderer = new Renderer(stateManager, scene, app)
  const hud = new HUD(scene, stateManager)

  // Set up custom snapshot handler (uses simulation systems not in orchestrator)
  networkClient.onSnapshot((snapshot: SnapshotMessage) => {
    // Update authoritative state from server
    stateManager.updateAuthoritative(snapshot)
    
    // Add snapshot to interpolation buffer
    interpolationSystem.addSnapshot(snapshot, performance.now())
    
    // Reconcile predicted state with authoritative
    reconciliationSystem.reconcile(snapshot)
  })

  // Create orchestrator with all subsystems
  const orchestrator = new AppOrchestrator(app, {
    networkClient,
    renderer,
    hud,
    inputHandler: keyboardInput,
    stateManager,
    scene
  })

  // Initialize orchestrator (initializes App, sets up event handlers)
  await orchestrator.init()

  // Connect to server
  try {
    await networkClient.connect(WS_URL)
  } catch (error) {
    console.error('Failed to connect to server:', error)
    // Continue anyway - might be testing without server
  }

  // Start orchestrator (starts render loop and game loop)
  orchestrator.start()

  // Set up input sending loop (custom logic, runs in parallel with game loop)
  let lastInputSendTime = 0
  let commandSequence = 0
  const inputLoop = () => {
    const now = performance.now()

    // Send input commands at regular intervals
    if (now - lastInputSendTime >= INPUT_SEND_INTERVAL_MS) {
      const thrust = keyboardInput.getThrust()
      const turn = keyboardInput.getTurn()

      // Only send if there's actual input
      if (thrust > 0 || turn !== 0) {
        commandSequence++
        commandHistory.addCommand(commandSequence, thrust, turn)
        networkClient.sendInput(commandSequence, thrust, turn)
        
        // Immediately predict locally for responsive feel
        predictionSystem.predict({ thrust, turn })
      }

      lastInputSendTime = now
    }

    requestAnimationFrame(inputLoop)
  }
  inputLoop()

  // Set up interpolation update loop (runs in parallel with game loop)
  const interpolationLoop = () => {
    interpolationSystem.update(performance.now())
    requestAnimationFrame(interpolationLoop)
  }
  interpolationLoop()

  // Cleanup is handled by orchestrator's destroy() on beforeunload
  // Orchestrator will clean up: render loop, input handler, network client, renderer, HUD, scene, app
}

init().catch(console.error)
