/**
 * Integration tests for Pallet Counter component.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { Container } from 'pixi.js'
import { PalletCounter } from './pallet-counter'

describe('Pallet Counter', () => {
  let container: Container

  beforeEach(() => {
    container = new Container()
  })

  afterEach(() => {
    container.destroy({ children: true })
  })

  describe('Creation', () => {
    it('creates pallet counter with default configuration', () => {
      const counter = new PalletCounter(container)

      expect(counter).toBeDefined()
      expect(container.children.length).toBeGreaterThan(0)
    })

    it('creates pallet counter with custom configuration', () => {
      const counter = new PalletCounter(container, {
        x: 100,
        y: 50
      })

      expect(counter).toBeDefined()
      expect(container.children.length).toBeGreaterThan(0)
    })
  })

  describe('Counter Updates', () => {
    it('updates counter text with active pallet count', () => {
      const counter = new PalletCounter(container)

      counter.update(5, 10)

      const counterContainer = container.children.find(child => 
        child.label === 'pallet-counter'
      )
      expect(counterContainer).toBeDefined()
      const textElement = counterContainer?.children.find(child => 
        child.label === 'pallet-counter-text'
      )
      expect(textElement).toBeDefined()
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toContain('5')
      }
    })

    it('handles zero active pallets', () => {
      const counter = new PalletCounter(container)

      counter.update(0, 10)

      const counterContainer = container.children.find(child => 
        child.label === 'pallet-counter'
      )
      const textElement = counterContainer?.children.find(child => 
        child.label === 'pallet-counter-text'
      )
      expect(textElement).toBeDefined()
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toContain('0')
      }
    })

    it('handles all pallets collected (active = 0, total > 0)', () => {
      const counter = new PalletCounter(container)

      counter.update(0, 5)

      const counterContainer = container.children.find(child => 
        child.label === 'pallet-counter'
      )
      const textElement = counterContainer?.children.find(child => 
        child.label === 'pallet-counter-text'
      )
      expect(textElement).toBeDefined()
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toContain('0')
      }
    })

    it('handles empty pallets array (active = 0, total = 0)', () => {
      const counter = new PalletCounter(container)

      counter.update(0, 0)

      const counterContainer = container.children.find(child => 
        child.label === 'pallet-counter'
      )
      const textElement = counterContainer?.children.find(child => 
        child.label === 'pallet-counter-text'
      )
      expect(textElement).toBeDefined()
      if (textElement && 'text' in textElement) {
        expect(textElement.text).toContain('0')
      }
    })

    it('updates counter when pallet state changes', () => {
      const counter = new PalletCounter(container)

      counter.update(5, 10)
      const counterContainer1 = container.children.find(child => 
        child.label === 'pallet-counter'
      )
      const textElement1 = counterContainer1?.children.find(child => 
        child.label === 'pallet-counter-text'
      )
      expect(textElement1).toBeDefined()
      if (textElement1 && 'text' in textElement1) {
        expect(textElement1.text).toContain('5')
      }

      counter.update(3, 10)
      const counterContainer2 = container.children.find(child => 
        child.label === 'pallet-counter'
      )
      const textElement2 = counterContainer2?.children.find(child => 
        child.label === 'pallet-counter-text'
      )
      expect(textElement2).toBeDefined()
      if (textElement2 && 'text' in textElement2) {
        expect(textElement2.text).toContain('3')
      }
    })
  })

  describe('Destruction', () => {
    it('destroys counter and cleans up resources', () => {
      const counter = new PalletCounter(container)
      const initialChildCount = container.children.length

      counter.destroy()

      expect(container.children.length).toBeLessThan(initialChildCount)
    })
  })
})

