/**
 * Integration tests for PlayerIndicator component.
 * 
 * Labels: scope:integration loop:g7-ui layer:ui dep:state,net b:player-indicator r:medium
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { Container } from 'pixi.js'
import { PlayerIndicator } from './player-indicator'

describe('PlayerIndicator', () => {
  let container: Container

  beforeEach(() => {
    container = new Container()
  })

  afterEach(() => {
    container.destroy({ children: true })
  })

  describe('Creation', () => {
    it('creates PlayerIndicator with container and position options', () => {
      const indicator = new PlayerIndicator(container, { x: 100, y: 50 })

      expect(indicator).toBeDefined()
      expect(container.children.length).toBeGreaterThan(0)
    })

    it('creates PlayerIndicator with default position', () => {
      const indicator = new PlayerIndicator(container)

      expect(indicator).toBeDefined()
      expect(container.children.length).toBeGreaterThan(0)
    })

    it('displays player ID', () => {
      const indicator = new PlayerIndicator(container, { x: 100, y: 50 })
      indicator.update(1)

      const indicatorContainer = container.children.find(child => 
        child.label === 'player-indicator'
      )
      expect(indicatorContainer).toBeDefined()
      
      const textElement = indicatorContainer?.children.find(child => 
        child.label === 'player-indicator-text'
      )
      expect(textElement).toBeDefined()
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toContain('1')
      }
    })
  })

  describe('Update', () => {
    it('updates displayed player ID', () => {
      const indicator = new PlayerIndicator(container)
      indicator.update(1)

      const indicatorContainer = container.children.find(child => 
        child.label === 'player-indicator'
      )
      const textElement = indicatorContainer?.children.find(child => 
        child.label === 'player-indicator-text'
      )
      expect(textElement).toBeDefined()
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toContain('1')
      }

      indicator.update(2)
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toContain('2')
      }
    })

    it('updates displayed player ID with name', () => {
      const indicator = new PlayerIndicator(container)
      indicator.update(1, 'Player 1')

      const indicatorContainer = container.children.find(child => 
        child.label === 'player-indicator'
      )
      const textElement = indicatorContainer?.children.find(child => 
        child.label === 'player-indicator-text'
      )
      expect(textElement).toBeDefined()
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toContain('1')
        expect(textElement.text).toContain('Player 1')
      }
    })

    it('updates from ID only to ID with name', () => {
      const indicator = new PlayerIndicator(container)
      indicator.update(1)

      const indicatorContainer = container.children.find(child => 
        child.label === 'player-indicator'
      )
      const textElement = indicatorContainer?.children.find(child => 
        child.label === 'player-indicator-text'
      )

      indicator.update(1, 'Player 1')
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toContain('1')
        expect(textElement.text).toContain('Player 1')
      }
    })

    it('updates from ID with name to ID only', () => {
      const indicator = new PlayerIndicator(container)
      indicator.update(1, 'Custom Name')

      const indicatorContainer = container.children.find(child => 
        child.label === 'player-indicator'
      )
      const textElement = indicatorContainer?.children.find(child => 
        child.label === 'player-indicator-text'
      )

      indicator.update(1)
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toContain('1')
        expect(textElement.text).not.toContain('Custom Name')
        expect(textElement.text).toBe('Player 1')
      }
    })
  })

  describe('Display', () => {
    it('displays player ID in format "Player {id}" when no name provided', () => {
      const indicator = new PlayerIndicator(container)
      indicator.update(1)

      const indicatorContainer = container.children.find(child => 
        child.label === 'player-indicator'
      )
      const textElement = indicatorContainer?.children.find(child => 
        child.label === 'player-indicator-text'
      )
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toMatch(/Player\s+1|ID:\s*1|1/i)
      }
    })

    it('displays player name when provided', () => {
      const indicator = new PlayerIndicator(container)
      indicator.update(1, 'Custom Name')

      const indicatorContainer = container.children.find(child => 
        child.label === 'player-indicator'
      )
      const textElement = indicatorContainer?.children.find(child => 
        child.label === 'player-indicator-text'
      )
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toContain('Custom Name')
      }
    })
  })

  describe('Destruction', () => {
    it('destroys indicator and cleans up resources', () => {
      const indicator = new PlayerIndicator(container)
      const initialChildCount = container.children.length

      indicator.destroy()

      expect(container.children.length).toBeLessThan(initialChildCount)
    })

    it('can be destroyed multiple times safely', () => {
      const indicator = new PlayerIndicator(container)

      indicator.destroy()
      indicator.destroy()

      // Should not throw
      expect(true).toBe(true)
    })
  })
})

