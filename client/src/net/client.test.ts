/**
 * Integration tests for network client with message handling.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:ws
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
    it('calls onSnapshot callback when SnapshotMessage is received', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onSnapshotSpy = vi.fn()
      client.onSnapshot(onSnapshotSpy)
      
      const snapshot: SnapshotMessage = {
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
      mockWs.simulateMessage(JSON.stringify(snapshot))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onSnapshotSpy).toHaveBeenCalledWith(snapshot)
    })

    it('deserializes SnapshotMessage correctly', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onSnapshotSpy = vi.fn()
      client.onSnapshot(onSnapshotSpy)
      
      const snapshot = {
        t: 'snapshot',
        tick: 42,
        ship: {
          pos: { x: 100, y: 200 },
          vel: { x: 10, y: -5 },
          rot: 1.57,
          energy: 75.5
        },
        sun: {
          pos: { x: 0, y: 0 },
          radius: 15.0
        },
        pallets: [
          { id: 1, pos: { x: 50, y: 50 }, active: true },
          { id: 2, pos: { x: -50, y: -50 }, active: false }
        ],
        done: false,
        win: false
      }
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify(snapshot))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onSnapshotSpy).toHaveBeenCalled()
      const receivedSnapshot = onSnapshotSpy.mock.calls[0][0]
      expect(receivedSnapshot).toEqual(snapshot)
      expect(receivedSnapshot.tick).toBe(42)
      expect(receivedSnapshot.pallets).toHaveLength(2)
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
})

