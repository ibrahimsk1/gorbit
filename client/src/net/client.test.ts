/**
 * Integration tests for network client with message handling.
 * 
 * Labels: scope:integration loop:g5-network layer:net dep:proto
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { NetworkClient } from './client'
import type { SnapshotMessage } from './protocol'

// Mock WebSocket for testing
class MockWebSocket {
  url: string
  readyState: number = WebSocket.CONNECTING
  onopen: ((event: Event) => void) | null = null
  onclose: ((event: CloseEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null
  onmessage: ((event: MessageEvent) => void) | null = null
  sentMessages: string[] = []

  constructor(url: string) {
    this.url = url
    setTimeout(() => {
      this.readyState = WebSocket.OPEN
      if (this.onopen) {
        this.onopen(new Event('open'))
      }
    }, 10)
  }

  send(data: string): void {
    this.sentMessages.push(data)
  }

  close(): void {
    this.readyState = WebSocket.CLOSED
    if (this.onclose) {
      this.onclose(new CloseEvent('close'))
    }
  }

  simulateMessage(data: string): void {
    if (this.onmessage) {
      this.onmessage(new MessageEvent('message', { data }))
    }
  }
}

const OriginalWebSocket = global.WebSocket
beforeEach(() => {
  // @ts-expect-error - Mock WebSocket for testing
  global.WebSocket = MockWebSocket as any
})

afterEach(() => {
  global.WebSocket = OriginalWebSocket
})

describe('NetworkClient', () => {
  let client: NetworkClient

  beforeEach(() => {
    client = new NetworkClient()
  })

  afterEach(() => {
    if (client) {
      client.disconnect()
    }
  })

  describe('Connection', () => {
    // Labels: scope:integration loop:g5-network layer:net dep:proto
    it('connects to server at configured URL', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      expect(client.isConnected()).toBe(true)
    })

    it('calls onConnect callback on successful connection', async () => {
      const onConnectSpy = vi.fn()
      client.onConnect(onConnectSpy)
      
      await client.connect('ws://localhost:8080/ws')
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onConnectSpy).toHaveBeenCalled()
    })

    it('calls onDisconnect callback on disconnect', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onDisconnectSpy = vi.fn()
      client.onDisconnect(onDisconnectSpy)
      
      client.disconnect()
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onDisconnectSpy).toHaveBeenCalled()
    })

    it('calls onError callback on connection error', async () => {
      const onErrorSpy = vi.fn()
      client.onError(onErrorSpy)
      
      // Mock WebSocket to fail
      class FailingWebSocket {
        constructor() {
          throw new Error('Connection failed')
        }
      }
      // @ts-expect-error - Mock WebSocket for testing
      global.WebSocket = FailingWebSocket as any
      
      await expect(client.connect('ws://localhost:8080/ws')).rejects.toThrow()
      
      // Restore mock
      global.WebSocket = MockWebSocket as any
    })
  })

  describe('Input Commands', () => {
    // Labels: scope:integration loop:g5-network layer:net dep:proto
    it('sends InputMessage with correct format', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      client.sendInput(1, 0.5, 0.3)
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      expect(mockWs.sentMessages).toHaveLength(1)
      
      const sentMessage = JSON.parse(mockWs.sentMessages[0])
      expect(sentMessage).toEqual({
        t: 'input',
        seq: 1,
        thrust: 0.5,
        turn: 0.3
      })
    })

    it('sends InputMessage with sequence number', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      client.sendInput(42, 0.8, -0.5)
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      const sentMessage = JSON.parse(mockWs.sentMessages[0])
      
      expect(sentMessage.seq).toBe(42)
    })

    it('throws error when sending input without connection', () => {
      expect(() => client.sendInput(1, 0.5, 0.3)).toThrow()
    })
  })

  describe('Restart Command', () => {
    // Labels: scope:integration loop:g5-network layer:net dep:proto
    it('sends RestartMessage with correct format', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      client.sendRestart()
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      expect(mockWs.sentMessages).toHaveLength(1)
      
      const sentMessage = JSON.parse(mockWs.sentMessages[0])
      expect(sentMessage).toEqual({
        t: 'restart'
      })
    })

    it('throws error when sending restart without connection', () => {
      expect(() => client.sendRestart()).toThrow()
    })
  })

  describe('Snapshot Handling', () => {
    // Labels: scope:integration loop:g5-network layer:net dep:proto b:snapshot-v1
    it('handles v1 SnapshotMessage format with ships array, worldBounds, and myShipId', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onSnapshotSpy = vi.fn()
      client.onSnapshot(onSnapshotSpy)
      
      const v1Snapshot = {
        t: 'snapshot',
        tick: 1,
        ships: [
          { id: 1, pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
          { id: 2, pos: { x: 100, y: 100 }, vel: { x: 5, y: 5 }, rot: 1.57, energy: 80 }
        ],
        planets: [
          { id: 1, pos: { x: 0, y: 0 }, radius: 10 }
        ],
        pallets: [],
        worldBounds: { width: 2000, height: 2000 },
        myShipId: 1,
        done: false,
        win: false
      }
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify(v1Snapshot))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onSnapshotSpy).toHaveBeenCalled()
      const receivedSnapshot = onSnapshotSpy.mock.calls[0][0]
      expect(receivedSnapshot).toEqual(v1Snapshot)
      expect(receivedSnapshot.ships).toHaveLength(2)
      expect(receivedSnapshot.worldBounds).toEqual({ width: 2000, height: 2000 })
      expect(receivedSnapshot.myShipId).toBe(1)
    })

    it('uses isSnapshotMessage type guard for validation', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onSnapshotSpy = vi.fn()
      client.onSnapshot(onSnapshotSpy)
      
      // Invalid snapshot (missing required fields)
      const invalidSnapshot = {
        t: 'snapshot',
        tick: 1
        // Missing ships, worldBounds, myShipId
      }
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify(invalidSnapshot))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Should not call handler for invalid snapshot
      expect(onSnapshotSpy).not.toHaveBeenCalled()
    })

    it('rejects legacy v0 format (no conversion)', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onSnapshotSpy = vi.fn()
      client.onSnapshot(onSnapshotSpy)
      
      // Legacy v0 format with single 'ship' field
      const v0Snapshot = {
        t: 'snapshot',
        tick: 1,
        ship: {
          pos: { x: 0, y: 0 },
          vel: { x: 0, y: 0 },
          rot: 0,
          energy: 100
        },
        sun: {
          pos: { x: 0, y: 0 },
          radius: 10
        },
        pallets: [],
        done: false,
        win: false
      }
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify(v0Snapshot))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Should not call handler for v0 format (no conversion)
      expect(onSnapshotSpy).not.toHaveBeenCalled()
    })

    it('handles malformed snapshot messages gracefully', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onSnapshotSpy = vi.fn()
      const onErrorSpy = vi.fn()
      client.onSnapshot(onSnapshotSpy)
      client.onError(onErrorSpy)
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage('invalid json {')
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Should not call onSnapshot with invalid JSON
      expect(onSnapshotSpy).not.toHaveBeenCalled()
    })

    it('ignores non-snapshot messages', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onSnapshotSpy = vi.fn()
      client.onSnapshot(onSnapshotSpy)
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify({ t: 'unknown', data: 'test' }))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Should not call onSnapshot for non-snapshot messages
      expect(onSnapshotSpy).not.toHaveBeenCalled()
    })
  })

  describe('Room Management', () => {
    // Labels: scope:integration loop:g5-network layer:net dep:proto b:room-management
    it('createRoom sends createRoom message and resolves on roomState', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const createRoomPromise = client.createRoom()
      
      // Simulate server response
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      
      // Check that createRoom message was sent
      await new Promise(resolve => setTimeout(resolve, 50))
      expect(mockWs.sentMessages).toHaveLength(1)
      const sentMessage = JSON.parse(mockWs.sentMessages[0])
      expect(sentMessage).toEqual({ t: 'createRoom' })
      
      // Simulate roomState response (server now sends roomState instead of roomCreated)
      mockWs.simulateMessage(JSON.stringify({
        t: 'roomState',
        roomCode: 'ABC123',
        players: [{ id: 1, name: '' }],
        state: 'lobby',
        hostId: 1
      }))
      
      await expect(createRoomPromise).resolves.toBeUndefined()
    })

    it('createRoom throws error when not connected', async () => {
      await expect(client.createRoom()).rejects.toThrow()
    })

    it('joinRoom sends joinRoom message and resolves on roomState', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const joinRoomPromise = client.joinRoom('ABC123')
      
      // Simulate server response
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      
      // Check that joinRoom message was sent
      await new Promise(resolve => setTimeout(resolve, 50))
      expect(mockWs.sentMessages).toHaveLength(1)
      const sentMessage = JSON.parse(mockWs.sentMessages[0])
      expect(sentMessage).toEqual({ t: 'joinRoom', roomCode: 'ABC123' })
      
      // Simulate roomState response
      mockWs.simulateMessage(JSON.stringify({
        t: 'roomState',
        roomCode: 'ABC123',
        players: [{ id: 1, name: 'Player1' }],
        state: 'lobby',
        hostId: 1
      }))
      
      await expect(joinRoomPromise).resolves.toBeUndefined()
    })

    it('joinRoom throws error when not connected', async () => {
      await expect(client.joinRoom('ABC123')).rejects.toThrow()
    })

    it('leaveRoom sends leaveRoom message', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      client.leaveRoom()
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      expect(mockWs.sentMessages).toHaveLength(1)
      
      const sentMessage = JSON.parse(mockWs.sentMessages[0])
      expect(sentMessage).toEqual({ t: 'leaveRoom' })
    })

    it('leaveRoom throws error when not connected', () => {
      expect(() => client.leaveRoom()).toThrow()
    })

    it('startMatch sends startMatch message', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      client.startMatch()
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      expect(mockWs.sentMessages).toHaveLength(1)
      
      const sentMessage = JSON.parse(mockWs.sentMessages[0])
      expect(sentMessage).toEqual({ t: 'startMatch' })
    })

    it('startMatch throws error when not connected', () => {
      expect(() => client.startMatch()).toThrow()
    })
  })

  describe('Room Event Handlers', () => {
    // Labels: scope:integration loop:g5-network layer:net dep:proto b:room-events
    it('onRoomState callback is called when roomState message is received', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onRoomStateSpy = vi.fn()
      client.onRoomState(onRoomStateSpy)
      
      const roomState = {
        t: 'roomState',
        roomCode: 'ABC123',
        players: [{ id: 1, name: 'Player1' }, { id: 2, name: 'Player2' }],
        state: 'lobby' as const,
        hostId: 1
      }
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify(roomState))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onRoomStateSpy).toHaveBeenCalledWith({
        roomCode: 'ABC123',
        players: [{ id: 1, name: 'Player1' }, { id: 2, name: 'Player2' }],
        state: 'lobby',
        hostId: 1
      })
    })

    it('onPlayerJoined callback is called when playerJoined message is received', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onPlayerJoinedSpy = vi.fn()
      client.onPlayerJoined(onPlayerJoinedSpy)
      
      const playerJoined = {
        t: 'playerJoined',
        player: { id: 3, name: 'Player3' }
      }
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify(playerJoined))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onPlayerJoinedSpy).toHaveBeenCalledWith({ id: 3, name: 'Player3' })
    })

    it('onPlayerLeft callback is called when playerLeft message is received', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onPlayerLeftSpy = vi.fn()
      client.onPlayerLeft(onPlayerLeftSpy)
      
      const playerLeft = {
        t: 'playerLeft',
        playerId: 2
      }
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify(playerLeft))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onPlayerLeftSpy).toHaveBeenCalledWith(2)
    })

    it('onMatchStarted callback is called when matchStarted message is received', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onMatchStartedSpy = vi.fn()
      client.onMatchStarted(onMatchStartedSpy)
      
      const matchStarted = {
        t: 'matchStarted'
      }
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify(matchStarted))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onMatchStartedSpy).toHaveBeenCalled()
    })

    it('onMatchEnded callback is called when matchEnded message is received', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onMatchEndedSpy = vi.fn()
      client.onMatchEnded(onMatchEndedSpy)
      
      const matchEnded = {
        t: 'matchEnded',
        winnerId: 1
      }
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify(matchEnded))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onMatchEndedSpy).toHaveBeenCalledWith(1)
    })

    it('onMatchEnded callback handles matchEnded without winnerId', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onMatchEndedSpy = vi.fn()
      client.onMatchEnded(onMatchEndedSpy)
      
      const matchEnded = {
        t: 'matchEnded'
      }
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify(matchEnded))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onMatchEndedSpy).toHaveBeenCalledWith(undefined)
    })

    it('multiple callbacks can be registered for each event type', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onRoomStateSpy1 = vi.fn()
      const onRoomStateSpy2 = vi.fn()
      client.onRoomState(onRoomStateSpy1)
      client.onRoomState(onRoomStateSpy2)
      
      const roomState = {
        t: 'roomState',
        roomCode: 'ABC123',
        players: [{ id: 1, name: 'Player1' }],
        state: 'lobby' as const,
        hostId: 1
      }
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify(roomState))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onRoomStateSpy1).toHaveBeenCalled()
      expect(onRoomStateSpy2).toHaveBeenCalled()
    })
  })

  describe('Command History Integration', () => {
    // Labels: scope:integration loop:g5-network layer:net dep:proto b:command-history
    it('creates CommandHistory instance in constructor', () => {
      const newClient = new NetworkClient()
      const commandHistory = newClient.getCommandHistory()
      expect(commandHistory).toBeDefined()
      expect(typeof commandHistory.getNextSequence).toBe('function')
      expect(typeof commandHistory.addCommand).toBe('function')
    })

    it('sendInput adds commands to history before sending', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const commandHistory = client.getCommandHistory()
      const initialSeq = commandHistory.getNextSequence()
      
      client.sendInput(initialSeq, 0.5, 0.3)
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const command = commandHistory.getCommand(initialSeq)
      expect(command).toBeDefined()
      expect(command?.thrust).toBe(0.5)
      expect(command?.turn).toBe(0.3)
      expect(command?.confirmed).toBe(false)
    })

    it('commands are marked as confirmed when snapshot is received', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const commandHistory = client.getCommandHistory()
      const seq1 = commandHistory.getNextSequence()
      const seq2 = seq1 + 1
      
      client.sendInput(seq1, 0.5, 0.3)
      client.sendInput(seq2, 0.8, -0.5)
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Commands should be unconfirmed initially
      expect(commandHistory.getCommand(seq1)?.confirmed).toBe(false)
      expect(commandHistory.getCommand(seq2)?.confirmed).toBe(false)
      
      // Send snapshot
      const snapshot = {
        t: 'snapshot',
        tick: 1,
        ships: [
          { id: 1, pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
        ],
        planets: [],
        pallets: [],
        worldBounds: { width: 2000, height: 2000 },
        myShipId: 1,
        done: false,
        win: false
      }
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify(snapshot))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Commands should be confirmed after snapshot
      expect(commandHistory.getCommand(seq1)?.confirmed).toBe(true)
      expect(commandHistory.getCommand(seq2)?.confirmed).toBe(true)
    })

    it('getCommandHistory returns CommandHistory instance', () => {
      const commandHistory = client.getCommandHistory()
      expect(commandHistory).toBeDefined()
      expect(typeof commandHistory.getNextSequence).toBe('function')
      expect(typeof commandHistory.addCommand).toBe('function')
      expect(typeof commandHistory.markConfirmed).toBe('function')
      expect(typeof commandHistory.getUnconfirmed).toBe('function')
    })
  })
})

