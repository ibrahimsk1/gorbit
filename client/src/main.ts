/**
 * Orbital Rush Client - Main entry point
 * Initializes Pixi Application, Scene, and Render Loop
 * 
 * Note: Full game integration (WebSocket, input, prediction, reconciliation, interpolation, renderer)
 * will be added in future CUs. This is the minimal initialization.
 */

import { App } from './core/app'
import { Scene } from './gfx/scene'
import { RenderLoop } from './core/render-loop'

async function init() {
  // Initialize Pixi Application
  const app = new App()
  await app.init()

  // Create scene hierarchy
  const scene = new Scene(app)

  // Start render loop
  const renderLoop = new RenderLoop(app)
  renderLoop.start()

  // Cleanup on page unload
  window.addEventListener('beforeunload', () => {
    renderLoop.stop()
    scene.destroy()
    app.destroy()
  })
}

init().catch(console.error)
