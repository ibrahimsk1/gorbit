/**
 * Integration tests for Game Banner component.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { Container } from 'pixi.js'
import { GameBanner } from './game-banner'

describe('Game Banner', () => {
  let container: Container

  beforeEach(() => {
    container = new Container()
  })

  afterEach(() => {
    container.destroy({ children: true })
  })

  describe('Creation', () => {
    it('creates game banner with default configuration', () => {
      const banner = new GameBanner(container)

      expect(banner).toBeDefined()
      // Banner should be hidden initially
      expect(container.children.length).toBeGreaterThan(0)
    })

    it('creates game banner with custom configuration', () => {
      const banner = new GameBanner(container, {
        winMessage: 'Victory!',
        loseMessage: 'Defeat!'
      })

      expect(banner).toBeDefined()
    })
  })

  describe('Win Message', () => {
    it('shows win message when showWin() called', () => {
      const banner = new GameBanner(container)

      banner.showWin()

      const bannerContainer = container.children.find(child => 
        child.label === 'game-banner'
      )
      expect(bannerContainer).toBeDefined()
      const textElement = bannerContainer?.children.find(child => 
        child.label === 'game-banner-text'
      )
      expect(textElement).toBeDefined()
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toContain('Win')
      }
    })

    it('displays correct win message text', () => {
      const banner = new GameBanner(container, {
        winMessage: 'You Win!'
      })

      banner.showWin()

      const bannerContainer = container.children.find(child => 
        child.label === 'game-banner'
      )
      const textElement = bannerContainer?.children.find(child => 
        child.label === 'game-banner-text'
      )
      expect(textElement).toBeDefined()
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toBe('You Win!')
      }
    })
  })

  describe('Lose Message', () => {
    it('shows lose message when showLose() called', () => {
      const banner = new GameBanner(container)

      banner.showLose()

      const bannerContainer = container.children.find(child => 
        child.label === 'game-banner'
      )
      const textElement = bannerContainer?.children.find(child => 
        child.label === 'game-banner-text'
      )
      expect(textElement).toBeDefined()
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toContain('Over')
      }
    })

    it('displays correct lose message text', () => {
      const banner = new GameBanner(container, {
        loseMessage: 'Game Over!'
      })

      banner.showLose()

      const bannerContainer = container.children.find(child => 
        child.label === 'game-banner'
      )
      const textElement = bannerContainer?.children.find(child => 
        child.label === 'game-banner-text'
      )
      expect(textElement).toBeDefined()
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toBe('Game Over!')
      }
    })
  })

  describe('Hide Banner', () => {
    it('hides banner when hide() called', () => {
      const banner = new GameBanner(container)

      banner.showWin()
      banner.hide()

      const bannerContainer = container.children.find(child => 
        child.label === 'game-banner'
      )
      expect(bannerContainer).toBeDefined()
      if (bannerContainer && 'visible' in bannerContainer) {
        expect(bannerContainer.visible).toBe(false)
      }
    })
  })

  describe('State Transitions', () => {
    it('handles state transitions (show -> hide -> show)', () => {
      const banner = new GameBanner(container)

      banner.showWin()
      const bannerContainer1 = container.children.find(child => 
        child.label === 'game-banner'
      )
      const textElement1 = bannerContainer1?.children.find(child => 
        child.label === 'game-banner-text'
      )
      expect(textElement1).toBeDefined()
      if (textElement1 && 'text' in textElement1) {
        expect(textElement1.text).toContain('Win')
      }

      banner.hide()
      expect(bannerContainer1).toBeDefined()
      if (bannerContainer1 && 'visible' in bannerContainer1) {
        expect(bannerContainer1.visible).toBe(false)
      }

      banner.showLose()
      const bannerContainer2 = container.children.find(child => 
        child.label === 'game-banner'
      )
      const textElement2 = bannerContainer2?.children.find(child => 
        child.label === 'game-banner-text'
      )
      expect(textElement2).toBeDefined()
      if (textElement2 && 'text' in textElement2) {
        expect(textElement2.text).toContain('Over')
      }
    })
  })

  describe('Destruction', () => {
    it('destroys banner and cleans up resources', () => {
      const banner = new GameBanner(container)
      const initialChildCount = container.children.length

      banner.destroy()

      expect(container.children.length).toBeLessThan(initialChildCount)
    })
  })
})

