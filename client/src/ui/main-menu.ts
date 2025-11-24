/**
 * Main menu component for game entry point.
 * 
 * Labels: scope:integration loop:g7-ui layer:ui dep:state,net b:main-menu r:high
 */

import { Container, Graphics, Text } from 'pixi.js'

/**
 * Main menu UI component with "Create Room" and "Join Room" options.
 */
export class MainMenu {
  private container: Container
  private onCreateRoom: () => void
  private onJoinRoom: (code: string) => void
  private titleText: Text
  private createButton: Graphics
  private createButtonText: Text
  private joinButton: Graphics
  private joinButtonText: Text
  private inputContainer: Container | null = null
  private inputField: Graphics | null = null
  private inputText: Text | null = null
  private submitButton: Graphics | null = null
  private submitButtonText: Text | null = null
  private roomCodeInput: string = ''
  private isInputVisible: boolean = false

  constructor(
    parent: Container,
    onCreateRoom: () => void,
    onJoinRoom: (code: string) => void
  ) {
    this.container = new Container()
    this.container.label = 'main-menu'
    parent.addChild(this.container)

    this.onCreateRoom = onCreateRoom
    this.onJoinRoom = onJoinRoom

    // Create title
    this.titleText = new Text({
      text: 'Orbital Rush',
      style: {
        fontFamily: 'Arial',
        fontSize: 48,
        fill: 0xffffff,
        align: 'center'
      }
    })
    this.titleText.label = 'main-menu-title'
    this.titleText.anchor.set(0.5)
    this.titleText.x = 0
    this.titleText.y = -150
    this.container.addChild(this.titleText)

    // Create "Create Room" button
    this.createButton = new Graphics()
    this.createButton.label = 'main-menu-create-button'
    this.createButton.eventMode = 'static'
    this.createButton.cursor = 'pointer'
    this.createButton.rect(-100, -20, 200, 40)
    this.createButton.fill(0x4a90e2)
    this.createButton.x = 0
    this.createButton.y = -50
    this.createButton.on('pointerdown', () => {
      this.createButton.clear()
      this.createButton.rect(-100, -20, 200, 40)
      this.createButton.fill(0x357abd)
    })
    this.createButton.on('pointerup', () => {
      this.createButton.clear()
      this.createButton.rect(-100, -20, 200, 40)
      this.createButton.fill(0x4a90e2)
      this.onCreateRoom()
    })
    this.createButton.on('pointerupoutside', () => {
      this.createButton.clear()
      this.createButton.rect(-100, -20, 200, 40)
      this.createButton.fill(0x4a90e2)
    })
    this.container.addChild(this.createButton)

    this.createButtonText = new Text({
      text: 'Create Room',
      style: {
        fontFamily: 'Arial',
        fontSize: 20,
        fill: 0xffffff
      }
    })
    this.createButtonText.anchor.set(0.5)
    this.createButtonText.x = 0
    this.createButtonText.y = -50
    this.container.addChild(this.createButtonText)

    // Create "Join Room" button
    this.joinButton = new Graphics()
    this.joinButton.label = 'main-menu-join-button'
    this.joinButton.eventMode = 'static'
    this.joinButton.cursor = 'pointer'
    this.joinButton.rect(-100, -20, 200, 40)
    this.joinButton.fill(0x50c878)
    this.joinButton.x = 0
    this.joinButton.y = 20
    this.joinButton.on('pointerdown', () => {
      this.joinButton.clear()
      this.joinButton.rect(-100, -20, 200, 40)
      this.joinButton.fill(0x3da05f)
    })
    this.joinButton.on('pointerup', () => {
      this.joinButton.clear()
      this.joinButton.rect(-100, -20, 200, 40)
      this.joinButton.fill(0x50c878)
      this.showInputField()
    })
    this.joinButton.on('pointerupoutside', () => {
      this.joinButton.clear()
      this.joinButton.rect(-100, -20, 200, 40)
      this.joinButton.fill(0x50c878)
    })
    this.container.addChild(this.joinButton)

    this.joinButtonText = new Text({
      text: 'Join Room',
      style: {
        fontFamily: 'Arial',
        fontSize: 20,
        fill: 0xffffff
      }
    })
    this.joinButtonText.anchor.set(0.5)
    this.joinButtonText.x = 0
    this.joinButtonText.y = 20
    this.container.addChild(this.joinButtonText)

    // Set container to center (will be positioned by parent)
    this.container.x = 0
    this.container.y = 0
  }

