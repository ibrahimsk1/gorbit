/**
 * Test setup file to provide WebGL context for Pixi.js in headless environment
 * 
 * Note: Full Pixi.js renderer initialization requires WebGL context.
 * In headless environments, we provide a minimal WebGL mock for structure tests.
 * Full integration tests should run in a browser environment.
 */

// Provide WebGL context mock for headless testing
if (typeof window !== 'undefined') {
  const originalGetContext = HTMLCanvasElement.prototype.getContext
  
  HTMLCanvasElement.prototype.getContext = function (contextType: string, options?: any) {
    if (contextType === 'webgl' || contextType === 'webgl2') {
      // Try to use real gl package if available
      try {
        // eslint-disable-next-line @typescript-eslint/no-require-imports
        const gl = require('gl')
        const width = this.width || 800
        const height = this.height || 600
        return gl(width, height, { preserveDrawingBuffer: true }) as any
      } catch {
        // Fallback: minimal WebGL mock
        return {
          getParameter: () => 0,
          getExtension: () => null,
          canvas: this,
          viewport: () => {},
          clear: () => {},
          clearColor: () => {},
        } as any
      }
    }
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return (originalGetContext as any).call(this, contextType, options)
  }
}

