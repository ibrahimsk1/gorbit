/**
 * Integration tests for input command generator.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { InputCommandGenerator } from './command'
import { KeyboardInputHandler } from './keyboard'
import { CommandHistory } from '../net/command-history'
import { NetworkClient } from '../net/client'

// Mock WebSocket for NetworkClient
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
}

const OriginalWebSocket = global.WebSocket

describe('InputCommandGenerator', () => {
  let generator: InputCommandGenerator
  let keyboardHandler: KeyboardInputHandler
  let commandHistory: CommandHistory
  let networkClient: NetworkClient

  beforeEach(() => {
    // @ts-expect-error - Mock WebSocket for testing
    global.WebSocket = MockWebSocket as any
    
    keyboardHandler = new KeyboardInputHandler()
    commandHistory = new CommandHistory()
    networkClient = new NetworkClient()
    generator = new InputCommandGenerator(keyboardHandler, commandHistory, networkClient)
  })

  afterEach(() => {
    generator.stop()
    networkClient.disconnect()
    global.WebSocket = OriginalWebSocket
    vi.clearAllTimers()
  })

  describe('Initial State', () => {
    it('does not send commands when not started', async () => {
      const sendInputSpy = vi.spyOn(networkClient, 'sendInput')
      
      keyboardHandler.onKeyDown('ArrowUp')
      
      // Wait a bit to ensure no commands are sent
      await new Promise(resolve => setTimeout(resolve, 100))
      
      expect(sendInputSpy).not.toHaveBeenCalled()
      
      sendInputSpy.mockRestore()
    })

    it('does not send commands when network client is not connected', async () => {
      const sendInputSpy = vi.spyOn(networkClient, 'sendInput')
      
      generator.start()
      keyboardHandler.onKeyDown('ArrowUp')
      
      // Wait a bit
      await new Promise(resolve => setTimeout(resolve, 100))
      
      expect(sendInputSpy).not.toHaveBeenCalled()
      
      generator.stop()
      sendInputSpy.mockRestore()
    })
  })

  describe('Command Generation', () => {
    it('generates commands with sequence numbers from CommandHistory', async () => {
      await networkClient.connect('ws://localhost:8080/ws')
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const sendInputSpy = vi.spyOn(networkClient, 'sendInput')
      
      generator.start()
      keyboardHandler.onKeyDown('ArrowUp')
      
      // Wait for command to be sent (60ms interval)
      await new Promise(resolve => setTimeout(resolve, 100))
      
      expect(sendInputSpy).toHaveBeenCalled()
      const firstCall = sendInputSpy.mock.calls[0]
      expect(firstCall[0]).toBe(1) // Sequence number should be 1
      
      generator.stop()
      sendInputSpy.mockRestore()
    })

    it('adds commands to history before sending', async () => {
      await networkClient.connect('ws://localhost:8080/ws')
      await new Promise(resolve => setTimeout(resolve, 50))
      
      generator.start()
      keyboardHandler.onKeyDown('ArrowUp')
      
      // Wait for command to be sent
      await new Promise(resolve => setTimeout(resolve, 100))
      
      const command = commandHistory.getCommand(1)
      expect(command).not.toBeNull()
      expect(command?.seq).toBe(1)
      expect(command?.thrust).toBe(1.0)
      expect(command?.turn).toBe(0.0)
      
      generator.stop()
    })

    it('sends commands with correct thrust and turn values', async () => {
      await networkClient.connect('ws://localhost:8080/ws')
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const sendInputSpy = vi.spyOn(networkClient, 'sendInput')
      
      generator.start()
      keyboardHandler.onKeyDown('ArrowUp')
      keyboardHandler.onKeyDown('ArrowLeft')
      
      // Wait for command to be sent
      await new Promise(resolve => setTimeout(resolve, 100))
      
      expect(sendInputSpy).toHaveBeenCalled()
      const call = sendInputSpy.mock.calls[0]
      expect(call[0]).toBe(1) // seq
      expect(call[1]).toBe(1.0) // thrust
      expect(call[2]).toBe(-1.0) // turn
      
      generator.stop()
      sendInputSpy.mockRestore()
    })

    it('increments sequence numbers for each command', async () => {
      await networkClient.connect('ws://localhost:8080/ws')
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const sendInputSpy = vi.spyOn(networkClient, 'sendInput')
      
      generator.start()
      keyboardHandler.onKeyDown('ArrowUp')
      
      // Wait for first command
      await new Promise(resolve => setTimeout(resolve, 100))
      
      keyboardHandler.onKeyUp('ArrowUp')
      keyboardHandler.onKeyDown('ArrowLeft')
      
      // Wait for second command
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // Stop generator to prevent more commands
      generator.stop()
      
      // Should have at least 2 commands, and sequence numbers should increment
      expect(sendInputSpy).toHaveBeenCalled()
      const calls = sendInputSpy.mock.calls
      expect(calls.length).toBeGreaterThanOrEqual(2)
      
      // Check that sequence numbers increment correctly
      expect(calls[0][0]).toBe(1) // First command seq = 1
      expect(calls[1][0]).toBe(2) // Second command seq = 2
      
      // Verify sequence numbers are sequential
      for (let i = 1; i < calls.length; i++) {
        expect(calls[i][0]).toBeGreaterThan(calls[i - 1][0])
      }
      
      sendInputSpy.mockRestore()
    })
  })

  describe('Input Clamping', () => {
    it('clamps thrust values to [0.0, 1.0]', async () => {
      await networkClient.connect('ws://localhost:8080/ws')
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const sendInputSpy = vi.spyOn(networkClient, 'sendInput')
      
      generator.start()
      keyboardHandler.onKeyDown('ArrowUp') // Thrust = 1.0
      
      await new Promise(resolve => setTimeout(resolve, 100))
      
      expect(sendInputSpy).toHaveBeenCalled()
      const call = sendInputSpy.mock.calls[0]
      expect(call[1]).toBeGreaterThanOrEqual(0.0)
      expect(call[1]).toBeLessThanOrEqual(1.0)
      
      generator.stop()
      sendInputSpy.mockRestore()
    })

    it('clamps turn values to [-1.0, 1.0]', async () => {
      await networkClient.connect('ws://localhost:8080/ws')
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const sendInputSpy = vi.spyOn(networkClient, 'sendInput')
      
      generator.start()
      keyboardHandler.onKeyDown('ArrowLeft') // Turn = -1.0
      
      await new Promise(resolve => setTimeout(resolve, 100))
      
      expect(sendInputSpy).toHaveBeenCalled()
      const call = sendInputSpy.mock.calls[0]
      expect(call[2]).toBeGreaterThanOrEqual(-1.0)
      expect(call[2]).toBeLessThanOrEqual(1.0)
      
      generator.stop()
      sendInputSpy.mockRestore()
    })
  })

  describe('Command Timing', () => {
    it('sends commands at appropriate intervals (~60ms)', async () => {
      await networkClient.connect('ws://localhost:8080/ws')
      await new Promise(resolve => setTimeout(resolve, 50))
      
      vi.useFakeTimers()
      const sendInputSpy = vi.spyOn(networkClient, 'sendInput')
      
      generator.start()
      keyboardHandler.onKeyDown('ArrowUp')
      
      // Fast-forward time
      vi.advanceTimersByTime(60)
      
      expect(sendInputSpy).toHaveBeenCalled()
      
      // Advance another interval
      vi.advanceTimersByTime(60)
      
      // Should send another command (even if input unchanged, for keep-alive)
      expect(sendInputSpy).toHaveBeenCalledTimes(2)
      
      generator.stop()
      vi.useRealTimers()
      sendInputSpy.mockRestore()
    })
  })

  describe('Connection State', () => {
    it('only sends commands when network client is connected', async () => {
      const sendInputSpy = vi.spyOn(networkClient, 'sendInput')
      
      generator.start()
      keyboardHandler.onKeyDown('ArrowUp')
      
      // Wait a bit - should not send since not connected
      await new Promise(resolve => setTimeout(resolve, 100))
      
      expect(sendInputSpy).not.toHaveBeenCalled()
      
      // Now connect
      await networkClient.connect('ws://localhost:8080/ws')
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Wait for command to be sent
      await new Promise(resolve => setTimeout(resolve, 100))
      
      expect(sendInputSpy).toHaveBeenCalled()
      
      generator.stop()
      sendInputSpy.mockRestore()
    })

    it('stops sending commands when network client disconnects', async () => {
      await networkClient.connect('ws://localhost:8080/ws')
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const sendInputSpy = vi.spyOn(networkClient, 'sendInput')
      
      generator.start()
      keyboardHandler.onKeyDown('ArrowUp')
      
      // Wait for first command
      await new Promise(resolve => setTimeout(resolve, 100))
      
      expect(sendInputSpy).toHaveBeenCalled()
      sendInputSpy.mockClear()
      
      // Disconnect
      networkClient.disconnect()
      await new Promise(resolve => setTimeout(resolve, 50))
      
      // Wait a bit - should not send more commands
      await new Promise(resolve => setTimeout(resolve, 100))
      
      expect(sendInputSpy).not.toHaveBeenCalled()
      
      generator.stop()
      sendInputSpy.mockRestore()
    })
  })

  describe('Lifecycle Management', () => {
    it('starts command generation loop when start() is called', async () => {
      await networkClient.connect('ws://localhost:8080/ws')
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const sendInputSpy = vi.spyOn(networkClient, 'sendInput')
      
      keyboardHandler.onKeyDown('ArrowUp')
      
      // Should not send before start()
      await new Promise(resolve => setTimeout(resolve, 100))
      expect(sendInputSpy).not.toHaveBeenCalled()
      
      generator.start()
      
      // Should send after start()
      await new Promise(resolve => setTimeout(resolve, 100))
      expect(sendInputSpy).toHaveBeenCalled()
      
      generator.stop()
      sendInputSpy.mockRestore()
    })

    it('stops command generation loop when stop() is called', async () => {
      await networkClient.connect('ws://localhost:8080/ws')
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const sendInputSpy = vi.spyOn(networkClient, 'sendInput')
      
      generator.start()
      keyboardHandler.onKeyDown('ArrowUp')
      
      // Wait for first command
      await new Promise(resolve => setTimeout(resolve, 100))
      
      expect(sendInputSpy).toHaveBeenCalled()
      sendInputSpy.mockClear()
      
      generator.stop()
      
      // Wait a bit - should not send more commands
      await new Promise(resolve => setTimeout(resolve, 100))
      
      expect(sendInputSpy).not.toHaveBeenCalled()
      
      sendInputSpy.mockRestore()
    })

    it('can be started and stopped multiple times', async () => {
      await networkClient.connect('ws://localhost:8080/ws')
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const sendInputSpy = vi.spyOn(networkClient, 'sendInput')
      
      generator.start()
      generator.stop()
      generator.start()
      generator.stop()
      
      keyboardHandler.onKeyDown('ArrowUp')
      generator.start()
      
      await new Promise(resolve => setTimeout(resolve, 100))
      
      expect(sendInputSpy).toHaveBeenCalled()
      
      generator.stop()
      sendInputSpy.mockRestore()
    })
  })

  describe('Manual Update', () => {
    it('generates and sends command when update() is called', async () => {
      await networkClient.connect('ws://localhost:8080/ws')
      await new Promise(resolve => setTimeout(resolve, 50))
      
      const sendInputSpy = vi.spyOn(networkClient, 'sendInput')
      
      keyboardHandler.onKeyDown('ArrowUp')
      
      // Manual update should send command immediately
      generator.update()
      
      expect(sendInputSpy).toHaveBeenCalled()
      const call = sendInputSpy.mock.calls[0]
      expect(call[0]).toBe(1) // seq
      expect(call[1]).toBe(1.0) // thrust
      expect(call[2]).toBe(0.0) // turn
      
      sendInputSpy.mockRestore()
    })

    it('only sends command when connected in manual update mode', async () => {
      const sendInputSpy = vi.spyOn(networkClient, 'sendInput')
      
      keyboardHandler.onKeyDown('ArrowUp')
      generator.update()
      
      // Should not send since not connected
      expect(sendInputSpy).not.toHaveBeenCalled()
      
      sendInputSpy.mockRestore()
    })
  })
})

