/**
 * Input command generator that creates and sends input commands to the server.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import { KeyboardInputHandler } from './keyboard'
import { CommandHistory } from '../net/command-history'
import { NetworkClient } from '../net/client'

/**
 * Input command generator that integrates keyboard input, command history,
 * and network client to generate and send input commands to the server.
 * 
 * Commands are sent at ~60ms intervals (16.67Hz) when the network client
 * is connected. Commands are added to the command history before sending
 * to enable client-side prediction and reconciliation.
 */
export class InputCommandGenerator {
  private keyboardHandler: KeyboardInputHandler
  private commandHistory: CommandHistory
  private networkClient: NetworkClient
  private intervalId: number | null = null
  private readonly commandInterval: number = 60 // ~60ms (16.67Hz)

  constructor(
    keyboardHandler: KeyboardInputHandler,
    commandHistory: CommandHistory,
    networkClient: NetworkClient
  ) {
    this.keyboardHandler = keyboardHandler
    this.commandHistory = commandHistory
    this.networkClient = networkClient
  }

  /**
   * Starts the command generation loop.
   * Commands will be sent periodically at the configured interval
   * when the network client is connected.
   */
  start(): void {
    if (this.intervalId !== null) {
      // Already started
      return
    }

    this.intervalId = window.setInterval(() => {
      this.update()
    }, this.commandInterval)
  }

  /**
   * Stops the command generation loop.
   */
  stop(): void {
    if (this.intervalId !== null) {
      window.clearInterval(this.intervalId)
      this.intervalId = null
    }
  }

  /**
   * Manually updates and sends a command if conditions are met.
   * This is called automatically by the interval timer when start() is called,
   * but can also be called manually for frame-based update loops.
   */
  update(): void {
    // Only send commands when connected
    if (!this.networkClient.isConnected()) {
      return
    }

    // Get current input state
    const thrust = this.clampThrust(this.keyboardHandler.getThrust())
    const turn = this.clampTurn(this.keyboardHandler.getTurn())

    // Get next sequence number
    const seq = this.commandHistory.getNextSequence()

    // Add command to history before sending
    this.commandHistory.addCommand(seq, thrust, turn)

    // Send command to server
    this.networkClient.sendInput(seq, thrust, turn)
  }

  /**
   * Clamps thrust value to [0.0, 1.0] range.
   */
  private clampThrust(value: number): number {
    if (value < 0.0) return 0.0
    if (value > 1.0) return 1.0
    return value
  }

  /**
   * Clamps turn value to [-1.0, 1.0] range.
   */
  private clampTurn(value: number): number {
    if (value < -1.0) return -1.0
    if (value > 1.0) return 1.0
    return value
  }
}

