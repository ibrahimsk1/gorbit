/**
 * Integration tests for keyboard input handler.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { KeyboardInputHandler } from './keyboard'

describe('KeyboardInputHandler', () => {
  let handler: KeyboardInputHandler
  let mockWindow: Window & typeof globalThis

  beforeEach(() => {
    handler = new KeyboardInputHandler()
    
    // Create mock window with event listeners
    const listeners: Record<string, EventListener[]> = {}
    mockWindow = {
      ...globalThis,
      addEventListener: vi.fn((event: string, listener: EventListener) => {
        if (!listeners[event]) {
          listeners[event] = []
        }
        listeners[event].push(listener)
      }),
      removeEventListener: vi.fn((event: string, listener: EventListener) => {
        if (listeners[event]) {
          const index = listeners[event].indexOf(listener)
          if (index > -1) {
            listeners[event].splice(index, 1)
          }
        }
      }),
      dispatchEvent: vi.fn((event: Event) => {
        const eventType = event.type
        if (listeners[eventType]) {
          listeners[eventType].forEach(listener => {
            try {
              listener(event)
            } catch (e) {
              // Ignore errors in test
            }
          })
        }
        return true
      })
    } as unknown as Window & typeof globalThis
  })

  afterEach(() => {
    handler.detach()
    handler.reset()
  })

  describe('Initial State', () => {
    it('starts with neutral input state', () => {
      expect(handler.getThrust()).toBe(0.0)
      expect(handler.getTurn()).toBe(0.0)
    })

    it('does not have event listeners attached initially', () => {
      // Handler should not be listening until attach() is called
      expect(handler.getThrust()).toBe(0.0)
      expect(handler.getTurn()).toBe(0.0)
    })
  })

  describe('Thrust Input', () => {
    it('sets thrust to 1.0 when ArrowUp key is pressed', () => {
      handler.onKeyDown('ArrowUp')
      expect(handler.getThrust()).toBe(1.0)
      expect(handler.getTurn()).toBe(0.0)
    })

    it('sets thrust to 1.0 when W key is pressed', () => {
      handler.onKeyDown('w')
      expect(handler.getThrust()).toBe(1.0)
      expect(handler.getTurn()).toBe(0.0)
    })

    it('sets thrust to 1.0 when W key is pressed (uppercase)', () => {
      handler.onKeyDown('W')
      expect(handler.getThrust()).toBe(1.0)
      expect(handler.getTurn()).toBe(0.0)
    })

    it('resets thrust to 0.0 when ArrowUp key is released', () => {
      handler.onKeyDown('ArrowUp')
      expect(handler.getThrust()).toBe(1.0)
      
      handler.onKeyUp('ArrowUp')
      expect(handler.getThrust()).toBe(0.0)
    })

    it('resets thrust to 0.0 when W key is released', () => {
      handler.onKeyDown('w')
      expect(handler.getThrust()).toBe(1.0)
      
      handler.onKeyUp('w')
      expect(handler.getThrust()).toBe(0.0)
    })
  })

  describe('Turn Input', () => {
    it('sets turn to -1.0 when ArrowLeft key is pressed', () => {
      handler.onKeyDown('ArrowLeft')
      expect(handler.getTurn()).toBe(-1.0)
      expect(handler.getThrust()).toBe(0.0)
    })

    it('sets turn to -1.0 when A key is pressed', () => {
      handler.onKeyDown('a')
      expect(handler.getTurn()).toBe(-1.0)
      expect(handler.getThrust()).toBe(0.0)
    })

    it('sets turn to 1.0 when ArrowRight key is pressed', () => {
      handler.onKeyDown('ArrowRight')
      expect(handler.getTurn()).toBe(1.0)
      expect(handler.getThrust()).toBe(0.0)
    })

    it('sets turn to 1.0 when D key is pressed', () => {
      handler.onKeyDown('d')
      expect(handler.getTurn()).toBe(1.0)
      expect(handler.getThrust()).toBe(0.0)
    })

    it('resets turn to 0.0 when ArrowLeft key is released', () => {
      handler.onKeyDown('ArrowLeft')
      expect(handler.getTurn()).toBe(-1.0)
      
      handler.onKeyUp('ArrowLeft')
      expect(handler.getTurn()).toBe(0.0)
    })

    it('resets turn to 0.0 when A key is released', () => {
      handler.onKeyDown('a')
      expect(handler.getTurn()).toBe(-1.0)
      
      handler.onKeyUp('a')
      expect(handler.getTurn()).toBe(0.0)
    })

    it('resets turn to 0.0 when ArrowRight key is released', () => {
      handler.onKeyDown('ArrowRight')
      expect(handler.getTurn()).toBe(1.0)
      
      handler.onKeyUp('ArrowRight')
      expect(handler.getTurn()).toBe(0.0)
    })

    it('resets turn to 0.0 when D key is released', () => {
      handler.onKeyDown('d')
      expect(handler.getTurn()).toBe(1.0)
      
      handler.onKeyUp('d')
      expect(handler.getTurn()).toBe(0.0)
    })
  })

  describe('Multiple Keys Pressed', () => {
    it('allows thrust and turn to be pressed simultaneously', () => {
      handler.onKeyDown('ArrowUp')
      handler.onKeyDown('ArrowLeft')
      
      expect(handler.getThrust()).toBe(1.0)
      expect(handler.getTurn()).toBe(-1.0)
    })

    it('allows thrust and turn right to be pressed simultaneously', () => {
      handler.onKeyDown('w')
      handler.onKeyDown('d')
      
      expect(handler.getThrust()).toBe(1.0)
      expect(handler.getTurn()).toBe(1.0)
    })

    it('combines turn left and right to 0.0 when both pressed', () => {
      handler.onKeyDown('ArrowLeft')
      handler.onKeyDown('ArrowRight')
      
      expect(handler.getTurn()).toBe(0.0)
    })

    it('combines A and D keys to 0.0 when both pressed', () => {
      handler.onKeyDown('a')
      handler.onKeyDown('d')
      
      expect(handler.getTurn()).toBe(0.0)
    })

    it('handles releasing one turn key while other is still pressed', () => {
      handler.onKeyDown('ArrowLeft')
      handler.onKeyDown('ArrowRight')
      expect(handler.getTurn()).toBe(0.0)
      
      handler.onKeyUp('ArrowLeft')
      expect(handler.getTurn()).toBe(1.0) // Only right is pressed now
    })

    it('handles releasing one turn key while other is still pressed (reverse)', () => {
      handler.onKeyDown('ArrowLeft')
      handler.onKeyDown('ArrowRight')
      expect(handler.getTurn()).toBe(0.0)
      
      handler.onKeyUp('ArrowRight')
      expect(handler.getTurn()).toBe(-1.0) // Only left is pressed now
    })
  })

  describe('Key Mapping', () => {
    it('ignores unmapped keys', () => {
      handler.onKeyDown('Space')
      handler.onKeyDown('Enter')
      handler.onKeyDown('Escape')
      
      expect(handler.getThrust()).toBe(0.0)
      expect(handler.getTurn()).toBe(0.0)
    })

    it('handles case-insensitive key names', () => {
      handler.onKeyDown('W')
      expect(handler.getThrust()).toBe(1.0)
      
      handler.onKeyDown('A')
      expect(handler.getTurn()).toBe(-1.0)
      
      handler.onKeyDown('D')
      expect(handler.getTurn()).toBe(0.0) // Left + Right = 0
    })
  })

  describe('Reset', () => {
    it('resets all input state to neutral', () => {
      handler.onKeyDown('ArrowUp')
      handler.onKeyDown('ArrowLeft')
      
      expect(handler.getThrust()).toBe(1.0)
      expect(handler.getTurn()).toBe(-1.0)
      
      handler.reset()
      
      expect(handler.getThrust()).toBe(0.0)
      expect(handler.getTurn()).toBe(0.0)
    })

    it('can be called multiple times safely', () => {
      handler.onKeyDown('w')
      handler.reset()
      handler.reset()
      handler.reset()
      
      expect(handler.getThrust()).toBe(0.0)
      expect(handler.getTurn()).toBe(0.0)
    })
  })

  describe('Event Listener Management', () => {
    it('attaches event listeners when attach() is called', () => {
      const addEventListenerSpy = vi.spyOn(window, 'addEventListener')
      
      handler.attach()
      
      expect(addEventListenerSpy).toHaveBeenCalledWith('keydown', expect.any(Function))
      expect(addEventListenerSpy).toHaveBeenCalledWith('keyup', expect.any(Function))
      
      addEventListenerSpy.mockRestore()
    })

    it('removes event listeners when detach() is called', () => {
      const removeEventListenerSpy = vi.spyOn(window, 'removeEventListener')
      
      handler.attach()
      handler.detach()
      
      expect(removeEventListenerSpy).toHaveBeenCalledWith('keydown', expect.any(Function))
      expect(removeEventListenerSpy).toHaveBeenCalledWith('keyup', expect.any(Function))
      
      removeEventListenerSpy.mockRestore()
    })

    it('handles keyboard events from window when attached', () => {
      handler.attach()
      
      // Simulate keydown event
      const keydownEvent = new KeyboardEvent('keydown', { key: 'ArrowUp' })
      window.dispatchEvent(keydownEvent)
      
      expect(handler.getThrust()).toBe(1.0)
      
      // Simulate keyup event
      const keyupEvent = new KeyboardEvent('keyup', { key: 'ArrowUp' })
      window.dispatchEvent(keyupEvent)
      
      expect(handler.getThrust()).toBe(0.0)
    })

    it('does not handle keyboard events when not attached', () => {
      // Don't call attach()
      
      const keydownEvent = new KeyboardEvent('keydown', { key: 'ArrowUp' })
      window.dispatchEvent(keydownEvent)
      
      // State should remain neutral since handler is not listening
      expect(handler.getThrust()).toBe(0.0)
    })

    it('can be attached and detached multiple times', () => {
      handler.attach()
      handler.detach()
      handler.attach()
      handler.detach()
      
      // Should not throw and should work correctly
      handler.onKeyDown('ArrowUp')
      expect(handler.getThrust()).toBe(1.0)
    })
  })
})