  /**
   * Shows the input field for room code entry.
   */
  private showInputField(): void {
    if (this.isInputVisible) {
      return
    }

    this.isInputVisible = true

    // Hide main buttons
    this.createButton.visible = false
    this.createButtonText.visible = false
    this.joinButton.visible = false
    this.joinButtonText.visible = false

    // Create input container
    this.inputContainer = new Container()
    this.inputContainer.label = 'main-menu-input-container'
    this.inputContainer.y = 20

    // Create input field background
    this.inputField = new Graphics()
    this.inputField.label = 'main-menu-join-input'
    this.inputField.rect(-120, -15, 240, 30)
    this.inputField.fill(0x333333)
    this.inputField.stroke({ width: 2, color: 0xffffff })
    this.inputField.x = 0
    this.inputField.y = 0
    this.inputContainer.addChild(this.inputField)

    // Create input text
    this.inputText = new Text({
      text: '',
      style: {
        fontFamily: 'Arial',
        fontSize: 18,
        fill: 0xffffff
      }
    })
    this.inputText.label = 'main-menu-join-input-text'
    this.inputText.anchor.set(0, 0.5)
    this.inputText.x = -110
    this.inputText.y = 0
    this.inputContainer.addChild(this.inputText)

    // Create placeholder text
    const placeholderText = new Text({
      text: 'Enter room code...',
      style: {
        fontFamily: 'Arial',
        fontSize: 18,
        fill: 0x888888
      }
    })
    placeholderText.label = 'main-menu-join-placeholder'
    placeholderText.anchor.set(0, 0.5)
    placeholderText.x = -110
    placeholderText.y = 0
    this.inputContainer.addChild(placeholderText)

    // Create submit button
    this.submitButton = new Graphics()
    this.submitButton.label = 'main-menu-join-submit'
    this.submitButton.eventMode = 'static'
    this.submitButton.cursor = 'pointer'
    this.submitButton.rect(-60, -15, 120, 30)
    this.submitButton.fill(0x50c878)
    this.submitButton.x = 0
    this.submitButton.y = 50
    this.submitButton.on('pointerdown', () => {
      this.submitButton!.clear()
      this.submitButton!.rect(-60, -15, 120, 30)
      this.submitButton!.fill(0x3da05f)
    })
    this.submitButton.on('pointerup', () => {
      this.submitButton!.clear()
      this.submitButton!.rect(-60, -15, 120, 30)
      this.submitButton!.fill(0x50c878)
      this.submitRoomCode()
    })
    this.submitButton.on('pointerupoutside', () => {
      this.submitButton!.clear()
      this.submitButton!.rect(-60, -15, 120, 30)
      this.submitButton!.fill(0x50c878)
    })
    this.inputContainer.addChild(this.submitButton)

    this.submitButtonText = new Text({
      text: 'Join',
      style: {
        fontFamily: 'Arial',
        fontSize: 18,
        fill: 0xffffff
      }
    })
    this.submitButtonText.anchor.set(0.5)
    this.submitButtonText.x = 0
    this.submitButtonText.y = 50
    this.inputContainer.addChild(this.submitButtonText)

    // Create cancel button
    const cancelButton = new Graphics()
    cancelButton.label = 'main-menu-join-cancel'
    cancelButton.eventMode = 'static'
    cancelButton.cursor = 'pointer'
    cancelButton.rect(-60, -15, 120, 30)
    cancelButton.fill(0x888888)
    cancelButton.x = 0
    cancelButton.y = 90
    cancelButton.on('pointerdown', () => {
      cancelButton.clear()
      cancelButton.rect(-60, -15, 120, 30)
      cancelButton.fill(0x666666)
    })
    cancelButton.on('pointerup', () => {
      cancelButton.clear()
      cancelButton.rect(-60, -15, 120, 30)
      cancelButton.fill(0x888888)
      this.hideInputField()
    })
    cancelButton.on('pointerupoutside', () => {
      cancelButton.clear()
      cancelButton.rect(-60, -15, 120, 30)
      cancelButton.fill(0x888888)
    })
    this.inputContainer.addChild(cancelButton)

    const cancelButtonText = new Text({
      text: 'Cancel',
      style: {
        fontFamily: 'Arial',
        fontSize: 18,
        fill: 0xffffff
      }
    })
    cancelButtonText.anchor.set(0.5)
    cancelButtonText.x = 0
    cancelButtonText.y = 90
    this.inputContainer.addChild(cancelButtonText)

    this.container.addChild(this.inputContainer)

    // Set up keyboard input handling
    this.setupKeyboardInput(placeholderText)
  }

