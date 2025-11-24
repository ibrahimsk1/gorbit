/**
 * Room lobby component for displaying room state and player management.
 * 
 * Labels: scope:integration loop:g7-ui layer:ui dep:state,net b:room-lobby r:high
 */

import { Container, Graphics, Text } from 'pixi.js'
import type { RoomState } from '../core/orchestrator'

/**
 * Room lobby UI component showing room code, player list, and match controls.
 */
export class RoomLobby {
  private container: Container
  private roomState: RoomState
  private isHost: boolean
  private onStartMatch: () => void
  private onLeaveRoom: () => void
  private roomCodeText: Text
  private playerListContainer: Container
  private startButton: Graphics | null = null
  private startButtonText: Text | null = null
  private leaveButton: Graphics
  private leaveButtonText: Text
  private waitingText: Text | null = null
  private playerTexts: Text[] = []

  constructor(
    parent: Container,
    roomState: RoomState,
    isHost: boolean,
    onStartMatch: () => void,
    onLeaveRoom: () => void
  ) {
    this.container = new Container()
    this.container.label = 'room-lobby'
    this.container.visible = false // Hidden by default
    parent.addChild(this.container)

    this.roomState = roomState
    this.isHost = isHost
    this.onStartMatch = onStartMatch
    this.onLeaveRoom = onLeaveRoom

    // Create room code display (prominently displayed, large text)
    this.roomCodeText = new Text({
      text: `Room Code: ${roomState.roomCode}`,
      style: {
        fontFamily: 'Arial',
        fontSize: 36,
        fill: 0xffffff,
        align: 'center'
      }
    })
    this.roomCodeText.label = 'room-lobby-code'
    this.roomCodeText.anchor.set(0.5)
    this.roomCodeText.x = 0
    this.roomCodeText.y = -200
    this.container.addChild(this.roomCodeText)

    // Create player list container
    this.playerListContainer = new Container()
    this.playerListContainer.label = 'room-lobby-players'
    this.playerListContainer.x = 0
    this.playerListContainer.y = -100
    this.container.addChild(this.playerListContainer)

    // Create player list title
    const playerListTitle = new Text({
      text: 'Players:',
      style: {
        fontFamily: 'Arial',
        fontSize: 24,
        fill: 0xffffff
      }
    })
    playerListTitle.anchor.set(0.5)
    playerListTitle.x = 0
    playerListTitle.y = 0
    this.playerListContainer.addChild(playerListTitle)

    // Create player list items
    this.updatePlayerList()

    // Create "Start Match" button (only for host)
    if (this.isHost) {
      this.startButton = new Graphics()
      this.startButton.label = 'room-lobby-start-button'
      const canStart = this.roomState.players.length >= 2
      this.startButton.eventMode = canStart ? 'static' : 'none'
      this.startButton.cursor = canStart ? 'pointer' : 'default'
      this.startButton.rect(-100, -20, 200, 40)
      this.startButton.fill(canStart ? 0x4a90e2 : 0x666666)
      this.startButton.x = 0
      this.startButton.y = 50
      if (canStart) {
        this.startButton.on('pointerdown', () => {
          this.startButton!.clear()
          this.startButton!.rect(-100, -20, 200, 40)
          this.startButton!.fill(0x357abd)
        })
        this.startButton.on('pointerup', () => {
          this.startButton!.clear()
          this.startButton!.rect(-100, -20, 200, 40)
          this.startButton!.fill(0x4a90e2)
          this.onStartMatch()
        })
        this.startButton.on('pointerupoutside', () => {
          this.startButton!.clear()
          this.startButton!.rect(-100, -20, 200, 40)
          this.startButton!.fill(0x4a90e2)
        })
      }
      this.container.addChild(this.startButton)

      this.startButtonText = new Text({
        text: canStart ? 'Start Match' : 'Start Match (Need 2+ Players)',
        style: {
          fontFamily: 'Arial',
          fontSize: 20,
          fill: 0xffffff
        }
      })
      this.startButtonText.anchor.set(0.5)
      this.startButtonText.x = 0
      this.startButtonText.y = 50
      this.container.addChild(this.startButtonText)
    }

    // Create "Leave Room" button (always visible)
    this.leaveButton = new Graphics()
    this.leaveButton.label = 'room-lobby-leave-button'
    this.leaveButton.eventMode = 'static'
    this.leaveButton.cursor = 'pointer'
    this.leaveButton.rect(-100, -20, 200, 40)
    this.leaveButton.fill(0xff4444)
    this.leaveButton.x = 0
    this.leaveButton.y = 110
    this.leaveButton.on('pointerdown', () => {
      this.leaveButton.clear()
      this.leaveButton.rect(-100, -20, 200, 40)
      this.leaveButton.fill(0xcc3333)
    })
    this.leaveButton.on('pointerup', () => {
      this.leaveButton.clear()
      this.leaveButton.rect(-100, -20, 200, 40)
      this.leaveButton.fill(0xff4444)
      this.onLeaveRoom()
    })
    this.leaveButton.on('pointerupoutside', () => {
      this.leaveButton.clear()
      this.leaveButton.rect(-100, -20, 200, 40)
      this.leaveButton.fill(0xff4444)
    })
    this.container.addChild(this.leaveButton)

    this.leaveButtonText = new Text({
      text: 'Leave Room',
      style: {
        fontFamily: 'Arial',
        fontSize: 20,
        fill: 0xffffff
      }
    })
    this.leaveButtonText.anchor.set(0.5)
    this.leaveButtonText.x = 0
    this.leaveButtonText.y = 110
    this.container.addChild(this.leaveButtonText)

    // Create waiting message (for non-host)
    if (!this.isHost) {
      this.waitingText = new Text({
        text: 'Waiting for host to start match...',
        style: {
          fontFamily: 'Arial',
          fontSize: 18,
          fill: 0xaaaaaa
        }
      })
      this.waitingText.label = 'room-lobby-waiting'
      this.waitingText.anchor.set(0.5)
      this.waitingText.x = 0
      this.waitingText.y = 50
      this.container.addChild(this.waitingText)
    }

    // Set container to center (will be positioned by parent)
    this.container.x = 0
    this.container.y = 0
  }

