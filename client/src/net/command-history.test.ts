/**
 * Integration tests for command history tracker.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import { describe, it, expect, beforeEach } from 'vitest'
import { CommandHistory } from './command-history'
import type { InputMessage } from './protocol'

describe('CommandHistory', () => {
  let history: CommandHistory

  beforeEach(() => {
    history = new CommandHistory()
  })

  describe('Sequence Number Generation', () => {
    it('starts sequence numbers at 1', () => {
      const seq = history.getNextSequence()
      expect(seq).toBe(1)
    })

    it('increments sequence numbers for each command', () => {
      history.addCommand(1, 0.5, 0.3)
      const seq2 = history.getNextSequence()
      expect(seq2).toBe(2)

      history.addCommand(2, 0.8, -0.5)
      const seq3 = history.getNextSequence()
      expect(seq3).toBe(3)
    })

    it('maintains sequence continuity after adding commands', () => {
      history.addCommand(1, 0.5, 0.3)
      history.addCommand(2, 0.8, -0.5)
      history.addCommand(3, 0.2, 0.1)

      const seq4 = history.getNextSequence()
      expect(seq4).toBe(4)
    })
  })

  describe('Adding Commands', () => {
    it('adds command to history with sequence number', () => {
      history.addCommand(1, 0.5, 0.3)

      const command = history.getCommand(1)
      expect(command).not.toBeNull()
      expect(command?.seq).toBe(1)
      expect(command?.thrust).toBe(0.5)
      expect(command?.turn).toBe(0.3)
    })

    it('adds multiple commands to history', () => {
      history.addCommand(1, 0.5, 0.3)
      history.addCommand(2, 0.8, -0.5)
      history.addCommand(3, 0.2, 0.1)

      expect(history.getCommand(1)).not.toBeNull()
      expect(history.getCommand(2)).not.toBeNull()
      expect(history.getCommand(3)).not.toBeNull()
    })

    it('stores commands as unconfirmed by default', () => {
      history.addCommand(1, 0.5, 0.3)
      history.addCommand(2, 0.8, -0.5)

      const unconfirmed = history.getUnconfirmed()
      expect(unconfirmed).toHaveLength(2)
      expect(unconfirmed[0].seq).toBe(1)
      expect(unconfirmed[1].seq).toBe(2)
    })

    it('tracks timestamp for each command', () => {
      const before = Date.now()
      history.addCommand(1, 0.5, 0.3)
      const after = Date.now()

      const command = history.getCommand(1)
      expect(command).not.toBeNull()
      expect(command?.timestamp).toBeGreaterThanOrEqual(before)
      expect(command?.timestamp).toBeLessThanOrEqual(after)
    })
  })

  describe('Command Confirmation', () => {
    it('marks single command as confirmed', () => {
      history.addCommand(1, 0.5, 0.3)
      history.addCommand(2, 0.8, -0.5)

      history.markConfirmed(1)

      const unconfirmed = history.getUnconfirmed()
      expect(unconfirmed).toHaveLength(1)
      expect(unconfirmed[0].seq).toBe(2)
    })

    it('marks multiple commands as confirmed up to sequence', () => {
      history.addCommand(1, 0.5, 0.3)
      history.addCommand(2, 0.8, -0.5)
      history.addCommand(3, 0.2, 0.1)
      history.addCommand(4, 0.9, -0.2)

      history.markConfirmedUpTo(2)

      const unconfirmed = history.getUnconfirmed()
      expect(unconfirmed).toHaveLength(2)
      expect(unconfirmed[0].seq).toBe(3)
      expect(unconfirmed[1].seq).toBe(4)
    })

    it('marks all commands up to sequence as confirmed', () => {
      history.addCommand(1, 0.5, 0.3)
      history.addCommand(2, 0.8, -0.5)
      history.addCommand(3, 0.2, 0.1)

      history.markConfirmedUpTo(3)

      const unconfirmed = history.getUnconfirmed()
      expect(unconfirmed).toHaveLength(0)
    })

    it('handles marking non-existent sequence as confirmed', () => {
      history.addCommand(1, 0.5, 0.3)

      // Should not throw
      history.markConfirmed(999)
      // markConfirmedUpTo should mark command 1 since 1 <= 999
      history.markConfirmedUpTo(999)

      const unconfirmed = history.getUnconfirmed()
      expect(unconfirmed).toHaveLength(0)
    })

    it('handles marking commands that are already confirmed', () => {
      history.addCommand(1, 0.5, 0.3)
      history.addCommand(2, 0.8, -0.5)

      history.markConfirmed(1)
      history.markConfirmed(1) // Mark again

      const unconfirmed = history.getUnconfirmed()
      expect(unconfirmed).toHaveLength(1)
      expect(unconfirmed[0].seq).toBe(2)
    })
  })

  describe('Querying Unconfirmed Commands', () => {
    it('returns empty array when no commands exist', () => {
      const unconfirmed = history.getUnconfirmed()
      expect(unconfirmed).toHaveLength(0)
    })

    it('returns all commands when none are confirmed', () => {
      history.addCommand(1, 0.5, 0.3)
      history.addCommand(2, 0.8, -0.5)
      history.addCommand(3, 0.2, 0.1)

      const unconfirmed = history.getUnconfirmed()
      expect(unconfirmed).toHaveLength(3)
      expect(unconfirmed.map(c => c.seq)).toEqual([1, 2, 3])
    })

    it('returns only unconfirmed commands in sequence order', () => {
      history.addCommand(1, 0.5, 0.3)
      history.addCommand(2, 0.8, -0.5)
      history.addCommand(3, 0.2, 0.1)
      history.addCommand(4, 0.9, -0.2)

      history.markConfirmedUpTo(2)

      const unconfirmed = history.getUnconfirmed()
      expect(unconfirmed).toHaveLength(2)
      expect(unconfirmed[0].seq).toBe(3)
      expect(unconfirmed[1].seq).toBe(4)
    })

    it('returns empty array when all commands are confirmed', () => {
      history.addCommand(1, 0.5, 0.3)
      history.addCommand(2, 0.8, -0.5)

      history.markConfirmedUpTo(2)

      const unconfirmed = history.getUnconfirmed()
      expect(unconfirmed).toHaveLength(0)
    })
  })

  describe('Command Lookup', () => {
    it('returns command by sequence number', () => {
      history.addCommand(1, 0.5, 0.3)
      history.addCommand(2, 0.8, -0.5)

      const command = history.getCommand(1)
      expect(command).not.toBeNull()
      expect(command?.seq).toBe(1)
      expect(command?.thrust).toBe(0.5)
      expect(command?.turn).toBe(0.3)
    })

    it('returns null for non-existent sequence number', () => {
      history.addCommand(1, 0.5, 0.3)

      const command = history.getCommand(999)
      expect(command).toBeNull()
    })

    it('returns command even after confirmation', () => {
      history.addCommand(1, 0.5, 0.3)
      history.markConfirmed(1)

      const command = history.getCommand(1)
      expect(command).not.toBeNull()
      expect(command?.seq).toBe(1)
    })
  })

  describe('History Cleanup', () => {
    it('clears all commands from history', () => {
      history.addCommand(1, 0.5, 0.3)
      history.addCommand(2, 0.8, -0.5)

      history.clear()

      expect(history.getUnconfirmed()).toHaveLength(0)
      expect(history.getCommand(1)).toBeNull()
      expect(history.getCommand(2)).toBeNull()
    })

    it('resets sequence number after clear', () => {
      history.addCommand(1, 0.5, 0.3)
      history.addCommand(2, 0.8, -0.5)

      history.clear()

      const nextSeq = history.getNextSequence()
      expect(nextSeq).toBe(1)
    })

    it('allows adding commands after clear', () => {
      history.addCommand(1, 0.5, 0.3)
      history.clear()

      history.addCommand(1, 0.8, -0.5)
      const command = history.getCommand(1)
      expect(command).not.toBeNull()
      expect(command?.thrust).toBe(0.8)
    })
  })

  describe('Edge Cases', () => {
    it('handles empty history gracefully', () => {
      expect(history.getUnconfirmed()).toHaveLength(0)
      expect(history.getCommand(1)).toBeNull()
      expect(history.getNextSequence()).toBe(1)
    })

    it('handles large sequence numbers', () => {
      history.addCommand(1000, 0.5, 0.3)
      history.addCommand(2000, 0.8, -0.5)

      expect(history.getCommand(1000)).not.toBeNull()
      expect(history.getCommand(2000)).not.toBeNull()

      const unconfirmed = history.getUnconfirmed()
      expect(unconfirmed).toHaveLength(2)
    })

    it('maintains command order regardless of confirmation order', () => {
      history.addCommand(1, 0.5, 0.3)
      history.addCommand(2, 0.8, -0.5)
      history.addCommand(3, 0.2, 0.1)

      // Confirm out of order
      history.markConfirmed(3)
      history.markConfirmed(1)

      const unconfirmed = history.getUnconfirmed()
      expect(unconfirmed).toHaveLength(1)
      expect(unconfirmed[0].seq).toBe(2)
    })
  })
})