  /**
   * Sets up keyboard input handling for room code entry.
   */
  private setupKeyboardInput(placeholderText: Text): void {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (!this.isInputVisible || !this.inputText) {
        return
      }

      if (event.key === 'Enter') {
        this.submitRoomCode()
        return
      }

      if (event.key === 'Escape') {
        this.hideInputField()
        return
      }

      if (event.key === 'Backspace') {
        this.roomCodeInput = this.roomCodeInput.slice(0, -1)
        this.updateInputText(placeholderText)
        return
      }

      // Only allow alphanumeric characters and limit to 6 characters
      if (/^[a-zA-Z0-9]$/.test(event.key) && this.roomCodeInput.length < 6) {
        this.roomCodeInput += event.key.toUpperCase()
        this.updateInputText(placeholderText)
      }
    }

    window.addEventListener('keydown', handleKeyDown)

    // Store handler for cleanup
    ;(this.inputContainer as any)._keyHandler = handleKeyDown
  }

  /**
   * Updates the input text display.
   */
  private updateInputText(placeholderText: Text): void {
    if (!this.inputText) {
      return
    }

    if (this.roomCodeInput.length > 0) {
      this.inputText.text = this.roomCodeInput
      placeholderText.visible = false
    } else {
      this.inputText.text = ''
      placeholderText.visible = true
    }
  }

  /**
   * Submits the room code and triggers onJoinRoom callback.
   */
  private submitRoomCode(): void {
    if (this.roomCodeInput.length > 0) {
      this.onJoinRoom(this.roomCodeInput)
      this.hideInputField()
    }
  }

  /**
   * Hides the input field and shows main buttons.
   */
  private hideInputField(): void {
    if (!this.isInputVisible) {
      return
    }

    this.isInputVisible = false
    this.roomCodeInput = ''

    // Remove keyboard handler
    if (this.inputContainer && (this.inputContainer as any)._keyHandler) {
      window.removeEventListener('keydown', (this.inputContainer as any)._keyHandler)
    }

    // Remove input container
    if (this.inputContainer && this.inputContainer.parent) {
      this.inputContainer.parent.removeChild(this.inputContainer)
      this.inputContainer.destroy({ children: true })
    }

    this.inputContainer = null
    this.inputField = null
    this.inputText = null
    this.submitButton = null
    this.submitButtonText = null

    // Show main buttons
    this.createButton.visible = true
    this.createButtonText.visible = true
    this.joinButton.visible = true
    this.joinButtonText.visible = true
  }

  /**
   * Shows the main menu.
   */
  show(): void {
    this.container.visible = true
  }

  /**
   * Hides the main menu.
   */
  hide(): void {
    this.container.visible = false
    // Also hide input field if visible
    if (this.isInputVisible) {
      this.hideInputField()
    }
  }

  /**
   * Destroys the main menu and cleans up resources.
   */
  destroy(): void {
    // Clean up keyboard handler
    if (this.inputContainer && (this.inputContainer as any)._keyHandler) {
      window.removeEventListener('keydown', (this.inputContainer as any)._keyHandler)
    }

    // Hide input field if visible
    if (this.isInputVisible) {
      this.hideInputField()
    }

    if (this.container.parent) {
      this.container.parent.removeChild(this.container)
    }
    this.container.destroy({ children: true })
  }
}

