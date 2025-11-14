/**
 * Integration tests for WebSocket connection and message handling.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:ws
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { WebSocketClient } from '../net/ws'
import { NetworkClient } from '../net/client'
import type { SnapshotMessage, InputMessage } from '../net/protocol'
import { MockWebSocket } from './test-helpers'

const OriginalWebSocket = global.WebSocket

beforeEach(() => {
  // @ts-expect-error - Mock WebSocket for testing
  global.WebSocket = MockWebSocket as any
})

afterEach(() => {
  global.WebSocket = OriginalWebSocket
})

describe('WebSocket Integration', () => {
  describe('WebSocketClient Connection Lifecycle', () => {
    let client: WebSocketClient

    beforeEach(() => {
      client = new WebSocketClient()
    })

    afterEach(() => {
      if (client) {
        client.disconnect()
      }
    })

    it('connects to server at configured URL', async () => {
      const url = 'ws://localhost:8080/ws'
      await client.connect(url)
      
      expect(client.isConnected()).toBe(true)
      expect(client.getReadyState()).toBe(WebSocket.OPEN)
    })

    it('tracks connection state correctly', async () => {
      expect(client.isConnected()).toBe(false)
      
      await client.connect('ws://localhost:8080/ws')
      
      expect(client.isConnected()).toBe(true)
    })

    it('calls onOpen callback on successful connection', async () => {
      const onOpenSpy = vi.fn()
      client.onOpen(onOpenSpy)
      
      await client.connect('ws://localhost:8080/ws')
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onOpenSpy).toHaveBeenCalled()
    })

    it('handles connection failure gracefully', async () => {
      class FailingWebSocket {
        constructor() {
          throw new Error('Connection failed')
        }
      }
      // @ts-expect-error - Mock WebSocket for testing
      global.WebSocket = FailingWebSocket as any
      
      const client2 = new WebSocketClient()
      const onErrorSpy = vi.fn()
      client2.onError(onErrorSpy)
      
      await expect(client2.connect('ws://localhost:8080/ws')).rejects.toThrow()
      
      // Restore mock
      global.WebSocket = MockWebSocket as any
    })

    it('disconnects and closes connection', async () => {
      await client.connect('ws://localhost:8080/ws')
      expect(client.isConnected()).toBe(true)
      
      client.disconnect()
      
      expect(client.isConnected()).toBe(false)
    })

    it('calls onClose callback on disconnect', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onCloseSpy = vi.fn()
      client.onClose(onCloseSpy)
      
      client.disconnect()
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onCloseSpy).toHaveBeenCalled()
    })
  })

  describe('WebSocket Message Serialization', () => {
    let client: WebSocketClient

    beforeEach(() => {
      client = new WebSocketClient()
    })

    afterEach(() => {
      if (client) {
        client.disconnect()
      }
    })

    it('serializes InputMessage correctly', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const message: InputMessage = {
        t: 'input',
        seq: 1,
        thrust: 0.5,
        turn: 0.3
      }
      
      // WebSocketClient.send() takes a message object and stringifies it
      client.send(message as any)
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const mockWs = (client as any).ws as MockWebSocket
      expect(mockWs.sentMessages).toHaveLength(1)
      
      const sentMessage = JSON.parse(mockWs.sentMessages[0])
      expect(sentMessage).toEqual(message)
    })

    it('deserializes SnapshotMessage correctly', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onMessageSpy = vi.fn()
      client.onMessage(onMessageSpy)
      
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 42,
        ship: {
          pos: { x: 100, y: 200 },
          vel: { x: 10, y: -5 },
          rot: 1.57,
          energy: 75.5
        },
        planets: [
          { pos: { x: 0, y: 0 }, radius: 15.0 }
        ],
        pallets: [
          { id: 1, pos: { x: 50, y: 50 }, active: true }
        ],
        done: false,
        win: false
      }
      
      const mockWs = (client as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify(snapshot))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onMessageSpy).toHaveBeenCalled()
      const receivedMessage = onMessageSpy.mock.calls[0][0]
      expect(receivedMessage).toEqual(snapshot)
    })

    it('handles malformed JSON gracefully', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onMessageSpy = vi.fn()
      const onErrorSpy = vi.fn()
      client.onMessage(onMessageSpy)
      client.onError(onErrorSpy)
      
      const mockWs = (client as any).ws as MockWebSocket
      mockWs.simulateMessage('invalid json {')
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Should not call onMessage with invalid JSON
      expect(onMessageSpy).not.toHaveBeenCalled()
    })
  })

  describe('NetworkClient Message Sending', () => {
    let client: NetworkClient

    beforeEach(() => {
      client = new NetworkClient()
    })

    afterEach(() => {
      if (client) {
        client.disconnect()
      }
    })

    it('sends InputMessage to server with correct format', async () => {
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

    it('throws error when sending without connection', () => {
      expect(() => client.sendInput(1, 0.5, 0.3)).toThrow()
      expect(() => client.sendRestart()).toThrow()
    })
  })

  describe('NetworkClient Message Receiving', () => {
    let client: NetworkClient

    beforeEach(() => {
      client = new NetworkClient()
    })

    afterEach(() => {
      if (client) {
        client.disconnect()
      }
    })

    it('calls onSnapshot callback when SnapshotMessage is received', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onSnapshotSpy = vi.fn()
      client.onSnapshot(onSnapshotSpy)
      
      const serverSnapshot = {
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
      mockWs.simulateMessage(JSON.stringify(serverSnapshot))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onSnapshotSpy).toHaveBeenCalled()
    })

    it('deserializes SnapshotMessage correctly', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onSnapshotSpy = vi.fn()
      client.onSnapshot(onSnapshotSpy)
      
      const serverSnapshot = {
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
          { id: 1, pos: { x: 50, y: 50 }, active: true }
        ],
        done: false,
        win: false
      }
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify(serverSnapshot))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onSnapshotSpy).toHaveBeenCalled()
      const receivedSnapshot = onSnapshotSpy.mock.calls[0][0]
      expect(receivedSnapshot.tick).toBe(42)
      expect(receivedSnapshot.planets).toHaveLength(1)
      expect(receivedSnapshot.pallets).toHaveLength(1)
    })

    it('handles multiple messages in order', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onSnapshotSpy = vi.fn()
      client.onSnapshot(onSnapshotSpy)
      
      const snapshot1 = {
        t: 'snapshot',
        tick: 1,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        sun: { pos: { x: 0, y: 0 }, radius: 10 },
        pallets: [],
        done: false,
        win: false
      }
      
      const snapshot2 = {
        t: 'snapshot',
        tick: 2,
        ship: { pos: { x: 10, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        sun: { pos: { x: 0, y: 0 }, radius: 10 },
        pallets: [],
        done: false,
        win: false
      }
      
      const wsClient = (client as any).wsClient
      const mockWs = (wsClient as any).ws as MockWebSocket
      mockWs.simulateMessage(JSON.stringify(snapshot1))
      mockWs.simulateMessage(JSON.stringify(snapshot2))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onSnapshotSpy).toHaveBeenCalledTimes(2)
      expect(onSnapshotSpy.mock.calls[0][0].tick).toBe(1)
      expect(onSnapshotSpy.mock.calls[1][0].tick).toBe(2)
    })
  })
})

