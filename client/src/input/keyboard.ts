/**
 * Keyboard input handler for game controls.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

/**
 * Keyboard input handler that tracks keyboard state and converts key presses
 * to game input actions (thrust and turn).
 * 
 * Key mappings:
 * - Thrust: ArrowUp or W key
 * - Turn left: ArrowLeft or A key
 * - Turn right: ArrowRight or D key
 */
export class KeyboardInputHandler {
  private thrustActive: boolean = false
  private turnLeftActive: boolean = false
  private turnRightActive: boolean = false
  private keydownListener: ((event: KeyboardEvent) => void) | null = null
  private keyupListener: ((event: KeyboardEvent) => void) | null = null

  /**
   * Gets the current thrust value [0.0, 1.0].
   * Returns 1.0 if thrust key is pressed, 0.0 otherwise.
   */
  getThrust(): number {
    return this.thrustActive ? 1.0 : 0.0
  }

  /**
   * Gets the current turn value [-1.0, 1.0].
   * Returns -1.0 if turn left key is pressed, 1.0 if turn right key is pressed,
   * 0.0 if neither or both are pressed.
   */
  getTurn(): number {
    if (this.turnLeftActive && this.turnRightActive) {
      return 0.0 // Both pressed = cancel out
    }
    if (this.turnLeftActive) {
      return -1.0
    }
    if (this.turnRightActive) {
      return 1.0
    }
    return 0.0
  }

  /**
   * Handles a key press event.
   * @param key Key name (e.g., 'ArrowUp', 'w', 'ArrowLeft', 'a', 'ArrowRight', 'd')
   */
  onKeyDown(key: string): void {
    const normalizedKey = key.toLowerCase()
    
    // Thrust keys
    if (normalizedKey === 'arrowup' || normalizedKey === 'w') {
      this.thrustActive = true
    }
    
    // Turn left keys
    if (normalizedKey === 'arrowleft' || normalizedKey === 'a') {
      this.turnLeftActive = true
    }
    
    // Turn right keys
    if (normalizedKey === 'arrowright' || normalizedKey === 'd') {
      this.turnRightActive = true
    }
  }

  /**
   * Handles a key release event.
   * @param key Key name (e.g., 'ArrowUp', 'w', 'ArrowLeft', 'a', 'ArrowRight', 'd')
   */
  onKeyUp(key: string): void {
    const normalizedKey = key.toLowerCase()
    
    // Thrust keys
    if (normalizedKey === 'arrowup' || normalizedKey === 'w') {
      this.thrustActive = false
    }
    
    // Turn left keys
    if (normalizedKey === 'arrowleft' || normalizedKey === 'a') {
      this.turnLeftActive = false
    }
    
    // Turn right keys
    if (normalizedKey === 'arrowright' || normalizedKey === 'd') {
      this.turnRightActive = false
    }
  }

  /**
   * Attaches event listeners to the window for keyboard events.
   * Call this to start listening for keyboard input.
   */
  attach(): void {
    if (this.keydownListener || this.keyupListener) {
      // Already attached
      return
    }

    this.keydownListener = (event: KeyboardEvent) => {
      this.onKeyDown(event.key)
    }

    this.keyupListener = (event: KeyboardEvent) => {
      this.onKeyUp(event.key)
    }

    window.addEventListener('keydown', this.keydownListener)
    window.addEventListener('keyup', this.keyupListener)
  }

  /**
   * Removes event listeners from the window.
   * Call this to stop listening for keyboard input.
   */
  detach(): void {
    if (this.keydownListener) {
      window.removeEventListener('keydown', this.keydownListener)
      this.keydownListener = null
    }

    if (this.keyupListener) {
      window.removeEventListener('keyup', this.keyupListener)
      this.keyupListener = null
    }
  }

  /**
   * Resets all input state to neutral (thrust=0, turn=0).
   * Useful when game state resets or when detaching.
   */
  reset(): void {
    this.thrustActive = false
    this.turnLeftActive = false
    this.turnRightActive = false
  }
}

