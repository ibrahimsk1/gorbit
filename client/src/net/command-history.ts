/**
 * Command history tracker for reconciliation with sequence numbers.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

/**
 * Command entry stored in history with sequence number, command data, and confirmation status.
 */
export interface CommandEntry {
  seq: number
  thrust: number
  turn: number
  timestamp: number
  confirmed: boolean
}

/**
 * Command history tracker that maintains sent commands with sequence numbers
 * for client-side prediction and server reconciliation.
 * 
 * - Tracks sent commands with sequence numbers
 * - Marks commands as confirmed when server acknowledges them
 * - Provides unconfirmed commands for re-application after rollback
 * - Generates sequence numbers starting at 1
 */
export class CommandHistory {
  private commands: Map<number, CommandEntry> = new Map()
  private nextSequence: number = 1

  /**
   * Gets the next sequence number for a new command.
   * Sequence numbers start at 1 and increment for each command.
   */
  getNextSequence(): number {
    return this.nextSequence
  }

  /**
   * Adds a command to the history.
   * @param seq Sequence number for the command
   * @param thrust Thrust value (0.0 to 1.0)
   * @param turn Turn value (-1.0 to 1.0)
   */
  addCommand(seq: number, thrust: number, turn: number): void {
    const entry: CommandEntry = {
      seq,
      thrust,
      turn,
      timestamp: Date.now(),
      confirmed: false
    }
    this.commands.set(seq, entry)

    // Update next sequence to be one more than the highest sequence
    if (seq >= this.nextSequence) {
      this.nextSequence = seq + 1
    }
  }

  /**
   * Marks a single command as confirmed.
   * @param seq Sequence number of the command to confirm
   */
  markConfirmed(seq: number): void {
    const entry = this.commands.get(seq)
    if (entry) {
      entry.confirmed = true
    }
  }

  /**
   * Marks all commands up to and including the specified sequence as confirmed.
   * This is useful when server acknowledges commands up to a certain tick.
   * Only marks commands that actually exist in the history.
   * @param seq Sequence number up to which commands should be confirmed
   */
  markConfirmedUpTo(seq: number): void {
    // Only mark commands that exist in history
    for (const [commandSeq, entry] of this.commands.entries()) {
      if (commandSeq <= seq) {
        entry.confirmed = true
      }
    }
  }

  /**
   * Gets all unconfirmed commands in sequence order.
   * These commands are used for re-application after rollback during reconciliation.
   * @returns Array of unconfirmed commands sorted by sequence number
   */
  getUnconfirmed(): CommandEntry[] {
    const unconfirmed: CommandEntry[] = []
    for (const entry of this.commands.values()) {
      if (!entry.confirmed) {
        unconfirmed.push(entry)
      }
    }
    // Sort by sequence number
    unconfirmed.sort((a, b) => a.seq - b.seq)
    return unconfirmed
  }

  /**
   * Gets a command by its sequence number.
   * @param seq Sequence number of the command
   * @returns Command entry or null if not found
   */
  getCommand(seq: number): CommandEntry | null {
    return this.commands.get(seq) || null
  }

  /**
   * Clears all commands from history and resets sequence number to 1.
   * Useful when starting a new game or resetting state.
   */
  clear(): void {
    this.commands.clear()
    this.nextSequence = 1
  }
}

