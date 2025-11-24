/**
 * Network client for game communication.
 * Handles WebSocket connection, room management, input commands, and snapshot handling.
 * 
 * Labels: scope:integration loop:g5-network layer:net dep:proto
 */

import { WebSocketClient } from './ws'
import { CommandHistory } from './command-history'
import type { 
  InputMessage, 
  RestartMessage, 
  SnapshotMessage,
  CreateRoomMessage,
  JoinRoomMessage,
  LeaveRoomMessage,
  StartMatchMessage,
  RoomCreatedMessage,
  RoomStateMessage
} from './protocol'
import {
  isRoomCreatedMessage,
  isRoomStateMessage,
  isPlayerJoinedMessage,
  isPlayerLeftMessage,
  isMatchStartedMessage,
  isMatchEndedMessage,
  isSnapshotMessage,
  createCreateRoomMessage,
  createJoinRoomMessage,
  createLeaveRoomMessage,
  createStartMatchMessage,
  type PlayerInfo,
  type RoomStateMessage,
  type PlayerJoinedMessage,
  type PlayerLeftMessage,
  type MatchEndedMessage
} from './protocol'

/**
 * Room state information (matches RoomStateMessage structure without 't' field).
 */
export interface RoomState {
  roomCode: string
  players: PlayerInfo[]
  state: 'lobby' | 'playing' | 'ended'
  hostId: number
}

export class NetworkClient {
  private wsClient: WebSocketClient
  private commandHistory: CommandHistory
  private snapshotHandlers: Array<(snapshot: SnapshotMessage) => void> = []
  private connectHandlers: Array<() => void> = []
  private disconnectHandlers: Array<() => void> = []
  private errorHandlers: Array<(error: Error) => void> = []
  private createRoomPromise: { resolve: (roomCode: string) => void; reject: (error: Error) => void } | null = null
  private joinRoomPromise: { resolve: () => void; reject: (error: Error) => void } | null = null
  private roomStateHandlers: Array<(state: RoomState) => void> = []
  private playerJoinedHandlers: Array<(player: PlayerInfo) => void> = []
  private playerLeftHandlers: Array<(playerId: number) => void> = []
  private matchStartedHandlers: Array<() => void> = []
  private matchEndedHandlers: Array<(winnerId?: number) => void> = []

