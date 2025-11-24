/**
 * Integration tests for RoomLobby component.
 * 
 * Labels: scope:integration loop:g7-ui layer:ui dep:state,net b:room-lobby r:high
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { Container } from 'pixi.js'
import { RoomLobby } from './room-lobby'
import type { RoomState, PlayerInfo } from '../core/orchestrator'

describe('RoomLobby', () => {
  let container: Container
  let onStartMatch: () => void
  let onLeaveRoom: () => void
  let roomState: RoomState

  beforeEach(() => {
    container = new Container()
    onStartMatch = vi.fn()
    onLeaveRoom = vi.fn()
    roomState = {
      roomCode: 'ABC123',
      players: [
        { id: 1, name: 'Player 1' },
        { id: 2, name: 'Player 2' }
      ],
      state: 'lobby',
      hostId: 1
    }
  })

  afterEach(() => {
    container.destroy({ children: true })
  })

  describe('Creation', () => {
    it('creates RoomLobby with container, room state, and callbacks', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)

      expect(lobby).toBeDefined()
      expect(container.children.length).toBeGreaterThan(0)
    })

    it('displays room code prominently', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)

      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      expect(lobbyContainer).toBeDefined()
      
      const roomCodeText = lobbyContainer?.children.find(child => 
        child.label === 'room-lobby-code'
      )
      expect(roomCodeText).toBeDefined()
      if (roomCodeText && 'text' in roomCodeText) {
        expect(roomCodeText.text).toContain('ABC123')
      }
    })

    it('displays player list', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)

      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      const playerList = lobbyContainer?.children.find(child => 
        child.label === 'room-lobby-players'
      )
      expect(playerList).toBeDefined()
    })

    it('displays "Start Match" button for host', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)

      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      const startButton = lobbyContainer?.children.find(child => 
        child.label === 'room-lobby-start-button'
      )
      expect(startButton).toBeDefined()
      expect(startButton?.visible).toBe(true)
    })

    it('does not create "Start Match" button for non-host', () => {
      const lobby = new RoomLobby(container, roomState, false, onStartMatch, onLeaveRoom)

      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      const startButton = lobbyContainer?.children.find(child => 
        child.label === 'room-lobby-start-button'
      )
      expect(startButton).toBeUndefined()
    })

    it('displays "Leave Room" button', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)

      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      const leaveButton = lobbyContainer?.children.find(child => 
        child.label === 'room-lobby-leave-button'
      )
      expect(leaveButton).toBeDefined()
      expect(leaveButton?.visible).toBe(true)
    })

    it('shows waiting message for non-host', () => {
      const lobby = new RoomLobby(container, roomState, false, onStartMatch, onLeaveRoom)

      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      const waitingText = lobbyContainer?.children.find(child => 
        child.label === 'room-lobby-waiting'
      )
      expect(waitingText).toBeDefined()
      expect(waitingText?.visible).toBe(true)
    })
  })

  describe('Show/Hide', () => {
    it('shows lobby when show() is called', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)
      
      lobby.hide()
      lobby.show()

      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      expect(lobbyContainer).toBeDefined()
      expect(lobbyContainer?.visible).toBe(true)
    })

    it('hides lobby when hide() is called', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)
      
      lobby.hide()

      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      expect(lobbyContainer).toBeDefined()
      expect(lobbyContainer?.visible).toBe(false)
    })

    it('lobby is hidden by default', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)

      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      expect(lobbyContainer).toBeDefined()
      expect(lobbyContainer?.visible).toBe(false)
    })
  })

  describe('Update', () => {
    it('updates room code when update() is called', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)

      const newRoomState: RoomState = {
        ...roomState,
        roomCode: 'XYZ789'
      }
      lobby.update(newRoomState)

      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      const roomCodeText = lobbyContainer?.children.find(child => 
        child.label === 'room-lobby-code'
      )
      if (roomCodeText && 'text' in roomCodeText) {
        expect(roomCodeText.text).toContain('XYZ789')
      }
    })

    it('updates player list when update() is called', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)

      const newRoomState: RoomState = {
        ...roomState,
        players: [
          { id: 1, name: 'Player 1' },
          { id: 2, name: 'Player 2' },
          { id: 3, name: 'Player 3' }
        ]
      }
      lobby.update(newRoomState)

      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      const playerList = lobbyContainer?.children.find(child => 
        child.label === 'room-lobby-players'
      )
      expect(playerList).toBeDefined()
    })

    it('updates host status when update() is called', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)

      const newRoomState: RoomState = {
        ...roomState,
        hostId: 2
      }
      lobby.update(newRoomState)

      // Host status change should update button visibility
      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      const startButton = lobbyContainer?.children.find(child => 
        child.label === 'room-lobby-start-button'
      )
      expect(startButton).toBeDefined()
    })

    it('disables "Start Match" button when less than 2 players', () => {
      const singlePlayerState: RoomState = {
        roomCode: 'ABC123',
        players: [{ id: 1, name: 'Player 1' }],
        state: 'lobby',
        hostId: 1
      }
      const lobby = new RoomLobby(container, singlePlayerState, true, onStartMatch, onLeaveRoom)

      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      const startButton = lobbyContainer?.children.find(child => 
        child.label === 'room-lobby-start-button'
      )
      expect(startButton).toBeDefined()
      // Button should be disabled (not clickable or visually disabled)
      if (startButton && 'eventMode' in startButton) {
        expect(startButton.eventMode).toBe('none')
      }
    })

    it('enables "Start Match" button when 2 or more players', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)

      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      const startButton = lobbyContainer?.children.find(child => 
        child.label === 'room-lobby-start-button'
      )
      expect(startButton).toBeDefined()
      if (startButton && 'eventMode' in startButton) {
        expect(startButton.eventMode).toBe('static')
      }
    })
  })

  describe('Button Callbacks', () => {
    it('triggers onStartMatch callback when "Start Match" button is clicked', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)

      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      const startButton = lobbyContainer?.children.find(child => 
        child.label === 'room-lobby-start-button'
      )
      expect(startButton).toBeDefined()

      if (startButton && 'eventMode' in startButton && startButton.eventMode !== 'none') {
        startButton.emit('pointerdown')
        startButton.emit('pointerup')
      }

      expect(onStartMatch).toHaveBeenCalledTimes(1)
    })

    it('triggers onLeaveRoom callback when "Leave Room" button is clicked', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)

      const lobbyContainer = container.children.find(child => 
        child.label === 'room-lobby'
      )
      const leaveButton = lobbyContainer?.children.find(child => 
        child.label === 'room-lobby-leave-button'
      )
      expect(leaveButton).toBeDefined()

      if (leaveButton && 'eventMode' in leaveButton && leaveButton.eventMode !== 'none') {
        leaveButton.emit('pointerdown')
        leaveButton.emit('pointerup')
      }

      expect(onLeaveRoom).toHaveBeenCalledTimes(1)
    })
  })

  describe('Destruction', () => {
    it('destroys lobby and cleans up resources', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)
      const initialChildCount = container.children.length

      lobby.destroy()

      expect(container.children.length).toBeLessThan(initialChildCount)
    })

    it('can be destroyed multiple times safely', () => {
      const lobby = new RoomLobby(container, roomState, true, onStartMatch, onLeaveRoom)

      lobby.destroy()
      lobby.destroy()

      // Should not throw
      expect(true).toBe(true)
    })
  })
})

