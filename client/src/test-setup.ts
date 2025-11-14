/**
 * Test setup file to provide WebGL context for Pixi.js in headless environment
 * 
 * Note: Full Pixi.js renderer initialization requires WebGL context.
 * We use vitest-webgl-canvas-mock for proper WebGL context mocking,
 * with fallback to gl package for more complete WebGL support.
 */

// Import vitest-webgl-canvas-mock for basic canvas/WebGL mocking
// This must be imported before any Pixi.js code runs
import 'vitest-webgl-canvas-mock'

// Provide enhanced WebGL context using gl package if available
if (typeof window !== 'undefined') {
  const originalGetContext = HTMLCanvasElement.prototype.getContext
  
  HTMLCanvasElement.prototype.getContext = function (contextType: string, options?: any) {
    if (contextType === 'webgl' || contextType === 'webgl2') {
      // Try to use real gl package for more complete WebGL support
      try {
        // eslint-disable-next-line @typescript-eslint/no-require-imports
        const gl = require('gl')
        const width = this.width || 800
        const height = this.height || 600
        const context = gl(width, height, { 
          preserveDrawingBuffer: true,
          stencil: true // Required by Pixi.js isWebGLSupported check
        })
        
        // Ensure canvas property is set (required by Pixi.js)
        if (!context.canvas) {
          context.canvas = this
        }
        
        // Add getContextAttributes method (required by Pixi.js isWebGLSupported check)
        if (!context.getContextAttributes) {
          context.getContextAttributes = () => ({
            stencil: true,
            preserveDrawingBuffer: true,
            antialias: false,
            depth: true,
            failIfMajorPerformanceCaveat: false
          })
        }
        
        // Add WEBGL_lose_context extension (checked by Pixi.js)
        // Wrap existing getExtension to add WEBGL_lose_context support
        const originalGetExtension = context.getExtension.bind(context)
        context.getExtension = (name: string) => {
          if (name === 'WEBGL_lose_context') {
            return {
              loseContext: () => {},
              restoreContext: () => {}
            }
          }
          // Call the original getExtension from gl package
          try {
            return originalGetExtension(name)
          } catch {
            return null
          }
        }
        
        // Set canvas dimensions if not already set
        if (!this.width) {
          this.width = width
        }
        if (!this.height) {
          this.height = height
        }
        
        return context as any
      } catch {
        // Fallback to default getContext (from vitest-webgl-canvas-mock or browser)
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        return (originalGetContext as any).call(this, contextType, options)
      }
    }
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return (originalGetContext as any).call(this, contextType, options)
  }
}