  constructor() {
    this.wsClient = new WebSocketClient()
    this.commandHistory = new CommandHistory()

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
      if (!data || typeof data !== 'object') {
        return
      }

      // Handle roomCreated message (response to createRoom)
      if (isRoomCreatedMessage(data)) {
        const msg = data as RoomCreatedMessage
        if (this.createRoomPromise) {
          this.createRoomPromise.resolve(msg.roomCode)
          this.createRoomPromise = null
        }
        return
      }

      // Handle roomState message (response to joinRoom)
      if (isRoomStateMessage(data)) {
        const msg = data as RoomStateMessage
        if (this.joinRoomPromise) {
          this.joinRoomPromise.resolve()
          this.joinRoomPromise = null
        }
        // Call room state event handlers
        const roomState: RoomState = {
          roomCode: msg.roomCode,
          players: msg.players,
          state: msg.state,
          hostId: msg.hostId
        }
        this.roomStateHandlers.forEach(handler => handler(roomState))
        return
      }

      // Handle playerJoined message
      if (isPlayerJoinedMessage(data)) {
        const msg = data as PlayerJoinedMessage
        this.playerJoinedHandlers.forEach(handler => handler(msg.player))
        return
      }

      // Handle playerLeft message
      if (isPlayerLeftMessage(data)) {
        const msg = data as PlayerLeftMessage
        this.playerLeftHandlers.forEach(handler => handler(msg.playerId))
        return
      }

      // Handle matchStarted message
      if (isMatchStartedMessage(data)) {
        this.matchStartedHandlers.forEach(handler => handler())
        return
      }

      // Handle matchEnded message
      if (isMatchEndedMessage(data)) {
        const msg = data as MatchEndedMessage
        this.matchEndedHandlers.forEach(handler => handler(msg.winnerId))
        return
      }

      // Handle snapshot messages (v1 multiplayer format)
      if (isSnapshotMessage(data)) {
        // v1 format: ships array, worldBounds, myShipId
        // Mark commands as confirmed when snapshot is received
        // For now, mark all unconfirmed commands as confirmed (simple heuristic)
        // This can be refined later to use tick-based confirmation
        const unconfirmed = this.commandHistory.getUnconfirmed()
        if (unconfirmed.length > 0) {
          // Mark all unconfirmed commands as confirmed
          // In a more sophisticated implementation, we'd use the snapshot tick
          // to determine which commands to confirm
          for (const cmd of unconfirmed) {
            this.commandHistory.markConfirmed(cmd.seq)
          }
        }
        this.snapshotHandlers.forEach(handler => handler(data))
        return
      }

      // Ignore other message types for now (will be handled in message routing CU)
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
    if (!this.isConnected()) {
      throw new Error('Not connected to server')
    }

    // Add command to history before sending
    this.commandHistory.addCommand(seq, thrust, turn)

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

  /**
   * Create a new room.
   * Sends createRoom message and waits for roomCreated response.
   * @returns Promise that resolves with the room code
   * @throws Error if not connected or operation fails
   */
  async createRoom(): Promise<string> {
    if (!this.isConnected()) {
      throw new Error('Not connected to server')
    }

    return new Promise<string>((resolve, reject) => {
      this.createRoomPromise = { resolve, reject }
      
      const message: CreateRoomMessage = createCreateRoomMessage()
      this.wsClient.send(message)

      // Timeout after 5 seconds
      setTimeout(() => {
        if (this.createRoomPromise) {
          this.createRoomPromise.reject(new Error('createRoom timeout'))
          this.createRoomPromise = null
        }
      }, 5000)
    })
  }

  /**
   * Join an existing room by room code.
   * Sends joinRoom message and waits for roomState response.
   * @param roomCode 6-character alphanumeric room code
   * @returns Promise that resolves when successfully joined
   * @throws Error if not connected or operation fails
   */
  async joinRoom(roomCode: string): Promise<void> {
    if (!this.isConnected()) {
      throw new Error('Not connected to server')
    }

    return new Promise<void>((resolve, reject) => {
      this.joinRoomPromise = { resolve, reject }
      
      const message: JoinRoomMessage = createJoinRoomMessage(roomCode)
      this.wsClient.send(message)

      // Timeout after 5 seconds
      setTimeout(() => {
        if (this.joinRoomPromise) {
          this.joinRoomPromise.reject(new Error('joinRoom timeout'))
          this.joinRoomPromise = null
        }
      }, 5000)
    })
  }

  /**
   * Leave the current room.
   * Sends leaveRoom message (fire-and-forget).
   * @throws Error if not connected
   */
  leaveRoom(): void {
    if (!this.isConnected()) {
      throw new Error('Not connected to server')
    }

    const message: LeaveRoomMessage = createLeaveRoomMessage()
    this.wsClient.send(message)
  }

  /**
   * Start the match (host only).
   * Sends startMatch message (fire-and-forget).
   * @throws Error if not connected
   */
  startMatch(): void {
    if (!this.isConnected()) {
      throw new Error('Not connected to server')
    }

    const message: StartMatchMessage = createStartMatchMessage()
    this.wsClient.send(message)
  }

  /**
   * Register callback for room state messages.
   * @param callback Function to call when a roomState message is received
   */
  onRoomState(callback: (state: RoomState) => void): void {
    this.roomStateHandlers.push(callback)
  }

  /**
   * Register callback for player joined messages.
   * @param callback Function to call when a playerJoined message is received
   */
  onPlayerJoined(callback: (player: PlayerInfo) => void): void {
    this.playerJoinedHandlers.push(callback)
  }

  /**
   * Register callback for player left messages.
   * @param callback Function to call when a playerLeft message is received
   */
  onPlayerLeft(callback: (playerId: number) => void): void {
    this.playerLeftHandlers.push(callback)
  }

  /**
   * Register callback for match started messages.
   * @param callback Function to call when a matchStarted message is received
   */
  onMatchStarted(callback: () => void): void {
    this.matchStartedHandlers.push(callback)
  }

  /**
   * Register callback for match ended messages.
   * @param callback Function to call when a matchEnded message is received
   */
  onMatchEnded(callback: (winnerId?: number) => void): void {
    this.matchEndedHandlers.push(callback)
  }

  /**
   * Get the CommandHistory instance.
   * This allows prediction systems to access command history for reconciliation.
   * @returns CommandHistory instance
   */
  getCommandHistory(): CommandHistory {
    return this.commandHistory
  }
}

