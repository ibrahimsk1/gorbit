/**
 * Integration tests for WebSocket client wrapper.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:ws
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { WebSocketClient } from './ws'

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
    // Simulate connection after a short delay
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

  // Test helpers
  simulateMessage(data: string): void {
    if (this.onmessage) {
      this.onmessage(new MessageEvent('message', { data }))
    }
  }

  simulateError(): void {
    if (this.onerror) {
      this.onerror(new Event('error'))
    }
  }
}

// Replace global WebSocket with mock
const OriginalWebSocket = global.WebSocket
beforeEach(() => {
  // @ts-expect-error - Mock WebSocket for testing
  global.WebSocket = MockWebSocket as any
})

afterEach(() => {
  global.WebSocket = OriginalWebSocket
})

describe('WebSocketClient', () => {
  let client: WebSocketClient

  beforeEach(() => {
    client = new WebSocketClient()
  })

  afterEach(() => {
    if (client) {
      client.disconnect()
    }
  })

  describe('Connection Lifecycle', () => {
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
      
      // Wait for async connection
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onOpenSpy).toHaveBeenCalled()
    })

    it('handles connection failure gracefully', async () => {
      // Mock WebSocket to fail immediately
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
      expect(client.getReadyState()).toBe(WebSocket.CLOSED)
    })

    it('calls onClose callback on disconnect', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onCloseSpy = vi.fn()
      client.onClose(onCloseSpy)
      
      client.disconnect()
      
      // Wait for close event
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onCloseSpy).toHaveBeenCalled()
    })

    it('calls onError callback on WebSocket error', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onErrorSpy = vi.fn()
      client.onError(onErrorSpy)
      
      // Get the mock WebSocket and simulate error
      const mockWs = (client as any).ws as MockWebSocket
      mockWs.simulateError()
      
      // Wait for error event
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onErrorSpy).toHaveBeenCalled()
    })
  })

  describe('Message Sending', () => {
    it('sends JSON message over WebSocket connection', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const message = { t: 'input', seq: 1, thrust: 0.5, turn: 0.3 }
      client.send(message)
      
      // Wait for send
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const mockWs = (client as any).ws as MockWebSocket
      expect(mockWs.sentMessages).toHaveLength(1)
      expect(JSON.parse(mockWs.sentMessages[0])).toEqual(message)
    })

    it('serializes message to JSON before sending', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const message = { t: 'restart' }
      client.send(message)
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const mockWs = (client as any).ws as MockWebSocket
      const sentData = mockWs.sentMessages[0]
      expect(() => JSON.parse(sentData)).not.toThrow()
      expect(JSON.parse(sentData)).toEqual(message)
    })

    it('throws error when sending without connection', () => {
      const message = { t: 'input', seq: 1, thrust: 0.5, turn: 0.3 }
      
      expect(() => client.send(message)).toThrow()
    })
  })

  describe('Message Receiving', () => {
    it('calls onMessage callback when message is received', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onMessageSpy = vi.fn()
      client.onMessage(onMessageSpy)
      
      const mockWs = (client as any).ws as MockWebSocket
      const message = { t: 'snapshot', tick: 1, ship: {}, sun: {}, pallets: [], done: false, win: false }
      mockWs.simulateMessage(JSON.stringify(message))
      
      // Wait for message event
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onMessageSpy).toHaveBeenCalledWith(message)
    })

    it('deserializes JSON message from server', async () => {
      await client.connect('ws://localhost:8080/ws')
      
      const onMessageSpy = vi.fn()
      client.onMessage(onMessageSpy)
      
      const mockWs = (client as any).ws as MockWebSocket
      const message = { t: 'snapshot', tick: 1, ship: { pos: { x: 0, y: 0 } }, sun: {}, pallets: [], done: false, win: false }
      mockWs.simulateMessage(JSON.stringify(message))
      
      await new Promise(resolve => setTimeout(resolve, 50))
      
      expect(onMessageSpy).toHaveBeenCalled()
      const receivedMessage = onMessageSpy.mock.calls[0][0]
      expect(receivedMessage).toEqual(message)
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
      // May call onError or ignore silently
      expect(onMessageSpy).not.toHaveBeenCalled()
    })
  })
})

