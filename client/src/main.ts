/**
 * Orbital Rush Client - Main entry point
 * Initializes all systems and wires them together for full game functionality.
 */

import { App } from './core/app'
import { Scene } from './gfx/scene'
import { RenderLoop } from './core/render-loop'
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
  await app.init()

  // Create scene hierarchy
  const scene = new Scene(app)

  // Initialize all systems in dependency order
  const stateManager = new StateManager()
  const localSimulator = new LocalSimulator()
  const commandHistory = new CommandHistory()
  const predictionSystem = new PredictionSystem(stateManager, localSimulator, commandHistory)
  const reconciliationSystem = new ReconciliationSystem(stateManager, localSimulator, commandHistory, predictionSystem)
  const interpolationSystem = new InterpolationSystem(stateManager)
  const renderer = new Renderer(stateManager, scene, app)
  const networkClient = new NetworkClient()
  const keyboardInput = new KeyboardInputHandler()
  const hud = new HUD(scene, stateManager)

  // Set up WebSocket connection and snapshot handling
  networkClient.onSnapshot((snapshot: SnapshotMessage) => {
    // Update authoritative state from server
    stateManager.updateAuthoritative(snapshot)
    
    // Add snapshot to interpolation buffer
    interpolationSystem.addSnapshot(snapshot, performance.now())
    
    // Reconcile predicted state with authoritative
    reconciliationSystem.reconcile(snapshot)
  })

  networkClient.onConnect(() => {
    console.log('Connected to game server')
  })

  networkClient.onDisconnect(() => {
    console.log('Disconnected from game server')
  })

  networkClient.onError((error) => {
    console.error('Network error:', error)
  })

  // Connect to server
  try {
    await networkClient.connect(WS_URL)
  } catch (error) {
    console.error('Failed to connect to server:', error)
    // Continue anyway - might be testing without server
  }

  // Set up keyboard input
  let lastInputSendTime = 0
  let commandSequence = 0

  window.addEventListener('keydown', (event) => {
    keyboardInput.onKeyDown(event.key)
  })

  window.addEventListener('keyup', (event) => {
    keyboardInput.onKeyUp(event.key)
  })

  // Start render loop
  const renderLoop = new RenderLoop(app)
  
  // Game loop that runs alongside render loop
  const gameLoop = () => {
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

    // Update interpolation system
    interpolationSystem.update(now)

    // Update renderer with current state
    renderer.update()

    // Update HUD
    hud.update()

    // Continue game loop
    requestAnimationFrame(gameLoop)
  }

  // Start both loops
  renderLoop.start()
  gameLoop()

  // Cleanup on page unload
  window.addEventListener('beforeunload', () => {
    renderLoop.stop()
    networkClient.disconnect()
    renderer.destroy()
    hud.destroy()
    scene.destroy()
    app.destroy()
  })
}

init().catch(console.error)