  /**
   * Updates the player list display.
   */
  private updatePlayerList(): void {
    // Clear existing player texts
    for (const text of this.playerTexts) {
      if (text.parent) {
        text.parent.removeChild(text)
      }
      text.destroy()
    }
    this.playerTexts = []

    // Create player list items
    this.roomState.players.forEach((player, index) => {
      const playerText = new Text({
        text: `${player.name} (ID: ${player.id})${player.id === this.roomState.hostId ? ' [Host]' : ''}`,
        style: {
          fontFamily: 'Arial',
          fontSize: 18,
          fill: player.id === this.roomState.hostId ? 0xffff00 : 0xffffff
        }
      })
      playerText.anchor.set(0.5)
      playerText.x = 0
      playerText.y = 40 + (index * 30)
      this.playerListContainer.addChild(playerText)
      this.playerTexts.push(playerText)
    })
  }

  /**
   * Shows the lobby.
   */
  show(): void {
    this.container.visible = true
  }

  /**
   * Hides the lobby.
   */
  hide(): void {
    this.container.visible = false
  }

  /**
   * Updates the lobby with new room state.
   * 
   * @param roomState New room state
   */
  update(roomState: RoomState): void {
    this.roomState = roomState

    // Update room code
    this.roomCodeText.text = `Room Code: ${roomState.roomCode}`

    // Update player list
    this.updatePlayerList()

    // Update "Start Match" button state (if host)
    if (this.isHost && this.startButton && this.startButtonText) {
      const canStart = roomState.players.length >= 2
      this.startButton.eventMode = canStart ? 'static' : 'none'
      this.startButton.cursor = canStart ? 'pointer' : 'default'
      
      // Update button color
      this.startButton.clear()
      this.startButton.rect(-100, -20, 200, 40)
      this.startButton.fill(canStart ? 0x4a90e2 : 0x666666)

      // Update button text
      this.startButtonText.text = canStart ? 'Start Match' : 'Start Match (Need 2+ Players)'

      // Re-setup event handlers if enabled
      if (canStart) {
        this.startButton.removeAllListeners()
        this.startButton.on('pointerdown', () => {
          this.startButton!.clear()
          this.startButton!.rect(-100, -20, 200, 40)
          this.startButton!.fill(0x357abd)
        })
        this.startButton.on('pointerup', () => {
          this.startButton!.clear()
          this.startButton!.rect(-100, -20, 200, 40)
          this.startButton!.fill(0x4a90e2)
          this.onStartMatch()
        })
        this.startButton.on('pointerupoutside', () => {
          this.startButton!.clear()
          this.startButton!.rect(-100, -20, 200, 40)
          this.startButton!.fill(0x4a90e2)
        })
      }
    }

    // Update host status (if changed)
    // Note: isHost is set in constructor, but we can check if current player is host
    // For now, we'll keep the original isHost value
  }

  /**
   * Destroys the lobby and cleans up resources.
   */
  destroy(): void {
    // Clean up player texts
    for (const text of this.playerTexts) {
      if (text.parent) {
        text.parent.removeChild(text)
      }
      text.destroy()
    }
    this.playerTexts = []

    if (this.container.parent) {
      this.container.parent.removeChild(this.container)
    }
    this.container.destroy({ children: true })
  }
}

