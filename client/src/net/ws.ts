/**
 * WebSocket client wrapper with connection lifecycle management.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:ws
 */

export class WebSocketClient {
  private ws: WebSocket | null = null
  private url: string | null = null
  private messageHandlers: Array<(data: any) => void> = []
  private openHandlers: Array<() => void> = []
  private closeHandlers: Array<() => void> = []
  private errorHandlers: Array<(error: Error) => void> = []

  /**
   * Connect to WebSocket server at the specified URL.
   * @param url WebSocket server URL (e.g., 'ws://localhost:8080/ws')
   * @returns Promise that resolves when connection is established
   */
  async connect(url: string): Promise<void> {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      // Already connected
      return
    }

    this.url = url

    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(url)

        this.ws.onopen = () => {
          this.openHandlers.forEach(handler => handler())
          resolve()
        }

        this.ws.onclose = () => {
          this.closeHandlers.forEach(handler => handler())
        }

        this.ws.onerror = (event) => {
          const error = new Error('WebSocket error')
          this.errorHandlers.forEach(handler => handler(error))
          reject(error)
        }

        this.ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data)
            this.messageHandlers.forEach(handler => handler(data))
          } catch (error) {
            // Malformed JSON - emit error
            const err = error instanceof Error ? error : new Error('Failed to parse message')
            this.errorHandlers.forEach(handler => handler(err))
          }
        }
      } catch (error) {
        const err = error instanceof Error ? error : new Error('Failed to create WebSocket')
        reject(err)
      }
    })
  }

  /**
   * Disconnect from WebSocket server.
   */
  disconnect(): void {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  /**
   * Send a message over the WebSocket connection.
   * @param message Message object to send (will be serialized to JSON)
   * @throws Error if not connected
   */
  send(message: object): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket is not connected')
    }

    const json = JSON.stringify(message)
    this.ws.send(json)
  }

  /**
   * Register a callback for incoming messages.
   * @param callback Function to call when a message is received
   */
  onMessage(callback: (data: any) => void): void {
    this.messageHandlers.push(callback)
  }

  /**
   * Register a callback for connection open events.
   * @param callback Function to call when connection is opened
   */
  onOpen(callback: () => void): void {
    this.openHandlers.push(callback)
  }

  /**
   * Register a callback for connection close events.
   * @param callback Function to call when connection is closed
   */
  onClose(callback: () => void): void {
    this.closeHandlers.push(callback)
  }

  /**
   * Register a callback for error events.
   * @param callback Function to call when an error occurs
   */
  onError(callback: (error: Error) => void): void {
    this.errorHandlers.push(callback)
  }

  /**
   * Check if WebSocket is connected.
   * @returns true if connected, false otherwise
   */
  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN
  }

  /**
   * Get WebSocket ready state.
   * @returns WebSocket ready state (CONNECTING, OPEN, CLOSING, CLOSED)
   */
  getReadyState(): number {
    return this.ws?.readyState ?? WebSocket.CLOSED
  }
}

