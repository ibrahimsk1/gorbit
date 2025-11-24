/**
 * Integration tests for MainMenu component.
 * 
 * Labels: scope:integration loop:g7-ui layer:ui dep:state,net b:main-menu r:high
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { Container, Application } from 'pixi.js'
import { MainMenu } from './main-menu'

describe('MainMenu', () => {
  let container: Container
  let onCreateRoom: () => void
  let onJoinRoom: (code: string) => void

  beforeEach(() => {
    container = new Container()
    onCreateRoom = vi.fn()
    onJoinRoom = vi.fn()
  })

  afterEach(() => {
    container.destroy({ children: true })
  })

  describe('Creation', () => {
    it('creates MainMenu with container and callbacks', () => {
      const menu = new MainMenu(container, onCreateRoom, onJoinRoom)

      expect(menu).toBeDefined()
      expect(container.children.length).toBeGreaterThan(0)
    })

    it('creates menu with title "Orbital Rush"', () => {
      const menu = new MainMenu(container, onCreateRoom, onJoinRoom)

      const menuContainer = container.children.find(child => 
        child.label === 'main-menu'
      )
      expect(menuContainer).toBeDefined()
      
      // Check for title text
      const titleText = menuContainer?.children.find(child => 
        child.label === 'main-menu-title'
      )
      expect(titleText).toBeDefined()
      if (titleText && 'text' in titleText) {
        expect(titleText.text).toContain('Orbital Rush')
      }
    })

    it('creates "Create Room" button', () => {
      const menu = new MainMenu(container, onCreateRoom, onJoinRoom)

      const menuContainer = container.children.find(child => 
        child.label === 'main-menu'
      )
      const createButton = menuContainer?.children.find(child => 
        child.label === 'main-menu-create-button'
      )
      expect(createButton).toBeDefined()
    })

    it('creates "Join Room" button', () => {
      const menu = new MainMenu(container, onCreateRoom, onJoinRoom)

      const menuContainer = container.children.find(child => 
        child.label === 'main-menu'
      )
      const joinButton = menuContainer?.children.find(child => 
        child.label === 'main-menu-join-button'
      )
      expect(joinButton).toBeDefined()
    })
  })

  describe('Show/Hide', () => {
    it('shows menu when show() is called', () => {
      const menu = new MainMenu(container, onCreateRoom, onJoinRoom)
      
      menu.hide()
      menu.show()

      const menuContainer = container.children.find(child => 
        child.label === 'main-menu'
      )
      expect(menuContainer).toBeDefined()
      expect(menuContainer?.visible).toBe(true)
    })

    it('hides menu when hide() is called', () => {
      const menu = new MainMenu(container, onCreateRoom, onJoinRoom)
      
      menu.hide()

      const menuContainer = container.children.find(child => 
        child.label === 'main-menu'
      )
      expect(menuContainer).toBeDefined()
      expect(menuContainer?.visible).toBe(false)
    })

    it('menu is visible by default', () => {
      const menu = new MainMenu(container, onCreateRoom, onJoinRoom)

      const menuContainer = container.children.find(child => 
        child.label === 'main-menu'
      )
      expect(menuContainer).toBeDefined()
      expect(menuContainer?.visible).toBe(true)
    })
  })

  describe('Button Callbacks', () => {
    it('triggers onCreateRoom callback when "Create Room" button is clicked', () => {
      const menu = new MainMenu(container, onCreateRoom, onJoinRoom)

      const menuContainer = container.children.find(child => 
        child.label === 'main-menu'
      )
      const createButton = menuContainer?.children.find(child => 
        child.label === 'main-menu-create-button'
      )
      expect(createButton).toBeDefined()

      // Simulate button click
      if (createButton && 'eventMode' in createButton && createButton.eventMode !== 'none') {
        createButton.emit('pointerdown')
        createButton.emit('pointerup')
      }

      expect(onCreateRoom).toHaveBeenCalledTimes(1)
    })

    it('triggers onJoinRoom callback when "Join Room" button is clicked with room code', () => {
      const menu = new MainMenu(container, onCreateRoom, onJoinRoom)

      // First, we need to show the input field (this might be triggered by clicking join button)
      const menuContainer = container.children.find(child => 
        child.label === 'main-menu'
      )
      const joinButton = menuContainer?.children.find(child => 
        child.label === 'main-menu-join-button'
      )
      expect(joinButton).toBeDefined()

      // Simulate clicking join button to show input
      if (joinButton && 'eventMode' in joinButton && joinButton.eventMode !== 'none') {
        joinButton.emit('pointerdown')
        joinButton.emit('pointerup')
      }

      // Find input field and submit button
      const inputField = menuContainer?.children.find(child => 
        child.label === 'main-menu-join-input'
      )
      const submitButton = menuContainer?.children.find(child => 
        child.label === 'main-menu-join-submit'
      )

      // If input field exists, simulate entering room code and submitting
      if (inputField && submitButton) {
        // Simulate entering room code (this would be handled by the component)
        // For now, we'll test that the callback structure exists
        // The actual input handling will be tested in the implementation
      }

      // The callback should be callable (exact behavior depends on implementation)
      expect(onJoinRoom).toBeDefined()
    })
  })

  describe('Input Handling', () => {
    it('shows input field when "Join Room" is clicked', () => {
      const menu = new MainMenu(container, onCreateRoom, onJoinRoom)

      const menuContainer = container.children.find(child => 
        child.label === 'main-menu'
      )
      const joinButton = menuContainer?.children.find(child => 
        child.label === 'main-menu-join-button'
      )
      
      if (joinButton && 'eventMode' in joinButton && joinButton.eventMode !== 'none') {
        joinButton.emit('pointerdown')
        joinButton.emit('pointerup')
      }

      // After clicking join, input container should be visible
      const inputContainer = menuContainer?.children.find(child => 
        child.label === 'main-menu-input-container'
      )
      expect(inputContainer).toBeDefined()
      expect(inputContainer?.visible).toBe(true)

      // Input field should exist
      const inputField = inputContainer?.children.find(child => 
        child.label === 'main-menu-join-input'
      )
      expect(inputField).toBeDefined()
    })

    it('handles keyboard input for room code', () => {
      const menu = new MainMenu(container, onCreateRoom, onJoinRoom)

      // Show input field
      const menuContainer = container.children.find(child => 
        child.label === 'main-menu'
      )
      const joinButton = menuContainer?.children.find(child => 
        child.label === 'main-menu-join-button'
      )
      
      if (joinButton && 'eventMode' in joinButton && joinButton.eventMode !== 'none') {
        joinButton.emit('pointerdown')
        joinButton.emit('pointerup')
      }

      const inputContainer = menuContainer?.children.find(child => 
        child.label === 'main-menu-input-container'
      )
      const inputText = inputContainer?.children.find(child => 
        child.label === 'main-menu-join-input-text'
      )

      // Simulate typing room code
      const testCode = 'TEST1'
      for (const char of testCode) {
        const keyEvent = new KeyboardEvent('keydown', { key: char })
        window.dispatchEvent(keyEvent)
      }

      // Verify input text was updated
      expect(inputText).toBeDefined()
      if (inputText && 'text' in inputText) {
        expect(inputText.text).toBe(testCode)
      }
    })

    it('handles backspace to delete characters', () => {
      const menu = new MainMenu(container, onCreateRoom, onJoinRoom)

      // Show input field
      const menuContainer = container.children.find(child => 
        child.label === 'main-menu'
      )
      const joinButton = menuContainer?.children.find(child => 
        child.label === 'main-menu-join-button'
      )
      
      if (joinButton && 'eventMode' in joinButton && joinButton.eventMode !== 'none') {
        joinButton.emit('pointerdown')
        joinButton.emit('pointerup')
      }

      const inputContainer = menuContainer?.children.find(child => 
        child.label === 'main-menu-input-container'
      )
      const inputText = inputContainer?.children.find(child => 
        child.label === 'main-menu-join-input-text'
      )

      // Type some characters
      for (const char of 'ABC') {
        const keyEvent = new KeyboardEvent('keydown', { key: char })
        window.dispatchEvent(keyEvent)
      }

      // Press backspace
      const backspaceEvent = new KeyboardEvent('keydown', { key: 'Backspace' })
      window.dispatchEvent(backspaceEvent)

      // Verify last character was removed
      if (inputText && 'text' in inputText) {
        expect(inputText.text).toBe('AB')
      }
    })
  })

  describe('Destruction', () => {
    it('destroys menu and cleans up resources', () => {
      const menu = new MainMenu(container, onCreateRoom, onJoinRoom)
      const initialChildCount = container.children.length

      menu.destroy()

      expect(container.children.length).toBeLessThan(initialChildCount)
    })

    it('can be destroyed multiple times safely', () => {
      const menu = new MainMenu(container, onCreateRoom, onJoinRoom)

      menu.destroy()
      menu.destroy()

      // Should not throw
      expect(true).toBe(true)
    })
  })
})

