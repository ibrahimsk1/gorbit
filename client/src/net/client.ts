/**
 * Network client for game communication.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:ws
 */

import { WebSocketClient } from './ws'
import type { InputMessage, RestartMessage, SnapshotMessage, PlanetSnapshot } from './protocol'

export class NetworkClient {
  private wsClient: WebSocketClient
  private snapshotHandlers: Array<(snapshot: SnapshotMessage) => void> = []
  private connectHandlers: Array<() => void> = []
  private disconnectHandlers: Array<() => void> = []
  private errorHandlers: Array<(error: Error) => void> = []

  constructor() {
    this.wsClient = new WebSocketClient()

    // Forward WebSocket events
    this.wsClient.onOpen(() => {
      this.connectHandlers.forEach(handler => handler())
    })

    this.wsClient.onClose(() => {
      this.disconnectHandlers.forEach(handler => handler())
    })

    this.wsClient.onError((error) => {
      this.errorHandlers.forEach(handler => handler(error))
    })

    // Handle incoming messages
    this.wsClient.onMessage((data) => {
      // Only process snapshot messages
      if (data && typeof data === 'object' && data.t === 'snapshot') {
        // Convert server's 'sun' field to client's 'planets' array for extensibility
        const serverSnapshot = data as Record<string, unknown>
        const clientSnapshot: SnapshotMessage = {
          t: 'snapshot',
          tick: serverSnapshot.tick as number,
          ship: serverSnapshot.ship as SnapshotMessage['ship'],
          planets: serverSnapshot.planets
            ? (serverSnapshot.planets as PlanetSnapshot[])
            : serverSnapshot.sun
              ? [serverSnapshot.sun as PlanetSnapshot] // Convert single sun to planets array
              : [],
          pallets: (serverSnapshot.pallets || []) as SnapshotMessage['pallets'],
          done: serverSnapshot.done as boolean,
          win: serverSnapshot.win as boolean,
          version: serverSnapshot.version as number | undefined
        }
        this.snapshotHandlers.forEach(handler => handler(clientSnapshot))
      }
      // Ignore other message types for now
    })
  }

  /**
   * Connect to game server.
   * @param url WebSocket server URL (e.g., 'ws://localhost:8080/ws')
   * @returns Promise that resolves when connected
   */
  async connect(url: string): Promise<void> {
    await this.wsClient.connect(url)
  }

  /**
   * Disconnect from game server.
   */
  disconnect(): void {
    this.wsClient.disconnect()
  }

  /**
   * Send input command to server.
   * @param seq Sequence number for command
   * @param thrust Thrust value (0.0 to 1.0)
   * @param turn Turn value (-1.0 to 1.0)
   * @throws Error if not connected
   */
  sendInput(seq: number, thrust: number, turn: number): void {
    const message: InputMessage = {
      t: 'input',
      seq,
      thrust,
      turn
    }
    this.wsClient.send(message)
  }

  /**
   * Send restart command to server.
   * @throws Error if not connected
   */
  sendRestart(): void {
    const message: RestartMessage = {
      t: 'restart'
    }
    this.wsClient.send(message)
  }

  /**
   * Register callback for snapshot messages.
   * @param callback Function to call when a snapshot is received
   */
  onSnapshot(callback: (snapshot: SnapshotMessage) => void): void {
    this.snapshotHandlers.push(callback)
  }

  /**
   * Register callback for connection events.
   * @param callback Function to call when connected
   */
  onConnect(callback: () => void): void {
    this.connectHandlers.push(callback)
  }

  /**
   * Register callback for disconnection events.
   * @param callback Function to call when disconnected
   */
  onDisconnect(callback: () => void): void {
    this.disconnectHandlers.push(callback)
  }

  /**
   * Register callback for error events.
   * @param callback Function to call when an error occurs
   */
  onError(callback: (error: Error) => void): void {
    this.errorHandlers.push(callback)
  }

  /**
   * Check if client is connected.
   * @returns true if connected, false otherwise
   */
  isConnected(): boolean {
    return this.wsClient.isConnected()
  }
}

